package cluster

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"tbunny/internal/rmq"
	"tbunny/internal/sl"
	"time"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type VirtualHostsListener interface {
	ClusterVirtualHostsChanged(cluster *Cluster)
}

type ActiveVirtualHostListener interface {
	ClusterActiveVirtualHostChanged(cluster *Cluster)
}

type InformationListener interface {
	ClusterInformationChanged(cluster *Cluster)
}

type ConnectionListener interface {
	ClusterConnectionLost(cluster *Cluster)
	ClusterConnectionRestored(cluster *Cluster)
}

type Cluster struct {
	*rmq.Client

	config       *Config
	info         Information
	virtualHosts []rabbithole.VhostInfo
	listeners    []any
	errorCount   atomic.Int32
	pollChan     chan struct{}
	connection   connection
	mx           sync.RWMutex
}

type Information struct {
	Name              string
	Username          string
	ClusterName       string
	RabbitMQVersion   string
	ErlangVersion     string
	ManagementVersion string
}

func NewCluster(ctx context.Context, cfg *Config) (c *Cluster, err error) {
	var client *rmq.Client

	conn, err := cfg.Connection.createConnection(ctx)
	if err != nil {
		return nil, err
	}

	uri := conn.Uri()
	if strings.HasPrefix(strings.ToLower(uri), "https://") {
		transport := &http.Transport{TLSClientConfig: &tls.Config{}}
		client, err = rmq.NewTLSClient(uri, cfg.Connection.Username, cfg.Connection.Password, transport)
	} else {
		client, err = rmq.NewClient(uri, cfg.Connection.Username, cfg.Connection.Password)
	}

	if err != nil {
		return nil, err
	}

	info, err := getClusterInfo(client, cfg)
	if err != nil {
		return nil, err
	}

	vhosts, err := client.ListVhosts()
	if err != nil {
		return nil, err
	}

	if cfg.Vhost != "" {
		vhostOk := slices.ContainsFunc(vhosts, func(vhost rabbithole.VhostInfo) bool {
			return vhost.Name == cfg.Vhost
		})
		if !vhostOk {
			cfg.Vhost = ""
		}
	}

	c = &Cluster{
		Client:       client,
		connection:   conn,
		config:       cfg,
		info:         info,
		virtualHosts: vhosts,
	}

	return c, nil
}

func (c *Cluster) IsAvailable() bool {
	return c.errorCount.Load() == 0
}

func (c *Cluster) AddListener(l any) {
	c.listeners = append(c.listeners, l)
}

func (c *Cluster) RemoveListener(l any) {
	for i, l2 := range c.listeners {
		if l2 == l {
			c.listeners = append(c.listeners[:i], c.listeners[i+1:]...)
			return
		}
	}
}

func (c *Cluster) Name() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.config.name
}

func (c *Cluster) Username() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.config.Connection.Username
}

func (c *Cluster) Information() Information {
	return c.info
}

func (c *Cluster) VirtualHosts() []rabbithole.VhostInfo {
	return c.virtualHosts
}

func (c *Cluster) FavoriteVhosts() []string {
	return c.config.FavoriteVhosts
}

func (c *Cluster) ActiveVirtualHost() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.config.Vhost
}

func (c *Cluster) SetActiveVirtualHost(vhost string) {
	c.mx.Lock()

	if c.config.Vhost == vhost {
		c.mx.Unlock()
		return
	}

	c.config.Vhost = vhost
	if vhost != "" {
		if len(c.config.FavoriteVhosts) < 9 && !slices.Contains(c.config.FavoriteVhosts, vhost) {
			c.config.FavoriteVhosts = append(c.config.FavoriteVhosts, vhost)
		}
	}

	c.saveConfig()

	c.mx.Unlock()

	c.notifyActiveVirtualHostChanged()
}

func (c *Cluster) DeleteVhost(name string) (res *http.Response, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	res, err = c.Client.DeleteVhost(name)
	if err != nil {
		return nil, err
	}

	if slices.Contains(c.config.FavoriteVhosts, name) {
		c.config.FavoriteVhosts = slices.DeleteFunc(c.config.FavoriteVhosts, func(v string) bool { return v == name })
		c.saveConfig()
	}

	return res, nil
}
func (c *Cluster) Refresh() {
	if c.pollChan != nil {
		select {
		case c.pollChan <- struct{}{}:
		default:
			// Refresh already requested
		}
	}
}

func (c *Cluster) start() {
	c.startPolling()
}

func (c *Cluster) stop() {
	c.connection.Close()
	c.stopPolling()
}

func (c *Cluster) startPolling() {
	if c.pollChan != nil {
		return
	}

	c.pollChan = make(chan struct{}, 1)

	go c.poll(c.pollChan)
}

func (c *Cluster) stopPolling() {
	if c.pollChan == nil {
		return
	}

	close(c.pollChan)

	c.pollChan = nil
}

