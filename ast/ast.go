package ast

import (
	"first-api/utils"
	"strconv"
	"strings"
)

const (
	_ int = iota
	lowest
	assignPrecedence
	logicalOrPrecedence
	logicalAndPrecedence
	comparePrecedence
	sumPrecedence
	productPrecedence
	prefixPrecedence
	callPrecedence
	indexPrecedence
)

func currentLiteral(p *utils.Parser) string {
	if p.Current() == nil {
		return ""
	}
	return p.Current().Literal
}

func currentType(p *utils.Parser) string {
	if p.Current() == nil {
		return ""
	}
	return p.Current().Type
}

func currentPosition(p *utils.Parser) utils.Position {
	if p.Current() == nil {
		return utils.Position{}
	}
	token := p.Current()
	return utils.Position{File: token.File, Line: token.Line, Column: token.Column}
}

func tokenPosition(token *utils.Token) utils.Position {
	if token == nil {
		return utils.Position{}
	}
	return utils.Position{File: token.File, Line: token.Line, Column: token.Column}
}

func tokenPrecedence(token *utils.Token) int {
	if token == nil {
		return lowest
	}
	switch token.Type {
	case "ASSIGNMENT":
		return assignPrecedence
	case "LOGICAL":
		if token.Literal == "|" {
			return logicalOrPrecedence
		}
		if token.Literal == "&" {
			return logicalAndPrecedence
		}
	case "COMPARISON":
		return comparePrecedence
	case "ARITHMETIC":
		if token.Literal == "+" || token.Literal == "-" {
			return sumPrecedence
		}
		return productPrecedence
	case "PAREN":
		if token.Literal == "(" {
			return callPrecedence
		}
	case "BRACKET":
		if token.Literal == "[" {
			return indexPrecedence
		}
	case "DOT":
		return indexPrecedence
	}
	return lowest
}

func isExpressionEnd(token *utils.Token) bool {
	if token == nil {
		return true
	}
	return token.Type == "EOF" ||
		token.Type == "SEMICOLON" ||
		token.Type == "COMMA" ||
		token.Literal == ")" ||
		token.Literal == "]" ||
		token.Literal == "}"
}

func parseLiteral(token *utils.Token) utils.Expression {
	if token == nil {
		return nil
	}
	switch token.Type {
	case "NUMBER":
		if strings.Contains(token.Literal, ".") {
			value, err := strconv.ParseFloat(token.Literal, 64)
			if err == nil {
				return utils.LiteralExpression{Pos: tokenPosition(token), Value: value}
			}
		}
		value, err := strconv.ParseInt(token.Literal, 10, 64)
		if err == nil {
			return utils.LiteralExpression{Pos: tokenPosition(token), Value: value}
		}
	case "STRING_LITERAL":
		value, err := strconv.Unquote(token.Literal)
		if err == nil {
			return utils.LiteralExpression{Pos: tokenPosition(token), Value: value}
		}
		return utils.LiteralExpression{Pos: tokenPosition(token), Value: token.Literal}
	case "KEYWORD":
		if token.Literal == "true" {
			return utils.LiteralExpression{Pos: tokenPosition(token), Value: true}
		}
		if token.Literal == "false" {
			return utils.LiteralExpression{Pos: tokenPosition(token), Value: false}
		}
	}
	return utils.ValueExpression{Pos: tokenPosition(token), Value: token.Literal}
}

