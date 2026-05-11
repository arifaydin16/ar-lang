package utils

type Parser struct {
	tokens []Token
	pos    int
}

type Token struct {
	Type    string
	Literal string
}

type Possibility struct {
	scope    bool
	var_name bool
	operator bool
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

type BlockStatement struct {
	Statements []Statement
}

type ComparisonExpression struct {
	Variable   string
	Comparison string
	Compared   string
}

type UnaryStatement struct {
	Variable string
	Unary    string
}

func (ce ComparisonExpression) expressionNode()      {}
func (ce ComparisonExpression) TokenLiteral() string { return ce.Variable }

type AssignmentStatement struct {
	Variable string
	Scope    string
	Type     string
	Value    interface{}
}

type ForStatement struct {
	Init      AssignmentStatement
	Condition Expression
	Post      Statement
	Body      *[]BlockStatement
}

func (bs BlockStatement) statementNode()            {}
func (bs ForStatement) statementNode()              {}
func (bs AssignmentStatement) statementNode()       {}
func (bs AssignmentStatement) TokenLiteral() string { return bs.Variable }
func (bs UnaryStatement) statementNode()            {}
func (bs UnaryStatement) TokenLiteral() string      { return bs.Variable }

type Codebase struct {
	Statements []Statement
}
