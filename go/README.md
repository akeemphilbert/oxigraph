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
`go.yml` CI workflow — see [lib/README.md](./lib/README.md), including
how library updates flow), then in the repository's `target/release`
as the development fallback. Consumers with a prebuilt library need
nothing but Go and a C toolchain — CI proves this by building
[`examples/quickstart`](./examples/quickstart) on a runner with the
Rust toolchain removed. Tagged releases attach a prebuilt library per
platform (see [Releases](#releases)); point the linker at its
directory: `CGO_LDFLAGS=-L/path/to/dir go build ./...`.

Binding developers build the library once from the repository root (the
first build compiles RocksDB and takes a while):

```sh
cargo build -p oxigraph-ffi --release
```

On Windows, build the GNU target instead — Go's cgo links with MinGW,
never MSVC (`rustup target add x86_64-pc-windows-gnu`, a `mingw-w64`
gcc on PATH, then add `--target x86_64-pc-windows-gnu`).

Then everything in this module builds and tests with a stock cgo
toolchain (the cgo directives link the platform C++ runtime: libc++ on
macOS, libstdc++ on Linux and Windows/MinGW):

```sh
cd go
go test ./...
```

## Releases

A release is cut by pushing a `go/vX.Y.Z` tag — the `go/` prefix is
how Go's module proxy resolves versions of a module living in the
`go/` subdirectory. A semver prerelease version is marked as a
prerelease on GitHub:

```sh
git tag go/v0.1.0-alpha.1
git push origin go/v0.1.0-alpha.1
```

The `go.yml` workflow runs its full gate (tests on Linux, macOS, and
Windows; the five static library builds; the no-Rust consumer proof)
and then publishes a GitHub release with each
`liboxigraph_ffi_<platform>.a.gz` and a `SHA256SUMS` file attached.

Consumers pin the module version and fetch the matching library:

```sh
go get github.com/akeemphilbert/oxigraph/go@v0.1.0-alpha.1
gh release download go/v0.1.0-alpha.1 --repo akeemphilbert/oxigraph \
  --pattern "liboxigraph_ffi_$(go env GOOS)_$(go env GOARCH).a.gz" \
  --pattern SHA256SUMS
shasum -a 256 --check --ignore-missing SHA256SUMS
mkdir -p oxigraph-lib
gunzip -c liboxigraph_ffi_*.a.gz > oxigraph-lib/liboxigraph_ffi.a
CGO_LDFLAGS="-L$PWD/oxigraph-lib" go build ./...
```

## Tests

- Unit tests sit beside the sources (`go test ./...`).
- The acceptance suite is Gherkin under `features/`, run with
  [godog](https://github.com/cucumber/godog) as part of the same
  `go test ./...` invocation.
