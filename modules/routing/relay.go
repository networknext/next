package routing

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/backend/modules/crypto"
)

const (
	// EncryptedRelayTokenSize ...
	EncryptedRelayTokenSize = crypto.KeySize + crypto.MACSize

	RelayTimeout = 30 * time.Second

	// MaxRelayAddressLength ...
	MaxRelayAddressLength = 256

	MaxRelayNameLength = 63
)

type RelayState uint32

func (state RelayState) String() string {
	switch state {
	case RelayStateEnabled:
		return "enabled"
	case RelayStateMaintenance:
		return "maintenance"
	case RelayStateDisabled:
		return "disabled"
	case RelayStateQuarantine:
		return "quarantine"
	case RelayStateDecommissioned:
		return "decommissioned"
	case RelayStateOffline:
		return "offline"
	default:
		return "unknown"
	}
}

// ParseRelayState returns a relay state type given the state in string form
func ParseRelayState(str string) (RelayState, error) {
	switch str {
	case "enabled":
		return RelayStateEnabled, nil
	case "maintenance":
		return RelayStateMaintenance, nil
	case "disabled":
		return RelayStateDisabled, nil
	case "quarantine":
		return RelayStateQuarantine, nil
	case "decommissioned":
		return RelayStateDecommissioned, nil
	case "offline":
		return RelayStateOffline, nil
	default:
		return RelayStateEnabled, fmt.Errorf("invalid relay state '%s'", str)
	}
}

func GetRelayStateSQL(state int64) (RelayState, error) {
	switch state {
	case 0:
		return RelayStateEnabled, nil
	case 1:
		return RelayStateMaintenance, nil
	case 2:
		return RelayStateDisabled, nil
	case 3:
		return RelayStateQuarantine, nil
	case 4:
		return RelayStateDecommissioned, nil
	case 5:
		return RelayStateOffline, nil
	default:
		return RelayStateDisabled, fmt.Errorf("invalid relay state '%d'", state)
	}
}

const (
	// RelayStateEnabled if running and communicating with backend
	RelayStateEnabled RelayState = 0
	// RelayStateMaintenance System does a clean shutdown of the relay process
	RelayStateMaintenance RelayState = 1
	// RelayStateDisabled System does a clean shutdown and is shown as disabled as a temporary state before decommissioning in case it needs to be re-enabled
	RelayStateDisabled RelayState = 2
	// RelayStateQuarantine Backend has detected an unexpected disruption in service from this relay and has removed it from getting packets until it is manually added back into the system
	RelayStateQuarantine RelayState = 3
	// RelayStateDecommissioned System is removed from the UI and lists. Reusable fields like IP and name are cleared in firestore. Most data is retained for historical purposes
	RelayStateDecommissioned RelayState = 4
	// RelayStateOffline Relay has stopped communicating with the backend for some unknown reason
	RelayStateOffline RelayState = 5
)

// BandWidthRule Flat / Burst / Pool: Relates to bandwidth
type BandWidthRule uint32

const (
	BWRuleNone  BandWidthRule = iota
	BWRuleFlat  BandWidthRule = iota // can not go over allocated amount
	BWRuleBurst BandWidthRule = iota // can go over amount
	BWRulePool  BandWidthRule = iota // supplier gives X amount of bandwidth for all relays in the pool
)

func ParseBandwidthRule(bwRule string) (BandWidthRule, error) {
	switch bwRule {
	case "none":
		return BWRuleNone, nil
	case "flat":
		return BWRuleFlat, nil
	case "burst":
		return BWRuleBurst, nil
	case "pool":
		return BWRulePool, nil
	default:
		return BWRuleNone, fmt.Errorf("invalid BandWidthRule '%s'", bwRule)
	}
}

func GetBandwidthRuleSQL(bwRule int64) (BandWidthRule, error) {
	switch bwRule {
	case 0:
		return BWRuleNone, nil
	case 1:
		return BWRuleFlat, nil
	case 2:
		return BWRuleBurst, nil
	case 3:
		return BWRulePool, nil
	default:
		return BWRuleNone, fmt.Errorf("invalid BandWidthRule '%d'", bwRule)
	}
}

// MachineType is the type of server the relay is running on
type MachineType uint32

const (
	NoneSpecified  MachineType = iota
	BareMetal      MachineType = iota
	VirtualMachine MachineType = iota
)

func ParseMachineType(machineType string) (MachineType, error) {
	switch machineType {
	case "none":
		return NoneSpecified, nil
	case "bare-metal":
		return BareMetal, nil
	case "vm":
		return VirtualMachine, nil
	default:
		return NoneSpecified, fmt.Errorf("invalid MachineType '%s'", machineType)
	}
}

