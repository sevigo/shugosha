package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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
	s.router.Get("/config", s.readConfigHandler)
	s.router.Post("/config", s.updateConfigHandler)
}

// Start starts the API server on the specified port.
func (s *Server) Start(port string) error {
	return http.ListenAndServe(port, s.router)
}
