package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"naevis/middleware"
	"naevis/ratelim"
	"naevis/routes"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

// Index is a simple health check handler.
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "200")
}

// setupRouter builds the router with all API and static routes.
func setupRouter(rateLimiter *ratelim.RateLimiter) *httprouter.Router {
	router := httprouter.New()
	router.GET("/health", Index)

	// API routes
	routes.AddFiledropRoutes(router, rateLimiter)

	// Static routes
	routes.AddStaticRoutes(router)

	return router
}

// parseAllowedOrigins parses comma-separated origins from environment variable.
func parseAllowedOrigins(env string) []string {
	if env == "" {
		return []string{
			"http://localhost:5173",
			"https://indium.netlify.app",
		}
	}
	parts := strings.Split(env, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func main() {
	// Load environment variables from .env if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment")
	}

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = ":6925"
	} else if port[0] != ':' {
		port = ":" + port
	}

	// Parse allowed origins
	allowedOrigins := parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))

	// Initialize rate limiter
	rateLimiter := ratelim.NewRateLimiter(1, 6, 10*time.Minute, 10000)

	// Build router with API + static routes
	router := setupRouter(rateLimiter)

	// Middleware chain: SecurityHeaders → Logging → Router
	innerHandler := middleware.LoggingMiddleware(middleware.SecurityHeaders(router))

	// CORS applied outermost
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Idempotency-Key", "X-Requested-With"},
		AllowCredentials: false,
	}).Handler(innerHandler)

	// Multiplexer: /health bypasses CORS
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "200")
	})
	mux.Handle("/", corsHandler)

	// Configure HTTP server
	server := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadTimeout:       7 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Graceful shutdown registration
	server.RegisterOnShutdown(func() {
		log.Println("🛑 Server shutting down...")
	})

	// Start server
	go func() {
		log.Printf("🚀 Server running on %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt or SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// Begin graceful shutdown
	log.Println("🛑 Shutdown signal received; stopping gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Graceful shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
