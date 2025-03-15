package mockaso

import (
	"net/http"
	"net/http/httptest"
)

type Server struct {
	server *httptest.Server
	logger Logger
}

func (s *Server) Start() error {
	if s.server == nil {
		s.server = s.newTestServer()
	}

	s.logger.Logf("server started at %s", s.server.URL)

	return nil
}

func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	s.server.Close()

	s.logger.Logf("server stopped at %s", s.server.URL)

	return nil
}

func (s *Server) MustStart() {
	if err := s.Start(); err != nil {
		panic(err)
	}
}

func (s *Server) MustShutdown() {
	if err := s.Shutdown(); err != nil {
		panic(err)
	}
}

func (s *Server) URL() string {
	if s.server == nil {
		return ""
	}

	return s.server.URL
}

func (s *Server) TestServer() *httptest.Server {
	return s.server
}

func (s *Server) Clear() {
	if s.server == nil {
		return
	}

	s.server.Close()
	s.logger.Logf("server cleared at %s", s.server.URL)
}

func (s *Server) newTestServer() *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: complete
	})

	return httptest.NewServer(h)
}

func NewServer(opts ...ServerOption) *Server {
	server := &Server{
		logger: &noLogger{},
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

type ServerOption func(*Server)

// WithLogger sets a Logger. Intended for use with testing.T
func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}
