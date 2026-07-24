package crap_test

import (
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func TestParseCoverFunc(t *testing.T) {
	// Mirrors the real `go tool cover -func` output: tabwriter pads
	// columns with a variable number of tabs, and the final "total:"
	// line must be skipped rather than treated as a function.
	output := "github.com/x/y.go:12:\tFuncName\t83.3%\n" +
		"github.com/x/y.go:34:\tOtherFunc\t\t0.0%\n" +
		"github.com/x/z.go:7:\tThird\t100.0%\n" +
		"total:\t\t\t(statements)\t76.5%\n"

	got, err := crap.ParseCoverFunc(output)
	if err != nil {
		t.Fatalf("ParseCoverFunc() error = %v", err)
	}

	want := map[string]float64{
		"github.com/x/y.go:12:FuncName":  83.3,
		"github.com/x/y.go:34:OtherFunc": 0.0,
		"github.com/x/z.go:7:Third":      100.0,
	}

	if len(got) != len(want) {
		t.Fatalf("ParseCoverFunc() returned %d entries, want %d: %+v", len(got), len(want), got)
	}
	for k, v := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing key %q in result %+v", k, got)
			continue
		}
		if gv != v {
			t.Errorf("key %q = %v, want %v", k, gv, v)
		}
	}
	if _, ok := got["total:(statements)"]; ok {
		t.Errorf("total line should be skipped, got entry for it: %+v", got)
	}
}

func TestParseCoverFunc_EmptyInput(t *testing.T) {
	got, err := crap.ParseCoverFunc("")
	if err != nil {
		t.Fatalf("ParseCoverFunc() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseCoverFunc(\"\") = %+v, want empty map", got)
	}
}

func TestParseCoverFunc_OnlyTotalLine(t *testing.T) {
	got, err := crap.ParseCoverFunc("total:\t\t\t(statements)\t100.0%\n")
	if err != nil {
		t.Fatalf("ParseCoverFunc() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseCoverFunc() = %+v, want empty map", got)
	}
}

func TestParseCoverFunc_RejectsInvalidPercent(t *testing.T) {
	_, err := crap.ParseCoverFunc("github.com/x/y.go:12:\tFuncName\t101.0%\n")
	if err == nil {
		t.Fatal("ParseCoverFunc() error = nil, want error for coverage above 100%")
	}
}
