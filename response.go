package mockaso

type StubResponseRule func(*stubResponse)

func WithStatusCode(statusCode int) StubResponseRule {
	return func(r *stubResponse) {
		r.statusCode = statusCode
	}
}
