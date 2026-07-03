package oxigraph

import (
	"errors"
	"testing"
)

func TestParseNQuadsLineRoundTrips(t *testing.T) {
	lines := []string{
		`<http://example.com/book/1> <http://purl.org/dc/terms/publisher> "Gallimard" .`,
		`<http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr <http://example.com/library> .`,
		`<http://example.com/book/1> <http://example.com/pageCount> "96"^^<http://www.w3.org/2001/XMLSchema#integer> .`,
		`_:author1 <http://xmlns.com/foaf/0.1/name> "Antoine de Saint-Exupéry" .`,
		`<http://example.com/s> <http://example.com/p> _:o _:g .`,
	}
	for _, line := range lines {
		q, err := ParseNQuadsLine(line)
		if err != nil {
			t.Errorf("ParseNQuadsLine(%q): %v", line, err)
			continue
		}
		if got := q.NQuadsLine(); got != line {
			t.Errorf("round trip of %q produced %q", line, got)
		}
	}
}

func TestParseNQuadsLineDefaultGraph(t *testing.T) {
	q, err := ParseNQuadsLine(`<http://example.com/s> <http://example.com/p> "o" .`)
	if err != nil {
		t.Fatalf("ParseNQuadsLine: %v", err)
	}
	if q.GraphName != GraphName(DefaultGraph()) {
		t.Errorf("GraphName = %v, want the default graph", q.GraphName)
	}
}

func TestParseNQuadsLineDecodesEscapes(t *testing.T) {
	q, err := ParseNQuadsLine(`<http://example.com/s> <http://example.com/p> "a\"b\ncAd\U0001F600e" .`)
	if err != nil {
		t.Fatalf("ParseNQuadsLine: %v", err)
	}
	want := "a\"b\nc" + "Ad" + "\U0001F600" + "e"
	if got := q.Object.(Literal).Value(); got != want {
		t.Errorf("Value() = %q, want %q", got, want)
	}
}

func TestParseNQuadsLineMalformed(t *testing.T) {
	malformed := []string{
		``,
		`<http://example.com/book/1> <http://purl.org/dc/terms/title> .`,
		`<http://example.com/book/1> "Gallimard" <http://example.com/book/2> .`,
		`_:b <http://example.com/p> "x"`,
		`<http://example.com/s> <http://example.com/p> "x" . trailing`,
		`<http://example.com/s> <http://example.com/p> "x" "g" .`,
		`<http://example.com/s> <http://example.com/p> "unterminated .`,
		`<http://example.com/s> <http://example.com/p> "bad\q" .`,
		`<http://example.com/s`,                                        // unclosed IRIREF
		`<http://example.com/s> <http://example.com/p> "x"^^"y" .`,     // missing '<' after '^^'
		`<http://example.com/s> <http://example.com/p> "x\`,            // trailing backslash
		"<http://example.com/s> <http://example.com/p> \"a\nb\" .",     // raw line jump in a literal
		`<http://example.com/\tab> <http://example.com/p> "x" .`,       // ECHAR is illegal in an IRIREF
		`<http://example.com/s> <http://example.com/p> "\uD800" .`,     // lone surrogate escape
		`<http://example.com/s> <http://example.com/p> "\U00110000" .`, // beyond U+10FFFF
	}
	for _, line := range malformed {
		if _, err := ParseNQuadsLine(line); !errors.Is(err, ErrSyntax) {
			t.Errorf("ParseNQuadsLine(%q) error = %v, want ErrSyntax", line, err)
		}
	}
}

func TestParseNQuadsLinePropagatesTermErrors(t *testing.T) {
	if _, err := ParseNQuadsLine(`<relative/iri> <http://example.com/p> "x" .`); !errors.Is(err, ErrInvalidIRI) {
		t.Errorf("error = %v, want ErrInvalidIRI", err)
	}
	if _, err := ParseNQuadsLine(`<http://example.com/s> <http://example.com/p> "x"@123 .`); !errors.Is(err, ErrInvalidLanguageTag) {
		t.Errorf("error = %v, want ErrInvalidLanguageTag", err)
	}
}

