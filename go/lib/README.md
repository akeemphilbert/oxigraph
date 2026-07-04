# Prebuilt oxigraph-ffi libraries

The cgo directives in this module look for `liboxigraph_ffi.a` first in
the per-platform directory here, then in the repository's
`target/release` as the development fallback:

| Directory | Platform |
|---|---|
| `darwin_arm64/` | macOS Apple Silicon |
| `darwin_amd64/` | macOS Intel |
| `linux_amd64/` | Linux x86_64 (glibc) |
| `linux_arm64/` | Linux aarch64 (glibc) |

The `.a` files are not committed; the `go.yml` CI workflow builds all
four (job `ffi_artifact`, mirroring `artifacts.yml`'s runner matrix) and
uploads them as workflow artifacts, and its `consumer_no_rust` job
proves a `go build` succeeds with one of them on a runner with the Rust
toolchain removed.

## How library updates flow

1. The engine changes (a new Oxigraph release lands on `main`, or the
   `oxigraph-ffi` crate changes).
2. CI rebuilds the four static libraries on the next push — no manual
   step, the artifact job runs on every change.
3. A release vendors the artifacts: download the four
   `liboxigraph-ffi-*` artifacts from the release build and place each
   `liboxigraph_ffi.a` in its directory above (or attach them to the
   GitHub release for consumers who vendor themselves).

## Binding developers (local fallback)

You do not need the prebuilt libraries to work on the binding — build
the library once with cargo and the fallback `-L` path picks it up:

```sh
cargo build -p oxigraph-ffi --release
cd go && go test ./...
```
