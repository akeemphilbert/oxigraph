Feature: Adding and removing quads
  As a Go developer embedding Oxigraph
  I want to add quads to a store and remove them
  So that my program maintains graph data the way a pyoxigraph program does

  @wip
  Scenario: Adding a quad makes it visible to queries
    Given an open in-memory store
    When the developer adds the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  @wip
  Scenario: A query scoped to a named graph finds a quad added to that graph
    Given an open in-memory store
    And the store contains the quad:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |
    When the developer runs the query:
      """
      SELECT ?title WHERE {
        GRAPH <http://example.com/library> {
          <http://example.com/book/1> <http://purl.org/dc/terms/title> ?title
        }
      }
      """
    Then the query returns these solutions in order:
      | title                |
      | "Le Petit Prince"@fr |

  @wip
  Scenario: Adding the same quad twice stores it once
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer adds the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the store contains exactly 1 quad

  @wip
  Scenario: Removing a quad leaves the store's other quads intact
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    And the store contains the quad:
      | subject   | <http://example.com/book/2>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Watership Down"@en              |
    When the developer removes the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the store does not contain the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    And the store contains the quad:
      | subject   | <http://example.com/book/2>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Watership Down"@en              |

  @wip
  Scenario: Removing a quad the store does not contain reports no error
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer removes the quad:
      | subject   | <http://example.com/book/2>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Watership Down"@en              |
    Then the remove reports no error
    And the store contains exactly 1 quad

  @wip
  Scenario: A blank-node subject survives the add and remove round trip
    Given an open in-memory store
    When the developer adds the quad:
      | subject   | _:author1                        |
      | predicate | <http://xmlns.com/foaf/0.1/name> |
      | object    | "Antoine de Saint-Exupéry"       |
    And the developer removes the quad:
      | subject   | _:author1                        |
      | predicate | <http://xmlns.com/foaf/0.1/name> |
      | object    | "Antoine de Saint-Exupéry"       |
    Then the store is empty

  @wip
  Scenario: Data added to an on-disk store survives closing and reopening
    Given an open on-disk store at "book-catalog"
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer closes the store
    And the developer opens an on-disk store at "book-catalog"
    Then the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  @wip
  Scenario: A quad cannot be added to a closed store
    Given an in-memory store that has been closed
    When the developer adds the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the add fails with a closed store error

  @wip
  Scenario: A quad cannot be removed from a closed store
    Given an in-memory store that has been closed
    When the developer removes the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    Then the remove fails with a closed store error
