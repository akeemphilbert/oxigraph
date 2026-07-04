Feature: Loading and dumping RDF documents
  As a Go developer embedding Oxigraph
  I want to load RDF documents into a store and dump the store back out
  So that my program exchanges whole datasets with other tools the way a pyoxigraph program does

  @wip
  Scenario: Loading a Turtle document makes its triples visible to queries
    Given an open in-memory store
    When the developer loads the Turtle document:
      """
      @prefix dcterms: <http://purl.org/dc/terms/> .
      <http://example.com/book/1> dcterms:title "Le Petit Prince"@fr ;
          dcterms:publisher "Gallimard" .
      """
    Then the store contains exactly 2 quads
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  @wip
  Scenario: Loading an N-Triples document places its triples in the default graph
    Given an open in-memory store
    When the developer loads the N-Triples document:
      """
      <http://example.com/book/2> <http://purl.org/dc/terms/title> "Watership Down"@en .
      """
    Then the store contains the quad:
      | subject   | <http://example.com/book/2>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Watership Down"@en              |

  @wip
  Scenario: Loading an N-Quads document places each quad in its named graph
    Given an open in-memory store
    When the developer loads the N-Quads document:
      """
      <http://example.com/book/3> <http://purl.org/dc/terms/title> "Vol de Nuit" <http://example.com/library> .
      """
    Then the store contains the quad:
      | subject    | <http://example.com/book/3>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Vol de Nuit"                    |
      | graph-name | <http://example.com/library>     |

  @wip
  Scenario: Loading a TriG document places each quad in its named graph
    Given an open in-memory store
    When the developer loads the TriG document:
      """
      @prefix dcterms: <http://purl.org/dc/terms/> .
      GRAPH <http://example.com/library> {
        <http://example.com/book/1> dcterms:title "Le Petit Prince"@fr .
      }
      """
    Then the store contains the quad:
      | subject    | <http://example.com/book/1>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Le Petit Prince"@fr             |
      | graph-name | <http://example.com/library>     |

  @wip
  Scenario: Loading a document keeps the quads already in the store
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer loads the N-Triples document:
      """
      <http://example.com/book/2> <http://purl.org/dc/terms/title> "Watership Down"@en .
      """
    Then the store contains exactly 2 quads
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |

  @wip
  Scenario: Loading the same document twice stores each quad once
    Given an open in-memory store
    When the developer loads the N-Triples document:
      """
      <http://example.com/book/2> <http://purl.org/dc/terms/title> "Watership Down"@en .
      """
    And the developer loads the N-Triples document:
      """
      <http://example.com/book/2> <http://purl.org/dc/terms/title> "Watership Down"@en .
      """
    Then the store contains exactly 1 quad

  @wip
  Scenario: Loading an empty document reports no error and adds nothing
    Given an open in-memory store
    When the developer loads the Turtle document:
      """
      """
    Then the load reports no error
    And the store is empty

  @wip
  Scenario: A malformed Turtle document is rejected with its line number
    Given an open in-memory store
    When the developer loads the Turtle document:
      """
      <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr .
      <http://example.com/book/2> <http://purl.org/dc/terms/title> .
      """
    Then the load fails with a syntax error
    And the error message mentions "line 2"
    And the store is empty

  @wip
  Scenario: A document cannot be loaded into a closed store
    Given an in-memory store that has been closed
    When the developer loads the Turtle document:
      """
      <http://example.com/book/1> <http://purl.org/dc/terms/title> "Le Petit Prince"@fr .
      """
    Then the load fails with a closed store error

  @wip
  Scenario: Loading with a format the library does not define is rejected
    Given an open in-memory store
    When the developer loads a document using an undefined format
    Then the load fails with an unsupported format error

  @wip
  Scenario Outline: A dataset-format dump captures quads from named graphs
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    And the store contains the quad:
      | subject    | <http://example.com/book/2>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Watership Down"@en              |
      | graph-name | <http://example.com/library>     |
    When the developer dumps the store as <format>
    And the developer loads the dump into a second in-memory store
    Then the second store contains the quad:
      | subject    | <http://example.com/book/2>      |
      | predicate  | <http://purl.org/dc/terms/title> |
      | object     | "Watership Down"@en              |
      | graph-name | <http://example.com/library>     |
    And the second store contains exactly 2 quads

    Examples:
      | format  |
      | N-Quads |
      | TriG    |

  @wip
  Scenario: A load, dump, and reload round trip preserves the quad count
    Given an open in-memory store
    When the developer loads the TriG document:
      """
      @prefix dcterms: <http://purl.org/dc/terms/> .
      <http://example.com/book/1> dcterms:title "Le Petit Prince"@fr .
      GRAPH <http://example.com/library> {
        <http://example.com/book/2> dcterms:title "Watership Down"@en .
      }
      """
    And the developer dumps the store as N-Quads
    And the developer loads the dump into a second in-memory store
    Then the second store contains exactly 2 quads

  @wip
  Scenario: Dumping an empty store produces a document with no quads
    Given an open in-memory store
    When the developer dumps the store as N-Quads
    And the developer loads the dump into a second in-memory store
    Then the second store is empty

  @wip
  Scenario Outline: Dumping the whole store to a triples-only format is rejected
    Given an open in-memory store
    And the store contains the quad:
      | subject   | <http://example.com/book/1>      |
      | predicate | <http://purl.org/dc/terms/title> |
      | object    | "Le Petit Prince"@fr             |
    When the developer dumps the store as <format>
    Then the dump fails with an unsupported format error
    And the error message mentions "<format>"

    Examples:
      | format    |
      | Turtle    |
      | N-Triples |

  @wip
  Scenario: A closed store cannot be dumped
    Given an in-memory store that has been closed
    When the developer dumps the store as N-Quads
    Then the dump fails with a closed store error

  @wip
  Scenario: Dumping with a format the library does not define is rejected
    Given an open in-memory store
    When the developer dumps the store using an undefined format
    Then the dump fails with an unsupported format error
