package mockaso

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
)

type Server struct {
	server *httptest.Server
	stubs  []*stub
	logger Logger
	mutex  sync.RWMutex
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

func (s *Server) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stubs = nil

	if s.server == nil {
		return
	}

	s.logger.Logf("server cleared at %s", s.server.URL)
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

func (s *Server) Client() *http.Client {
	if s.server == nil {
		return nil
	}

	client := s.server.Client()
	client.Transport = newTransportWithBaseURL(client.Transport, s.URL())

	return client
}

func (s *Server) Logger() Logger {
	return s.logger
}

func (s *Server) Stub(method string, url URLMatcher) Stub {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	st := &stub{response: newStubResponse(), matchers: defaultMatchers(method, url)}
	s.stubs = append(s.stubs, st)

	return st
}

func (s *Server) newTestServer() *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		for _, st := range s.stubs {
			if st.match(r) {
				st.write(w)
				return
			}
		}

		// http request does not match with any stub
		s.logger.Logf("no stub matched for %s %s", r.Method, r.URL.String())
		writeNoMatch(w, r)
	})

	return httptest.NewServer(h)
}

func NewServer(opts ...ServerOption) *Server {
	server := &Server{
		logger: &noLogger{},
		stubs:  make([]*stub, 0),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func MustStartNewServer(opts ...ServerOption) *Server {
	server := NewServer(opts...)
	server.MustStart()

	return server
}

const demonCode = 666

func writeNoMatch(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(demonCode)
	_, _ = fmt.Fprintf(w, "no stubs for %s %s", r.Method, r.URL)
}

type ServerOption func(*Server)

// WithLogger sets a Logger. Intended for use with testing.T
func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithSlogLogger sets a Logger from slog.Logger.
// level is the slog.LogLevel that will be used.
func WithSlogLogger(logger *slog.Logger, level slog.Level) ServerOption {
	return func(s *Server) {
		s.logger = NewSlogLogger(logger, level)
	}
}

// WithLogLogger sets a Logger from log.Logger.
func WithLogLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.logger = NewLogLogger(logger)
	}
}
