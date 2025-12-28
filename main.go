package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gaming-leaderboard/config"
	"gaming-leaderboard/initilizer"
	"gaming-leaderboard/router"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// Root context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load config & initialize dependencies
	config.InitConfig()
	initilizer.Initialize(ctx)

	// Create Gin engine
	app := gin.New()
	app.Use(gin.Recovery())

	// Register routes
	router.RegisterPublicRoutes(ctx, app)

	// HTTP server with timeouts
	server := &http.Server{
		Addr:         viper.GetString("server.port"),
		Handler:      app,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("[INFO] server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] server failed: %v", err)
		}
	}()

	// Listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("[INFO] shutdown signal received")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] server shutdown failed: %v", err)
	}

	// Cancel root context (DB, Redis, etc.)
	cancel()

	log.Println("[INFO] server gracefully stopped")
}
