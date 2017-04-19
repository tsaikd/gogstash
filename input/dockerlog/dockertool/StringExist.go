package dockertool

import (
	"sync"
)

// StringExist a struct to check string state
type StringExist struct {
	data  map[string]bool
	mutex *sync.RWMutex
}

// NewStringExist create a StringExist
func NewStringExist() StringExist {
	return StringExist{
		data:  map[string]bool{},
		mutex: &sync.RWMutex{},
	}
}

// Exist return true if id exist
func (t StringExist) Exist(id string) bool {
	t.mutex.RLock()
	_, exist := t.data[id]
	t.mutex.RUnlock()
	return exist
}

// Add id to data
func (t *StringExist) Add(id string) {
	t.mutex.Lock()
	t.data[id] = true
	t.mutex.Unlock()
}

// Remove id from data
func (t *StringExist) Remove(id string) {
	t.mutex.Lock()
	delete(t.data, id)
	t.mutex.Unlock()
}
