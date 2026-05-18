package ast

import (
	"first-api/utils"
	"fmt"
)

type SemanticError struct {
	Code    string
	Message string
	File    string
	Line    int
	Column  int
}

func (e SemanticError) Error() string {
	location := e.File
	if location == "" {
		location = "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d [%s] %s", location, e.Line, e.Column, e.Code, e.Message)
}

type Symbol struct {
	Name  string
	Types []string
	Scope string
}

type TypeInfo struct {
	Name    string
	Kind    string
	Values  []utils.Expression
	Members map[string]utils.Expression
	Fields  map[string]string
}

type Analyzer struct {
	Symbols   map[string]Symbol
	Types     map[string]TypeInfo
	Functions map[string]*utils.FunctionStatement
	Errors    []SemanticError
	scopes    []map[string]Symbol
}

func NewAnalyzer() *Analyzer {
	analyzer := &Analyzer{
		Symbols:   map[string]Symbol{},
		Types:     map[string]TypeInfo{},
		Functions: map[string]*utils.FunctionStatement{},
		Errors:    []SemanticError{},
		scopes:    []map[string]Symbol{{}},
	}
	for _, primitive := range []string{"int", "float", "bool", "string", "object", "any"} {
		analyzer.Types[primitive] = TypeInfo{Name: primitive, Kind: "primitive"}
	}
	return analyzer
}

func Analyze(program *utils.Codebase) []SemanticError {
	analyzer := NewAnalyzer()
	analyzer.Analyze(program)
	return analyzer.Errors
}

func (a *Analyzer) Analyze(program *utils.Codebase) {
	if program == nil {
		return
	}
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt, "")
	}
}

func (a *Analyzer) addError(code string, pos utils.Position, format string, args ...interface{}) {
	a.Errors = append(a.Errors, SemanticError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		File:    pos.File,
		Line:    pos.Line,
		Column:  pos.Column,
	})
}

func (a *Analyzer) pushScope() {
	a.scopes = append(a.scopes, map[string]Symbol{})
}

func (a *Analyzer) popScope() {
	if len(a.scopes) > 1 {
		a.scopes = a.scopes[:len(a.scopes)-1]
	}
}

func (a *Analyzer) declare(symbol Symbol) {
	current := a.scopes[len(a.scopes)-1]
	if _, exists := current[symbol.Name]; exists {
		a.addError("ARLANG_SEMANTIC_ERR_001", utils.Position{}, "symbol %q already declared in this scope", symbol.Name)
		return
	}
	current[symbol.Name] = symbol
	if len(a.scopes) == 1 {
		a.Symbols[symbol.Name] = symbol
	}
}

func (a *Analyzer) resolve(name string) (Symbol, bool) {
	for i := len(a.scopes) - 1; i >= 0; i-- {
		if symbol, ok := a.scopes[i][name]; ok {
			return symbol, true
		}
	}
	return Symbol{}, false
}

func (a *Analyzer) analyzeStatement(stmt utils.Statement, expectedReturn string) {
	switch node := stmt.(type) {
	case utils.TypeStatement:
		a.registerType(node)
	case *utils.TypeStatement:
		a.registerType(*node)
	case utils.InterfaceStatement:
		a.registerInterface(node)
	case *utils.InterfaceStatement:
		a.registerInterface(*node)
	case utils.AssignmentStatement:
		a.analyzeAssignment(node)
	case utils.ReassignmentStatement:
		a.analyzeReassignment(node)
	case utils.ReturnStatement:
		a.analyzeReturn(node, expectedReturn)
	case utils.ExpressionStatement:
		a.inferExpression(node.Expression)
	case utils.UnaryStatement:
		a.analyzeUnary(node)
	case *utils.IfStatement:
		a.analyzeIf(node, expectedReturn)
	case *utils.ForStatement:
		a.analyzeFor(node, expectedReturn)
	case *utils.FunctionStatement:
		a.analyzeFunction(node)
	case *utils.ExportStatement:
		a.analyzeExport(node)
	case *utils.ImportStatement:
		a.analyzeImport(node)
	}
}

func (a *Analyzer) registerType(stmt utils.TypeStatement) {
	if stmt.Name == "" {
		a.addError("ARLANG_SEMANTIC_ERR_100", stmt.Pos, "type declaration missing name")
		return
	}
	if _, exists := a.Types[stmt.Name]; exists {
		a.addError("ARLANG_SEMANTIC_ERR_101", stmt.Pos, "type %q already declared", stmt.Name)
		return
	}
	info := TypeInfo{
		Name:    stmt.Name,
		Kind:    stmt.Kind,
		Values:  stmt.Values,
		Members: map[string]utils.Expression{},
	}
	for _, member := range stmt.Members {
		info.Members[member.Name] = member.Value
	}
	a.Types[stmt.Name] = info
}

