package mockaso

// Logger abstraction intended for use with testing.T.
type Logger interface {
	Log(...any)
	Logf(string, ...any)
}

// noLogger is a Logger that does not log anything.
type noLogger struct{}

func (n noLogger) Log(...any)          {}
func (n noLogger) Logf(string, ...any) {}
