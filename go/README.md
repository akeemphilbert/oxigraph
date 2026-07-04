# Oxigraph for Go

[Oxigraph](https://oxigraph.org) is a graph database implementing the
[SPARQL](https://www.w3.org/TR/sparql11-overview/) standard. This module
embeds it in Go programs with a
[pyoxigraph](https://oxigraph.org/pyoxigraph/)-style API: RDF terms
(`NamedNode`, `BlankNode`, `Literal`, `Triple`, `Quad`) are pure Go
values, and the `Store` is backed by the Rust engine through the
`oxigraph-ffi` C ABI (see `docs/decisions/0001-embed-oxigraph-in-go-via-c-abi-ffi.md`).

```go
store, err := oxigraph.Open("./data") // or oxigraph.NewStore() in memory
if err != nil {
    log.Fatal(err)
}
defer store.Close()
```

## Development builds

The store requires the `oxigraph-ffi` static library. Build it once from
the repository root (the first build compiles RocksDB and takes a while):

```sh
cargo build -p oxigraph-ffi --release
```

Then everything in this module builds and tests with a stock cgo
toolchain:

```sh
cd go
go test ./...
```

The cgo directives in `store.go` link `target/release/liboxigraph_ffi.a`
plus the platform C++ runtime (`-lc++` on macOS). Prebuilt static
libraries for consumers outside this repository ship in a later story, so
`go get` users will not need Rust.

## Tests

- Unit tests sit beside the sources (`go test ./...`).
- The acceptance suite is Gherkin under `features/`, run with
  [godog](https://github.com/cucumber/godog) as part of the same
  `go test ./...` invocation.
