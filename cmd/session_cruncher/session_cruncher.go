package main

import (
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"
	"fmt"
	"os"

	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/encoding"
)

const TopSessionsCount = 10000

const NumBuckets = 1000

const SessionBatchVersion = uint64(0)

const TopSessionsVersion = uint64(0)

const BuyerStatsVersion = uint64(0)

const MapPointsVersion = uint64(0)

type SessionUpdate struct {
	sessionId uint64
	buyerId   uint64
	next      uint8
	latitude  float32
	longitude float32
}

type TopSessions struct {
	nextSessions   uint32
	totalSessions  uint32
	numTopSessions int
	topSessions    [TopSessionsCount]uint64
}

type Bucket struct {
	index                int
	mutex                sync.Mutex
	sessionUpdateChannel chan []SessionUpdate
	totalSessions        *SortedSet
	nextSessions         *SortedSet
	buyerTotalSessions   map[uint64]map[uint64]bool
	buyerNextSessions    map[uint64]map[uint64]bool
	mapEntries           map[uint64]MapEntry
}

var buckets []Bucket

var topSessionsMutex sync.Mutex
var topSessions *TopSessions
var topSessionsData []byte

type BuyerStats struct {
	buyerIds      []uint64
	totalSessions []uint32
	nextSessions  []uint32
}

var buyerDataMutex sync.Mutex
var buyerData []byte

type MapEntry struct {
	latitude  float32
	longitude float32
	next      uint8
}

type MapPoint struct {
	sessionId uint64
	next      uint8
	latitude  float32
	longitude float32
}

type MapPoints struct {
	numMapPoints int
	mapPoints    [TopSessionsCount]MapPoint
}

var mapDataMutex sync.Mutex
var mapData []byte

var timeSeriesPublisher *common.RedisTimeSeriesPublisher

func main() {

	redisPortalCluster := envvar.GetStringArray("REDIS_PORTAL_CLUSTER", []string{})
	redisPortalHostname := envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")

	core.Debug("redis portal cluster: %s", redisPortalCluster)
	core.Debug("redis portal hostname: %s", redisPortalHostname)

	service := common.CreateService("session_cruncher")

	service.Router.HandleFunc("/session_batch", sessionBatchHandler).Methods("POST")
	service.Router.HandleFunc("/top_sessions", topSessionsHandler).Methods("GET")
	service.Router.HandleFunc("/buyer_data", buyerDataHandler).Methods("GET")
	service.Router.HandleFunc("/map_data", mapDataHandler).Methods("GET")

	buckets = make([]Bucket, NumBuckets)
	for i := range buckets {
		buckets[i].index = i
		buckets[i].sessionUpdateChannel = make(chan []SessionUpdate, 1000000)
		buckets[i].nextSessions = NewSortedSet()
		buckets[i].totalSessions = NewSortedSet()
		buckets[i].buyerTotalSessions = make(map[uint64]map[uint64]bool)
		buckets[i].buyerNextSessions = make(map[uint64]map[uint64]bool)
		buckets[i].mapEntries = make(map[uint64]MapEntry, 10000)
		StartProcessThread(&buckets[i])
	}

	timeSeriesConfig := common.RedisTimeSeriesConfig{
		RedisHostname: redisPortalHostname,
		RedisCluster:  redisPortalCluster,
	}

	var err error
	timeSeriesPublisher, err = common.CreateRedisTimeSeriesPublisher(service.Context, timeSeriesConfig)
	if err != nil {
		core.Error("could not create redis time series publisher: %v", err)
		os.Exit(1)
	}

	UpdateTopSessions(&TopSessions{})

	UpdateBuyerData(&BuyerStats{})

	UpdateMapData(&MapPoints{})

	//go TestThread()

	go TopSessionsThread()

	service.StartWebServer()

	service.WaitForShutdown()
}

func TestThread() {
	for {
		for index := 0; index < NumBuckets; index++ {
			batch := make([]SessionUpdate, 1000)
			for i := 0; i < len(batch); i++ {
				batch[i].sessionId = rand.Uint64()
				batch[i].buyerId = uint64(common.RandomInt(0, 9))
				batch[i].next = uint8(i % 2)
				batch[i].latitude = rand.Float32()
				batch[i].longitude = rand.Float32()
			}
			buckets[index].sessionUpdateChannel <- batch
			time.Sleep(2 * time.Millisecond)
		}
	}
}

