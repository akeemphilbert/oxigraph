//go:build (darwin || linux) && (amd64 || arm64)

package features_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func registerQuerySteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^an open in-memory store$`, func() error {
		store, err := oxigraph.NewStore()
		if err != nil {
			return err
		}
		w.store = store
		return nil
	})
	sc.Step(`^the developer runs the query:$`, func(query *godog.DocString) error {
		return w.runsQuery(query.Content)
	})
	sc.Step(`^the developer runs the query "(.*)"$`, func(query string) error {
		return w.runsQuery(query)
	})

	sc.Step(`^the query returns these solutions in order:$`, func(table *godog.Table) error {
		results, err := w.queryResults(oxigraph.QuerySolutions)
		if err != nil {
			return err
		}
		if len(table.Rows) < 2 {
			return errors.New("the table needs a header row and at least one solution row")
		}
		header := table.Rows[0]
		want := len(table.Rows) - 1
		if len(results.Solutions) != want {
			return fmt.Errorf("got %d solutions, want %d", len(results.Solutions), want)
		}
		for i, row := range table.Rows[1:] {
			solution := results.Solutions[i]
			for j, cell := range row.Cells {
				variable := header.Cells[j].Value
				expected, err := oxigraph.ParseTerm(cell.Value)
				if err != nil {
					return fmt.Errorf("row %d, %q: %w", i+1, variable, err)
				}
				got, ok := solution.Get(variable)
				if !ok {
					return fmt.Errorf("solution %d does not bind %q", i+1, variable)
				}
				if got != expected {
					return fmt.Errorf("solution %d binds %q to %s, want %s", i+1, variable, got, expected)
				}
			}
		}
		return nil
	})
	sc.Step(`^the query returns no solutions$`, func() error {
		results, err := w.queryResults(oxigraph.QuerySolutions)
		if err != nil {
			return err
		}
		if len(results.Solutions) != 0 {
			return fmt.Errorf("got %d solutions, want none", len(results.Solutions))
		}
		return nil
	})
	sc.Step(`^the query returns 1 solution$`, func() error {
		results, err := w.queryResults(oxigraph.QuerySolutions)
		if err != nil {
			return err
		}
		if len(results.Solutions) != 1 {
			return fmt.Errorf("got %d solutions, want 1", len(results.Solutions))
		}
		return nil
	})
	sc.Step(`^the solution binds "([^"]*)" to the literal "([^"]*)"$`, func(variable, value string) error {
		results, err := w.queryResults(oxigraph.QuerySolutions)
		if err != nil {
			return err
		}
		if len(results.Solutions) == 0 {
			return errors.New("there is no solution")
		}
		got, ok := results.Solutions[0].Get(variable)
		if !ok {
			return fmt.Errorf("%q is unbound", variable)
		}
		if got != oxigraph.Term(oxigraph.NewLiteral(value)) {
			return fmt.Errorf("%q is bound to %s, want %q", variable, got, value)
		}
		return nil
	})
	sc.Step(`^the solution does not bind "([^"]*)"$`, func(variable string) error {
		results, err := w.queryResults(oxigraph.QuerySolutions)
		if err != nil {
			return err
		}
		if len(results.Solutions) == 0 {
			return errors.New("there is no solution")
		}
		if term, ok := results.Solutions[0].Get(variable); ok {
			return fmt.Errorf("%q is unexpectedly bound to %s", variable, term)
		}
		return nil
	})
	sc.Step(`^the query answers (true|false)$`, func(answer string) error {
		results, err := w.queryResults(oxigraph.QueryBoolean)
		if err != nil {
			return err
		}
		if results.Bool != (answer == "true") {
			return fmt.Errorf("the query answered %v, want %s", results.Bool, answer)
		}
		return nil
	})
	sc.Step(`^the query returns exactly these triples:$`, func(table *godog.Table) error {
		results, err := w.queryResults(oxigraph.QueryTriples)
		if err != nil {
			return err
		}
		if len(table.Rows) < 2 {
			return errors.New("the table needs a header row and at least one triple row")
		}
		want := map[oxigraph.Triple]bool{}
		for _, row := range table.Rows[1:] {
			if len(row.Cells) != 3 {
				return fmt.Errorf("expected subject/predicate/object cells, got %d", len(row.Cells))
			}
			triple, err := tripleFromFields(map[string]string{
				"subject":   row.Cells[0].Value,
				"predicate": row.Cells[1].Value,
				"object":    row.Cells[2].Value,
			})
			if err != nil {
				return err
			}
			want[triple] = true
		}
		if len(results.Triples) != len(want) {
			return fmt.Errorf("got %d triples, want %d", len(results.Triples), len(want))
		}
		for _, triple := range results.Triples {
			if !want[triple] {
				return fmt.Errorf("unexpected triple %s", triple)
			}
		}
		return nil
	})
	sc.Step(`^the query returns no triples$`, func() error {
		results, err := w.queryResults(oxigraph.QueryTriples)
		if err != nil {
			return err
		}
		if len(results.Triples) != 0 {
			return fmt.Errorf("got %d triples, want none", len(results.Triples))
		}
		return nil
	})
	sc.Step(`^the query fails with a syntax error$`, func() error {
		return w.failsWith(oxigraph.ErrSyntax)
	})
	sc.Step(`^the query fails with an evaluation error$`, func() error {
		return w.failsWith(oxigraph.ErrEvaluation)
	})
	sc.Step(`^the query fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
	sc.Step(`^the error message mentions "([^"]*)"$`, func(fragment string) error {
		if w.err == nil {
			return errors.New("there is no error")
		}
		if !strings.Contains(w.err.Error(), fragment) {
			return fmt.Errorf("error %q does not mention %q", w.err, fragment)
		}
		return nil
	})
}

// runsQuery records the outcome of a query; success or failure is
// asserted by a later Then step.
func (w *world) runsQuery(query string) error {
	if w.store == nil {
		return errors.New("no store is open")
	}
	results, err := w.store.Query(query)
	w.err = err
	w.query = results
	w.hasQuery = err == nil
	return nil
}

// queryResults returns the last query's results, requiring the kind the
// asserting step expects.
func (w *world) queryResults(kind oxigraph.QueryResultsKind) (oxigraph.QueryResults, error) {
	if w.err != nil {
		return oxigraph.QueryResults{}, fmt.Errorf("the query failed: %w", w.err)
	}
	if !w.hasQuery {
		return oxigraph.QueryResults{}, errors.New("no query was run")
	}
	if w.query.Kind != kind {
		return oxigraph.QueryResults{}, fmt.Errorf("the query returned kind %v, want %v", w.query.Kind, kind)
	}
	return w.query, nil
}
