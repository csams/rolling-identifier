package ingress

import (
	"fmt"
	"time"

	"github.com/csams/rolling-identifier/pkg/inventory"
	"github.com/csams/rolling-identifier/pkg/storage"
	"github.com/csams/rolling-identifier/pkg/trie"
	"github.com/google/uuid"
)

// type aliases
type Receipt = storage.Id
type Payload = inventory.System
type Inventory = inventory.Inventory
type AnnounceFunc func(Receipt)
type Index = *trie.Trie[time.Time]
type IndexNode = Index
type Storage = storage.Storage[Payload]

// Request contains either a payload or a Receipt. It shouldn't contain both.
type Request struct {
	Payload inventory.System
	Receipt Receipt
}

// Response tells the client whether it should update its local key and come back with the receipt
type Response struct {
	ComeBack bool
	Receipt  Receipt
	Key      trie.Key
	Error    error
}

// Ingress is what a client would use to check in, providing its current key and a request
type Ingress interface {
	CheckIn(trie.Key, Request) Response
}

type IngressImpl struct {
	EnoughSeconds float64

	Inventory          Inventory
	Index              Index
	S3                 storage.Storage[Payload]
	AnnounceNewArchive AnnounceFunc
}

func New(enoughSeconds float64, inv Inventory, ind Index, st Storage, announce AnnounceFunc) Ingress {
	return &IngressImpl{
		EnoughSeconds:      enoughSeconds,
		Inventory:          inv,
		Index:              ind,
		S3:                 st,
		AnnounceNewArchive: announce,
	}
}

func (ing *IngressImpl) CheckIn(k trie.Key, req Request) Response {
	now := time.Now()
	receipt := uuid.New().String()
	resp := Response{
		Key:     k,
		Receipt: receipt,
	}
	resp.Error = ing.Index.WithLongestPrefix(k, func(remainder []trie.Key, pos IndexNode) error {
		var err error

		// we need to make a new system
		newSystem := func(r Receipt) {
			fmt.Println("newSystem")
			ing.Inventory.Create(k, req.Payload)
			ing.S3.Put(r, req.Payload)
			ing.AnnounceNewArchive(r)

			resp.Receipt = r
		}

		// regular old checkin
		checkIn := func(r Receipt) {
			fmt.Println("checkIn")
			ing.Inventory.Update(k, k, req.Payload)
			ing.S3.Put(r, req.Payload)
			ing.AnnounceNewArchive(r)

			resp.Receipt = r
		}

		// if we tell a client to come back, we go ahead and store the archive it sent so it doesn't have to post it again
		comeBack := func(r Receipt) {
			fmt.Println("comeBack")
			ing.S3.Put(r, req.Payload)

			resp.ComeBack = true
			resp.Key = trie.ExtendKey(k, r) // just reusing the storage receipt - could be anything unique
			resp.Receipt = r
		}

		// the client came back with a receipt
		cameBack := func(r Receipt) {
			fmt.Println("cameBack")
			prevId := trie.TrimKeySuffix(k, trie.NewKey(remainder))
			payload, found, e := ing.S3.Get(r)
			err = e
			if found {
				ing.Inventory.Update(prevId, k, payload)
				ing.AnnounceNewArchive(r)
			}
			resp.Receipt = r
		}

		// none of the system's history is in the index (the longest common prefix was empty)
		if pos == ing.Index {
			_, found := ing.Inventory.Get(k)
			pos.Extend(remainder, now)
			if found { // the key is in inventory, though. Freshen the index and do a routine checkin.
				checkIn(receipt)
			} else { // it's not in the inventory either.. must be a new system.
				newSystem(receipt)
			}
		} else { // at least part of its history is in the index
			if len(remainder) == 0 { // we found the entire key
				if len(pos.Children) == 0 { // the node in the trie has no children, so it's an exact match
					if time.Since(pos.Value).Seconds() < ing.EnoughSeconds { // it hasn't been long enough since we saw it last..
						comeBack(receipt)
					} else { // it's been long enough. Assume it's a routine check-in.
						pos.Value = now
						checkIn(receipt)
					}
				} else { // we found the key, but it's already been extended. Either a backup or some other clone is trying to check in.
					comeBack(receipt)
				}
			} else { // not all key components are found
				pos.Extend(remainder, now)
				if len(pos.Children) == 0 { // the system was previously given a new id and told to come back
					cameBack(req.Receipt)
				} else { // we found a common prefix but have diverged - this is a clone.
					newSystem(req.Receipt)
				}
			}
		}
		return err
	})
	return resp
}
