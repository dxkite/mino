package xxor

import (
	"container/heap"
	"sync"
	"time"
)

type sessionItem struct {
	timestamp  int64
	sessionKey string
}

type sessionHeap []*sessionItem

func (s sessionHeap) Len() int { return len(s) }

func (s sessionHeap) Less(i, j int) bool {
	return s[i].timestamp < s[j].timestamp
}

func (s sessionHeap) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *sessionHeap) Push(x interface{}) {
	item := x.(*sessionItem)
	*s = append(*s, item)
}

func (pq *sessionHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type TtlSession struct {
	maxSize  int
	liveTime int
	session  map[string]int64
	timeHeap *sessionHeap
	mtx      sync.Mutex
}

const MaxSessionSize = 1024
const LiveTime = 60

var XorTtlSession = &TtlSession{
	maxSize:  MaxSessionSize,
	liveTime: LiveTime,
	session:  map[string]int64{},
	timeHeap: &sessionHeap{},
}

func (ts *TtlSession) Push(id string, timestamp int64) {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()
	ts.session[id] = timestamp
	heap.Push(ts.timeHeap, &sessionItem{
		timestamp:  timestamp,
		sessionKey: id,
	})
	if ts.timeHeap.Len() > ts.maxSize {
		delete(ts.session, (*ts.timeHeap)[0].sessionKey)
		heap.Remove(ts.timeHeap, 0)
	}
}

func (ts *TtlSession) clear() {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()
	curTime := time.Now().Unix()
	for ts.timeHeap.Len() > 0 && (*ts.timeHeap)[0].timestamp+int64(ts.liveTime) < curTime {
		delete(ts.session, (*ts.timeHeap)[0].sessionKey)
		heap.Remove(ts.timeHeap, 0)
	}
}

func (ts *TtlSession) Has(id string) bool {
	t, ok := ts.session[id]
	if t+int64(ts.liveTime) < time.Now().Unix() {
		go ts.clear()
		return false
	}
	return ok
}
