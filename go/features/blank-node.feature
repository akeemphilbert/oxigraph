Feature: Blank node terms
  As a Go developer embedding Oxigraph
  I want to construct anonymous nodes as native Go values
  So that I can model resources that have no IRI

  @wip
  Scenario: Creating a blank node with a chosen identifier
    When the developer creates a blank node with the identifier "author1"
    Then the blank node's identifier is "author1"
    And the term's N-Quads form is "_:author1"

  @wip
  Scenario: Blank nodes created without an identifier are distinct
    When the developer creates two blank nodes without identifiers
    Then each blank node has a non-empty identifier
    And the two blank nodes are not equal

  @wip
  Scenario Outline: An invalid blank node identifier is rejected
    When the developer creates a blank node with the identifier "<identifier>"
    Then the construction fails with an invalid blank node identifier error

    Examples:
      | identifier |
      |            |
      | invoice 42 |
      | -order     |