func TestParseNQuadsLineValueRoundTrips(t *testing.T) {
	integer := mustNamedNode(t, "http://www.w3.org/2001/XMLSchema#integer")
	graph := mustNamedNode(t, "http://example.com/library")
	s := mustNamedNode(t, "http://example.com/s")
	p := mustNamedNode(t, "http://example.com/p")
	bnode, _ := NewBlankNode("g1")

	quads := []Quad{
		NewQuad(s, p, NewLiteral("line one\nline two"), nil),
		NewQuad(s, p, NewLiteral(`quote " and backslash \`), graph),
		NewQuad(s, p, NewLiteral("nul \x00 and ￾"), nil),
		NewQuad(s, p, NewTypedLiteral("96", integer), graph),
		NewQuad(s, p, NewLiteral("x"), bnode),
	}
	for _, q := range quads {
		parsed, err := ParseNQuadsLine(q.NQuadsLine())
		if err != nil {
			t.Errorf("ParseNQuadsLine(%s): %v", q.NQuadsLine(), err)
			continue
		}
		if parsed != q {
			t.Errorf("round trip of %s produced a different quad %v", q.NQuadsLine(), parsed)
		}
	}
}

func TestParseNQuadsLineBlankNodeDots(t *testing.T) {
	// A dot glued to a blank node label terminates the statement...
	q, err := ParseNQuadsLine(`<http://example.com/s> <http://example.com/p> "x" _:g.`)
	if err != nil {
		t.Fatalf("ParseNQuadsLine: %v", err)
	}
	if got := q.GraphName.String(); got != "_:g" {
		t.Errorf("graph name = %s, want _:g", got)
	}
	// ...while an interior dot stays part of the label.
	q, err = ParseNQuadsLine(`_:a.b <http://example.com/p> "x" .`)
	if err != nil {
		t.Fatalf("ParseNQuadsLine: %v", err)
	}
	if got := q.Subject.String(); got != "_:a.b" {
		t.Errorf("subject = %s, want _:a.b", got)
	}
}

func TestParseNQuadsLineRejectsBareLangString(t *testing.T) {
	line := `<http://example.com/s> <http://example.com/p> "x"^^<http://www.w3.org/1999/02/22-rdf-syntax-ns#langString> .`
	if _, err := ParseNQuadsLine(line); !errors.Is(err, ErrSyntax) {
		t.Errorf("error = %v, want ErrSyntax", err)
	}
}

func TestParseTermDecodesIRIEscapes(t *testing.T) {
	input := "<http://example.com/" + `\` + "u00E9>" // the u00E9 escape must decode to é
	n, err := ParseNamedNode(input)
	if err != nil {
		t.Fatalf("ParseNamedNode(%q): %v", input, err)
	}
	if got := n.Value(); got != "http://example.com/é" {
		t.Errorf("Value() = %q", got)
	}
}

func TestParseTerm(t *testing.T) {
	term, err := ParseTerm(`"Le Petit Prince"@fr`)
	if err != nil {
		t.Fatalf("ParseTerm: %v", err)
	}
	want, _ := NewLanguageTaggedLiteral("Le Petit Prince", "fr")
	if term != Term(want) {
		t.Errorf("ParseTerm = %v, want %v", term, want)
	}

	if _, err := ParseTerm(`<http://example.com/a> <http://example.com/b>`); !errors.Is(err, ErrSyntax) {
		t.Errorf("trailing content error = %v, want ErrSyntax", err)
	}
}

func TestParseTypedParsers(t *testing.T) {
	n, err := ParseNamedNode(`<http://example.com/a>`)
	if err != nil || n.Value() != "http://example.com/a" {
		t.Errorf("ParseNamedNode = %v, %v", n, err)
	}
	b, err := ParseBlankNode(`_:b1`)
	if err != nil || b.Value() != "b1" {
		t.Errorf("ParseBlankNode = %v, %v", b, err)
	}
	l, err := ParseLiteral(`"96"^^<http://www.w3.org/2001/XMLSchema#integer>`)
	if err != nil || l.Value() != "96" {
		t.Errorf("ParseLiteral = %v, %v", l, err)
	}
	if _, err := ParseNamedNode(`"not an iri"`); !errors.Is(err, ErrSyntax) {
		t.Errorf("ParseNamedNode on a literal = %v, want ErrSyntax", err)
	}
}
