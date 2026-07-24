package crap_test

import (
	"testing"

	"github.com/jadenmaciel/gauntlet/internal/crap"
)

func TestComplexityOfFile(t *testing.T) {
	tests := []struct {
		name           string
		src            string
		wantFunc       string
		wantLine       int
		wantComplexity int
	}{
		{
			name: "straight_line_cc1",
			src: `package p

func F1() int {
	x := 1
	y := 2
	return x + y
}
`,
			wantFunc:       "F1",
			wantLine:       3,
			wantComplexity: 1,
		},
		{
			name: "two_ifs_cc3",
			src: `package p

func F2(a, b int) int {
	if a > 0 {
		return a
	}
	if b > 0 {
		return b
	}
	return 0
}
`,
			wantFunc:       "F2",
			wantLine:       3,
			wantComplexity: 3,
		},
		{
			name: "switch_case_default",
			src: `package p

func F3(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	default:
		return "other"
	}
}
`,
			wantFunc:       "F3",
			wantLine:       3,
			wantComplexity: 3,
		},
		{
			name: "for_range",
			src: `package p

func F4(items []int) int {
	sum := 0
	for _, v := range items {
		sum += v
	}
	return sum
}
`,
			wantFunc:       "F4",
			wantLine:       3,
			wantComplexity: 2,
		},
		{
			name: "cc9_plus",
			src: `package p

func F5(a, b, c, d int) int {
	if a > 0 && b > 0 {
		return 1
	}
	if a > 0 || b > 0 {
		return 2
	}
	for i := 0; i < a; i++ {
		if i == b {
			return 3
		}
	}
	switch c {
	case 1:
		return 4
	case 2:
		return 5
	case 3:
		return 6
	default:
		return 7
	}
	return 0
}
`,
			wantFunc:       "F5",
			wantLine:       3,
			wantComplexity: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcs, err := crap.ComplexityOfFile([]byte(tt.src))
			if err != nil {
				t.Fatalf("ComplexityOfFile() error = %v", err)
			}
			if len(funcs) != 1 {
				t.Fatalf("ComplexityOfFile() returned %d functions, want 1", len(funcs))
			}
			got := funcs[0]
			if got.Func != tt.wantFunc {
				t.Errorf("Func = %q, want %q", got.Func, tt.wantFunc)
			}
			if got.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", got.Line, tt.wantLine)
			}
			if got.Complexity != tt.wantComplexity {
				t.Errorf("Complexity = %d, want %d", got.Complexity, tt.wantComplexity)
			}
		})
	}
}

func TestComplexityOfFile_MultipleFunctionsAndMethods(t *testing.T) {
	src := `package p

type T struct{}

func Plain() {
}

func (t *T) Method(x int) int {
	if x > 0 {
		return x
	}
	return -x
}
`
	funcs, err := crap.ComplexityOfFile([]byte(src))
	if err != nil {
		t.Fatalf("ComplexityOfFile() error = %v", err)
	}
	if len(funcs) != 2 {
		t.Fatalf("ComplexityOfFile() returned %d functions, want 2", len(funcs))
	}
	byName := map[string]crap.FunctionComplexity{}
	for _, f := range funcs {
		byName[f.Func] = f
	}
	if f, ok := byName["Plain"]; !ok || f.Complexity != 1 {
		t.Errorf("Plain: got %+v, want complexity 1", f)
	}
	if f, ok := byName["Method"]; !ok || f.Complexity != 2 {
		t.Errorf("Method: got %+v, want complexity 2", f)
	}
}

func TestComplexityOfFile_InvalidSource(t *testing.T) {
	_, err := crap.ComplexityOfFile([]byte("this is not { go source"))
	if err == nil {
		t.Fatal("ComplexityOfFile() error = nil, want error for invalid source")
	}
}
