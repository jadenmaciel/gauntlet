// Command gauntlet manages a monotonic ratchet file for quality
// thresholds: floors that may only rise, ceilings that may only fall.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jadenmaciel/gauntlet/internal/thresholds"
)

const defaultThresholdsPath = ".gauntlet/thresholds.yml"

// placeholderCeiling is deliberately not a recommendation. A brownfield repo
// almost never starts under 8, so initializing there makes the very first
// scorer run fail on code nobody just wrote. Measure a baseline instead:
//
//	crap4go --dir . --coverage coverage.out --ceiling 100000 --format json
//
// then pass the reported summary.max_crap to "gauntlet init --crap-ceiling".
const placeholderCeiling = 8

const starterTemplateFormat = `# Quality threshold ratchet. Floors (direction: min) may only rise;
# ceilings (direction: max) may only fall. Use "gauntlet ratchet" to move
# a value, never hand-edit it downward/upward.
#
# The 0-valued floors below (mutation_efficacy, mutation_mcover) are
# placeholders. They mean nothing until the first "gauntlet ratchet" call
# sets them to a real measured baseline.
metrics:
  crap_ceiling:%s
    direction: max
    value: %s
  mutation_efficacy:
    direction: min
    value: 0
  mutation_mcover:
    direction: min
    value: 0
  coverage_min:
    direction: min
    value: 80
`

// placeholderNote is emitted only when the caller did not supply a measured
// ceiling, so a hand-written baseline does not get labelled a guess.
const placeholderNote = `
    # PLACEHOLDER, not a recommendation. Score your tree first and re-run
    # "gauntlet init --force --crap-ceiling <measured summary.max_crap>",
    # then ratchet it down from there.`

func starterTemplate(ceiling float64, measured bool) string {
	note := placeholderNote
	if measured {
		note = ""
	}
	return fmt.Sprintf(starterTemplateFormat, note, formatCeiling(ceiling))
}

func formatCeiling(value float64) string {
	if value == math.Trunc(value) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gauntlet <init|verify|ratchet> ...")
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "verify":
		return runVerify(args[1:])
	case "ratchet":
		return runRatchet(args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q (want init, verify, or ratchet)", args[0])
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite an existing thresholds file")
	ceiling := fs.Float64("crap-ceiling", placeholderCeiling, "measured baseline CRAP ceiling to start the ratchet at")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("init accepts at most one path")
	}
	if *ceiling <= 0 {
		return fmt.Errorf("--crap-ceiling must be positive, got %v", *ceiling)
	}

	measured := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "crap-ceiling" {
			measured = true
		}
	})

	path := defaultThresholdsPath
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	if err := writeStarter(path, *force, *ceiling, measured); err != nil {
		return err
	}

	fmt.Printf("wrote starter thresholds file to %s\n", path)
	if !measured {
		fmt.Fprintf(os.Stderr,
			"gauntlet: crap_ceiling is a placeholder %s, not a measured baseline. "+
				"Score your tree, then re-run with --force --crap-ceiling <summary.max_crap>.\n",
			formatCeiling(*ceiling))
	}
	return nil
}

func writeStarter(path string, force bool, ceiling float64, measured bool) error {
	if !force {
		if err := requireNewFile(path); err != nil {
			return err
		}
	}
	if err := createParentDirectory(path); err != nil {
		return err
	}
	body := starterTemplate(ceiling, measured)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("writing thresholds file %q: %w", path, err)
	}
	return nil
}

func requireNewFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("refusing to overwrite existing file %q (use --force)", path)
	}
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("checking existing file %q: %w", path, err)
}

func createParentDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}
	return nil
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	path := fs.String("thresholds", defaultThresholdsPath, "path to thresholds file")
	metric := fs.String("metric", "", "metric name")
	value := fs.Float64("value", 0, "measured value")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("verify accepts flags only")
	}
	if *metric == "" {
		return fmt.Errorf("--metric is required")
	}

	tf, err := thresholds.Load(*path)
	if err != nil {
		return err
	}

	if err := thresholds.Verify(tf, *metric, *value); err != nil {
		return err
	}
	fmt.Printf("%s %v OK\n", *metric, *value)
	return nil
}

func runRatchet(args []string) error {
	path, metric, value, err := ratchetArgs(args)
	if err != nil {
		return err
	}
	tf, err := thresholds.Load(path)
	if err != nil {
		return err
	}
	newValue, changed, err := thresholds.Ratchet(tf, metric, value)
	if err != nil {
		return err
	}
	if err := saveRatchet(path, tf, changed); err != nil {
		return err
	}
	fmt.Printf("%s ratcheted to %v (changed=%v)\n", metric, newValue, changed)
	return nil
}

func ratchetArgs(args []string) (string, string, float64, error) {
	fs := flag.NewFlagSet("ratchet", flag.ContinueOnError)
	path := fs.String("thresholds", defaultThresholdsPath, "path to thresholds file")
	metric := fs.String("metric", "", "metric name")
	value := fs.Float64("value", 0, "measured value")
	if err := fs.Parse(args); err != nil {
		return "", "", 0, err
	}
	if fs.NArg() != 0 {
		return "", "", 0, fmt.Errorf("ratchet accepts flags only")
	}
	if *metric == "" {
		return "", "", 0, fmt.Errorf("--metric is required")
	}
	return *path, *metric, *value, nil
}

func saveRatchet(path string, tf *thresholds.File, changed bool) error {
	if !changed {
		return nil
	}
	return thresholds.Save(path, tf)
}
