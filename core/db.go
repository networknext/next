package core

import "net"

type ServerEntry struct {
	PublicKey   []byte
	PrivateAddr *net.UDPAddr

	VersionMajor int32
	VersionMinor int32
	VersionPatch int32

	DatacenterID      uint64
	DatacenterName    string
	DatacenterEnabled int32
}

func (se *ServerEntry) Serialize(stream Stream) error {
	stream.SerializeBytes(se.PublicKey)
	stream.SerializeAddress(se.PrivateAddr)
	stream.SerializeInteger(&se.VersionMajor, 0, SDKVersionMajorMax)
	stream.SerializeInteger(&se.VersionMinor, 0, SDKVersionMinorMax)
	stream.SerializeInteger(&se.VersionPatch, 0, SDKVersionPatchMax)
	stream.SerializeUint64(&se.DatacenterID)
	stream.SerializeBytes([]byte(se.DatacenterName))
	stream.SerializeInteger(&se.DatacenterEnabled, 0, 1)
	return stream.Error()
}
