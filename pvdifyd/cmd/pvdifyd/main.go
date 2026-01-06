package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/philoveracity/pvdifyd/internal/api"
	"github.com/philoveracity/pvdifyd/internal/config"
	"github.com/philoveracity/pvdifyd/internal/db"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "/etc/pvdify/pvdifyd.yaml", "config file path")
		dev        = flag.Bool("dev", false, "development mode")
		version    = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *version {
		fmt.Printf("pvdifyd %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	if *dev {
		cfg.Dev = true
		cfg.Log.Level = "debug"
		cfg.Log.Format = "text"
	}

	// Setup logger
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Log.Level),
	}
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(handler)

	logger.Info("starting pvdifyd",
		"version", Version,
		"dev", cfg.Dev,
		"listen", cfg.Listen,
	)

	// Initialize database
	database, err := db.New(cfg.Database)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("database initialized", "path", cfg.Database)

	// Create state directories
	if err := ensureStateDir(cfg.StateDir); err != nil {
		logger.Error("failed to create state directory", "error", err)
		os.Exit(1)
	}

	// Create API server
	server := api.New(database, cfg, logger)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	// Start server
	if err := server.ListenAndServe(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("pvdifyd stopped")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func ensureStateDir(stateDir string) error {
	dirs := []string{
		stateDir,
		stateDir + "/apps",
		stateDir + "/releases",
		stateDir + "/config",
		stateDir + "/tunnels",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
