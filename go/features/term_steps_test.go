//go:build (darwin || linux) && (amd64 || arm64)

package features_test

import (
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func registerTermSteps(sc *godog.ScenarioContext, w *world) {
	// Construction.
	sc.Step(`^the developer creates a named node from the IRI "(.*)"$`, func(iri string) error {
		w.result, w.err = oxigraph.NewNamedNode(iri)
		return nil
	})
	sc.Step(`^the developer creates a blank node with the identifier "(.*)"$`, func(id string) error {
		w.result, w.err = oxigraph.NewBlankNode(id)
		return nil
	})
	sc.Step(`^the developer creates two blank nodes without identifiers$`, func() error {
		w.blankNodes = []oxigraph.BlankNode{oxigraph.NewBlankNodeRandom(), oxigraph.NewBlankNodeRandom()}
		return nil
	})
	sc.Step(`^the developer creates the literal "(.*)"$`, func(value string) error {
		w.result, w.err = oxigraph.NewLiteral(value), nil
		return nil
	})
	sc.Step(`^the developer creates a literal with the value:$`, func(value *godog.DocString) error {
		w.result, w.err = oxigraph.NewLiteral(value.Content), nil
		return nil
	})
	sc.Step(`^the developer creates the language-tagged literal "([^"]*)" with the language "(.*)"$`, func(value, language string) error {
		w.result, w.err = oxigraph.NewLanguageTaggedLiteral(value, language)
		return nil
	})
	sc.Step(`^the developer creates the typed literal "([^"]*)" with the datatype "([^"]*)"$`, func(value, datatype string) error {
		dt, err := oxigraph.NewNamedNode(datatype)
		if err != nil {
			return fmt.Errorf("the scenario's datatype IRI is invalid: %w", err)
		}
		w.result, w.err = oxigraph.NewTypedLiteral(value, dt), nil
		return nil
	})
	sc.Step(`^the developer creates a default graph value$`, func() error {
		w.result, w.err = oxigraph.DefaultGraph(), nil
		return nil
	})

	// Construction outcomes.
	sc.Step(`^the named node's value is "([^"]*)"$`, func(want string) error {
		if w.err != nil {
			return fmt.Errorf("the construction failed: %w", w.err)
		}
		n, ok := w.result.(oxigraph.NamedNode)
		if !ok {
			return fmt.Errorf("the last value is a %T, not a NamedNode", w.result)
		}
		if n.Value() != want {
			return fmt.Errorf("value is %q, want %q", n.Value(), want)
		}
		return nil
	})
	sc.Step(`^the blank node's identifier is "([^"]*)"$`, func(want string) error {
		if w.err != nil {
			return fmt.Errorf("the construction failed: %w", w.err)
		}
		b, ok := w.result.(oxigraph.BlankNode)
		if !ok {
			return fmt.Errorf("the last value is a %T, not a BlankNode", w.result)
		}
		if b.Value() != want {
			return fmt.Errorf("identifier is %q, want %q", b.Value(), want)
		}
		return nil
	})
	sc.Step(`^each blank node has a non-empty identifier$`, func() error {
		if len(w.blankNodes) != 2 {
			return errors.New("expected two blank nodes")
		}
		for _, b := range w.blankNodes {
			if b.Value() == "" {
				return errors.New("a blank node has an empty identifier")
			}
		}
		return nil
	})
	sc.Step(`^the two blank nodes are not equal$`, func() error {
		if len(w.blankNodes) != 2 {
			return errors.New("expected two blank nodes")
		}
		if w.blankNodes[0] == w.blankNodes[1] {
			return fmt.Errorf("the blank nodes are both %v", w.blankNodes[0])
		}
		return nil
	})
	sc.Step(`^the literal's value is "([^"]*)"$`, func(want string) error {
		l, err := w.literal()
		if err != nil {
			return err
		}
		if l.Value() != want {
			return fmt.Errorf("value is %q, want %q", l.Value(), want)
		}
		return nil
	})
	sc.Step(`^the literal has no language$`, func() error {
		l, err := w.literal()
		if err != nil {
			return err
		}
		if lang, ok := l.Language(); ok {
			return fmt.Errorf("the literal has language %q", lang)
		}
		return nil
	})
	sc.Step(`^the literal's language is "([^"]*)"$`, func(want string) error {
		l, err := w.literal()
		if err != nil {
			return err
		}
		lang, ok := l.Language()
		if !ok {
			return errors.New("the literal has no language")
		}
		if lang != want {
			return fmt.Errorf("language is %q, want %q", lang, want)
		}
		return nil
	})
	sc.Step(`^the literal's datatype is "([^"]*)"$`, func(want string) error {
		l, err := w.literal()
		if err != nil {
			return err
		}
		if got := l.Datatype().Value(); got != want {
			return fmt.Errorf("datatype is %q, want %q", got, want)
		}
		return nil
	})
	sc.Step(`^the term's N-Quads form is "([^"]*)"$`, func(want string) error {
		s, err := w.stringer()
		if err != nil {
			return err
		}
		if got := s.String(); got != want {
			return fmt.Errorf("N-Quads form is %s, want %s", got, want)
		}
		return nil
	})
	sc.Step(`^the term's N-Quads form is:$`, func(want *godog.DocString) error {
		s, err := w.stringer()
		if err != nil {
			return err
		}
		if got := s.String(); got != want.Content {
			return fmt.Errorf("N-Quads form is %s, want %s", got, want.Content)
		}
		return nil
	})
	sc.Step(`^the default graph's string form is "([^"]*)"$`, func(want string) error {
		s, err := w.stringer()
		if err != nil {
			return err
		}
		if _, ok := w.result.(oxigraph.DefaultGraphTerm); !ok {
			return fmt.Errorf("the last value is a %T, not a DefaultGraphTerm", w.result)
		}
		if got := s.String(); got != want {
			return fmt.Errorf("string form is %q, want %q", got, want)
		}
		return nil
	})

	// Construction failures.
	sc.Step(`^the construction fails with an invalid IRI error$`, func() error {
		return w.failsWith(oxigraph.ErrInvalidIRI)
	})
	sc.Step(`^the construction fails with an invalid blank node identifier error$`, func() error {
		return w.failsWith(oxigraph.ErrInvalidBlankNodeID)
	})
	sc.Step(`^the construction fails with an invalid language tag error$`, func() error {
		return w.failsWith(oxigraph.ErrInvalidLanguageTag)
	})

	// Equality: the Given steps seed the stack, the When compares.
	sc.Step(`^the named node "([^"]*)"$`, func(iri string) error {
		n, err := oxigraph.NewNamedNode(iri)
		if err != nil {
			return err
		}
		w.pushTerm(n)
		return nil
	})
	sc.Step(`^the blank node with the identifier "([^"]*)"$`, func(id string) error {
		b, err := oxigraph.NewBlankNode(id)
		if err != nil {
			return err
		}
		w.pushTerm(b)
		return nil
	})
	sc.Step(`^the literal "([^"]*)"$`, func(value string) error {
		w.pushTerm(oxigraph.NewLiteral(value))
		return nil
	})
	sc.Step(`^the typed literal "([^"]*)" with the datatype "([^"]*)"$`, func(value, datatype string) error {
		dt, err := oxigraph.NewNamedNode(datatype)
		if err != nil {
			return err
		}
		w.pushTerm(oxigraph.NewTypedLiteral(value, dt))
		return nil
	})
	sc.Step(`^the language-tagged literal "([^"]*)" with the language "([^"]*)"$`, func(value, language string) error {
		l, err := oxigraph.NewLanguageTaggedLiteral(value, language)
		if err != nil {
			return err
		}
		w.pushTerm(l)
		return nil
	})
	sc.Step(`^a default graph value$`, func() error {
		w.pushTerm(oxigraph.DefaultGraph())
		return nil
	})
	registerEqualitySteps(sc, w, "term", &w.terms)
}
