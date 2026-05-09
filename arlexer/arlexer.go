package arlexer

import (
	"first-api/utils"
	"fmt"
	"strings"
)

func arlexerDefiner(define string, tokens *[]utils.Token) {
	var result string

	if len(define) > 0 && define[0] == '"' {
		result = "STRING_LITERAL"
	} else {
		switch define {
		case "const", "var", "int", "float", "bool", "string", "if", "else", "elsif", "return", "for", "while", "true", "false":
			result = "KEYWORD"
		case "+", "-", "*", "/":
			result = "ARITHMETIC"
		case "=", ":=":
			result = "ASSIGNMENT"
		case "--", "++":
			result = "UNARY"
		case "==", "!=", "<", ">", "<=", ">=":
			result = "COMPARISON"
		case "&&", "||", "!":
			result = "LOGICAL"
		case "(", ")":
			result = "PAREN"
		case "{", "}":
			result = "BRACE"
		case "[", "]":
			result = "BRACKET"
		case ",":
			result = "COMMA"
		case ";":
			result = "SEMICOLON"
		default:
			// 2. Dinamik Kontroller
			if utils.IsNumberStr(define) {
				result = "NUMBER"
			} else if utils.IsLetter(define[0]) {
				result = "IDENTIFIER"
			} else {
				result = "ILLEGAL"
			}
		}
	}

	*tokens = append(*tokens, utils.Token{
		Type:    result,
		Literal: define,
	})
}

func Arlexer(line string) []utils.Token {
	var lineLength int = len(line)
	var word []rune
	var tokens []utils.Token
	var previousChar byte
	var isInString bool = false
	for i := 0; i < lineLength; i++ {
		if line[i] == '"' {
			if isInString {
				// String bitti, tırnağı içeri alıp (veya almayıp) kapatıyoruz
				word = append(word, rune(line[i]))
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
				isInString = false
			} else {
				// Eğer daha önceden kalan bir word varsa onu gönder
				if len(word) > 0 {
					arlexerDefiner(string(word), &tokens)
					word = []rune{}
				}
				isInString = true
				word = append(word, rune(line[i]))
			}
			previousChar = line[i]
			continue // Tırnak işlemini bitir, sonraki karaktere geç
		}
		if isInString {
			word = append(word, rune(line[i]))
			previousChar = line[i]
			continue
		}
		if line[i] == ' ' && len(word) > 0 {
			arlexerDefiner(string(word), &tokens)
			word = []rune{}
		} else if utils.IsLetter(line[i]) {
			if !utils.IsLetter(previousChar) && len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
			}
			word = append(word, rune(line[i]))
			previousChar = line[i]
		} else if utils.IsNumber(line[i]) {
			if !utils.IsNumber(previousChar) && len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
			}
			word = append(word, rune(line[i]))
			previousChar = line[i]
		} else if utils.IsComparisonProbability(line[i]) {
			if !utils.IsComparisonProbability(previousChar) && len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
			}
			var nextLetter = line[i+1]
			if len(line) > i+1 {
				var mergedWord string = strings.Trim(string(line[i])+string(nextLetter), " ")
				if len(mergedWord) == 1 && utils.IsAssignment(mergedWord[0]) && !utils.IsComparisonProbability(previousChar) {
					arlexerDefiner(mergedWord, &tokens)
					word = []rune{}
				} else {
					if utils.IsComparison(mergedWord) {
						arlexerDefiner(mergedWord, &tokens)
						previousChar = line[i]
						word = []rune{}
					} else if len(word) > 0 {
						arlexerDefiner(string(word), &tokens)
					}
				}
			} else if len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
			}

		} else if utils.IsAssignment(line[i]) {
			if !utils.IsAssignment(previousChar) && len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
			}
			word = append(word, rune(line[i]))
			previousChar = line[i]
		} else if utils.IsOperator(line[i]) {
			if !utils.IsOperator(previousChar) && len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
			}
			word = append(word, rune(line[i]))
			previousChar = line[i]
		} else if utils.IsPunctuation(line[i]) {
			if len(word) > 0 {
				arlexerDefiner(string(word), &tokens)
				word = []rune{}
			}
			arlexerDefiner(string(line[i]), &tokens)
			previousChar = line[i]
		}
	}
	if len(word) > 0 {
		arlexerDefiner(string(word), &tokens)
	}
	fmt.Println(tokens)
	return tokens
}
