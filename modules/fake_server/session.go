package fake_server

import (
	"crypto/ed25519"
	"math/rand"
	"net"
	"time"

	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
)

const (
	// SessionLengthRange is the range of session length, that is maximum possible session length - minimum possible session length
	SessionLengthRange = time.Minute * 5

	// SessionLengthMin is the minimum possible session length
	SessionLengthMin = time.Minute * 5

	// HighDirectRTTSessionChance is the percent chance that a session will have high direct RTT
	HighDirectRTTSessionChance = 50.0

	// HighNextRTTSessionChance is the percent chance that a session will have high nxt RTT
	HighNextRTTSessionChance = 5.0

	// RTTRange is the range of RTT, that is maximum possible RTT - minimum possible RTT
	RTTRange = 30.0

	// RTTMin is the minimum possible RTT
	RTTMin = 5.0

	// NearRelayRTTRange is the range of RTT for near relays, that is maximum possible RTT - minimum possible RTT
	NearRelayRTTRange = 15.0

	// NearRelayRTTMin is the minimum possible RTT for a near relay
	NearRelayRTTMin = 1.0

	// HighJitterSessionChance is the percent chance that a session will have occasional spikes of high jitter
	HighJitterSessionChance = 25.0

	// JitterPercent is the percentage of the session that will have spikes in jitter
	JitterPercent = 10.0

	// JitterRange is the range of jitter, that is maximum possible jitter - minimum possible jitter
	JitterRange = 15.0

	// JitterMin is the minimum possible jitter
	JitterMin = 5.0

	// PacketLossSessionChance is the percent chance that a session will have sporadic packet loss
	PacketLossSessionChance = 10.0

	// PacketLossPercent is the percentage of the session that will have spikes in packet loss
	PacketLossPercent = 5.0

	// MaxPacketsLostPercent is the maximum percent of packets to lose during each slice, if packets are being lost
	MaxPacketsLostPercent = 12.5

	// NearRelayJitterPacketLossChance is the percent change a near relay will use the calculated jitter and packet loss instead of the normal values
	NearRelayJitterPacketLossChance = 30.0
)

// Session contains the necessary info for the server to keep track of sessions
// so that each session update request is consistent with the previous response.
type Session struct {
	// Per session config
	startTime          time.Time
	duration           time.Duration
	highDirectRTT      bool
	highNextRTT        bool
	sporaticJitter     bool
	sporaticPacketLoss bool

	// Constant throughout the length of the session
	sessionID            uint64
	userHash             uint64
	clientAddress        net.UDPAddr
	clientRoutePublicKey []byte
	platformType         int32
	connectionType       int32

	// Volatile per slice
	sliceNumber         uint32
	sessionDataBytes    int32
	sessionData         [transport.MaxSessionDataSize]byte
	upgraded            bool
	next                bool
	committed           bool
	directRTT           float32
	directJitter        float32
	directPacketLoss    float32
	nextRTT             float32
	nextJitter          float32
	nextPacketLoss      float32
	numNearRelays       int32
	nearRelayIDs        []uint64
	nearRelayRTT        []int32
	nearRelayJitter     []int32
	nearRelayPacketLoss []int32
	packetsSent         uint64
	packetsLost         uint64
	jitter              float32
}

// NewSession returns a new, randomly generated session.
func NewSession() (Session, error) {
	var highDirectRTT bool
	if rand.Float32()*100 < HighDirectRTTSessionChance {
		highDirectRTT = true
	}

	var highNextRTT bool
	if rand.Float32()*100 < HighNextRTTSessionChance {
		highNextRTT = true
	}

	var sporadicJitter bool
	if rand.Float32()*100 < HighJitterSessionChance {
		sporadicJitter = true
	}

	var sporadicPacketLoss bool
	if rand.Float32()*100 < PacketLossSessionChance {
		sporadicPacketLoss = true
	}

	// Random session length duration
	duration := time.Duration(rand.Intn(int(SessionLengthRange+1)) + int(SessionLengthMin))

	randIPBytes := make([]byte, 0)

	for i := 0; i < 4; i++ {
		randIPBytes = append(randIPBytes, byte(rand.Intn(255)))
	}

	randPort := rand.Intn(65536)

	randomAddress := net.UDPAddr{
		IP:   net.IPv4(randIPBytes[0], randIPBytes[1], randIPBytes[2], randIPBytes[3]),
		Port: randPort,
	}

	randomPublicKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return Session{}, err
	}

	randPlatformType := int32(rand.Intn(transport.PlatformTypeMax))
	randConnectionType := int32(rand.Intn(transport.ConnectionTypeMax))

	session := Session{
		startTime:          time.Now(),
		duration:           duration,
		highDirectRTT:      highDirectRTT,
		highNextRTT:        highNextRTT,
		sporaticJitter:     sporadicJitter,
		sporaticPacketLoss: sporadicPacketLoss,

		sessionID:            rand.Uint64(),
		clientAddress:        randomAddress,
		clientRoutePublicKey: randomPublicKey,
		userHash:             rand.Uint64(),
		platformType:         randPlatformType,
		connectionType:       randConnectionType,
	}

	return session, nil
}

