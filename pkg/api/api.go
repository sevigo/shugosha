package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/sevigo/shugosha/pkg/api/config"
	"github.com/sevigo/shugosha/pkg/model"
)

// Server represents the API server.
type Server struct {
	configManager model.ConfigManager
	router        *chi.Mux
}

// NewServer creates a new API server.
func NewServer(cm model.ConfigManager) *Server {
	s := &Server{
		configManager: cm,
		router:        chi.NewRouter(),
	}

	s.routes()

	return s
}

// routes defines all the routes for the API server.
func (s *Server) routes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	configHandler := config.NewConfigHandler(s.configManager)

	s.router.Get("/api/config", configHandler.ReadConfigHandler)
	s.router.Post("/api/config", configHandler.UpdateConfigHandler)
}

// Start starts the API server on the specified port.
func (s *Server) Start(ctx context.Context, port string) error {
	return http.ListenAndServe(port, s.router)
}
