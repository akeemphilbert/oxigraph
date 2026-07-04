Feature: SPARQL query execution
  As a Go developer embedding Oxigraph
  I want to run SPARQL queries against a store and read typed results
  So that my program answers graph questions the way a pyoxigraph program does

  Scenario: A SELECT query returns ordered, typed solutions
    Given an open in-memory store
    When the developer runs the query:
      """
      SELECT ?book ?title ?pages WHERE {
        VALUES (?book ?title ?pages) {
          (<http://example.com/book/1> "Le Petit Prince"@fr 96)
          (<http://example.com/book/2> "Watership Down"@en 476)
        }
      }
      ORDER BY DESC(?pages)
      """
    Then the query returns these solutions in order:
      | book                        | title                | pages                                             |
      | <http://example.com/book/2> | "Watership Down"@en  | "476"^^<http://www.w3.org/2001/XMLSchema#integer> |
      | <http://example.com/book/1> | "Le Petit Prince"@fr | "96"^^<http://www.w3.org/2001/XMLSchema#integer>  |

  Scenario: A SELECT query with no matching data returns zero solutions
    Given an open in-memory store
    When the developer runs the query:
      """
      SELECT ?title WHERE { ?book <http://purl.org/dc/terms/title> ?title }
      """
    Then the query returns no solutions

  Scenario: A solution distinguishes an unbound variable from a bound one
    Given an open in-memory store
    When the developer runs the query:
      """
      SELECT ?name ?birthYear WHERE {
        VALUES (?name ?birthYear) { ("Homer" UNDEF) }
      }
      """
    Then the query returns 1 solution
    And the solution binds "name" to the literal "Homer"
    But the solution does not bind "birthYear"

  Scenario: An ASK query answers true when its condition holds
    Given an open in-memory store
    When the developer runs the query "ASK { FILTER(2 + 2 = 4) }"
    Then the query answers true

  Scenario: An ASK query answers false when the store has no matching data
    Given an open in-memory store
    When the developer runs the query "ASK { ?book <http://purl.org/dc/terms/title> ?title }"
    Then the query answers false

  Scenario: A CONSTRUCT query returns the triples it builds
    Given an open in-memory store
    When the developer runs the query:
      """
      CONSTRUCT { ?book <http://purl.org/dc/terms/title> ?title }
      WHERE {
        VALUES (?book ?title) {
          (<http://example.com/book/1> "Le Petit Prince"@fr)
          (<http://example.com/book/2> "Watership Down"@en)
        }
      }
      """
    Then the query returns exactly these triples:
      | subject                     | predicate                        | object               |
      | <http://example.com/book/1> | <http://purl.org/dc/terms/title> | "Le Petit Prince"@fr |
      | <http://example.com/book/2> | <http://purl.org/dc/terms/title> | "Watership Down"@en  |

  Scenario: A DESCRIBE query for a resource with no data returns no triples
    Given an open in-memory store
    When the developer runs the query "DESCRIBE <http://example.com/book/1>"
    Then the query returns no triples

  Scenario Outline: A malformed query is rejected
    Given an open in-memory store
    When the developer runs the query "<query>"
    Then the query fails with a syntax error

    Examples:
      | query                                  |
      | SELCT ?title WHERE { ?book ?p ?title } |
      | SELECT ?title WHERE { ?book ?p ?title  |
      |                                        |

  Scenario: A query calling an unknown function fails with an evaluation error
    Given an open in-memory store
    When the developer runs the query:
      """
      SELECT ?age WHERE {
        VALUES ?age { 44 }
        FILTER(<http://example.com/functions/isPrime>(?age))
      }
      """
    Then the query fails with an evaluation error
    And the error message mentions "http://example.com/functions/isPrime"

  Scenario: A closed store cannot be queried
    Given an in-memory store that has been closed
    When the developer runs the query "SELECT ?title WHERE { ?book ?p ?title }"
    Then the query fails with a closed store error
