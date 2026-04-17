package decorator

// MetricsClient records command/query execution (optional; use NoopMetrics when unset).
type MetricsClient interface {
	IncCommand(name string)
	IncQuery(name string)
}

// NoopMetrics is a no-op MetricsClient.
type NoopMetrics struct{}

func (NoopMetrics) IncCommand(string) {}
func (NoopMetrics) IncQuery(string)   {}
