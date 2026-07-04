# Oxigraph FFI

An internal, deliberately coarse C ABI over the
[`oxigraph`](https://crates.io/crates/oxigraph) crate, built as a static
library for embedding Oxigraph in other languages. It exists for the
in-tree Go binding ([`go/`](../go)); see
[ADR 0001](../docs/decisions/0001-embed-oxigraph-in-go-via-c-abi-ffi.md)
for the design and its boundaries.

The surface is string-based and whole-operation-sized: no iterators, term
objects or transactions cross the boundary. Fallible functions report
failure through a `char **error_out` out-parameter; every pointer the
library hands out is caller-owned and released through exactly one
matching function (`oxigraph_close`, `oxigraph_free_string`). The C
declarations live in [`oxigraph_ffi.h`](./oxigraph_ffi.h).

Build it with:

```sh
cargo build -p oxigraph-ffi --release
```

This crate is not published; its only consumers are in this repository.

## License

Licensed under either of Apache License, Version 2.0 or MIT license, at
your option.
