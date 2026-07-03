package iri

import "testing"

func TestValidateAccepts(t *testing.T) {
	valid := []string{
		"http://example.com",
		"http://example.com/",
		"http://example.com/person/alice",
		"http://example.com/a%20b",
		"https://user:secret@example.com:8080/path?q=1&r=2#frag",
		"http://[2001:db8::1]/path",
		"http://[::ffff:192.168.0.1]:80/",
		"http://[vFe.some+future:host]/",
		"urn:isbn:0451450523",
		"mailto:akeem@example.com",
		"file:///tmp/data",
		"http://example.com/été",        // ucschar in path
		"http://example.com/?private=", // iprivate in query
		"a+b-c.d://host",                // full scheme charset
		"http://example.com/path?q#f",   // query and fragment
		"scheme:",                       // empty path
		"http://",                       // empty host
		"http://example.com:/",          // empty port
		"tag:example.com,2026:x?y#z",
	}
	for _, iri := range valid {
		if err := Validate(iri); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", iri, err)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	invalid := []string{
		"",                         // empty
		"person/alice",             // no scheme
		"//example.com/x",          // no scheme
		"http://example.com/a b",   // space in path
		"1http://example.com",      // scheme starting with a digit
		"http://example.com/a%2",   // truncated percent encoding
		"http://example.com/a%zz",  // invalid percent digits
		"http://example.com/<x>",   // forbidden code point
		"http://example.com/",     // iprivate outside the query
		"http://[2001:db8::1/",     // unclosed bracket
		"http://[not-an-ip]/",      // invalid IPv6
		"http://[fe80::1%25eth0]/", // zone identifier
		"http://example.com:8a/",   // non-digit port
		"http://example.com/x#a#b", // '#' inside the fragment
		"http://example.com/\x7f",  // DEL control character
		"http://example.com/",     // C1 control character (not ucschar)
	}
	for _, iri := range invalid {
		if err := Validate(iri); err == nil {
			t.Errorf("Validate(%q) = nil, want error", iri)
		}
	}
}
