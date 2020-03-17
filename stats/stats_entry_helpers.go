package stats

func NewEntityID(kind string, relayName string) *EntityId {
	return &EntityId{
		Kind: kind,
		Name: relayName,
	}
}
