Feature: N-Quads and N-Triples line round-trips
  As a Go developer embedding Oxigraph
  I want N-Quads and N-Triples lines to round-trip through native Go quads
  So that later load and dump operations preserve data exactly

  @wip
  Scenario Outline: A canonical line round-trips unchanged
    When the developer parses the N-Quads line "<line>"
    Then serializing the quad as an N-Quads line reproduces the original line

    Examples:
      | line                                                                                                          |
      | <http://example.com/book/1> <http://purl.org/dc/terms/publisher> "Gallimard" .                                |
      | <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr <http://example.com/library> . |
      | <http://example.com/book/1> <http://example.com/pageCount> "96"^^<http://www.w3.org/2001/XMLSchema#integer> . |
      | _:author1 <http://xmlns.com/foaf/0.1/name> "Antoine de Saint-Exupéry" .                                       |

  @wip
  Scenario: An N-Triples line parses into the default graph
    When the developer parses the N-Quads line "<http://example.com/book/1> <http://purl.org/dc/terms/publisher> "Gallimard" ."
    Then the quad's graph name is the default graph
    And the quad's subject is the named node "http://example.com/book/1"

  @wip
  Scenario: Escape sequences decode when a line is parsed
    When the developer parses the N-Quads line:
      """
      <http://example.com/note/1> <http://example.com/text> "He said \"bonjour\"\nand left" .
      """
    Then the quad's object is a literal with the value:
      """
      He said "bonjour"
      and left
      """

  @wip
  Scenario Outline: A malformed line is rejected
    When the developer parses the N-Quads line "<line>"
    Then the parsing fails with a syntax error

    Examples:
      | line                                                                   |
      | <http://example.com/book/1> <http://purl.org/dc/terms/title> .        |
      | <http://example.com/book/1> "Gallimard" <http://example.com/book/2> . |
