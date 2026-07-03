package features_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

// world holds one scenario's state. A fresh value is installed before each
// scenario, so steps never see another scenario's terms or errors.
type world struct {
	result     any   // last constructed or parsed value (term, default graph, ...)
	err        error // error from the last construction or parse
	terms      []any // comparison stack seeded by the Given steps
	triples    []oxigraph.Triple
	quads      []oxigraph.Quad
	blankNodes []oxigraph.BlankNode
	quad       oxigraph.Quad // last constructed or parsed quad
	hasQuad    bool
	triple     oxigraph.Triple // last constructed triple
	hasTriple  bool
	line       string // original N-Quads line, for round-trip checks
	equal      bool   // outcome of the last comparison
	compared   bool
}

func InitializeScenario(sc *godog.ScenarioContext) {
	w := &world{}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*w = world{}
		return ctx, nil
	})
	registerTermSteps(sc, w)
	registerStatementSteps(sc, w)
	registerParsingSteps(sc, w)
}

// stringer returns the last constructed value as a fmt.Stringer, failing
// if the preceding construction errored.
func (w *world) stringer() (fmt.Stringer, error) {
	if w.err != nil {
		return nil, fmt.Errorf("the preceding construction failed: %w", w.err)
	}
	s, ok := w.result.(fmt.Stringer)
	if !ok {
		return nil, fmt.Errorf("no term was constructed (got %T)", w.result)
	}
	return s, nil
}

// literal returns the last constructed value as a Literal.
func (w *world) literal() (oxigraph.Literal, error) {
	if w.err != nil {
		return oxigraph.Literal{}, fmt.Errorf("the preceding construction failed: %w", w.err)
	}
	l, ok := w.result.(oxigraph.Literal)
	if !ok {
		return oxigraph.Literal{}, fmt.Errorf("the last value is a %T, not a Literal", w.result)
	}
	return l, nil
}

// lastQuad returns the last constructed or parsed quad.
func (w *world) lastQuad() (oxigraph.Quad, error) {
	if w.err != nil {
		return oxigraph.Quad{}, fmt.Errorf("the preceding parse failed: %w", w.err)
	}
	if !w.hasQuad {
		return oxigraph.Quad{}, errors.New("no quad was constructed or parsed")
	}
	return w.quad, nil
}

// failsWith checks that the last construction or parse failed with the
// given error kind.
func (w *world) failsWith(kind error) error {
	if w.err == nil {
		return fmt.Errorf("expected a failure matching %v, but the operation succeeded with %v", kind, w.result)
	}
	if !errors.Is(w.err, kind) {
		return fmt.Errorf("expected an error matching %v, got %v", kind, w.err)
	}
	return nil
}

// pushTerm seeds the comparison stack used by the equality scenarios.
func (w *world) pushTerm(term any) {
	w.terms = append(w.terms, term)
}

// termTable turns a vertical two-column Gherkin table (field | N-Quads
// term expression) into a field → parsed value map.
func termTable(table *godog.Table) (map[string]string, error) {
	fields := map[string]string{}
	for _, row := range table.Rows {
		if len(row.Cells) != 2 {
			return nil, fmt.Errorf("expected a two-column table row, got %d cells", len(row.Cells))
		}
		fields[row.Cells[0].Value] = row.Cells[1].Value
	}
	return fields, nil
}

// registerEqualitySteps wires the compare/equal/not-equal step trio for a
// kind of value ("term", "triple", "quad") over its comparison stack.
func registerEqualitySteps[T comparable](sc *godog.ScenarioContext, w *world, noun string, items *[]T) {
	sc.Step(`^the developer compares the two `+noun+`s$`, func() error {
		if len(*items) != 2 {
			return fmt.Errorf("expected two %ss on the stack, have %d", noun, len(*items))
		}
		w.equal, w.compared = (*items)[0] == (*items)[1], true
		return nil
	})
	sc.Step(`^the `+noun+`s are equal$`, func() error {
		if !w.compared {
			return errors.New("no comparison happened")
		}
		if !w.equal {
			return fmt.Errorf("the %ss %v and %v are not equal", noun, (*items)[0], (*items)[1])
		}
		return nil
	})
	sc.Step(`^the `+noun+`s are not equal$`, func() error {
		if !w.compared {
			return errors.New("no comparison happened")
		}
		if w.equal {
			return fmt.Errorf("the %ss %v and %v are equal", noun, (*items)[0], (*items)[1])
		}
		return nil
	})
}

// tripleFromTable builds a Triple from a subject/predicate/object table,
// parsing each cell with the same term parser the library ships.
func tripleFromTable(table *godog.Table) (oxigraph.Triple, error) {
	fields, err := termTable(table)
	if err != nil {
		return oxigraph.Triple{}, err
	}
	return tripleFromFields(fields)
}

// tripleFromFields builds a Triple from an already-collected field map.
func tripleFromFields(fields map[string]string) (oxigraph.Triple, error) {
	subjectTerm, err := oxigraph.ParseTerm(fields["subject"])
	if err != nil {
		return oxigraph.Triple{}, fmt.Errorf("subject: %w", err)
	}
	subject, ok := subjectTerm.(oxigraph.Subject)
	if !ok {
		return oxigraph.Triple{}, fmt.Errorf("subject %q is not a named or blank node", fields["subject"])
	}
	predicate, err := oxigraph.ParseNamedNode(fields["predicate"])
	if err != nil {
		return oxigraph.Triple{}, fmt.Errorf("predicate: %w", err)
	}
	object, err := oxigraph.ParseTerm(fields["object"])
	if err != nil {
		return oxigraph.Triple{}, fmt.Errorf("object: %w", err)
	}
	return oxigraph.NewTriple(subject, predicate, object), nil
}

// quadFromTable builds a Quad from a triple table with an optional
// graph-name row; a missing row means the default graph.
func quadFromTable(table *godog.Table) (oxigraph.Quad, error) {
	fields, err := termTable(table)
	if err != nil {
		return oxigraph.Quad{}, err
	}
	triple, err := tripleFromFields(fields)
	if err != nil {
		return oxigraph.Quad{}, err
	}
	graphName := oxigraph.GraphName(oxigraph.DefaultGraph())
	if expr, ok := fields["graph-name"]; ok {
		graphTerm, err := oxigraph.ParseTerm(expr)
		if err != nil {
			return oxigraph.Quad{}, fmt.Errorf("graph-name: %w", err)
		}
		name, ok := graphTerm.(oxigraph.GraphName)
		if !ok {
			return oxigraph.Quad{}, fmt.Errorf("graph-name %q is not a named or blank node", expr)
		}
		graphName = name
	}
	return oxigraph.NewQuad(triple.Subject, triple.Predicate, triple.Object, graphName), nil
}
