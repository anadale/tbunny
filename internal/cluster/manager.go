package cluster

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path"
	"strings"
	"sync"
	"tbunny/internal/sl"

	"gopkg.in/yaml.v3"
)

type ManagerListener interface {
	ClusterChanged(*Cluster)
}

type Manager struct {
	cluster *Cluster

	config      *clustersConfig
	clusters    map[string]*Config
	configFile  string
	clustersDir string

	listeners []ManagerListener
	mx        sync.RWMutex
}

func NewManager(configDir string) *Manager {
	clustersDir := path.Join(configDir, "clusters")
	configFile := path.Join(configDir, "clusters.yaml")

	clusters, config := loadClusters(configFile, clustersDir)

	m := Manager{
		config:      config,
		clusters:    clusters,
		configFile:  configFile,
		clustersDir: clustersDir,
	}

	return &m
}

func loadClusters(configFile, clustersDir string) (clusters map[string]*Config, config *clustersConfig) {
	var content []byte

	clusters = make(map[string]*Config)

	entries, err := os.ReadDir(clustersDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()

			if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
				clusterFile := path.Join(clustersDir, name)
				name = strings.TrimSuffix(name, path.Ext(name))

				content, err = os.ReadFile(clusterFile)
				if err != nil {
					slog.Error("Failed to read cluster config", sl.Error, err, sl.File, name)
					continue
				}

				var cluster *Config

				err = yaml.Unmarshal(content, &cluster)
				if err != nil {
					slog.Error("Failed to parse cluster config", sl.Error, err, sl.File, name)
					continue
				}

				cluster.name = name
				cluster.fileName = clusterFile

				clusters[name] = cluster
			}
		}
	}

	content, err = os.ReadFile(configFile)
	if err == nil {
		err = yaml.Unmarshal(content, &config)
		if err != nil {
			slog.Error("Failed to parse clusters config file", sl.Error, err, sl.File, configFile)
		}
	} else {
		slog.Error("Failed to read clusters config file", sl.Error, err, sl.File, configFile)
	}

	if config == nil {
		config = &clustersConfig{}
	}

	_, ok := clusters[config.ActiveCluster]
	if !ok {
		slog.Error("Active cluster not found", sl.Cluster, config.ActiveCluster)

		for key := range clusters {
			slog.Debug("Settings active cluster", sl.Cluster, key)

			config.ActiveCluster = key
			break
		}
	}

	return clusters, config
}

func (m *Manager) GetClusters() map[string]*Config {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return maps.Clone(m.clusters)
}

func (m *Manager) GetActiveClusterName() string {
	return m.config.ActiveCluster
}

func (m *Manager) ConnectToCluster(name string) (*Cluster, error) {
	var cfg *Config
	var ok bool

	if cfg, ok = m.clusters[name]; !ok {
		return nil, fmt.Errorf("active cluster %s not found", name)
	}

	newCluster, err := NewCluster(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cluster %s: %w", name, err)
	}

	m.setCluster(newCluster)

	return newCluster, nil
}

func (m *Manager) Create(name string, parameters ConnectionParameters) error {
	if _, ok := m.clusters[name]; ok {
		return fmt.Errorf("cluster %s already exists", name)
	}

	config := &Config{
		Connection: parameters,
		name:       name,
		fileName:   path.Join(m.clustersDir, name+".yaml"),
	}

	err := config.save()
	if err != nil {
		return err
	}

	m.clusters[name] = config

	return nil
}

func (m *Manager) Delete(name string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	cluster, ok := m.clusters[name]
	if !ok {
		return fmt.Errorf("cluster %s not found", name)
	}

	delete(m.clusters, name)

	_ = os.Remove(cluster.fileName)

	return nil
}

func (m *Manager) Cluster() *Cluster {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return m.cluster
}

func (m *Manager) AddListener(l ManagerListener) {
	m.listeners = append(m.listeners, l)
}

func (m *Manager) RemoveListener(l ManagerListener) {
	for i, l2 := range m.listeners {
		if l2 == l {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			return
		}
	}
}

func (m *Manager) setCluster(c *Cluster) {
	var oldCluster *Cluster

	m.mx.Lock()

	oldCluster = m.cluster
	m.cluster = c

	var shouldSave bool

	if m.cluster != nil {
		if m.config.ActiveCluster != c.Name() {
			m.config.ActiveCluster = c.Name()
			shouldSave = true
		}
	} else {
		if m.config.ActiveCluster != "" {
			m.config.ActiveCluster = ""
			shouldSave = true
		}
	}

	m.mx.Unlock()

	if oldCluster != nil {
		oldCluster.stopPolling()
	}

	if c != nil {
		c.startPolling()
	}

	if shouldSave {
		m.saveConfig()
	}

	m.notifyClusterChanged()
}

func (m *Manager) notifyClusterChanged() {
	for _, l := range m.listeners {
		l.ClusterChanged(m.cluster)
	}
}

func (m *Manager) saveConfig() {
	m.mx.RLock()
	defer m.mx.RUnlock()

	content, err := yaml.Marshal(m.config)
	if err != nil {
		slog.Error("Failed to marshal cluster config", sl.Error, err, sl.File, m.configFile)
		return
	}

	err = os.MkdirAll(m.clustersDir, 0755)
	if err != nil {
		slog.Error("Failed to create clusters directory", sl.Error, err, sl.File, m.clustersDir)
		return
	}

	err = os.WriteFile(m.configFile, content, 0600)
	if err != nil {
		slog.Error("Failed to save cluster config", sl.Error, err, sl.File, m.configFile)
	}

	slog.Info("Saved clusters config", sl.File, m.configFile)
}
