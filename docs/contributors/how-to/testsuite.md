# Run and debug the conformance testsuite

**Goal:** run the W3C conformance tests (plus Oxigraph's own regression tests)
against your working tree, scope the run to the spec area you're touching, and
dig into a single failing test.

The `testsuite/` crate drives manifest-based test suites: the official W3C
suites (vendored as git submodules — `rdf-tests`, `json-ld-api`, `rdf-canon`,
`N3`, …) and Oxigraph's own tests under `testsuite/oxigraph-tests/`. If those
directories are empty, fetch the submodules first: `git submodule update --init`.

## Run everything

```sh
cargo test -p oxigraph-testsuite
```

## Scope the run to one spec area

The suites are grouped into integration-test files — `sparql`, `parser`,
`canonicalization`, `oxigraph` (Oxigraph's regression tests), and `serd`:

```sh
cargo test -p oxigraph-testsuite --test sparql
cargo test -p oxigraph-testsuite --test parser
```

Within a file, filter by test-function name (they are named after the suite):

```sh
cargo test -p oxigraph-testsuite --test sparql sparql10_w3c_query_evaluation
```

## Debug a single failing test

The testsuite also builds as a binary that runs any manifest directly and
prints a per-test report — handy when one conformance test fails and you want
to see which:

```sh
cargo run -p oxigraph-testsuite -- https://github.com/oxigraph/oxigraph/tests/sparql/manifest.ttl
```

Manifest URLs are mapped to local files (see `testsuite/src/files.rs`):
`https://github.com/oxigraph/oxigraph/tests/…` resolves to
`testsuite/oxigraph-tests/…`, and `https://w3c.github.io/…` resolves to the
matching W3C submodule. So you can run exactly the manifest — W3C or local —
that contains the failing test, read which assertion failed, and open the
test's `.rq`/`.ttl`/`.srx` files referenced by the manifest to reproduce it.

## Related

- Adding a test for a fix: [add a conformance test](add-conformance-test.md).
- The full workspace test run is covered in
  [the build-run-test tutorial](../tutorials/build-run-test.md).