func ParseExpression(p *utils.Parser, precedence int) utils.Expression {
	token := p.Current()
	if token == nil || isExpressionEnd(token) {
		return nil
	}

	var left utils.Expression
	switch token.Type {
	case "IDENTIFIER":
		left = utils.IdentifierExpression{Pos: tokenPosition(token), Value: token.Literal}
		p.Next()
	case "NUMBER", "STRING_LITERAL", "KEYWORD":
		left = parseLiteral(token)
		p.Next()
	case "LOGICAL", "ARITHMETIC":
		if token.Literal == "!" || token.Literal == "-" {
			operator := token.Literal
			p.Next()
			left = utils.PrefixExpression{
				Pos:      tokenPosition(token),
				Operator: operator,
				Right:    ParseExpression(p, prefixPrecedence),
			}
		} else {
			left = parseLiteral(token)
			p.Next()
		}
	case "PAREN":
		if token.Literal == "(" {
			p.Next()
			left = ParseExpression(p, lowest)
			if p.Expect(")", "LITERAL") {
				p.Next()
			}
		}
	case "BRACKET":
		if token.Literal == "[" {
			left = parseArrayExpression(p)
		}
	case "BRACE":
		if token.Literal == "{" {
			left = parseObjectExpression(p)
		}
	default:
		left = parseLiteral(token)
		p.Next()
	}

	for !isExpressionEnd(p.Current()) && precedence < tokenPrecedence(p.Current()) {
		token = p.Current()
		switch token.Type {
		case "ARITHMETIC", "COMPARISON", "LOGICAL":
			operator := token.Literal
			nextPrecedence := tokenPrecedence(token)
			p.Next()
			right := ParseExpression(p, nextPrecedence)
			left = utils.InfixExpression{
				Pos:      tokenPosition(token),
				Left:     left,
				Operator: operator,
				Right:    right,
			}
		case "PAREN":
			if token.Literal != "(" {
				return left
			}
			left = parseCallExpression(p, left)
		case "BRACKET":
			if token.Literal != "[" {
				return left
			}
			left = parseIndexExpression(p, left)
		case "DOT":
			left = parseMemberExpression(p, left)
		default:
			return left
		}
	}

	return left
}

func parseExpressionList(p *utils.Parser, end string) []utils.Expression {
	expressions := []utils.Expression{}
	if p.Expect(end, "LITERAL") {
		p.Next()
		return expressions
	}

	for p.Current() != nil && !p.Expect(end, "LITERAL") {
		expr := ParseExpression(p, lowest)
		if expr != nil {
			expressions = append(expressions, expr)
		}
		if p.Expect(",", "LITERAL") {
			p.Next()
			continue
		}
	}
	if p.Expect(end, "LITERAL") {
		p.Next()
	}
	return expressions
}

func parseCallExpression(p *utils.Parser, left utils.Expression) utils.Expression {
	name := left.TokenLiteral()
	p.Next()
	return utils.CallExpression{
		Pos:      currentPosition(p),
		Function: name,
		Args:     parseExpressionList(p, ")"),
	}
}

func parseArrayExpression(p *utils.Parser) utils.Expression {
	p.Next()
	return utils.ArrayExpression{Pos: currentPosition(p), Elements: parseExpressionList(p, "]")}
}

func parseIndexExpression(p *utils.Parser, left utils.Expression) utils.Expression {
	p.Next()
	index := ParseExpression(p, lowest)
	if p.Expect("]", "LITERAL") {
		p.Next()
	}
	return utils.IndexExpression{Pos: currentPosition(p), Left: left, Index: index}
}

func parseMemberExpression(p *utils.Parser, left utils.Expression) utils.Expression {
	p.Next()
	property := ""
	if p.Expect("IDENTIFIER", "TYPE") || p.Expect("KEYWORD", "TYPE") {
		property = currentLiteral(p)
		p.Next()
	}
	return utils.MemberExpression{Pos: currentPosition(p), Object: left, Property: property}
}

func parseObjectExpression(p *utils.Parser) utils.Expression {
	pos := currentPosition(p)
	properties := []utils.ObjectProperty{}
	p.Next()
	for p.Current() != nil && !p.Expect("}", "LITERAL") && !p.Expect("EOF", "TYPE") {
		key := currentLiteral(p)
		if p.Expect("IDENTIFIER", "TYPE") || p.Expect("KEYWORD", "TYPE") || p.Expect("STRING_LITERAL", "TYPE") {
			p.Next()
		}
		if p.Expect(":", "LITERAL") {
			p.Next()
		}
		value := ParseExpression(p, lowest)
		properties = append(properties, utils.ObjectProperty{Key: strings.Trim(key, `"`), Value: value})
		if p.Expect(",", "LITERAL") || p.Expect(";", "LITERAL") {
			p.Next()
		}
	}
	if p.Expect("}", "LITERAL") {
		p.Next()
	}
	return utils.ObjectExpression{Pos: pos, Properties: properties}
}

