package stats

func NewEntityID(kind string, name string) *EntityId {
	return &EntityId{
		Kind: kind,
		Name: name,
	}
}
