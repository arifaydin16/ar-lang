package arlexer

import "first-api/utils"

func defineToken(literal string) string {
	switch literal {
	case "const", "var", "int", "float", "bool", "string", "object", "any", "true", "false":
		return "KEYWORD"
	case "func":
		return "FUNCTION"
	case "return":
		return "RETURN"
	case "type":
		return "TYPEDEF"
	case "interface":
		return "INTERFACE"
	case "import":
		return "IMPORT"
	case "export":
		return "EXPORT"
	case "default":
		return "DEFAULT"
	case "from":
		return "FROM"
	case "as":
		return "AS"
	case "enum":
		return "ENUM"
	case "if", "else", "elsif":
		return "LOGIC"
	case "for":
		return "LOOP"
	case "+", "-", "*", "/", "%":
		return "ARITHMETIC"
	case "=", ":=", "+=", "-=", "*=", "/=":
		return "ASSIGNMENT"
	case "--", "++":
		return "UNARY"
	case "==", "!=", "<", ">", "<=", ">=":
		return "COMPARISON"
	case "&", "|", "!":
		return "LOGICAL"
	case "(", ")":
		return "PAREN"
	case "{", "}":
		return "BRACE"
	case "[", "]":
		return "BRACKET"
	case ",":
		return "COMMA"
	case ";":
		return "SEMICOLON"
	case ":":
		return "COLON"
	case ".":
		return "DOT"
	}

	if len(literal) > 0 && literal[0] == '"' {
		return "STRING_LITERAL"
	}
	if utils.IsNumberStr(literal) || (len(literal) > 1 && literal[0] == '-' && utils.IsNumberStr(literal[1:])) {
		return "NUMBER"
	}
	if len(literal) > 0 && utils.IsLetter(literal[0]) {
		return "IDENTIFIER"
	}
	return "ILLEGAL"
}

func appendToken(tokens *[]utils.Token, literal string, file string, line int, column int) {
	if literal == "" {
		return
	}
	*tokens = append(*tokens, utils.Token{
		Type:    defineToken(literal),
		Literal: literal,
		File:    file,
		Line:    line,
		Column:  column,
	})
}

func isIdentifierChar(ch byte) bool {
	return utils.IsLetter(ch) || (ch >= '0' && ch <= '9')
}

func canStartNegativeNumber(tokens []utils.Token) bool {
	if len(tokens) == 0 {
		return true
	}
	last := tokens[len(tokens)-1]
	return last.Type == "ASSIGNMENT" || last.Type == "COMPARISON" || last.Type == "LOGICAL" || last.Literal == "(" || last.Literal == "[" || last.Literal == "," || last.Literal == "return"
}

func Arlexer(line string) []utils.Token {
	return ArlexerWithLine(line, 0)
}

func ArlexerWithLine(line string, lineNumber int) []utils.Token {
	return ArlexerWithLineAndFile(line, "", lineNumber)
}

func ArlexerWithLineAndFile(line string, file string, lineNumber int) []utils.Token {
	var tokens []utils.Token

	for i := 0; i < len(line); {
		ch := line[i]
		column := i + 1

		if ch == ' ' || ch == '\t' || ch == '\r' {
			i++
			continue
		}

		if ch == '/' && i+1 < len(line) && line[i+1] == '/' {
			break
		}
		if ch == '/' && i+1 < len(line) && line[i+1] == '*' {
			i += 2
			for i+1 < len(line) && !(line[i] == '*' && line[i+1] == '/') {
				i++
			}
			if i+1 < len(line) {
				i += 2
			}
			continue
		}

		if utils.IsLetter(ch) {
			start := i
			for i < len(line) && isIdentifierChar(line[i]) {
				i++
			}
			appendToken(&tokens, line[start:i], file, lineNumber, column)
			continue
		}

		if utils.IsNumber(ch) || (ch == '-' && i+1 < len(line) && utils.IsNumber(line[i+1]) && canStartNegativeNumber(tokens)) {
			start := i
			if ch == '-' {
				i++
			}
			for i < len(line) && utils.IsNumber(line[i]) {
				i++
			}
			appendToken(&tokens, line[start:i], file, lineNumber, column)
			continue
		}

		if ch == '"' {
			start := i
			i++
			for i < len(line) {
				if line[i] == '\\' && i+1 < len(line) {
					i += 2
					continue
				}
				if line[i] == '"' {
					i++
					break
				}
				i++
			}
			appendToken(&tokens, line[start:i], file, lineNumber, column)
			continue
		}

		if i+1 < len(line) {
			two := line[i : i+2]
			switch two {
			case "==", "!=", "<=", ">=", ":=", "++", "--", "+=", "-=", "*=", "/=":
				appendToken(&tokens, two, file, lineNumber, column)
				i += 2
				continue
			}
		}

		switch ch {
		case '+', '-', '*', '/', '%', '=', '<', '>', '&', '|', '!', '(', ')', '{', '}', '[', ']', ',', ';', ':', '.':
			appendToken(&tokens, string(ch), file, lineNumber, column)
			i++
		default:
			appendToken(&tokens, string(ch), file, lineNumber, column)
			i++
		}
	}

	return tokens
}

func EOFToken(lineNumber int) utils.Token {
	return EOFTokenWithFile("", lineNumber)
}

func EOFTokenWithFile(file string, lineNumber int) utils.Token {
	return utils.Token{
		Type:    "EOF",
		Literal: "",
		File:    file,
		Line:    lineNumber,
		Column:  0,
	}
}
