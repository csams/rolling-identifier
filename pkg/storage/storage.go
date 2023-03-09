package storage

import "sync"

type Id = string

type Storage[V any] interface {
    Put(Id, V) error
    Get(Id) (V, bool, error)
}

type MemoryStorage[V any] struct {
    sync.RWMutex
    Store map[Id]V
}

func New[V any]() Storage[V] {
    return &MemoryStorage[V] {
        Store: map[Id]V{},
    }
}

func (s *MemoryStorage[V]) Put(id Id, v V) error {
    s.Lock()
    defer s.Unlock()
    s.Store[id] = v
    return nil
}

func (s *MemoryStorage[V]) Get(id Id) (V, bool, error) {
    s.RLock()
    defer s.RUnlock()
    obj, found := s.Store[id]
    return obj, found, nil
}
