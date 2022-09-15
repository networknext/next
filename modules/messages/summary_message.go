package messages

import (
	"time"

	"cloud.google.com/go/bigquery"
)

type SummaryMessage struct {
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
	Tags                            [BillingMessageMaxTags]uint64
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
	NearRelayIDs                    [BillingMessageMaxNearRelays]uint64
	NearRelayRTTs                   [BillingMessageMaxNearRelays]int32
	NearRelayJitters                [BillingMessageMaxNearRelays]int32
	NearRelayPacketLosses           [BillingMessageMaxNearRelays]int32
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
func (message *BillingMessage) GetSummary() *SummaryMessage {
	return &SummaryMessage{
		SessionID:                       message.SessionID,
		Summary:                         message.Summary,
		BuyerID:                         message.BuyerID,
		UserHash:                        message.UserHash,
		DatacenterID:                    message.DatacenterID,
		StartTimestamp:                  message.StartTimestamp,
		Latitude:                        message.Latitude,
		Longitude:                       message.Longitude,
		ISP:                             message.ISP,
		ConnectionType:                  message.ConnectionType,
		PlatformType:                    message.PlatformType,
		NumTags:                         message.NumTags,
		Tags:                            message.Tags,
		ABTest:                          message.ABTest,
		Pro:                             message.Pro,
		SDKVersion:                      message.SDKVersion,
		EnvelopeBytesUp:                 message.EnvelopeBytesUp,
		EnvelopeBytesDown:               message.EnvelopeBytesDown,
		ClientToServerPacketsSent:       message.ClientToServerPacketsSent,
		ServerToClientPacketsSent:       message.ServerToClientPacketsSent,
		ClientToServerPacketsLost:       message.ClientToServerPacketsLost,
		ServerToClientPacketsLost:       message.ServerToClientPacketsLost,
		ClientToServerPacketsOutOfOrder: message.ClientToServerPacketsOutOfOrder,
		ServerToClientPacketsOutOfOrder: message.ServerToClientPacketsOutOfOrder,
		NumNearRelays:                   message.NumNearRelays,
		NearRelayIDs:                    message.NearRelayIDs,
		NearRelayRTTs:                   message.NearRelayRTTs,
		NearRelayJitters:                message.NearRelayJitters,
		NearRelayPacketLosses:           message.NearRelayPacketLosses,
		EverOnNext:                      message.EverOnNext,
		SessionDuration:                 message.SessionDuration,
		TotalPriceSum:                   message.TotalPriceSum,
		EnvelopeBytesUpSum:              message.EnvelopeBytesUpSum,
		EnvelopeBytesDownSum:            message.EnvelopeBytesDownSum,
		DurationOnNext:                  message.DurationOnNext,
		ClientAddress:                   message.ClientAddress,
		ServerAddress:                   message.ServerAddress,
	}
}

// todo: WHY
func (message *SummaryMessage) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	/*
		1. Always

		These values are written for every slice.
	*/

	e["sessionID"] = int(message.SessionID)

	if message.Summary {

		/*
			2. First slice and summary slice only

			These values are serialized only for slice 0 and the summary slice.
		*/

		e["datacenterID"] = int(message.DatacenterID)
		e["buyerID"] = int(message.BuyerID)
		e["userHash"] = int(message.UserHash)
		e["envelopeBytesUp"] = int(message.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(message.EnvelopeBytesDown)
		e["latitude"] = message.Latitude
		e["longitude"] = message.Longitude
		e["clientAddress"] = message.ClientAddress
		e["serverAddress"] = message.ServerAddress
		e["isp"] = message.ISP
		e["connectionType"] = int(message.ConnectionType)
		e["platformType"] = int(message.PlatformType)
		e["sdkVersion"] = message.SDKVersion

		if message.NumTags > 0 {
			tags := make([]bigquery.Value, message.NumTags)
			for i := 0; i < int(message.NumTags); i++ {
				tags[i] = int(message.Tags[i])
			}
			e["tags"] = tags
		}

		if message.ABTest {
			e["abTest"] = true
		}

		if message.Pro {
			e["pro"] = true
		}

		/*
			3. Summary slice only

			These values are serialized only for the summary slice (at the end of the session)
		*/

		e["clientToServerPacketsSent"] = int(message.ClientToServerPacketsSent)
		e["serverToClientPacketsSent"] = int(message.ServerToClientPacketsSent)
		e["clientToServerPacketsLost"] = int(message.ClientToServerPacketsLost)
		e["serverToClientPacketsLost"] = int(message.ServerToClientPacketsLost)
		e["clientToServerPacketsOutOfOrder"] = int(message.ClientToServerPacketsOutOfOrder)
		e["serverToClientPacketsOutOfOrder"] = int(message.ServerToClientPacketsOutOfOrder)

		if message.NumNearRelays > 0 {

			nearRelayIDs := make([]bigquery.Value, message.NumNearRelays)
			nearRelayRTTs := make([]bigquery.Value, message.NumNearRelays)
			nearRelayJitters := make([]bigquery.Value, message.NumNearRelays)
			nearRelayPacketLosses := make([]bigquery.Value, message.NumNearRelays)

			for i := 0; i < int(message.NumNearRelays); i++ {
				nearRelayIDs[i] = int(message.NearRelayIDs[i])
				nearRelayRTTs[i] = int(message.NearRelayRTTs[i])
				nearRelayJitters[i] = int(message.NearRelayJitters[i])
				nearRelayPacketLosses[i] = int(message.NearRelayPacketLosses[i])
			}

			e["nearRelayIDs"] = nearRelayIDs
			e["nearRelayRTTs"] = nearRelayRTTs
			e["nearRelayJitters"] = nearRelayJitters
			e["nearRelayPacketLosses"] = nearRelayPacketLosses

		}

		if message.EverOnNext {
			e["everOnNext"] = message.EverOnNext
		}

		e["sessionDuration"] = int(message.SessionDuration)

		if message.EverOnNext {
			e["totalPriceSum"] = int(message.TotalPriceSum)
			e["envelopeBytesUpSum"] = int(message.EnvelopeBytesUpSum)
			e["envelopeBytesDownSum"] = int(message.EnvelopeBytesDownSum)
			e["durationOnNext"] = int(message.DurationOnNext)
		}

		if message.StartTimestamp == 0 {
			// In case startTimestamp is 0 during transition
			e["startTimestamp"] = int(time.Now().Add(time.Duration(message.SessionDuration) * time.Second * -1).Unix())
		} else {
			e["startTimestamp"] = int(message.StartTimestamp)
		}
	}

	return e, "", nil
}
