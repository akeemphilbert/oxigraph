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

## Building

The store links the `oxigraph-ffi` static library. The cgo directives
look for it first in `lib/<goos>_<goarch>/` (prebuilt, produced by the
`go.yml` CI workflow and vendored at release time — see
[lib/README.md](./lib/README.md), including how library updates flow),
then in the repository's `target/release` as the development fallback.
Consumers with a vendored prebuilt library need nothing but Go and a C
toolchain — CI proves this by building
[`examples/quickstart`](./examples/quickstart) on a runner with the
Rust toolchain removed. A module consumer whose platform library is not
vendored can point the linker at a downloaded CI artifact instead:
`CGO_LDFLAGS=-L/path/to/dir go build ./...`.

Binding developers build the library once from the repository root (the
first build compiles RocksDB and takes a while):

```sh
cargo build -p oxigraph-ffi --release
```

Then everything in this module builds and tests with a stock cgo
toolchain (the platform C++ runtime is linked automatically: libc++ on
macOS, libstdc++ on Linux):

```sh
cd go
go test ./...
```

## Tests

- Unit tests sit beside the sources (`go test ./...`).
- The acceptance suite is Gherkin under `features/`, run with
  [godog](https://github.com/cucumber/godog) as part of the same
  `go test ./...` invocation.
