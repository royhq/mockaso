package mockaso

import (
	"net/http"
)

type StubMatcherRule func() requestMatcherFunc

type Stub interface {
	StubResponder
	Match(...StubMatcherRule) StubResponder
}

type StubResponder interface {
	Respond(...StubResponseRule)
}

type stub struct {
	matchers []requestMatcherFunc
	response *stubResponse
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

func (s *stub) write(w http.ResponseWriter) {
	for k, v := range s.response.headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(s.response.statusCode)
	_, _ = w.Write(s.response.body)
}

type stubResponse struct {
	statusCode int
	body       []byte
	headers    map[string]string
}

func newStubResponse() *stubResponse {
	return &stubResponse{
		statusCode: http.StatusOK,
		headers:    make(map[string]string),
	}
}
