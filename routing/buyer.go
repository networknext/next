package routing

type Buyer struct {
	ID        uint64
	Name      string
	Active    bool
	Live      bool
	PublicKey []byte
	Config    RoutingRulesSettings
}
