package metrics

type DecisionMetrics struct {
	NoChange            Counter
	ForceDirect         Counter
	ForceNext           Counter
	NoNextRoute         Counter
	ABTestDirect        Counter
	RTTReduction        Counter
	PacketLossMultipath Counter
	JitterMultipath     Counter
	VetoRTT             Counter
	RTTMultipath        Counter
	VetoPacketLoss      Counter
	FallbackToDirect    Counter
	VetoYOLO            Counter
	VetoNoRoute         Counter
	InitialSlice        Counter
	VetoRTTYOLO         Counter
	VetoPacketLossYOLO  Counter
	RTTIncrease         Counter
}

var EmptyDecisionMetrics DecisionMetrics = DecisionMetrics{
	NoChange:            &EmptyCounter{},
	ForceDirect:         &EmptyCounter{},
	ForceNext:           &EmptyCounter{},
	NoNextRoute:         &EmptyCounter{},
	ABTestDirect:        &EmptyCounter{},
	RTTReduction:        &EmptyCounter{},
	PacketLossMultipath: &EmptyCounter{},
	JitterMultipath:     &EmptyCounter{},
	VetoRTT:             &EmptyCounter{},
	RTTMultipath:        &EmptyCounter{},
	VetoPacketLoss:      &EmptyCounter{},
	FallbackToDirect:    &EmptyCounter{},
	VetoYOLO:            &EmptyCounter{},
	VetoNoRoute:         &EmptyCounter{},
	InitialSlice:        &EmptyCounter{},
	VetoRTTYOLO:         &EmptyCounter{},
	VetoPacketLossYOLO:  &EmptyCounter{},
	RTTIncrease:         &EmptyCounter{},
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
