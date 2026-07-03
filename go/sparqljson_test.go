package oxigraph

import (
	"errors"
	"testing"
)

func TestParseSPARQLJSONTerm(t *testing.T) {
	alice := mustNamedNode(t, "http://example.com/person/alice")
	b0, _ := NewBlankNode("b0")
	alicia, _ := NewLanguageTaggedLiteral("Alicia", "es")
	integer := mustNamedNode(t, "http://www.w3.org/2001/XMLSchema#integer")

	cases := []struct {
		json string
		want Term
	}{
		{`{"type": "uri", "value": "http://example.com/person/alice"}`, alice},
		{`{"type": "bnode", "value": "b0"}`, b0},
		{`{"type": "literal", "value": "Alice"}`, NewLiteral("Alice")},
		{`{"type": "literal", "value": "Alicia", "xml:lang": "es"}`, alicia},
		{`{"type": "literal", "value": "42", "datatype": "http://www.w3.org/2001/XMLSchema#integer"}`, NewTypedLiteral("42", integer)},
		// The legacy alias and the implicit xsd:string datatype.
		{`{"type": "typed-literal", "value": "42", "datatype": "http://www.w3.org/2001/XMLSchema#integer"}`, NewTypedLiteral("42", integer)},
		{`{"type": "literal", "value": "x", "datatype": "http://www.w3.org/2001/XMLSchema#string"}`, NewLiteral("x")},
	}
	for _, c := range cases {
		got, err := ParseSPARQLJSONTerm([]byte(c.json))
		if err != nil {
			t.Errorf("ParseSPARQLJSONTerm(%s): %v", c.json, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseSPARQLJSONTerm(%s) = %v, want %v", c.json, got, c.want)
		}
	}
}

func TestParseSPARQLJSONTermErrors(t *testing.T) {
	cases := []struct {
		json string
		want error
	}{
		{`{"type": "graph", "value": "http://example.com/library"}`, ErrUnsupportedTermType},
		{`{"type": "uri"}`, ErrMalformedTerm},
		{`{"value": "x"}`, ErrMalformedTerm},
		{`not json`, ErrMalformedTerm},
		{`{"type": "uri", "value": "person/alice"}`, ErrInvalidIRI},
		{`{"type": "bnode", "value": "invoice 42"}`, ErrInvalidBlankNodeID},
		{`{"type": "literal", "value": "x", "xml:lang": "en_US"}`, ErrInvalidLanguageTag},
		{`{"type": "literal", "value": "x", "xml:lang": "en", "datatype": "http://example.com/dt"}`, ErrMalformedTerm},
	}
	for _, c := range cases {
		_, err := ParseSPARQLJSONTerm([]byte(c.json))
		if !errors.Is(err, c.want) {
			t.Errorf("ParseSPARQLJSONTerm(%s) error = %v, want %v", c.json, err, c.want)
		}
	}
}
