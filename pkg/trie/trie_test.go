package trie

import (
    "reflect"
    "testing"
)

func TestKeyComponents(t *testing.T) {
    key := ""
    expected := []string{""}
    got := KeyComponents(key)

    if ! reflect.DeepEqual(expected, got) {
        t.Fatalf("%+v", got)
    }

    key = "a"
    expected = []string{"a"}
    got = KeyComponents(key)

    if ! reflect.DeepEqual(expected, got) {
        t.Fatalf("%+v", got)
    }

    key = "a:b"
    expected = []string{"a", "b"}
    got = KeyComponents(key)

    if ! reflect.DeepEqual(expected, got) {
        t.Fatalf("%+v", got)
    }
}

func TestExtend(t *testing.T) {
    root := New[int]()
    cur := root.Extend([]string{"a"}, 0)

    if _, ok := root.Children["a"]; !ok {
        t.Fatal("Expected a to exist")
    }

    cur.Extend([]string{"b"}, 1)
    if _, ok := cur.Children["b"]; !ok {
        t.Fatal("Expected a to exist")
    }

    if root.Children["a"].Children["b"].Value != 1 {
        t.Fatal("Expected 1 to exist")
    }

    cur.Extend([]string{"c"}, 2)
    if root.Children["a"].Children["c"].Value != 2 {
        t.Fatal("Expected 2 to exist")
    }

    root.Extend([]string{"d", "e", "f"}, 3)
    if root.Children["d"].Children["e"].Children["f"].Value != 3 {
        t.Fatal("Expected 3 to exist")
    }
}

func TestWithLongestPrefix(t *testing.T) {
    root := New[int]()

    root.WithLongestPrefix("a", func(comps []Key, pos *Trie[int]) error {
        if pos != root {
            t.Fatalf("%+v", pos)
        }
        if len(comps) != 1 {
            t.Fatalf("%+v", comps)
        }

        return nil
    })

    cur := root.Extend([]string{"a"}, 0)
    cur.Extend([]string{"b"}, 1)
    cur.Extend([]string{"c"}, 2)
    root.Extend([]string{"d", "e", "f"}, 3)

    root.WithLongestPrefix("a", func(comps []Key, pos *Trie[int]) error {
        if len(comps) != 0 {
            t.Fatalf("%+v", comps)
        }

        if len(pos.Children) != 2 {
            t.Fatalf("%+v", pos)
        }

        return nil
    })

    root.WithLongestPrefix("a:c", func(comps []Key, pos *Trie[int]) error {
        if len(comps) != 0 {
            t.Fatalf("%+v", comps)
        }

        if len(pos.Children) != 0 {
            t.Fatalf("%+v", pos)
        }

        return nil
    })

    root.WithLongestPrefix("a:d", func(comps []Key, pos *Trie[int]) error {
        if len(comps) != 1 && comps[0] == "d" {
            t.Fatalf("%+v", comps)
        }

        if len(pos.Children) != 2 {
            t.Fatalf("%+v", pos)
        }

        return nil
    })
}
