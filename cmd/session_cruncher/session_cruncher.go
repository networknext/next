package main

import (
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/encoding"
)

const TopSessionsCount = 100000

const NumBuckets = 1000

const SessionBatchVersion = uint64(0)

const TopSessionsVersion = uint64(0)

type SessionUpdate struct {
	sessionId uint64
	score     int32
	next      bool
}

type SessionDelete struct {
	sessionId uint64
}

type TopSessions struct {
	nextSessions   uint32
	totalSessions  uint32
	numTopSessions int
	topSessions    [TopSessionsCount]uint64
}

type Bucket struct {
	mutex                sync.Mutex
	sessionUpdateChannel chan SessionUpdate
	sessionDeleteChannel chan SessionDelete
	currentSessions      *SortedSet
	previousSessions     *SortedSet
	currentNext          map[uint64]bool
	previousNext         map[uint64]bool
}

var buckets []Bucket

var topSessionsMutex sync.Mutex
var topSessions *TopSessions
var topSessionsData []byte

func main() {

	service := common.CreateService("session_cruncher")

	service.Router.HandleFunc("/session_batch", sessionBatchHandler).Methods("POST")
	service.Router.HandleFunc("/top_sessions", topSessionsHandler).Methods("GET")

	buckets = make([]Bucket, NumBuckets)
	for i := range buckets {
		buckets[i].sessionUpdateChannel = make(chan SessionUpdate, 1000000)
		buckets[i].currentSessions = NewSortedSet()
		buckets[i].previousSessions = NewSortedSet()
		buckets[i].currentNext = make(map[uint64]bool)
		buckets[i].previousNext = make(map[uint64]bool)
		StartProcessThread(&buckets[i])
	}

	topSessionsMutex.Lock()
	topSessions = &TopSessions{}
	topSessionsMutex.Unlock()

	go TopSessionsThread()

	service.StartWebServer()

	service.WaitForShutdown()
}

func GetBucketIndex(score int32) int {
	index := score
	if index < 0 {
		index = 0
	} else if index > NumBuckets-1 {
		index = NumBuckets - 1
	}
	return int(index)
}

func StartProcessThread(bucket *Bucket) {

	minute := time.Now().Unix() / 60

	go func() {
		for {
			ticker := time.NewTicker(time.Second)
			select {

			case sessionUpdate := <-bucket.sessionUpdateChannel:

				bucket.mutex.Lock()
				bucket.currentSessions.Insert(sessionUpdate.sessionId, sessionUpdate.score)
				if sessionUpdate.next {
					bucket.currentNext[sessionUpdate.sessionId] = true
				}
				bucket.mutex.Unlock()

			case sessionDelete := <-bucket.sessionDeleteChannel:

				bucket.mutex.Lock()
				bucket.currentSessions.Delete(sessionDelete.sessionId)
				delete(bucket.currentNext, sessionDelete.sessionId)
				bucket.mutex.Unlock()

			case <-ticker.C:

				currentTime := time.Now().Unix()

				currentMinute := currentTime / 60

				if currentMinute > minute {
					bucket.mutex.Lock()
					bucket.previousSessions = bucket.currentSessions
					bucket.currentSessions = NewSortedSet()
					bucket.previousNext = bucket.currentNext
					bucket.currentNext = make(map[uint64]bool)
					bucket.mutex.Unlock()
					minute = currentMinute
				}
			}
		}
	}()
}

