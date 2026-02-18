package config

import "time"

type (
	Config struct {
		UI UI `yaml:"ui" json:"ui"`
	}

	UI struct {
		EnableMouse    bool          `yaml:"enableMouse" json:"enableMouse"`
		SplashDuration time.Duration `yaml:"splashDuration" json:"splashDuration"`
	}
)

func newConfig() *Config {
	return &Config{
		UI: UI{
			EnableMouse:    true,
			SplashDuration: 1 * time.Second,
		},
	}
}
