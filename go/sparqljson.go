package oxigraph

import (
	"encoding/json"
	"fmt"
)

// sparqlJSONTerm is the wire shape of one RDF term in the SPARQL 1.1 Query
// Results JSON Format. Pointer fields distinguish absent keys from empty
// strings.
type sparqlJSONTerm struct {
	Type     *string `json:"type"`
	Value    *string `json:"value"`
	Language *string `json:"xml:lang"`
	Datatype *string `json:"datatype"`
}

// ParseSPARQLJSONTerm parses a single RDF term in the SPARQL 1.1 Query
// Results JSON encoding, e.g. {"type":"uri","value":"http://..."}. The
// returned error matches ErrMalformedTerm for structurally broken terms,
// ErrUnsupportedTermType for unknown "type" values, and ErrInvalidIRI /
// ErrInvalidBlankNodeID / ErrInvalidLanguageTag when the term's value
// fails validation.
func ParseSPARQLJSONTerm(data json.RawMessage) (Term, error) {
	malformed := func(detail string) error {
		return &ParseError{Kind: ErrMalformedTerm, Input: string(data), Detail: detail}
	}

	var raw sparqlJSONTerm
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, malformed(err.Error())
	}
	if raw.Type == nil {
		return nil, malformed("a term must have a 'type' key")
	}
	switch *raw.Type {
	case "uri", "bnode", "literal", "typed-literal":
		// "typed-literal" is the legacy SPARQL 1.0 draft alias for
		// "literal"; Oxigraph accepts it on input and never emits it.
	default:
		return nil, &ParseError{
			Kind:   ErrUnsupportedTermType,
			Input:  *raw.Type,
			Detail: "supported term types are 'uri', 'bnode' and 'literal'",
		}
	}
	if raw.Value == nil {
		return nil, malformed(fmt.Sprintf("a %q term must have a 'value' key", *raw.Type))
	}

	switch *raw.Type {
	case "uri":
		n, err := NewNamedNode(*raw.Value)
		if err != nil {
			return nil, err
		}
		return n, nil
	case "bnode":
		b, err := NewBlankNode(*raw.Value)
		if err != nil {
			return nil, err
		}
		return b, nil
	default: // "literal", "typed-literal"
		if raw.Language != nil {
			if raw.Datatype != nil && *raw.Datatype != rdfLangString {
				return nil, malformed("the 'xml:lang' key conflicts with a datatype other than rdf:langString")
			}
			l, err := NewLanguageTaggedLiteral(*raw.Value, *raw.Language)
			if err != nil {
				return nil, err
			}
			return l, nil
		}
		if raw.Datatype != nil {
			datatype, err := NewNamedNode(*raw.Datatype)
			if err != nil {
				return nil, err
			}
			return NewTypedLiteral(*raw.Value, datatype), nil
		}
		return NewLiteral(*raw.Value), nil
	}
}
