package mockaso

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type requestMatcherFunc func(*stub, *http.Request) bool

type URLMatcher func(*url.URL) bool

// URL will match http request when the value specified is equals to the full request URL.
func URL(u string) URLMatcher {
	return func(url *url.URL) bool {
		return u == url.String()
	}
}

// Path will match http request when the value specified is equals to the request URL path part.
func Path(path string) URLMatcher {
	return func(url *url.URL) bool {
		return url.Path == strings.TrimSuffix(path, "/")
	}
}

func defaultMatchers(method string, url URLMatcher) []requestMatcherFunc {
	return []requestMatcherFunc{
		methodMatcher(method),
		urlMatcher(url),
	}
}

func methodMatcher(method string) requestMatcherFunc {
	return func(_ *stub, r *http.Request) bool {
		return r.Method == method
	}
}

func urlMatcher(matcher URLMatcher) requestMatcherFunc {
	return func(_ *stub, r *http.Request) bool {
		return matcher(r.URL)
	}
}

type StubMatcherRule func() requestMatcherFunc

type RequestMatcherFunc func(*http.Request) bool

// MatchHeader sets a rule to match the http request with the given header value.
func MatchHeader(key, value string) StubMatcherRule {
	matcher := RequestMatcherFunc(func(r *http.Request) bool {
		return r.Header.Get(key) == value
	})

	return MatchRequest(matcher)
}

// MatchNoBody sets a rule to match the http request with empty body.
func MatchNoBody() StubMatcherRule {
	matcher := RequestMatcherFunc(func(r *http.Request) bool {
		realReqBody := mustReadBody(r)
		return len(realReqBody) == 0
	})

	return MatchRequest(matcher)
}

// MatchRequest sets a rule to match the http request given a custom matcher.
func MatchRequest(requestMatcher RequestMatcherFunc) StubMatcherRule {
	matcher := requestMatcherFunc(func(_ *stub, r *http.Request) bool {
		return requestMatcher(r)
	})

	return func() requestMatcherFunc { return matcher }
}

func mustReadBody(r *http.Request) []byte {
	buff := new(bytes.Buffer)
	tee := io.TeeReader(r.Body, buff)

	data, err := io.ReadAll(tee)
	if err != nil {
		panic(fmt.Errorf("read request body failed: %w", err))
	}

	r.Body = io.NopCloser(buff)

	return data
}
