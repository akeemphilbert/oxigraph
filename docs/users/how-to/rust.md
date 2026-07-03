# Use Oxigraph as a Rust library

**Goal:** embed a SPARQL-capable RDF store inside a Rust program, no server needed.

Add the dependency:

```sh
cargo add oxigraph
```

Insert a triple and query it back:

```rust
use oxigraph::model::*;
use oxigraph::sparql::QueryResults;
use oxigraph::store::Store;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let store = Store::new()?;

    let subject = NamedNodeRef::new("http://example.com/oxigraph")?;
    let label = NamedNodeRef::new("http://www.w3.org/2000/01/rdf-schema#label")?;
    store.insert(QuadRef::new(
        subject,
        label,
        LiteralRef::new_simple_literal("Oxigraph"),
        GraphNameRef::DefaultGraph,
    ))?;

    if let QueryResults::Solutions(solutions) = store.query("SELECT ?name WHERE { ?s ?p ?name }")? {
        for solution in solutions {
            println!("{}", solution?.get("name").unwrap());
        }
    }
    Ok(())
}
```

`Store::new()` gives you an in-memory store. To persist data on disk, enable the
`rocksdb` storage backend and open a directory instead:

```rust
let store = Store::open("path/to/data")?;
```

For the full API — loading RDF files, updates, transactions, named graphs — see
[the `oxigraph` crate documentation on docs.rs](https://docs.rs/oxigraph).
