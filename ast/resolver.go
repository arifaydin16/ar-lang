package ast

import (
	"bufio"
	"first-api/arlexer"
	"first-api/utils"
	"fmt"
	"os"
	"path/filepath"
)

type ExportSymbol struct {
	Name  string
	Kind  string
	Types []string
}

type ExportTable struct {
	Source  string
	Default *ExportSymbol
	Named   map[string]ExportSymbol
}

type Module struct {
	Path     string
	Program  *utils.Codebase
	Analyzer *Analyzer
	Exports  ExportTable
}

type Resolver struct {
	Root    string
	Modules map[string]*Module
	Errors  []SemanticError
	loading map[string]bool
}

func NewResolver(root string) *Resolver {
	return &Resolver{
		Root:    root,
		Modules: map[string]*Module{},
		Errors:  []SemanticError{},
		loading: map[string]bool{},
	}
}

func ResolveModule(entry string) (*Module, []SemanticError) {
	resolver := NewResolver(filepath.Dir(entry))
	module := resolver.Resolve(entry)
	return module, resolver.Errors
}

func (r *Resolver) Resolve(path string) *Module {
	resolvedPath := r.resolvePath(path, "")
	if module, ok := r.Modules[resolvedPath]; ok {
		return module
	}
	if r.loading[resolvedPath] {
		r.addResolverError("ARLANG_MODULE_ERR_004", utils.Position{File: resolvedPath}, "circular import detected for %q", resolvedPath)
		return nil
	}

	r.loading[resolvedPath] = true
	program, parseErrors := ParseFile(resolvedPath)
	if len(parseErrors) > 0 {
		r.Errors = append(r.Errors, parseErrors...)
		delete(r.loading, resolvedPath)
		return nil
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)
	r.Errors = append(r.Errors, analyzer.Errors...)

	module := &Module{
		Path:     resolvedPath,
		Program:  program,
		Analyzer: analyzer,
		Exports:  BuildExportTable(resolvedPath, program, analyzer),
	}
	r.Modules[resolvedPath] = module

	r.resolveImports(module)
	delete(r.loading, resolvedPath)
	return module
}

func ParseFile(path string) (*utils.Codebase, []SemanticError) {
	file, err := os.Open(path)
	if err != nil {
		return nil, []SemanticError{{
			Code:    "ARLANG_MODULE_ERR_001",
			Message: "cannot open module file " + path + ": " + err.Error(),
			File:    path,
		}}
	}
	defer file.Close()

	tokens := []utils.Token{}
	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		tokens = append(tokens, arlexer.ArlexerWithLineAndFile(scanner.Text(), path, lineNumber)...)
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		return nil, []SemanticError{{
			Code:    "ARLANG_MODULE_ERR_002",
			Message: "cannot read module file " + path + ": " + err.Error(),
			File:    path,
		}}
	}
	tokens = append(tokens, arlexer.EOFTokenWithFile(path, lineNumber))
	return ParseARLang(utils.NewParser(tokens)), nil
}

func BuildExportTable(path string, program *utils.Codebase, analyzer *Analyzer) ExportTable {
	table := ExportTable{
		Source: path,
		Named:  map[string]ExportSymbol{},
	}
	if program == nil {
		return table
	}

	for _, stmt := range program.Statements {
		exportStmt, ok := stmt.(*utils.ExportStatement)
		if !ok {
			continue
		}

		if exportStmt.Declaration != nil {
			symbol := exportSymbolFromStatement(exportStmt.Declaration, analyzer)
			if symbol.Name != "" {
				if exportStmt.Default {
					table.Default = &symbol
				} else {
					table.Named[symbol.Name] = symbol
				}
			}
			continue
		}

		for _, name := range exportStmt.Names {
			if symbol, ok := exportSymbolByName(name, analyzer); ok {
				table.Named[name] = symbol
			}
		}

		if exportStmt.Default && exportStmt.Value != nil {
			if identifier, ok := exportStmt.Value.(utils.IdentifierExpression); ok {
				if symbol, ok := exportSymbolByName(identifier.Value, analyzer); ok {
					table.Default = &symbol
				}
			} else {
				table.Default = &ExportSymbol{Name: "default", Kind: "expression", Types: []string{analyzer.inferExpression(exportStmt.Value)}}
			}
		}
	}

	return table
}

func exportSymbolFromStatement(stmt utils.Statement, analyzer *Analyzer) ExportSymbol {
	switch node := stmt.(type) {
	case utils.AssignmentStatement:
		return ExportSymbol{Name: node.Variable, Kind: "variable", Types: node.Types}
	case *utils.FunctionStatement:
		return ExportSymbol{Name: node.Name, Kind: "function", Types: typeList(node.ReturnType)}
	case *utils.TypeStatement:
		return ExportSymbol{Name: node.Name, Kind: "type", Types: []string{node.Name}}
	case *utils.InterfaceStatement:
		return ExportSymbol{Name: node.Name, Kind: "interface", Types: []string{node.Name}}
	}
	return ExportSymbol{}
}

func exportSymbolByName(name string, analyzer *Analyzer) (ExportSymbol, bool) {
	if symbol, ok := analyzer.Symbols[name]; ok {
		return ExportSymbol{Name: name, Kind: "variable", Types: symbol.Types}, true
	}
	if fn, ok := analyzer.Functions[name]; ok {
		return ExportSymbol{Name: name, Kind: "function", Types: typeList(fn.ReturnType)}, true
	}
	if typeInfo, ok := analyzer.Types[name]; ok && typeInfo.Kind != "primitive" {
		return ExportSymbol{Name: name, Kind: typeInfo.Kind, Types: []string{name}}, true
	}
	return ExportSymbol{}, false
}

func (r *Resolver) resolveImports(module *Module) {
	for _, stmt := range module.Program.Statements {
		importStmt, ok := stmt.(*utils.ImportStatement)
		if !ok {
			continue
		}

		importedPath := r.resolvePath(importStmt.Source, filepath.Dir(module.Path))
		importedModule := r.Resolve(importedPath)
		if importedModule == nil {
			r.addResolverError("ARLANG_MODULE_ERR_003", importStmt.Pos, "cannot resolve import %q", importStmt.Source)
			continue
		}

		if importStmt.DefaultImport != "" && importedModule.Exports.Default == nil {
			r.addResolverError("ARLANG_MODULE_ERR_005", importStmt.Pos, "module %q has no default export", importStmt.Source)
		}

		for _, namedImport := range importStmt.NamedImports {
			if _, ok := importedModule.Exports.Named[namedImport.Name]; !ok {
				r.addResolverError("ARLANG_MODULE_ERR_006", importStmt.Pos, "module %q has no named export %q", importStmt.Source, namedImport.Name)
			}
		}
	}
}

func (r *Resolver) resolvePath(path string, baseDir string) string {
	if filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			return filepath.Clean(abs)
		}
		return filepath.Clean(path)
	}
	if baseDir == "" {
		baseDir = r.Root
	}
	if baseDir == "" {
		baseDir = "."
	}
	abs, err := filepath.Abs(filepath.Join(baseDir, path))
	if err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}

func (r *Resolver) addResolverError(code string, pos utils.Position, format string, args ...interface{}) {
	err := SemanticError{
		Code:    code,
		Message: formatMessage(format, args...),
		File:    pos.File,
		Line:    pos.Line,
		Column:  pos.Column,
	}
	r.Errors = append(r.Errors, err)
}

func formatMessage(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
