package utils

import "unicode"

func IsLetter(let byte) bool {
	return (let >= 'a' && let <= 'z') || (let >= 'A' && let <= 'Z') || let == '_'
}

func IsNumber(let byte) bool {
	return (let >= '0' && let <= '9') || let == '.'
}

func IsOperator(let byte) bool {
	return let == '+' || let == '-' || let == '*' || let == '/' || let == '%'
}

func isLogicalOperator(let byte) bool {
	return let == '&' || let == '|'
}

func IsAssignment(let byte) bool {
	return let == '='
}
func IsComparison(wrd string) bool {
	return wrd == "==" || wrd == "<=" || wrd == ">=" || wrd == "!="
}
func HasComparisonProbability(let byte) bool {
	return let == '=' || let == '<' || let == '>'
}
func ContainsNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func IsNumberStr(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			return false
		}
	}
	return true
}

func IsPunctuation(ch byte) bool {
	return ch == '(' || ch == ')' || ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ',' || ch == ';'
}
