package config

import (
	"ngstaticserver/test"
	"path/filepath"
	"testing"
)

func TestShouldParseDotEnv(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(".env", "ENV =production\nPORT =8080 \nDELAY = 200")

	var result map[string]*string
	CreateDotEnv(context.Path, func(variables map[string]*string) {
		result = variables
	})

	test.AssertEqual(t, len(result), 3)
	test.AssertEqual(t, readValue(t, result, "ENV"), "production")
	test.AssertEqual(t, readValue(t, result, "PORT"), "8080")
	test.AssertEqual(t, readValue(t, result, "DELAY"), "200")
}

func TestShouldParseEmptyDotEnv(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(".env", "")

	var result map[string]*string
	CreateDotEnv(context.Path, func(variables map[string]*string) {
		result = variables
	})

	test.AssertEqual(t, len(result), 0)
}

func TestShouldSkipMissingDotEnv(t *testing.T) {
	context := test.NewTestDir(t)

	var result map[string]*string
	CreateDotEnv(filepath.Join(context.Path, "missing"), func(variables map[string]*string) {
		result = variables
	})

	test.AssertEqual(t, len(result), 0)
}

func TestShouldSkipMalformedDotEnv(t *testing.T) {
	context := test.NewTestDir(t)
	context.WriteFile(".env", "{}")

	var result map[string]*string
	CreateDotEnv(context.Path, func(variables map[string]*string) {
		result = variables
	})

	test.AssertEqual(t, len(result), 0)
}

func readValue(t *testing.T, variables map[string]*string, key string) string {
	value, ok := variables[key]
	if !ok {
		t.Fatalf("Variable %v not defined", key)
		return ""
	}

	return *value
}
