package http

import (
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents an http server
type Server struct {
	address            string
	listener           net.Listener
	router             chi.Router
	JsonEateryProvider AsyncProvider
	XmlGrillProvider   AsyncProvider
}

// NewServer returns a new sever instance
func NewServer(opts ...func(*Server)) *Server {
	s := &Server{
		address: ":8080",
	}

	s.router = chi.NewRouter()

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithAddress sets the listening address and port for the server
// Format is https://my-address.com:80
// defaults to 127.0.0.1 on random port
func WithAddress(address string) func(*Server) {
	return func(s *Server) {
		s.address = address
	}
}

// Open opens the server and listens at the specifed address
func (s *Server) Open() error {
	// create the routes for the server
	s.routes()

	// Open socket.
	ln, err := net.Listen("tcp", s.address)

	if err != nil {
		return err
	}

	s.listener = ln

	// Start HTTP server. Note this is non-blocking so
	// we must block in the calling code
	go func() { http.Serve(s.listener, s.router) }()

	log.Println("Server started listening on port " + s.address[1:] + "....")

	return nil
}

// Close closes the socket.
func (s *Server) Close() error {
	if s.listener != nil {
		defer s.listener.Close()
	}

	return nil
}

// routes maps all route handlers to their respoective paths
func (s *Server) routes() {
	// basic middlewares for our server
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Route("/providers/locations/", func(r chi.Router) {
		r.Get(
			"/{id}",
			s.handleGetProviderInfo(),
		)
		r.Get(
			"/{id}/menu",
			s.handleGetFullMenu(),
		)
	})
}
