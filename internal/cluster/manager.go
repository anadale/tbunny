package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path"
	"strings"
	"sync"
	"tbunny/internal/config"
	"tbunny/internal/sl"

	"gopkg.in/yaml.v3"
)

type Listener interface {
	ClusterChanged(*Cluster)
}

var (
	cluster *Cluster

	clustersConfig *clustersConfiguration
	clusters       map[string]*Config
	configFile     string
	clustersDir    string

	listeners []Listener
	mx        sync.RWMutex
)

func Init(configDir string) {
	clustersDir = path.Join(configDir, "clusters")
	configFile = path.Join(configDir, "clusters.yaml")

	clusters, clustersConfig = loadClusters(configFile, clustersDir)
}

func Clusters() map[string]*Config {
	mx.RLock()
	defer mx.RUnlock()

	return maps.Clone(clusters)
}

func ActiveClusterName() string {
	return clustersConfig.ActiveCluster
}

func Connect(name string) (*Cluster, error) {
	var cfg *Config
	var ok bool

	if cfg, ok = clusters[name]; !ok {
		return nil, fmt.Errorf("active cluster %s not found", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Current().ConnectionTimeout)
	defer cancel()

	newCluster, err := NewCluster(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cluster %s: %w", name, err)
	}

	setCluster(newCluster)

	return newCluster, nil
}

func Create(name string, parameters ConnectionParameters) error {
	mx.RLock()
	if _, ok := clusters[name]; ok {
		mx.RUnlock()
		return fmt.Errorf("cluster %s already exists", name)
	}

	mx.RUnlock()
	mx.Lock()
	defer mx.Unlock()

	clusterConfig := &Config{
		Connection: parameters,
		name:       name,
		fileName:   path.Join(clustersDir, name+".yaml"),
	}

	err := clusterConfig.save()
	if err != nil {
		return err
	}

	clusters[name] = clusterConfig

	return nil
}

func Delete(name string) error {
	mx.Lock()
	defer mx.Unlock()

	c, ok := clusters[name]
	if !ok {
		return fmt.Errorf("cluster %s not found", name)
	}

	delete(clusters, name)

	_ = os.Remove(c.fileName)

	return nil
}

func Current() *Cluster {
	mx.RLock()
	defer mx.RUnlock()

	return cluster
}

func AddListener(l Listener) {
	listeners = append(listeners, l)
}

func RemoveListener(l Listener) {
	for i, l2 := range listeners {
		if l2 == l {
			listeners = append(listeners[:i], listeners[i+1:]...)
			return
		}
	}
}

func loadClusters(configFile, clustersDir string) (clusters map[string]*Config, config *clustersConfiguration) {
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

				var clusterConfig *Config

				err = yaml.Unmarshal(content, &clusterConfig)
				if err != nil {
					slog.Error("Failed to parse cluster config", sl.Error, err, sl.File, name)
					continue
				}

				clusterConfig.migrate()

				clusterConfig.name = name
				clusterConfig.fileName = clusterFile

				clusters[name] = clusterConfig
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
		config = &clustersConfiguration{}
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

func setCluster(c *Cluster) {
	var oldCluster *Cluster

	mx.Lock()

	oldCluster = cluster
	cluster = c

	var shouldSave bool

	if cluster != nil {
		if clustersConfig.ActiveCluster != c.Name() {
			clustersConfig.ActiveCluster = c.Name()
			shouldSave = true
		}
	} else {
		if clustersConfig.ActiveCluster != "" {
			clustersConfig.ActiveCluster = ""
			shouldSave = true
		}
	}

	mx.Unlock()

	if oldCluster != nil {
		oldCluster.stop()
	}

	if c != nil {
		c.start()
	}

	if shouldSave {
		saveClustersConfig()
	}

	notifyClusterChanged()
}

func notifyClusterChanged() {
	for _, l := range listeners {
		l.ClusterChanged(cluster)
	}
}

func saveClustersConfig() {
	mx.RLock()
	defer mx.RUnlock()

	content, err := yaml.Marshal(clustersConfig)
	if err != nil {
		slog.Error("Failed to marshal cluster config", sl.Error, err, sl.File, configFile)
		return
	}

	err = os.MkdirAll(clustersDir, 0755)
	if err != nil {
		slog.Error("Failed to create clusters directory", sl.Error, err, sl.File, clustersDir)
		return
	}

	err = os.WriteFile(configFile, content, 0600)
	if err != nil {
		slog.Error("Failed to save cluster config", sl.Error, err, sl.File, configFile)
	}

	slog.Info("Saved clusters config", sl.File, configFile)
}
