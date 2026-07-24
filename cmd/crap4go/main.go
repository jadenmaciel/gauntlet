// Command crap4go computes CRAP (Change Risk Anti-Patterns) scores for
// Go functions by combining cyclomatic complexity with test coverage,
// and fails the build when any function exceeds a configurable ceiling.
package main

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

var generatedHeaderRE = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

// cliFunctionReport and cliReport mirror the exact JSON schema required
// by the CLI; they intentionally drop internal-only fields (like
// FunctionReport.Matched) that aren't part of that schema.
type cliFunctionReport struct {
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Func       string  `json:"func"`
	Complexity int     `json:"complexity"`
	Coverage   float64 `json:"coverage"`
	CRAP       float64 `json:"crap"`
	Pass       bool    `json:"pass"`
}

type cliSummary struct {
	Total   int     `json:"total"`
	Failing int     `json:"failing"`
	MaxCRAP float64 `json:"max_crap"`
}

type cliReport struct {
	Ceiling   float64             `json:"ceiling"`
	Functions []cliFunctionReport `json:"functions"`
	Summary   cliSummary          `json:"summary"`
}

type cliOptions struct {
	profile string
	dir     string
	ceiling float64
	changed string
	format  string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		return 2
	}
	functions, err := scanDir(options.dir)
	if err != nil {
		fmt.Fprintf(stderr, "crap4go: %v\n", err)
		return 2
	}
	functions, err = filterChangedFunctions(options.dir, options.changed, functions)
	if err != nil {
		fmt.Fprintf(stderr, "crap4go: %v\n", err)
		return 2
	}
	coverage, err := coverageFor(options.profile)
	if err != nil {
		fmt.Fprintf(stderr, "crap4go: %v\n", err)
		return 2
	}
	sortFunctions(functions)
	report := crap.Evaluate(functions, coverage, options.ceiling)
	printReport(stdout, options.format, report)
	return reportExitCode(report)
}

func parseOptions(args []string, stderr *os.File) (cliOptions, error) {
	fs := flag.NewFlagSet("crap4go", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profile := fs.String("profile", "", "path to a go coverage profile (fed through `go tool cover -func`)")
	dir := fs.String("dir", ".", "directory to scan for .go source files")
	ceiling := fs.Float64("ceiling", 8, "CRAP ceiling; a function passes when CRAP <= ceiling")
	changed := fs.String("changed", "", "git ref; when set, scope to files changed since this ref (via git diff --name-only)")
	format := fs.String("format", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(stderr, "crap4go: invalid --format %q, want \"text\" or \"json\"\n", *format)
		return cliOptions{}, fmt.Errorf("invalid format")
	}
	return cliOptions{profile: *profile, dir: *dir, ceiling: *ceiling, changed: *changed, format: *format}, nil
}

func filterChangedFunctions(dir, changed string, functions []crap.FunctionComplexity) ([]crap.FunctionComplexity, error) {
	if changed == "" {
		return functions, nil
	}
	files, err := gitChangedFiles(dir, changed)
	if err != nil {
		return nil, err
	}
	return crap.FilterByChangedFiles(functions, files), nil
}

func coverageFor(profile string) (map[string]float64, error) {
	if profile == "" {
		return map[string]float64{}, nil
	}
	return coverageFromProfile(profile)
}

func sortFunctions(functions []crap.FunctionComplexity) {
	slices.SortFunc(functions, func(a, b crap.FunctionComplexity) int {
		return cmp.Or(
			cmp.Compare(a.File, b.File),
			cmp.Compare(a.Line, b.Line),
			cmp.Compare(a.Func, b.Func),
		)
	})
}

func printReport(stdout *os.File, format string, report crap.Report) {
	if format == "json" {
		printJSON(stdout, report)
		return
	}
	printText(stdout, report)
}

func reportExitCode(report crap.Report) int {
	if report.Summary.Failing > 0 {
		return 1
	}
	return 0
}

// scanDir walks dir for .go source files (skipping _test.go, vendor
// directories, and generated files) and returns their function
// complexities with File rewritten to the module-import-style path
// (e.g. "github.com/x/y/pkg/file.go") that `go tool cover -func` uses,
// so Evaluate's file:line:func join lines up.
func scanDir(dir string) ([]crap.FunctionComplexity, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving scan directory %q: %w", dir, err)
	}

	modulePath, moduleDir, err := goModule(absDir)
	if err != nil {
		return nil, err
	}

	var results []crap.FunctionComplexity
	walkErr := filepath.Walk(absDir, func(p string, info os.FileInfo, err error) error {
		funcs, err := scanPath(modulePath, moduleDir, p, info, err)
		results = append(results, funcs...)
		return err
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return results, nil
}

func scanPath(modulePath, moduleDir, p string, info os.FileInfo, walkErr error) ([]crap.FunctionComplexity, error) {
	if walkErr != nil {
		return nil, walkErr
	}
	if skipPath(p, info) {
		return nil, nil
	}
	return scanSourcePath(modulePath, moduleDir, p)
}

func scanSourcePath(modulePath, moduleDir, p string) ([]crap.FunctionComplexity, error) {
	src, err := sourceForPath(p)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, nil
	}
	funcs, err := crap.ComplexityOfFile(src)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", p, err)
	}
	importPath, err := importStylePath(modulePath, moduleDir, p)
	if err != nil {
		return nil, err
	}
	for i := range funcs {
		funcs[i].File = importPath
	}
	return funcs, nil
}

