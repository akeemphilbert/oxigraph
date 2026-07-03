package oxigraph

import (
	"errors"
	"testing"
)

func TestNewBlankNode(t *testing.T) {
	b, err := NewBlankNode("author1")
	if err != nil {
		t.Fatalf("NewBlankNode: %v", err)
	}
	if got := b.Value(); got != "author1" {
		t.Errorf("Value() = %q", got)
	}
	if got := b.String(); got != "_:author1" {
		t.Errorf("String() = %q", got)
	}
	// Dots are allowed inside an identifier but not at its end.
	if _, err := NewBlankNode("a.b"); err != nil {
		t.Errorf("NewBlankNode(\"a.b\") = %v, want nil", err)
	}
}

func TestNewBlankNodeInvalid(t *testing.T) {
	for _, id := range []string{"", "invoice 42", "-order", "ends.", "\xff"} {
		_, err := NewBlankNode(id)
		if !errors.Is(err, ErrInvalidBlankNodeID) {
			t.Errorf("NewBlankNode(%q) error = %v, want ErrInvalidBlankNodeID", id, err)
		}
	}
}

func TestNewBlankNodeRandom(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		b := NewBlankNodeRandom()
		id := b.Value()
		if id == "" {
			t.Fatal("random blank node has an empty identifier")
		}
		if id[0] < 'a' || id[0] > 'f' {
			t.Fatalf("random identifier %q must start with a hex letter", id)
		}
		for j := 0; j < len(id); j++ {
			c := id[j]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Fatalf("random identifier %q is not lowercase hex", id)
			}
		}
		if seen[id] {
			t.Fatalf("random identifier %q repeated", id)
		}
		seen[id] = true
	}
}
