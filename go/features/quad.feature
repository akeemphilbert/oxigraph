Feature: Quads and the default graph
  As a Go developer embedding Oxigraph
  I want to assemble terms into quads as native Go values
  So that I can describe statements destined for a store

  @wip
  Scenario: A quad exposes its four components
    When the developer creates a quad with:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    Then the quad's subject is the named node "http://example.com/book/1"
    And the quad's predicate is the named node "http://purl.org/dc/terms/title"
    And the quad's object is the language-tagged literal "Le Petit Prince" with the language "fr"
    And the quad's graph name is the named node "http://example.com/library"

  @wip
  Scenario: A quad exposes its underlying triple
    When the developer creates a quad with:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    Then the quad's triple equals the triple:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  @wip
  Scenario: A quad in a named graph serializes with four terms
    When the developer creates a quad with:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    Then the quad serializes to:
      """
      <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr <http://example.com/library>
      """

  @wip
  Scenario: A quad without a graph name belongs to the default graph
    When the developer creates a quad with:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the quad's graph name is the default graph
    And the quad serializes to:
      """
      <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr
      """

  @wip
  Scenario: Quads with the same components are equal
    Given the quad:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    And the quad:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    When the developer compares the two quads
    Then the quads are equal

  @wip
  Scenario: Quads that differ only by graph name are not equal
    Given the quad:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    And the quad:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/archive>     |
    When the developer compares the two quads
    Then the quads are not equal

  @wip
  Scenario: The default graph has a stable string form
    When the developer creates a default graph value
    Then the default graph's string form is "DEFAULT"
