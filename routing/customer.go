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
}