func (a *Analyzer) registerInterface(stmt utils.InterfaceStatement) {
	if stmt.Name == "" {
		a.addError("ARLANG_SEMANTIC_ERR_110", stmt.Pos, "interface declaration missing name")
		return
	}
	if _, exists := a.Types[stmt.Name]; exists {
		a.addError("ARLANG_SEMANTIC_ERR_101", stmt.Pos, "type %q already declared", stmt.Name)
		return
	}
	fields := map[string]string{}
	for _, field := range stmt.Fields {
		fields[field.Name] = field.Type
		a.ensureTypeExists(field.Type)
	}
	a.Types[stmt.Name] = TypeInfo{Name: stmt.Name, Kind: "interface", Fields: fields}
}

func (a *Analyzer) analyzeAssignment(stmt utils.AssignmentStatement) {
	if stmt.Variable == "" {
		a.addError("ARLANG_SEMANTIC_ERR_200", stmt.Pos, "assignment missing variable name")
		return
	}
	for _, typ := range stmt.Types {
		a.ensureTypeExists(typ)
	}
	if len(stmt.Types) > 0 && !a.expressionAssignableTo(stmt.Value, stmt.Types) {
		a.addError("ARLANG_SEMANTIC_ERR_201", stmt.Pos, "cannot assign value of type %q to %s %q", a.inferExpression(stmt.Value), stmt.Type, stmt.Variable)
	}
	a.declare(Symbol{Name: stmt.Variable, Types: stmt.Types, Scope: stmt.Scope})
}

func (a *Analyzer) analyzeReassignment(stmt utils.ReassignmentStatement) {
	symbol, ok := a.resolve(stmt.Variable)
	if !ok {
		a.addError("ARLANG_SEMANTIC_ERR_202", stmt.Pos, "cannot assign to undeclared symbol %q", stmt.Variable)
		return
	}
	if symbol.Scope == "const" {
		a.addError("ARLANG_SEMANTIC_ERR_203", stmt.Pos, "cannot reassign const symbol %q", stmt.Variable)
	}
	if len(symbol.Types) > 0 && !a.expressionAssignableTo(stmt.Value, symbol.Types) {
		a.addError("ARLANG_SEMANTIC_ERR_201", stmt.Pos, "cannot assign value of type %q to %q", a.inferExpression(stmt.Value), stmt.Variable)
	}
}

func (a *Analyzer) analyzeReturn(stmt utils.ReturnStatement, expectedReturn string) {
	if expectedReturn == "" || expectedReturn == "any" {
		a.inferExpression(stmt.Value)
		return
	}
	if !a.expressionAssignableTo(stmt.Value, []string{expectedReturn}) {
		a.addError("ARLANG_SEMANTIC_ERR_300", stmt.Pos, "return value of type %q is not assignable to %q", a.inferExpression(stmt.Value), expectedReturn)
	}
}

func (a *Analyzer) analyzeUnary(stmt utils.UnaryStatement) {
	if _, ok := a.resolve(stmt.Variable); !ok {
		a.addError("ARLANG_SEMANTIC_ERR_204", stmt.Pos, "cannot use unary operator on undeclared symbol %q", stmt.Variable)
	}
}

func (a *Analyzer) analyzeIf(stmt *utils.IfStatement, expectedReturn string) {
	a.inferExpression(stmt.Condition)
	a.analyzeBlock(stmt.Consequence, expectedReturn)
	for _, elsif := range stmt.ElseIfs {
		a.inferExpression(elsif.Condition)
		a.analyzeBlock(elsif.Consequence, expectedReturn)
	}
	a.analyzeBlock(stmt.Alternative, expectedReturn)
}

func (a *Analyzer) analyzeFor(stmt *utils.ForStatement, expectedReturn string) {
	a.pushScope()
	if stmt.Init != nil {
		a.analyzeStatement(stmt.Init, expectedReturn)
	}
	a.inferExpression(stmt.Condition)
	if stmt.Post != nil {
		a.analyzeStatement(stmt.Post, expectedReturn)
	}
	a.analyzeBlock(stmt.Body, expectedReturn)
	a.popScope()
}

func (a *Analyzer) analyzeFunction(stmt *utils.FunctionStatement) {
	if stmt.Name == "" {
		a.addError("ARLANG_SEMANTIC_ERR_400", stmt.Pos, "function declaration missing name")
		return
	}
	if _, exists := a.Functions[stmt.Name]; exists {
		a.addError("ARLANG_SEMANTIC_ERR_401", stmt.Pos, "function %q already declared", stmt.Name)
		return
	}
	if stmt.ReturnType != "" {
		a.ensureTypeExists(stmt.ReturnType)
	}
	a.Functions[stmt.Name] = stmt

	a.pushScope()
	for _, arg := range stmt.Args {
		if arg.Type != "" {
			a.ensureTypeExists(arg.Type)
		}
		a.declare(Symbol{Name: arg.Name, Types: typeList(arg.Type), Scope: "arg"})
	}
	a.analyzeBlock(stmt.Body, stmt.ReturnType)
	a.popScope()
}

