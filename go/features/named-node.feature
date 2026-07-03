Feature: Named node terms
  As a Go developer embedding Oxigraph
  I want to construct and inspect IRI terms as native Go values
  So that I can identify RDF resources without a running store

  Scenario: Creating a named node from an absolute IRI
    When the developer creates a named node from the IRI "http://example.com/person/alice"
    Then the named node's value is "http://example.com/person/alice"
    And the term's N-Quads form is "<http://example.com/person/alice>"

  Scenario Outline: An invalid IRI is rejected
    When the developer creates a named node from the IRI "<iri>"
    Then the construction fails with an invalid IRI error

    Examples:
      | iri                    |
      | person/alice           |
      | http://example.com/a b |
      |                        |
