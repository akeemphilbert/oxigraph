package oxigraph

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadTurtle(t *testing.T) {
	store := mustInMemoryStore(t)
	turtle := `@prefix dcterms: <http://purl.org/dc/terms/> .
<http://example.com/book/1> dcterms:title "Le Petit Prince"@fr ;
    dcterms:publisher "Gallimard" .`
	if err := store.Load(strings.NewReader(turtle), Turtle); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !askTrue(t, store, `ASK { <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr }`) {
		t.Error("loaded triples must be queryable")
	}
}

func TestLoadNQuadsNamedGraph(t *testing.T) {
	store := mustInMemoryStore(t)
	nquads := `<http://example.com/book/3> <http://purl.org/dc/terms/title> "Vol de Nuit" <http://example.com/library> .`
	if err := store.Load(strings.NewReader(nquads), NQuads); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !askTrue(t, store, `ASK { GRAPH <http://example.com/library> { ?s ?p ?o } }`) {
		t.Error("the quad must land in its named graph")
	}
}

func TestLoadMalformedIsAtomic(t *testing.T) {
	store := mustInMemoryStore(t)
	broken := `<http://example.com/book/1> <http://purl.org/dc/terms/title> "ok" .
<http://example.com/book/2> <http://purl.org/dc/terms/title> .`
	err := store.Load(strings.NewReader(broken), Turtle)
	if !errors.Is(err, ErrSyntax) {
		t.Fatalf("Load error = %v, want ErrSyntax", err)
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("error %q must mention the line", err)
	}
	if askTrue(t, store, `ASK { ?s ?p ?o }`) {
		t.Error("a failed load must not leave partial data")
	}
}

func TestLoadUnsupportedFormat(t *testing.T) {
	store := mustInMemoryStore(t)
	if err := store.Load(strings.NewReader(""), RdfFormat(99)); !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Load error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestLoadReaderError(t *testing.T) {
	store := mustInMemoryStore(t)
	boom := errors.New("boom")
	if err := store.Load(&failingReader{err: boom}, Turtle); !errors.Is(err, boom) {
		t.Errorf("Load error = %v, want the reader's error", err)
	}
}

type failingReader struct{ err error }

func (r *failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestDumpRoundTrip(t *testing.T) {
	store := mustInMemoryStore(t)
	trig := `@prefix dcterms: <http://purl.org/dc/terms/> .
<http://example.com/book/1> dcterms:title "Le Petit Prince"@fr .
GRAPH <http://example.com/library> {
  <http://example.com/book/2> dcterms:title "Watership Down"@en .
}`
	if err := store.Load(strings.NewReader(trig), TriG); err != nil {
		t.Fatalf("Load: %v", err)
	}
	var dump strings.Builder
	if err := store.Dump(&dump, NQuads); err != nil {
		t.Fatalf("Dump: %v", err)
	}
	reloaded := mustInMemoryStore(t)
	if err := reloaded.Load(strings.NewReader(dump.String()), NQuads); err != nil {
		t.Fatalf("reload: %v", err)
	}
	results, err := reloaded.Query(`SELECT (COUNT(*) AS ?n) WHERE { { ?s ?p ?o } UNION { GRAPH ?g { ?s ?p ?o } } }`)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	n, _ := results.Solutions[0].Get("n")
	if got := n.(Literal).Value(); got != "2" {
		t.Errorf("reloaded quad count = %s, want 2", got)
	}
}

func TestDumpTriplesOnlyFormatRejected(t *testing.T) {
	store := mustInMemoryStore(t)
	var dump strings.Builder
	err := store.Dump(&dump, Turtle)
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Dump error = %v, want ErrUnsupportedFormat", err)
	}
	if err == nil || !strings.Contains(err.Error(), "Turtle") {
		t.Errorf("error %v must name the format", err)
	}
}

func TestIOOnClosedStore(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	store.Close()
	if err := store.Load(strings.NewReader(""), Turtle); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("Load = %v, want ErrStoreClosed", err)
	}
	var dump strings.Builder
	if err := store.Dump(&dump, NQuads); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("Dump = %v, want ErrStoreClosed", err)
	}
}

func TestRdfFormatString(t *testing.T) {
	names := map[RdfFormat]string{Turtle: "Turtle", NTriples: "N-Triples", NQuads: "N-Quads", TriG: "TriG"}
	for format, want := range names {
		if got := format.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	}
}
