package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Est3banj/phone-tracker/internal/adapters/handler"
	"github.com/Est3banj/phone-tracker/internal/adapters/repository"
	"github.com/Est3banj/phone-tracker/internal/config"
	"github.com/Est3banj/phone-tracker/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting phone-tracker server...")

	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize repos
	userRepo := repository.NewUserRepo(db)
	deviceRepo := repository.NewDeviceRepo(db)
	tokenRepo := repository.NewTokenRepo(db)
	locationRepo := repository.NewLocationRepo(db)
	eventRepo := repository.NewEventRepo(db)
	commandRepo := repository.NewCommandRepo(db)

	// Initialize hub
	hub := handler.NewHub()

	// Initialize services
	trackerService := service.NewTracker(locationRepo, eventRepo, hub)
	commanderService := service.NewCommander(commandRepo, hub)

	hub.SetTracker(trackerService)
	hub.SetCommander(commanderService)
	hub.SetRepos(locationRepo, eventRepo)

	// Create background context for loops
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start ping loop
	go hub.PingLoop(ctx, cfg.WSPingInterval)

	// Start command timeout loop
	go commanderService.TimeoutLoop(ctx, 30*time.Second)

	// Initialize middleware
	mw := handler.NewMiddleware(userRepo, cfg.JWTSecret)

	// Initialize auth handler
	authHandler := handler.NewAuthHandler(cfg, userRepo, deviceRepo, tokenRepo)

	// Initialize HTTP server
	httpServer := handler.NewHTTPServer(hub, authHandler, mw, cfg, userRepo, deviceRepo, tokenRepo)

	// Register routes
	mux := http.NewServeMux()
	httpServer.RegisterRoutes(mux)

	// Wrap with CORS
	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: mw.CORS(mux),
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		srv.Close()
	}()

	log.Printf("Server listening on %s", cfg.ServerAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
