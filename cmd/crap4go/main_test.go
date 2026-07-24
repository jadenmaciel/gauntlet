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
	if code := runCrap(t, []string{"--dir", writeScanModule(t), "--profile", filepath.Join(t.TempDir(), "missing.out")}); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunScansModule(t *testing.T) {
	if code := runCrap(t, []string{"--dir", writeScanModule(t), "--format", "json"}); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
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
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/scan\n\ngo 1.26.4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scan.go"), []byte("package scan\n\nfunc F() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCoverageFromProfileRejectsMissingFile(t *testing.T) {
	if _, err := coverageFromProfile(filepath.Join(t.TempDir(), "missing.out")); err == nil {
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