func (c *Cluster) poll(ch chan struct{}) {
	slog.Debug("Cluster availability monitoring started", sl.Cluster, c.config.name)

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				slog.Debug("Cluster availability monitoring stopped", sl.Cluster, c.config.name)
				return
			}
			slog.Debug("Cluster information refresh triggered", sl.Cluster, c.config.name)
		case <-time.After(5 * time.Second):
		}

		c.probeConnection()
	}
}

func (c *Cluster) probeConnection() {
	var info Information
	var vhosts []rabbithole.VhostInfo
	var err error

	info, err = getClusterInfo(c.Client, c.config)
	if err == nil {
		vhosts, err = c.Client.ListVhosts()
	}

	if err != nil {
		if c.errorCount.Add(1) == 2 {
			slog.Debug("Cluster connection lost", sl.Cluster, c.config.name)
			c.notifyConnectionLost()
		}

		return
	}

	old := c.errorCount.Swap(0)
	if old > 1 {
		slog.Debug("Cluster connection restored", sl.Cluster, c.config.name)
		c.notifyConnectionRestored()
	}

	isInformationChanged := c.info != info
	isVirtualHostsChanged := !isEqualVhosts(c.virtualHosts, vhosts)

	if !isInformationChanged && !isVirtualHostsChanged {
		return
	}

	c.mx.Lock()

	if isInformationChanged {
		c.info = info
	}

	if isVirtualHostsChanged {
		c.virtualHosts = vhosts

		if c.sanitizeFavoriteVhosts() {
			c.saveConfig()
		}

		c.notifyVirtualHostsChanged()
	}

	c.mx.Unlock()

	if isInformationChanged {
		slog.Debug(fmt.Sprintf("Cluster name has been changed to %s", info.ClusterName), sl.Cluster, c.config.name)
		c.notifyInformationChanged()
	}

	if isVirtualHostsChanged {
		slog.Debug("List of cluster virtual hosts has been changed", sl.Cluster, c.config.name)
		c.notifyVirtualHostsChanged()
	}
}

func (c *Cluster) notifyActiveVirtualHostChanged() {
	for _, l := range c.listeners {
		if t, ok := l.(ActiveVirtualHostListener); ok {
			t.ClusterActiveVirtualHostChanged(c)
		}
	}
}

func (c *Cluster) notifyVirtualHostsChanged() {
	for _, l := range c.listeners {
		if t, ok := l.(VirtualHostsListener); ok {
			t.ClusterVirtualHostsChanged(c)
		}
	}
}

func (c *Cluster) notifyInformationChanged() {
	for _, l := range c.listeners {
		if t, ok := l.(InformationListener); ok {
			t.ClusterInformationChanged(c)
		}
	}
}

func (c *Cluster) notifyConnectionLost() {
	for _, l := range c.listeners {
		if t, ok := l.(ConnectionListener); ok {
			t.ClusterConnectionLost(c)
		}
	}
}

func (c *Cluster) notifyConnectionRestored() {
	for _, l := range c.listeners {
		if t, ok := l.(ConnectionListener); ok {
			t.ClusterConnectionRestored(c)
		}
	}
}

func (c *Cluster) saveConfig() {
	err := c.config.save()
	if err != nil {
		slog.Error("Failed to save cluster config", sl.Cluster, c.config.name, sl.Error, err)
		return
	}

	slog.Info("Saved cluster config", sl.Cluster, c.config.name)
}

func (c *Cluster) sanitizeFavoriteVhosts() bool {
	var victims []string

	for _, favorite := range c.config.FavoriteVhosts {
		if !slices.ContainsFunc(c.virtualHosts, func(vhost rabbithole.VhostInfo) bool { return vhost.Name == favorite }) {
			victims = append(victims, favorite)
		}
	}

	if len(victims) == 0 {
		return false
	}

	favorites := make([]string, 0, len(c.config.FavoriteVhosts)-len(victims))
	for _, favorite := range c.config.FavoriteVhosts {
		if !slices.Contains(victims, favorite) {
			favorites = append(favorites, favorite)
		}
	}

	c.config.FavoriteVhosts = favorites

	slog.Info("Sanitized favorite vhosts", sl.Cluster, c.config.name)

	return true
}

func isEqualVhosts(vhosts1 []rabbithole.VhostInfo, vhosts2 []rabbithole.VhostInfo) bool {
	if len(vhosts1) != len(vhosts2) {
		return false
	}

	for i, vhost1 := range vhosts1 {
		if vhost1.Name != vhosts2[i].Name {
			return false
		}
	}

	return true
}

func getClusterInfo(client *rmq.Client, config *Config) (Information, error) {
	clusterName, err := client.GetClusterName()
	if err != nil {
		return Information{}, err
	}

	overview, err := client.Overview()
	if err != nil {
		return Information{}, err
	}

	info := Information{
		Name:              config.name,
		ClusterName:       clusterName.Name,
		Username:          config.Connection.Username,
		RabbitMQVersion:   overview.RabbitMQVersion,
		ErlangVersion:     overview.ErlangVersion,
		ManagementVersion: overview.ManagementVersion,
	}

	return info, nil
}
