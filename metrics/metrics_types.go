package metrics

type DecisionMetrics struct {
	VetoedSessions Counter
}

var EmptyDecisionMetrics DecisionMetrics = DecisionMetrics{
	VetoedSessions: &EmptyCounter{},
}

type SessionMetrics struct {
	Invocations     Counter
	DirectSessions  Counter
	NextSessions    Counter
	DurationGauge   Gauge
	DecisionMetrics DecisionMetrics
}

var EmptySessionMetrics SessionMetrics = SessionMetrics{
	Invocations:     &EmptyCounter{},
	DirectSessions:  &EmptyCounter{},
	NextSessions:    &EmptyCounter{},
	DurationGauge:   &EmptyGauge{},
	DecisionMetrics: EmptyDecisionMetrics,
}

type ServerUpdateMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
}

var EmptyServerUpdateMetrics ServerUpdateMetrics = ServerUpdateMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
}
