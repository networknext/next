package core

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
)

type RouteContext struct {
	RelayAddresses  []*net.UDPAddr
	RelayPublicKeys [][]byte
	RouteMatrix     *RouteMatrix
	RelayIdToIndex  map[RelayId]int
}

type Route struct {
	RTT        float32
	Jitter     float32
	PacketLoss float32
	RelayIds   []RelayId
}

func GetRouteHash(relayIds []RelayId) uint64 {
	hash := fnv.New64a()
	for _, v := range relayIds {
		a := make([]byte, 4)
		binary.LittleEndian.PutUint32(a, uint32(v))
		hash.Write(a)
	}
	return hash.Sum64()
}

type RouteStats struct {
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

type RelayStats struct {
	Id RelayId
	RouteStats
}

type RouteSample struct {
	DirectStats RouteStats
	NextStats   RouteStats
	NearRelays  []RelayStats
}

type RouteSlice struct {
	Flags          uint64
	RouteSample    RouteSample
	PredictedRoute *Route
}

const RouteSliceVersion = uint32(0)

const RouteSliceMagic = uint32(0x12345678)

func (slice *RouteSlice) Serialize(stream Stream) error {

	var magic uint32
	if stream.IsWriting() {
		magic = RouteSliceMagic
	}
	stream.SerializeUint32(&magic)
	if stream.IsReading() && magic != RouteSliceMagic {
		return fmt.Errorf("expected route slice magic %x, got %x", RouteSliceMagic, magic)
	}

	var version uint32
	if stream.IsWriting() {
		version = RouteSliceVersion
	}
	stream.SerializeUint32(&version)
	if stream.IsReading() && version != RouteSliceVersion {
		return fmt.Errorf("expected route slice version %d, got %d", RouteSliceVersion, version)
	}

	stream.SerializeUint64(&slice.Flags)

	stream.SerializeFloat32(&slice.RouteSample.DirectStats.RTT)
	stream.SerializeFloat32(&slice.RouteSample.DirectStats.Jitter)
	stream.SerializeFloat32(&slice.RouteSample.DirectStats.PacketLoss)

	stream.SerializeFloat32(&slice.RouteSample.NextStats.RTT)
	stream.SerializeFloat32(&slice.RouteSample.NextStats.Jitter)
	stream.SerializeFloat32(&slice.RouteSample.NextStats.PacketLoss)

	hasPredictedRoute := stream.IsWriting() && slice.PredictedRoute != nil
	stream.SerializeBool(&hasPredictedRoute)
	if hasPredictedRoute {
		if stream.IsReading() {
			slice.PredictedRoute = &Route{}
		}
		stream.SerializeFloat32(&slice.PredictedRoute.RTT)
		stream.SerializeFloat32(&slice.PredictedRoute.Jitter)
		stream.SerializeFloat32(&slice.PredictedRoute.PacketLoss)
		numRelayIds := uint32(len(slice.PredictedRoute.RelayIds))
		stream.SerializeUint32(&numRelayIds)
		if stream.IsReading() {
			if numRelayIds > MaxRelays {
				return fmt.Errorf("too many relays in route: %d", numRelayIds)
			}
			slice.PredictedRoute.RelayIds = make([]RelayId, numRelayIds)
		}
		for i := range slice.PredictedRoute.RelayIds {
			relayId := uint64(slice.PredictedRoute.RelayIds[i])
			stream.SerializeUint64(&relayId)
			if stream.IsReading() {
				slice.PredictedRoute.RelayIds[i] = RelayId(relayId)
			}
		}

		numNearRelays := uint32(len(slice.RouteSample.NearRelays))
		stream.SerializeUint32(&numNearRelays)
		if stream.IsReading() {
			if numNearRelays > MaxNearRelays {
				return fmt.Errorf("too many near relays in route slice: %d", numNearRelays)
			}
			slice.RouteSample.NearRelays = make([]RelayStats, numNearRelays)
		}
		for i := 0; i < int(numNearRelays); i++ {
			relayId := uint64(slice.RouteSample.NearRelays[i].Id)
			stream.SerializeUint64(&relayId)
			if stream.IsReading() {
				slice.RouteSample.NearRelays[i].Id = RelayId(relayId)
			}
			stream.SerializeFloat32(&slice.RouteSample.NearRelays[i].RTT)
			stream.SerializeFloat32(&slice.RouteSample.NearRelays[i].Jitter)
			stream.SerializeFloat32(&slice.RouteSample.NearRelays[i].PacketLoss)
		}
	}

	return stream.Error()
}
