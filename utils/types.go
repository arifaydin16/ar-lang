package utils

type Parser struct {
	Tokens []Token
	Pos    int
	Errors []string
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		Tokens: tokens,
		Pos:    0,
		Errors: []string{},
	}
}

func (p *Parser) AddError(message string) {
	p.Errors = append(p.Errors, message)
}

func (p *Parser) GetTokens() []Token {
	return p.Tokens
}

func (p *Parser) PeekExpect(expect string, pos int, t string) bool {
	token := p.Peek(pos)
	if token == nil {
		return false
	}
	if t == "LITERAL" {
		return token.Literal == expect
	}
	return token.Type == expect
}

func (p *Parser) SetTokens(tokens []Token) {
	p.Tokens = tokens
}

func (p *Parser) Current() *Token {
	if p.Pos < 0 || p.Pos >= len(p.Tokens) {
		return nil
	}
	return &p.Tokens[p.Pos]
}

func (p *Parser) Expect(v string, t string) bool {
	token := p.Current()
	if token == nil {
		return false
	}
	if t == "LITERAL" {
		return token.Literal == v
	}
	return token.Type == v
}

func (p *Parser) Next() (int, bool) {
	if p.Pos < len(p.Tokens) {
		p.Pos++
		return p.Pos, true
	}
	return p.Pos, false
}

func (p *Parser) Peek(pos int) *Token {
	index := p.Pos + pos
	if index >= 0 && index < len(p.Tokens) {
		return &p.Tokens[index]
	}
	return nil
}

type Token struct {
	Type    string
	Literal string
	File    string
	Line    int
	Column  int
}

type Position struct {
	File   string
	Line   int
	Column int
}

type Node interface {
	TokenLiteral() string
}

type Expression interface {
	Node
	expressionNode()
}

type Statement interface {
	Node
	statementNode()
}

type Codebase struct {
	Statements []Statement
	Errors     []string
}

type BlockStatement struct {
	Statements []Statement
}

func (bs BlockStatement) statementNode()       {}
func (bs BlockStatement) TokenLiteral() string { return "block" }

type IdentifierExpression struct {
	Pos   Position
	Value string
}

type LiteralExpression struct {
	Pos   Position
	Value interface{}
}

type ValueExpression struct {
	Pos   Position
	Value string
}

type PrefixExpression struct {
	Pos      Position
	Operator string
	Right    Expression
}

type InfixExpression struct {
	Pos      Position
	Left     Expression
	Operator string
	Right    Expression
}

type ComparisonExpression struct {
	Variable   string
	Comparison string
	Compared   string
}

type LogicalExpression struct {
	Pos      Position
	Left     Expression
	Operator string
	Right    Expression
}

type CallExpression struct {
	Pos      Position
	Function string
	Args     []Expression
}

type ArrayExpression struct {
	Pos      Position
	Elements []Expression
}

type IndexExpression struct {
	Pos   Position
	Left  Expression
	Index Expression
}

type MemberExpression struct {
	Pos      Position
	Object   Expression
	Property string
}

type ObjectProperty struct {
	Key   string
	Value Expression
}

type ObjectExpression struct {
	Pos        Position
	Properties []ObjectProperty
}

func (ie IdentifierExpression) expressionNode()      {}
func (ie IdentifierExpression) TokenLiteral() string { return ie.Value }
func (le LiteralExpression) expressionNode()         {}
func (le LiteralExpression) TokenLiteral() string    { return "" }
func (ve ValueExpression) expressionNode()           {}
func (ve ValueExpression) TokenLiteral() string      { return ve.Value }
func (pe PrefixExpression) expressionNode()          {}
func (pe PrefixExpression) TokenLiteral() string     { return pe.Operator }
func (ie InfixExpression) expressionNode()           {}
func (ie InfixExpression) TokenLiteral() string      { return ie.Operator }
func (ce ComparisonExpression) expressionNode()      {}
func (ce ComparisonExpression) TokenLiteral() string { return ce.Variable }
func (le LogicalExpression) expressionNode()         {}
func (le LogicalExpression) TokenLiteral() string    { return le.Operator }
func (ce CallExpression) expressionNode()            {}
func (ce CallExpression) TokenLiteral() string       { return ce.Function }
func (ae ArrayExpression) expressionNode()           {}
func (ae ArrayExpression) TokenLiteral() string      { return "array" }
func (ie IndexExpression) expressionNode()           {}
func (ie IndexExpression) TokenLiteral() string      { return "index" }
func (me MemberExpression) expressionNode()          {}
func (me MemberExpression) TokenLiteral() string     { return me.Property }
func (oe ObjectExpression) expressionNode()          {}
func (oe ObjectExpression) TokenLiteral() string     { return "object" }

