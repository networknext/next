package routing

import (
	"crypto/rand"
	"encoding/binary"
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

type NextRouteDecision struct {
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
	token      []byte
	offset     int
}

func (r *NextRouteDecision) Encrypt(privateKey []byte) []byte {
	r.privateKey = make([]byte, crypto.KeySize)
	rand.Read(r.privateKey)

	r.token = make([]byte, 117*(len(r.Relays)+2))

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

func (r *NextRouteDecision) encryptToken(addr net.UDPAddr, receiverPublicKey []byte, senderPrivateKey []byte) {
	// Create space for the entire encoded node
	node := make([]byte, EncryptedNextRouteTokenSize)

	// Create an copy a nonce to the start of the node
	nonce := make([]byte, crypto.NonceSize)
	rand.Read(nonce)
	copy(node[0:], nonce)

	// Encode the data into the rest of the node
	binary.LittleEndian.PutUint64(node[crypto.NonceSize:], r.Expires)
	binary.LittleEndian.PutUint64(node[crypto.NonceSize+8:], r.SessionId)
	node[crypto.NonceSize+8+8] = r.SessionVersion
	node[crypto.NonceSize+8+8+1] = r.SessionFlags
	binary.LittleEndian.PutUint32(node[crypto.NonceSize+8+8+1+1:], r.KbpsUp)
	binary.LittleEndian.PutUint32(node[crypto.NonceSize+8+8+1+1+4:], r.KbpsUp)
	encoding.WriteAddress(node[crypto.NonceSize+8+8+1+1+4+4:], &r.Client.Addr)
	copy(node[crypto.NonceSize+8+8+1+1+4+4+encoding.AddressSize:], r.privateKey)

	copy(r.token[r.offset:], crypto.Seal(node[crypto.NonceSize:], nonce, receiverPublicKey, senderPrivateKey))

	r.offset += 117
}
