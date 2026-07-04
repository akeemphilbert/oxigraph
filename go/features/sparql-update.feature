Feature: SPARQL update execution
  As a Go developer embedding Oxigraph
  I want to run SPARQL updates against a store
  So that my program rewrites graph data the way a pyoxigraph program does

  Scenario: An INSERT DATA update makes its data visible to queries
    Given an open in-memory store
    When the developer runs the update:
      """
      INSERT DATA {
        <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr
      }
      """
    Then the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  Scenario: A DELETE DATA update removes the data it names
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer runs the update:
      """
      DELETE DATA {
        <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr
      }
      """
    Then the store is empty

  Scenario: A DELETE DATA update may not reference a blank node
    Given an open in-memory store
    When the developer runs the update:
      """
      DELETE DATA {
        _:author1 <http://xmlns.com/foaf/0.1/name> "Antoine de Saint-Exupéry"
      }
      """
    Then the update fails with a syntax error

  Scenario: A rejected update leaves the store unchanged
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer runs the update:
      """
      INSERT DATA {
        <http://example.com/book/2> <http://purl.org/dc/terms/title> "Watership Down"@en
      """
    Then the update fails with a syntax error
    And the store contains exactly 1 quad
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  Scenario Outline: A malformed update is rejected
    Given an open in-memory store
    When the developer runs the update "<update>"
    Then the update fails with a syntax error

    Examples:
      | update                                                                                    |
      | INSRT DATA { <http://example.com/book/1> <http://purl.org/dc/terms/title> "Vol de Nuit" } |
      | INSERT DATA { <http://example.com/book/1> <http://purl.org/dc/terms/title> "Vol de Nuit"  |
      | SELECT ?title WHERE { ?book ?p ?title }                                                   |

  Scenario: A closed store cannot be updated
    Given an in-memory store that has been closed
    When the developer runs the update "CLEAR ALL"
    Then the update fails with a closed store error
