# Add a conformance test for a spec-compliance fix

**Goal:** pin down a parsing or SPARQL bug fix with a manifest-based test, the
same format the W3C suites use.

The W3C suites live in read-only submodules, so new tests go into Oxigraph's
own suite under `testsuite/oxigraph-tests/`, which mirrors the W3C layout.
Each area (`sparql/`, `parser/`, `jsonld/`, `sparql-results/`, …) has a
`manifest.ttl` listing its tests.

## 1. Add the test files

For a SPARQL evaluation test you typically add a query and its expected
results, plus input data if needed:

```
testsuite/oxigraph-tests/sparql/my_fix.rq    # the query
testsuite/oxigraph-tests/sparql/my_fix.srx   # expected results (SPARQL XML)
testsuite/oxigraph-tests/sparql/my_fix.ttl   # input data, if any
```

## 2. Register it in the manifest

Add an entry to `testsuite/oxigraph-tests/sparql/manifest.ttl` — its id in the
`mf:entries` list and a description block like the existing ones:

```turtle
:my_fix rdf:type mf:QueryEvaluationTest ;
    mf:name "short description of what is being checked" ;
    mf:action [ qt:query <my_fix.rq> ; qt:data <my_fix.ttl> ] ;
    mf:result <my_fix.srx> .
```

Parser tests use the same pattern with the eval/positive/negative syntax test
types you'll find in the neighboring manifests — copy the closest existing
entry and adjust.

## 3. Run it

```sh
cargo run -p oxigraph-testsuite -- https://github.com/oxigraph/oxigraph/tests/sparql/manifest.ttl
```

Your test should appear in the report — failing before your fix, passing
after. The scoped `cargo test -p oxigraph-testsuite --test sparql` run picks it
up too, which is what CI executes.
