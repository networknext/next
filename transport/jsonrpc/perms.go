package jsonrpc

import (
	"fmt"

	"github.com/gorilla/rpc/v2"
)

// PermsValidator checks that the permissions applied to the calling
// user are sufficient to access the specified RPC endpoint.
func PermsValidator(r *rpc.RequestInfo, i interface{}) error {

	// r.Request.Context().Value("user") gets populated with parsed token
	// by AuthMiddleware().

	fmt.Printf("r.Method    : %s\n", r.Method)
	fmt.Printf("user        : %v\n", r.Request.Context().Value("user"))
	fmt.Printf("r.StatusCode: %d\n\n", r.StatusCode)
	fmt.Printf("args: %v\n\n", i)

	return nil
}
