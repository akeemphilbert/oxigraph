# Feature flags of the `oxigraph` crate

Declared in [`lib/oxigraph/Cargo.toml`](../../../lib/oxigraph/Cargo.toml):

| Feature | Effect |
| ------- | ------ |
| `rocksdb` *(default)* | On-disk storage backend via `oxrocksdb-sys`, compiling the vendored RocksDB. Without it, only the in-memory store is available |
| `rocksdb-pkg-config` | Link a system-provided RocksDB found through pkg-config instead of building the vendored copy |
| `rocksdb-debug` | Development aid: extra checking in the RocksDB bindings |
| `js` | Adjustments for JavaScript/WebAssembly targets (e.g. randomness via the JS API) |
| `http-client` | HTTP client (oxhttp) so `SERVICE` clauses can call remote SPARQL endpoints |
| `http-client-native-tls` | `http-client` with the platform TLS stack |
| `http-client-rustls-webpki` | `http-client` with rustls and the webpki root certificates |
| `http-client-rustls-native` | `http-client` with rustls and the platform root certificates |
| `rdf-12` | Preliminary support for the RDF 1.2 / SPARQL 1.2 drafts |
| `geosparql` | GeoSPARQL functions in queries, via `spargeo` |

The CLI crate re-exposes several of these (see `cli/Cargo.toml`): its defaults
add `native-tls`, `geosparql`, and `rdf-12` on top of the library's.

Downstream crates each declare their own smaller feature sets — check the
`[features]` section of the crate you're working on.
