package routing

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	RouteTypeDirect   = 0
	RouteTypeNew      = 1
	RouteTypeContinue = 2

	NextRouteTokenSize          = 101
	EncryptedNextRouteTokenSize = NextRouteTokenSize + crypto.MACSize

	ContinueRouteTokenSize          = 42
	EncryptedContinueRouteTokenSize = ContinueRouteTokenSize + crypto.MACSize

	NextRouteTokenSize4          = 100
	EncryptedNextRouteTokenSize4 = NextRouteTokenSize4 + crypto.MACSize

	ContinueRouteTokenSize4          = 41
	EncryptedContinueRouteTokenSize4 = ContinueRouteTokenSize4 + crypto.MACSize
)

type Token interface {
	Type() int
	Encrypt([]byte) ([]byte, int, error)
}

type Client struct {
	Addr      net.UDPAddr
	PublicKey []byte
}

type Server struct {
	Addr      net.UDPAddr
	PublicKey []byte
}

type RelayToken struct {
	ID        uint64
	Addr      net.UDPAddr
	PublicKey []byte
}

type ContinueRouteToken struct {
	Expires uint64

	SessionID      uint64
	SessionVersion uint8
	SessionFlags   uint8 // unused

	Client Client
	Server Server
	Relays []RelayToken

	privateKey []byte
	tokens     []byte
	offset     int
}

func (r *ContinueRouteToken) Type() int {
	return RouteTypeContinue
}

func (r *ContinueRouteToken) Encrypt(privateKey []byte) ([]byte, int, error) {
	r.privateKey = make([]byte, crypto.KeySize)
	rand.Read(r.privateKey)

	r.tokens = make([]byte, EncryptedContinueRouteTokenSize*(len(r.Relays)+2))

	// Encrypt the first node with the client public key
	// and point it to the FIRST relay in the route
	r.encryptToken(r.Relays[0].Addr, r.Client.PublicKey, privateKey)

	for i := range r.Relays {
		// If this is the last relay in the route
		// encrypt it with its public key, but point
		// it to the server
		if i == len(r.Relays)-1 {
			r.encryptToken(r.Server.Addr, r.Relays[i].PublicKey, privateKey)
			break
		}

		// All internal relay node get encrypted with their own
		// public keys and point to the next relay in the route
		r.encryptToken(r.Relays[i+1].Addr, r.Relays[i].PublicKey, privateKey)
	}

	// Encrypt the last node with the server public key
	// and point it to the server itself signifying the end
	r.encryptToken(r.Server.Addr, r.Server.PublicKey, privateKey)

	// We add 2 to the number of tokens to account for the client and the server of route
	// tokens with a client, server, and 5 relays is a total of 7 tokens generated
	return r.tokens, len(r.Relays) + 2, nil
}

func (r *ContinueRouteToken) encryptToken(addr net.UDPAddr, receiverPublicKey []byte, senderPrivateKey []byte) {
	tokenstart := r.offset
	tokenend := r.offset + ContinueRouteTokenSize

	noncestart := tokenstart
	nonceend := noncestart + crypto.NonceSize

	datastart := nonceend
	dataend := tokenend

	rand.Read(r.tokens[noncestart:nonceend])

	// Encode the data into the rest of the node
	var index int
	encoding.WriteUint64(r.tokens[nonceend:], &index, r.Expires)
	encoding.WriteUint64(r.tokens[nonceend:], &index, r.SessionID)
	encoding.WriteUint8(r.tokens[nonceend:], &index, r.SessionVersion)
	encoding.WriteUint8(r.tokens[nonceend:], &index, r.SessionFlags)

	enc := crypto.Seal(r.tokens[datastart:dataend], r.tokens[noncestart:nonceend], receiverPublicKey, senderPrivateKey)
	copy(r.tokens[datastart:], enc)

	r.offset += EncryptedContinueRouteTokenSize
}

