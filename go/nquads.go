package oxigraph

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ParseNQuadsLine parses a single N-Quads (or N-Triples) statement line,
// including its terminating dot. A three-term line yields a quad in the
// default graph. Structural problems return an error matching ErrSyntax;
// an invalid IRI, blank node identifier or language tag inside an
// otherwise well-formed line surfaces as its own error kind.
func ParseNQuadsLine(line string) (Quad, error) {
	syntaxErr := func(detail string) error {
		return &ParseError{Kind: ErrSyntax, Input: line, Detail: detail}
	}

	subjectTerm, rest, err := readTerm(line)
	if err != nil {
		return Quad{}, err
	}
	subject, ok := subjectTerm.(Subject)
	if !ok {
		return Quad{}, syntaxErr("the subject of a statement must be a named or blank node")
	}
	predicateTerm, rest, err := readTerm(rest)
	if err != nil {
		return Quad{}, err
	}
	predicate, ok := predicateTerm.(NamedNode)
	if !ok {
		return Quad{}, syntaxErr("the predicate of a statement must be a named node")
	}
	object, rest, err := readTerm(rest)
	if err != nil {
		return Quad{}, err
	}

	graphName := GraphName(DefaultGraph())
	rest = skipWhitespace(rest)
	if !strings.HasPrefix(rest, ".") {
		graphTerm, afterGraph, err := readTerm(rest)
		if err != nil {
			return Quad{}, err
		}
		name, ok := graphTerm.(GraphName)
		if !ok {
			return Quad{}, syntaxErr("the graph name of a statement must be a named or blank node")
		}
		graphName = name
		rest = skipWhitespace(afterGraph)
		if !strings.HasPrefix(rest, ".") {
			return Quad{}, syntaxErr("a statement must end with '.'")
		}
	}
	if skipWhitespace(rest[1:]) != "" {
		return Quad{}, syntaxErr("unexpected content after the terminating '.'")
	}
	return Quad{Subject: subject, Predicate: predicate, Object: object, GraphName: graphName}, nil
}

// NQuadsLine returns the quad as a full N-Quads statement line with the
// terminating dot, the form ParseNQuadsLine accepts.
func (q Quad) NQuadsLine() string {
	return q.String() + " ."
}

// ParseTerm parses a single term in N-Quads syntax: <iri>, _:id, or a
// literal with an optional language tag or datatype.
func ParseTerm(s string) (Term, error) {
	term, rest, err := readTerm(s)
	if err != nil {
		return nil, err
	}
	if skipWhitespace(rest) != "" {
		return nil, &ParseError{Kind: ErrSyntax, Input: s, Detail: "unexpected content after the term"}
	}
	return term, nil
}

// parseTermAs parses a term and requires it to be a specific concrete
// term type, described by label in the error on mismatch.
func parseTermAs[T Term](s, label string) (T, error) {
	var zero T
	term, err := ParseTerm(s)
	if err != nil {
		return zero, err
	}
	t, ok := term.(T)
	if !ok {
		return zero, &ParseError{Kind: ErrSyntax, Input: s, Detail: "expected a " + label}
	}
	return t, nil
}

// ParseNamedNode parses a named node in N-Quads syntax: <iri>.
func ParseNamedNode(s string) (NamedNode, error) {
	return parseTermAs[NamedNode](s, "named node")
}

// ParseBlankNode parses a blank node in N-Quads syntax: _:id.
func ParseBlankNode(s string) (BlankNode, error) {
	return parseTermAs[BlankNode](s, "blank node")
}

// ParseLiteral parses a literal in N-Quads syntax:
// "value"[@lang | ^^<datatype>].
func ParseLiteral(s string) (Literal, error) {
	return parseTermAs[Literal](s, "literal")
}

// readTerm reads one term from the front of s, after leading whitespace,
// returning the term and the unconsumed remainder.
func readTerm(s string) (Term, string, error) {
	s = skipWhitespace(s)
	if s == "" {
		return nil, "", &ParseError{Kind: ErrSyntax, Input: s, Detail: "expected a term, found the end of the input"}
	}
	switch s[0] {
	case '<':
		return readNamedNodeTerm(s)
	case '_':
		return readBlankNodeTerm(s)
	case '"':
		return readLiteralTerm(s)
	default:
		return nil, "", &ParseError{Kind: ErrSyntax, Input: s, Detail: fmt.Sprintf("expected a term, found %q", s[0])}
	}
}

func readNamedNodeTerm(s string) (Term, string, error) {
	n, rest, err := readNamedNode(s)
	if err != nil {
		return nil, "", err
	}
	return n, rest, nil
}

func readBlankNodeTerm(s string) (Term, string, error) {
	if !strings.HasPrefix(s, "_:") {
		return nil, "", &ParseError{Kind: ErrSyntax, Input: s, Detail: "a blank node must start with '_:'"}
	}
	start := len("_:")
	i := start
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !isBlankNodeIDPart(r) {
			break
		}
		i += size
	}
	// Trailing dots belong to the statement, not the identifier.
	end := i
	for end > start && s[end-1] == '.' {
		end--
	}
	b, err := NewBlankNode(s[start:end])
	if err != nil {
		return nil, "", err
	}
	return b, s[end:], nil
}

