package routing_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestRouteMatrix(t *testing.T) {
	t.Run("RouteMatrix", func(t *testing.T) {
		t.Run("UnmarshalBinary()", func(t *testing.T) {
			unmarshalAssertionsVer0 := func(t *testing.T, matrix *routing.RouteMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, entries []routing.RouteMatrixEntry) {
				assert.Len(t, matrix.RelayIDs, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.Entries, len(entries))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIDs, id&0xFFFFFFFF)
				}

				for _, addr := range relayAddrs {
					tmp := make([]byte, len(addr))
					copy(tmp, addr)
					assert.Contains(t, matrix.RelayAddresses, tmp)
				}

				for _, pk := range publicKeys {
					assert.Contains(t, matrix.RelayPublicKeys, pk)
				}

				for i := 0; i < numDatacenters; i++ {
					assert.Contains(t, matrix.DatacenterRelays, datacenters[i]&0xFFFFFFFF)

					relays := matrix.DatacenterRelays[datacenters[i]]
					for j := 0; j < len(datacenterRelays[i]); j++ {
						assert.Contains(t, relays, datacenterRelays[i][j]&0xFFFFFFFF)
					}
				}

				for i, expected := range entries {
					actual := matrix.Entries[i]

					assert.Equal(t, expected.DirectRTT, actual.DirectRTT)
					assert.Equal(t, expected.NumRoutes, actual.NumRoutes)
					assert.Equal(t, expected.RouteRTT, actual.RouteRTT)
					assert.Equal(t, expected.RouteNumRelays, actual.RouteNumRelays)

					for i, ids := range expected.RouteRelays {
						for j, id := range ids {
							assert.Equal(t, id&0xFFFFFFFF, actual.RouteRelays[i][j])
						}
					}
				}
			}

			unmarshalAssertionsVer1 := func(t *testing.T, matrix *routing.RouteMatrix, relayNames []string) {
				assert.Len(t, matrix.RelayNames, len(relayNames))
				assert.Len(t, matrix.RelayIDs, len(relayNames))
				for _, name := range relayNames {
					assert.Contains(t, matrix.RelayNames, name)
				}
			}

			unmarshalAssertionsVer2 := func(t *testing.T, matrix *routing.RouteMatrix, datacenterIDs []uint64, datacenterNames []string) {
				assert.Len(t, matrix.DatacenterIDs, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))
				assert.Len(t, matrix.DatacenterIDs, len(matrix.DatacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIDs, id&0xFFFFFFFF)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			unmarshalAssertionsVer3 := func(t *testing.T, matrix *routing.RouteMatrix, numRelays, numDatacenters int, relayIDs, datacenters []uint64, relayAddrs []string, datacenterRelays [][]uint64, publicKeys [][]byte, entries []routing.RouteMatrixEntry, relayNames []string, datacenterIDs []uint64, datacenterNames []string) {
				assert.Len(t, matrix.RelayIDs, numRelays)
				assert.Len(t, matrix.RelayAddresses, numRelays)
				assert.Len(t, matrix.RelayPublicKeys, numRelays)
				assert.Len(t, matrix.DatacenterRelays, numDatacenters)
				assert.Len(t, matrix.Entries, len(entries))

				for _, id := range relayIDs {
					assert.Contains(t, matrix.RelayIDs, id)
				}

				for _, addr := range relayAddrs {
					tmp := make([]byte, routing.MaxRelayAddressLength)
					copy(tmp, addr)
					assert.Contains(t, matrix.RelayAddresses, tmp)
				}

				for _, pk := range publicKeys {
					assert.Contains(t, matrix.RelayPublicKeys, pk)
				}

				for i := 0; i < numDatacenters; i++ {
					assert.Contains(t, matrix.DatacenterRelays, datacenters[i])

					relays := matrix.DatacenterRelays[datacenters[i]]
					for j := 0; j < len(datacenterRelays[i]); j++ {
						assert.Contains(t, relays, datacenterRelays[i][j])
					}
				}

				for i, expected := range entries {
					actual := matrix.Entries[i]

					assert.Equal(t, expected.DirectRTT, actual.DirectRTT)
					assert.Equal(t, expected.NumRoutes, actual.NumRoutes)
					assert.Equal(t, expected.RouteRTT, actual.RouteRTT)
					assert.Equal(t, expected.RouteNumRelays, actual.RouteNumRelays)

					for i, ids := range expected.RouteRelays {
						for j, id := range ids {
							assert.Equal(t, id, actual.RouteRelays[i][j])
						}
					}
				}

				unmarshalAssertionsVer1(t, matrix, relayNames)

				assert.Len(t, matrix.DatacenterIDs, len(datacenterIDs))
				assert.Len(t, matrix.DatacenterNames, len(datacenterNames))

				for _, id := range datacenterIDs {
					assert.Contains(t, matrix.DatacenterIDs, id)
				}

				for _, name := range datacenterNames {
					assert.Contains(t, matrix.DatacenterNames, name)
				}
			}

			unmarshalAssertionsVer4 := func(t *testing.T, matrix *routing.RouteMatrix, sellers []routing.Seller) {
				assert.Len(t, matrix.RelaySellers, len(sellers))
				for i, seller := range sellers {
					assert.Equal(t, matrix.RelaySellers[i].ID, seller.ID)
					assert.Equal(t, matrix.RelaySellers[i].Name, seller.Name)
					assert.Equal(t, matrix.RelaySellers[i].IngressPriceCents, seller.IngressPriceCents)
					assert.Equal(t, matrix.RelaySellers[i].EgressPriceCents, seller.EgressPriceCents)
				}
			}

			t.Run("version number 0", func(t *testing.T) {
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}

				relayIDs := addrsToIDs(relayAddrs)

				numRelays := len(relayAddrs)

				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}

				datacenters := []uint64{0, 1, 2, 3, 4}

				numDatacenters := len(datacenters)

				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}

				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				// the size of each route entry
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries)
			})

			t.Run("version number 1", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				numDatacenters := len(datacenters)
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 1)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames) //version >= 1
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
			})

			t.Run("version number 2", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				numDatacenters := len(datacenters)
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 2)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                           // version 1
				putDatacenterStuffOld(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer0(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries)
				unmarshalAssertionsVer1(t, &matrix, relayNames)
				unmarshalAssertionsVer2(t, &matrix, datacenters, datacenterNames)
			})

			t.Run("version number 3", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				numDatacenters := len(datacenters)
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddress(relayAddrs)
				buffSize += sizeofRelayPublicKeys(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRouteMatrixEntry(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 3)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putEntries(buff, &offset, entries)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer3(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries, relayNames, datacenters, datacenterNames)
			})

			t.Run("version number 4", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				numDatacenters := len(datacenters)
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

				// version 4 stuff
				sellers := []routing.Seller{
					routing.Seller{ID: "id0", Name: "name0", IngressPriceCents: 1, EgressPriceCents: 2},
					routing.Seller{ID: "id1", Name: "name1", IngressPriceCents: 3, EgressPriceCents: 4},
					routing.Seller{ID: "id2", Name: "name2", IngressPriceCents: 5, EgressPriceCents: 6},
					routing.Seller{ID: "id3", Name: "name3", IngressPriceCents: 7, EgressPriceCents: 8},
					routing.Seller{ID: "id4", Name: "name4", IngressPriceCents: 9, EgressPriceCents: 10},
				}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddress(relayAddrs)
				buffSize += sizeofRelayPublicKeys(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRouteMatrixEntry(entries)
				buffSize += sizeofSellers(sellers)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 4)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putEntries(buff, &offset, entries)
				putSellers(buff, &offset, sellers)

				var matrix routing.RouteMatrix
				err := matrix.UnmarshalBinary(buff)
				assert.Nil(t, err)
				unmarshalAssertionsVer3(t, &matrix, numRelays, numDatacenters, relayIDs, datacenters, relayAddrs, datacenterRelays, publicKeys, entries, relayNames, datacenters, datacenterNames)
				unmarshalAssertionsVer4(t, &matrix, sellers)
			})

			t.Run("Error cases - v0", func(t *testing.T) {
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}

				relayIDs := addrsToIDs(relayAddrs)

				numRelays := len(relayAddrs)

				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}

				datacenters := []uint64{0, 1, 2, 3, 4}

				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}

				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				// the size of each route entry
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 0)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 6)
					var matrix routing.RouteMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown route matrix version: 6")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver < 3")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver < 3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver < 3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver < 3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver < 3")
				})

				t.Run("Invalid matrix entry read", func(t *testing.T) {
					t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
					})

					t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
					})

					t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 32 + 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
					})

					t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 32 + 32 + 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
					})

					t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := sizeofRouteMatrixEntryOld(entries) + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver < 3")
					})
				})
			})

			t.Run("Error cases - v1", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 1)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames) //version >= 1
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 6)
					var matrix routing.RouteMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown route matrix version: 6")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver < 3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver < 3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver < 3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver < 3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver < 3")
				})

				t.Run("Invalid matrix entry read", func(t *testing.T) {
					t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
					})

					t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
					})

					t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 32 + 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
					})

					t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 32 + 32 + 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
					})

					t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := sizeofRouteMatrixEntryOld(entries) + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver < 3")
					})
				})
			})

			t.Run("Error cases - v2", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddressOld(relayAddrs)
				buffSize += sizeofRelayPublicKeysOld(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs32(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs32(relayIDs)
				buffSize += sizeofRouteMatrixEntryOld(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 2)
				putRelayIDsOld(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                           // version 1
				putDatacenterStuffOld(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddressesOld(buff, &offset, relayAddrs)
				putRelayPublicKeysOld(buff, &offset, publicKeys)
				putDatacentersOld(buff, &offset, datacenters, datacenterRelays)
				putEntriesOld(buff, &offset, entries)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 6)
					var matrix routing.RouteMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown route matrix version: 6")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver < 3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay names")
				})

				t.Run("Invalid datacenter count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter count")
				})

				t.Run("Invalid datacenter id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter ids - ver < 3")
				})

				t.Run("Invalid datacenter name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver < 3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver < 3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver < 3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 4 + 4 + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver < 3")
				})

				t.Run("Invalid matrix entry read", func(t *testing.T) {
					t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
					})

					t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
					})

					t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 32 + 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
					})

					t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 32 + 32 + 4 + 4 + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
					})

					t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := sizeofRouteMatrixEntryOld(entries) + sizeofRelayIDs32(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs32(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeysOld(publicKeys) + sizeofRelayAddressOld(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs32(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs32(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver < 3")
					})
				})
			})

			t.Run("Error cases - v3", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddress(relayAddrs)
				buffSize += sizeofRelayPublicKeys(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRouteMatrixEntry(entries)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 3)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putEntries(buff, &offset, entries)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 6)
					var matrix routing.RouteMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown route matrix version: 6")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver >= v3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay names")
				})

				t.Run("Invalid datacenter count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter count")
				})

				t.Run("Invalid datacenter id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter ids - ver >= v3")
				})

				t.Run("Invalid datacenter name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDatacenterNames(datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver >= v3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver >= v3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver >= v3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver >= v3")
				})

				t.Run("Invalid matrix entry read", func(t *testing.T) {
					t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
					})

					t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
					})

					t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
					})

					t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + 4 + 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
					})

					t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := sizeofRouteMatrixEntry(entries) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver >= v3")
					})
				})
			})

			t.Run("Error cases - v4", func(t *testing.T) {
				// version 0 stuff
				relayAddrs := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
				relayIDs := addrsToIDs(relayAddrs)
				numRelays := len(relayAddrs)
				publicKeys := [][]byte{
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
					RandomPublicKey(),
				}
				datacenters := []uint64{0, 1, 2, 3, 4}
				datacenterRelays := [][]uint64{{relayIDs[0]}, {relayIDs[1]}, {relayIDs[2]}, {relayIDs[3]}, {relayIDs[4]}}
				numEntries := routing.TriMatrixLength(numRelays)
				entries := make([]routing.RouteMatrixEntry, numEntries)
				generateRouteMatrixEntries(entries)

				// version 1 stuff
				relayNames := []string{"a name", "another name", "oh boy another", "they just keep coming", "i'm out of sarcasm"}

				// version 2 stuff
				// resusing datacenters for the ID array
				datacenterNames := []string{"a datacenter", "another datacenter", "third", "fourth", "fifth"}

				// version 4 stuff
				sellers := []routing.Seller{
					routing.Seller{ID: "id0", Name: "name0", IngressPriceCents: 1, EgressPriceCents: 2},
					routing.Seller{ID: "id1", Name: "name1", IngressPriceCents: 3, EgressPriceCents: 4},
					routing.Seller{ID: "id2", Name: "name2", IngressPriceCents: 5, EgressPriceCents: 6},
					routing.Seller{ID: "id3", Name: "name3", IngressPriceCents: 7, EgressPriceCents: 8},
					routing.Seller{ID: "id4", Name: "name4", IngressPriceCents: 9, EgressPriceCents: 10},
				}

				buffSize := 0
				buffSize += sizeofVersionNumber()
				buffSize += sizeofRelayCount()
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRelayNames(relayNames)
				buffSize += sizeofDatacenterCount()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofDatacenterNames(datacenterNames)
				buffSize += sizeofRelayAddress(relayAddrs)
				buffSize += sizeofRelayPublicKeys(publicKeys)
				buffSize += sizeofDataCenterCount2()
				buffSize += sizeofDatacenterIDs64(datacenters)
				buffSize += sizeofRelaysInDatacenterCount(datacenters)
				buffSize += sizeofRelayIDs64(relayIDs)
				buffSize += sizeofRouteMatrixEntry(entries)
				buffSize += sizeofSellers(sellers)

				buff := make([]byte, buffSize)

				offset := 0
				putVersionNumber(buff, &offset, 4)
				putRelayIDs(buff, &offset, addrsToIDs(relayAddrs))
				putRelayNames(buff, &offset, relayNames)                        // version 1
				putDatacenterStuff(buff, &offset, datacenters, datacenterNames) // version 2
				putRelayAddresses(buff, &offset, relayAddrs)
				putRelayPublicKeys(buff, &offset, publicKeys)
				putDatacenters(buff, &offset, datacenters, datacenterRelays)
				putEntries(buff, &offset, entries)
				putSellers(buff, &offset, sellers)

				t.Run("version of incoming bin data too high", func(t *testing.T) {
					buff := make([]byte, 4)
					offset := 0
					putVersionNumber(buff, &offset, 6)
					var matrix routing.RouteMatrix

					err := matrix.UnmarshalBinary(buff)

					assert.EqualError(t, err, "unknown route matrix version: 6")
				})

				t.Run("Invalid version read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at version number")
				})

				t.Run("Invalid relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays")
				})

				t.Run("Invalid relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids - ver >= v3")
				})

				t.Run("Invalid relay name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay names")
				})

				t.Run("Invalid datacenter count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter count")
				})

				t.Run("Invalid datacenter id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 8 + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter ids - ver >= v3")
				})

				t.Run("Invalid datacenter name read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDatacenterNames(datacenterNames) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter names")
				})

				t.Run("Invalid relay address read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay addresses - ver >= v3")
				})

				t.Run("Invalid relay public key read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay public keys - ver >= v3")
				})

				t.Run("Invalid datacenter count read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of datacenters (second time)")
				})

				t.Run("Invalid datacenter id read second time", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at datacenter id - ver >= v3")
				})

				t.Run("Invalid datacenter relay count read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in datacenter")
				})

				t.Run("Invalid datacenter relay id read", func(t *testing.T) {
					var matrix routing.RouteMatrix
					offset := 8 + 4 + 8 + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
					err := matrix.UnmarshalBinary(buff[:offset])
					assert.EqualError(t, err, "[RouteMatrix] invalid read at relay ids for datacenter - ver >= v3")
				})

				t.Run("Invalid matrix entry read", func(t *testing.T) {
					t.Run("Invalid direct route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at direct rtt")
					})

					t.Run("Invalid route count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of routes")
					})

					t.Run("Invalid route RTT read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at route rtt")
					})

					t.Run("Invalid relay count read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + 4 + 4 + 4 + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at number of relays in route")
					})

					t.Run("Invalid relay read in matrix entry", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := sizeofRouteMatrixEntry(entries) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read at relays in route - ver >= v3")
					})
				})

				t.Run("Invalid seller read", func(t *testing.T) {
					t.Run("Invalid seller ID read", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + len(sellers[0].ID) + sizeofRouteMatrixEntry(entries) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller ID")
					})

					t.Run("Invalid seller name read", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 4 + len(sellers[0].Name) + 4 + len(sellers[0].ID) + sizeofRouteMatrixEntry(entries) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller name")
					})

					t.Run("Invalid seller ingress price read", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 8 + 4 + len(sellers[0].Name) + 4 + len(sellers[0].ID) + sizeofRouteMatrixEntry(entries) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller ingress price")
					})

					t.Run("Invalid seller egress price read", func(t *testing.T) {
						var matrix routing.RouteMatrix
						offset := 8 + 8 + 4 + len(sellers[0].Name) + 4 + len(sellers[0].ID) + sizeofRouteMatrixEntry(entries) + sizeofRelayIDs64(relayIDs) + sizeofRelaysInDatacenterCount(datacenters) + sizeofDatacenterIDs64(datacenters) + sizeofDataCenterCount2() + sizeofRelayPublicKeys(publicKeys) + sizeofRelayAddress(relayAddrs) + sizeofDatacenterNames(datacenterNames) + sizeofDatacenterIDs64(datacenters) + sizeofDatacenterCount() + sizeofRelayNames(relayNames) + sizeofRelayIDs64(relayIDs) + sizeofRelayCount() + sizeofVersionNumber() - 1
						err := matrix.UnmarshalBinary(buff[:offset])
						assert.EqualError(t, err, "[RouteMatrix] invalid read on relay seller egress price")
					})
				})
			})
		})

		t.Run("MarshalBinary()", func(t *testing.T) {
			t.Run("MarshalBinary -> UnmarshalBinary equality", func(t *testing.T) {
				matrix := getPopulatedRouteMatrix(false)

				var other routing.RouteMatrix

				bin, err := matrix.MarshalBinary()

				// essentialy this asserts the result of MarshalBinary(),
				// if Unmarshal tests pass then the binary data from Marshal
				// is valid if unmarshaling equals the original
				other.UnmarshalBinary(bin)

				assert.Nil(t, err)
				assert.Equal(t, matrix, &other)
			})

			t.Run("Relay ID and name buffers different sizes", func(t *testing.T) {
				var matrix routing.RouteMatrix

				matrix.RelayIDs = make([]uint64, 2)
				matrix.RelayIDs[0] = 123
				matrix.RelayIDs[1] = 456

				matrix.RelayNames = make([]string, 1) // Only 1 name but 2 IDs
				matrix.RelayNames[0] = "first"

				_, err := matrix.MarshalBinary()
				errorString := fmt.Errorf("length of Relay IDs not equal to length of Relay Names: %d != %d", len(matrix.RelayIDs), len(matrix.RelayNames))
				assert.EqualError(t, err, errorString.Error())
			})

			t.Run("Datacenter ID and name buffers different sizes", func(t *testing.T) {
				var matrix routing.RouteMatrix

				matrix.DatacenterIDs = make([]uint64, 2)
				matrix.DatacenterIDs[0] = 999
				matrix.DatacenterIDs[1] = 111

				matrix.DatacenterNames = make([]string, 1) // Only 1 name but 2 IDs
				matrix.DatacenterNames[0] = "a name"

				_, err := matrix.MarshalBinary()
				errorString := fmt.Errorf("length of Datacenter IDs not equal to length of Datacenter Names: %d != %d", len(matrix.DatacenterIDs), len(matrix.DatacenterNames))
				assert.EqualError(t, err, errorString.Error())
			})
		})

		t.Run("ServeHTTP()", func(t *testing.T) {
			t.Run("Failure to serve HTTP", func(t *testing.T) {
				// Create and populate a malformed route matrix
				matrix := getPopulatedRouteMatrix(true)

				// Create a dummy http request to test ServeHTTP
				recorder := httptest.NewRecorder()
				request, err := http.NewRequest("GET", "/", nil)
				assert.NoError(t, err)

				matrix.ServeHTTP(recorder, request)

				// Get the response
				response := recorder.Result()

				assert.Equal(t, 500, response.StatusCode)
			})

			t.Run("Successful Serve", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				// Create a dummy http request to test ServeHTTP
				recorder := httptest.NewRecorder()
				request, err := http.NewRequest("GET", "/", nil)
				assert.NoError(t, err)

				matrix.ServeHTTP(recorder, request)

				// Get the response
				response := recorder.Result()

				// Read the response body
				body, err := ioutil.ReadAll(response.Body)
				assert.NoError(t, err)
				response.Body.Close()

				// Create a new matrix to store the response
				var receivedMatrix routing.RouteMatrix
				err = receivedMatrix.UnmarshalBinary(body)
				assert.NoError(t, err)

				// Validate the response
				assert.Equal(t, "application/octet-stream", response.Header.Get("Content-Type"))
				assert.Equal(t, matrix, &receivedMatrix)
			})
		})

		t.Run("WriteTo()", func(t *testing.T) {
			t.Run("Error during MarshalBinary()", func(t *testing.T) {
				// Create and populate a malformed route matrix
				matrix := getPopulatedRouteMatrix(true)

				var buff bytes.Buffer
				_, err := matrix.WriteTo(&buff)
				assert.EqualError(t, err, fmt.Sprintf("length of Relay IDs not equal to length of Relay Names: %v != %v", len(matrix.RelayIDs), len(matrix.RelayNames)))
			})

			t.Run("Error during write", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				var buff ErrorBuffer
				_, err := matrix.WriteTo(&buff)
				assert.Error(t, err)
			})

			t.Run("Success", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				var buff bytes.Buffer
				_, err := matrix.WriteTo(&buff)
				assert.NoError(t, err)
			})
		})

		t.Run("ReadFrom()", func(t *testing.T) {
			t.Run("Nil reader", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				// Try to read from nil reader
				_, err := matrix.ReadFrom(nil)
				assert.EqualError(t, err, "reader is nil")
			})

			t.Run("Error during read", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				// Try to read into the ErrorBuffer
				var buff ErrorBuffer
				_, err := matrix.ReadFrom(&buff)
				assert.Error(t, err)
			})

			t.Run("Error during UnmarshalBinary()", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				// Marshal the route matrix, modify it, then attempt to unmarshal it
				buff, err := matrix.MarshalBinary()
				assert.NoError(t, err)

				buffSlice := buff[:3] // Only send the first 3 bytes so that the version read fails and throws an error

				_, err = matrix.ReadFrom(bytes.NewBuffer(buffSlice))
				assert.Error(t, err)
			})

			t.Run("Success", func(t *testing.T) {
				// Create and populate a route matrix
				matrix := getPopulatedRouteMatrix(false)

				// Marshal the route matrix so we can read it in
				buff, err := matrix.MarshalBinary()
				assert.NoError(t, err)

				// Read into a byte buffer
				_, err = matrix.ReadFrom(bytes.NewBuffer(buff))
				assert.NoError(t, err)
			})
		})
	})

	t.Run("Old tests from core/core_test.go", func(t *testing.T) {
		analyze := func(t *testing.T, route_matrix *routing.RouteMatrix) {
			src := route_matrix.RelayIDs
			dest := route_matrix.RelayIDs

			numRelayPairs := 0
			numValidRelayPairs := 0
			numValidRelayPairsWithoutImprovement := 0

			buckets := make([]int, 11)

			for i := range src {
				for j := range dest {
					if j < i {
						numRelayPairs++
						abFlatIndex := routing.TriMatrixIndex(i, j)
						if len(route_matrix.Entries[abFlatIndex].RouteRTT) > 0 {
							numValidRelayPairs++
							improvement := route_matrix.Entries[abFlatIndex].DirectRTT - route_matrix.Entries[abFlatIndex].RouteRTT[0]
							if improvement > 0.0 {
								if improvement <= 5 {
									buckets[0]++
								} else if improvement <= 10 {
									buckets[1]++
								} else if improvement <= 15 {
									buckets[2]++
								} else if improvement <= 20 {
									buckets[3]++
								} else if improvement <= 25 {
									buckets[4]++
								} else if improvement <= 30 {
									buckets[5]++
								} else if improvement <= 35 {
									buckets[6]++
								} else if improvement <= 40 {
									buckets[7]++
								} else if improvement <= 45 {
									buckets[8]++
								} else if improvement <= 50 {
									buckets[9]++
								} else {
									buckets[10]++
								}
							} else {
								numValidRelayPairsWithoutImprovement++
							}
						}
					}
				}
			}

			assert.Equal(t, 43916, numValidRelayPairsWithoutImprovement, "optimizer is broken")

			expected := []int{2561, 8443, 6531, 4690, 3208, 2336, 1775, 1364, 1078, 749, 5159}

			assert.Equal(t, expected, buckets, "optimizer is broken")
		}

		t.Run("TestRouteMatrixSanity() - test using version 2 example data", func(t *testing.T) {
			var cmatrix routing.CostMatrix
			var rmatrix routing.RouteMatrix

			raw, err := ioutil.ReadFile("test_data/cost-for-sanity-check.bin")
			assert.Nil(t, err)

			err = cmatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			err = cmatrix.Optimize(&rmatrix, 1.0)
			assert.Nil(t, err)

			src := rmatrix.RelayIDs
			dest := rmatrix.RelayIDs

			for i := range src {
				for j := range dest {
					if j < i {
						ijFlatIndex := routing.TriMatrixIndex(i, j)

						entries := rmatrix.Entries[ijFlatIndex]
						for k := 0; k < int(entries.NumRoutes); k++ {
							numRelays := entries.RouteNumRelays[k]
							firstRelay := entries.RouteRelays[k][0]
							lastRelay := entries.RouteRelays[k][numRelays-1]

							assert.Equal(t, src[firstRelay], dest[i], "invalid route entry #%d at (%d,%d), near relay %d (idx %d) != %d (idx %d)\n", k, i, j, src[firstRelay], firstRelay, dest[i], i)
							assert.Equal(t, src[lastRelay], dest[j], "invalid route entry #%d at (%d,%d), dest relay %d (idx %d) != %d (idx %d)\n", k, i, j, src[lastRelay], lastRelay, dest[j], j)
						}
					}
				}
			}
		})

		t.Run("TestRouteMatrix() - another test with different version 0 sample data", func(t *testing.T) {
			raw, err := ioutil.ReadFile("test_data/cost.bin")
			assert.Nil(t, err)
			assert.Equal(t, len(raw), 355188, "cost.bin should be 355188 bytes")

			var costMatrix routing.CostMatrix
			err = costMatrix.UnmarshalBinary(raw)
			assert.Nil(t, err)

			costMatrixData, err := costMatrix.MarshalBinary()
			assert.Nil(t, err)

			var readCostMatrix routing.CostMatrix
			err = readCostMatrix.UnmarshalBinary(costMatrixData)
			assert.Nil(t, err)

			var routeMatrix routing.RouteMatrix
			costMatrix.Optimize(&routeMatrix, 5)
			assert.NotNil(t, &routeMatrix)
			assert.Equal(t, costMatrix.RelayIDs, routeMatrix.RelayIDs, "relay id mismatch")
			assert.Equal(t, costMatrix.RelayAddresses, routeMatrix.RelayAddresses, "relay address mismatch")
			assert.Equal(t, costMatrix.RelayPublicKeys, routeMatrix.RelayPublicKeys, "relay public key mismatch")

			routeMatrixData, err := routeMatrix.MarshalBinary()
			assert.Nil(t, err)

			var readRouteMatrix routing.RouteMatrix
			err = readRouteMatrix.UnmarshalBinary(routeMatrixData)
			assert.Nil(t, err)

			assert.Equal(t, routeMatrix.RelayIDs, readRouteMatrix.RelayIDs, "relay id mismatch")
			// todo: relay names soon
			// this was the old line however because relay addresses are written with extra 0's this is how they must be checked
			// assert.Equal(t, routeMatrix.RelayAddresses, readRouteMatrix.RelayAddresses, "relay address mismatch")

			assert.Len(t, readCostMatrix.RelayAddresses, len(costMatrix.RelayAddresses))
			for i, addr := range costMatrix.RelayAddresses {
				assert.Equal(t, string(addr), strings.Trim(string(readCostMatrix.RelayAddresses[i]), string([]byte{0x0})))
			}
			assert.Equal(t, routeMatrix.RelayPublicKeys, readRouteMatrix.RelayPublicKeys, "relay public key mismatch")
			assert.Equal(t, routeMatrix.DatacenterRelays, readRouteMatrix.DatacenterRelays, "datacenter relays mismatch")

			equal := true

			assert.Len(t, readRouteMatrix.Entries, len(routeMatrix.Entries))
			for i := 0; i < len(routeMatrix.Entries); i++ {

				if routeMatrix.Entries[i].DirectRTT != readRouteMatrix.Entries[i].DirectRTT {
					t.Errorf("DirectRTT mismatch: %d != %d\n", routeMatrix.Entries[i].DirectRTT, readRouteMatrix.Entries[i].DirectRTT)
					equal = false
					break
				}

				if routeMatrix.Entries[i].NumRoutes != readRouteMatrix.Entries[i].NumRoutes {
					t.Errorf("NumRoutes mismatch\n")
					equal = false
					break
				}

				for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

					if routeMatrix.Entries[i].RouteRTT[j] != readRouteMatrix.Entries[i].RouteRTT[j] {
						t.Errorf("RouteRTT mismatch\n")
						equal = false
						break
					}

					if routeMatrix.Entries[i].RouteNumRelays[j] != readRouteMatrix.Entries[i].RouteNumRelays[j] {
						t.Errorf("RouteNumRelays mismatch\n")
						equal = false
						break
					}

					for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
						if routeMatrix.Entries[i].RouteRelays[j][k] != readRouteMatrix.Entries[i].RouteRelays[j][k] {
							t.Errorf("RouteRelayID mismatch\n")
							equal = false
							break
						}
					}
				}
			}

			assert.True(t, equal, "route matrix entries mismatch")
			analyze(t, &readRouteMatrix)
		})
	})
}
