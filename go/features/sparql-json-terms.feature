Feature: SPARQL JSON result term parsing
  As a Go developer embedding Oxigraph
  I want SPARQL JSON result terms to parse into native Go term values
  So that later query results surface as typed terms instead of raw JSON

  Scenario: A uri term parses into a named node
    When the developer parses the SPARQL JSON term:
      """
      {"type": "uri", "value": "http://example.com/person/alice"}
      """
    Then the parsed term equals the named node "http://example.com/person/alice"

  Scenario: A bnode term parses into a blank node
    When the developer parses the SPARQL JSON term:
      """
      {"type": "bnode", "value": "b0"}
      """
    Then the parsed term equals the blank node with the identifier "b0"

  Scenario: A plain literal term parses into a literal
    When the developer parses the SPARQL JSON term:
      """
      {"type": "literal", "value": "Alice"}
      """
    Then the parsed term equals the literal "Alice"

  Scenario: A language-tagged literal term parses with its language
    When the developer parses the SPARQL JSON term:
      """
      {"type": "literal", "value": "Alicia", "xml:lang": "es"}
      """
    Then the parsed term equals the language-tagged literal "Alicia" with the language "es"

  Scenario: A typed literal term parses with its datatype
    When the developer parses the SPARQL JSON term:
      """
      {"type": "literal", "value": "42", "datatype": "http://www.w3.org/2001/XMLSchema#integer"}
      """
    Then the parsed term equals the typed literal "42" with the datatype "http://www.w3.org/2001/XMLSchema#integer"

  Scenario: A uri term with an invalid IRI is rejected
    When the developer parses the SPARQL JSON term:
      """
      {"type": "uri", "value": "person/alice"}
      """
    Then the parsing fails with an invalid IRI error

  Scenario: A term with an unrecognized type is rejected
    When the developer parses the SPARQL JSON term:
      """
      {"type": "graph", "value": "http://example.com/library"}
      """
    Then the parsing fails with an unsupported term type error

  Scenario: A term missing its value is rejected
    When the developer parses the SPARQL JSON term:
      """
      {"type": "uri"}
      """
    Then the parsing fails with a malformed term error
