package ast

import (
	"os"
	"path/filepath"
	"testing"
)

func writeModule(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write module %s: %v", name, err)
	}
	return path
}

func TestResolverBuildsExportTableAndValidatesImports(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "a.go", `
const int age = 18;
const string name = "Arif";
export default age;
export { name };
`)
	entry := writeModule(t, dir, "b.go", `import age, { name } from "a.go"`)

	resolver := NewResolver(dir)
	module := resolver.Resolve(entry)
	if module == nil {
		t.Fatal("expected resolved module")
	}
	if len(resolver.Errors) != 0 {
		t.Fatalf("expected no resolver errors, got %+v", resolver.Errors)
	}

	moduleA := resolver.Modules[filepath.Clean(filepath.Join(dir, "a.go"))]
	if moduleA == nil {
		t.Fatal("expected imported module to be cached")
	}
	if moduleA.Exports.Default == nil || moduleA.Exports.Default.Name != "age" {
		t.Fatalf("unexpected default export: %+v", moduleA.Exports.Default)
	}
	if _, ok := moduleA.Exports.Named["name"]; !ok {
		t.Fatalf("expected named export name, got %+v", moduleA.Exports.Named)
	}
}

func TestResolverReportsMissingNamedExport(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "a.go", `
const int age = 18;
export default age;
`)
	entry := writeModule(t, dir, "b.go", `import age, { name } from "a.go"`)

	resolver := NewResolver(dir)
	resolver.Resolve(entry)
	if len(resolver.Errors) == 0 {
		t.Fatal("expected missing export error")
	}
	if resolver.Errors[0].Code != "ARLANG_MODULE_ERR_006" {
		t.Fatalf("expected ARLANG_MODULE_ERR_006, got %+v", resolver.Errors)
	}
	if resolver.Errors[0].File == "" || resolver.Errors[0].Line == 0 || resolver.Errors[0].Column == 0 {
		t.Fatalf("expected resolver error location, got %+v", resolver.Errors[0])
	}
}

func TestResolverReportsMissingDefaultExport(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "a.go", `
const string name = "Arif";
export { name };
`)
	entry := writeModule(t, dir, "b.go", `import age, { name } from "a.go"`)

	resolver := NewResolver(dir)
	resolver.Resolve(entry)
	if len(resolver.Errors) == 0 {
		t.Fatal("expected missing default export error")
	}
	if resolver.Errors[0].Code != "ARLANG_MODULE_ERR_005" {
		t.Fatalf("expected ARLANG_MODULE_ERR_005, got %+v", resolver.Errors)
	}
}
