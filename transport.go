package mockaso

import (
	"fmt"
	"net/http"
	"net/url"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func newTransportWithBaseURL(baseTransport http.RoundTripper, baseURL string) http.RoundTripper {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse base URL: %s", err))
	}

	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		copyRequest := *r

		if !copyRequest.URL.IsAbs() { // only modify relative URL
			copyRequest.URL = parsedBaseURL.ResolveReference(copyRequest.URL)
			copyRequest.Host = copyRequest.URL.Host
		}

		return baseTransport.RoundTrip(&copyRequest)
	})
}
