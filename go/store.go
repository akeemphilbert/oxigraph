// The embedded store requires cgo and a platform the oxigraph-ffi
// static library is built for; the RDF term model in the rest of the
// package is pure Go and builds everywhere.
//go:build ((darwin || linux) && (amd64 || arm64)) || (windows && amd64)

package oxigraph

/*
#include <stdlib.h>
#include "oxigraph_ffi.h"
*/
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// hasNUL reports whether s contains a NUL byte, which a C string cannot
// carry — C.CString would silently truncate at it.
func hasNUL(s string) bool { return strings.IndexByte(s, 0) >= 0 }

// Store is an embedded Oxigraph store, mirroring pyoxigraph's Store. It
// is backed by the Rust engine through the oxigraph-ffi C ABI; see the
// repository's go/README.md for how the static library is built.
//
// A Store must be released with Close; Go's garbage collector does not
// release the engine's resources (for an on-disk store, the RocksDB
// directory lock). A Store is safe for concurrent use.
type Store struct {
	mu  sync.RWMutex
	ptr *C.OxigraphStore
}

// Open opens a read-write store persisted at path with RocksDB, creating
// the leaf directory if missing — parent directories are not created,
// matching the engine. The returned error matches ErrStorage, including
// when the directory is locked by another open store.
func Open(path string) (*Store, error) {
	if hasNUL(path) {
		return nil, fmt.Errorf("%w: the store path contains a NUL byte", ErrStorage)
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cError *C.char
	ptr := C.oxigraph_open(cPath, &cError)
	if ptr == nil {
		return nil, storageError(fmt.Sprintf("opening %q", path), cError)
	}
	return &Store{ptr: ptr}, nil
}

// NewStore opens an in-memory store, mirroring pyoxigraph's Store().
func NewStore() (*Store, error) {
	var cError *C.char
	ptr := C.oxigraph_open_in_memory(&cError)
	if ptr == nil {
		return nil, storageError("opening an in-memory store", cError)
	}
	return &Store{ptr: ptr}, nil
}

// Close releases the store's resources; for an on-disk store that
// releases the RocksDB directory lock, so the path can be opened again.
// Closing an already-closed store returns ErrStoreClosed, like os.File.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptr == nil {
		return ErrStoreClosed
	}
	C.oxigraph_close(s.ptr)
	s.ptr = nil
	return nil
}

// storageError turns a caller-owned FFI error string into a Go error
// matching ErrStorage, freeing the C string.
func storageError(context string, cError *C.char) error {
	message := "unknown storage error"
	if cError != nil {
		message = C.GoString(cError)
		C.oxigraph_free_string(cError)
	}
	return fmt.Errorf("%w: %s: %s", ErrStorage, context, message)
}
