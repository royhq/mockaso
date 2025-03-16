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
