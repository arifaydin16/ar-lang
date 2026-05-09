package main

var LETTERS_ARRAY = []rune("abcdefghijklmnopqrstuvwxyz")
var NUMS = []rune("1234567890")
var SINGLE_CHAR_OPS = map[rune]string{
	'+': "PLUS",
	'-': "MINUS",
	'*': "ASTERISK",
	'/': "SLASH",
	'=': "ASSIGN",
	'<': "LT",
	'>': "GT",
}

var MULTI_CHAR_OPS = map[string]string{
	"==": "EQ",
	"!=": "NOT_EQ",
	"<=": "LT_EQ",
	">=": "GT_EQ",
	"&&": "AND",
	"||": "OR",
}
