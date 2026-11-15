package livecoding

import (
	"reflect"
	"testing"
)

// §4.1 — arrays are VALUES in Go: assignment copies everything.
func TestPrimitiveArrayCopy(t *testing.T) {
	a := [3]int{1, 2, 3}
	b := a
	b[0] = 99

	if a[0] != 1 {
		t.Errorf("a was mutated: %v (arrays must be copied by value)", a)
	}
	if !reflect.DeepEqual(a, [3]int{1, 2, 3}) {
		t.Errorf("a = %v, want [1 2 3]", a)
	}
}

// §4.1 — struct array copy: values are duplicated, but pointer fields inside
// stay shared. Mutating the pointee through the copy is visible in the original.
func TestStructArrayCopyShallow(t *testing.T) {
	type person struct {
		name   string
		friend *string
	}

	alice := "alice"
	original := [2]person{{name: "p1", friend: &alice}, {name: "p2"}}
	copied := original

	copied[0].name = "renamed"      // value field: copy is independent
	*copied[0].friend = "mallory"   // pointer field: SHARED backing storage

	if original[0].name != "p1" {
		t.Errorf("name field leaked through copy: %q", original[0].name)
	}
	if alice != "mallory" {
		t.Errorf("pointee mutation invisible: %q (pointer fields are shared)", alice)
	}
}

// §4.1 — value receiver mutates a copy; pointer receiver mutates the original.
type badge struct{ level int }

func (b badge) bumpValue()    { b.level++ } // acts on a copy — no lasting effect
func (b *badge) bumpPointer() { b.level++ } // acts on the original

func TestStructReceiverSemantics(t *testing.T) {
	b := badge{level: 1}

	b.bumpValue()
	if b.level != 1 {
		t.Errorf("value receiver mutated the original: %d", b.level)
	}

	b.bumpPointer()
	if b.level != 2 {
		t.Errorf("pointer receiver did not mutate the original: %d", b.level)
	}
}

// §4.1 — a slice "copy" shares the backing array; mutations leak.
func TestSliceCopySharesBackingArray(t *testing.T) {
	a := []int{1, 2, 3}
	b := a // NO copy — same backing array
	b[0] = 99
	if a[0] != 99 {
		t.Errorf("expected shared backing array, a = %v", a)
	}

	// Real copy: either copy() or append trick.
	c := append([]int(nil), a...)
	c[0] = 1
	if a[0] != 99 {
		t.Errorf("copy() should have been independent, a = %v", a)
	}
}
