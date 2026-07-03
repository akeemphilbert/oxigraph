# Use Oxigraph from JavaScript

**Goal:** run an RDF store with SPARQL support inside Node.js or the browser
through the `oxigraph` npm package (Oxigraph compiled to WebAssembly).

Install it:

```sh
npm install oxigraph
```

Insert a triple and query it back (Node.js):

```js
const oxigraph = require("oxigraph");

const store = new oxigraph.Store();

store.add(oxigraph.triple(
    oxigraph.namedNode("http://example.com/oxigraph"),
    oxigraph.namedNode("http://www.w3.org/2000/01/rdf-schema#label"),
    oxigraph.literal("Oxigraph")
));

for (const binding of store.query("SELECT ?name WHERE { ?s ?p ?name }")) {
    console.log(binding.get("name").value);
}
```

(This is the API of the current 0.5 releases, which this snippet was verified
against. The upcoming 0.6 removes the built-in term constructors like
`oxigraph.namedNode` in favor of RDF/JS data-model libraries — see the
migration guide in the package README before upgrading.)

The store lives in memory: the WebAssembly build has no on-disk persistence. If
you need durable storage, run [the server](../tutorials/getting-started.md) and
query it [over HTTP](http.md) instead.

The same package works in the browser with any bundler that supports WebAssembly
modules. See [the `oxigraph` package on npm](https://www.npmjs.com/package/oxigraph)
and [the JS binding README](../../../js/README.md) for the full API, including
RDF/JS-style term construction and dataset loading.
