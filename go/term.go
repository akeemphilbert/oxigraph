package oxigraph

import "fmt"

// Term is any RDF term that may appear in the object position of a triple
// or quad: NamedNode, BlankNode or Literal. It is sealed: only types in
// this package implement it. Every implementation is a comparable value,
// so two terms are equal exactly when == reports true.
type Term interface {
	fmt.Stringer
	isTerm()
}

// Subject is an RDF term valid in the subject position: NamedNode or
// BlankNode. RDF forbids literal subjects, and so does the compiler.
type Subject interface {
	Term
	isSubject()
}

// GraphName is an RDF term valid in the graph-name position of a quad:
// NamedNode, BlankNode or DefaultGraphTerm.
type GraphName interface {
	fmt.Stringer
	isGraphName()
}

// DefaultGraphTerm designates a store's default graph, mirroring
// pyoxigraph's DefaultGraph. Obtain it with DefaultGraph.
type DefaultGraphTerm struct{}

// DefaultGraph returns the graph name designating a store's default graph.
func DefaultGraph() DefaultGraphTerm { return DefaultGraphTerm{} }

// String returns "DEFAULT", the SPARQL keyword pyoxigraph uses for
// str(DefaultGraph()).
func (DefaultGraphTerm) String() string { return "DEFAULT" }

func (DefaultGraphTerm) isGraphName() {}
