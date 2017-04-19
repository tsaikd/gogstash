package inputdockerstats

import (
	"sync"
	"time"
)

type sincemap struct {
	data  map[string]*time.Time
	mutex *sync.RWMutex
}

func newSinceMap() sincemap {
	return sincemap{
		data:  map[string]*time.Time{},
		mutex: &sync.RWMutex{},
	}
}

func (t sincemap) ensure(id string) *time.Time {
	t.mutex.Lock()
	since, ok := t.data[id]
	if !ok || since == nil {
		since = &time.Time{}
		t.data[id] = since
	}
	t.mutex.Unlock()
	return since
}
