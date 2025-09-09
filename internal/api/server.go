package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Franklyne-kibet/aster-scheduler/internal/api/handlers"
	"github.com/Franklyne-kibet/aster-scheduler/internal/config"
	"github.com/Franklyne-kibet/aster-scheduler/internal/db/store"
)

// Server represents the HTTP API server
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, jobStore *store.JobStore, runStore *store.RunStore, logger *zap.Logger) *Server {
	// Create handlers
	jobHandler := handlers.NewJobHandler(jobStore, logger)
	runHandler := handlers.NewRunHandler(runStore, logger)

	// Create router
	router := mux.NewRouter()

	// Add middleware in order
	router.Use(panicRecoveryMiddleware(logger))
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware(cfg.AllowedOrigins))

	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Job routes
	apiRouter.HandleFunc("/jobs", jobHandler.CreateJob).Methods("POST")
	apiRouter.HandleFunc("/jobs", jobHandler.ListJobs).Methods("GET")
	apiRouter.HandleFunc("/jobs/{id}", jobHandler.GetJob).Methods("GET")
	apiRouter.HandleFunc("/jobs/{id}", jobHandler.UpdateJob).Methods("PUT")
	apiRouter.HandleFunc("/jobs/{id}", jobHandler.DeleteJob).Methods("DELETE")

	// Run routes
	apiRouter.HandleFunc("/runs", runHandler.ListRuns).Methods("GET")
	apiRouter.HandleFunc("/runs/{id}", runHandler.GetRun).Methods("GET")

	// Health check
	router.HandleFunc("/health", healthHandler).Methods("GET")

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.APIPort),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// Middleware

// panicRecoveryMiddleware recovers from panics and returns 500
func panicRecoveryMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method))

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom ResponseWriter to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(ww, r)

			// Log the request
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", ww.statusCode),
				zap.Duration("duration", time.Since(start)),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", getClientIP(r)))
		})
	}
}

// corsMiddleware adds CORS headers with configurable origins
func corsMiddleware(allowedOrigins []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

// isOriginAllowed checks if origin is in allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	return slices.Contains(allowedOrigins, origin)
}

// getClientIP extracts client IP, handling proxies
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take first IP in case of multiple
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "aster-api",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
