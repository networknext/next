package routing

const (
	RouteTypeDirect   = 0
	RouteTypeNew      = 1
	RouteTypeContinue = 2
)

type Route struct {
	Type      int
	NumTokens int
	Tokens    []byte
	Multipath bool
}