type NextRouteToken struct {
	Expires uint64

	SessionID      uint64
	SessionVersion uint8
	SessionFlags   uint8

	KbpsUp   uint32
	KbpsDown uint32

	Client Client
	Server Server
	Relays []RelayToken

	privateKey []byte
	tokens     []byte
	offset     int
}

func (r *NextRouteToken) Type() int {
	return RouteTypeNew
}

func (r *NextRouteToken) Encrypt(privateKey []byte) ([]byte, int, error) {
	if len(r.Relays) <= 0 {
		return nil, 0, errors.New("at least 1 relay is required")
	}

	r.privateKey = make([]byte, crypto.KeySize)
	rand.Read(r.privateKey)

	r.tokens = make([]byte, EncryptedNextRouteTokenSize*(len(r.Relays)+2))

	// Encrypt the first node with the client public key
	// and point it to the FIRST relay in the route
	if r.Client.PublicKey == nil {
		return nil, 0, errors.New("client public key cannot be nil")
	}
	r.encryptToken(&r.Relays[0].Addr, r.Client.PublicKey, privateKey)

	for i := range r.Relays {
		if r.Relays[i].PublicKey == nil {
			return nil, 0, fmt.Errorf("relay public key at index %d cannot be nil", i)
		}

		// If this is the last relay in the route
		// encrypt it with its public key, but point
		// it to the server
		if i == len(r.Relays)-1 {
			r.encryptToken(&r.Server.Addr, r.Relays[i].PublicKey, privateKey)
			break
		}

		// All internal relay node get encrypted with their own
		// public keys and point to the next relay in the route
		r.encryptToken(&r.Relays[i+1].Addr, r.Relays[i].PublicKey, privateKey)
	}

	// Encrypt the last node with the server public key
	// and point it to the server itself signifying the end
	if r.Server.PublicKey == nil {
		return nil, 0, errors.New("server public key cannot be nil")
	}
	r.encryptToken(nil, r.Server.PublicKey, privateKey)

	// We add 2 to the number of tokens to account for the client and the server of route
	// tokens with a client, server, and 5 relays is a total of 7 tokens generated
	return r.tokens, len(r.Relays) + 2, nil
}

// encryptToken works on each token in the chain of tokens
// and increments the overall offset to move across and ecrypt
// each token of the tokens slice to avoid a lot of copy ops
func (r *NextRouteToken) encryptToken(addr *net.UDPAddr, receiverPublicKey []byte, senderPrivateKey []byte) {
	tokenstart := r.offset
	tokenend := r.offset + NextRouteTokenSize

	noncestart := tokenstart
	nonceend := noncestart + crypto.NonceSize

	datastart := nonceend
	dataend := tokenend

	rand.Read(r.tokens[noncestart:nonceend])

	var index int
	encoding.WriteUint64(r.tokens[nonceend:], &index, r.Expires)
	encoding.WriteUint64(r.tokens[nonceend:], &index, r.SessionID)
	encoding.WriteUint8(r.tokens[nonceend:], &index, r.SessionVersion)
	encoding.WriteUint8(r.tokens[nonceend:], &index, r.SessionFlags)
	encoding.WriteUint32(r.tokens[nonceend:], &index, r.KbpsUp)
	encoding.WriteUint32(r.tokens[nonceend:], &index, r.KbpsDown)
	byteaddr := make([]byte, encoding.AddressSize)
	encoding.WriteAddress(byteaddr, addr)
	encoding.WriteBytes(r.tokens[nonceend:], &index, byteaddr, encoding.AddressSize)
	encoding.WriteBytes(r.tokens[nonceend:], &index, r.privateKey, crypto.KeySize)

	enc := crypto.Seal(r.tokens[datastart:dataend], r.tokens[noncestart:nonceend], receiverPublicKey, senderPrivateKey)
	copy(r.tokens[datastart:], enc)

	r.offset += EncryptedNextRouteTokenSize
}