func GetMachineTypeSQL(machineType int64) (MachineType, error) {
	switch machineType {
	case 0:
		return NoneSpecified, nil
	case 1:
		return BareMetal, nil
	case 2:
		return VirtualMachine, nil
	default:
		return NoneSpecified, fmt.Errorf("invalid MachineType '%d'", machineType)
	}
}

type RelayVersion struct {
	Major int32
	Minor int32
	Patch int32
}

const (
	RelayVersionEqual = iota
	RelayVersionOlder
	RelayVersionNewer
)

func (a RelayVersion) Compare(b RelayVersion) int {
	if a.Major > b.Major {
		return RelayVersionNewer
	}
	if a.Major == b.Major {
		if a.Minor > b.Minor {
			return RelayVersionNewer
		}

		if a.Minor == b.Minor {
			if a.Patch == b.Patch {
				return RelayVersionEqual
			}

			if a.Patch > b.Patch {
				return RelayVersionNewer
			}

			if a.Patch < b.Patch {
				return RelayVersionOlder
			}
		}
	}
	return RelayVersionOlder
}

func (a RelayVersion) AtLeast(b RelayVersion) bool {
	return a.Compare(b) != RelayVersionOlder
}

func (v RelayVersion) Parse(version string) error {
	components := strings.Split(version, ".")
	if len(components) != 3 {
		return fmt.Errorf("version string does not follow major.minor.patch format: %s", version)
	}

	major, err := strconv.ParseInt(components[0], 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse major component %s as an int", components[0])
	}

	minor, err := strconv.ParseInt(components[1], 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse minor component %s as an int", components[1])
	}

	patch, err := strconv.ParseInt(components[2], 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse patch component %s as an int", components[2])
	}

	v.Major = int32(major)
	v.Minor = int32(minor)
	v.Patch = int32(patch)

	return nil
}