func GetBucketIndex(score uint32) int {
	index := score
	if index < 0 {
		index = 0
	} else if index > NumBuckets-1 {
		index = NumBuckets - 1
	}
	return int(index)
}

func StartProcessThread(bucket *Bucket) {
	go func() {
		for {
			select {
			case batch := <-bucket.sessionUpdateChannel:
				bucket.mutex.Lock()
				for i := range batch {

					bucket.totalSessions.Insert(batch[i].sessionId, uint32(bucket.index))

					buyerTotalSessions, exists := bucket.buyerTotalSessions[batch[i].buyerId]
					if !exists {
						buyerTotalSessions = make(map[uint64]bool)
						bucket.buyerTotalSessions[batch[i].buyerId] = buyerTotalSessions
					}
					buyerTotalSessions[batch[i].sessionId] = true

					if batch[i].next != 0 {

						bucket.nextSessions.Insert(batch[i].sessionId, uint32(bucket.index))

						buyerNextSessions, exists := bucket.buyerNextSessions[batch[i].buyerId]
						if !exists {
							buyerNextSessions = make(map[uint64]bool)
							bucket.buyerNextSessions[batch[i].buyerId] = buyerNextSessions
						}
						buyerNextSessions[batch[i].sessionId] = true
					}

					bucket.mapEntries[batch[i].sessionId] = MapEntry{next: batch[i].next, latitude: batch[i].latitude, longitude: batch[i].longitude}
				}
				bucket.mutex.Unlock()
			}
		}
	}()
}

func UpdateTopSessions(newTopSessions *TopSessions) {

	data := make([]byte, 8+4+4+4+8*newTopSessions.numTopSessions)

	index := 0

	encoding.WriteUint64(data[:], &index, TopSessionsVersion)
	encoding.WriteUint32(data[:], &index, newTopSessions.nextSessions)
	encoding.WriteUint32(data[:], &index, newTopSessions.totalSessions)

	for i := 0; i < newTopSessions.numTopSessions; i++ {
		encoding.WriteUint64(data[:], &index, newTopSessions.topSessions[i])
	}

	topSessionsMutex.Lock()
	topSessions = newTopSessions
	topSessionsData = data
	topSessionsMutex.Unlock()
}

func UpdateBuyerData(newBuyerStats *BuyerStats) {

	data := make([]byte, 8+4+len(newBuyerStats.buyerIds)*(8+4+4))

	index := 0

	encoding.WriteUint64(data[:], &index, BuyerStatsVersion)
	encoding.WriteUint32(data[:], &index, uint32(len(newBuyerStats.buyerIds)))

	for i := 0; i < len(newBuyerStats.buyerIds); i++ {
		encoding.WriteUint64(data[:], &index, newBuyerStats.buyerIds[i])
		encoding.WriteUint32(data[:], &index, newBuyerStats.totalSessions[i])
		encoding.WriteUint32(data[:], &index, newBuyerStats.nextSessions[i])
	}

	buyerDataMutex.Lock()
	buyerData = data
	buyerDataMutex.Unlock()
}

func UpdateMapData(newMapPoints *MapPoints) {

	data := make([]byte, 8+4+newMapPoints.numMapPoints*(8+1+4+4))

	index := 0

	encoding.WriteUint64(data[:], &index, MapPointsVersion)
	encoding.WriteUint32(data[:], &index, uint32(newMapPoints.numMapPoints))

	for i := 0; i < newMapPoints.numMapPoints; i++ {
		encoding.WriteUint64(data[:], &index, newMapPoints.mapPoints[i].sessionId)
		encoding.WriteUint8(data[:], &index, newMapPoints.mapPoints[i].next)
		encoding.WriteFloat32(data[:], &index, newMapPoints.mapPoints[i].latitude)
		encoding.WriteFloat32(data[:], &index, newMapPoints.mapPoints[i].longitude)
	}

	mapDataMutex.Lock()
	mapData = data
	mapDataMutex.Unlock()
}