func (a *Analyzer) analyzeExport(stmt *utils.ExportStatement) {
	if stmt.Declaration != nil {
		a.analyzeStatement(stmt.Declaration, "")
		return
	}
	for _, name := range stmt.Names {
		if _, ok := a.resolve(name); !ok {
			if _, fnOK := a.Functions[name]; !fnOK {
				a.addError("ARLANG_SEMANTIC_ERR_500", stmt.Pos, "cannot export undeclared symbol %q", name)
			}
		}
	}
	if stmt.Value != nil {
		a.inferExpression(stmt.Value)
	}
}

func (a *Analyzer) analyzeImport(stmt *utils.ImportStatement) {
	if stmt.DefaultImport != "" {
		a.declare(Symbol{Name: stmt.DefaultImport, Types: []string{"any"}, Scope: "import"})
	}
	for _, imported := range stmt.NamedImports {
		name := imported.Name
		if imported.Alias != "" {
			name = imported.Alias
		}
		a.declare(Symbol{Name: name, Types: []string{"any"}, Scope: "import"})
	}
}

func (a *Analyzer) analyzeBlock(block *utils.BlockStatement, expectedReturn string) {
	if block == nil {
		return
	}
	a.pushScope()
	for _, stmt := range block.Statements {
		a.analyzeStatement(stmt, expectedReturn)
	}
	a.popScope()
}

func (a *Analyzer) ensureTypeExists(typeName string) {
	if typeName == "" || typeName == "any" {
		return
	}
	if _, ok := a.Types[typeName]; ok {
		return
	}
	a.addError("ARLANG_SEMANTIC_ERR_102", utils.Position{}, "unknown type %q", typeName)
}

func (a *Analyzer) inferExpression(expr utils.Expression) string {
	switch node := expr.(type) {
	case nil:
		return "void"
	case utils.LiteralExpression:
		switch node.Value.(type) {
		case int, int64:
			return "int"
		case float32, float64:
			return "float"
		case bool:
			return "bool"
		case string:
			return "string"
		}
	case utils.ValueExpression:
		return "any"
	case utils.IdentifierExpression:
		symbol, ok := a.resolve(node.Value)
		if !ok {
			a.addError("ARLANG_SEMANTIC_ERR_205", node.Pos, "undeclared symbol %q", node.Value)
			return "any"
		}
		if len(symbol.Types) == 1 {
			return symbol.Types[0]
		}
		return "any"
	case utils.ArrayExpression:
		for _, element := range node.Elements {
			a.inferExpression(element)
		}
		return "array"
	case utils.ObjectExpression:
		for _, property := range node.Properties {
			a.inferExpression(property.Value)
		}
		return "object"
	case utils.MemberExpression:
		return a.inferMemberExpression(node)
	case utils.CallExpression:
		return a.inferCallExpression(node)
	case utils.IndexExpression:
		a.inferExpression(node.Left)
		a.inferExpression(node.Index)
		return "any"
	case utils.PrefixExpression:
		right := a.inferExpression(node.Right)
		if node.Operator == "!" {
			return "bool"
		}
		return right
	case utils.InfixExpression:
		left := a.inferExpression(node.Left)
		right := a.inferExpression(node.Right)
		if node.Operator == "&" || node.Operator == "|" || node.Operator == "==" || node.Operator == "!=" || node.Operator == "<" || node.Operator == ">" || node.Operator == "<=" || node.Operator == ">=" {
			return "bool"
		}
		if left == "float" || right == "float" {
			return "float"
		}
		if left == right {
			return left
		}
		return "any"
	}
	return "any"
}

func (a *Analyzer) inferMemberExpression(expr utils.MemberExpression) string {
	identifier, ok := expr.Object.(utils.IdentifierExpression)
	if !ok {
		a.inferExpression(expr.Object)
		return "any"
	}
	typeInfo, ok := a.Types[identifier.Value]
	if !ok {
		if _, symbolOK := a.resolve(identifier.Value); !symbolOK {
			a.addError("ARLANG_SEMANTIC_ERR_206", expr.Pos, "undeclared symbol or type %q", identifier.Value)
		}
		return "any"
	}
	if typeInfo.Kind == "enum" {
		if _, exists := typeInfo.Members[expr.Property]; !exists {
			a.addError("ARLANG_SEMANTIC_ERR_120", expr.Pos, "enum %q has no member %q", identifier.Value, expr.Property)
		}
		return identifier.Value
	}
	return "any"
}

