package langtag

import "testing"

func TestNormalizeAccepts(t *testing.T) {
	cases := []struct{ in, want string }{
		{"en", "en"},
		{"en-US", "en-us"},
		{"fr-CA", "fr-ca"},
		{"zh-cmn-Hans-CN", "zh-cmn-hans-cn"}, // extlang + script + region
		{"sl-rozaj-biske", "sl-rozaj-biske"}, // multiple variants
		{"de-CH-1901", "de-ch-1901"},         // digit-led variant
		{"hy-Latn-IT-arevela", "hy-latn-it-arevela"},
		{"en-a-bbb-x-a-b", "en-a-bbb-x-a-b"}, // extension then private use
		{"x-private", "x-private"},           // private-use-only tag
		{"i-klingon", "i-klingon"},           // grandfathered
		{"en-GB-oed", "en-gb-oed"},           // grandfathered, mixed case
		{"az-Arab-x-AZE-derbend", "az-arab-x-aze-derbend"},
		{"und", "und"},
		{"qaa-Qaaa-QM-x-southern", "qaa-qaaa-qm-x-southern"},
		{"zh-aaa-bbb-ccc", "zh-aaa-bbb-ccc"}, // three extlangs allowed
		{"en-a-bb-c-dd", "en-a-bb-c-dd"},     // two singleton chains
		{"en-a-bb-a-cc", "en-a-bb-a-cc"},     // duplicate singleton: well-formedness only
	}
	for _, c := range cases {
		got, err := Normalize(c.in)
		if err != nil {
			t.Errorf("Normalize(%q) = error %v, want %q", c.in, err, c.want)
			continue
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeRejects(t *testing.T) {
	invalid := []string{
		"",                   // empty
		"en_US",              // underscore separator
		"123",                // digits as primary language
		"a",                  // single-letter primary language (only x/i via grandfathered)
		"en-",                // empty subtag
		"-en",                // empty subtag
		"en--us",             // empty subtag
		"verylongtag",        // primary language longer than 8
		"en-a",               // extension singleton with no subtags
		"en-x",               // private use with no subtags
		"x",                  // private-use-only with no subtags
		"en-us-us",           // duplicate region position
		"en-é",               // non-ASCII
		"en-aaaaaaaaa",       // subtag longer than 8
		"zh-aaa-bbb-ccc-ddd", // fourth extlang
		"de-419-DE",          // subtag after region that fits nothing
	}
	for _, tag := range invalid {
		if _, err := Normalize(tag); err == nil {
			t.Errorf("Normalize(%q) = nil error, want error", tag)
		}
	}
}
