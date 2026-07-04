//go:build darwin || linux

package features_test

import (
	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func registerParsingSteps(sc *godog.ScenarioContext, w *world) {
	// SPARQL JSON terms.
	sc.Step(`^the developer parses the SPARQL JSON term:$`, func(doc *godog.DocString) error {
		w.result, w.err = oxigraph.ParseSPARQLJSONTerm([]byte(doc.Content))
		return nil
	})
	sc.Step(`^the parsed term equals the named node "([^"]*)"$`, func(iri string) error {
		want, err := oxigraph.NewNamedNode(iri)
		if err != nil {
			return err
		}
		return w.expectParsedTerm(want)
	})
	sc.Step(`^the parsed term equals the blank node with the identifier "([^"]*)"$`, func(id string) error {
		want, err := oxigraph.NewBlankNode(id)
		if err != nil {
			return err
		}
		return w.expectParsedTerm(want)
	})
	sc.Step(`^the parsed term equals the literal "([^"]*)"$`, func(value string) error {
		return w.expectParsedTerm(oxigraph.NewLiteral(value))
	})
	sc.Step(`^the parsed term equals the language-tagged literal "([^"]*)" with the language "([^"]*)"$`, func(value, language string) error {
		want, err := oxigraph.NewLanguageTaggedLiteral(value, language)
		if err != nil {
			return err
		}
		return w.expectParsedTerm(want)
	})
	sc.Step(`^the parsed term equals the typed literal "([^"]*)" with the datatype "([^"]*)"$`, func(value, datatype string) error {
		dt, err := oxigraph.NewNamedNode(datatype)
		if err != nil {
			return err
		}
		return w.expectParsedTerm(oxigraph.NewTypedLiteral(value, dt))
	})

	// N-Quads lines. The inline variant needs a greedy capture: the line
	// itself contains quoted literals, so the capture must run to the
	// last '"' on the step, not the first embedded one.
	sc.Step(`^the developer parses the N-Quads line "(.*)"$`, func(line string) error {
		w.line = line
		quad, err := oxigraph.ParseNQuadsLine(line)
		w.quad, w.hasQuad, w.err = quad, err == nil, err
		return nil
	})
	sc.Step(`^the developer parses the N-Quads line:$`, func(doc *godog.DocString) error {
		w.line = doc.Content
		quad, err := oxigraph.ParseNQuadsLine(doc.Content)
		w.quad, w.hasQuad, w.err = quad, err == nil, err
		return nil
	})
	sc.Step(`^serializing the quad as an N-Quads line reproduces the original line$`, func() error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		if got := quad.NQuadsLine(); got != w.line {
			return &lineMismatchError{got: got, want: w.line}
		}
		return nil
	})

	// Parse failures.
	sc.Step(`^the parsing fails with an invalid IRI error$`, func() error {
		return w.failsWith(oxigraph.ErrInvalidIRI)
	})
	sc.Step(`^the parsing fails with an unsupported term type error$`, func() error {
		return w.failsWith(oxigraph.ErrUnsupportedTermType)
	})
	sc.Step(`^the parsing fails with a malformed term error$`, func() error {
		return w.failsWith(oxigraph.ErrMalformedTerm)
	})
	sc.Step(`^the parsing fails with a syntax error$`, func() error {
		return w.failsWith(oxigraph.ErrSyntax)
	})
}

// expectParsedTerm checks the outcome of the last parse against a term
// value, comparing with ==.
func (w *world) expectParsedTerm(want oxigraph.Term) error {
	if w.err != nil {
		return w.err
	}
	got, ok := w.result.(oxigraph.Term)
	if !ok {
		return &lineMismatchError{got: "no parsed term", want: want.String()}
	}
	if got != want {
		return &lineMismatchError{got: got.String(), want: want.String()}
	}
	return nil
}

// lineMismatchError formats got/want pairs without quoting, keeping the
// already-quoted N-Quads syntax readable in failure output.
type lineMismatchError struct {
	got  string
	want string
}

func (e *lineMismatchError) Error() string {
	return "got  " + e.got + "\nwant " + e.want
}
