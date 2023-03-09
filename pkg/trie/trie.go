package trie

import (
	"fmt"
	"strings"
)

type Callback[V any] func([]Key, *Trie[V]) error

type Key = string
type KeyComponent = Key

func ExtendKey(base, suffix Key) Key {
	return strings.Join([]string{base, suffix}, ":")
}

func KeyComponents(k Key) []KeyComponent {
	return strings.Split(k, ":")
}

func NewKey(comps []KeyComponent) Key {
	return strings.Join(comps, ":")
}

// TrimKeySuffix just hides that we need to account for the fact that keys are joined by ":"
// It'll take suffix from k and then remove the trailing ":".
func TrimKeySuffix(k, suffix Key) Key {
	if len(k) == len(suffix) {
		return ""
	}
	fmt.Printf("%s %s\n", k, suffix)
	trimLen := len(suffix) + 1
	return k[:len(k)-trimLen]
}

// Trie is a simple struct to represent a tree structure where the children are indexed by a map
type Trie[V any] struct {
	Value    V
	Children map[Key]*Trie[V]
}

// Simple visualization of the trie
func (t Trie[V]) String() string {
	builder := strings.Builder{}
	builder.WriteString("\n")
	var inner func(c Trie[V], depth int)
	inner = func(c Trie[V], depth int) {
		prefix := strings.Repeat("┃  ", depth)
		for k, v := range c.Children {
			builder.WriteString(prefix)
			builder.WriteString(fmt.Sprintf("%s: %+v", k, v.Value))
			builder.WriteString("\n")
			inner(*v, depth+1)
		}
	}

	inner(t, 0)
	return builder.String()
}

// Return a new trie
func New[V any]() *Trie[V] {
	return &Trie[V]{
		Value:    *new(V),
		Children: map[Key]*Trie[V]{},
	}
}

// WithLongestPrefix accepts a key and will invoke a callback function with the remainder of the key's componets after
// walking down the trie, along with the node that represents the end of the path.
func (t *Trie[V]) WithLongestPrefix(k Key, cb Callback[V]) error {
	cur := t
	l := -1
	components := KeyComponents(k)
	for i, c := range components {
		y, found := cur.Children[c]
		if found {
			l = i
			cur = y
		} else {
			return cb(components[l+1:], cur)
		}
	}
	return cb(components[l+1:], cur)
}

// Extend will extend the tree downward for each component and assign v to the last node.
func (t *Trie[V]) Extend(comps []KeyComponent, v V) *Trie[V] {
	cur := t
	for _, c := range comps {
		n := New[V]()
		cur.Children[c] = n
		cur = n
	}
	cur.Value = v
	return cur
}
