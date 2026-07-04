package oxigraph

import "github.com/akeemphilbert/oxigraph/go/internal/iri"

// NamedNode is an RDF IRI term, mirroring pyoxigraph's NamedNode. The zero
// value is not a valid term; build named nodes with NewNamedNode.
type NamedNode struct {
	iri string
}

// NewNamedNode creates a named node from an absolute IRI, validated
// against RFC 3987. The returned error matches ErrInvalidIRI.
func NewNamedNode(value string) (NamedNode, error) {
	if err := iri.Validate(value); err != nil {
		return NamedNode{}, &ParseError{Kind: ErrInvalidIRI, Input: value, Detail: err.Error()}
	}
	return NamedNode{iri: value}, nil
}

// Value returns the IRI, mirroring pyoxigraph's NamedNode.value.
func (n NamedNode) Value() string { return n.iri }

// String returns the N-Quads form "<iri>".
func (n NamedNode) String() string { return "<" + n.iri + ">" }

func (NamedNode) isTerm()      {}
func (NamedNode) isSubject()   {}
func (NamedNode) isGraphName() {}
