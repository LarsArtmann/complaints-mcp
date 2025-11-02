package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/complaints-mcp/internal/config"
	mcpdelivery "github.com/larsartmann/complaints-mcp/internal/delivery/mcp"
	"github.com/larsartmann/complaints-mcp/internal/repo"
	"github.com/larsartmann/complaints-mcp/internal/service"
	"github.com/larsartmann/complaints-mcp/internal/tracing"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "complaints-mcp",
	Short: "MCP server for filing structured complaints",
	Long:  `A Model Context Protocol server that allows AI agents to file structured complaints about missing or confusing information they encounter during development tasks.`,
	RunE:  runServer,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (trace, debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("dev", "d", false, "development mode")
	rootCmd.PersistentFlags().Bool("version", false, "show version information")

	// Cache configuration flags
	rootCmd.PersistentFlags().Bool("cache-enabled", true, "enable complaint caching for performance")
	rootCmd.PersistentFlags().Int("cache-max-size", 1000, "maximum number of complaints to cache")
	rootCmd.PersistentFlags().String("cache-eviction", "lru", "cache eviction policy (lru, fifo, none)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check version flag
	showVersion, _ := cmd.Flags().GetBool("version")
	if showVersion {
		fmt.Printf("complaints-mcp version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		return nil
	}

	// Initialize logging
	logLevel, _ := cmd.Flags().GetString("log-level")
	devMode, _ := cmd.Flags().GetBool("dev")

	var logger *log.Logger
	if devMode {
		level, _ := log.ParseLevel(logLevel)
		logger = log.NewWithOptions(os.Stderr, log.Options{
			Level:           level,
			ReportTimestamp: true,
			ReportCaller:    true,
		})
	} else {
		level, _ := log.ParseLevel(logLevel)
		logger = log.NewWithOptions(os.Stderr, log.Options{
			Level:           level,
			ReportTimestamp: true,
			ReportCaller:    false,
		})
	}

	ctx = log.WithContext(ctx, logger)

	logger.Info("Starting complaints-mcp server",
		"version", version,
		"commit", commit,
		"log_level", logLevel,
		"dev_mode", devMode)

	// Load configuration
	cfg, err := config.Load(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize dependencies
	tracerConfig := tracing.DefaultTracerConfig()
	tracer := tracing.NewTracer(tracerConfig)
	complaintRepo := repo.NewRepositoryFromConfig(cfg)
	complaintService := service.NewComplaintService(complaintRepo, tracer, logger)
	mcpServer := mcpdelivery.NewServer(cfg.Server.Name, version, complaintService, logger, tracer)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		// Set config for MCP server
		mcpServer.SetConfig(cfg)
		if err := mcpServer.Start(ctx); err != nil {
			serverErrChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", "signal", sig.String())

	case err := <-serverErrChan:
		logger.Error("Server error occurred", "error", err)
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	logger.Info("Initiating graceful shutdown")
	if err := mcpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
	} else {
		logger.Info("Server stopped gracefully")
	}

	return nil
}
