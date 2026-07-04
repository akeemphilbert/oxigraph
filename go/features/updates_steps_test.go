package features_test

import (
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func registerUpdateSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the developer adds the quad:$`, func(table *godog.Table) error {
		quad, err := quadFromTable(table)
		if err != nil {
			return err
		}
		if w.store == nil {
			return errors.New("no store is open")
		}
		w.err = w.store.Add(quad)
		return nil
	})
	sc.Step(`^the developer removes the quad:$`, func(table *godog.Table) error {
		quad, err := quadFromTable(table)
		if err != nil {
			return err
		}
		if w.store == nil {
			return errors.New("no store is open")
		}
		w.err = w.store.Remove(quad)
		return nil
	})
	sc.Step(`^the developer runs the update:$`, func(update *godog.DocString) error {
		return w.runsUpdate(update.Content)
	})
	sc.Step(`^the developer runs the update "(.*)"$`, func(update string) error {
		return w.runsUpdate(update)
	})

	// "the store contains the quad:" is setup sugar as a Given (add it,
	// failing the step on error) and an assertion as a Then (prove it
	// through a query).
	sc.Given(`^the store contains the quad:$`, func(table *godog.Table) error {
		quad, err := quadFromTable(table)
		if err != nil {
			return err
		}
		if w.store == nil {
			return errors.New("no store is open")
		}
		return w.store.Add(quad)
	})
	sc.Then(`^the store contains the quad:$`, func(table *godog.Table) error {
		found, err := w.storeContains(table)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("the store does not contain the quad")
		}
		return nil
	})
	sc.Then(`^the store does not contain the quad:$`, func(table *godog.Table) error {
		found, err := w.storeContains(table)
		if err != nil {
			return err
		}
		if found {
			return errors.New("the store unexpectedly contains the quad")
		}
		return nil
	})
	sc.Step(`^the store contains exactly (\d+) quads?$`, func(want int) error {
		count, err := w.storeQuadCount()
		if err != nil {
			return err
		}
		if count != want {
			return fmt.Errorf("the store contains %d quads, want %d", count, want)
		}
		return nil
	})
	sc.Step(`^the store is empty$`, func() error {
		count, err := w.storeQuadCount()
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("the store contains %d quads, want none", count)
		}
		return nil
	})
	sc.Step(`^the remove reports no error$`, func() error {
		if w.err != nil {
			return fmt.Errorf("the remove failed: %w", w.err)
		}
		return nil
	})
	sc.Step(`^the add fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
	sc.Step(`^the remove fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
	sc.Step(`^the update fails with a syntax error$`, func() error {
		return w.failsWith(oxigraph.ErrSyntax)
	})
	sc.Step(`^the update fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
}

// runsUpdate records the outcome of a SPARQL update; success or failure
// is asserted by a later Then step.
func (w *world) runsUpdate(update string) error {
	if w.store == nil {
		return errors.New("no store is open")
	}
	w.err = w.store.Update(update)
	return nil
}

// storeContains asks the store whether the quad described by the table
// is present, scoping the pattern to its graph.
func (w *world) storeContains(table *godog.Table) (bool, error) {
	return containsQuad(w.store, table)
}

// containsQuad asks a store whether the quad described by the table is
// present, scoping the pattern to its graph.
func containsQuad(store *oxigraph.Store, table *godog.Table) (bool, error) {
	if store == nil {
		return false, errors.New("no store is open")
	}
	quad, err := quadFromTable(table)
	if err != nil {
		return false, err
	}
	pattern := quad.Triple().String()
	query := "ASK { " + pattern + " }"
	if quad.GraphName != oxigraph.GraphName(oxigraph.DefaultGraph()) {
		query = "ASK { GRAPH " + quad.GraphName.String() + " { " + pattern + " } }"
	}
	results, err := store.Query(query)
	if err != nil {
		return false, err
	}
	return results.Bool, nil
}

// storeQuadCount counts the quads across the default and named graphs.
func (w *world) storeQuadCount() (int, error) {
	return quadCount(w.store)
}

// quadCount counts a store's quads across the default and named graphs.
func quadCount(store *oxigraph.Store) (int, error) {
	if store == nil {
		return 0, errors.New("no store is open")
	}
	results, err := store.Query(
		`SELECT (COUNT(*) AS ?n) WHERE { { ?s ?p ?o } UNION { GRAPH ?g { ?s ?p ?o } } }`)
	if err != nil {
		return 0, err
	}
	if len(results.Solutions) != 1 {
		return 0, fmt.Errorf("count query returned %d solutions", len(results.Solutions))
	}
	term, ok := results.Solutions[0].Get("n")
	if !ok {
		return 0, errors.New("count query bound no ?n")
	}
	literal, ok := term.(oxigraph.Literal)
	if !ok {
		return 0, fmt.Errorf("?n is a %T, not a literal", term)
	}
	var count int
	if _, err := fmt.Sscanf(literal.Value(), "%d", &count); err != nil {
		return 0, fmt.Errorf("?n = %q is not a number", literal.Value())
	}
	return count, nil
}
