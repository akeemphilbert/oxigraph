# Work on the Python bindings (pyoxigraph)

**Goal:** build your local checkout as a Python package and run its test suite.

The bindings live in `python/`, written in Rust with
[PyO3](https://pyo3.rs/) — `python/src/` wraps the `oxigraph` crate, so most
functional changes happen in the Rust code, not in Python.

## Build and install the development version

From `python/`, into whatever virtualenv you use:

```sh
python -m venv .venv && source .venv/bin/activate   # if you need one
pip install .
```

`pip` drives the Rust build (via maturin), so this compiles the crate — the
first run takes a while, like any Oxigraph build.

## Run the tests

The tests are standard `unittest` cases in `python/tests/`:

```sh
cd tests && python -m unittest
```

CI runs them with [uv](https://docs.astral.sh/uv/) instead of a manual venv —
`uv run --locked -m unittest` from `python/tests` — plus `ruff` for
formatting/linting, type-stub checks (`generate_stubs.py`, `mypy.stubtest`),
and a Sphinx docs build from `python/docs`. If you touch the Python API
surface, remember the `.pyi` stubs are generated, not hand-edited.

## Iterate

After a Rust change, reinstall to rebuild the extension module:

```sh
pip install . && (cd tests && python -m unittest)
```

The Rust side of the bindings is linted like the rest of the workspace:
`cargo clippy --all-targets -- -D warnings -D clippy::all` from `python/`.
