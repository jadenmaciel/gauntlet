package crap

import (
	"path/filepath"
	"strings"
)

// FilterByChangedFiles keeps only the functions whose File is among
// changedFiles. Matching allows a path-segment suffix match (not just
// exact equality) because `git diff --name-only` always returns
// repo-root-relative paths, while scanned files may be relative to a
// --dir that sits below the repo root, or vice versa.
func FilterByChangedFiles(functions []FunctionComplexity, changedFiles []string) []FunctionComplexity {
	result := make([]FunctionComplexity, 0, len(functions))
	if len(changedFiles) == 0 {
		return result
	}
	for _, fn := range functions {
		for _, cf := range changedFiles {
			if filesMatch(fn.File, cf) {
				result = append(result, fn)
				break
			}
		}
	}
	return result
}

func filesMatch(a, b string) bool {
	a = filepath.ToSlash(a)
	b = filepath.ToSlash(b)
	if a == b {
		return true
	}
	return strings.HasSuffix(a, "/"+b) || strings.HasSuffix(b, "/"+a)
}
