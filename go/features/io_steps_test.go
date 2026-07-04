package features_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

// rdfFormatsByName maps the contract's format names to the Go enum.
var rdfFormatsByName = map[string]oxigraph.RdfFormat{
	"Turtle":    oxigraph.Turtle,
	"N-Triples": oxigraph.NTriples,
	"N-Quads":   oxigraph.NQuads,
	"TriG":      oxigraph.TriG,
}

func registerIOSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the developer loads the (Turtle|N-Triples|N-Quads|TriG) document:$`, func(format string, doc *godog.DocString) error {
		if w.store == nil {
			return errors.New("no store is open")
		}
		w.err = w.store.Load(strings.NewReader(doc.Content), rdfFormatsByName[format])
		return nil
	})
	sc.Step(`^the developer loads a document using an undefined format$`, func() error {
		if w.store == nil {
			return errors.New("no store is open")
		}
		w.err = w.store.Load(strings.NewReader(""), oxigraph.RdfFormat(99))
		return nil
	})
	sc.Step(`^the load reports no error$`, func() error {
		if w.err != nil {
			return fmt.Errorf("the load failed: %w", w.err)
		}
		return nil
	})
	sc.Step(`^the load fails with a syntax error$`, func() error {
		return w.failsWith(oxigraph.ErrSyntax)
	})
	sc.Step(`^the load fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
	sc.Step(`^the load fails with an unsupported format error$`, func() error {
		return w.failsWith(oxigraph.ErrUnsupportedFormat)
	})

	sc.Step(`^the developer dumps the store as (Turtle|N-Triples|N-Quads|TriG)$`, func(format string) error {
		if w.store == nil {
			return errors.New("no store is open")
		}
		var dump strings.Builder
		w.dumpFormat = rdfFormatsByName[format]
		w.err = w.store.Dump(&dump, w.dumpFormat)
		if w.err == nil {
			w.dump = dump.String()
			w.hasDump = true
		}
		return nil
	})
	sc.Step(`^the developer dumps the store using an undefined format$`, func() error {
		if w.store == nil {
			return errors.New("no store is open")
		}
		w.err = w.store.Dump(&strings.Builder{}, oxigraph.RdfFormat(99))
		return nil
	})
	sc.Step(`^the developer loads the dump into a second in-memory store$`, func() error {
		if !w.hasDump {
			return fmt.Errorf("no dump was produced (last error: %v)", w.err)
		}
		second, err := oxigraph.NewStore()
		if err != nil {
			return err
		}
		w.secondStore = second
		return second.Load(strings.NewReader(w.dump), w.dumpFormat)
	})
	sc.Step(`^the second store contains the quad:$`, func(table *godog.Table) error {
		found, err := containsQuad(w.secondStore, table)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("the second store does not contain the quad")
		}
		return nil
	})
	sc.Step(`^the second store contains exactly (\d+) quads?$`, func(want int) error {
		count, err := quadCount(w.secondStore)
		if err != nil {
			return err
		}
		if count != want {
			return fmt.Errorf("the second store contains %d quads, want %d", count, want)
		}
		return nil
	})
	sc.Step(`^the second store is empty$`, func() error {
		count, err := quadCount(w.secondStore)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("the second store contains %d quads, want none", count)
		}
		return nil
	})
	sc.Step(`^the dump fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
	sc.Step(`^the dump fails with an unsupported format error$`, func() error {
		return w.failsWith(oxigraph.ErrUnsupportedFormat)
	})
}
