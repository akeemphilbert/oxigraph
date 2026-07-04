# Work on the Go binding

**Goal:** build your local checkout's Go binding and run both of its test
suites.

The binding is two layers that version together:
[`oxigraph-ffi/`](../../../oxigraph-ffi) — a Rust `staticlib` crate exposing
the C ABI ([ADR 0001](../../decisions/0001-embed-oxigraph-in-go-via-c-abi-ffi.md))
— and [`go/`](../../../go), the public Go module that wraps it with cgo. The
RDF term model is pure Go; only whole operations (query, update, load, dump)
cross the FFI as strings.

## Build the FFI crate

From the repository root (the first build compiles RocksDB and takes a
while):

```sh
cargo build -p oxigraph-ffi --release
```

The Go module's cgo directives find `target/release/liboxigraph_ffi.a`
automatically as the development fallback (prebuilt libraries in
`go/lib/<goos>_<goarch>/` win when present — see
[`go/lib/README.md`](../../../go/lib/README.md)).

## Run both test suites

The Rust side (FFI unit tests; `--release` reuses the build above):

```sh
cargo test -p oxigraph-ffi --release
```

The Go side — unit tests and the godog acceptance suite under
`go/features/` run in one command:

```sh
cd go && go test ./...
```

The Rust code is linted like the rest of the workspace (`cargo clippy -p
oxigraph-ffi --release --all-targets` and `cargo fmt`); the Go code with
`gofmt` and `go vet ./...`.

On Windows, build the GNU target — Go's cgo links with MinGW, never
MSVC: `rustup target add x86_64-pc-windows-gnu`, put a `mingw-w64` gcc
on PATH, and pass `--target x86_64-pc-windows-gnu` to the cargo
commands above.

## Refresh the prebuilt libraries

CI's `go.yml` workflow rebuilds the five platform libraries (macOS
arm64/x86_64, Linux x86_64/arm64, Windows x86_64) on every change — the
`ffi_artifact` job —
and proves a consumer builds without Rust (`consumer_no_rust`). To vendor
fresh libraries, download the `liboxigraph-ffi-*` artifacts from a workflow
run and place each `liboxigraph_ffi.a` in its `go/lib/<goos>_<goarch>/`
directory. The flow from engine release to vendored bump is documented in
[`go/lib/README.md`](../../../go/lib/README.md).
