# Run a fuzz target

**Goal:** throw generated inputs at a parser or the SPARQL engine to shake out
panics and crashes, especially after changing parsing code.

Fuzzing uses [cargo-fuzz](https://rust-fuzz.github.io/book/cargo-fuzz.html),
which needs a nightly toolchain:

```sh
rustup toolchain install nightly
cargo install cargo-fuzz
```

List the available targets (one per parser/serializer plus SPARQL query and
update evaluation):

```sh
cargo fuzz list
```

You'll see targets like `trig`, `nquads`, `n3`, `rdf_xml`, `jsonld`,
`sparql_query`, `sparql_update`, `sparql_query_eval`, `sparql_results_json`, …
(they live in `fuzz/fuzz_targets/`).

Run one:

```sh
cargo +nightly fuzz run sparql_query
```

It runs until you stop it (Ctrl-C) or it finds a crash. Crashing inputs are
saved under `fuzz/artifacts/<target>/`; reproduce one deterministically with:

```sh
cargo +nightly fuzz run sparql_query fuzz/artifacts/sparql_query/<crash-file>
```

`fuzz/build_corpus.py` can seed the corpus from the conformance-test inputs to
give the fuzzer a head start.

If you fix a crash the fuzzer found, add the crashing input as a
[conformance test](add-conformance-test.md) so it never comes back.
