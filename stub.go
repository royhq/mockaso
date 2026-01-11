package mockaso

import (
	"net/http"
	"time"
)

// Stub represents a configured HTTP endpoint mock with request matching rules
// and response configuration. It provides a fluent API for chaining matchers
// and response rules.
//
// The typical usage pattern is:
//  1. Create a stub with Server.Stub(method, urlMatcher)
//  2. Optionally add matchers with Match(...)
//  3. Configure the response with Respond(...)
//
// Example:
//
//	server.Stub("POST", mockaso.Path("/users")).
//		Match(
//			mockaso.MatchHeader("Content-Type", "application/json"),
//			mockaso.MatchJSONBody(map[string]string{"name": "Alice"}),
//		).
//		Respond(
//			mockaso.WithStatusCode(201),
//			mockaso.WithJSON(map[string]any{"id": 123, "name": "Alice"}),
//		)
type Stub interface {
	StubResponder
	Match(...StubMatcherRule) StubResponder
}

// StubResponder provides the final step in the fluent API for configuring
// a stub's response. After all matchers are defined, call Respond to set
// the HTTP response that will be returned when the stub matches.
//
// Example:
//
//	stub.Respond(
//		mockaso.WithStatusCode(200),
//		mockaso.WithJSON(responseData),
//		mockaso.WithHeader("X-Custom", "value"),
//	)
type StubResponder interface {
	Respond(...StubResponseRule)
}

type stub struct {
	matchers      []requestMatcherFunc
	response      *stubResponse
	patternParams map[string]string
}

func (s *stub) Match(rules ...StubMatcherRule) StubResponder {
	for _, rule := range rules {
		s.matchers = append(s.matchers, rule())
	}

	return s
}

func (s *stub) Respond(rules ...StubResponseRule) {
	for _, rule := range rules {
		rule(s.response)
	}
}

func (s *stub) match(r *http.Request) bool {
	for _, match := range s.matchers {
		if !match(s, r) {
			return false
		}
	}

	return true
}

func (s *stub) write(w http.ResponseWriter, r *http.Request) {
	for k, v := range s.response.headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(s.response.statusCode)

	if s.response.delay > 0 {
		select {
		case <-time.After(s.response.delay):
			// Delay completed normally
		case <-r.Context().Done():
			// Request was cancelled, abort
			return
		}
	}

	_, _ = w.Write(s.response.body)
}

type stubResponse struct {
	statusCode int
	body       []byte
	headers    map[string]string
	delay      time.Duration
}

func (r *stubResponse) setHeader(key, value string) {
	r.headers[key] = value
}

func (r *stubResponse) setHeaders(headers map[string]string) {
	for k, v := range headers {
		r.headers[k] = v
	}
}

func (r *stubResponse) setJSON(content []byte) {
	r.headers["Content-Type"] = "application/json"
	r.body = content
}

func newStubResponse() *stubResponse {
	return &stubResponse{
		statusCode: http.StatusOK,
		headers:    make(map[string]string),
	}
}
