package oxigraph

import (
	"errors"
	"fmt"
)

// Sentinel errors classifying every failure this package can produce.
// Match them with errors.Is; the concrete error is usually a *ParseError.
var (
	ErrInvalidIRI          = errors.New("oxigraph: invalid IRI")
	ErrInvalidLanguageTag  = errors.New("oxigraph: invalid language tag")
	ErrInvalidBlankNodeID  = errors.New("oxigraph: invalid blank node identifier")
	ErrUnsupportedTermType = errors.New("oxigraph: unsupported term type")
	ErrMalformedTerm       = errors.New("oxigraph: malformed term")
	ErrSyntax              = errors.New("oxigraph: syntax error")
	ErrStorage             = errors.New("oxigraph: storage error")
	ErrStoreClosed         = errors.New("oxigraph: store already closed")
	ErrEvaluation          = errors.New("oxigraph: evaluation error")
	ErrUnsupportedFormat   = errors.New("oxigraph: unsupported RDF format")
)

// ParseError reports why an input was rejected. Kind is one of the Err*
// sentinels, exposed through Unwrap so errors.Is(err, ErrInvalidIRI) and
// friends work on any error returned by this package.
type ParseError struct {
	Kind   error  // the Err* sentinel classifying the failure
	Input  string // the rejected input
	Detail string // what exactly was wrong with it
}

func (e *ParseError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("%s: %q", e.Kind, e.Input)
	}
	return fmt.Sprintf("%s %q: %s", e.Kind, e.Input, e.Detail)
}

func (e *ParseError) Unwrap() error { return e.Kind }