func (v RelayVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

type Relay struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`

	Addr         net.UDPAddr `json:"addr"`
	InternalAddr net.UDPAddr `json:"internal_addr"`
	PublicKey    []byte      `json:"public_key"`

	Seller          Seller     `json:"seller"`          // TODO: chopping block
	BillingSupplier string     `json:"billingSupplier"` // Seller FK
	Datacenter      Datacenter `json:"datacenter"`

	NICSpeedMbps        int32 `json:"nicSpeedMbps"`
	IncludedBandwidthGB int32 `json:"includedBandwidthGB"`
	MaxBandwidthMbps    int32 `json:"maxBandwidthMbps"`

	State RelayState `json:"state"`

	ManagementAddr string `json:"management_addr"` // TODO: convert to a legit network type
	SSHUser        string `json:"ssh_user"`
	SSHPort        int64  `json:"ssh_port"`

	MaxSessions uint32 `json:"max_sessions"`

	// EgressPriceOverride (nibblins/GB) is used to calculate route price instead of seller price when > 0
	EgressPriceOverride Nibblin `json:"egressPriceOverride"`
	// MRC is the monthly recurring cost for the relay
	MRC Nibblin `json:"monthlyRecurringChargeNibblins"`
	// Overage is the charge/penalty if we exceed the bandwidth alloted for the relay
	Overage Nibblin       `json:"overage"`
	BWRule  BandWidthRule `json:"bandwidthRule"`
	// ContractTerm is the term in months
	ContractTerm int32 `json:"contractTerm"`
	// StartDate is the date the contract term starts
	StartDate time.Time `json:"startDate"`
	// EndDate is the date the contract term ends
	EndDate time.Time   `json:"endDate"`
	Type    MachineType `json:"machineType"`

	SignedID int64 `json:"signed_id"` // TODO: chopping block

	// SQL id (PK)
	DatabaseID int64

	// Simple text field for Ops to save data unique to each relay
	Notes string `json:"notes"`

	// Version, checked by fleet relays to see if they need to update
	Version string `json:"relay_version"`

	// Relay is prioritized when looking up near relays since it is
	// in the same datacenter as the destination datacenter.
	DestFirst bool `json:"destFirst"`

	// Relay can receive pings from any client via its internal address,
	// which is used for pinging near relays and if this relay is the first
	// hop in the route token. Other servers and relays should ping this
	// relay via external address (unless belong to same supplier of course).
	// IMPORTANT: relay must have InternalAddr field filled out
	InternalAddressClientRoutable bool `json:"internalAddressClientRoutable"`
}

func (r *Relay) EncodedPublicKey() string {
	return base64.StdEncoding.EncodeToString(r.PublicKey)
}

type Stats struct {
	RTT        float64 `json:"rtt"`
	Jitter     float64 `json:"jitter"`
	PacketLoss float64 `json:"packet_loss"`
}

func (s Stats) String() string {
	return fmt.Sprintf("RTT(%f) J(%f) PL(%f)", s.RTT, s.Jitter, s.PacketLoss)
}

func (s Stats) RedisString() string {
	return fmt.Sprintf("%.2f|%.2f|%.2f", s.RTT, s.Jitter, s.PacketLoss)
}

func (s *Stats) ParseRedisString(values []string) error {
	var index int
	var err error

	if s.RTT, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Stats] failed to read RTT from redis data: %v", err)
	}
	index++

	if s.Jitter, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Stats] failed to read jitter from redis data: %v", err)
	}
	index++

	if s.PacketLoss, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Stats] failed to read packet loss from redis data: %v", err)
	}
	index++

	return nil
}

type RelayPingData struct {
	ID      uint64 `json:"relay_id"`
	Address string `json:"relay_address"`
}

func RelayAddrs(relays []Relay) string {
	var b strings.Builder
	for _, relay := range relays {
		b.WriteString("{")
		b.WriteString(relay.Addr.String())
		b.WriteString("}")
	}
	return b.String()
}

func (r *Relay) String() string {
	relay := "\nrouting.Relay:\n"

	relay += "\tID                 				: " + fmt.Sprintf("%d", r.ID) + "\n"
	relay += "\tHex ID             				: " + fmt.Sprintf("%016x", r.ID) + "\n"
	relay += "\tName              		 		: " + r.Name + "\n"
	relay += "\tAddr              				: " + r.Addr.String() + "\n"
	relay += "\tInternalAddr       				: " + r.InternalAddr.String() + "\n"
	relay += "\tPublicKey          				: " + base64.StdEncoding.EncodeToString(r.PublicKey) + "\n"
	relay += "\tSeller             				: " + fmt.Sprintf("%d", r.Seller.DatabaseID) + "\n"
	relay += "\tBillingSupplier    				: " + fmt.Sprintf("%s", r.BillingSupplier) + "\n"
	relay += "\tDatacenter         				: " + fmt.Sprintf("%016x", r.Datacenter.ID) + "\n"
	relay += "\tNICSpeedMbps       				: " + fmt.Sprintf("%d", r.NICSpeedMbps) + "\n"
	relay += "\tIncludedBandwidthGB				: " + fmt.Sprintf("%d", r.IncludedBandwidthGB) + "\n"
	relay += "\tMaxBandwidthMbps   				: " + fmt.Sprintf("%d", r.MaxBandwidthMbps) + "\n"
	// relay += "\tLastUpdateTime     			: " + r.LastUpdateTime.String() + "\n"
	relay += "\tState              				: " + fmt.Sprintf("%v", r.State) + "\n"
	relay += "\tManagementAddr     				: " + r.ManagementAddr + "\n"
	relay += "\tSSHUser            				: " + r.SSHUser + "\n"
	relay += "\tSSHPort            				: " + fmt.Sprintf("%d", r.SSHPort) + "\n"
	relay += "\tMaxSessions        				: " + fmt.Sprintf("%d", r.MaxSessions) + "\n"
	// relay += "\tCPUUsage           			: " + fmt.Sprintf("%f", r.CPUUsage) + "\n"
	// relay += "\tMemUsage           			: " + fmt.Sprintf("%f", r.MemUsage) + "\n"
	relay += "\tEgressPriceOverride				: " + fmt.Sprintf("%v", r.EgressPriceOverride) + "\n"
	relay += "\tMRC                				: " + fmt.Sprintf("%v", r.MRC) + "\n"
	relay += "\tOverage            				: " + fmt.Sprintf("%v", r.Overage) + "\n"
	relay += "\tBWRule             				: " + fmt.Sprintf("%v", r.BWRule) + "\n"
	relay += "\tContractTerm       				: " + fmt.Sprintf("%d", r.ContractTerm) + "\n"
	relay += "\tStartDate          				: " + r.StartDate.String() + "\n"
	relay += "\tEndDate            				: " + r.EndDate.String() + "\n"
	relay += "\tType               				: " + fmt.Sprintf("%v", r.Type) + "\n"
	relay += "\tDatabaseID         				: " + fmt.Sprintf("%d", r.DatabaseID) + "\n"
	relay += "\tVersion            				: " + r.Version + "\n"
	relay += "\tDestFirst          				: " + fmt.Sprintf("%v", r.DestFirst) + "\n"
	relay += "\tInternalAddressClientRoutable	: " + fmt.Sprintf("%v", r.InternalAddressClientRoutable) + "\n"
	relay += "\tNotes:\n" + fmt.Sprintf("%v", r.Notes) + "\n"

	return relay
}
