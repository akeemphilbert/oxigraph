Feature: Triple terms
  As a Go developer embedding Oxigraph
  I want to assemble subject, predicate, and object into triples as native Go values
  So that I can describe statements independently of any graph

  @wip
  Scenario: A triple exposes its three components
    When the developer creates a triple with:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the triple's subject is the named node "http://example.com/book/1"
    And the triple's predicate is the named node "http://purl.org/dc/terms/title"
    And the triple's object is the language-tagged literal "Le Petit Prince" with the language "fr"

  @wip
  Scenario: A triple serializes with three terms
    When the developer creates a triple with:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the triple serializes to:
      """
      <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr
      """

  @wip
  Scenario: Triples with the same components are equal
    Given the triple:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    And the triple:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer compares the two triples
    Then the triples are equal

  @wip
  Scenario: Triples with different objects are not equal
    Given the triple:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    And the triple:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "The Little Prince"@en           |
    When the developer compares the two triples
    Then the triples are not equal
