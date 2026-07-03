// Package iri validates absolute IRIs following RFC 3987. It mirrors the
// subset of the oxiri crate that backs oxrdf's NamedNode: parse-only,
// absolute-only (a scheme is required), with no base resolution.
package iri

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"unicode/utf8"
)

// Validate checks that s is an absolute IRI per RFC 3987: a scheme followed
// by an optional authority (with bracketed IPv6/IPvFuture host support),
// path, query and fragment, each restricted to its component's character
// class (iunreserved / sub-delims plus component-specific extras, with the
// iprivate ranges allowed only in the query) and %XX percent-encoding.
func Validate(s string) error {
	if !utf8.ValidString(s) {
		return errors.New("the IRI is not valid UTF-8")
	}
	i := strings.IndexAny(s, ":/?#")
	if i < 0 || s[i] != ':' {
		return errors.New("no scheme found in an absolute IRI")
	}
	if err := validateScheme(s[:i]); err != nil {
		return err
	}
	rest := s[i+1:]

	var fragment string
	hasFragment := false
	if h := strings.IndexByte(rest, '#'); h >= 0 {
		fragment = rest[h+1:]
		hasFragment = true
		rest = rest[:h]
	}
	var query string
	hasQuery := false
	if q := strings.IndexByte(rest, '?'); q >= 0 {
		query = rest[q+1:]
		hasQuery = true
		rest = rest[:q]
	}

	path := rest
	if strings.HasPrefix(rest, "//") {
		authority := rest[2:]
		path = ""
		if slash := strings.IndexByte(authority, '/'); slash >= 0 {
			path = authority[slash:]
			authority = authority[:slash]
		}
		if err := validateAuthority(authority); err != nil {
			return err
		}
	}
	if err := checkChars(path, isPathChar); err != nil {
		return err
	}
	if hasQuery {
		if err := checkChars(query, isQueryChar); err != nil {
			return err
		}
	}
	if hasFragment {
		if err := checkChars(fragment, isFragmentChar); err != nil {
			return err
		}
	}
	return nil
}

func validateScheme(scheme string) error {
	if scheme == "" {
		return errors.New("no scheme found in an absolute IRI")
	}
	if !isASCIIAlpha(scheme[0]) {
		return fmt.Errorf("invalid scheme start character %q", scheme[0])
	}
	for i := 1; i < len(scheme); i++ {
		c := scheme[i]
		if !isASCIIAlpha(c) && !isASCIIDigit(c) && c != '+' && c != '-' && c != '.' {
			return fmt.Errorf("invalid scheme character %q", c)
		}
	}
	return nil
}

func validateAuthority(authority string) error {
	host := authority
	if at := strings.IndexByte(authority, '@'); at >= 0 {
		userinfo := authority[:at]
		host = authority[at+1:]
		if err := checkChars(userinfo, isUserinfoChar); err != nil {
			return err
		}
	}
	if strings.HasPrefix(host, "[") {
		end := strings.IndexByte(host, ']')
		if end < 0 {
			return errors.New("unclosed '[' in the host")
		}
		if err := validateIPLiteral(host[1:end]); err != nil {
			return err
		}
		port := host[end+1:]
		if port == "" {
			return nil
		}
		if port[0] != ':' {
			return fmt.Errorf("unexpected character %q after the ']' closing the host", port[0])
		}
		return validatePort(port[1:])
	}
	if colon := strings.IndexByte(host, ':'); colon >= 0 {
		if err := validatePort(host[colon+1:]); err != nil {
			return err
		}
		host = host[:colon]
	}
	return checkChars(host, isIUnreservedOrSubDelims)
}

