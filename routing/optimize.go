package routing

type CostMatrix struct{}

func (m *CostMatrix) UnmarshalBinary(data []byte) error {
	return nil
}

func (m CostMatrix) MarshalBinary() ([]byte, error) {
	return nil, nil
}

// Optimize will fill up a *RouteMatrix with the optimized routes based on cost.
func (m *CostMatrix) Optimize(routes *RouteMatrix) error {
	return nil
}

type RouteMatrix struct{}

func (m *RouteMatrix) UnmarshalBinary(data []byte) error {
	return nil
}

func (m RouteMatrix) MarshalBinary() ([]byte, error) {
	return nil, nil
}
