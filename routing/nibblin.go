package routing

// Nibblin is a fake monetary currency to help with float point imprecision
// during calculations (1,000,000,000 Nibblins = $0.01 USD)
type Nibblin uint64

func CentsToNibblins(cents float64) Nibblin {
	return Nibblin(cents * 1e9)
}

func (n Nibblin) ToCents() float64 {
	return float64(n) / 1e9
}
