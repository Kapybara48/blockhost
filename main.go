package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"blockhost/internal/api"
	"blockhost/internal/mcprotocol"
	"blockhost/internal/store"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	dataDir := flag.String("data", "./data", "Directory for state persistence, properties and backups")
	webDir := flag.String("web", ".", "Directory containing static web files (index.html, style.css, app.js)")
	flag.Parse()

	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		absDataDir = *dataDir
	}

	absWebDir, err := filepath.Abs(*webDir)
	if err != nil {
		absWebDir = *webDir
	}

	log.Printf("Starting BlockHost Go Backend...")
	log.Printf("Data directory: %s", absDataDir)
	log.Printf("Web directory:  %s", absWebDir)

	// Initialize state store
	st, err := store.New(absDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer st.Close()

	// Initialize API server
	apiServer := api.NewServer(st, absWebDir)
	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)

	handler := api.Middleware(mux)

	serverAddr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // 0 for streaming SSE
	}

	// Initialize Minecraft Protocol TCP Server (Port 25565)
	mcServer := mcprotocol.NewServer(st)
	if st.GetState().ServerStatus == "online" {
		if err := mcServer.Start(); err != nil {
			log.Printf("Warning: Failed to start Minecraft TCP server on port 25565: %v", err)
		}
	}
	defer mcServer.Stop()

	// Dynamic listener state synchronization
	subChan := st.Subscribe()
	go func() {
		for state := range subChan {
			if state.ServerStatus == "online" {
				_ = mcServer.Start()
			} else if state.ServerStatus == "offline" {
				mcServer.Stop()
			}
		}
	}()

	// Server startup banner
	mcPort := st.GetState().Config.Port
	if mcPort <= 0 {
		mcPort = 25565
	}
	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("  🎮  BlockHost - Minecraft Server Management Backend")
	fmt.Println("  🚀  Backend running in Go (Golang)")
	fmt.Println("=========================================================")
	fmt.Printf("  🌐 Web Dashboard:    http://localhost:%d/\n", *port)
	fmt.Printf("  📡 REST API:         http://localhost:%d/api/state\n", *port)
	fmt.Printf("  ⚡ SSE Stream:       http://localhost:%d/api/events/stream\n", *port)
	fmt.Printf("  ⛏️  Minecraft Server: localhost:%d (TCP)\n", mcPort)
	fmt.Println("=========================================================")
	fmt.Println("  Press Ctrl+C to stop the server.")
	fmt.Println()

	// Run HTTP server in background
	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("Server error: %v", err)
	case sig := <-quit:
		log.Printf("Received signal %v. Gracefully shutting down...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown warning: %v", err)
	} else {
		log.Println("BlockHost server stopped gracefully.")
	}
}
