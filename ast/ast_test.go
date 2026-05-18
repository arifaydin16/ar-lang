package ast

import (
	"first-api/arlexer"
	"first-api/utils"
	"testing"
)

func parseSource(lines ...string) *utils.Codebase {
	tokens := []utils.Token{}
	for i, line := range lines {
		tokens = append(tokens, arlexer.ArlexerWithLine(line, i+1)...)
	}
	tokens = append(tokens, arlexer.EOFToken(len(lines)+1))
	return ParseARLang(utils.NewParser(tokens))
}

func TestParseFunctionSignatureAndReturn(t *testing.T) {
	program := parseSource(
		"func int main(string arg1, int arg2, arg3){",
		"  return 0;",
		"}",
	)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
	fn, ok := program.Statements[0].(*utils.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if fn.ReturnType != "int" || fn.Name != "main" {
		t.Fatalf("unexpected function signature: %+v", fn)
	}
	if len(fn.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(fn.Args))
	}
	if fn.Args[0].Type != "string" || fn.Args[0].Name != "arg1" {
		t.Fatalf("unexpected first arg: %+v", fn.Args[0])
	}
	if fn.Args[2].Type != "" || fn.Args[2].Name != "arg3" {
		t.Fatalf("unexpected untyped arg: %+v", fn.Args[2])
	}
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body.Statements))
	}
}

func TestParseIfElseAndExpressions(t *testing.T) {
	program := parseSource(
		"func int main(){",
		"  var int[] nums = [1, 2, -3];",
		"  if(data == 250 & (data2 > 59 | data3 < 24)){",
		"    return sum(nums[0], call(\"x\\\"y\"));",
		"  } elsif(!active) {",
		"    return -1;",
		"  } else {",
		"    return 0;",
		"  }",
		"}",
	)

	fn := program.Statements[0].(*utils.FunctionStatement)
	if len(fn.Body.Statements) != 2 {
		t.Fatalf("expected declaration and if statement, got %d", len(fn.Body.Statements))
	}
	ifStmt, ok := fn.Body.Statements[1].(*utils.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", fn.Body.Statements[1])
	}
	if ifStmt.Condition == nil || ifStmt.Consequence == nil || ifStmt.Alternative == nil {
		t.Fatalf("if statement missing condition or blocks: %+v", ifStmt)
	}
	if len(ifStmt.ElseIfs) != 1 {
		t.Fatalf("expected 1 elsif, got %d", len(ifStmt.ElseIfs))
	}
	if len(ifStmt.Consequence.Statements) != 1 {
		t.Fatalf("expected 1 consequence statement, got %d", len(ifStmt.Consequence.Statements))
	}
}

func TestParseUnionTypeDeclaration(t *testing.T) {
	program := parseSource(`const int - string - float data = "bu bir string"`)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(utils.AssignmentStatement)
	if !ok {
		t.Fatalf("expected AssignmentStatement, got %T", program.Statements[0])
	}
	if stmt.Variable != "data" {
		t.Fatalf("unexpected variable: %s", stmt.Variable)
	}
	if stmt.Type != "int - string - float" {
		t.Fatalf("unexpected type string: %s", stmt.Type)
	}
	if len(stmt.Types) != 3 || stmt.Types[0] != "int" || stmt.Types[1] != "string" || stmt.Types[2] != "float" {
		t.Fatalf("unexpected types: %+v", stmt.Types)
	}
}

func TestParseCustomTypesAndObjectLiteral(t *testing.T) {
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

	if len(program.Statements) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(program.Statements))
	}
	enumStmt, ok := program.Statements[0].(*utils.TypeStatement)
	if !ok || enumStmt.Kind != "enum" || len(enumStmt.Members) != 2 {
		t.Fatalf("unexpected enum statement: %#v", program.Statements[0])
	}
	aliasStmt, ok := program.Statements[1].(*utils.TypeStatement)
	if !ok || aliasStmt.Kind != "alias" || len(aliasStmt.Values) != 3 {
		t.Fatalf("unexpected alias statement: %#v", program.Statements[1])
	}
	interfaceStmt, ok := program.Statements[2].(*utils.InterfaceStatement)
	if !ok || len(interfaceStmt.Fields) != 3 {
		t.Fatalf("unexpected interface statement: %#v", program.Statements[2])
	}
	assignment, ok := program.Statements[3].(utils.AssignmentStatement)
	if !ok {
		t.Fatalf("expected assignment, got %T", program.Statements[3])
	}
	if assignment.Type != "User" || assignment.Variable != "user" {
		t.Fatalf("unexpected user assignment: %+v", assignment)
	}
	object, ok := assignment.Value.(utils.ObjectExpression)
	if !ok || len(object.Properties) != 3 {
		t.Fatalf("expected object expression, got %#v", assignment.Value)
	}
	if _, ok := object.Properties[1].Value.(utils.MemberExpression); !ok {
		t.Fatalf("expected member expression for gender, got %#v", object.Properties[1].Value)
	}
}

func TestParseImportExportStatements(t *testing.T) {
	moduleA := parseSource(
		"const int age = 18;",
		"const string name = 20;",
		"export default age;",
	)

	if len(moduleA.Statements) != 3 {
		t.Fatalf("expected 3 statements in module A, got %d", len(moduleA.Statements))
	}
	exportStmt, ok := moduleA.Statements[2].(*utils.ExportStatement)
	if !ok {
		t.Fatalf("expected ExportStatement, got %T", moduleA.Statements[2])
	}
	if !exportStmt.Default {
		t.Fatalf("expected default export: %+v", exportStmt)
	}
	identifier, ok := exportStmt.Value.(utils.IdentifierExpression)
	if !ok || identifier.Value != "age" {
		t.Fatalf("unexpected export value: %#v", exportStmt.Value)
	}

	moduleB := parseSource(`import age, { name } from "a.go"`)
	if len(moduleB.Statements) != 1 {
		t.Fatalf("expected 1 statement in module B, got %d", len(moduleB.Statements))
	}
	importStmt, ok := moduleB.Statements[0].(*utils.ImportStatement)
	if !ok {
		t.Fatalf("expected ImportStatement, got %T", moduleB.Statements[0])
	}
	if importStmt.DefaultImport != "age" || importStmt.Source != "a.go" {
		t.Fatalf("unexpected import statement: %+v", importStmt)
	}
	if len(importStmt.NamedImports) != 1 || importStmt.NamedImports[0].Name != "name" {
		t.Fatalf("unexpected named imports: %+v", importStmt.NamedImports)
	}
}

func TestParseObjectAndAnyTypes(t *testing.T) {
	program := parseSource(
		"const object dt = {",
		`  name: "Arif"`,
		"}",
		"const any value = dt",
	)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	objectDecl, ok := program.Statements[0].(utils.AssignmentStatement)
	if !ok {
		t.Fatalf("expected object assignment, got %T", program.Statements[0])
	}
	if objectDecl.Type != "object" || objectDecl.Variable != "dt" {
		t.Fatalf("unexpected object declaration: %+v", objectDecl)
	}
	objectExpr, ok := objectDecl.Value.(utils.ObjectExpression)
	if !ok || len(objectExpr.Properties) != 1 {
		t.Fatalf("expected object expression, got %#v", objectDecl.Value)
	}

	anyDecl, ok := program.Statements[1].(utils.AssignmentStatement)
	if !ok {
		t.Fatalf("expected any assignment, got %T", program.Statements[1])
	}
	if anyDecl.Type != "any" || anyDecl.Variable != "value" {
		t.Fatalf("unexpected any declaration: %+v", anyDecl)
	}
}
