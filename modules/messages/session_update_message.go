package messages

import (
	"fmt"
	"net"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/encoding"
)

const (
	SessionUpdateMessageVersion_Min   = 1
	SessionUpdateMessageVersion_Max   = 1
	SessionUpdateMessageVersion_Write = 1
)

type SessionUpdateMessage struct {

	// always

	Version          byte
	Timestamp        uint64
	SessionId        uint64
	SliceNumber      uint32
	RealPacketLoss   float32
	RealJitter       float32
	RealOutOfOrder   float32
	SessionFlags     uint64
	GameEvents       uint64
	DirectRTT        float32
	DirectJitter     float32
	DirectPacketLoss float32
	DirectKbpsUp     uint32
	DirectKbpsDown   uint32

	// next only

	NextRTT            float32
	NextJitter         float32
	NextPacketLoss     float32
	NextKbpsUp         uint32
	NextKbpsDown       uint32
	NextPredictedRTT   uint32
	NextNumRouteRelays uint32
	NextRouteRelayId   [constants.MaxRouteRelays]uint64

	// first slice only

	NumTags byte
	Tags    [constants.MaxTags]uint64

	// first slice and summary slice only

	DatacenterId     uint64
	BuyerId          uint64
	MatchId          uint64
	UserHash         uint64
	Latitude         float32
	Longitude        float32
	ClientAddress    net.UDPAddr
	ServerAddress    net.UDPAddr
	ConnectionType   byte
	PlatformType     byte
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte

	// summary slice only

	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	ClientToServerPacketsOutOfOrder uint64
	ServerToClientPacketsOutOfOrder uint64
	TotalEnvelopeBytesUp            uint64
	TotalEnvelopeBytesDown          uint64
	DurationOnNext                  uint32
	SessionDuration                 uint32
	StartTimestamp                  uint64
}

func (message *SessionUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < SessionUpdateMessageVersion_Min || message.Version > SessionUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid session update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	// always

	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint32(buffer, &index, message.SliceNumber)
	encoding.WriteFloat32(buffer, &index, message.RealPacketLoss)
	encoding.WriteFloat32(buffer, &index, message.RealJitter)
	encoding.WriteFloat32(buffer, &index, message.RealOutOfOrder)
	encoding.WriteUint64(buffer, &index, message.SessionFlags)
	encoding.WriteUint64(buffer, &index, message.GameEvents)
	encoding.WriteFloat32(buffer, &index, message.DirectRTT)
	encoding.WriteFloat32(buffer, &index, message.DirectJitter)
	encoding.WriteFloat32(buffer, &index, message.DirectPacketLoss)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsUp)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsDown)

	// next only

	if (message.SessionFlags & constants.SessionFlags_Next) != 0 {
		encoding.WriteFloat32(buffer, &index, message.NextRTT)
		encoding.WriteFloat32(buffer, &index, message.NextJitter)
		encoding.WriteFloat32(buffer, &index, message.NextPacketLoss)
		encoding.WriteUint32(buffer, &index, message.NextKbpsUp)
		encoding.WriteUint32(buffer, &index, message.NextKbpsDown)
		encoding.WriteUint32(buffer, &index, message.NextPredictedRTT)
		encoding.WriteUint32(buffer, &index, message.NextNumRouteRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			encoding.WriteUint64(buffer, &index, message.NextRouteRelayId[i])
		}
	}

	// first slice only

	if message.SliceNumber == 0 {
		encoding.WriteUint8(buffer, &index, message.NumTags)
		for i := 0; i < int(message.NumTags); i++ {
			encoding.WriteUint64(buffer, &index, message.Tags[i])
		}
	}

	// first slice or summary slice

	if message.SliceNumber == 0 || (message.SessionFlags&constants.SessionFlags_Summary) != 0 {
		encoding.WriteUint64(buffer, &index, message.DatacenterId)
		encoding.WriteUint64(buffer, &index, message.BuyerId)
		encoding.WriteUint64(buffer, &index, message.MatchId)
		encoding.WriteUint64(buffer, &index, message.UserHash)
		encoding.WriteFloat32(buffer, &index, message.Latitude)
		encoding.WriteFloat32(buffer, &index, message.Longitude)
		encoding.WriteAddress(buffer, &index, &message.ClientAddress)
		encoding.WriteAddress(buffer, &index, &message.ServerAddress)
		encoding.WriteUint8(buffer, &index, message.ConnectionType)
		encoding.WriteUint8(buffer, &index, message.PlatformType)
		encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
		encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
		encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
	}

	// summary slice only

	if (message.SessionFlags & constants.SessionFlags_Summary) != 0 {
		encoding.WriteUint64(buffer, &index, message.ClientToServerPacketsSent)
		encoding.WriteUint64(buffer, &index, message.ServerToClientPacketsSent)
		encoding.WriteUint64(buffer, &index, message.ClientToServerPacketsLost)
		encoding.WriteUint64(buffer, &index, message.ServerToClientPacketsLost)
		encoding.WriteUint64(buffer, &index, message.ClientToServerPacketsOutOfOrder)
		encoding.WriteUint64(buffer, &index, message.ServerToClientPacketsOutOfOrder)
		encoding.WriteUint32(buffer, &index, message.SessionDuration)
		encoding.WriteUint64(buffer, &index, message.TotalEnvelopeBytesUp)
		encoding.WriteUint64(buffer, &index, message.TotalEnvelopeBytesDown)
		encoding.WriteUint32(buffer, &index, message.DurationOnNext)
		encoding.WriteUint64(buffer, &index, message.StartTimestamp)
	}

	return buffer[:index]
}

