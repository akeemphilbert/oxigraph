# Query and update a running server over HTTP

**Goal:** talk to an Oxigraph server from any language, using nothing but HTTP.

These recipes assume a server started as in
[the getting-started tutorial](../tutorials/getting-started.md):

```sh
oxigraph serve --location ./data
```

Oxigraph implements the standard
[SPARQL Protocol](https://www.w3.org/TR/sparql11-protocol/) and
[Graph Store Protocol](https://www.w3.org/TR/sparql11-http-rdf-update/), so any
SPARQL client library works. With plain `curl`:

## Run a query

```sh
curl -X POST http://localhost:7878/query \
  -H 'Content-Type: application/sparql-query' \
  --data 'SELECT * WHERE { ?s ?p ?o } LIMIT 10'
```

Ask for a specific result format with an `Accept` header, e.g.
`-H 'Accept: application/sparql-results+json'`.

## Run an update

```sh
curl -X POST http://localhost:7878/update \
  -H 'Content-Type: application/sparql-update' \
  --data 'INSERT DATA { <http://example.com/s> <http://example.com/p> "o" }'
```

## Load an RDF file

Upload into the default graph through the Graph Store Protocol:

```sh
curl -X POST 'http://localhost:7878/store?default' \
  -H 'Content-Type: text/turtle' \
  --data-binary @example.ttl
```

Replace `default` with `graph=http://example.com/g` to target a named graph.

## Read data back out

```sh
curl 'http://localhost:7878/store?default' -H 'Accept: text/turtle'
```

For all endpoints and options (authentication, read-only mode, CORS, …) see the
[CLI README](../../../cli/README.md).
