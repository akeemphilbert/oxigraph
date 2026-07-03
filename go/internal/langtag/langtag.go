// Package langtag validates BCP47 language tags the way the oxilangtag
// crate does when called from oxrdf: the whole tag is ASCII-lowercased
// first and the grammar is checked on the lowercased form, so no BCP47
// mixed-case canonicalization (script title case, region upper case) is
// ever applied.
package langtag

import (
	"errors"
	"fmt"
	"strings"
)

// grandfathered holds the RFC 5646 grandfathered tags, pre-lowercased to
// match the normalization applied before lookup.
var grandfathered = map[string]bool{
	"art-lojban":  true,
	"cel-gaulish": true,
	"en-gb-oed":   true,
	"i-ami":       true,
	"i-bnn":       true,
	"i-default":   true,
	"i-enochian":  true,
	"i-hak":       true,
	"i-klingon":   true,
	"i-lux":       true,
	"i-mingo":     true,
	"i-navajo":    true,
	"i-pwn":       true,
	"i-tao":       true,
	"i-tay":       true,
	"i-tsu":       true,
	"no-bok":      true,
	"no-nyn":      true,
	"sgn-be-fr":   true,
	"sgn-be-nl":   true,
	"sgn-ch-de":   true,
	"zh-guoyu":    true,
	"zh-hakka":    true,
	"zh-min":      true,
	"zh-min-nan":  true,
	"zh-xiang":    true,
}

// Normalize ASCII-lowercases tag and validates the result against the
// RFC 5646 langtag grammar, returning the lowercased form.
func Normalize(tag string) (string, error) {
	lower := asciiLower(tag)
	if err := validate(lower); err != nil {
		return "", err
	}
	return lower, nil
}

func asciiLower(s string) string {
	if !strings.ContainsFunc(s, func(r rune) bool { return r >= 'A' && r <= 'Z' }) {
		return s
	}
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

func validate(tag string) error {
	if tag == "" {
		return errors.New("the language tag is empty")
	}
	if grandfathered[tag] {
		return nil
	}
	subtags := strings.Split(tag, "-")
	for _, subtag := range subtags {
		if subtag == "" {
			return errors.New("empty subtag")
		}
		if len(subtag) > 8 {
			return fmt.Errorf("subtag %q is longer than 8 characters", subtag)
		}
		for i := 0; i < len(subtag); i++ {
			if !isAlphanum(subtag[i]) {
				return fmt.Errorf("invalid character %q in subtag %q", subtag[i], subtag)
			}
		}
	}

	// privateuse-only tag: "x" 1*("-" (1*8alphanum))
	if subtags[0] == "x" {
		if len(subtags) == 1 {
			return errors.New("a private use tag needs at least one subtag after 'x'")
		}
		return nil
	}

	// language = 2*3ALPHA ["-" extlang] / 4ALPHA / 5*8ALPHA
	primary := subtags[0]
	if len(primary) < 2 || !isAllAlpha(primary) {
		return fmt.Errorf("invalid primary language subtag %q", primary)
	}
	i := 1
	if len(primary) < 4 {
		// extlang = 3ALPHA *2("-" 3ALPHA)
		for n := 0; n < 3 && i < len(subtags) && len(subtags[i]) == 3 && isAllAlpha(subtags[i]); n++ {
			i++
		}
	}
	// script = 4ALPHA
	if i < len(subtags) && len(subtags[i]) == 4 && isAllAlpha(subtags[i]) {
		i++
	}
	// region = 2ALPHA / 3DIGIT
	if i < len(subtags) &&
		((len(subtags[i]) == 2 && isAllAlpha(subtags[i])) ||
			(len(subtags[i]) == 3 && isAllDigit(subtags[i]))) {
		i++
	}
	// variant = 5*8alphanum / (DIGIT 3alphanum)
	for i < len(subtags) && (len(subtags[i]) >= 5 || (len(subtags[i]) == 4 && isDigit(subtags[i][0]))) {
		i++
	}
	// extension = singleton 1*("-" (2*8alphanum)), singleton != "x"
	for i < len(subtags) && len(subtags[i]) == 1 && subtags[i] != "x" {
		i++
		n := 0
		for i < len(subtags) && len(subtags[i]) >= 2 {
			i++
			n++
		}
		if n == 0 {
			return errors.New("an extension needs at least one subtag after its singleton")
		}
	}
	// privateuse = "x" 1*("-" (1*8alphanum))
	if i < len(subtags) && subtags[i] == "x" {
		if i == len(subtags)-1 {
			return errors.New("a private use section needs at least one subtag after 'x'")
		}
		i = len(subtags)
	}
	if i != len(subtags) {
		return fmt.Errorf("unexpected subtag %q", subtags[i])
	}
	return nil
}

func isAlphanum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || isDigit(c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAllAlpha(s string) bool {
	for i := 0; i < len(s); i++ {
		if !((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z')) {
			return false
		}
	}
	return true
}

func isAllDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return false
		}
	}
	return true
}
