package ast

import (
	"first-api/utils"
	"fmt"
	"strconv"
)

func ParseAssignment(tokens []utils.Token, token utils.Token, i int) utils.AssignmentStatement {
	var assignmentValue interface{} = tokens[i+2].Literal
	assignmentType := tokens[i-1].Literal
	var err error

	// Tip dönüşümlerini güvenli hale getirelim
	switch assignmentType {
	case "int":
		// Literal her zaman string gelir, direkt parse edebiliriz
		assignmentValue, err = strconv.ParseInt(tokens[i+2].Literal, 10, 64)
	case "float":
		assignmentValue, err = strconv.ParseFloat(tokens[i+2].Literal, 64)
	case "bool":
		assignmentValue, err = strconv.ParseBool(tokens[i+2].Literal)
	}

	if err != nil {
		fmt.Printf("Conversation error: in %s line %s excepted.\n", string(i+2), assignmentType)
	}

	var ass = utils.AssignmentStatement{
		Variable: token.Literal,
		Scope:    tokens[i-2].Literal,
		Type:     tokens[i-1].Literal,
		Value:    assignmentValue,
	}
	return ass
}

func ParseComparison(tokens []utils.Token, i int) utils.ComparisonExpression {
	var token = tokens[i]
	var com = utils.ComparisonExpression{}
	if token.Literal == "IDENTIFIER" && tokens[i+1].Literal == "COMPARISON" {
		com.Variable = token.Literal
		com.Comparison = tokens[i+1].Type
		com.Compared = tokens[i+2].Literal
	}
	return com
}
func ParseUnary(tokens []utils.Token, i int) utils.UnaryStatement {
	var token = tokens[i]
	var un = utils.UnaryStatement{}
	if token.Literal == "UNARY" && tokens[i-1].Literal == "IDENTIFIER" {
		un.Variable = tokens[i-1].Literal
		un.Unary = token.Literal
	}
	return un
}

func ParseARLang(tokens []utils.Token) *utils.Codebase {
	var program = &utils.Codebase{}
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token.Type == "IDENTIFIER" {
			if i+1 < len(tokens) && tokens[i+1].Literal == "=" {
				var assignment = ParseAssignment(tokens, token, i)
				program.Statements = append(program.Statements, assignment)

				i += 2
			}
		} else if token.Type == "KEYWORD" {
			switch token.Literal {
			case "for":
				var ForStatement = &utils.ForStatement{}
				i++
				if token.Type == "PAREN" && token.Literal == "{" {
					i++
				}
				for tokens[i].Literal != ";" {
					i++
				}

				var assStatement = ParseAssignment(tokens, token, i)
				ForStatement.Init = assStatement
				for tokens[i].Literal != ";" {
					i++
				}
				var comStatement = ParseComparison(tokens, i)
				ForStatement.Condition = comStatement
				for tokens[i].Literal != ";" {
					i++
				}
				var unaryStatement = ParseUnary(tokens, i+1)
				ForStatement.Post = unaryStatement

			}
		}
	}
	return program
}
