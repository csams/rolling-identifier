package inventory

import (
	"fmt"
	"sync"

	"github.com/csams/rolling-identifier/pkg/trie"
)

type Key = trie.Key

// System is a standin for anything we want to track
type System struct {
	Id   int  // this is actually the stable primary key other databases would depend on
	Data int
}

// Inventory is a respository of Systems
type Inventory interface {
	Create(Key, System) error
	Get(Key) (System, bool)
	Update(Key, Key, System) error
}

// MemoryInventory just stores inventory in a map
type MemoryInventory struct {
	sync.RWMutex
	Counter int
	Store   map[Key]System
}

func New() Inventory {
	return &MemoryInventory{
		Store: map[Key]System{},
	}
}

func (inv *MemoryInventory) Create(k Key, sys System) error {
	inv.Lock()
	defer inv.Unlock()
	if _, found := inv.Store[k]; !found {
        sys.Id = inv.Counter
        inv.Counter += 1
		inv.Store[k] = sys
		return nil
	}
	return fmt.Errorf("Key already exists: %s", k)
}

func (inv *MemoryInventory) Update(prev, cur Key, sys System) error {
	inv.Lock()
	defer inv.Unlock()
	if obj, found := inv.Store[prev]; found {
        delete(inv.Store, prev)
		inv.Store[cur] = sys
        sys.Id = obj.Id
		return nil
	}
	return fmt.Errorf("Key doesn't exist: %s", prev)
}

func (inv *MemoryInventory) Get(k Key) (System, bool) {
	inv.RLock()
	defer inv.RUnlock()
	s, found := inv.Store[k]
	return s, found
}
