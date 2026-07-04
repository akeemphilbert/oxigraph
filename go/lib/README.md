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
| `windows_amd64/` | Windows x86_64 (MinGW — built for `x86_64-pc-windows-gnu`; Go's cgo links with MinGW, so an MSVC-built library cannot be used) |

The `.a` files are not committed; the `go.yml` CI workflow builds all
five (job `ffi_artifact`, mirroring `artifacts.yml`'s runner matrix) and
uploads them as workflow artifacts, and its `consumer_no_rust` job
proves a `go build` succeeds with one of them on a runner with the Rust
toolchain removed.

## How library updates flow

1. The engine changes (a new Oxigraph release lands on `main`, or the
   `oxigraph-ffi` crate changes).
2. CI rebuilds the five static libraries on the next push — no manual
   step, the artifact job runs on every change.
3. A maintainer cuts a release by tagging — `git tag go/v0.1.0-alpha.1
   && git push origin go/v0.1.0-alpha.1`. The `go/` tag prefix is what
   lets Go's module proxy resolve `go get …/oxigraph/go@v0.1.0-alpha.1`
   for a module in the `go/` subdirectory, and a semver prerelease
   version is marked as a prerelease on GitHub. The tag runs the full `go.yml` gate (tests
   on three OSes, the five library builds, the no-Rust consumer proof)
   and publishes a GitHub release with each library attached as
   `liboxigraph_ffi_<platform>.a.gz` plus a `SHA256SUMS` file. The
   libraries are release assets, not files committed into the tag:
   GitHub rejects files over 100 MB and the libraries are already
   73–97 MB each.
4. Consumers download the asset for their platform, check it against
   `SHA256SUMS`, and place the unpacked `liboxigraph_ffi.a` where the
   linker looks (its directory above in a repository checkout, or any
   directory passed via `CGO_LDFLAGS=-L…`).

## Binding developers (local fallback)

You do not need the prebuilt libraries to work on the binding — build
the library once with cargo and the fallback `-L` path picks it up:

```sh
cargo build -p oxigraph-ffi --release
cd go && go test ./...
```
