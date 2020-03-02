package core

import (
	"net"
)

type SessionToken struct {
	expireTimestamp uint64
	sessionID       uint64
	sessionVersion  uint8
	sessionFlags    uint8
	kbpsUp          uint32
	kbpsDown        uint32
	nextAddress     *net.UDPAddr
	privateKey      []byte
}

type ContinueToken struct {
	expireTimestamp uint64
	sessionID       uint64
	sessionVersion  uint8
	sessionFlags    uint8
}
