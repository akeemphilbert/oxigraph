package features_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func registerStoreSteps(sc *godog.ScenarioContext, w *world) {
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		// Close every store the scenario opened so a leaked RocksDB lock
		// cannot poison later scenarios, then drop the workspace.
		for _, store := range w.storesByName {
			_ = store.Close()
		}
		if w.store != nil {
			_ = w.store.Close()
		}
		if w.workspace != "" {
			_ = os.RemoveAll(w.workspace)
		}
		return ctx, nil
	})

	sc.Step(`^no directory exists at "([^"]*)"$`, func(name string) error {
		if _, err := os.Stat(w.storePath(name)); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%q unexpectedly exists", name)
		}
		return nil
	})
	sc.Step(`^an open on-disk store at "([^"]*)"$`, func(name string) error {
		store, err := oxigraph.Open(w.storePath(name))
		if err != nil {
			return err
		}
		w.store = store
		w.storesByName[name] = store
		return nil
	})
	sc.Step(`^an in-memory store that has been closed$`, func() error {
		store, err := oxigraph.NewStore()
		if err != nil {
			return err
		}
		if err := store.Close(); err != nil {
			return err
		}
		w.store = store
		return nil
	})
	sc.Step(`^the developer opens an on-disk store at "([^"]*)"$`, w.opensStoreAt)
	sc.Step(`^the developer opens a second on-disk store at "([^"]*)"$`, w.opensStoreAt)
	sc.Step(`^the developer opens an in-memory store$`, func() error {
		store, err := oxigraph.NewStore()
		w.err = err
		if err == nil {
			w.store = store
		}
		return nil
	})
	sc.Step(`^the developer closes the store$`, w.closesTheStore)
	sc.Step(`^the developer closes the store again$`, w.closesTheStore)
	sc.Step(`^the developer closes the store at "([^"]*)"$`, func(name string) error {
		store, ok := w.storesByName[name]
		if !ok {
			return fmt.Errorf("no store was opened at %q", name)
		}
		w.err = store.Close()
		return nil
	})

	sc.Step(`^the store is open$`, func() error {
		if w.err != nil {
			return fmt.Errorf("the open failed: %w", w.err)
		}
		if w.store == nil {
			return errors.New("no store was opened")
		}
		return nil
	})
	sc.Step(`^a directory exists at "([^"]*)"$`, func(name string) error {
		info, err := os.Stat(w.storePath(name))
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("%q is not a directory", name)
		}
		return nil
	})
	sc.Step(`^no data directory is created$`, func() error {
		if w.workspace == "" {
			return nil // no on-disk path was ever touched
		}
		entries, err := os.ReadDir(w.workspace)
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return fmt.Errorf("the workspace unexpectedly contains %d entries", len(entries))
		}
		return nil
	})
	sc.Step(`^the store at "([^"]*)" is still open$`, func(name string) error {
		store, ok := w.storesByName[name]
		if !ok {
			return fmt.Errorf("no store was opened at %q", name)
		}
		// With no read surface yet, a successful Close is the probe that
		// the handle was still open; a closed one returns ErrStoreClosed.
		if err := store.Close(); err != nil {
			return fmt.Errorf("the store at %q was not open: %w", name, err)
		}
		delete(w.storesByName, name)
		return nil
	})
	sc.Step(`^the open fails with a storage error$`, func() error {
		return w.failsWith(oxigraph.ErrStorage)
	})
	sc.Step(`^the close fails with a closed store error$`, func() error {
		return w.failsWith(oxigraph.ErrStoreClosed)
	})
}

// opensStoreAt records the outcome of an Open attempt without failing the
// step: the contract asserts success or failure in a later Then step.
func (w *world) opensStoreAt(name string) error {
	store, err := oxigraph.Open(w.storePath(name))
	w.err = err
	if err == nil {
		w.store = store
		w.storesByName[name] = store
	}
	return nil
}

// closesTheStore closes the most recently referenced store, recording the
// outcome for a later Then step.
func (w *world) closesTheStore() error {
	if w.store == nil {
		return errors.New("no store to close")
	}
	w.err = w.store.Close()
	return nil
}

// storePath roots a scenario's store name in its temp workspace, creating
// the workspace lazily.
func (w *world) storePath(name string) string {
	if w.workspace == "" {
		dir, err := os.MkdirTemp("", "oxigraph-acceptance-")
		if err != nil {
			panic(err)
		}
		w.workspace = dir
	}
	return filepath.Join(w.workspace, name)
}
