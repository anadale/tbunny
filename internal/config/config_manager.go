package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
)

type Listener interface {
	ConfigChanged(*Config)
}

type Manager struct {
	config    *Config
	configDir string
	listeners []Listener
}

func NewManager(configDir string) *Manager {
	if configDir == "" {
		configDir = path.Join(xdg.ConfigHome, "tbunny")
	}

	configFile := path.Join(configDir, "config.yaml")
	config, err := loadConfigFromFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			config = newConfig()
		} else {
			panic(fmt.Sprintf("unable to load configuration from file %s: %v", configFile, err))
		}
	}

	return &Manager{
		config:    config,
		configDir: configDir,
	}
}

func loadConfigFromFile(filePath string) (config *Config, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, &config)

	return config, err
}

func (m *Manager) Config() *Config {
	return m.config
}

func (m *Manager) ConfigDir() string {
	return m.configDir
}

func (m *Manager) AddListener(l Listener) {
	m.listeners = append(m.listeners, l)
}

func (m *Manager) RemoveListener(l Listener) {
	for i, l2 := range m.listeners {
		if l2 == l {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			return
		}
	}
}

func (m *Manager) notifyConfigChanged() {
	for _, l := range m.listeners {
		l.ConfigChanged(m.config)
	}
}
