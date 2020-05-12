package routing

const (
	ConnectionTypeUnknown  = 0
	ConnectionTypeWired    = 1
	ConnectionTypeWifi     = 2
	ConnectionTypeCellular = 3
)

// ConnectionTypeText is similar to http.StatusText(int) which converts the code to a readable text format
func ConnectionTypeText(conntype int32) string {
	switch conntype {
	case ConnectionTypeWired:
		return "wired"
	case ConnectionTypeWifi:
		return "wifi"
	case ConnectionTypeCellular:
		return "cellular"
	default:
		return "unknown"
	}
}
