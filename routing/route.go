package routing

const (
	RouteTypeDirect = iota
	RouteTypeNew
	RouteTypeContinue
)

type Route struct {
	Type      int
	NumTokens int
	Tokens    []byte
	Multipath bool
}
