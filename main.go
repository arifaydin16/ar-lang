package main

import (
	"bufio"
	"encoding/json"
	"first-api/arlexer"
	"first-api/ast"
	"first-api/utils"
	"fmt"
	"os"
)

func main() {
	sourcePath := "./ar-lang/index.ar"
	var file, err = os.Open(sourcePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var tokens []utils.Token = []utils.Token{}
	var parser = utils.NewParser(tokens)
	var scanner = bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		var line = scanner.Text()
		println(line)
		var lineTokens = arlexer.ArlexerWithLineAndFile(line, sourcePath, lineNumber)
		parser.SetTokens(append(parser.Tokens, lineTokens...))
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	parser.SetTokens(append(parser.Tokens, arlexer.EOFTokenWithFile(sourcePath, lineNumber)))
	fmt.Println("------------------------------ arlexer.go")
	fmt.Printf("%+v\n", parser.Tokens)
	var astezied = ast.ParseARLang(parser)
	fmt.Println("------------------------------ ast.go")
	b, err := json.MarshalIndent(astezied, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
	resolver := ast.NewResolver(".")
	resolver.Resolve(sourcePath)
	if len(resolver.Errors) > 0 {
		fmt.Println("------------------------------ semantic.go")
		for _, resolverError := range resolver.Errors {
			fmt.Println(resolverError.Error())
		}
	}

}
