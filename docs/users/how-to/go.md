# Use Oxigraph from Go

**Goal:** work with RDF data and SPARQL from Go through the
`github.com/akeemphilbert/oxigraph/go` module — the engine embedded in your
process, no server to run.

Add it to your module (the binding uses cgo, so you need a C toolchain; the
Rust engine ships as a prebuilt static library, so you do **not** need Rust —
see [the module README](../../../go/README.md) for how the library is found):

```sh
go get github.com/akeemphilbert/oxigraph/go
```

Insert a quad and query it back:

```go
package main

import (
	"fmt"
	"log"

	oxigraph "github.com/akeemphilbert/oxigraph/go"
)

func main() {
	store, err := oxigraph.NewStore() // in-memory; oxigraph.Open("path/to/data") persists to disk
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	subject, err := oxigraph.NewNamedNode("http://example.com/oxigraph")
	if err != nil {
		log.Fatal(err)
	}
	label, err := oxigraph.NewNamedNode("http://www.w3.org/2000/01/rdf-schema#label")
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Add(oxigraph.NewQuad(subject, label, oxigraph.NewLiteral("Oxigraph"), oxigraph.DefaultGraph())); err != nil {
		log.Fatal(err)
	}

	results, err := store.Query(`SELECT ?name WHERE { ?s ?p ?name }`)
	if err != nil {
		log.Fatal(err)
	}
	for _, solution := range results.Solutions {
		if name, ok := solution.Get("name"); ok {
			fmt.Println(name)
		}
	}
}
```

`Query` also answers ASK (`results.Bool`) and CONSTRUCT/DESCRIBE
(`results.Triples`) — `results.Kind` says which. The store loads and dumps
whole RDF documents too (Turtle, N-Triples, N-Quads, TriG):

```go
err = store.Load(file, oxigraph.Turtle)
```

For the API reference run `go doc github.com/akeemphilbert/oxigraph/go`, and
see [the runnable example](../../../go/examples/quickstart/main.go) this
recipe mirrors.
