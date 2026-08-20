package gosnippet

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeAddsBoilerplate(t *testing.T) {
	// The shape almost every article uses: a bare main with no package
	// clause and no imports.
	code := "func main() {\n\tfmt.Println(\"hi\")\n}\n"
	p, err := Normalize(code, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !p.AddedPkg || !p.HasMain || p.AddedMain {
		t.Errorf("flags = %+v", p)
	}
	for _, want := range []string{"package main", `"fmt"`, `fmt.Println("hi")`} {
		if !strings.Contains(p.Source, want) {
			t.Errorf("source is missing %q:\n%s", want, p.Source)
		}
	}
}

func TestNormalizeAddsMain(t *testing.T) {
	// A snippet that only declares things still has to be a runnable
	// program to compile and to be shareable on the playground.
	p, err := Normalize("func f[T comparable](T) {}\n", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !p.AddedMain || !p.HasMain {
		t.Errorf("flags = %+v", p)
	}
	if !strings.Contains(p.Source, "func main() {}") {
		t.Errorf("no main added:\n%s", p.Source)
	}
}

func TestNormalizeKeepsExistingPackage(t *testing.T) {
	p, err := Normalize("package sample\n\nvar X = 1\n", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if p.AddedPkg || p.AddedMain || p.Package != "sample" {
		t.Errorf("flags = %+v", p)
	}
}

func TestNormalizeForcedImports(t *testing.T) {
	// goimports would pick crypto/rand here; the directive settles it.
	code := "func main() {\n\t_ = rand.Intn(2)\n}\n"
	p, err := Normalize(code, Options{Imports: []string{"math/rand"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.Source, `"math/rand"`) {
		t.Errorf("math/rand not imported:\n%s", p.Source)
	}
}

func TestNormalizeReportsSnippetPositions(t *testing.T) {
	// Line 2 of the snippet is bad; the wrapper line must not shift it.
	p, err := Normalize("var x int\n~int | ~string\n", Options{})
	if err == nil {
		t.Fatalf("expected an error, got:\n%s", p.Source)
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("error is %T, want *ParseError", err)
	}
	if len(pe.Problems) == 0 || pe.Problems[0].Line != 2 {
		t.Errorf("problems = %+v, want the first one on line 2", pe.Problems)
	}
}

func TestHashIsStableAndEmptyForFailures(t *testing.T) {
	a, err := Normalize("func main() {}\n", Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Normalize("func main() {}", Options{Filename: "other.go"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Hash() != b.Hash() {
		t.Errorf("hashes differ: %s vs %s", a.Hash(), b.Hash())
	}
	if (Program{}).Hash() != "" {
		t.Errorf("empty program should hash to the empty string")
	}
}
