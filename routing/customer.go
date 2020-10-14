package routing

import "cloud.google.com/go/firestore"

type Customer struct {
	Code                   string
	Name                   string
	AutomaticSignInDomains string
	Active                 bool
	Debug                  bool
	BuyerRef               *firestore.DocumentRef
	SellerRef              *firestore.DocumentRef
	SqlID                  int64
	BuyerID                uint64 // binary.LittleEndian.Uint64(publicKey[:8]),
	SellerID               string // ID: name
}
