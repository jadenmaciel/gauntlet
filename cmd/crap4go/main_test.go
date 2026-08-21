package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func TestScanDirWithRelativePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/scan\n\ngo 1.26.4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scan.go"), []byte("package scan\n\nfunc F() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Dir(dir)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	functions, err := scanDir(filepath.Base(dir))
	if err != nil {
		t.Fatalf("scanDir() error = %v", err)
	}
	if len(functions) != 1 {
		t.Fatalf("scanDir() returned %d functions, want 1", len(functions))
	}
	if functions[0].File != "example.com/scan/scan.go" {
		t.Errorf("File = %q, want example.com/scan/scan.go", functions[0].File)
	}
}

func TestGitChangedFilesIncludesUntrackedFiles(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "tracked.go"), []byte("package p\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "tracked.go")
	runGit(t, dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
	if err := os.WriteFile(filepath.Join(dir, "new.go"), []byte("package p\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	files, err := gitChangedFiles(dir, "HEAD")
	if err != nil {
		t.Fatalf("gitChangedFiles() error = %v", err)
	}
	if len(files) != 1 || files[0] != "new.go" {
		t.Fatalf("gitChangedFiles() = %v, want [new.go]", files)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Ignore the developer's global and system git config. A global
	// core.hooksPath can drop generated files into the fixture on commit,
	// which then show up as untracked changes and skew the assertions.
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestRunRejectsInvalidFormat(t *testing.T) {
	if code := runCrap(t, []string{"--format", "xml"}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunRejectsInvalidFlag(t *testing.T) {
	if code := runCrap(t, []string{"--unknown"}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunRejectsMissingCoverageProfile(t *testing.T) {
	dir := writeScanModule(t)
	args := []string{"--dir", dir, "--ceiling", "10", "--profile", filepath.Join(t.TempDir(), "missing.out")}
	if code := runCrap(t, args); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunScansModule(t *testing.T) {
	dir := writeScanModule(t)
	args := []string{"--dir", dir, "--coverage", filepath.Join(dir, "coverage.out"), "--ceiling", "10", "--format", "json"}
	if code := runCrap(t, args); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func TestRunRequiresCoverage(t *testing.T) {
	if code := runCrap(t, []string{"--dir", writeScanModule(t), "--ceiling", "10"}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunAcceptsDeprecatedProfileAlias(t *testing.T) {
	dir := writeScanModule(t)
	args := []string{"--dir", dir, "--profile", filepath.Join(dir, "coverage.out"), "--ceiling", "10"}
	if code := runCrap(t, args); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func TestRunRejectsConflictingCoverageFlags(t *testing.T) {
	dir := writeScanModule(t)
	profile := filepath.Join(dir, "coverage.out")
	args := []string{"--dir", dir, "--coverage", profile, "--profile", profile + ".other", "--ceiling", "10"}
	if code := runCrap(t, args); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunReadsCeilingFromThresholdsFile(t *testing.T) {
	for _, tc := range []struct {
		ceiling string
		want    int
	}{{"3", 0}, {"2", 1}} {
		dir := writeScanModule(t)
		thresholdsPath := writeThresholds(t, dir, tc.ceiling)
		args := []string{"--dir", dir, "--coverage", filepath.Join(dir, "coverage.out"), "--thresholds", thresholdsPath}
		if code := runCrap(t, args); code != tc.want {
			t.Errorf("ceiling %s: run() = %d, want %d", tc.ceiling, code, tc.want)
		}
	}
}

func TestRunCeilingFlagOverridesThresholdsFile(t *testing.T) {
	dir := writeScanModule(t)
	thresholdsPath := writeThresholds(t, dir, "2")
	args := []string{"--dir", dir, "--coverage", filepath.Join(dir, "coverage.out"), "--thresholds", thresholdsPath, "--ceiling", "10"}
	if code := runCrap(t, args); code != 0 {
		t.Fatalf("run() = %d, want 0 (explicit --ceiling must win)", code)
	}
}

func TestRunRequiresACeilingSource(t *testing.T) {
	dir := writeScanModule(t)
	args := []string{"--dir", dir, "--coverage", filepath.Join(dir, "coverage.out"), "--thresholds", filepath.Join(dir, "absent.yml")}
	if code := runCrap(t, args); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunRejectsThresholdsFileWithoutCrapCeiling(t *testing.T) {
	dir := writeScanModule(t)
	path := filepath.Join(dir, "other.yml")
	if err := os.WriteFile(path, []byte("metrics:\n  coverage_floor:\n    direction: min\n    value: 80\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--dir", dir, "--coverage", filepath.Join(dir, "coverage.out"), "--thresholds", path}
	if code := runCrap(t, args); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func writeThresholds(t *testing.T, dir, ceiling string) string {
	t.Helper()
	path := filepath.Join(dir, "thresholds.yml")
	body := "metrics:\n  crap_ceiling:\n    direction: max\n    value: " + ceiling + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runCrap(t *testing.T, args []string) int {
	t.Helper()
	out, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	errOut, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = out.Close() })
	t.Cleanup(func() { _ = errOut.Close() })
	return run(args, out, errOut)
}

func writeScanModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":  "module example.com/scan\n\ngo 1.26.4\n",
		"scan.go": "package scan\n\nfunc F(n int) int {\n\tif n > 0 {\n\t\treturn 1\n\t}\n\treturn 0\n}\n",
		"coverage.out": "mode: set\n" +
			"example.com/scan/scan.go:3.19,4.11 1 1\n" +
			"example.com/scan/scan.go:4.11,6.3 1 1\n" +
			"example.com/scan/scan.go:7.2,7.10 1 0\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestCoverageFromProfileRejectsMissingFile(t *testing.T) {
	if _, err := coverageFromProfile(t.TempDir(), filepath.Join(t.TempDir(), "missing.out")); err == nil {
		t.Fatal("coverageFromProfile() error = nil, want missing profile error")
	}
}

func TestPrintText(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "report")
	if err != nil {
		t.Fatal(err)
	}
	printText(file, crap.Report{Ceiling: 8, Functions: []crap.FunctionReport{{File: "a.go", Line: 1, Func: "A", Pass: true, Matched: true}, {File: "b.go", Line: 2, Func: "B", Pass: false}}, Summary: crap.Summary{Total: 2, Failing: 1}})
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "PASS") || !strings.Contains(string(data), "FAIL") {
		t.Fatalf("printText() = %q, want PASS and FAIL", data)
	}
}
