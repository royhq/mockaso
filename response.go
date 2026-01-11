package mockaso

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// StubResponseRule is a function that configures a stub's HTTP response.
// Multiple rules can be applied to a single response, and they are processed
// in the order they are provided to Respond.
//
// Example:
//
//	stub.Respond(
//		mockaso.WithStatusCode(201),
//		mockaso.WithJSON(createdResource),
//		mockaso.WithHeader("Location", "/api/resources/123"),
//	)
type StubResponseRule func(*stubResponse)

// WithStatusCode sets the response status code.
// If not specified, the default status code is 200 (OK).
//
// Example:
//
//	stub.Respond(mockaso.WithStatusCode(404))
func WithStatusCode(statusCode int) StubResponseRule {
	return func(r *stubResponse) {
		r.statusCode = statusCode
	}
}

// WithBody sets the response body.
// Accepts []byte, string, json.RawMessage, or io.Reader.
// For other types, the value is converted using fmt.Sprintf("%v", value).
//
// Example:
//
//	stub.Respond(mockaso.WithBody("Hello, World!"))
//	stub.Respond(mockaso.WithBody([]byte{0x89, 0x50, 0x4E, 0x47}))
func WithBody(body any) StubResponseRule {
	data, err := anyBodyToBytes(body)
	if err != nil {
		panic(fmt.Errorf("WithBody err: failed to read body: %w", err))
	}

	return func(r *stubResponse) {
		r.body = data
	}
}

// WithRawJSON sets the response content with the given JSON.
// The response will include the Content-Type:application/json header.
// The provided JSON is validated before being set.
//
// Example:
//
//	stub.Respond(mockaso.WithRawJSON(`{"status":"ok","count":42}`))
func WithRawJSON[T string | []byte | json.RawMessage](raw T) StubResponseRule {
	data := []byte(raw)

	if !json.Valid(data) {
		panic(fmt.Errorf("json is not valid: %s", data))
	}

	return func(r *stubResponse) {
		r.setJSON(data)
	}
}

// WithJSON sets the response content with the marshal output of the given body.
// The response will include the Content-Type:application/json header.
// Panics if the body cannot be marshaled to JSON.
//
// Example:
//
//	user := User{ID: 1, Name: "Alice"}
//	stub.Respond(mockaso.WithJSON(user))
//
//	// Or with inline data
//	stub.Respond(mockaso.WithJSON(map[string]any{
//		"users": []string{"Alice", "Bob"},
//		"total": 2,
//	}))
func WithJSON(body any) StubResponseRule {
	return func(r *stubResponse) {
		data, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Errorf("WithJSON err: body marshal failed: %w", err))
		}

		r.setJSON(data)
	}
}

// WithHeader sets a response header.
// If the key already exists it will be overwritten.
//
// Example:
//
//	stub.Respond(
//		mockaso.WithHeader("Content-Type", "text/plain"),
//		mockaso.WithHeader("X-Request-ID", "abc-123"),
//	)
func WithHeader(key, value string) StubResponseRule {
	return func(r *stubResponse) {
		r.setHeader(key, value)
	}
}

// WithHeaders sets a set of response headers.
// These headers will be added to the already specified headers.
// If any key already exists it will be overwritten.
//
// Example:
//
//	headers := map[string]string{
//		"X-RateLimit-Limit": "100",
//		"X-RateLimit-Remaining": "99",
//	}
//	stub.Respond(mockaso.WithHeaders(headers))
func WithHeaders(headers map[string]string) StubResponseRule {
	return func(r *stubResponse) {
		r.setHeaders(headers)
	}
}

// WithDelay sets a delay time to the response.
// The delay is applied before writing the response body.
// If the request context is cancelled during the delay, the response is aborted.
//
// Example:
//
//	stub.Respond(
//		mockaso.WithJSON(data),
//		mockaso.WithDelay(500 * time.Millisecond), // Simulate slow API
//	)
func WithDelay(d time.Duration) StubResponseRule {
	if d < 0 {
		panic(fmt.Errorf("WithDelay: delay must be non-negative, got %s", d))
	}

	return func(r *stubResponse) {
		r.delay = d
	}
}

func anyBodyToBytes(body any) ([]byte, error) {
	switch v := body.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case json.RawMessage:
		return v, nil
	case io.Reader:
		return io.ReadAll(v)
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}
