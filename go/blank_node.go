package oxigraph

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"unicode/utf8"
)

// BlankNode is an RDF blank node term, mirroring pyoxigraph's BlankNode.
// The zero value is not a valid term; build blank nodes with NewBlankNode
// or NewBlankNodeRandom.
type BlankNode struct {
	id string
}

// NewBlankNode creates a blank node with the given identifier, validated
// against oxrdf's blank node identifier grammar. The returned error
// matches ErrInvalidBlankNodeID.
func NewBlankNode(id string) (BlankNode, error) {
	if err := validateBlankNodeID(id); err != nil {
		return BlankNode{}, &ParseError{Kind: ErrInvalidBlankNodeID, Input: id, Detail: err.Error()}
	}
	return BlankNode{id: id}, nil
}

// NewBlankNodeRandom creates a blank node with a fresh random identifier,
// mirroring pyoxigraph's BlankNode(): the lowercase hex form of a random
// 128-bit integer, re-drawn until its first digit is a letter so the
// identifier stays distinguishable from numbered serialization labels.
func NewBlankNodeRandom() BlankNode {
	for {
		id := formatUint128Hex(rand.Uint64(), rand.Uint64())
		if id[0] >= 'a' && id[0] <= 'f' {
			return BlankNode{id: id}
		}
	}
}

// Value returns the identifier, mirroring pyoxigraph's BlankNode.value.
func (b BlankNode) Value() string { return b.id }

// String returns the N-Quads form "_:id".
func (b BlankNode) String() string { return "_:" + b.id }

func (BlankNode) isTerm()      {}
func (BlankNode) isSubject()   {}
func (BlankNode) isGraphName() {}

// formatUint128Hex renders hi<<64|lo in lowercase hex without leading
// zeros, matching Rust's "{:x}" formatting of a u128.
func formatUint128Hex(hi, lo uint64) string {
	if hi == 0 {
		return strconv.FormatUint(lo, 16)
	}
	low := strconv.FormatUint(lo, 16)
	return strconv.FormatUint(hi, 16) + strings.Repeat("0", 16-len(low)) + low
}

func validateBlankNodeID(id string) error {
	if id == "" {
		return errors.New("the identifier is empty")
	}
	if !utf8.ValidString(id) {
		return errors.New("the identifier is not valid UTF-8")
	}
	first := true
	for _, r := range id {
		if first {
			if !isBlankNodeIDStart(r) {
				return fmt.Errorf("disallowed leading character %q", r)
			}
			first = false
		} else if !isBlankNodeIDPart(r) {
			return fmt.Errorf("disallowed character %q", r)
		}
	}
	if id[len(id)-1] == '.' {
		return errors.New("the identifier must not end with a dot")
	}
	return nil
}

// isBlankNodeIDStart matches oxrdf's allowed first characters: ASCII
// alphanumerics, '_', ':' and the PN_CHARS_BASE Unicode ranges.
func isBlankNodeIDStart(r rune) bool {
	switch {
	case r >= '0' && r <= '9',
		r == '_', r == ':',
		r >= 'A' && r <= 'Z',
		r >= 'a' && r <= 'z',
		r >= 0x00C0 && r <= 0x00D6,
		r >= 0x00D8 && r <= 0x00F6,
		r >= 0x00F8 && r <= 0x02FF,
		r >= 0x0370 && r <= 0x037D,
		r >= 0x037F && r <= 0x1FFF,
		r >= 0x200C && r <= 0x200D,
		r >= 0x2070 && r <= 0x218F,
		r >= 0x2C00 && r <= 0x2FEF,
		r >= 0x3001 && r <= 0xD7FF,
		r >= 0xF900 && r <= 0xFDCF,
		r >= 0xFDF0 && r <= 0xFFFD,
		r >= 0x10000 && r <= 0xEFFFF:
		return true
	}
	return false
}

// isBlankNodeIDPart additionally allows '.', '-' and the PN_CHARS
// combining ranges in non-leading positions.
func isBlankNodeIDPart(r rune) bool {
	if isBlankNodeIDStart(r) {
		return true
	}
	switch {
	case r == '.', r == '-', r == 0x00B7,
		r >= 0x0300 && r <= 0x036F,
		r >= 0x203F && r <= 0x2040:
		return true
	}
	return false
}
