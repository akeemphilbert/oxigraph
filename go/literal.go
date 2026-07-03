package oxigraph

import (
	"fmt"
	"strings"

	"github.com/akeemphilbert/oxigraph/go/internal/langtag"
)

// IRIs of the implicit literal datatypes.
const (
	xsdString     = "http://www.w3.org/2001/XMLSchema#string"
	rdfLangString = "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"
)

// Literal is an RDF literal term, mirroring pyoxigraph's Literal. An
// empty datatype field encodes xsd:string, so == compares literals
// exactly as pyoxigraph does: by lexical form, never by value space, with
// an xsd:string-typed literal indistinguishable from the plain literal —
// and the zero value is the empty plain literal.
type Literal struct {
	value    string
	language string
	datatype string
}

// NewLiteral creates a plain literal; its datatype is xsd:string.
func NewLiteral(value string) Literal {
	return Literal{value: value}
}

// NewLanguageTaggedLiteral creates a language-tagged literal. The tag is
// validated against BCP47 and normalized to lowercase, and the datatype is
// rdf:langString. The returned error matches ErrInvalidLanguageTag.
func NewLanguageTaggedLiteral(value, language string) (Literal, error) {
	normalized, err := langtag.Normalize(language)
	if err != nil {
		return Literal{}, &ParseError{Kind: ErrInvalidLanguageTag, Input: language, Detail: err.Error()}
	}
	return Literal{value: value, language: normalized, datatype: rdfLangString}, nil
}

// NewTypedLiteral creates a literal with an explicit datatype. A datatype
// of xsd:string (or the zero-value NamedNode) yields the exact
// representation NewLiteral produces, so the two compare equal, as in
// oxrdf. Passing rdf:langString does not create a language-tagged literal
// — use NewLanguageTaggedLiteral for that.
func NewTypedLiteral(value string, datatype NamedNode) Literal {
	if datatype.iri == xsdString {
		return Literal{value: value}
	}
	return Literal{value: value, datatype: datatype.iri}
}

// Value returns the lexical form, mirroring pyoxigraph's Literal.value.
func (l Literal) Value() string { return l.value }

// Language returns the language tag and true for a language-tagged
// literal, or "" and false otherwise (where pyoxigraph's Literal.language
// is None).
func (l Literal) Language() (string, bool) { return l.language, l.language != "" }

// Datatype returns the datatype IRI, mirroring pyoxigraph's
// Literal.datatype: xsd:string for plain literals, rdf:langString for
// language-tagged ones.
func (l Literal) Datatype() NamedNode {
	if l.datatype == "" {
		return NamedNode{iri: xsdString}
	}
	return NamedNode{iri: l.datatype}
}

// String returns the N-Quads form of the literal.
func (l Literal) String() string {
	var b strings.Builder
	writeQuotedString(&b, l.value)
	if l.language != "" {
		b.WriteByte('@')
		b.WriteString(l.language)
	} else if l.datatype != "" {
		b.WriteString("^^<")
		b.WriteString(l.datatype)
		b.WriteByte('>')
	}
	return b.String()
}

func (Literal) isTerm() {}

// writeQuotedString writes value quoted and escaped exactly as oxrdf's
// print_quoted_str: named escapes for backspace, tab, newline, form feed,
// carriage return, '"' and '\', and uppercase \uXXXX for the remaining C0
// controls, DEL and the U+FFFE/U+FFFF non-characters.
func writeQuotedString(b *strings.Builder, value string) {
	b.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			if r < 0x20 || r == 0x7F || r == 0xFFFE || r == 0xFFFF {
				fmt.Fprintf(b, `\u%04X`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}