func (a *Analyzer) inferCallExpression(expr utils.CallExpression) string {
	fn, ok := a.Functions[expr.Function]
	if !ok {
		a.addError("ARLANG_SEMANTIC_ERR_402", expr.Pos, "call to undeclared function %q", expr.Function)
		for _, arg := range expr.Args {
			a.inferExpression(arg)
		}
		return "any"
	}
	if len(fn.Args) != len(expr.Args) {
		a.addError("ARLANG_SEMANTIC_ERR_403", expr.Pos, "function %q expects %d args, got %d", expr.Function, len(fn.Args), len(expr.Args))
	}
	for i, arg := range expr.Args {
		if i < len(fn.Args) && fn.Args[i].Type != "" && !a.expressionAssignableTo(arg, []string{fn.Args[i].Type}) {
			a.addError("ARLANG_SEMANTIC_ERR_404", expr.Pos, "argument %d for function %q has type %q, expected %q", i+1, expr.Function, a.inferExpression(arg), fn.Args[i].Type)
		} else {
			a.inferExpression(arg)
		}
	}
	if fn.ReturnType == "" {
		return "void"
	}
	return fn.ReturnType
}

func (a *Analyzer) expressionAssignableTo(expr utils.Expression, targetTypes []string) bool {
	if len(targetTypes) == 0 {
		return true
	}
	for _, target := range targetTypes {
		if target == "any" {
			a.inferExpression(expr)
			return true
		}
		if a.expressionMatchesType(expr, target) {
			return true
		}
	}
	return false
}

func (a *Analyzer) expressionMatchesType(expr utils.Expression, target string) bool {
	if target == "" || target == "any" {
		a.inferExpression(expr)
		return true
	}
	if info, ok := a.Types[target]; ok {
		switch info.Kind {
		case "primitive":
			return a.inferExpression(expr) == target
		case "alias":
			return a.matchesAlias(expr, info)
		case "enum":
			if member, ok := expr.(utils.MemberExpression); ok {
				if identifier, ok := member.Object.(utils.IdentifierExpression); ok && identifier.Value == target {
					_, exists := info.Members[member.Property]
					if !exists {
						a.addError("ARLANG_SEMANTIC_ERR_120", member.Pos, "enum %q has no member %q", target, member.Property)
					}
					return exists
				}
			}
			return false
		case "interface":
			return a.matchesInterface(expr, info)
		}
	}
	return a.inferExpression(expr) == target
}

func (a *Analyzer) matchesAlias(expr utils.Expression, info TypeInfo) bool {
	exprValue, ok := literalComparableValue(expr)
	if !ok {
		return a.inferExpression(expr) == info.Name
	}
	for _, allowed := range info.Values {
		allowedValue, allowedOK := literalComparableValue(allowed)
		if allowedOK && allowedValue == exprValue {
			return true
		}
	}
	return false
}

func (a *Analyzer) matchesInterface(expr utils.Expression, info TypeInfo) bool {
	object, ok := expr.(utils.ObjectExpression)
	if !ok {
		return false
	}
	properties := map[string]utils.Expression{}
	for _, property := range object.Properties {
		properties[property.Key] = property.Value
	}
	for field, fieldType := range info.Fields {
		value, exists := properties[field]
		if !exists {
			a.addError("ARLANG_SEMANTIC_ERR_130", expressionPosition(expr), "object missing field %q for interface %q", field, info.Name)
			return false
		}
		if !a.expressionAssignableTo(value, []string{fieldType}) {
			a.addError("ARLANG_SEMANTIC_ERR_131", expressionPosition(value), "field %q is not assignable to type %q", field, fieldType)
			return false
		}
	}
	return true
}

func literalComparableValue(expr utils.Expression) (interface{}, bool) {
	switch node := expr.(type) {
	case utils.LiteralExpression:
		return node.Value, true
	case utils.ValueExpression:
		return node.Value, true
	}
	return nil, false
}

func expressionPosition(expr utils.Expression) utils.Position {
	switch node := expr.(type) {
	case utils.IdentifierExpression:
		return node.Pos
	case utils.LiteralExpression:
		return node.Pos
	case utils.ValueExpression:
		return node.Pos
	case utils.PrefixExpression:
		return node.Pos
	case utils.InfixExpression:
		return node.Pos
	case utils.LogicalExpression:
		return node.Pos
	case utils.CallExpression:
		return node.Pos
	case utils.ArrayExpression:
		return node.Pos
	case utils.IndexExpression:
		return node.Pos
	case utils.MemberExpression:
		return node.Pos
	case utils.ObjectExpression:
		return node.Pos
	}
	return utils.Position{}
}

func typeList(typeName string) []string {
	if typeName == "" {
		return nil
	}
	return []string{typeName}
}