func validateIPLiteral(inner string) error {
	if inner == "" {
		return errors.New("empty IP literal host")
	}
	if inner[0] == 'v' || inner[0] == 'V' {
		// IPvFuture = "v" 1*HEXDIG "." 1*( unreserved / sub-delims / ":" )
		rest := inner[1:]
		dot := strings.IndexByte(rest, '.')
		if dot <= 0 {
			return errors.New("invalid IPvFuture host")
		}
		for i := range dot {
			if !isASCIIHexDigit(rest[i]) {
				return errors.New("invalid IPvFuture host version")
			}
		}
		tail := rest[dot+1:]
		if tail == "" {
			return errors.New("empty IPvFuture host address")
		}
		for i := 0; i < len(tail); i++ {
			if !isIPvFutureChar(tail[i]) {
				return fmt.Errorf("invalid IPvFuture host character %q", tail[i])
			}
		}
		return nil
	}
	if strings.ContainsRune(inner, '%') {
		return errors.New("zone identifiers are not allowed in IPv6 hosts")
	}
	addr, err := netip.ParseAddr(inner)
	if err != nil || !addr.Is6() {
		return fmt.Errorf("invalid IPv6 host %q", inner)
	}
	return nil
}

func validatePort(port string) error {
	for i := 0; i < len(port); i++ {
		if !isASCIIDigit(port[i]) {
			return fmt.Errorf("invalid port character %q", port[i])
		}
	}
	return nil
}

// checkChars validates every rune of s against ok, allowing %XX
// percent-encoded escapes anywhere.
func checkChars(s string, ok func(rune) bool) error {
	for i := 0; i < len(s); {
		if s[i] == '%' {
			if i+3 > len(s) || !isASCIIHexDigit(s[i+1]) || !isASCIIHexDigit(s[i+2]) {
				return errors.New("invalid IRI percent encoding")
			}
			i += 3
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if !ok(r) {
			return fmt.Errorf("invalid IRI code point %q", r)
		}
		i += size
	}
	return nil
}

func isPathChar(r rune) bool {
	return isIUnreservedOrSubDelims(r) || r == ':' || r == '@' || r == '/'
}

func isQueryChar(r rune) bool {
	return isPathChar(r) || r == '?' || isIPrivate(r)
}

func isFragmentChar(r rune) bool {
	return isPathChar(r) || r == '?'
}

func isUserinfoChar(r rune) bool {
	return isIUnreservedOrSubDelims(r) || r == ':'
}

func isIUnreservedOrSubDelims(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}
	switch r {
	case '!', '$', '&', '\'', '(', ')', '*', '+', ',', '-', '.', ';', '=', '_', '~':
		return true
	}
	return isUcschar(r)
}

// isUcschar reports whether r is in RFC 3987's ucschar production: the
// BMP ranges U+00A0-D7FF, U+F900-FDCF and U+FDF0-FFEF, plus every
// supplementary plane up to U+EFFFD minus the plane-final non-characters
// and the U+E0000-E0FFF range (RFC 3987 resumes plane 14 at U+E1000).
func isUcschar(r rune) bool {
	switch {
	case r >= 0xA0 && r <= 0xD7FF,
		r >= 0xF900 && r <= 0xFDCF,
		r >= 0xFDF0 && r <= 0xFFEF:
		return true
	}
	if r < 0x10000 || r > 0xEFFFD {
		return false
	}
	if r >= 0xE0000 && r <= 0xE0FFF {
		return false
	}
	return r&0xFFFF <= 0xFFFD
}

// isIPrivate reports whether r is in RFC 3987's iprivate production,
// allowed only in the query component.
func isIPrivate(r rune) bool {
	return (r >= 0xE000 && r <= 0xF8FF) ||
		(r >= 0xF0000 && r <= 0xFFFFD) ||
		(r >= 0x100000 && r <= 0x10FFFD)
}

func isASCIIAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isASCIIDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isASCIIHexDigit(c byte) bool {
	return isASCIIDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isIPvFutureChar(c byte) bool {
	if isASCIIAlpha(c) || isASCIIDigit(c) {
		return true
	}
	switch c {
	case '-', '.', '_', '~', // unreserved
		'!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', // sub-delims
		':':
		return true
	}
	return false
}
