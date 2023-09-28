package main

import (
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/encoding"
)

const MaxServerAddressLength = 64 // IMPORTANT: Enough for IPv4 and IPv6 + port number

const TopServersCount = 10000

const NumBuckets = 1000

const ServerBatchVersion = uint64(0)

const TopServersVersion = uint64(0)

type ServerUpdate struct {
	serverAddress string
}

type TopServers struct {
	totalServerCount uint32
	numTopServers    int
	topServers       [TopServersCount]string
}

type Bucket struct {
	index               int
	mutex               sync.Mutex
	serverUpdateChannel chan []ServerUpdate
	totalServers        *SortedSet
}

var buckets []Bucket

var topServersMutex sync.Mutex
var topServers *TopServers
var topServersData []byte

func main() {

	service := common.CreateService("server_cruncher")

	service.Router.HandleFunc("/server_batch", serverBatchHandler).Methods("POST")
	service.Router.HandleFunc("/top_servers", topServersHandler).Methods("GET")

	buckets = make([]Bucket, NumBuckets)
	for i := range buckets {
		buckets[i].index = i
		buckets[i].serverUpdateChannel = make(chan []ServerUpdate, 1000000)
		buckets[i].totalServers = NewSortedSet()
		StartProcessThread(&buckets[i])
	}

	UpdateTopServers(&TopServers{})

	// go TestThread()

	go TopSessionsThread()

	service.StartWebServer()

	service.WaitForShutdown()
}

func TestThread() {
	for {
		for index := 0; index < NumBuckets; index++ {
			batch := make([]ServerUpdate, 1000)
			for i := 0; i < len(batch); i++ {
				serverAddress := common.RandomAddress()
				batch[i].serverAddress = serverAddress.String()
			}
			buckets[index].serverUpdateChannel <- batch
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
			case batch := <-bucket.serverUpdateChannel:
				bucket.mutex.Lock()
				for i := range batch {
					bucket.totalServers.Insert(batch[i].serverAddress, uint32(bucket.index))
				}
				bucket.mutex.Unlock()
			}
		}
	}()
}

func UpdateTopServers(newTopServers *TopServers) {

	data := make([]byte, 8+4+4+newTopServers.numTopServers*(4+MaxServerAddressLength))

	index := 0

	encoding.WriteUint64(data[:], &index, TopServersVersion)
	encoding.WriteUint32(data[:], &index, newTopServers.totalServerCount)
	encoding.WriteUint32(data[:], &index, uint32(newTopServers.numTopServers))

	for i := 0; i < newTopServers.numTopServers; i++ {
		encoding.WriteString(data[:], &index, newTopServers.topServers[i], MaxServerAddressLength)
	}

	topServersMutex.Lock()
	topServers = newTopServers
	topServersData = data[:index]
	topServersMutex.Unlock()
}

func TopSessionsThread() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:

			totalServers := make([]*SortedSet, NumBuckets)

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Lock()
			}

			for i := 0; i < NumBuckets; i++ {
				totalServers[i] = buckets[i].totalServers
				buckets[i].totalServers = NewSortedSet()
			}

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Unlock()
			}

			start := time.Now()

			maxTotalServers := 0

			for i := 0; i < NumBuckets; i++ {
				maxTotalServers += totalServers[i].GetCount()
			}

			totalServersMap := make(map[string]bool, maxTotalServers)

			type Server struct {
				serverAddress string
				score         uint32
			}

			servers := make([]Server, 0, TopServersCount)

			for i := 0; i < NumBuckets; i++ {
				bucketTotalServers := totalServers[i].GetByRankRange(1, -1)
				for j := range bucketTotalServers {
					if _, exists := totalServersMap[bucketTotalServers[j].Key]; !exists {
						totalServersMap[bucketTotalServers[j].Key] = true
						if len(servers) < TopServersCount {
							servers = append(servers, Server{serverAddress: bucketTotalServers[j].Key, score: bucketTotalServers[j].Score})
						}
					}
				}
			}

			totalServerCount := len(totalServersMap)

			newTopServers := &TopServers{}
			newTopServers.totalServerCount = uint32(totalServerCount)
			newTopServers.numTopServers = len(servers)
			for i := range servers {
				newTopServers.topServers[i] = servers[i].serverAddress
			}

			UpdateTopServers(newTopServers)

			duration := time.Since(start)

			core.Log("top %d of %d servers (%.6fms)", len(servers), totalServerCount, float64(duration.Nanoseconds())/1000000.0)
		}
	}
}

func serverBatchHandler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("could not read server batch body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	if len(body) < 8 {
		core.Error("server batch is too small")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	version := binary.LittleEndian.Uint64(body[0:8])

	if version != ServerBatchVersion {
		core.Error("server batch has unknown version %d, expected %d\n", version, ServerBatchVersion)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body = body[8:]

	index := 0
	for j := 0; j < NumBuckets; j++ {
		var numUpdates uint32
		encoding.ReadUint32(body[:], &index, &numUpdates)
		batch := make([]ServerUpdate, numUpdates)
		if numUpdates > 0 {
			for i := 0; i < int(numUpdates); i++ {
				encoding.ReadString(body[:], &index, &batch[i].serverAddress, MaxServerAddressLength)
			}
			buckets[j].serverUpdateChannel <- batch
		}
	}
}

func topServersHandler(w http.ResponseWriter, r *http.Request) {
	topServersMutex.Lock()
	data := topServersData
	topServersMutex.Unlock()
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
	dict   map[string]*SortedSetNode
}

type SortedSetLevel struct {
	forward *SortedSetNode
	span    int64
}

type SortedSetNode struct {
	Key      string // unique key of this node
	Score    uint32 // score to determine the order of this node in the set
	backward *SortedSetNode
	level    []SortedSetLevel
}

func createNode(level int, score uint32, key string) *SortedSetNode {
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

func (this *SortedSet) insertNode(score uint32, key string) *SortedSetNode {
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

func (this *SortedSet) delete(score uint32, key string) bool {
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
		dict:  make(map[string]*SortedSetNode, 10000),
	}
	sortedSet.header = createNode(SKIPLIST_MAXLEVEL, 0, "")
	return &sortedSet
}

func (this *SortedSet) GetCount() int {
	return int(this.length)
}

func (this *SortedSet) Insert(key string, score uint32) bool {
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

func (this *SortedSet) Delete(key string) *SortedSetNode {
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
