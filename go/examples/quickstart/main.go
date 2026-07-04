// Command quickstart is the smallest useful Oxigraph Go program: it
// opens an in-memory store, adds a quad, and queries it back. CI builds
// it on a runner with no Rust toolchain to prove the prebuilt static
// library is all a consumer needs.
package main

import (
	"fmt"
	"log"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func main() {
	store, err := oxigraph.NewStore()
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	book, err := oxigraph.NewNamedNode("http://example.com/book/1")
	if err != nil {
		log.Fatal(err)
	}
	title, err := oxigraph.NewNamedNode("http://purl.org/dc/terms/title")
	if err != nil {
		log.Fatal(err)
	}
	object, err := oxigraph.NewLanguageTaggedLiteral("Le Petit Prince", "fr")
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Add(oxigraph.NewQuad(book, title, object, oxigraph.DefaultGraph())); err != nil {
		log.Fatal(err)
	}

	results, err := store.Query(`SELECT ?title WHERE { ?book <http://purl.org/dc/terms/title> ?title }`)
	if err != nil {
		log.Fatal(err)
	}
	for _, solution := range results.Solutions {
		if value, ok := solution.Get("title"); ok {
			fmt.Println(value)
		}
	}
}
