//go:build darwin || linux

package oxigraph

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesLeafDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book-catalog")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Errorf("Open must create the leaf directory: %v", err)
	}
}

func TestNewStoreInMemory(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestCloseReleasesTheLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book-catalog")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen after Close: %v", err)
	}
	reopened.Close()
}

func TestOpenLockedDirectoryFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book-catalog")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()
	if _, err := Open(path); !errors.Is(err, ErrStorage) {
		t.Errorf("second Open error = %v, want ErrStorage", err)
	}
}

func TestNULBytesRejected(t *testing.T) {
	if _, err := Open("book\x00catalog"); !errors.Is(err, ErrStorage) {
		t.Errorf("Open with a NUL byte = %v, want ErrStorage", err)
	}
	store := mustInMemoryStore(t)
	if _, err := store.Query("ASK { }\x00"); !errors.Is(err, ErrSyntax) {
		t.Errorf("Query with a NUL byte = %v, want ErrSyntax", err)
	}
	if err := store.Update("INSERT DATA { }\x00"); !errors.Is(err, ErrSyntax) {
		t.Errorf("Update with a NUL byte = %v, want ErrSyntax", err)
	}
}

func TestOpenMissingParentFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "archive", "book-catalog")
	if _, err := Open(path); !errors.Is(err, ErrStorage) {
		t.Errorf("Open error = %v, want ErrStorage", err)
	}
}

func TestDoubleCloseFails(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := store.Close(); !errors.Is(err, ErrStoreClosed) {
		t.Errorf("second Close error = %v, want ErrStoreClosed", err)
	}
}

func TestStoresAreIndependent(t *testing.T) {
	dir := t.TempDir()
	books, err := Open(filepath.Join(dir, "book-catalog"))
	if err != nil {
		t.Fatalf("Open books: %v", err)
	}
	music, err := Open(filepath.Join(dir, "music-catalog"))
	if err != nil {
		t.Fatalf("Open music: %v", err)
	}
	if err := books.Close(); err != nil {
		t.Fatalf("Close books: %v", err)
	}
	if err := music.Close(); err != nil {
		t.Errorf("music store must survive closing the book store: %v", err)
	}
}
