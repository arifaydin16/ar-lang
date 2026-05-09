package main

import (
	"bufio"
	"first-api/arlexer"
	"first-api/ast"
	"first-api/utils"
	"fmt"
	"os"
)

func main() {
	var file, err = os.Open("./ar-lang/index.ar")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var tokens []utils.Token = []utils.Token{}
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		println(line)
		var lineTokens = arlexer.Arlexer(line)
		tokens = append(tokens, lineTokens...)
	}
	fmt.Println("------------------------------")
	fmt.Printf("%+v\n", tokens)
	var astezied = ast.ParseARLang(tokens)
	fmt.Println("------------------------------")
	fmt.Printf("%+v\n", astezied)

}
