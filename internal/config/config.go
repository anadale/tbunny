package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

type Config struct {
	UI UI `yaml:"ui" json:"ui"`
}

type UI struct {
	EnableMouse    bool          `yaml:"enableMouse" json:"enableMouse"`
	SplashDuration time.Duration `yaml:"splashDuration" json:"splashDuration"`
}

type Listener interface {
	ConfigChanged(*Config)
}

var (
	config        *Config
	rootDirectory string
	listeners     []Listener
)

func Init(configRoot string) {
	var err error

	if configRoot != "" {
		if strings.HasPrefix(configRoot, "~") {
			var home string

			home, err = os.UserHomeDir()
			if err != nil {
				panic(fmt.Sprintf("unable to get user home directory: %v", err))
			}

			configRoot = filepath.Join(home, strings.TrimPrefix(configRoot, "~"))
		}

		rootDirectory, err = filepath.Abs(configRoot)
		if err != nil {
			panic(fmt.Sprintf("unable to get absolute path for config root: %v", err))
		}
	} else {
		rootDirectory = filepath.Join(xdg.ConfigHome, "tbunny")
	}

	configFile := filepath.Join(rootDirectory, "config.yaml")

	config, err = loadConfigFromFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			config = newConfig()
		} else {
			panic(fmt.Sprintf("unable to load configuration from file %s: %v", configFile, err))
		}
	}
}

func Current() *Config {
	return config
}

// RootDirectory returns the root directory of the configuration.
func RootDirectory() string {
	return rootDirectory
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

func loadConfigFromFile(filePath string) (config *Config, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, &config)

	return config, err
}

func newConfig() *Config {
	return &Config{
		UI: UI{
			EnableMouse:    true,
			SplashDuration: 1 * time.Second,
		},
	}
}

func notifyConfigChanged() {
	for _, l := range listeners {
		l.ConfigChanged(config)
	}
}