func ParseControlLogic(p *utils.Parser) *utils.LogicalExpression {
	expr := ParseExpression(p, lowest)
	if expr == nil {
		return nil
	}
	if logical, ok := expr.(utils.LogicalExpression); ok {
		return &logical
	}
	if infix, ok := expr.(utils.InfixExpression); ok {
		return &utils.LogicalExpression{
			Pos:      infix.Pos,
			Left:     infix.Left,
			Operator: infix.Operator,
			Right:    infix.Right,
		}
	}
	return &utils.LogicalExpression{Left: expr}
}

func ParseCondition(p *utils.Parser) *utils.ComparisonExpression {
	var variable string = currentLiteral(p)
	var cond string
	var literal string
	if p.Peek(1) != nil {
		cond = p.Peek(1).Literal
	}
	if p.Peek(2) != nil {
		literal = p.Peek(2).Literal
	}
	return &utils.ComparisonExpression{
		Variable:   variable,
		Comparison: cond,
		Compared:   literal,
	}
}

func ParseComparison(p *utils.Parser) utils.ComparisonExpression {
	com := utils.ComparisonExpression{}
	if p.Expect("IDENTIFIER", "TYPE") && p.PeekExpect("COMPARISON", 1, "TYPE") {
		com.Variable = currentLiteral(p)
		com.Comparison = p.Peek(1).Literal
		if p.Peek(2) != nil {
			com.Compared = p.Peek(2).Literal
		}
	}
	return com
}

func ParseUnary(p *utils.Parser) utils.UnaryStatement {
	pos := currentPosition(p)
	if p.Expect("IDENTIFIER", "TYPE") && p.PeekExpect("UNARY", 1, "TYPE") {
		un := utils.UnaryStatement{
			Pos:      pos,
			Variable: currentLiteral(p),
			Unary:    p.Peek(1).Literal,
		}
		p.Next()
		p.Next()
		return un
	}
	if p.Expect("UNARY", "TYPE") && p.PeekExpect("IDENTIFIER", -1, "TYPE") {
		un := utils.UnaryStatement{
			Pos:      pos,
			Variable: p.Peek(-1).Literal,
			Unary:    currentLiteral(p),
		}
		p.Next()
		return un
	}
	return utils.UnaryStatement{}
}

func ParseAssignment(p *utils.Parser) utils.AssignmentStatement {
	pos := currentPosition(p)
	scope := ""
	valueTypes := []string{}

	if p.Expect("KEYWORD", "TYPE") && (currentLiteral(p) == "const" || currentLiteral(p) == "var") {
		scope = currentLiteral(p)
		p.Next()
	}

	for p.Expect("KEYWORD", "TYPE") || p.Expect("IDENTIFIER", "TYPE") {
		valueTypes = append(valueTypes, currentLiteral(p))
		p.Next()
		if p.Expect("[", "LITERAL") && p.PeekExpect("]", 1, "LITERAL") {
			valueTypes[len(valueTypes)-1] += "[]"
			p.Next()
			p.Next()
		}
		if p.Expect("-", "LITERAL") {
			p.Next()
			continue
		}
		break
	}

	variable := currentLiteral(p)
	if p.Expect("IDENTIFIER", "TYPE") {
		p.Next()
	}
	if p.Expect("ASSIGNMENT", "TYPE") {
		p.Next()
	}

	value := ParseExpression(p, lowest)
	return utils.AssignmentStatement{
		Pos:      pos,
		Variable: variable,
		Scope:    scope,
		Type:     strings.Join(valueTypes, " - "),
		Types:    valueTypes,
		Value:    value,
	}
}

