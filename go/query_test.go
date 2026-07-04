//go:build (darwin || linux) && (amd64 || arm64)

package oxigraph

import (
	"errors"
	"strings"
	"testing"
)

func mustInMemoryStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestQuerySelect(t *testing.T) {
	store := mustInMemoryStore(t)
	results, err := store.Query(`
		SELECT ?name ?age WHERE {
			VALUES (?name ?age) { ("Antoine" 44) ("Richard" 52) }
		} ORDER BY ?age`)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if results.Kind != QuerySolutions {
		t.Fatalf("Kind = %v, want QuerySolutions", results.Kind)
	}
	if len(results.Solutions) != 2 {
		t.Fatalf("got %d solutions, want 2", len(results.Solutions))
	}
	name, ok := results.Solutions[0].Get("name")
	if !ok || name != Term(NewLiteral("Antoine")) {
		t.Errorf(`first name = %v (%v), want "Antoine"`, name, ok)
	}
	integer := mustNamedNode(t, "http://www.w3.org/2001/XMLSchema#integer")
	age, ok := results.Solutions[1].Get("age")
	if !ok || age != Term(NewTypedLiteral("52", integer)) {
		t.Errorf("second age = %v (%v), want 52", age, ok)
	}
}

func TestQuerySelectUnbound(t *testing.T) {
	store := mustInMemoryStore(t)
	results, err := store.Query(`SELECT ?a ?b WHERE { VALUES (?a ?b) { ("x" UNDEF) } }`)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results.Solutions) != 1 {
		t.Fatalf("got %d solutions, want 1", len(results.Solutions))
	}
	if _, ok := results.Solutions[0].Get("b"); ok {
		t.Error("b must be unbound")
	}
}

func TestQueryAsk(t *testing.T) {
	store := mustInMemoryStore(t)
	for query, want := range map[string]bool{
		"ASK { FILTER(2 + 2 = 4) }": true,
		"ASK { ?s ?p ?o }":          false,
		"ASK { FILTER(1 > 2) }":     false,
	} {
		results, err := store.Query(query)
		if err != nil {
			t.Fatalf("Query(%s): %v", query, err)
		}
		if results.Kind != QueryBoolean || results.Bool != want {
			t.Errorf("Query(%s) = %v/%v, want boolean %v", query, results.Kind, results.Bool, want)
		}
	}
}

func TestQueryConstruct(t *testing.T) {
	store := mustInMemoryStore(t)
	results, err := store.Query(`
		CONSTRUCT { <http://example.com/s> <http://example.com/p> ?o }
		WHERE { VALUES ?o { "first" "second" } }`)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if results.Kind != QueryTriples || len(results.Triples) != 2 {
		t.Fatalf("Kind %v with %d triples, want QueryTriples with 2", results.Kind, len(results.Triples))
	}
	want := NewTriple(
		mustNamedNode(t, "http://example.com/s"),
		mustNamedNode(t, "http://example.com/p"),
		NewLiteral("first"),
	)
	if results.Triples[0] != want && results.Triples[1] != want {
		t.Errorf("constructed triples %v do not include %v", results.Triples, want)
	}
}

func TestQueryDescribeEmpty(t *testing.T) {
	store := mustInMemoryStore(t)
	results, err := store.Query(`DESCRIBE <http://example.com/book/1>`)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if results.Kind != QueryTriples || len(results.Triples) != 0 {
		t.Errorf("Kind %v with %d triples, want QueryTriples with 0", results.Kind, len(results.Triples))
	}
}

func TestQueryErrors(t *testing.T) {
	store := mustInMemoryStore(t)
	if _, err := store.Query("SELCT ?x WHERE {}"); !errors.Is(err, ErrSyntax) {
		t.Errorf("malformed query error = %v, want ErrSyntax", err)
	}
	_, err := store.Query(`SELECT ?x WHERE { VALUES ?x { 1 } FILTER(<http://example.com/nofn>(?x)) }`)
	if !errors.Is(err, ErrEvaluation) {
		t.Errorf("unknown function error = %v, want ErrEvaluation", err)
	}
	if err == nil || !strings.Contains(err.Error(), "http://example.com/nofn") {
		t.Errorf("error %v must carry the engine's message naming the function", err)
	}
}

func TestQueryClosedStore(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	store.Close()
	if _, err := store.Query("ASK { }"); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("Query on closed store = %v, want ErrStoreClosed", err)
	}
}

// BenchmarkQuerySmallSelect measures the per-call overhead of a small
// SELECT crossing the FFI (target: well under 1 ms per story #14).
func BenchmarkQuerySmallSelect(b *testing.B) {
	store, err := NewStore()
	if err != nil {
		b.Fatalf("NewStore: %v", err)
	}
	defer store.Close()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.Query(`SELECT ?x WHERE { VALUES ?x { 1 2 3 } }`); err != nil {
			b.Fatal(err)
		}
	}
}
