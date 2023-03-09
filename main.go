package main

/*
[[202303070850|A possible solution to the duplicate id problem]]
*/

import (
	"fmt"
	"time"

	"github.com/csams/rolling-identifier/pkg/ingress"
	"github.com/csams/rolling-identifier/pkg/inventory"
	"github.com/csams/rolling-identifier/pkg/storage"
	"github.com/csams/rolling-identifier/pkg/trie"
)

func main() {
	announce := func(r ingress.Receipt) { fmt.Printf("Hey ya'll! Process this archive: %s\n", r) }
	inv := inventory.New()
	index := trie.New[time.Time]()
	store := storage.New[ingress.Payload]()

	ing := ingress.New(1, inv, index, store, announce)

	req := ingress.Request{
		Payload: inventory.System{},
	}

	resp := ing.CheckIn("1", req)
	fmt.Printf("%+v\n", resp)
	fmt.Printf("%s\n", index)

	resp = ing.CheckIn("1", req)
	fmt.Printf("%+v\n", resp)
	fmt.Printf("%s\n", index)

	if !resp.ComeBack {
		fmt.Println("Expected a come back message")
		return
	}

	// go back with the new key we were issued and the receipt
	resp = ing.CheckIn(resp.Key, ingress.Request{Receipt: resp.Receipt})
	fmt.Printf("%+v\n", resp)
	fmt.Printf("%s\n", index)

	// go back again with the same key but delaying for 2 seconds so we're outside the window of suspicion.
	time.Sleep(2 * time.Second)
	resp = ing.CheckIn(resp.Key, ingress.Request{Payload: inventory.System{}})
	fmt.Printf("%+v\n", resp)
	fmt.Printf("%s\n", index)
}