func ParseTypeStatement(p *utils.Parser) *utils.TypeStatement {
	stmt := &utils.TypeStatement{Pos: currentPosition(p), Values: []utils.Expression{}, Members: []utils.EnumMember{}}
	p.Next()
	if p.Expect("IDENTIFIER", "TYPE") {
		stmt.Name = currentLiteral(p)
		p.Next()
	}

	if p.Expect("AS", "TYPE") {
		p.Next()
		if p.Expect("ENUM", "TYPE") {
			stmt.Kind = "enum"
			p.Next()
			if p.Expect("{", "LITERAL") {
				p.Next()
			}
			for p.Current() != nil && !p.Expect("}", "LITERAL") && !p.Expect("EOF", "TYPE") {
				memberName := currentLiteral(p)
				if p.Expect("IDENTIFIER", "TYPE") {
					p.Next()
				}
				if p.Expect(":", "LITERAL") {
					p.Next()
				}
				value := ParseExpression(p, lowest)
				stmt.Members = append(stmt.Members, utils.EnumMember{Name: memberName, Value: value})
				if p.Expect(",", "LITERAL") || p.Expect(";", "LITERAL") {
					p.Next()
				}
			}
			if p.Expect("}", "LITERAL") {
				p.Next()
			}
			return stmt
		}
	}

	stmt.Kind = "alias"
	if p.Expect("=", "LITERAL") {
		p.Next()
	}
	for p.Current() != nil && !p.Expect("EOF", "TYPE") && !p.Expect("}", "LITERAL") && !p.Expect(";", "LITERAL") {
		value := ParseExpression(p, lowest)
		if value != nil {
			stmt.Values = append(stmt.Values, value)
		}
		if p.Expect(",", "LITERAL") {
			p.Next()
			continue
		}
		break
	}
	consumeStatementEnd(p)
	return stmt
}

func ParseInterfaceStatement(p *utils.Parser) *utils.InterfaceStatement {
	stmt := &utils.InterfaceStatement{Pos: currentPosition(p), Fields: []utils.InterfaceField{}}
	p.Next()
	if p.Expect("IDENTIFIER", "TYPE") {
		stmt.Name = currentLiteral(p)
		p.Next()
	}
	if p.Expect("{", "LITERAL") {
		p.Next()
	}
	for p.Current() != nil && !p.Expect("}", "LITERAL") && !p.Expect("EOF", "TYPE") {
		field := utils.InterfaceField{}
		if p.Expect("IDENTIFIER", "TYPE") {
			field.Name = currentLiteral(p)
			p.Next()
		}
		if p.Expect(":", "LITERAL") {
			p.Next()
		}
		if p.Expect("KEYWORD", "TYPE") || p.Expect("IDENTIFIER", "TYPE") {
			field.Type = currentLiteral(p)
			p.Next()
		}
		if field.Name != "" {
			stmt.Fields = append(stmt.Fields, field)
		}
		if p.Expect(",", "LITERAL") || p.Expect(";", "LITERAL") {
			p.Next()
		}
	}
	if p.Expect("}", "LITERAL") {
		p.Next()
	}
	return stmt
}

func parseReassignment(p *utils.Parser) utils.Statement {
	pos := currentPosition(p)
	variable := currentLiteral(p)
	p.Next()
	operator := currentLiteral(p)
	p.Next()
	value := ParseExpression(p, lowest)
	return utils.ReassignmentStatement{
		Pos:      pos,
		Variable: variable,
		Operator: operator,
		Value:    value,
	}
}

func ParseReturn(p *utils.Parser) utils.ReturnStatement {
	pos := currentPosition(p)
	p.Next()
	return utils.ReturnStatement{Pos: pos, Value: ParseExpression(p, lowest)}
}

func ParseBlockStatement(p *utils.Parser) *utils.BlockStatement {
	block := &utils.BlockStatement{Statements: []utils.Statement{}}
	if p.Expect("{", "LITERAL") {
		p.Next()
	}

	for p.Current() != nil && !p.Expect("}", "LITERAL") && !p.Expect("EOF", "TYPE") {
		stmt := parseStatement(p)
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			p.Next()
		}
	}

	if p.Expect("}", "LITERAL") {
		p.Next()
	}
	return block
}

func ParseIfStatement(p *utils.Parser) *utils.IfStatement {
	stmt := &utils.IfStatement{Pos: currentPosition(p), ElseIfs: []utils.ElseIfStatement{}}
	p.Next()

	if p.Expect("(", "LITERAL") {
		p.Next()
	}
	stmt.Condition = ParseExpression(p, lowest)
	if p.Expect(")", "LITERAL") {
		p.Next()
	}
	stmt.Consequence = ParseBlockStatement(p)

	for p.Expect("LOGIC", "TYPE") && currentLiteral(p) == "elsif" {
		p.Next()
		if p.Expect("(", "LITERAL") {
			p.Next()
		}
		condition := ParseExpression(p, lowest)
		if p.Expect(")", "LITERAL") {
			p.Next()
		}
		stmt.ElseIfs = append(stmt.ElseIfs, utils.ElseIfStatement{
			Condition:   condition,
			Consequence: ParseBlockStatement(p),
		})
	}

	if p.Expect("LOGIC", "TYPE") && currentLiteral(p) == "else" {
		p.Next()
		stmt.Alternative = ParseBlockStatement(p)
	}

	return stmt
}

