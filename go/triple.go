package oxigraph

// Triple is an RDF triple, mirroring pyoxigraph's Triple. The fields hold
// already-validated terms, so they are exported; two triples are equal
// exactly when == reports true. The zero value is not a valid triple:
// methods panic on nil terms, so construct triples with NewTriple.
type Triple struct {
	Subject   Subject
	Predicate NamedNode
	Object    Term
}

// NewTriple assembles a triple from its three components.
func NewTriple(subject Subject, predicate NamedNode, object Term) Triple {
	return Triple{Subject: subject, Predicate: predicate, Object: object}
}

// String returns the three terms in N-Triples syntax, space-separated
// without a terminating dot, matching pyoxigraph's str(Triple).
func (t Triple) String() string {
	return t.Subject.String() + " " + t.Predicate.String() + " " + t.Object.String()
}
