package oxigraph

import (
	"errors"
	"path/filepath"
	"testing"
)

func testQuadWith(t *testing.T, object Term, graph GraphName) Quad {
	t.Helper()
	return NewQuad(
		mustNamedNode(t, "http://example.com/book/1"),
		mustNamedNode(t, "http://purl.org/dc/terms/title"),
		object,
		graph,
	)
}

func askTrue(t *testing.T, store *Store, query string) bool {
	t.Helper()
	results, err := store.Query(query)
	if err != nil {
		t.Fatalf("Query(%s): %v", query, err)
	}
	return results.Bool
}

func TestAddThenQuery(t *testing.T) {
	store := mustInMemoryStore(t)
	quad := testQuadWith(t, NewLiteral("Le Petit Prince"), nil)
	if err := store.Add(quad); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !askTrue(t, store, `ASK { <http://example.com/book/1> ?p ?o }`) {
		t.Error("the added quad must be visible to queries")
	}
	// Adding the same quad twice keeps set semantics.
	if err := store.Add(quad); err != nil {
		t.Fatalf("second Add: %v", err)
	}
	results, err := store.Query(`SELECT (COUNT(*) AS ?n) WHERE { ?s ?p ?o }`)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	n, _ := results.Solutions[0].Get("n")
	if got := n.(Literal).Value(); got != "1" {
		t.Errorf("quad count = %s, want 1", got)
	}
}

func TestAddToNamedGraph(t *testing.T) {
	store := mustInMemoryStore(t)
	library := mustNamedNode(t, "http://example.com/library")
	if err := store.Add(testQuadWith(t, NewLiteral("Le Petit Prince"), library)); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !askTrue(t, store, `ASK { GRAPH <http://example.com/library> { ?s ?p ?o } }`) {
		t.Error("the quad must be in the named graph")
	}
	if askTrue(t, store, `ASK { ?s ?p ?o }`) {
		t.Error("the default graph must stay empty")
	}
}

func TestRemove(t *testing.T) {
	store := mustInMemoryStore(t)
	quad := testQuadWith(t, NewLiteral("Le Petit Prince"), nil)
	if err := store.Add(quad); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := store.Remove(quad); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if askTrue(t, store, `ASK { ?s ?p ?o }`) {
		t.Error("the removed quad must not be visible")
	}
	// Removing an absent quad is a silent no-op.
	if err := store.Remove(quad); err != nil {
		t.Errorf("Remove of an absent quad = %v, want nil", err)
	}
}

func TestBlankNodeRoundTrip(t *testing.T) {
	store := mustInMemoryStore(t)
	author, err := NewBlankNode("author1")
	if err != nil {
		t.Fatal(err)
	}
	quad := NewQuad(author, mustNamedNode(t, "http://xmlns.com/foaf/0.1/name"),
		NewLiteral("Antoine de Saint-Exupéry"), nil)
	if err := store.Add(quad); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := store.Remove(quad); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if askTrue(t, store, `ASK { ?s ?p ?o }`) {
		t.Error("blank node identity must survive the boundary so the remove matches the add")
	}
}

func TestUpdate(t *testing.T) {
	store := mustInMemoryStore(t)
	err := store.Update(`INSERT DATA { <http://example.com/s> <http://example.com/p> "v" }`)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !askTrue(t, store, `ASK { <http://example.com/s> ?p ?o }`) {
		t.Error("inserted data must be visible")
	}
	err = store.Update(`DELETE DATA { <http://example.com/s> <http://example.com/p> "v" }`)
	if err != nil {
		t.Fatalf("Update delete: %v", err)
	}
	if askTrue(t, store, `ASK { ?s ?p ?o }`) {
		t.Error("deleted data must not be visible")
	}
}

func TestUpdateSyntaxError(t *testing.T) {
	store := mustInMemoryStore(t)
	if err := store.Update("INSRT DATA { }"); !errors.Is(err, ErrSyntax) {
		t.Errorf("malformed update error = %v, want ErrSyntax", err)
	}
}

func TestWriteDurability(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book-catalog")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := store.Add(testQuadWith(t, NewLiteral("Le Petit Prince"), nil)); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	if !askTrue(t, reopened, `ASK { <http://example.com/book/1> ?p ?o }`) {
		t.Error("on-disk data must survive close and reopen")
	}
}

func TestWritesOnClosedStore(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	store.Close()
	quad := testQuadWith(t, NewLiteral("x"), nil)
	if err := store.Add(quad); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("Add = %v, want ErrStoreClosed", err)
	}
	if err := store.Remove(quad); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("Remove = %v, want ErrStoreClosed", err)
	}
	if err := store.Update("INSERT DATA { }"); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("Update = %v, want ErrStoreClosed", err)
	}
}
