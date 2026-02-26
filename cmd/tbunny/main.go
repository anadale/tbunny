package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"tbunny/internal/cluster"
	"tbunny/internal/config"
	"tbunny/internal/sl"
	"tbunny/internal/view/application"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:     "tbunny",
		Short:   "A fast, keyboard-driven terminal UI for managing RabbitMQ clusters",
		Version: version,
		RunE:    run,
	}
	logFilePath string
	configDir   string
)

func init() {
	initFlags()
}

func initFlags() {
	rootCmd.SetVersionTemplate("TBunny version {{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&logFilePath, "log-file", "", "Specify the log file")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "Specify the configuration directory")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) error {
	var logHandler slog.Handler

	if len(logFilePath) > 0 {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open log file %q: %w", logFilePath, err)
		}

		defer func() {
			if logFile != nil {
				_ = logFile.Close()
			}
		}()

		logHandler = tint.NewHandler(logFile, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
			NoColor:    true,
		})
	} else {
		logHandler = slog.DiscardHandler
	}

	var initialized bool

	defer func() {
		if r := recover(); r != nil {
			if !initialized {
				slog.Error("TBunny initialization failed", sl.Error, r, sl.Stack, string(debug.Stack()))
				fmt.Printf("Unable to initialize: %v\n", r)
			} else {
				slog.Error("TBunny runtime error", sl.Error, r, sl.Stack, string(debug.Stack()))
				fmt.Printf("Unexpected error: %v\n", r)
			}
		}
	}()

	slog.SetDefault(slog.New(logHandler))

	cfm := loadConfiguration()
	clm := createClusterManager(cfm)

	app := application.NewApp(clm, cfm, version)

	if err := app.Init(); err != nil {
		return err
	}

	initialized = true

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}

func loadConfiguration() *config.Manager {
	return config.NewManager(configDir)
}

func createClusterManager(cfm *config.Manager) *cluster.Manager {
	clm := cluster.NewManager(cfm.ConfigDir())

	activeClusterName := clm.GetActiveClusterName()
	if activeClusterName != "" {
		if _, err := clm.ConnectToCluster(activeClusterName); err != nil {
			slog.Error("Failed to connect to active RabbitMQ cluster", sl.Cluster, activeClusterName, sl.Error, err)
		}
	}

	return clm
}
