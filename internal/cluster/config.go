package cluster

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Connection     ConnectionParameters `yaml:"connection" json:"connection"`
	Vhost          string               `yaml:"vhost" json:"vhost"`
	FavoriteVhosts []string             `yaml:"favoriteVhosts" json:"favoriteVhosts"`

	name     string
	fileName string
}

type ConnectionParameters struct {
	Uri      string `yaml:"uri" json:"uri"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

type clustersConfig struct {
	ActiveCluster string `yaml:"activeCluster" json:"activeCluster"`
}

func (c *Config) save() error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster config: %w", err)
	}

	err = os.WriteFile(c.fileName, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to save cluster config: %w", err)
	}

	return nil
}