func TopSessionsThread() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:

			core.Log("-------------------------------------------------------------------")

			nextSessions := make([]*SortedSet, NumBuckets)
			totalSessions := make([]*SortedSet, NumBuckets)
			buyerTotalSessions := make([]map[uint64]map[uint64]bool, NumBuckets)
			buyerNextSessions := make([]map[uint64]map[uint64]bool, NumBuckets)
			mapEntries := make([]map[uint64]MapEntry, NumBuckets)

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Lock()
			}

			for i := 0; i < NumBuckets; i++ {
				nextSessions[i] = buckets[i].nextSessions
				totalSessions[i] = buckets[i].totalSessions
				mapEntries[i] = buckets[i].mapEntries
				buyerTotalSessions[i] = buckets[i].buyerTotalSessions
				buyerNextSessions[i] = buckets[i].buyerNextSessions
				buckets[i].nextSessions = NewSortedSet()
				buckets[i].totalSessions = NewSortedSet()
				buckets[i].buyerTotalSessions = make(map[uint64]map[uint64]bool)
				buckets[i].buyerNextSessions = make(map[uint64]map[uint64]bool)
				buckets[i].mapEntries = make(map[uint64]MapEntry, 10000)
			}

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Unlock()
			}

			start := time.Now()

			// build global top sessions list and total sessions / next sessions counts

			maxNextSessions := 0
			maxTotalSessions := 0

			for i := 0; i < NumBuckets; i++ {
				maxNextSessions += nextSessions[i].GetCount()
				maxTotalSessions += totalSessions[i].GetCount()
			}

			nextSessionsMap := make(map[uint64]bool, maxNextSessions)
			totalSessionsMap := make(map[uint64]bool, maxTotalSessions)

			type Session struct {
				sessionId uint64
				score     uint32
			}

			sessions := make([]Session, 0, TopSessionsCount)

			for i := 0; i < NumBuckets; i++ {
				bucketNextSessions := nextSessions[i].GetByRankRange(1, -1)
				bucketTotalSessions := totalSessions[i].GetByRankRange(1, -1)
				for j := range bucketNextSessions {
					if _, exists := nextSessionsMap[bucketNextSessions[j].Key]; !exists {
						nextSessionsMap[bucketNextSessions[j].Key] = true
					}
				}
				for j := range bucketTotalSessions {
					if _, exists := totalSessionsMap[bucketTotalSessions[j].Key]; !exists {
						totalSessionsMap[bucketTotalSessions[j].Key] = true
						if len(sessions) < TopSessionsCount {
							sessions = append(sessions, Session{sessionId: bucketTotalSessions[j].Key, score: bucketTotalSessions[j].Score})
						}
					}
				}
			}

			nextCount := len(nextSessionsMap)
			totalCount := len(totalSessionsMap)

			newTopSessions := &TopSessions{}
			newTopSessions.nextSessions = uint32(nextCount)
			newTopSessions.totalSessions = uint32(totalCount)
			newTopSessions.numTopSessions = len(sessions)
			for i := range sessions {
				newTopSessions.topSessions[i] = sessions[i].sessionId
			}

			UpdateTopSessions(newTopSessions)

			// build data for the map, derived from the top sessions list

			newMapPoints := MapPoints{}
			newMapPoints.numMapPoints = len(sessions)
			for i := range sessions {
				newMapPoints.mapPoints[i].sessionId = sessions[i].sessionId
				score := sessions[i].score
				entry := mapEntries[score][sessions[i].sessionId]
				newMapPoints.mapPoints[i].next = entry.next
				newMapPoints.mapPoints[i].latitude = entry.latitude
				newMapPoints.mapPoints[i].longitude = entry.longitude
			}

			UpdateMapData(&newMapPoints)

			// build per-buyer total sessions / next sessions counts

			buyerMap := make(map[uint64]bool)

			for i := 0; i < NumBuckets; i++ {
				for k := range buyerTotalSessions[i] {
					buyerMap[k] = true
				}
			}

			buyers := make([]uint64, len(buyerMap))
			index := 0
			for k := range buyerMap {
				buyers[index] = k
				index++
			}

			sort.Slice(buyers, func(i, j int) bool { return buyers[i] < buyers[j] })

			core.Log("buyers: %v", buyers)

			buyerStats := BuyerStats{}
			buyerStats.buyerIds = make([]uint64, len(buyers))
			buyerStats.totalSessions = make([]uint32, len(buyers))
			buyerStats.nextSessions = make([]uint32, len(buyers))

			for i := range buyers {
				buyerStats.buyerIds[i] = buyers[i]
				for j := 0; j < NumBuckets; j++ {
					buyerStats.totalSessions[i] += uint32(len(buyerTotalSessions[j][buyers[i]]))
					buyerStats.nextSessions[i] += uint32(len(buyerNextSessions[j][buyers[i]]))
				}
				core.Log("buyer %d => %d/%d sessions", buyerStats.buyerIds[i], buyerStats.nextSessions[i], buyerStats.totalSessions[i])
			}

			UpdateBuyerData(&buyerStats)

			duration := time.Since(start)

			core.Log("top %d of %d/%d sessions (%.6fms)", len(sessions), nextCount, totalCount, float64(duration.Nanoseconds())/1000000.0)

			// publish time series data to redis

			message := common.RedisTimeSeriesMessage{}

			message.Timestamp = uint64(time.Now().Unix())

			message.Keys = []string{"total_sessions", "next_sessions", "accelerated_percent"}

			acceleratedPercent := 0.0
			if totalCount > 0 {
				acceleratedPercent = float64(nextCount / totalCount) * 100.0
			}

			message.Values = []float64{float64(totalCount), float64(nextCount), acceleratedPercent}

			for i := range buyers {
				message.Keys = append(message.Keys, fmt.Sprintf("%016x_total_sessions", buyers[i]))
				message.Keys = append(message.Keys, fmt.Sprintf("%016x_next_sessions", buyers[i]))
				message.Keys = append(message.Keys, fmt.Sprintf("%016x_accelerated_percent", buyers[i]))
				buyerAcceleratedPercent := 0.0
				if buyerStats.totalSessions[i] > 0 {
					buyerAcceleratedPercent = float64(buyerStats.nextSessions[i] / buyerStats.totalSessions[i]) * 100.0
				}
				message.Values = append(message.Values, float64(buyerStats.totalSessions[i]))
				message.Values = append(message.Values, float64(buyerStats.nextSessions[i]))
				message.Values = append(message.Values, float64(buyerAcceleratedPercent))
			}

			timeSeriesPublisher.MessageChannel <- &message
		}
	}
}

