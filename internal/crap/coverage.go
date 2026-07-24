package crap

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseCoverFunc parses the stdout of `go tool cover -func=<profile>`.
// Each function line looks like "file.go:12:\tFuncName\t83.3%" (tabwriter
// pads with a variable number of tabs); the trailing "total:\t...\t76.5%"
// line is skipped. The returned map is keyed "file:line:funcname".
func ParseCoverFunc(output string) (map[string]float64, error) {
	result := make(map[string]float64)
	for _, line := range strings.Split(output, "\n") {
		key, percent, ok, err := parseCoverLine(line)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result[key] = percent
	}
	return result, nil
}

func parseCoverLine(line string) (key string, percent float64, ok bool, err error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", 0, false, nil
	}
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return "", 0, false, fmt.Errorf("malformed cover -func line: %q", line)
	}
	if fields[0] == "total:" {
		return "", 0, false, nil
	}

	file, lineNum, err := parseCoverLocation(fields[0])
	if err != nil {
		return "", 0, false, err
	}
	percent, err = parseCoverPercent(fields[len(fields)-1])
	if err != nil {
		return "", 0, false, err
	}
	return buildKey(file, lineNum, fields[1]), percent, true, nil
}

func parseCoverLocation(value string) (string, int, error) {
	loc := strings.TrimSuffix(value, ":")
	idx := strings.LastIndex(loc, ":")
	if idx < 0 {
		return "", 0, fmt.Errorf("malformed file:line prefix: %q", value)
	}
	line, err := strconv.Atoi(loc[idx+1:])
	if err != nil {
		return "", 0, fmt.Errorf("malformed line number in %q: %w", value, err)
	}
	return loc[:idx], line, nil
}
func parseCoverPercent(value string) (float64, error) {
	percent, err := strconv.ParseFloat(strings.TrimSuffix(value, "%"), 64)
	if err != nil {
		return 0, fmt.Errorf("malformed coverage percent %q: %w", value, err)
	}
	if percent < 0 || percent > 100 {
		return 0, fmt.Errorf("coverage percent %q is outside 0-100", value)
	}
	return percent, nil
}