type AssignmentStatement struct {
	Pos      Position
	Variable string
	Scope    string
	Type     string
	Types    []string
	Value    Expression
}

type ReassignmentStatement struct {
	Pos      Position
	Variable string
	Operator string
	Value    Expression
}

type UnaryStatement struct {
	Pos      Position
	Variable string
	Unary    string
}

type ExpressionStatement struct {
	Pos        Position
	Expression Expression
}

type ReturnStatement struct {
	Pos   Position
	Value Expression
}

type IfStatement struct {
	Pos         Position
	Condition   Expression
	Consequence *BlockStatement
	ElseIfs     []ElseIfStatement
	Alternative *BlockStatement
}

type EnumMember struct {
	Name  string
	Value Expression
}

type TypeStatement struct {
	Pos     Position
	Name    string
	Kind    string
	Values  []Expression
	Members []EnumMember
}

type InterfaceField struct {
	Name string
	Type string
}

type InterfaceStatement struct {
	Pos    Position
	Name   string
	Fields []InterfaceField
}

type ImportSpecifier struct {
	Name  string
	Alias string
}

type ImportStatement struct {
	Pos           Position
	DefaultImport string
	NamedImports  []ImportSpecifier
	Source        string
}

type ExportStatement struct {
	Pos         Position
	Default     bool
	Names       []string
	Declaration Statement
	Value       Expression
}

type ElseIfStatement struct {
	Condition   Expression
	Consequence *BlockStatement
}

type FunctionArgument struct {
	Type string
	Name string
}

type FunctionStatement struct {
	Pos        Position
	ReturnType string
	Name       string
	Args       []FunctionArgument
	Body       *BlockStatement
}

type ForStatement struct {
	Pos       Position
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (as AssignmentStatement) statementNode()       {}
func (as AssignmentStatement) TokenLiteral() string { return as.Variable }
func (rs ReassignmentStatement) statementNode()     {}
func (rs ReassignmentStatement) TokenLiteral() string {
	return rs.Variable
}
func (us UnaryStatement) statementNode()       {}
func (us UnaryStatement) TokenLiteral() string { return us.Variable }
func (es ExpressionStatement) statementNode()  {}
func (es ExpressionStatement) TokenLiteral() string {
	if es.Expression == nil {
		return ""
	}
	return es.Expression.TokenLiteral()
}
func (rs ReturnStatement) statementNode()       {}
func (rs ReturnStatement) TokenLiteral() string { return "return" }
func (is IfStatement) statementNode()           {}
func (is IfStatement) TokenLiteral() string     { return "if" }
func (ts TypeStatement) statementNode()         {}
func (ts TypeStatement) TokenLiteral() string   { return ts.Name }
func (is InterfaceStatement) statementNode()    {}
func (is InterfaceStatement) TokenLiteral() string {
	return is.Name
}
func (is ImportStatement) statementNode()       {}
func (is ImportStatement) TokenLiteral() string { return "import" }
func (es ExportStatement) statementNode()       {}
func (es ExportStatement) TokenLiteral() string { return "export" }
func (fs FunctionStatement) statementNode()     {}
func (fs FunctionStatement) TokenLiteral() string {
	return fs.Name
}
func (fs ForStatement) statementNode()       {}
func (fs ForStatement) TokenLiteral() string { return "for" }