func readLiteralTerm(s string) (Term, string, error) {
	value, rest, err := readQuotedString(s)
	if err != nil {
		return nil, "", err
	}
	if strings.HasPrefix(rest, "@") {
		i := 1
		for i < len(rest) && (isASCIIAlphanumeric(rest[i]) || rest[i] == '-') {
			i++
		}
		l, err := NewLanguageTaggedLiteral(value, rest[1:i])
		if err != nil {
			return nil, "", err
		}
		return l, rest[i:], nil
	}
	if strings.HasPrefix(rest, "^^") {
		afterCarets := rest[2:]
		if !strings.HasPrefix(afterCarets, "<") {
			return nil, "", &ParseError{Kind: ErrSyntax, Input: s, Detail: "expected '<' after '^^'"}
		}
		datatype, rest, err := readNamedNode(afterCarets)
		if err != nil {
			return nil, "", err
		}
		// Mirrors oxttl: a literal without a language tag must not carry
		// the rdf:langString datatype.
		if datatype.iri == rdfLangString {
			return nil, "", &ParseError{Kind: ErrSyntax, Input: s, Detail: "a literal without a language tag cannot have the rdf:langString datatype"}
		}
		return NewTypedLiteral(value, datatype), rest, nil
	}
	return NewLiteral(value), rest, nil
}

// readNamedNode reads an IRIREF ("<...>", with \uXXXX / \UXXXXXXXX escapes
// decoded) from the front of s.
func readNamedNode(s string) (NamedNode, string, error) {
	var b strings.Builder
	i := 1 // past '<'
	for i < len(s) {
		switch s[i] {
		case '>':
			n, err := NewNamedNode(b.String())
			if err != nil {
				return NamedNode{}, "", err
			}
			return n, s[i+1:], nil
		case '\\':
			r, next, err := readEscape(s, i, false)
			if err != nil {
				return NamedNode{}, "", err
			}
			b.WriteRune(r)
			i = next
		default:
			r, size := utf8.DecodeRuneInString(s[i:])
			b.WriteRune(r)
			i += size
		}
	}
	return NamedNode{}, "", &ParseError{Kind: ErrSyntax, Input: s, Detail: "unclosed '<'"}
}

// readQuotedString reads a STRING_LITERAL_QUOTE from the front of s
// (starting at its opening '"'), decoding escape sequences, and returns
// the value and the remainder just past the closing '"'.
func readQuotedString(s string) (string, string, error) {
	var b strings.Builder
	i := 1 // past '"'
	for i < len(s) {
		switch s[i] {
		case '"':
			return b.String(), s[i+1:], nil
		case '\\':
			r, next, err := readEscape(s, i, true)
			if err != nil {
				return "", "", err
			}
			b.WriteRune(r)
			i = next
		case '\n', '\r':
			return "", "", &ParseError{Kind: ErrSyntax, Input: s, Detail: `line jumps are not allowed in string literals, use \n`}
		default:
			r, size := utf8.DecodeRuneInString(s[i:])
			b.WriteRune(r)
			i += size
		}
	}
	return "", "", &ParseError{Kind: ErrSyntax, Input: s, Detail: `unclosed '"'`}
}

// readEscape decodes the escape sequence whose backslash sits at s[i],
// returning the rune and the index just past the sequence. String literals
// accept the named ECHAR escapes and \uXXXX / \UXXXXXXXX; IRIs (withEchar
// false) accept only the latter two.
func readEscape(s string, i int, withEchar bool) (rune, int, error) {
	if i+1 >= len(s) {
		return 0, 0, &ParseError{Kind: ErrSyntax, Input: s, Detail: "trailing backslash"}
	}
	c := s[i+1]
	if withEchar {
		switch c {
		case 't':
			return '\t', i + 2, nil
		case 'b':
			return '\b', i + 2, nil
		case 'n':
			return '\n', i + 2, nil
		case 'r':
			return '\r', i + 2, nil
		case 'f':
			return '\f', i + 2, nil
		case '"':
			return '"', i + 2, nil
		case '\'':
			return '\'', i + 2, nil
		case '\\':
			return '\\', i + 2, nil
		}
	}
	switch c {
	case 'u':
		return readHexRune(s, i+2, 4)
	case 'U':
		return readHexRune(s, i+2, 8)
	default:
		return 0, 0, &ParseError{Kind: ErrSyntax, Input: s, Detail: fmt.Sprintf("unexpected escape character %q", c)}
	}
}

func readHexRune(s string, start, digits int) (rune, int, error) {
	if start+digits > len(s) {
		return 0, 0, &ParseError{Kind: ErrSyntax, Input: s, Detail: "truncated unicode escape"}
	}
	hex := s[start : start+digits]
	for j := range digits {
		if !isASCIIHexDigit(hex[j]) {
			return 0, 0, &ParseError{Kind: ErrSyntax, Input: s, Detail: fmt.Sprintf("invalid unicode escape digit %q", hex[j])}
		}
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil || !utf8.ValidRune(rune(v)) {
		return 0, 0, &ParseError{Kind: ErrSyntax, Input: s, Detail: fmt.Sprintf(`the unicode escape \%s%s is not a valid code point`, map[int]string{4: "u", 8: "U"}[digits], hex)}
	}
	return rune(v), start + digits, nil
}

func skipWhitespace(s string) string {
	return strings.TrimLeft(s, " \t")
}

func isASCIIAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isASCIIHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
