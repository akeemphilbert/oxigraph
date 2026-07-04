//go:build (darwin || linux) && (amd64 || arm64)

package features_test

import (
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func registerStatementSteps(sc *godog.ScenarioContext, w *world) {
	// Triples.
	sc.Step(`^the developer creates a triple with:$`, func(table *godog.Table) error {
		triple, err := tripleFromTable(table)
		if err != nil {
			return err
		}
		w.triple, w.hasTriple = triple, true
		return nil
	})
	sc.Step(`^the triple:$`, func(table *godog.Table) error {
		triple, err := tripleFromTable(table)
		if err != nil {
			return err
		}
		w.triples = append(w.triples, triple)
		return nil
	})
	sc.Step(`^the triple's subject is the named node "([^"]*)"$`, func(iri string) error {
		if !w.hasTriple {
			return errors.New("no triple was constructed")
		}
		return expectNamedNode(w.triple.Subject, iri)
	})
	sc.Step(`^the triple's predicate is the named node "([^"]*)"$`, func(iri string) error {
		if !w.hasTriple {
			return errors.New("no triple was constructed")
		}
		return expectNamedNode(w.triple.Predicate, iri)
	})
	sc.Step(`^the triple's object is the language-tagged literal "([^"]*)" with the language "([^"]*)"$`, func(value, language string) error {
		if !w.hasTriple {
			return errors.New("no triple was constructed")
		}
		return expectLanguageTaggedLiteral(w.triple.Object, value, language)
	})
	sc.Step(`^the triple serializes to:$`, func(want *godog.DocString) error {
		if !w.hasTriple {
			return errors.New("no triple was constructed")
		}
		if got := w.triple.String(); got != want.Content {
			return fmt.Errorf("the triple serializes to %s, want %s", got, want.Content)
		}
		return nil
	})
	registerEqualitySteps(sc, w, "triple", &w.triples)

	// Quads.
	sc.Step(`^the developer creates a quad with:$`, func(table *godog.Table) error {
		quad, err := quadFromTable(table)
		if err != nil {
			return err
		}
		w.quad, w.hasQuad = quad, true
		return nil
	})
	sc.Step(`^the quad:$`, func(table *godog.Table) error {
		quad, err := quadFromTable(table)
		if err != nil {
			return err
		}
		w.quads = append(w.quads, quad)
		return nil
	})
	sc.Step(`^the quad's subject is the named node "([^"]*)"$`, func(iri string) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		return expectNamedNode(quad.Subject, iri)
	})
	sc.Step(`^the quad's predicate is the named node "([^"]*)"$`, func(iri string) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		return expectNamedNode(quad.Predicate, iri)
	})
	sc.Step(`^the quad's object is the language-tagged literal "([^"]*)" with the language "([^"]*)"$`, func(value, language string) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		return expectLanguageTaggedLiteral(quad.Object, value, language)
	})
	sc.Step(`^the quad's object is a literal with the value:$`, func(want *godog.DocString) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		l, ok := quad.Object.(oxigraph.Literal)
		if !ok {
			return fmt.Errorf("the object is a %T, not a Literal", quad.Object)
		}
		if l.Value() != want.Content {
			return fmt.Errorf("the object's value is %q, want %q", l.Value(), want.Content)
		}
		return nil
	})
	sc.Step(`^the quad's graph name is the named node "([^"]*)"$`, func(iri string) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		return expectNamedNode(quad.GraphName, iri)
	})
	sc.Step(`^the quad's graph name is the default graph$`, func() error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		if quad.GraphName != oxigraph.GraphName(oxigraph.DefaultGraph()) {
			return fmt.Errorf("the graph name is %v, not the default graph", quad.GraphName)
		}
		return nil
	})
	sc.Step(`^the quad's triple equals the triple:$`, func(table *godog.Table) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		want, err := tripleFromTable(table)
		if err != nil {
			return err
		}
		if got := quad.Triple(); got != want {
			return fmt.Errorf("the quad's triple is %v, want %v", got, want)
		}
		return nil
	})
	sc.Step(`^the quad serializes to:$`, func(want *godog.DocString) error {
		quad, err := w.lastQuad()
		if err != nil {
			return err
		}
		if got := quad.String(); got != want.Content {
			return fmt.Errorf("the quad serializes to %s, want %s", got, want.Content)
		}
		return nil
	})
	registerEqualitySteps(sc, w, "quad", &w.quads)
}

// expectNamedNode checks that value is the named node with the given IRI.
// It accepts any of the term interfaces, so it works for subjects,
// predicates, objects and graph names alike.
func expectNamedNode(value any, iri string) error {
	n, ok := value.(oxigraph.NamedNode)
	if !ok {
		return fmt.Errorf("%v is a %T, not a NamedNode", value, value)
	}
	if n.Value() != iri {
		return fmt.Errorf("the named node is %q, want %q", n.Value(), iri)
	}
	return nil
}

func expectLanguageTaggedLiteral(value any, want, language string) error {
	l, ok := value.(oxigraph.Literal)
	if !ok {
		return fmt.Errorf("%v is a %T, not a Literal", value, value)
	}
	if l.Value() != want {
		return fmt.Errorf("the literal's value is %q, want %q", l.Value(), want)
	}
	lang, ok := l.Language()
	if !ok || lang != language {
		return fmt.Errorf("the literal's language is %q, want %q", lang, language)
	}
	return nil
}
