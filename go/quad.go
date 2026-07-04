package oxigraph

// Quad is an RDF triple in a graph, mirroring pyoxigraph's Quad. The
// fields hold already-validated terms, so they are exported; two quads are
// equal exactly when == reports true. The zero value is not a valid quad:
// methods panic on nil Subject or Object, so construct quads with NewQuad.
// A struct literal that leaves GraphName nil renders like a default-graph
// quad but does not compare equal to one whose GraphName is
// DefaultGraph().
type Quad struct {
	Subject   Subject
	Predicate NamedNode
	Object    Term
	GraphName GraphName
}

// NewQuad assembles a quad from its four components. A nil graphName means
// the default graph, mirroring pyoxigraph's Quad(..., graph_name=None).
func NewQuad(subject Subject, predicate NamedNode, object Term, graphName GraphName) Quad {
	if graphName == nil {
		graphName = DefaultGraph()
	}
	return Quad{Subject: subject, Predicate: predicate, Object: object, GraphName: graphName}
}

// Triple returns the quad's triple, dropping the graph name, mirroring
// pyoxigraph's Quad.triple.
func (q Quad) Triple() Triple {
	return Triple{Subject: q.Subject, Predicate: q.Predicate, Object: q.Object}
}

// String returns the quad in N-Quads syntax without a terminating dot; the
// graph name is omitted when it is the default graph, matching
// pyoxigraph's str(Quad).
func (q Quad) String() string {
	s := q.Triple().String()
	if q.GraphName == nil || q.GraphName == GraphName(DefaultGraph()) {
		return s
	}
	return s + " " + q.GraphName.String()
}
