# Mockaso

A lightweight Go library for creating HTTP mock servers in tests. Built around Go's `httptest.Server`, Mockaso provides a fluent API for stubbing HTTP endpoints with powerful matching rules and flexible responses.

## Features

- **Fluent API**: Chain methods for readable and maintainable test code
- **Flexible Matching**: Match requests by method, URL, headers, body, query parameters, and more
- **Pattern Matching**: Extract URL parameters with patterns like `/users/{id}` and use them in matchers
- **Multiple URL Matching Strategies**: Exact match, regex, and pattern-based matching
- **JSON Support**: Built-in JSON response handling with automatic Content-Type headers
- **Response Delays**: Simulate network latency with configurable delays
- **Relative URLs**: Use relative paths in tests without worrying about server addresses
- **Thread-Safe**: Safely add and clear stubs from concurrent tests
- **Order-Based Matching**: First matching stub wins, allowing specific-to-general stub ordering
- **Optional Logging**: Integrate with `testing.T`, `slog`, or `log` for debugging

## Installation

```bash
go get github.com/royhq/mockaso
```

## Quick Start

```go
package mypackage_test

import (
    "testing"
    "github.com/roycodev/mockaso"
)

func TestAPIClient(t *testing.T) {
    // Create and start mock server
    server := mockaso.NewServer(mockaso.WithLogger(t))
    server.MustStart()
    defer server.MustShutdown()

    // Stub an endpoint
    server.Stub("GET", mockaso.Path("/api/users")).
        Respond(
            mockaso.WithJSON([]map[string]any{
                {"id": 1, "name": "Alice"},
                {"id": 2, "name": "Bob"},
            }),
            mockaso.WithStatusCode(200),
        )

    // Use the server in your tests
    client := server.Client()
    resp, err := client.Get("/api/users")

    // ... your assertions
}
```

## Usage Examples

### Basic Stub

```go
server.Stub("POST", mockaso.Path("/api/users")).
    Respond(
        mockaso.WithStatusCode(201),
        mockaso.WithBody([]byte(`{"id": 1, "name": "Alice"}`)),
    )
```

### Match Request Headers

```go
server.Stub("GET", mockaso.Path("/api/users")).
    Match(mockaso.MatchHeader("Authorization", "Bearer token123")).
    Respond(mockaso.WithJSON(users))
```

### Match Request Body

```go
server.Stub("POST", mockaso.Path("/api/users")).
    Match(mockaso.MatchBody([]byte(`{"name":"Alice"}`))).
    Respond(mockaso.WithStatusCode(201))
```

### URL Pattern with Parameters

```go
// Match /users/123, /users/456, etc.
server.Stub("GET", mockaso.PathPattern("/users/{id}")).
    Match(mockaso.MatchParam("id", "123")).
    Respond(mockaso.WithJSON(map[string]any{"id": 123, "name": "Alice"}))
```

### Regex Matching

```go
// Match any UUID in the path
server.Stub("GET", mockaso.PathRegex(`/users/[0-9a-f-]+`)).
    Respond(mockaso.WithJSON(user))
```

### Query Parameter Matching

```go
server.Stub("GET", mockaso.Path("/api/users")).
    Match(mockaso.MatchQuery("role", "admin")).
    Respond(mockaso.WithJSON(adminUsers))
```

### Response Delay

```go
// Simulate slow network
server.Stub("GET", mockaso.Path("/api/slow")).
    Respond(
        mockaso.WithJSON(data),
        mockaso.WithDelay(2 * time.Second),
    )
```

### Multiple Stubs (Order Matters)

```go
// Specific case first
server.Stub("GET", mockaso.Path("/api/users/special")).
    Respond(mockaso.WithJSON(specialUser))

// General case second
server.Stub("GET", mockaso.PathPattern("/api/users/{id}")).
    Respond(mockaso.WithJSON(regularUser))
```

### Clearing Stubs

```go
func TestMultipleCases(t *testing.T) {
    server := mockaso.NewServer()
    server.MustStart()
    defer server.MustShutdown()

    t.Run("case1", func(t *testing.T) {
        server.Stub("GET", mockaso.Path("/api/data")).
            Respond(mockaso.WithJSON(data1))
        // ... test logic
        server.Clear() // Clean up for next test
    })

    t.Run("case2", func(t *testing.T) {
        server.Stub("GET", mockaso.Path("/api/data")).
            Respond(mockaso.WithJSON(data2))
        // ... test logic
    })
}
```

## How It Works

1. **Create a server**: Wraps Go's `httptest.Server` with stub management
2. **Add stubs**: Define expected requests and their responses
3. **First match wins**: Stubs are evaluated in order; the first matching stub serves the response
4. **No match**: Returns HTTP 666 (demon code) when no stub matches
5. **Use in tests**: Access via `server.Client()` for relative URLs or `server.URL()` for full address

## License

MIT License - see [LICENSE](LICENSE) file for details
