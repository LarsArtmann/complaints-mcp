package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/server"
	"github.com/spf13/viper"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize configuration
	initConfig()

	// Create MCP server
	srv := server.New()

	// TODO: Register tools and handlers
	// tools := []server.Tool{
	// 	{
	// 		Name:        "file_complaint",
	// 		Description: "File a structured complaint about missing or confusing information",
	// 		InputSchema: map[string]interface{}{
	// 			"type": "object",
	// 			"properties": map[string]interface{}{
	// 				"agent_name":        {"type": "string"},
	// 				"task_description": {"type": "string"},
	// 				"severity":          {"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
	// 			},
	// 			"required": []string{"agent_name", "task_description", "severity"},
	// 		},
	// 	},
	// }

	// Start server in goroutine
	go func() {
		if err := srv.Serve(ctx); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")

	// Graceful shutdown
	cancel()
	
	log.Println("Server stopped")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.complaints-mcp")
	
	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log.level", "info")
	
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults")
		} else {
			log.Printf("Error reading config file: %v", err)
		}
	}
	
	flag.Parse()
}