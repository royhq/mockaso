// Package mockaso provides a fluent API for creating HTTP mock servers in tests.
// It wraps Go's httptest.Server with a powerful stubbing system that supports
// request matching, pattern-based URLs, and flexible response configuration.
//
// # Basic Usage
//
// Create a server, add stubs, and use it in your tests:
//
//	server := mockaso.NewServer(mockaso.WithLogger(t))
//	server.MustStart()
//	defer server.MustShutdown()
//
//	server.Stub("GET", mockaso.Path("/users")).
//		Respond(mockaso.WithJSON([]User{{ID: 1, Name: "Alice"}}))
//
//	client := server.Client()
//	resp, _ := client.Get("/users") // Uses relative URL
//
// # Stub Matching
//
// Stubs are evaluated in the order they are added. The first matching stub
// handles the request. If no stub matches, the server returns status 666.
//
// URL matching supports exact paths, regex patterns, and parameter extraction:
//
//	mockaso.Path("/users")                    // Exact path
//	mockaso.PathRegex(`/users/\d+`)          // Regex
//	mockaso.PathPattern("/users/{id}")       // Pattern with params
//
// Additional matchers refine which requests a stub handles:
//
//	server.Stub("POST", mockaso.Path("/users")).
//		Match(
//			mockaso.MatchHeader("Content-Type", "application/json"),
//			mockaso.MatchJSONBody(expectedData),
//		).
//		Respond(mockaso.WithStatusCode(201))
//
// # Response Configuration
//
// Responses support JSON, custom headers, status codes, and delays:
//
//	stub.Respond(
//		mockaso.WithStatusCode(200),
//		mockaso.WithJSON(data),
//		mockaso.WithHeader("X-Custom", "value"),
//		mockaso.WithDelay(100 * time.Millisecond),
//	)
package mockaso

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
)

// Server is a wrapper around httptest.Server that provides HTTP endpoint mocking
// capabilities with a fluent API for stubbing responses.
//
// When no stub matches an incoming request, the server responds with status code 666
// (demon code) and a descriptive error message.
//
// Example:
//
//	server := mockaso.NewServer(mockaso.WithLogger(t))
//	server.MustStart()
//	defer server.MustShutdown()
//
//	server.Stub("GET", mockaso.Path("/users")).
//		Respond(
//			mockaso.WithJSON([]User{{ID: 1, Name: "Alice"}}),
//			mockaso.WithStatusCode(200),
//		)
//
//	client := server.Client()
//	resp, _ := client.Get("/users")
type Server struct {
	server *httptest.Server
	stubs  []*stub
	logger Logger
	mutex  sync.RWMutex
}

// Start initializes and starts the HTTP test server.
// It is safe to call Start multiple times - subsequent calls are no-ops.
// Returns nil on success.
func (s *Server) Start() error {
	if s.server == nil {
		s.server = s.newTestServer()
	}

	s.logger.Logf("server started at %s", s.server.URL)

	return nil
}

// Shutdown stops the HTTP test server and cleans up resources.
// It is safe to call Shutdown on an already stopped server.
// Returns nil on success.
func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	s.server.Close()

	s.logger.Logf("server stopped at %s", s.server.URL)

	return nil
}

// MustStart starts the server and panics if an error occurs.
// This is a convenience method for tests where startup failure should be fatal.
func (s *Server) MustStart() {
	if err := s.Start(); err != nil {
		panic(err)
	}
}

// MustShutdown stops the server and panics if an error occurs.
// This is a convenience method for tests, typically used with defer.
func (s *Server) MustShutdown() {
	if err := s.Shutdown(); err != nil {
		panic(err)
	}
}

// Clear removes all registered stubs from the server.
// The server continues running but will return 666 for all requests
// until new stubs are added.
// This method is thread-safe.
func (s *Server) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stubs = nil

	if s.server == nil {
		return
	}

	s.logger.Logf("server cleared at %s", s.server.URL)
}

