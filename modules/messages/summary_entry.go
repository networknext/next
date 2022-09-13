package messages

import (
	"time"

	"cloud.google.com/go/bigquery"
)

type SummaryEntry struct {
	SessionID                       uint64
	Summary                         bool
	BuyerID                         uint64
	UserHash                        uint64
	DatacenterID                    uint64
	StartTimestamp                  uint32
	Latitude                        float32
	Longitude                       float32
	ISP                             string
	ConnectionType                  int32
	PlatformType                    int32
	NumTags                         int32
	Tags                            [BillingEntryMaxTags]uint64
	ABTest                          bool
	Pro                             bool
	SDKVersion                      string
	EnvelopeBytesUp                 uint64
	EnvelopeBytesDown               uint64
	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	ClientToServerPacketsOutOfOrder uint64
	ServerToClientPacketsOutOfOrder uint64
	NumNearRelays                   int32
	NearRelayIDs                    [BillingEntryMaxNearRelays]uint64
	NearRelayRTTs                   [BillingEntryMaxNearRelays]int32
	NearRelayJitters                [BillingEntryMaxNearRelays]int32
	NearRelayPacketLosses           [BillingEntryMaxNearRelays]int32
	EverOnNext                      bool
	SessionDuration                 uint32
	TotalPriceSum                   uint64
	EnvelopeBytesUpSum              uint64
	EnvelopeBytesDownSum            uint64
	DurationOnNext                  uint32
	ClientAddress                   string
	ServerAddress                   string
}

// todo: WHY
func (entry *BillingEntry) GetSummary() *SummaryEntry {
	return &SummaryEntry{
		SessionID:                       entry.SessionID,
		Summary:                         entry.Summary,
		BuyerID:                         entry.BuyerID,
		UserHash:                        entry.UserHash,
		DatacenterID:                    entry.DatacenterID,
		StartTimestamp:                  entry.StartTimestamp,
		Latitude:                        entry.Latitude,
		Longitude:                       entry.Longitude,
		ISP:                             entry.ISP,
		ConnectionType:                  entry.ConnectionType,
		PlatformType:                    entry.PlatformType,
		NumTags:                         entry.NumTags,
		Tags:                            entry.Tags,
		ABTest:                          entry.ABTest,
		Pro:                             entry.Pro,
		SDKVersion:                      entry.SDKVersion,
		EnvelopeBytesUp:                 entry.EnvelopeBytesUp,
		EnvelopeBytesDown:               entry.EnvelopeBytesDown,
		ClientToServerPacketsSent:       entry.ClientToServerPacketsSent,
		ServerToClientPacketsSent:       entry.ServerToClientPacketsSent,
		ClientToServerPacketsLost:       entry.ClientToServerPacketsLost,
		ServerToClientPacketsLost:       entry.ServerToClientPacketsLost,
		ClientToServerPacketsOutOfOrder: entry.ClientToServerPacketsOutOfOrder,
		ServerToClientPacketsOutOfOrder: entry.ServerToClientPacketsOutOfOrder,
		NumNearRelays:                   entry.NumNearRelays,
		NearRelayIDs:                    entry.NearRelayIDs,
		NearRelayRTTs:                   entry.NearRelayRTTs,
		NearRelayJitters:                entry.NearRelayJitters,
		NearRelayPacketLosses:           entry.NearRelayPacketLosses,
		EverOnNext:                      entry.EverOnNext,
		SessionDuration:                 entry.SessionDuration,
		TotalPriceSum:                   entry.TotalPriceSum,
		EnvelopeBytesUpSum:              entry.EnvelopeBytesUpSum,
		EnvelopeBytesDownSum:            entry.EnvelopeBytesDownSum,
		DurationOnNext:                  entry.DurationOnNext,
		ClientAddress:                   entry.ClientAddress,
		ServerAddress:                   entry.ServerAddress,
	}
}

// todo: WHY
func (entry *SummaryEntry) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	/*
		1. Always

		These values are written for every slice.
	*/

	e["sessionID"] = int(entry.SessionID)

	if entry.Summary {

		/*
			2. First slice and summary slice only

			These values are serialized only for slice 0 and the summary slice.
		*/

		e["datacenterID"] = int(entry.DatacenterID)
		e["buyerID"] = int(entry.BuyerID)
		e["userHash"] = int(entry.UserHash)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)
		e["latitude"] = entry.Latitude
		e["longitude"] = entry.Longitude
		e["clientAddress"] = entry.ClientAddress
		e["serverAddress"] = entry.ServerAddress
		e["isp"] = entry.ISP
		e["connectionType"] = int(entry.ConnectionType)
		e["platformType"] = int(entry.PlatformType)
		e["sdkVersion"] = entry.SDKVersion

		if entry.NumTags > 0 {
			tags := make([]bigquery.Value, entry.NumTags)
			for i := 0; i < int(entry.NumTags); i++ {
				tags[i] = int(entry.Tags[i])
			}
			e["tags"] = tags
		}

		if entry.ABTest {
			e["abTest"] = true
		}

		if entry.Pro {
			e["pro"] = true
		}

		/*
			3. Summary slice only

			These values are serialized only for the summary slice (at the end of the session)
		*/

		e["clientToServerPacketsSent"] = int(entry.ClientToServerPacketsSent)
		e["serverToClientPacketsSent"] = int(entry.ServerToClientPacketsSent)
		e["clientToServerPacketsLost"] = int(entry.ClientToServerPacketsLost)
		e["serverToClientPacketsLost"] = int(entry.ServerToClientPacketsLost)
		e["clientToServerPacketsOutOfOrder"] = int(entry.ClientToServerPacketsOutOfOrder)
		e["serverToClientPacketsOutOfOrder"] = int(entry.ServerToClientPacketsOutOfOrder)

		if entry.NumNearRelays > 0 {

			nearRelayIDs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayRTTs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayJitters := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayPacketLosses := make([]bigquery.Value, entry.NumNearRelays)

			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayIDs[i] = int(entry.NearRelayIDs[i])
				nearRelayRTTs[i] = int(entry.NearRelayRTTs[i])
				nearRelayJitters[i] = int(entry.NearRelayJitters[i])
				nearRelayPacketLosses[i] = int(entry.NearRelayPacketLosses[i])
			}

			e["nearRelayIDs"] = nearRelayIDs
			e["nearRelayRTTs"] = nearRelayRTTs
			e["nearRelayJitters"] = nearRelayJitters
			e["nearRelayPacketLosses"] = nearRelayPacketLosses

		}

		if entry.EverOnNext {
			e["everOnNext"] = entry.EverOnNext
		}

		e["sessionDuration"] = int(entry.SessionDuration)

		if entry.EverOnNext {
			e["totalPriceSum"] = int(entry.TotalPriceSum)
			e["envelopeBytesUpSum"] = int(entry.EnvelopeBytesUpSum)
			e["envelopeBytesDownSum"] = int(entry.EnvelopeBytesDownSum)
			e["durationOnNext"] = int(entry.DurationOnNext)
		}

		if entry.StartTimestamp == 0 {
			// In case startTimestamp is 0 during transition
			e["startTimestamp"] = int(time.Now().Add(time.Duration(entry.SessionDuration) * time.Second * -1).Unix())
		} else {
			e["startTimestamp"] = int(entry.StartTimestamp)
		}
	}

	return e, "", nil
}
