package api

import (
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/philoveracity/pvdifyd/internal/config"
	"github.com/philoveracity/pvdifyd/internal/db"
)

// Server represents the HTTP API server
type Server struct {
	router *chi.Mux
	db     *db.DB
	cfg    *config.Config
	logger *slog.Logger
}

// New creates a new API server
func New(database *db.DB, cfg *config.Config, logger *slog.Logger) *Server {
	s := &Server{
		router: chi.NewRouter(),
		db:     database,
		cfg:    cfg,
		logger: logger,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	r := s.router

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(s.loggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check (no auth)
	r.Get("/health", s.handleHealth)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// CORS for Admin UI
		r.Use(corsMiddleware)

		// Cloudflare integration
		r.Route("/cloudflare", func(r chi.Router) {
			r.Get("/zones", s.handleListCloudflareZones)
			r.Get("/dns", s.handleListCloudflareDNS)
		})

		// Apps
		r.Route("/apps", func(r chi.Router) {
			r.Get("/", s.handleListApps)
			r.Post("/", s.handleCreateApp)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", s.handleGetApp)
				r.Patch("/", s.handleUpdateApp)
				r.Delete("/", s.handleDeleteApp)

				// Releases
				r.Route("/releases", func(r chi.Router) {
					r.Get("/", s.handleListReleases)
					r.Post("/", s.handleCreateRelease)
					r.Get("/{version}", s.handleGetRelease)
				})
				r.Post("/rollback", s.handleRollback)

				// Config
				r.Route("/config", func(r chi.Router) {
					r.Get("/", s.handleGetConfig)
					r.Put("/", s.handleSetConfig)
					r.Delete("/{key}", s.handleUnsetConfig)
				})

				// Domains
				r.Route("/domains", func(r chi.Router) {
					r.Get("/", s.handleListDomains)
					r.Post("/", s.handleAddDomain)
					r.Delete("/{domain}", s.handleRemoveDomain)
					// Cloudflare DNS integration
					r.Post("/{domain}/cloudflare", s.handleCreateCloudflareDNS)
				})

				// Processes
				r.Route("/ps", func(r chi.Router) {
					r.Get("/", s.handleListProcesses)
					r.Post("/scale", s.handleScale)
					r.Post("/restart", s.handleRestart)
				})

				// Logs
				r.Get("/logs", s.handleLogs)
			})
		})
	})

	// Serve Admin UI static files (SPA with fallback to index.html)
	if s.cfg.StaticDir != "" {
		s.logger.Info("serving static files", "dir", s.cfg.StaticDir)
		r.Get("/*", s.spaHandler(s.cfg.StaticDir))
	}
}

// loggerMiddleware logs requests
func (s *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start).String(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// ListenAndServe starts the server
func (s *Server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.cfg.Listen,
		Handler: s.router,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	s.logger.Info("starting server", "addr", s.cfg.Listen)

	if s.cfg.TLS.Enabled && !s.cfg.Dev {
		return srv.ListenAndServeTLS(s.cfg.TLS.Cert, s.cfg.TLS.Key)
	}
	return srv.ListenAndServe()
}

// JSON response helpers
func (s *Server) json(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) error(w http.ResponseWriter, status int, message string) {
	s.json(w, status, map[string]string{"error": message})
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.json(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": "0.1.0",
	})
}

// corsMiddleware handles CORS for API requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// spaHandler serves static files with SPA fallback to index.html
func (s *Server) spaHandler(staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = ""
		}

		// Try to serve the exact file
		fullPath := filepath.Join(staticDir, path)
		info, err := os.Stat(fullPath)

		// If file exists and is not a directory, serve it directly
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, fullPath)
			return
		}

		// For paths without extension, try adding .html
		if !strings.Contains(filepath.Base(path), ".") && path != "" {
			htmlPath := fullPath + ".html"
			if _, err := os.Stat(htmlPath); err == nil {
				http.ServeFile(w, r, htmlPath)
				return
			}
		}

		// Fallback to root index.html for SPA routing
		indexPath := filepath.Join(staticDir, "index.html")
		http.ServeFile(w, r, indexPath)
	}
}

// Silence unused import warning
var _ = fs.FS(nil)
