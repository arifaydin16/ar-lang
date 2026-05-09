package utils

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

type AssignmentStatement struct {
	Variable string
	Scope    string
	Type     string
	Value    string
}

type Codebase struct {
	Statements []AssignmentStatement
}