func ParseImportStatement(p *utils.Parser) *utils.ImportStatement {
	stmt := &utils.ImportStatement{Pos: currentPosition(p), NamedImports: []utils.ImportSpecifier{}}
	p.Next()

	if p.Expect("IDENTIFIER", "TYPE") {
		stmt.DefaultImport = currentLiteral(p)
		p.Next()
	}

	if p.Expect(",", "LITERAL") {
		p.Next()
	}

	if p.Expect("{", "LITERAL") {
		p.Next()
		for p.Current() != nil && !p.Expect("}", "LITERAL") && !p.Expect("EOF", "TYPE") {
			spec := utils.ImportSpecifier{}
			if p.Expect("IDENTIFIER", "TYPE") {
				spec.Name = currentLiteral(p)
				p.Next()
			}
			if p.Expect("AS", "TYPE") {
				p.Next()
				if p.Expect("IDENTIFIER", "TYPE") {
					spec.Alias = currentLiteral(p)
					p.Next()
				}
			}
			if spec.Name != "" {
				stmt.NamedImports = append(stmt.NamedImports, spec)
			}
			if p.Expect(",", "LITERAL") {
				p.Next()
			}
		}
		if p.Expect("}", "LITERAL") {
			p.Next()
		}
	}

	if p.Expect("FROM", "TYPE") {
		p.Next()
	}
	if p.Expect("STRING_LITERAL", "TYPE") {
		source, err := strconv.Unquote(currentLiteral(p))
		if err == nil {
			stmt.Source = source
		} else {
			stmt.Source = currentLiteral(p)
		}
		p.Next()
	}
	consumeStatementEnd(p)
	return stmt
}

func ParseExportStatement(p *utils.Parser) *utils.ExportStatement {
	stmt := &utils.ExportStatement{Pos: currentPosition(p), Names: []string{}}
	p.Next()

	if p.Expect("DEFAULT", "TYPE") {
		stmt.Default = true
		p.Next()
	}

	if p.Expect("{", "LITERAL") {
		p.Next()
		for p.Current() != nil && !p.Expect("}", "LITERAL") && !p.Expect("EOF", "TYPE") {
			if p.Expect("IDENTIFIER", "TYPE") {
				stmt.Names = append(stmt.Names, currentLiteral(p))
				p.Next()
			}
			if p.Expect(",", "LITERAL") {
				p.Next()
			}
		}
		if p.Expect("}", "LITERAL") {
			p.Next()
		}
		consumeStatementEnd(p)
		return stmt
	}

	switch currentType(p) {
	case "KEYWORD":
		if currentLiteral(p) == "const" || currentLiteral(p) == "var" {
			decl := ParseAssignment(p)
			stmt.Declaration = decl
			consumeStatementEnd(p)
			return stmt
		}
	case "FUNCTION":
		stmt.Declaration = ParseFunction(p)
		return stmt
	case "TYPEDEF":
		stmt.Declaration = ParseTypeStatement(p)
		return stmt
	case "INTERFACE":
		stmt.Declaration = ParseInterfaceStatement(p)
		return stmt
	}

	stmt.Value = ParseExpression(p, lowest)
	consumeStatementEnd(p)
	return stmt
}

func ParseFunction(p *utils.Parser) *utils.FunctionStatement {
	fn := &utils.FunctionStatement{Pos: currentPosition(p), Args: []utils.FunctionArgument{}}
	p.Next()

	if p.Current() == nil {
		return fn
	}

	if p.Current().Type == "KEYWORD" && p.Peek(1) != nil && p.Peek(1).Type == "IDENTIFIER" && p.Peek(2) != nil && p.Peek(2).Literal == "(" {
		fn.ReturnType = currentLiteral(p)
		p.Next()
	}

	if p.Expect("IDENTIFIER", "TYPE") {
		fn.Name = currentLiteral(p)
		p.Next()
	}

	if p.Expect("(", "LITERAL") {
		p.Next()
	}
	for p.Current() != nil && !p.Expect(")", "LITERAL") {
		argType := ""
		argName := ""
		if p.Current().Type == "KEYWORD" && p.Peek(1) != nil && p.Peek(1).Type == "IDENTIFIER" {
			argType = currentLiteral(p)
			p.Next()
		}
		if p.Expect("IDENTIFIER", "TYPE") {
			argName = currentLiteral(p)
			p.Next()
		}
		if argName != "" {
			fn.Args = append(fn.Args, utils.FunctionArgument{Type: argType, Name: argName})
		}
		if p.Expect(",", "LITERAL") {
			p.Next()
		}
	}
	if p.Expect(")", "LITERAL") {
		p.Next()
	}
	fn.Body = ParseBlockStatement(p)
	return fn
}

