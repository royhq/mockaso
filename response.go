package mockaso

import (
	"encoding/json"
	"fmt"
	"io"
)

type StubResponseRule func(*stubResponse)

// WithStatusCode sets the response status code.
func WithStatusCode(statusCode int) StubResponseRule {
	return func(r *stubResponse) {
		r.statusCode = statusCode
	}
}

// WithBody sets the response body.
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
func WithRawJSON[T string | []byte | json.RawMessage](raw T) StubResponseRule {
	return func(r *stubResponse) {
		r.setJSON([]byte(raw))
	}
}

// WithHeader sets a response header.
// If the key already exists it will be overwritten.
func WithHeader(key, value string) StubResponseRule {
	return func(r *stubResponse) {
		r.setHeader(key, value)
	}
}

// WithHeaders sets a set of response headers.
// These headers will be added to the already specified headers.
// If any key already exists it will be overwritten.
func WithHeaders(headers map[string]string) StubResponseRule {
	return func(r *stubResponse) {
		r.setHeaders(headers)
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
