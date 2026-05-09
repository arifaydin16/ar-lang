package ast

import (
	"first-api/utils"
	"fmt"
	"strconv"
)

func ParseARLang(tokens []utils.Token) *utils.Codebase {
	var program = &utils.Codebase{}
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token.Type == "IDENTIFIER" {
			if i+1 < len(tokens) && tokens[i+1].Literal == "=" {
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
					fmt.Printf("Conversation error: in %v line %s excepted.\n", tokens[i+2].Line, assignmentType)
				}

				assignment := utils.AssignmentStatement{
					Variable: token.Literal,
					Scope:    tokens[i-2].Literal,
					Type:     tokens[i-1].Literal,
					Value:    assignmentValue,
				}

				program.Statements = append(program.Statements, assignment)

				// Atama satırını tükettik
				i += 2
			}
		}
	}
	return program
}