func parseForPart(p *utils.Parser) utils.Statement {
	if p.Expect("KEYWORD", "TYPE") {
		return ParseAssignment(p)
	}
	if p.Expect("IDENTIFIER", "TYPE") && p.PeekExpect("ASSIGNMENT", 1, "TYPE") {
		return parseReassignment(p)
	}
	if p.Expect("IDENTIFIER", "TYPE") && p.PeekExpect("UNARY", 1, "TYPE") {
		return ParseUnary(p)
	}
	expr := ParseExpression(p, lowest)
	if expr != nil {
		return utils.ExpressionStatement{Expression: expr}
	}
	return nil
}

func ParseForLoop(p *utils.Parser) *utils.ForStatement {
	stmt := &utils.ForStatement{Pos: currentPosition(p)}
	p.Next()
	if p.Expect("(", "LITERAL") {
		p.Next()
	}

	stmt.Init = parseForPart(p)
	if p.Expect(";", "LITERAL") {
		p.Next()
	}
	stmt.Condition = ParseExpression(p, lowest)
	if p.Expect(";", "LITERAL") {
		p.Next()
	}
	stmt.Post = parseForPart(p)
	if p.Expect(")", "LITERAL") {
		p.Next()
	}
	stmt.Body = ParseBlockStatement(p)
	return stmt
}

func consumeStatementEnd(p *utils.Parser) {
	for p.Expect(";", "LITERAL") {
		p.Next()
	}
}

func parseStatement(p *utils.Parser) utils.Statement {
	if p.Current() == nil || p.Expect("EOF", "TYPE") {
		return nil
	}

	switch currentType(p) {
	case "KEYWORD":
		if currentLiteral(p) == "const" || currentLiteral(p) == "var" {
			stmt := ParseAssignment(p)
			consumeStatementEnd(p)
			return stmt
		}
	case "IDENTIFIER":
		if p.PeekExpect("ASSIGNMENT", 1, "TYPE") {
			stmt := parseReassignment(p)
			consumeStatementEnd(p)
			return stmt
		}
		if p.PeekExpect("UNARY", 1, "TYPE") {
			stmt := ParseUnary(p)
			consumeStatementEnd(p)
			return stmt
		}
	case "LOOP":
		return ParseForLoop(p)
	case "LOGIC":
		if currentLiteral(p) == "if" {
			return ParseIfStatement(p)
		}
	case "FUNCTION":
		return ParseFunction(p)
	case "TYPEDEF":
		return ParseTypeStatement(p)
	case "INTERFACE":
		return ParseInterfaceStatement(p)
	case "IMPORT":
		return ParseImportStatement(p)
	case "EXPORT":
		return ParseExportStatement(p)
	case "RETURN":
		stmt := ParseReturn(p)
		consumeStatementEnd(p)
		return stmt
	case "SEMICOLON":
		p.Next()
		return nil
	}

	expr := ParseExpression(p, lowest)
	consumeStatementEnd(p)
	if expr != nil {
		return utils.ExpressionStatement{Expression: expr}
	}
	return nil
}

func ParseARLang(parser *utils.Parser) *utils.Codebase {
	program := &utils.Codebase{Statements: []utils.Statement{}}
	for parser.Current() != nil && !parser.Expect("EOF", "TYPE") && !parser.Expect("}", "LITERAL") {
		stmt := parseStatement(parser)
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		} else if parser.Current() != nil && !parser.Expect("EOF", "TYPE") && !parser.Expect("}", "LITERAL") {
			parser.Next()
		}
	}
	program.Errors = parser.Errors
	return program
}
