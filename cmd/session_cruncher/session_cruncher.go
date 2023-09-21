package main

import (
	"fmt"
	"time"
	"net/http"
	"io/ioutil"
	"math/rand"
	"encoding/binary"
	"sync"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
)

type SessionUpdate struct {
	timestamp     uint64
	sessionId	  uint64
	previousScore int32
	newScore      int32
	delete        bool
}

type Bucket struct {
	mutex 			sync.Mutex
	sessionChannel 	chan SessionUpdate
	set 			*SortedSet
}

const NumBuckets = 1000

var buckets []Bucket

func main() {

	service := common.CreateService("session_cruncher")

	service.Router.HandleFunc("/session_batch", sessionBatchHandler).Methods("POST")
	service.Router.HandleFunc("/session_counts", sessionCountsHandler).Methods("GET")
	service.Router.HandleFunc("/sessions/{begin}/{end}", sessionsHandler).Methods("GET")

	buckets = make([]Bucket, NumBuckets)

	for i := range buckets {
		buckets[i].sessionChannel = make(chan SessionUpdate, 1000000)
		buckets[i].set = NewSortedSet()
		StartProcessThread(i)
	}

	go SortThread()

	go TestThread()

	service.StartWebServer()

	service.WaitForShutdown()
}

func GetBucketIndex(score int32) int {
	index := score
	if index < 0 {
		index = 0
	} else if index > NumBuckets - 1 {
		index = NumBuckets - 1
	}
	return int(index)
}

func TestThread() {
	i := uint64(0) 
	for {
		session := SessionUpdate{}
		session.timestamp = uint64(time.Now().Unix())
		session.sessionId = rand.Uint64()
		session.newScore = int32(i%NumBuckets)
		session.previousScore = int32((i+5)%NumBuckets)						// todo: mock worst case, removing from bucket each update...
		previousIndex := GetBucketIndex(session.previousScore)
		newIndex := GetBucketIndex(session.newScore)
		buckets[newIndex].sessionChannel <- session
		if previousIndex != newIndex {
			session.delete = true
			buckets[previousIndex].sessionChannel <- session
		}
		i++
	}
}

func StartProcessThread(i int) {
	bucket := &buckets[i]
	go func() {
		for {
			select {
			case session := <-bucket.sessionChannel:
				if !session.delete {
					bucket.mutex.Lock()
					bucket.set.Insert(session.sessionId, session.newScore)
					bucket.mutex.Unlock()
				} else {
					bucket.mutex.Lock()
					bucket.set.Delete(session.sessionId)
					bucket.mutex.Unlock()
				}
			}
		}
	}()
}

func SortThread() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:

			start := time.Now()

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Lock()
			}

			const TopSessions = 100000

			sessions := make([]*SortedSetNode, 0)

			for i := 0; i < NumBuckets; i++ {
				bucketSessions := buckets[i].set.GetByRankRange(1, TopSessions)
				sessions = append(sessions, bucketSessions...)
				if len(sessions) >= TopSessions {
					sessions = sessions[:TopSessions]
					break
				}
			}

			totalCount := uint64(0)
			for i := 0; i < NumBuckets; i++ {
				totalCount += uint64(buckets[i].set.GetCount())
			}

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Unlock()
			}

			fmt.Printf("top %d/%d sessions: %.6fms\n", len(sessions), totalCount, float64(time.Since(start).Nanoseconds())/1000000.0)
		}
	}
}

func sessionBatchHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("session batch handler")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("could not read body")
		return
	}
	defer r.Body.Close()
	if len(body) % 16 != 0 {
		core.Error("session batch should be multiple of 16 bytes")
		return
	}
	numSessions := len(body) / 16
	index := 0
	currentTime := uint64(time.Now().Unix())
	session := SessionUpdate{}
	for i := 0; i < numSessions; i++ {
		session.timestamp = currentTime
		session.sessionId = binary.LittleEndian.Uint64(body[index:index+8])
		session.newScore = int32(binary.LittleEndian.Uint32(body[index+8:index+12]))
		session.previousScore = int32(binary.LittleEndian.Uint32(body[index+12:index+16]))
		previousIndex := GetBucketIndex(session.previousScore)
		newIndex := GetBucketIndex(session.newScore)
		buckets[newIndex].sessionChannel <- session
		if previousIndex != newIndex {
			session.delete = true
			buckets[previousIndex].sessionChannel <- session
		}
		index += 16
    }
}

func sessionCountsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("session counts handler")
}

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("sessions handler")
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
	Key      uint64      // unique key of this node
	Score    int32       // score to determine the order of this node in the set
	backward *SortedSetNode
	level    []SortedSetLevel
}

func createNode(level int, score int32, key uint64) *SortedSetNode {
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

func (this *SortedSet) insertNode(score int32, key uint64) *SortedSetNode {
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

func (this *SortedSet) delete(score int32, key uint64) bool {
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
		dict:  make(map[uint64]*SortedSetNode),
	}
	sortedSet.header = createNode(SKIPLIST_MAXLEVEL, 0, 0)
	return &sortedSet
}

func (this *SortedSet) GetCount() int {
	return int(this.length)
}

func (this *SortedSet) Insert(key uint64, score int32) bool {
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
