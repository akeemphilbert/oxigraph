# Getting started: from install to your first SPARQL query

In this tutorial you will install Oxigraph, start its server, load a small RDF
file, and run your first SPARQL query against it. It takes about fifteen minutes.

## 1. Install the Oxigraph CLI

Pick one of the following.

**With cargo** (needs a [Rust toolchain](https://rustup.rs/) and a C/C++ compiler
such as clang — Oxigraph's storage engine is compiled from source):

```sh
cargo install oxigraph-cli
```

**With Docker** (no Rust toolchain needed):

```sh
docker pull ghcr.io/oxigraph/oxigraph
```

The rest of this tutorial shows the `oxigraph` binary. If you use Docker, skip
step 3 and start the server directly (the Docker equivalent of step 4):

```sh
docker run --rm -v "$PWD"/data:/data -p 7878:7878 ghcr.io/oxigraph/oxigraph serve --location /data --bind 0.0.0.0:7878
```

Then, once the server is up, load `example.ttl` from step 2 over HTTP instead of
with the `load` command — the container cannot see files outside the mounted
`data` volume:

```sh
curl -X POST 'http://localhost:7878/store?default' \
  -H 'Content-Type: text/turtle' --data-binary @example.ttl
```

Check the install worked:

```sh
oxigraph --version
```

## 2. Create a small RDF file

Save this as `example.ttl`. It describes one thing — Oxigraph itself — in the
[Turtle](https://www.w3.org/TR/turtle/) format:

```turtle
@prefix ex: <http://example.com/> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

ex:oxigraph a ex:Database ;
    rdfs:label "Oxigraph" ;
    ex:writtenIn "Rust" .
```

## 3. Load it into a store

Oxigraph keeps its data in a directory you choose. Load the file into a new store
under `./data`:

```sh
oxigraph load --location ./data --file example.ttl
```

## 4. Start the server

```sh
oxigraph serve --location ./data
```

The server is now listening on <http://localhost:7878>. Leave it running and open
that address in your browser: you get a query editor served by Oxigraph.

## 5. Run your first query

Paste this into the query editor and run it:

```sparql
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>

SELECT ?name WHERE {
  ?database rdfs:label ?name .
}
```

You should see one result: `"Oxigraph"`.

The same query works from a second terminal with `curl`, because the server
implements the standard [SPARQL Protocol](https://www.w3.org/TR/sparql11-protocol/):

```sh
curl -X POST http://localhost:7878/query \
  -H 'Content-Type: application/sparql-query' \
  --data 'PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#> SELECT ?name WHERE { ?database rdfs:label ?name }'
```

## Where to go next

- Query the server from your own programs: [over HTTP](../how-to/http.md), from
  [Rust](../how-to/rust.md), [Python](../how-to/python.md), or
  [JavaScript](../how-to/javascript.md).
- All server options: `oxigraph serve --help`, and the
  [CLI README](../../../cli/README.md).
