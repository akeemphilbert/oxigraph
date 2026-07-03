package oxigraph

import "testing"

func testQuad(t *testing.T) Quad {
	t.Helper()
	title, _ := NewLanguageTaggedLiteral("Le Petit Prince", "fr")
	return NewQuad(
		mustNamedNode(t, "http://example.com/book/1"),
		mustNamedNode(t, "http://purl.org/dc/terms/title"),
		title,
		mustNamedNode(t, "http://example.com/library"),
	)
}

func TestQuadString(t *testing.T) {
	q := testQuad(t)
	want := `<http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr <http://example.com/library>`
	if got := q.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestQuadDefaultGraphString(t *testing.T) {
	q := testQuad(t)
	q.GraphName = DefaultGraph()
	want := `<http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr`
	if got := q.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestNewQuadNilGraphName(t *testing.T) {
	q := testQuad(t)
	withNil := NewQuad(q.Subject, q.Predicate, q.Object, nil)
	if withNil.GraphName != GraphName(DefaultGraph()) {
		t.Error("a nil graph name must mean the default graph")
	}
}

func TestQuadTriple(t *testing.T) {
	q := testQuad(t)
	tr := q.Triple()
	if tr != NewTriple(q.Subject, q.Predicate, q.Object) {
		t.Error("Triple() must drop only the graph name")
	}
	want := `<http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr`
	if got := tr.String(); got != want {
		t.Errorf("Triple().String() = %q, want %q", got, want)
	}
}

func TestQuadEquality(t *testing.T) {
	a, b := testQuad(t), testQuad(t)
	if a != b {
		t.Error("identical quads must be equal")
	}
	b.GraphName = mustNamedNode(t, "http://example.com/archive")
	if a == b {
		t.Error("quads differing by graph name must not be equal")
	}
}

func TestDefaultGraph(t *testing.T) {
	if DefaultGraph() != DefaultGraph() {
		t.Error("default graph values must be equal")
	}
	if got := DefaultGraph().String(); got != "DEFAULT" {
		t.Errorf("String() = %q, want DEFAULT", got)
	}
}
