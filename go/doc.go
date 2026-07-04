// Package oxigraph provides Go-native RDF terms for the Oxigraph graph
// database, mirroring pyoxigraph's model classes: NamedNode, BlankNode,
// Literal, Triple, Quad and the default graph.
//
// All term types are immutable comparable values: two terms are equal
// exactly when == reports true, matching pyoxigraph's equality semantics
// (lexical-form comparison; a typed literal with datatype xsd:string is
// the plain literal). Term positions are enforced at compile time through
// the sealed Term, Subject and GraphName interfaces.
//
// The embedded Store is backed by the oxigraph-ffi C ABI; per ADR 0001,
// terms never cross that boundary as structured data — they are built,
// validated and compared in pure Go, and whole operations cross as
// strings.
package oxigraph