func sourceForPath(p string) ([]byte, error) {
	src, err := os.ReadFile(p)
	if err != nil || generatedHeaderRE.Match(src) {
		return nil, err
	}
	return src, nil
}

func skipPath(p string, info os.FileInfo) bool {
	return info.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") || pathHasSegment(p, "vendor")
}

func pathHasSegment(p, segment string) bool {
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		if part == segment {
			return true
		}
	}
	return false
}

func importStylePath(modulePath, moduleDir, filePath string) (string, error) {
	rel, err := filepath.Rel(moduleDir, filePath)
	if err != nil {
		return "", fmt.Errorf("computing module-relative path for %s: %w", filePath, err)
	}
	return path.Join(modulePath, filepath.ToSlash(rel)), nil
}

// goModule returns the module path and module root directory for the
// module containing dir, via `go list -m`.
func goModule(dir string) (modulePath, moduleDir string, err error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Path}} {{.Dir}}")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("go list -m in %s: %w", dir, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		return "", "", fmt.Errorf("unexpected `go list -m` output: %q", out)
	}
	return fields[0], fields[1], nil
}

// coverageFromProfile runs `go tool cover -func=<profile>` and parses
// its stdout.
func coverageFromProfile(profile string) (map[string]float64, error) {
	cmd := exec.Command("go", "tool", "cover", "-func="+profile)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go tool cover -func=%s: %w", profile, err)
	}
	coverage, err := crap.ParseCoverFunc(string(out))
	if err != nil {
		return nil, fmt.Errorf("parsing coverage output: %w", err)
	}
	return coverage, nil
}

// gitChangedFiles runs `git diff --name-only <merge-base>` in dir so the
// changed scope starts at the merge base of ref and HEAD and includes local
// worktree changes.
func gitChangedFiles(dir, ref string) ([]string, error) {
	mergeBase := exec.Command("git", "merge-base", ref, "HEAD")
	mergeBase.Dir = dir
	base, err := mergeBase.Output()
	if err != nil {
		return nil, fmt.Errorf("git merge-base %s HEAD: %w", ref, err)
	}

	cmd := exec.Command("git", "diff", "--name-only", strings.TrimSpace(string(base)))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only from merge base with %s: %w", ref, err)
	}
	untracked := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untracked.Dir = dir
	untrackedOut, err := untracked.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files --others: %w", err)
	}
	var files []string
	for _, line := range strings.Split(string(out)+string(untrackedOut), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func printJSON(stdout *os.File, report crap.Report) {
	out := cliReport{
		Ceiling:   report.Ceiling,
		Functions: make([]cliFunctionReport, 0, len(report.Functions)),
		Summary: cliSummary{
			Total:   report.Summary.Total,
			Failing: report.Summary.Failing,
			MaxCRAP: report.Summary.MaxCRAP,
		},
	}
	for _, fr := range report.Functions {
		out.Functions = append(out.Functions, cliFunctionReport{
			File:       fr.File,
			Line:       fr.Line,
			Func:       fr.Func,
			Complexity: fr.Complexity,
			Coverage:   fr.Coverage,
			CRAP:       fr.CRAP,
			Pass:       fr.Pass,
		})
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func printText(stdout *os.File, report crap.Report) {
	for _, fr := range report.Functions {
		status := "PASS"
		if !fr.Pass {
			status = "FAIL"
		}
		note := ""
		if !fr.Matched {
			note = " (no coverage data)"
		}
		fmt.Fprintf(stdout, "%s:%d:%s\tcomplexity=%d\tcoverage=%.1f%%\tcrap=%.2f\t%s%s\n",
			fr.File, fr.Line, fr.Func, fr.Complexity, fr.Coverage, fr.CRAP, status, note)
	}
	fmt.Fprintf(stdout, "\nceiling=%.2f total=%d failing=%d max_crap=%.2f\n",
		report.Ceiling, report.Summary.Total, report.Summary.Failing, report.Summary.MaxCRAP)
}