func TopSessionsThread() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:

			bucketDistribution := make([]uint64, 0)

			start := time.Now()

			sessions_a := make([]*SortedSetNode, 0, len(TopSessionsCount)*2)
			sessions_b := make([]*SortedSetNode, 0, len(TopSessionsCount)*2)

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Lock()
			}

			for i := 0; i < NumBuckets; i++ {
				bucketSessions := buckets[i].currentSessions.GetByRankRange(1, TopSessionsCount)
				sessions_a = append(sessions_a, bucketSessions...)
				if len(sessions_a) >= TopSessionsCount {
					sessions_a = sessions_a[:TopSessionsCount]
					break
				}
			}

			for i := 0; i < NumBuckets; i++ {
				bucketSessions := buckets[i].previousSessions.GetByRankRange(1, TopSessionsCount)
				sessions_b = append(sessions_b, bucketSessions...)
				if len(sessions_b) >= TopSessionsCount {
					sessions_b = sessions_b[:TopSessionsCount]
					break
				}
			}

			totalCount_a := uint64(0)
			totalCount_b := uint64(0)
			nextCount_a := uint64(0)
			nextCount_b := uint64(0)
			for i := 0; i < NumBuckets; i++ {
				totalCount_a += uint64(buckets[i].currentSessions.GetCount())
				totalCount_b += uint64(buckets[i].previousSessions.GetCount())
				nextCount_a += uint64(len(buckets[i].currentNext))
				nextCount_b += uint64(len(buckets[i].previousNext))
			}

			for i := 0; i < NumBuckets; i++ {
				buckets[i].mutex.Unlock()
			}

			totalCount := totalCount_a
			if totalCount_b > totalCount {
				totalCount = totalCount_b
			}

			nextCount := nextCount_a
			if nextCount_b > nextCount {
				nextCount = nextCount_b
			}

			sessionMap := make(map[uint64]int32)

			for i := range sessions_a {
				sessionMap[sessions_a[i].Key] = sessions_a[i].Score
			}

			for i := range sessions_b {
				sessionMap[sessions_b[i].Key] = sessions_b[i].Score
			}

			type Session struct {
				sessionId uint64
				score     int32
			}

			sessions := make([]Session, len(sessionMap))
			index := 0
			for k, v := range sessionMap {
				sessions[index].sessionId = k
				sessions[index].score = v
				index++
			}

			sort.Slice(sessions, func(i int, j int) bool { return sessions[i].score < sessions[j].score })

			if len(sessions) > TopSessionsCount {
				sessions = sessions[:TopSessionsCount]
			}

			newTopSessions := &TopSessions{}
			newTopSessions.nextSessions = uint32(nextCount)
			newTopSessions.totalSessions = uint32(totalCount)
			newTopSessions.numTopSessions = len(sessions)
			for i := range sessions {
				newTopSessions.topSessions[i] = sessions[i].sessionId
			}

			data := make([]byte, 8+8+8+4+8*newTopSessions.numTopSessions)

			index = 0

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

			duration := time.Since(start)

			core.Log("top %d of %d/%d sessions (%.6fms)", len(sessions), nextCount, totalCount, float64(duration.Nanoseconds())/1000000.0)

			if duration > time.Second {
				core.Warn("session cruncher can't keep up!")
			}
		}
	}
}

func sessionBatchHandler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("could not read session batch body")
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

	numSessionUpdates := len(body) / 17

	index := 0
	for i := 0; i < numSessionUpdates; i++ {
		sessionId := binary.LittleEndian.Uint64(body[index : index+8])
		currentScore := int32(binary.LittleEndian.Uint32(body[index+8 : index+12]))
		previousScore := int32(binary.LittleEndian.Uint32(body[index+12 : index+16]))
		var next bool
		if body[index+16] != 0 {
			next = true
		}
		currentIndex := GetBucketIndex(currentScore)
		previousIndex := GetBucketIndex(previousScore)
		buckets[currentIndex].sessionUpdateChannel <- SessionUpdate{sessionId: sessionId, score: currentScore, next: next}
		if previousIndex != currentIndex {
			buckets[previousIndex].sessionDeleteChannel <- SessionDelete{sessionId: sessionId}
		}
		index += 17
	}
}

func topSessionsHandler(w http.ResponseWriter, r *http.Request) {
	topSessionsMutex.Lock()
	data := topSessionsData
	topSessionsMutex.Unlock()
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
	Score    int32  // score to determine the order of this node in the set
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
