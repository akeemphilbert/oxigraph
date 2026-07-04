package oxigraph

/*
#include <stdlib.h>
#include "oxigraph_ffi.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"
)

// QueryResultsKind identifies which field of QueryResults a query filled.
type QueryResultsKind int

const (
	// QuerySolutions marks a SELECT result carried in Solutions.
	QuerySolutions QueryResultsKind = iota + 1
	// QueryBoolean marks an ASK result carried in Bool.
	QueryBoolean
	// QueryTriples marks a CONSTRUCT or DESCRIBE result carried in Triples.
	QueryTriples
)

// QueryResults carries the result of one SPARQL query: Solutions for
// SELECT, Bool for ASK, Triples for CONSTRUCT and DESCRIBE — Kind says
// which, mirroring pyoxigraph's per-form result classes.
type QueryResults struct {
	Kind      QueryResultsKind
	Solutions []Solution
	Bool      bool
	Triples   []Triple
}

// Solution is one row of a SELECT result, mirroring pyoxigraph's
// QuerySolution.
type Solution struct {
	bindings map[string]Term
}

// Get returns the term bound to the variable and true, or nil and false
// when the variable is unbound in this solution (where pyoxigraph
// returns None).
func (s Solution) Get(name string) (Term, bool) {
	term, ok := s.bindings[name]
	return term, ok
}

// Query evaluates a SPARQL query against the store, fully materializing
// the result. The returned error matches ErrSyntax for malformed SPARQL,
// ErrEvaluation for queries that fail during evaluation (carrying the
// engine's message), ErrStorage for engine storage failures, and
// ErrStoreClosed after Close.
func (s *Store) Query(query string) (QueryResults, error) {
	if hasNUL(query) {
		return QueryResults{}, fmt.Errorf("%w: the query contains a NUL byte", ErrSyntax)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ptr == nil {
		return QueryResults{}, ErrStoreClosed
	}
	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))
	var kind C.int
	var cError *C.char
	cResult := C.oxigraph_query(s.ptr, cQuery, &kind, &cError)
	if cResult == nil {
		return QueryResults{}, queryError(kind, cError)
	}
	payload := C.GoString(cResult)
	C.oxigraph_free_string(cResult)

	switch kind {
	case C.OXIGRAPH_RESULT_SOLUTIONS, C.OXIGRAPH_RESULT_BOOLEAN:
		return parseJSONQueryResults(payload)
	case C.OXIGRAPH_RESULT_TRIPLES:
		return parseNTriplesQueryResults(payload)
	default:
		return QueryResults{}, fmt.Errorf("%w: unexpected result kind %d", ErrEvaluation, int(kind))
	}
}

// queryError classifies a failed query call by the kind the FFI
// reported, freeing the C error string.
func queryError(kind C.int, cError *C.char) error {
	message := "unknown query error"
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
	}
	return fmt.Errorf("%w: %s", sentinel, message)
}

// sparqlJSONResults is the SPARQL 1.1 Query Results JSON envelope for
// SELECT (results.bindings) and ASK (boolean) results.
type sparqlJSONResults struct {
	Boolean *bool `json:"boolean"`
	Results *struct {
		Bindings []map[string]json.RawMessage `json:"bindings"`
	} `json:"results"`
}

func parseJSONQueryResults(payload string) (QueryResults, error) {
	var envelope sparqlJSONResults
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return QueryResults{}, fmt.Errorf("%w: invalid SPARQL JSON results: %v", ErrEvaluation, err)
	}
	if envelope.Boolean != nil {
		return QueryResults{Kind: QueryBoolean, Bool: *envelope.Boolean}, nil
	}
	if envelope.Results == nil {
		return QueryResults{}, fmt.Errorf("%w: SPARQL JSON results carry neither bindings nor a boolean", ErrEvaluation)
	}
	solutions := make([]Solution, 0, len(envelope.Results.Bindings))
	for _, binding := range envelope.Results.Bindings {
		bindings := make(map[string]Term, len(binding))
		for variable, raw := range binding {
			term, err := ParseSPARQLJSONTerm(raw)
			if err != nil {
				return QueryResults{}, err
			}
			bindings[variable] = term
		}
		solutions = append(solutions, Solution{bindings: bindings})
	}
	return QueryResults{Kind: QuerySolutions, Solutions: solutions}, nil
}

func parseNTriplesQueryResults(payload string) (QueryResults, error) {
	triples := []Triple{}
	for _, line := range strings.Split(payload, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		quad, err := ParseNQuadsLine(line)
		if err != nil {
			return QueryResults{}, err
		}
		triples = append(triples, quad.Triple())
	}
	return QueryResults{Kind: QueryTriples, Triples: triples}, nil
}
