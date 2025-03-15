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
