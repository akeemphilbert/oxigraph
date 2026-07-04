Feature: Store lifecycle
  As a Go developer embedding Oxigraph
  I want to open and close on-disk and in-memory stores
  So that graph storage starts and stops safely inside my program's process

  @wip
  Scenario: Opening an on-disk store creates its data directory
    Given no directory exists at "book-catalog"
    When the developer opens an on-disk store at "book-catalog"
    Then the store is open
    And a directory exists at "book-catalog"

  @wip
  Scenario: Opening an in-memory store needs no data directory
    When the developer opens an in-memory store
    Then the store is open
    And no data directory is created

  @wip
  Scenario: Closing an on-disk store releases it for reopening
    Given an open on-disk store at "book-catalog"
    When the developer closes the store
    And the developer opens an on-disk store at "book-catalog"
    Then the store is open

  @wip
  Scenario: A directory in use by an open store cannot be opened again
    Given an open on-disk store at "book-catalog"
    When the developer opens a second on-disk store at "book-catalog"
    Then the open fails with a storage error

  @wip
  Scenario: Closing one store leaves other stores open
    Given an open on-disk store at "book-catalog"
    And an open on-disk store at "music-catalog"
    When the developer closes the store at "book-catalog"
    Then the store at "music-catalog" is still open

  @wip
  Scenario: A closed store cannot be closed again
    Given an in-memory store that has been closed
    When the developer closes the store again
    Then the close fails with a closed store error

  @wip
  Scenario: Opening a store under a missing parent directory fails
    Given no directory exists at "archive"
    When the developer opens an on-disk store at "archive/book-catalog"
    Then the open fails with a storage error