func sessionBatchHandler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("could not read session batch body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	if len(body) < 8 {
		core.Error("session batch is too small")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	version := binary.LittleEndian.Uint64(body[0:8])

	if version != SessionBatchVersion {
		core.Error("session batch has unknown version %d, expected %d\n", version, SessionBatchVersion)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body = body[8:]

	index := 0
	for j := 0; j < NumBuckets; j++ {
		var numUpdates uint32
		encoding.ReadUint32(body[:], &index, &numUpdates)
		batch := make([]SessionUpdate, numUpdates)
		if numUpdates > 0 {
			for i := 0; i < int(numUpdates); i++ {
				encoding.ReadUint64(body[:], &index, &batch[i].sessionId)
				encoding.ReadUint64(body[:], &index, &batch[i].buyerId)
				encoding.ReadUint8(body[:], &index, &batch[i].next)
				encoding.ReadFloat32(body[:], &index, &batch[i].latitude)
				encoding.ReadFloat32(body[:], &index, &batch[i].longitude)
			}
			buckets[j].sessionUpdateChannel <- batch
		}
	}
}

func topSessionsHandler(w http.ResponseWriter, r *http.Request) {
	topSessionsMutex.Lock()
	data := topSessionsData
	topSessionsMutex.Unlock()
	w.Write(data)
}

func buyerDataHandler(w http.ResponseWriter, r *http.Request) {
	buyerDataMutex.Lock()
	data := buyerData
	buyerDataMutex.Unlock()
	w.Write(data)
}

func mapDataHandler(w http.ResponseWriter, r *http.Request) {
	mapDataMutex.Lock()
	data := mapData
	mapDataMutex.Unlock()
	w.Write(data)
}

// ---------------------------------------------------------------------------------------

// Copyright (c) 2016, Jerry.Wang
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//  this list of conditions and the following disclaimer in the documentation
//  and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

const SKIPLIST_MAXLEVEL = 32 /* Should be enough for 2^32 elements */
const SKIPLIST_P = 0.25      /* Skiplist P = 1/4 */

type SortedSet struct {
	header *SortedSetNode
	tail   *SortedSetNode
	length int64
	level  int
	dict   map[uint64]*SortedSetNode
}

type SortedSetLevel struct {
	forward *SortedSetNode
	span    int64
}

type SortedSetNode struct {
	Key      uint64 // unique key of this node
	Score    uint32 // score to determine the order of this node in the set
	backward *SortedSetNode
	level    []SortedSetLevel
}

func createNode(level int, score uint32, key uint64) *SortedSetNode {
	node := SortedSetNode{
		Score: score,
		Key:   key,
		level: make([]SortedSetLevel, level),
	}
	return &node
}

func randomLevel() int {
	level := 1
	for float64(rand.Int31()&0xFFFF) < float64(SKIPLIST_P*0xFFFF) {
		level += 1
	}
	if level < SKIPLIST_MAXLEVEL {
		return level
	}
	return SKIPLIST_MAXLEVEL
}

func (this *SortedSet) insertNode(score uint32, key uint64) *SortedSetNode {
	var update [SKIPLIST_MAXLEVEL]*SortedSetNode
	var rank [SKIPLIST_MAXLEVEL]int64

	x := this.header
	for i := this.level - 1; i >= 0; i-- {
		/* store rank that is crossed to reach the insert position */
		if this.level-1 == i {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score && // score is the same but the key is different
					x.level[i].forward.Key < key)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* we assume the key is not already inside, since we allow duplicated
	 * scores, and the re-insertion of score and redis object should never
	 * happen since the caller of Insert() should test in the hash table
	 * if the element is already inside or not. */
	level := randomLevel()

	if level > this.level { // add a new level
		for i := this.level; i < level; i++ {
			rank[i] = 0
			update[i] = this.header
			update[i].level[i].span = this.length
		}
		this.level = level
	}

	x = createNode(level, score, key)
	for i := 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	/* increment span for untouched levels */
	for i := level; i < this.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == this.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		this.tail = x
	}
	this.length++

	return x
}

func (this *SortedSet) deleteNode(x *SortedSetNode, update [SKIPLIST_MAXLEVEL]*SortedSetNode) {
	for i := 0; i < this.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span -= 1
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		this.tail = x.backward
	}
	for this.level > 1 && this.header.level[this.level-1].forward == nil {
		this.level--
	}
	this.length--
	delete(this.dict, x.Key)
}

