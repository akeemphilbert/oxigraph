# Run the benchmarks

**Goal:** measure whether your change helps or hurts performance, at two levels
of granularity.

## Microbenchmarks (Criterion)

The store and the parsers have [Criterion](https://github.com/bheisler/criterion.rs)
benchmarks:

```sh
cargo bench -p oxigraph              # store operations (lib/oxigraph/benches/store.rs)
cargo bench -p oxigraph-testsuite    # parsers (testsuite/benches/parser.rs)
```

Criterion prints timing changes against the previous run of the same bench, so
run once on `main`, switch to your branch, and run again to get a comparison.
Results (and HTML reports) land in `target/criterion/`.

## End-to-end benchmark (BSBM)

The `bench/` directory automates the
[Berlin SPARQL Benchmark](http://wifo5-03.informatik.uni-mannheim.de/bizer/berlinsparqlbenchmark/):
it generates a dataset, loads it into a server built from your working tree,
and fires the benchmark's query mixes at it. The `bsbm-tools` submodule must be
present (`git submodule update --init`).

```sh
cd bench
./bsbm_oxigraph.sh
```

Dataset size and other parameters are variables at the top of the script.
Matching scripts exist for other stores (`bsbm_jena.sh`, `bsbm_virtuoso.sh`, …)
if you want a cross-engine comparison, and `bench/README.md` shows previously
published results.

Expect a BSBM run to take a while and to be sensitive to whatever else your
machine is doing — prefer the Criterion benches for quick iteration and BSBM
for before/after evidence on substantial engine changes.