// URL returns the base URL of the running server.
// Returns an empty string if the server has not been started.
//
// Example:
//
//	server.MustStart()
//	fmt.Println(server.URL()) // http://127.0.0.1:12345
func (s *Server) URL() string {
	if s.server == nil {
		return ""
	}

	return s.server.URL
}

// TestServer returns the underlying httptest.Server.
// This is useful for advanced scenarios that require direct access
// to the test server.
func (s *Server) TestServer() *httptest.Server {
	return s.server
}

// Client returns an HTTP client configured to work with this mock server.
// The returned client uses a custom transport that automatically resolves
// relative URLs (like "/api/users") to the server's base URL.
//
// Example:
//
//	client := server.Client()
//	resp, err := client.Get("/api/users") // Automatically uses server.URL()
func (s *Server) Client() *http.Client {
	if s.server == nil {
		return nil
	}

	client := s.server.Client()
	client.Transport = newTransportWithBaseURL(client.Transport, s.URL())

	return client
}

// Logger returns the configured logger for this server.
func (s *Server) Logger() Logger {
	return s.logger
}

// Stub creates and registers a new stub for the given HTTP method and URL matcher.
// Stubs are evaluated in the order they are added - the first matching stub
// handles the request.
//
// The method parameter should be an HTTP method like "GET", "POST", "PUT", etc.
// The url parameter is a URLMatcher that determines which requests this stub handles.
//
// Example:
//
//	server.Stub("GET", mockaso.Path("/users")).
//		Match(mockaso.MatchHeader("Accept", "application/json")).
//		Respond(mockaso.WithJSON(users), mockaso.WithStatusCode(200))
func (s *Server) Stub(method string, url URLMatcher) Stub {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	st := &stub{
		response:      newStubResponse(),
		matchers:      defaultMatchers(method, url),
		patternParams: make(map[string]string),
	}

	s.stubs = append(s.stubs, st)

	return st
}

func (s *Server) newTestServer() *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		for _, st := range s.stubs {
			if st.match(r) {
				st.write(w, r)
				return
			}
		}

		// http request does not match with any stub
		s.logger.Logf("no stub matched for %s %s", r.Method, r.URL.String())
		writeNoMatch(w, r)
	})

	return httptest.NewServer(h)
}

// NewServer creates a new mock server with the given options.
// The server is not started automatically - call Start or MustStart to begin
// accepting requests.
//
// By default, the server uses a silent logger that produces no output.
// Use WithLogger, WithSlogLogger, or WithLogLogger to enable logging.
//
// Example:
//
//	server := mockaso.NewServer(mockaso.WithLogger(t))
//	server.MustStart()
//	defer server.MustShutdown()
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

// MustStartNewServer creates a new mock server and starts it immediately.
// This is a convenience function equivalent to calling NewServer followed by MustStart.
// It panics if the server fails to start.
//
// Example:
//
//	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
//	defer server.MustShutdown()
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

// ServerOption is a function that configures a Server during initialization.
// Options are applied in the order they are provided to NewServer.
type ServerOption func(*Server)

// WithLogger sets a Logger. Intended for use with testing.T.
//
// Example:
//
//	func TestAPI(t *testing.T) {
//		server := mockaso.NewServer(mockaso.WithLogger(t))
//		// Server will log to test output
//	}
func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithSlogLogger sets a Logger from slog.Logger.
// level is the slog.LogLevel that will be used.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	server := mockaso.NewServer(mockaso.WithSlogLogger(logger, slog.LevelInfo))
func WithSlogLogger(logger *slog.Logger, level slog.Level) ServerOption {
	return func(s *Server) {
		s.logger = NewSlogLogger(logger, level)
	}
}

// WithLogLogger sets a Logger from log.Logger.
//
// Example:
//
//	logger := log.New(os.Stdout, "[MOCKASO] ", log.LstdFlags)
//	server := mockaso.NewServer(mockaso.WithLogLogger(logger))
func WithLogLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.logger = NewLogLogger(logger)
	}
}