func (this *SortedSet) delete(score uint32, key uint64) bool {
	var update [SKIPLIST_MAXLEVEL]*SortedSetNode

	x := this.header
	for i := this.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Key < key)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* We may have multiple elements with the same score, what we need
	 * is to find the element with both the right score and object. */
	x = x.level[0].forward
	if x != nil && score == x.Score && x.Key == key {
		this.deleteNode(x, update)
		// free x
		return true
	}
	return false /* not found */
}

func NewSortedSet() *SortedSet {
	sortedSet := SortedSet{
		level: 1,
		dict:  make(map[uint64]*SortedSetNode, 10000),
	}
	sortedSet.header = createNode(SKIPLIST_MAXLEVEL, 0, 0)
	return &sortedSet
}

func (this *SortedSet) GetCount() int {
	return int(this.length)
}

func (this *SortedSet) Insert(key uint64, score uint32) bool {
	var newNode *SortedSetNode = nil
	found := this.dict[key]
	if found != nil {
		if found.Score != score {
			this.delete(found.Score, found.Key)
			newNode = this.insertNode(score, key)
		}
	} else {
		newNode = this.insertNode(score, key)
	}
	if newNode != nil {
		this.dict[key] = newNode
	}
	return found == nil
}

func (this *SortedSet) Delete(key uint64) *SortedSetNode {
	found := this.dict[key]
	if found != nil {
		this.delete(found.Score, found.Key)
		return found
	}
	return nil
}

func (this *SortedSet) sanitizeIndexes(start int, end int) (int, int) {
	if start < 0 {
		start = int(this.length) + start + 1
	}
	if end < 0 {
		end = int(this.length) + end + 1
	}
	if start <= 0 {
		start = 1
	}
	if end <= 0 {
		end = 1
	}
	return start, end
}

func (this *SortedSet) findNodeByRank(start int) (traversed int, x *SortedSetNode) {
	x = this.header
	for i := this.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			traversed+int(x.level[i].span) < start {
			traversed += int(x.level[i].span)
			x = x.level[i].forward
		}
		if traversed+1 == start {
			break
		}
	}
	return
}

func (this *SortedSet) GetByRankRange(start int, end int) []*SortedSetNode {

	start, end = this.sanitizeIndexes(start, end)

	var nodes []*SortedSetNode

	traversed, x := this.findNodeByRank(start)

	traversed++
	x = x.level[0].forward
	for x != nil && traversed <= end {
		next := x.level[0].forward
		nodes = append(nodes, x)
		traversed++
		x = next
	}

	return nodes
}
