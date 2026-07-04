// The embedded store requires cgo and a platform the oxigraph-ffi
// static library is built for; the RDF term model in the rest of the
// package is pure Go and builds everywhere.
//go:build darwin || linux

package oxigraph

/*
#include <stdlib.h>
#include "oxigraph_ffi.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Add inserts a quad into the store, mirroring pyoxigraph's Store.add.
// Inserting an already-present quad is a no-op, per RDF set semantics.
// The quad crosses the FFI as one N-Quads statement, so blank node
// identity is preserved. The returned error matches ErrStorage,
// ErrStoreClosed, or ErrSyntax when the quad's serialized form is
// rejected (for example a NUL byte in a term).
func (s *Store) Add(quad Quad) error {
	return s.statement(func(ptr *C.OxigraphStore, line *C.char, cError **C.char) C.int {
		return C.oxigraph_add(ptr, line, cError)
	}, quad.NQuadsLine())
}

// Remove deletes a quad from the store, mirroring pyoxigraph's
// Store.remove. Removing an absent quad is a no-op. Errors classify as
// Add's do.
func (s *Store) Remove(quad Quad) error {
	return s.statement(func(ptr *C.OxigraphStore, line *C.char, cError **C.char) C.int {
		return C.oxigraph_remove(ptr, line, cError)
	}, quad.NQuadsLine())
}

// Update executes a SPARQL update against the store, applied atomically:
// either the whole operation succeeds or nothing is written. The
// returned error matches ErrSyntax for malformed SPARQL, ErrEvaluation
// or ErrStorage for execution failures, and ErrStoreClosed after Close.
func (s *Store) Update(update string) error {
	return s.statement(func(ptr *C.OxigraphStore, text *C.char, cError **C.char) C.int {
		return C.oxigraph_update(ptr, text, cError)
	}, update)
}

// statement runs one string-carrying FFI write call under the store's
// read lock (the engine handles its own write synchronization) and maps
// the status/kind convention to Go errors.
func (s *Store) statement(call func(*C.OxigraphStore, *C.char, **C.char) C.int, text string) error {
	if hasNUL(text) {
		return fmt.Errorf("%w: the statement contains a NUL byte", ErrSyntax)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ptr == nil {
		return ErrStoreClosed
	}
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	var cError *C.char
	if status := call(s.ptr, cText, &cError); status != 0 {
		return statementError(status, cError)
	}
	return nil
}

// statementError classifies a failed write call by the kind the FFI
// reported, freeing the C error string.
func statementError(kind C.int, cError *C.char) error {
	message := "unknown error"
	if cError != nil {
		message = C.GoString(cError)
		C.oxigraph_free_string(cError)
	}
	sentinel := ErrEvaluation
	switch kind {
	case C.OXIGRAPH_ERROR_SYNTAX:
		sentinel = ErrSyntax
	case C.OXIGRAPH_ERROR_STORAGE:
		sentinel = ErrStorage
	case C.OXIGRAPH_ERROR_UNSUPPORTED_FORMAT:
		sentinel = ErrUnsupportedFormat
	}
	return fmt.Errorf("%w: %s", sentinel, message)
}
