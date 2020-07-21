package billing

// Note: Nibblins is a made up unit in the old backend presumably to deal with floating point issues. 1000000000 Niblins = $0.01 USD
func NibblinsToCents(nibblins uint64) uint64 {
	return nibblins / 1e9
}

// Note: Nibblins is a made up unit in the old backend presumably to deal with floating point issues. 1000000000 Niblins = $0.01 USD
func CentsToNibblins(cents uint64) uint64 {
	return cents * 1e9
}
