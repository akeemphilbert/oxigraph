//go:build (darwin || linux) && (amd64 || arm64)

// Package features_test runs the Gherkin acceptance contract in this
// directory against the public oxigraph API with godog. It is an external
// test package on purpose: every step definition can only reach exported
// identifiers, so the suite doubles as a check that the contract is
// satisfiable through the public surface alone.
package features_test

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned from the godog test suite")
	}
}
