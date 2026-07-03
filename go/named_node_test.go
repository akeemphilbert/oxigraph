package oxigraph

import (
	"errors"
	"testing"
)

func TestNewNamedNode(t *testing.T) {
	n, err := NewNamedNode("http://example.com/person/alice")
	if err != nil {
		t.Fatalf("NewNamedNode: %v", err)
	}
	if got := n.Value(); got != "http://example.com/person/alice" {
		t.Errorf("Value() = %q", got)
	}
	if got := n.String(); got != "<http://example.com/person/alice>" {
		t.Errorf("String() = %q", got)
	}
}

func TestNewNamedNodeInvalid(t *testing.T) {
	for _, iri := range []string{"", "person/alice", "http://example.com/a b"} {
		_, err := NewNamedNode(iri)
		if !errors.Is(err, ErrInvalidIRI) {
			t.Errorf("NewNamedNode(%q) error = %v, want ErrInvalidIRI", iri, err)
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) || parseErr.Input != iri {
			t.Errorf("NewNamedNode(%q) error should be a *ParseError carrying the input", iri)
		}
	}
}

func TestNamedNodeEquality(t *testing.T) {
	a := mustNamedNode(t, "http://example.com/a")
	b := mustNamedNode(t, "http://example.com/a")
	c := mustNamedNode(t, "http://example.com/c")
	if a != b {
		t.Error("identical named nodes must be equal")
	}
	if a == c {
		t.Error("different named nodes must not be equal")
	}
	if Term(a) != Term(b) {
		t.Error("equality must hold through the Term interface")
	}
}
