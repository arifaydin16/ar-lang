package ast

import (
	"first-api/arlexer"
	"first-api/utils"
	"strings"
	"testing"
)

func hasSemanticError(errors []SemanticError, part string) bool {
	for _, err := range errors {
		if strings.Contains(err.Error(), part) {
			return true
		}
	}
	return false
}

func parseSourceFile(file string, lines ...string) *utils.Codebase {
	tokens := []utils.Token{}
	for i, line := range lines {
		tokens = append(tokens, arlexer.ArlexerWithLineAndFile(line, file, i+1)...)
	}
	tokens = append(tokens, arlexer.EOFTokenWithFile(file, len(lines)+1))
	return ParseARLang(utils.NewParser(tokens))
}

func TestSemanticAnalyzerAcceptsCustomTypes(t *testing.T) {
	program := parseSource(
		"type GenderEnum as enum {",
		`  MALE:"MAN"`,
		`  FEMALE:"WOMAN"`,
		"}",
		`type AgeClass = "young", "adult", "old"`,
		"interface User {",
		"  name: string",
		"  gender: GenderEnum",
		"  age: AgeClass",
		"}",
		"const User user = {",
		`  name: "Arif Aydın"`,
		"  gender: GenderEnum.MALE",
		`  age: "adult"`,
		"}",
	)

	errors := Analyze(program)
	if len(errors) != 0 {
		t.Fatalf("expected no semantic errors, got %+v", errors)
	}
}

func TestSemanticAnalyzerReportsTypeErrors(t *testing.T) {
	program := parseSource(
		`type AgeClass = "young", "adult", "old"`,
		`const AgeClass age = "ancient"`,
		"type GenderEnum as enum {",
		`  MALE:"MAN"`,
		"}",
		"const GenderEnum gender = GenderEnum.UNKNOWN",
	)

	errors := Analyze(program)
	if !hasSemanticError(errors, "AgeClass") && !hasSemanticError(errors, "cannot assign") {
		t.Fatalf("expected AgeClass assignment error, got %+v", errors)
	}
	if !hasSemanticError(errors, `enum "GenderEnum" has no member "UNKNOWN"`) {
		t.Fatalf("expected enum member error, got %+v", errors)
	}
}

func TestSemanticAnalyzerConstAndFunctionReturn(t *testing.T) {
	program := parseSource(
		"const int age = 18",
		"age = 19",
		"func int main(){",
		`  return "nope"`,
		"}",
	)

	errors := Analyze(program)
	if !hasSemanticError(errors, "cannot reassign const") {
		t.Fatalf("expected const reassignment error, got %+v", errors)
	}
	if !hasSemanticError(errors, "return value") {
		t.Fatalf("expected return type error, got %+v", errors)
	}
}

func TestSemanticErrorIncludesCodeAndLocation(t *testing.T) {
	program := parseSourceFile(
		"sample.ar",
		"const int age = 18",
		"age = 19",
	)

	errors := Analyze(program)
	if len(errors) == 0 {
		t.Fatal("expected semantic error")
	}
	if errors[0].Code == "" || errors[0].File != "sample.ar" || errors[0].Line == 0 || errors[0].Column == 0 {
		t.Fatalf("expected code and location, got %+v", errors[0])
	}
	if !strings.Contains(errors[0].Error(), "sample.ar:2:1") {
		t.Fatalf("unexpected formatted error: %s", errors[0].Error())
	}
}
