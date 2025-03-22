package mockaso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
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

// URLRegex will match http request when the regex pattern specified match to the request URL.
func URLRegex(pattern string) URLMatcher {
	regex := regexp.MustCompile(pattern)
	return func(url *url.URL) bool { return regex.MatchString(url.String()) }
}

// PathRegex will match http request when the regex pattern specified match to the request URL path part.
func PathRegex(pattern string) URLMatcher {
	regex := regexp.MustCompile(pattern)
	return func(url *url.URL) bool { return regex.MatchString(url.Path) }
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

// MatchQuery sets a rule to match the http request with the given query string value.
func MatchQuery(key, value string) StubMatcherRule {
	matcher := RequestMatcherFunc(func(r *http.Request) bool {
		return r.URL.Query().Get(key) == value
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

// MatchRawJSONBody sets a rule to match the http request with the given raw JSON body.
func MatchRawJSONBody[T string | []byte | json.RawMessage](raw T) StubMatcherRule {
	return MatchJSONBody(json.RawMessage(raw))
}

// MatchJSONBody sets a rule to match the http request with the given JSON body.
// The specified body will be marshaled and compared with the real body.
func MatchJSONBody(body any) StubMatcherRule {
	data, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Errorf("MatchJSONBody err: marshal body failed: %w", err))
	}

	matcher := RequestMatcherFunc(func(r *http.Request) bool {
		reqBody := mustReadBody(r)

		equals, equalsErr := equalJSON(reqBody, data)
		if equalsErr != nil {
			panic(fmt.Errorf("MatchJSONBody err: equals failed: %w", equalsErr))
		}

		return equals
	})

	return MatchRequest(matcher)
}

type BodyMatcherMapFunc func(map[string]any) bool

// MatchBodyMapFunc sets a rule to match the http request with the given matcher based on the body as a map.
// The matcher is a func that receives the body parameters as a map. If the body is empty the map will be empty.
func MatchBodyMapFunc(bodyMatcher BodyMatcherMapFunc) StubMatcherRule {
	matcher := RequestMatcherFunc(func(r *http.Request) bool {
		reqBody := mustReadBody(r)

		if len(reqBody) == 0 { // empty body
			return bodyMatcher(make(map[string]any)) // empty map
		}

		var bodyMap map[string]any

		if err := json.Unmarshal(reqBody, &bodyMap); err != nil {
			panic(fmt.Errorf("MatchBodyMapFunc err: unmarshal body failed: %w", err))
		}

		return bodyMatcher(bodyMap)
	})

	return MatchRequest(matcher)
}

type BodyMatcherStringFunc func(string) bool

// MatchBodyStringFunc sets a rule to match the http request with the given matcher based on the body as string.
// The matcher is a func that receives the body as plain text.
func MatchBodyStringFunc(bodyMatcher BodyMatcherStringFunc) StubMatcherRule {
	matcher := RequestMatcherFunc(func(r *http.Request) bool {
		reqBody := mustReadBody(r)
		return bodyMatcher(string(reqBody))
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

func equalJSON(v1, v2 []byte) (bool, error) {
	var json1, json2 any

	if len(v1) > 0 {
		if err := json.Unmarshal(v1, &json1); err != nil {
			return false, fmt.Errorf("failed to unmarshal JSON v1: %w", err)
		}
	}

	if len(v2) > 0 {
		if err := json.Unmarshal(v2, &json2); err != nil {
			return false, fmt.Errorf("failed to unmarshal JSON v2: %w", err)
		}
	}

	return reflect.DeepEqual(json1, json2), nil
}
