# One engine, many interfaces

Oxigraph is not four products — a server, a Rust crate, a Python package, and a
JavaScript package. It is one database engine, written in Rust, that ships in
four forms. Understanding that relationship explains most of the behavior you
will notice as a user: why the SPARQL dialect is identical everywhere, why the
packages version together, and why the JavaScript package alone has no on-disk
storage.

## The core

The engine — RDF data model, the format parsers, the SPARQL parser, optimizer,
and evaluator, and the storage layer — lives in
[the `oxigraph` Rust crate](https://docs.rs/oxigraph). Everything you can
install is a wrapper around that crate:

- **The server and CLI** are a small Rust program (an HTTP front end and
  command-line tools) linked against the crate.
- **`pyoxigraph`** is the crate compiled into a native Python extension module.
  Your Python process calls straight into the engine — there is no server, no
  socket, no serialization boundary.
- **The `oxigraph` npm package** is the crate compiled to WebAssembly, running
  inside your JavaScript runtime.
- **Rust applications** just use the crate directly.

Because every interface runs the same parser, optimizer, and evaluator, a query
behaves the same everywhere: same SPARQL conformance, same supported RDF
formats, same results — and the same release notes apply to all of them, which
is why the packages share version numbers.

## Where your data lives

The engine has two storage backends, and which one you get is decided by the
platform, not by the query language:

- **On disk (RocksDB).** The server, the CLI, pyoxigraph, and the Rust crate
  store quads in a [RocksDB](https://rocksdb.org/) key-value store — a
  log-structured merge tree design that trades a little read work for fast
  writes. Oxigraph keeps the same quads under several sort orders so that any
  triple pattern can be answered by a range scan. Data written by one of these
  interfaces can be opened by another, because it is literally the same format.
- **In memory.** Every interface can also run purely in memory. For the
  JavaScript package it is the *only* option: WebAssembly has no filesystem, so
  the wasm build simply leaves RocksDB out. If you need persistence from
  JavaScript, run [the server](../tutorials/getting-started.md) and query it
  [over HTTP](../how-to/http.md).

## Queries are evaluated lazily

The evaluator follows the classic Volcano iterator model: results are pulled
one solution at a time as you iterate, streaming matching quads out of storage
on demand. A query that you stop consuming early does not pay for the results
you never read.

## Going deeper

- [The upstream wiki's architecture page](https://github.com/oxigraph/oxigraph/wiki/Architecture)
  covers the storage encoding and query evaluation in depth.
- If you are curious how the wrappers are built — or want a binding for another
  language — see the contributor explanation of
  [how the language bindings work](../../contributors/explanation/language-bindings.md).
