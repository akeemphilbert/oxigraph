# Crate map and repository layout

## Workspace crates

| Crate | Path | Responsibility |
| ----- | ---- | -------------- |
| [`oxigraph`](https://docs.rs/oxigraph) | `lib/oxigraph` | The database itself: the `Store` API, storage backends (RocksDB and in-memory), SPARQL entry points |
| [`oxrdf`](https://docs.rs/oxrdf) | `lib/oxrdf` | RDF data structures everything else shares: terms, triples, quads, graphs, datasets |
| [`oxrdfio`](https://docs.rs/oxrdfio) | `lib/oxrdfio` | Unified RDF parser/serializer API over the format crates below |
| [`oxttl`](https://docs.rs/oxttl) | `lib/oxttl` | Turtle, TriG, N-Triples, N-Quads, and N3 parsing/serialization |
| [`oxrdfxml`](https://docs.rs/oxrdfxml) | `lib/oxrdfxml` | RDF/XML parsing/serialization |
| [`oxjsonld`](https://docs.rs/oxjsonld) | `lib/oxjsonld` | JSON-LD parsing/serialization |
| [`oxsdatatypes`](https://docs.rs/oxsdatatypes) | `lib/oxsdatatypes` | XML Schema datatypes (`xsd:decimal`, `xsd:dateTime`, …) used by SPARQL |
| [`spargebra`](https://docs.rs/spargebra) | `lib/spargebra` | SPARQL parser: query/update strings → algebra |
| [`sparopt`](https://docs.rs/sparopt) | `lib/sparopt` | SPARQL optimizer: rewrites the algebra |
| [`spareval`](https://docs.rs/spareval) | `lib/spareval` | SPARQL evaluator: executes the algebra against a store |
| [`sparesults`](https://docs.rs/sparesults) | `lib/sparesults` | SPARQL result formats: JSON, XML, CSV/TSV |
| [`spargeo`](https://docs.rs/spargeo) | `lib/spargeo` | GeoSPARQL functions |
| [`sparql-smith`](https://docs.rs/sparql-smith) | `lib/sparql-smith` | SPARQL test-case generator, used by the fuzzers |
| `oxigraph-cli` | `cli` | The `oxigraph` binary: command-line tools and the SPARQL HTTP server |
| `oxigraph-js` | `js` | JavaScript/WebAssembly bindings ([`oxigraph` on npm](https://www.npmjs.com/package/oxigraph)) |
| `pyoxigraph` | `python` | Python bindings ([`pyoxigraph` on PyPI](https://pypi.org/project/pyoxigraph/)) |
| `oxrocksdb-sys` | `oxrocksdb-sys` | RocksDB C++ bindings; builds the vendored RocksDB submodule |
| `oxigraph-testsuite` | `testsuite` | W3C conformance and regression test harness (not published) |

How they fit together is the subject of the
[architecture overview](../explanation/architecture.md).

## Repository layout

| Directory | Contents |
| --------- | -------- |
| `lib/` | The library crates above |
| `cli/` | The CLI/server crate (`server/` is a legacy symlink to it) |
| `python/`, `js/` | The language bindings |
| `testsuite/` | Conformance harness, plus the W3C test suites as submodules (`rdf-tests`, `json-ld-api`, `rdf-canon`, `N3`, …) and Oxigraph's own tests in `oxigraph-tests/` |
| `oxrocksdb-sys/` | RocksDB bindings, with `rocksdb/` and `lz4/` submodules |
| `bench/` | BSBM end-to-end benchmark scripts (`bsbm-tools` submodule) |
| `fuzz/` | cargo-fuzz targets (`fuzz_targets/`) and corpus tooling |
| `docs/` | This documentation |
| `lints/` | Maintenance scripts for lint configuration and repo checks |
| `.github/` | CI workflows — the checks described in [dev commands](dev-commands.md) |
