package routing

// Nibblin is the quantum used to calculate costs in the route_matrix and cost_matrix
type Nibblin uint64

// ToCents returns the nibblin count converted to cents
func (n Nibblin) ToCents() float64 {
	return float64(n) / 1e9
}
