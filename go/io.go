package oxigraph

/*
#include <stdlib.h>
#include "oxigraph_ffi.h"
*/
import "C"

import (
	"fmt"
	"io"
	"runtime"
	"unsafe"
)

// RdfFormat identifies an RDF serialization, mirroring pyoxigraph's
// RdfFormat. The zero value is not a valid format.
type RdfFormat int

const (
	// Turtle is https://www.w3.org/TR/turtle/ (triples only).
	Turtle RdfFormat = iota + 1
	// NTriples is https://www.w3.org/TR/n-triples/ (triples only).
	NTriples
	// NQuads is https://www.w3.org/TR/n-quads/ (datasets).
	NQuads
	// TriG is https://www.w3.org/TR/trig/ (datasets).
	TriG
)

// String returns the format's canonical name.
func (f RdfFormat) String() string {
	switch f {
	case Turtle:
		return "Turtle"
	case NTriples:
		return "N-Triples"
	case NQuads:
		return "N-Quads"
	case TriG:
		return "TriG"
	default:
		return fmt.Sprintf("RdfFormat(%d)", int(f))
	}
}

// supportsDatasets reports whether the format can carry named graphs.
func (f RdfFormat) supportsDatasets() bool {
	return f == NQuads || f == TriG
}

func (f RdfFormat) valid() bool {
	return f >= Turtle && f <= TriG
}

// Load reads an RDF document from r and inserts its data into the
// store, mirroring pyoxigraph's Store.load: quads are added to what the
// store already contains, atomically — either the whole document loads
// or nothing does. The document is buffered and crosses the FFI in one
// call. The returned error matches ErrUnsupportedFormat for a format
// this package does not define, ErrSyntax for a malformed document
// (carrying the engine's message with line and column), ErrStorage, or
// ErrStoreClosed.
func (s *Store) Load(r io.Reader, format RdfFormat) error {
	if !format.valid() {
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
	// Fail fast on a closed store before draining the reader; the check
	// repeats under the lock below in case of a concurrent Close.
	s.mu.RLock()
	closed := s.ptr == nil
	s.mu.RUnlock()
	if closed {
		return ErrStoreClosed
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ptr == nil {
		return ErrStoreClosed
	}
	cFormat := C.CString(format.String())
	defer C.free(unsafe.Pointer(cFormat))
	var cData *C.char
	if len(data) > 0 {
		cData = (*C.char)(unsafe.Pointer(&data[0]))
	}
	var cError *C.char
	status := C.oxigraph_load(s.ptr, cFormat, cData, C.size_t(len(data)), &cError)
	// cgo already pins pointer arguments for the call's duration; the
	// KeepAlive makes the slice's liveness explicit regardless.
	runtime.KeepAlive(data)
	if status != 0 {
		return statementError(status, cError)
	}
	return nil
}

// Dump serializes the whole store — default and named graphs — to w,
// mirroring pyoxigraph's Store.dump. The format must be a dataset
// format (NQuads or TriG): triples-only formats return
// ErrUnsupportedFormat, as pyoxigraph's dump refuses them without a
// graph scope. The returned error also matches ErrStoreClosed and
// ErrStorage.
func (s *Store) Dump(w io.Writer, format RdfFormat) error {
	if !format.valid() {
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
	if !format.supportsDatasets() {
		return fmt.Errorf("%w: %s stores triples only; dump with a dataset format such as N-Quads or TriG",
			ErrUnsupportedFormat, format)
	}
	s.mu.RLock()
	if s.ptr == nil {
		s.mu.RUnlock()
		return ErrStoreClosed
	}
	cFormat := C.CString(format.String())
	var cError *C.char
	cDump := C.oxigraph_dump(s.ptr, cFormat, &cError)
	C.free(unsafe.Pointer(cFormat))
	s.mu.RUnlock()
	if cDump == nil {
		message := "unknown storage error"
		if cError != nil {
			message = C.GoString(cError)
			C.oxigraph_free_string(cError)
		}
		return fmt.Errorf("%w: dumping as %s: %s", ErrStorage, format, message)
	}
	payload := C.GoString(cDump)
	C.oxigraph_free_string(cDump)
	_, err := io.WriteString(w, payload)
	return err
}
