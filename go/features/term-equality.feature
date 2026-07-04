Feature: Term equality
  As a Go developer embedding Oxigraph
  I want terms to compare exactly as pyoxigraph terms do
  So that deduplication and assertions behave identically across bindings

  Scenario: Named nodes with the same IRI are equal
    Given the named node "http://example.com/person/alice"
    And the named node "http://example.com/person/alice"
    When the developer compares the two terms
    Then the terms are equal

  Scenario: Named nodes with different IRIs are not equal
    Given the named node "http://example.com/person/alice"
    And the named node "http://example.com/person/bob"
    When the developer compares the two terms
    Then the terms are not equal

  Scenario: Blank nodes with the same identifier are equal
    Given the blank node with the identifier "author1"
    And the blank node with the identifier "author1"
    When the developer compares the two terms
    Then the terms are equal

  Scenario: A typed literal with the xsd:string datatype equals the plain literal
    Given the literal "Oxigraph"
    And the typed literal "Oxigraph" with the datatype "http://www.w3.org/2001/XMLSchema#string"
    When the developer compares the two terms
    Then the terms are equal

  Scenario: A language-tagged literal does not equal the plain literal with the same value
    Given the literal "chat"
    And the language-tagged literal "chat" with the language "fr"
    When the developer compares the two terms
    Then the terms are not equal

  Scenario: Typed literals compare by lexical form, not by value
    Given the typed literal "42" with the datatype "http://www.w3.org/2001/XMLSchema#integer"
    And the typed literal "042" with the datatype "http://www.w3.org/2001/XMLSchema#integer"
    When the developer compares the two terms
    Then the terms are not equal

  Scenario: Two default graph values are equal
    Given a default graph value
    And a default graph value
    When the developer compares the two terms
    Then the terms are equal