func (message *SessionUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read session update message version")
	}

	if message.Version < SessionUpdateMessageVersion_Min || message.Version > SessionUpdateMessageVersion_Max {
		return fmt.Errorf("invalid session update message version %d", message.Version)
	}

	// always

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SliceNumber) {
		return fmt.Errorf("failed to read slice number")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealPacketLoss) {
		return fmt.Errorf("failed to read real packet loss")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealJitter) {
		return fmt.Errorf("failed to read real jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealOutOfOrder) {
		return fmt.Errorf("failed to read real out of order")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionFlags) {
		return fmt.Errorf("failed to read session flags")
	}

	if !encoding.ReadUint64(buffer, &index, &message.GameEvents) {
		return fmt.Errorf("failed to read game events")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectRTT) {
		return fmt.Errorf("failed to read direct rtt")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectJitter) {
		return fmt.Errorf("failed to read direct jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectPacketLoss) {
		return fmt.Errorf("failed to read direct packet loss")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DirectKbpsUp) {
		return fmt.Errorf("failed to read direct kbps up")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DirectKbpsDown) {
		return fmt.Errorf("failed to read direct kbps down")
	}

	// next only

	if (message.SessionFlags & constants.SessionFlags_Next) != 0 {

		if !encoding.ReadFloat32(buffer, &index, &message.NextRTT) {
			return fmt.Errorf("failed to read next rtt")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NextJitter) {
			return fmt.Errorf("failed to read next jitter")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NextPacketLoss) {
			return fmt.Errorf("failed to read next packet loss")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextKbpsUp) {
			return fmt.Errorf("failed to read next kbps up")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextKbpsDown) {
			return fmt.Errorf("failed to read next kbps down")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextPredictedRTT) {
			return fmt.Errorf("failed to read next predicted rtt")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextNumRouteRelays) {
			return fmt.Errorf("failed to read next num route relays")
		}

		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			if !encoding.ReadUint64(buffer, &index, &message.NextRouteRelayId[i]) {
				return fmt.Errorf("failed to read next route relay id")
			}
		}
	}

	// first slice only

	if message.SliceNumber == 0 {

		if !encoding.ReadUint8(buffer, &index, &message.NumTags) {
			return fmt.Errorf("failed to read num tags")
		}

		for i := 0; i < int(message.NumTags); i++ {
			if !encoding.ReadUint64(buffer, &index, &message.Tags[i]) {
				return fmt.Errorf("failed to read tags")
			}
		}
	}

	// summary only

	if (message.SessionFlags & constants.SessionFlags_Summary) != 0 {

		if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
			return fmt.Errorf("failed to read datacenter id")
		}

		if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
			return fmt.Errorf("failed to read buyer id")
		}

		if !encoding.ReadUint64(buffer, &index, &message.MatchId) {
			return fmt.Errorf("failed to read match id")
		}

		if !encoding.ReadUint64(buffer, &index, &message.UserHash) {
			return fmt.Errorf("failed to read user hash")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.Latitude) {
			return fmt.Errorf("failed to read latitude")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.Longitude) {
			return fmt.Errorf("failed to read longitude")
		}

		if !encoding.ReadAddress(buffer, &index, &message.ClientAddress) {
			return fmt.Errorf("failed to read client address")
		}

		if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
			return fmt.Errorf("failed to read server address")
		}

		if !encoding.ReadUint8(buffer, &index, &message.ConnectionType) {
			return fmt.Errorf("failed to read connection type")
		}

		if !encoding.ReadUint8(buffer, &index, &message.PlatformType) {
			return fmt.Errorf("failed to read platform type")
		}

		if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Major) {
			return fmt.Errorf("failed to read sdk version major")
		}

		if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Minor) {
			return fmt.Errorf("failed to read sdk version minor")
		}

		if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Patch) {
			return fmt.Errorf("failed to read sdk version patch")
		}

		if !encoding.ReadUint64(buffer, &index, &message.ClientToServerPacketsSent) {
			return fmt.Errorf("failed to read client to server packets sent")
		}

		if !encoding.ReadUint64(buffer, &index, &message.ServerToClientPacketsSent) {
			return fmt.Errorf("failed to read server to client packets sent")
		}

		if !encoding.ReadUint64(buffer, &index, &message.ClientToServerPacketsLost) {
			return fmt.Errorf("failed to read client to server packets lost")
		}

		if !encoding.ReadUint64(buffer, &index, &message.ServerToClientPacketsLost) {
			return fmt.Errorf("failed to read server to client packets lost")
		}

		if !encoding.ReadUint64(buffer, &index, &message.ClientToServerPacketsOutOfOrder) {
			return fmt.Errorf("failed to read client to server packets out of order")
		}

		if !encoding.ReadUint64(buffer, &index, &message.ServerToClientPacketsOutOfOrder) {
			return fmt.Errorf("failed to read server to client packets out of order")
		}

		if !encoding.ReadUint32(buffer, &index, &message.SessionDuration) {
			return fmt.Errorf("failed to read session duration")
		}

		if !encoding.ReadUint64(buffer, &index, &message.TotalEnvelopeBytesUp) {
			return fmt.Errorf("failed to read total envelope bytes up sum")
		}

		if !encoding.ReadUint64(buffer, &index, &message.TotalEnvelopeBytesDown) {
			return fmt.Errorf("failed to read total envelope bytes down sum")
		}

		if !encoding.ReadUint32(buffer, &index, &message.DurationOnNext) {
			return fmt.Errorf("failed to read duration on next")
		}

		if !encoding.ReadUint64(buffer, &index, &message.StartTimestamp) {
			return fmt.Errorf("failed to read start timestamp")
		}
	}

	return nil
}
