package oxigraph

import (
	"errors"
	"testing"
)

func mustNamedNode(t *testing.T, iri string) NamedNode {
	t.Helper()
	n, err := NewNamedNode(iri)
	if err != nil {
		t.Fatalf("NewNamedNode(%q): %v", iri, err)
	}
	return n
}

func TestNewLiteral(t *testing.T) {
	l := NewLiteral("Oxigraph")
	if got := l.Value(); got != "Oxigraph" {
		t.Errorf("Value() = %q", got)
	}
	if lang, ok := l.Language(); ok || lang != "" {
		t.Errorf("Language() = %q, %v; want \"\", false", lang, ok)
	}
	if got := l.Datatype().Value(); got != xsdString {
		t.Errorf("Datatype() = %q", got)
	}
	if got := l.String(); got != `"Oxigraph"` {
		t.Errorf("String() = %q", got)
	}
}

func TestNewLanguageTaggedLiteral(t *testing.T) {
	l, err := NewLanguageTaggedLiteral("Bonjour", "fr-CA")
	if err != nil {
		t.Fatalf("NewLanguageTaggedLiteral: %v", err)
	}
	if lang, ok := l.Language(); !ok || lang != "fr-ca" {
		t.Errorf("Language() = %q, %v; want \"fr-ca\", true", lang, ok)
	}
	if got := l.Datatype().Value(); got != rdfLangString {
		t.Errorf("Datatype() = %q", got)
	}
	if got := l.String(); got != `"Bonjour"@fr-ca` {
		t.Errorf("String() = %q", got)
	}
}

func TestNewLanguageTaggedLiteralInvalid(t *testing.T) {
	for _, lang := range []string{"", "en_US", "123"} {
		_, err := NewLanguageTaggedLiteral("hello", lang)
		if !errors.Is(err, ErrInvalidLanguageTag) {
			t.Errorf("NewLanguageTaggedLiteral(_, %q) error = %v, want ErrInvalidLanguageTag", lang, err)
		}
	}
}

func TestNewTypedLiteral(t *testing.T) {
	integer := mustNamedNode(t, "http://www.w3.org/2001/XMLSchema#integer")
	l := NewTypedLiteral("96", integer)
	if got := l.Value(); got != "96" {
		t.Errorf("Value() = %q", got)
	}
	if got := l.Datatype(); got != integer {
		t.Errorf("Datatype() = %v", got)
	}
	if got := l.String(); got != `"96"^^<http://www.w3.org/2001/XMLSchema#integer>` {
		t.Errorf("String() = %q", got)
	}
}

func TestXSDStringCollapse(t *testing.T) {
	xsd := mustNamedNode(t, xsdString)
	if NewLiteral("Oxigraph") != NewTypedLiteral("Oxigraph", xsd) {
		t.Error("an xsd:string typed literal must equal the plain literal")
	}
}

func TestLiteralEquality(t *testing.T) {
	integer := mustNamedNode(t, "http://www.w3.org/2001/XMLSchema#integer")
	fr, _ := NewLanguageTaggedLiteral("chat", "fr")
	if NewLiteral("chat") == fr {
		t.Error("a language-tagged literal must not equal the plain literal")
	}
	if NewTypedLiteral("42", integer) == NewTypedLiteral("042", integer) {
		t.Error("typed literals must compare by lexical form")
	}
}

func TestLiteralEscaping(t *testing.T) {
	cases := []struct{ value, want string }{
		{"He said \"bonjour\"\nand left", `"He said \"bonjour\"\nand left"`},
		{"C:\\graphs\\oxigraph", `"C:\\graphs\\oxigraph"`},
		{"tab\there", `"tab\there"`},
		{"bell\bform\ffeed", `"bell\bform\ffeed"`},
		{"return\rhere", `"return\rhere"`},
		{"nul\x00escape\x1f", "\"nul\\u0000escape\\u001F\""},
		{"del\x7f", "\"del\\u007F\""},
		{"noncharacter�", "\"noncharacter�\""}, // U+FFFD passes through; only U+FFFE/U+FFFF escape
		{"café", `"café"`},                     // non-ASCII is never escaped
	}
	for _, c := range cases {
		if got := NewLiteral(c.value).String(); got != c.want {
			t.Errorf("NewLiteral(%q).String() = %s, want %s", c.value, got, c.want)
		}
	}
}