// Advance moves the session forward by one slice, based on the
// session response received from the server backend.
func (session *Session) Advance(response transport.SessionResponsePacket) {
	session.sliceNumber++
	session.sessionDataBytes = response.SessionDataBytes
	session.sessionData = response.SessionData

	session.upgraded = true

	session.next = false
	if response.RouteType != routing.RouteTypeDirect {
		session.next = true
	}

	if response.NearRelaysChanged {
		session.numNearRelays = response.NumNearRelays
		session.nearRelayIDs = response.NearRelayIDs
	}

	session.committed = response.Committed

	// Recalculate the near relay stats, effectively simulating "pinging" the near relays
	session.nearRelayRTT = make([]int32, int(session.numNearRelays))
	session.nearRelayJitter = make([]int32, int(session.numNearRelays))
	session.nearRelayPacketLoss = make([]int32, int(session.numNearRelays))

	// TODO: Once we have a fake_relay that reports up fake ping stats,
	// we need to coordinate those stats so that the predictedRTT the server backend calculates
	// (based on the route matrix) is similar to the nextRTT and nearRelayRTT that this code is generating.
	// Jitter and packet loss are less important for this, since the server backend doesn't to attempt to predict them.

	// These are the base stats for near relays, before randomly deciding if we should make them worse
	for i := 0; i < int(session.numNearRelays); i++ {
		session.nearRelayRTT[i] = int32(rand.Float32()*NearRelayRTTRange + NearRelayRTTMin)
		session.nearRelayJitter[i] = int32(rand.Float32()*JitterRange + JitterMin)
		session.nearRelayPacketLoss[i] = 0
	}

	// These are the base RTTs for direct and next, before randomly deciding if we should make them worse
	directRTT := rand.Float32()*RTTRange + RTTMin
	nextRTT := rand.Float32()*RTTRange + RTTMin

	if session.highDirectRTT {
		directRTT += 250.0
	}

	if session.highNextRTT {
		nextRTT += 250.0

		// If we are simulating high network next RTT,
		// make the near relays have high RTT as well
		for i := 0; i < int(session.numNearRelays); i++ {
			session.nearRelayRTT[i] += 100.0
		}
	}

	// This is the base jitter, before randomly deciding if we should make it worse
	session.jitter = rand.Float32()*JitterRange + JitterMin

	if session.sporaticJitter && rand.Float32()*100 < JitterPercent {
		session.jitter += 30.0 + rand.Float32()*50
	}

	// Packet loss is calculated by actually generating the number of packets sent and lost.
	// Start with assuming 600 packets sent and 0 lost (60 packets per second for 10 seconds).
	packetsSent := 600
	packetsLost := 0
	packetLoss := 0.0

	// Have a chance to randomly lose a percentage of the total packets sent
	if session.sporaticPacketLoss && rand.Float32()*100 < PacketLossPercent {
		packetsLost = int(float32(packetsSent) * MaxPacketsLostPercent / 100.0)
		packetsSent -= packetsLost
		packetLoss = float64(packetsLost) / float64(packetsSent)
	}

	session.packetsSent += uint64(packetsSent)
	session.packetsLost += uint64(packetsLost)

	// Update the direct and next stats based on the above calcuated data
	if session.next {
		// The session is taking network next, so apply the jitter
		// and packet loss we calculated above to the next stats
		session.directRTT = directRTT
		session.directJitter = rand.Float32()*JitterRange + JitterMin
		session.directPacketLoss = 0

		session.nextRTT = nextRTT
		session.nextJitter = session.jitter
		session.nextPacketLoss = float32(packetLoss)
	} else {
		// The session is taking the direct route, so apply the jitter
		// and packet loss we calculated above to the direct stats and zero out the next stats
		session.directRTT = directRTT
		session.directJitter = session.jitter
		session.directPacketLoss = float32(packetLoss)

		session.nextRTT = 0
		session.nextJitter = 0
		session.nextPacketLoss = 0
	}

	// Allow for a random chance on each near relay to use the worse
	// jitter and packet loss value
	for i := 0; i < int(session.numNearRelays); i++ {
		if rand.Float32()*100.0 < NearRelayJitterPacketLossChance {
			session.nearRelayJitter[i] = int32(session.jitter)
			session.nearRelayPacketLoss[i] = int32(packetLoss)
		} else {
			session.nearRelayJitter[i] = int32(rand.Float32()*JitterRange + JitterMin)
			session.nearRelayPacketLoss[i] = 0
		}
	}
}
