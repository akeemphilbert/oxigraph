# Work on the JavaScript bindings

**Goal:** build your local checkout as the `oxigraph` npm package (WebAssembly)
and run its test suite.

The bindings live in `js/`, written in Rust with
[wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/) — `js/src/` wraps
the `oxigraph` crate for the WebAssembly target.

## Set up and test

You need the WebAssembly target once per toolchain:

```sh
rustup target add wasm32-unknown-unknown
```

Then:

```sh
cd js
npm install
npm test
```

`npm test` first compiles a debug WebAssembly build (`npm run build-debug`,
which drives `build_package.py`) and then runs the JavaScript tests with
vitest, so it is the one command that proves a Rust change works end to end.

The build script invokes `python`, so make sure a Python 3 answers to that name
on your `PATH` (on macOS, where only `python3` exists by default, an alias or
shim is enough).

## Other useful scripts

```sh
npm run build   # release build of the npm package into pkg/
npm run fmt     # format + lint the JS code with Biome
npm run bench   # benchmarks
```

The Rust side follows the workspace conventions (`cargo fmt`, `cargo clippy`).
Note that the WebAssembly build has no RocksDB — the JS package is in-memory
only, which is why storage changes rarely affect it but API changes always do.
