package routing

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	DecisionTypeDirect   = 0
	DecisionTypeNew      = 1
	DecisionTypeContinue = 2

	EncryptedNextRouteTokenSize = 117 // Need to find out how this size is made
)

type Client struct {
	Addr      net.UDPAddr
	PublicKey []byte
}

type Server struct {
	Addr      net.UDPAddr
	PublicKey []byte
}

type Route struct {
	Relays []Relay
	Stats  Stats
}

type ContinueRouteDecition struct {
	Expires uint64

	SessionId      uint64
	SessionVersion uint8
	SessionFlags   uint8

	Client Client
	Server Server
	Relays []Relay

	privateKey []byte
	token      []byte
	offset     int
}

func (r *ContinueRouteDecition) Encrypt(privateKey []byte) []byte {
	r.privateKey = make([]byte, crypto.KeySize)
	rand.Read(r.privateKey)

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

	return r.token
}

func (r *ContinueRouteDecition) encryptToken(addr net.UDPAddr, receiverPublicKey []byte, senderPrivateKey []byte) {
	// Create space for the entire encoded node
	node := make([]byte, 58)

	// Create an copy a nonce to the start of the node
	nonce := make([]byte, crypto.NonceSize)
	rand.Read(nonce)
	copy(node[0:], nonce)

	// Encode the data into the rest of the node
	binary.LittleEndian.PutUint64(node[crypto.NonceSize:], r.Expires)
	binary.LittleEndian.PutUint64(node[crypto.NonceSize+8:], r.SessionId)
	node[crypto.NonceSize+8+8] = r.SessionVersion
	node[crypto.NonceSize+8+8+1] = r.SessionFlags

	copy(r.token[r.offset:], crypto.Seal(node[crypto.NonceSize:], nonce, receiverPublicKey, senderPrivateKey))

	r.offset += 58
}

type NextRouteToken struct {
	Expires uint64

	SessionId      uint64
	SessionVersion uint8
	SessionFlags   uint8

	KbpsUp   uint32
	KbpsDown uint32

	Client Client
	Server Server
	Relays []Relay

	privateKey []byte
	tokens     []byte
	offset     int
}

func (r *NextRouteToken) Encrypt(privateKey []byte) ([]byte, error) {
	if len(r.Relays) <= 0 {
		return nil, errors.New("at least 1 relay is required")
	}

	r.privateKey = make([]byte, crypto.KeySize)
	rand.Read(r.privateKey)

	r.tokens = make([]byte, EncryptedNextRouteTokenSize*(len(r.Relays)+2))

	// Encrypt the first node with the client public key
	// and point it to the FIRST relay in the route
	if r.Client.PublicKey == nil {
		return nil, errors.New("client public key cannot be nil")
	}
	r.encryptToken(r.Relays[0].Addr, r.Client.PublicKey, privateKey)

	for i := range r.Relays {
		if r.Relays[i].PublicKey == nil {
			return nil, fmt.Errorf("relay public key at index %d cannot be nil", i)
		}

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
	if r.Server.PublicKey == nil {
		return nil, errors.New("server public key cannot be nil")
	}
	r.encryptToken(r.Server.Addr, r.Server.PublicKey, privateKey)

	return r.tokens, nil
}

// encryptToken works on each token in the chain of tokens
// and increments the overall offset to move across and ecrypt
// each token of the tokens slice to avoid a lot of copy ops
func (r *NextRouteToken) encryptToken(addr net.UDPAddr, receiverPublicKey []byte, senderPrivateKey []byte) {
	tokenstart := r.offset
	tokenend := r.offset + EncryptedNextRouteTokenSize

	noncestart := tokenstart
	nonceend := noncestart + crypto.NonceSize

	datastart := nonceend
	dataend := tokenend

	rand.Read(r.tokens[noncestart:nonceend])

	var index int
	encoding.WriteUint64(r.tokens[nonceend:], &index, r.Expires)
	encoding.WriteUint64(r.tokens[nonceend:], &index, r.SessionId)
	encoding.WriteUint8(r.tokens[nonceend:], &index, r.SessionVersion)
	encoding.WriteUint8(r.tokens[nonceend:], &index, r.SessionFlags)
	encoding.WriteUint32(r.tokens[nonceend:], &index, r.KbpsUp)
	encoding.WriteUint32(r.tokens[nonceend:], &index, r.KbpsDown)
	byteaddr := make([]byte, encoding.AddressSize)
	encoding.WriteAddress(byteaddr, &r.Client.Addr)
	encoding.WriteBytes(r.tokens[nonceend:], &index, byteaddr, encoding.AddressSize)
	encoding.WriteBytes(r.tokens[nonceend:], &index, r.privateKey, crypto.KeySize)

	enc := crypto.Seal(r.tokens[datastart:dataend], r.tokens[noncestart:nonceend], receiverPublicKey, senderPrivateKey)
	copy(r.tokens[datastart:dataend], enc)

	r.offset += EncryptedNextRouteTokenSize
}
