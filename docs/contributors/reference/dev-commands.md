# Development commands

The commands you'll reach for day to day, matching what CI enforces
(`.github/workflows/tests.yml`). Everything runs from the repository root
unless a directory is given.

## Build

| Task | Command |
| ---- | ------- |
| Build everything | `cargo build` |
| Release build of the CLI | `cargo build --release -p oxigraph-cli` |

## Test

| Task | Command |
| ---- | ------- |
| Whole workspace | `cargo test` |
| One crate | `cargo test -p oxrdf` |
| Filter by test name | `cargo test -p oxigraph store` |
| Conformance suites | `cargo test -p oxigraph-testsuite` |
| One suite file | `cargo test -p oxigraph-testsuite --test sparql` |
| One manifest, with report | `cargo run -p oxigraph-testsuite -- <manifest-url>` |
| Python bindings | `pip install .` then `python -m unittest` (from `python/`, `python/tests/`) |
| JavaScript bindings | `npm test` (from `js/`) |

## Lint & format

| Task | Command |
| ---- | ------- |
| Format | `cargo fmt` |
| Format check (CI) | `cargo fmt -- --check` |
| Lints (CI treats warnings as errors) | `cargo clippy --all-targets -- -D warnings -D clippy::all` |
| Spelling (CI) | `typos` — install with `cargo install typos-cli` |
| Dependency policy (CI) | `cargo deny check` — install with `cargo install cargo-deny` |

## Performance & robustness

| Task | Command |
| ---- | ------- |
| Store microbenchmarks | `cargo bench -p oxigraph` |
| Parser microbenchmarks | `cargo bench -p oxigraph-testsuite` |
| End-to-end benchmark | `bench/bsbm_oxigraph.sh` (see [the benchmarks guide](../how-to/benchmarks.md)) |
| Fuzz a parser | `cargo +nightly fuzz run <target>` (see [the fuzzing guide](../how-to/fuzzing.md)) |

The how-to guides cover the workflows behind these one-liners; the
[first-change tutorial](../tutorials/first-change.md) shows the minimal gate to
run before every pull request.
