package mockaso

import (
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

// MatchRequest sets a rule to match the http request given a custom matcher.
func MatchRequest(requestMatcher RequestMatcherFunc) StubMatcherRule {
	matcher := requestMatcherFunc(func(_ *stub, r *http.Request) bool {
		return requestMatcher(r)
	})

	return func() requestMatcherFunc { return matcher }
}
