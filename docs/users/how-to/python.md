# Use Oxigraph from Python

**Goal:** work with RDF data and SPARQL from Python through the `pyoxigraph`
package.

Install it (wheels are published for the common platforms, so there is nothing to
compile):

```sh
pip install pyoxigraph
```

Insert a triple and query it back:

```python
from pyoxigraph import Store, NamedNode, Literal, Quad

store = Store()  # in-memory; Store("path/to/data") persists to disk

store.add(Quad(
    NamedNode("http://example.com/oxigraph"),
    NamedNode("http://www.w3.org/2000/01/rdf-schema#label"),
    Literal("Oxigraph"),
))

for solution in store.query("SELECT ?name WHERE { ?s ?p ?name }"):
    print(solution["name"].value)
```

`pyoxigraph` also parses and serializes RDF files (Turtle, N-Triples, RDF/XML,
JSON-LD, …) and can load them straight into a store:

```python
store.load(path="example.ttl")
```

For the full API see
[the pyoxigraph documentation](https://pyoxigraph.readthedocs.io/) and
[the package on PyPI](https://pypi.org/project/pyoxigraph/).
