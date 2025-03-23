package mockaso

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

type requestMatcherFunc func(*stub, *http.Request) bool

type URLMatcher func(*url.URL, *stub) bool

// URL will match http request when the value specified is equals to the full request URL.
func URL(u string) URLMatcher {
	return func(url *url.URL, _ *stub) bool {
		return u == url.String()
	}
}

// Path will match http request when the value specified is equals to the request URL path part.
func Path(path string) URLMatcher {
	ensureHasNotQueryStringParams(path)

	return func(url *url.URL, _ *stub) bool {
		return url.Path == strings.TrimSuffix(path, "/")
	}
}

// URLRegex will match http request when the regex pattern specified match to the request URL.
func URLRegex(pattern string) URLMatcher {
	regex := regexp.MustCompile(pattern)
	return func(url *url.URL, _ *stub) bool { return regex.MatchString(url.String()) }
}

// PathRegex will match http request when the regex pattern specified match to the request URL path part.
func PathRegex(pattern string) URLMatcher {
	regex := regexp.MustCompile(pattern)
	return func(url *url.URL, _ *stub) bool { return regex.MatchString(url.Path) }
}

// URLPattern will match http request when the given URL pattern match to the request URL.
// Can specify path params with {param_name} notation and then use it in matcher.
// Can use parameters in query string.
//
// Example:
//
//	URLPattern("/api/users/{user_id}")
//	URLPattern("/api/users/{user_id}?attrs={attrs}")
func URLPattern(pattern string) URLMatcher {
	source := func(u *url.URL) string { return u.String() } // use complete url as source
	return patternMatcher(source, pattern)
}

// PathPattern will match http request when the given URL pattern match to the request URL path part.
// Can specify path params with {param_name} notation and then use it in matcher.
// Can't use parameters in query string, only path will be evaluated.
//
// Example:
//
//	PathPattern("/api/users/{user_id}")
func PathPattern(pattern string) URLMatcher {
	ensureHasNotQueryStringParams(pattern)
	source := func(u *url.URL) string { return u.Path } // use url path as source

	return patternMatcher(source, pattern)
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
	return func(st *stub, r *http.Request) bool {
		return matcher(r.URL, st)
	}
}

func patternMatcher(source func(*url.URL) string, pattern string) URLMatcher {
	expr, paramKeys := convertPatternToRegex(pattern)
	regex := regexp.MustCompile(expr)

	return func(url *url.URL, s *stub) bool {
		match := regex.FindStringSubmatch(source(url))
		if match == nil {
			return false
		}

		params := make(map[string]string)
		for i, paramKey := range paramKeys {
			params[paramKey] = match[i+1]
		}

		s.patternParams = params

		return true
	}
}

func convertPatternToRegex(urlPattern string) (string, []string) {
	urlPattern = escapeURLPattern(urlPattern)

	var paramNames []string

	re := regexp.MustCompile(`\{(\w+)}`) // to identify parameters like {param_name} within pattern

	urlPattern = re.ReplaceAllStringFunc(urlPattern, func(match string) string {
		paramName := re.FindStringSubmatch(match)[1]
		paramNames = append(paramNames, paramName)

		return fmt.Sprintf(`(?P<%s>[^/?&]+)`, paramName)
	})

	return "^" + urlPattern + "$", paramNames
}

func escapeURLPattern(urlPattern string) string {
	escaped := strings.ReplaceAll(urlPattern, "?", `\?`)
	escaped = strings.ReplaceAll(escaped, "&", `\&`)
	escaped = strings.ReplaceAll(escaped, "=", `\=`)

	return escaped
}

func ensureHasNotQueryStringParams(pattern string) {
	parsed, err := url.Parse(pattern)
	if err != nil {
		panic(fmt.Errorf("not valid url"))
	}

	if len(parsed.Query()) > 0 {
		panic(errors.New("pattern must not contain any query string parameters"))
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

// MatchParam sets a rule to match the http request with the given path param value.
// This needs that the URL must be specified with URLPattern.
func MatchParam(key, value string) StubMatcherRule {
	matcher := requestMatcherFunc(func(st *stub, r *http.Request) bool {
		return st.patternParams[key] == value
	})

	return func() requestMatcherFunc { return matcher }
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
