package crap_test

import (
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func TestFilterByChangedFiles(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "pkg/a.go", Line: 1, Func: "A"},
		{File: "pkg/b.go", Line: 2, Func: "B"},
		{File: "pkg/c.go", Line: 3, Func: "C"},
	}

	got := crap.FilterByChangedFiles(functions, []string{"pkg/a.go", "pkg/c.go"})

	if len(got) != 2 {
		t.Fatalf("got %d functions, want 2: %+v", len(got), got)
	}
	names := map[string]bool{}
	for _, f := range got {
		names[f.Func] = true
	}
	if !names["A"] || !names["C"] {
		t.Errorf("got funcs %+v, want A and C", got)
	}
	if names["B"] {
		t.Errorf("B should have been filtered out, got %+v", got)
	}
}

func TestFilterByChangedFiles_NoMatches(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "pkg/a.go", Line: 1, Func: "A"},
	}
	got := crap.FilterByChangedFiles(functions, []string{"pkg/other.go"})
	if len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestFilterByChangedFiles_EmptyChangedList(t *testing.T) {
	functions := []crap.FunctionComplexity{
		{File: "pkg/a.go", Line: 1, Func: "A"},
	}
	got := crap.FilterByChangedFiles(functions, nil)
	if len(got) != 0 {
		t.Errorf("got %+v, want empty when no files changed", got)
	}
}

func TestFilterByChangedFiles_SuffixMatchAcrossPathBases(t *testing.T) {
	// changed files come from `git diff --name-only`, which is always
	// repo-root relative, while scanned function files may be relative
	// to a --dir that is itself relative to the repo root (or vice
	// versa). A suffix match on path segments reconciles the two.
	functions := []crap.FunctionComplexity{
		{File: "sub/pkg/a.go", Line: 1, Func: "A"},
	}
	got := crap.FilterByChangedFiles(functions, []string{"pkg/a.go"})
	if len(got) != 1 {
		t.Fatalf("got %+v, want suffix match to include A", got)
	}
}
