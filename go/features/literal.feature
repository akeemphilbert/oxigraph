Feature: Literal terms
  As a Go developer embedding Oxigraph
  I want to construct plain, language-tagged, and typed literals as native Go values
  So that I can represent RDF values exactly as pyoxigraph does

  @wip
  Scenario: Creating a plain literal
    When the developer creates the literal "Oxigraph"
    Then the literal's value is "Oxigraph"
    And the literal has no language
    And the literal's datatype is "http://www.w3.org/2001/XMLSchema#string"
    And the term's N-Quads form is:
      """
      "Oxigraph"
      """

  @wip
  Scenario: Creating a language-tagged literal
    When the developer creates the language-tagged literal "graph database" with the language "en"
    Then the literal's value is "graph database"
    And the literal's language is "en"
    And the literal's datatype is "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"
    And the term's N-Quads form is:
      """
      "graph database"@en
      """

  @wip
  Scenario: Language tags are normalized to lowercase
    When the developer creates the language-tagged literal "Bonjour" with the language "fr-CA"
    Then the literal's language is "fr-ca"
    And the term's N-Quads form is:
      """
      "Bonjour"@fr-ca
      """

  @wip
  Scenario Outline: An invalid language tag is rejected
    When the developer creates the language-tagged literal "hello" with the language "<language>"
    Then the construction fails with an invalid language tag error

    Examples:
      | language |
      | en_US    |
      | 123      |
      |          |

  @wip
  Scenario: Creating a typed literal
    When the developer creates the typed literal "96" with the datatype "http://www.w3.org/2001/XMLSchema#integer"
    Then the literal's value is "96"
    And the literal has no language
    And the literal's datatype is "http://www.w3.org/2001/XMLSchema#integer"
    And the term's N-Quads form is:
      """
      "96"^^<http://www.w3.org/2001/XMLSchema#integer>
      """

  @wip
  Scenario: Double quotes and newlines in a literal value are escaped in N-Quads
    When the developer creates a literal with the value:
      """
      He said "bonjour"
      and left
      """
    Then the term's N-Quads form is:
      """
      "He said \"bonjour\"\nand left"
      """

  @wip
  Scenario: Backslashes in a literal value are escaped in N-Quads
    When the developer creates the literal "C:\graphs\oxigraph"
    Then the term's N-Quads form is:
      """
      "C:\\graphs\\oxigraph"
      """
