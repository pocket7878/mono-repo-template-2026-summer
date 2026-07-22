package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var excludedPrefixes = []string{
	"cmd/",
	"internal/handler/api/",
	"internal/repository/internal/db/",
	"internal/lint/",
}

type coverage struct {
	covered int
	total   int
}

func main() {
	coverProfile := flag.String("coverprofile", "", "path to Go coverage profile")
	threshold := flag.Float64("threshold", 85.0, "minimum scoped coverage")
	flag.Parse()
	if *coverProfile == "" {
		fmt.Fprintln(os.Stderr, "coverprofile is required")
		os.Exit(2)
	}

	result, err := scopedCoverage(*coverProfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	percentage := roundToTenth(float64(result.covered) * 100 / float64(result.total))
	fmt.Printf("backend scoped coverage: %.1f%% (%d/%d statements)\n", percentage, result.covered, result.total)
	if percentage+1e-9 < *threshold {
		fmt.Fprintf(os.Stderr, "coverage gate failed: %.1f%% is below %.1f%%\n", percentage, *threshold)
		os.Exit(1)
	}
	fmt.Printf("coverage gate passed: %.1f%% >= %.1f%%\n", percentage, *threshold)
}

func scopedCoverage(path string) (coverage, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return coverage{}, err
	}
	defer file.Close()

	result := coverage{}
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		if line == 1 {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			return coverage{}, fmt.Errorf("invalid coverage row %q", scanner.Text())
		}
		filePath := strings.SplitN(fields[0], ":", 2)[0]
		filePath = strings.TrimPrefix(filePath, "github.com/pocket7878/mono-repo-template-2026-summer/backend/")
		if isExcluded(filePath) {
			continue
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil {
			return coverage{}, err
		}
		executed, err := strconv.Atoi(fields[2])
		if err != nil {
			return coverage{}, err
		}
		result.total += statements
		if executed > 0 {
			result.covered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		return coverage{}, err
	}
	if result.total == 0 {
		return coverage{}, fmt.Errorf("scoped coverage has no statements")
	}
	return result, nil
}

func isExcluded(path string) bool {
	for _, prefix := range excludedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func roundToTenth(value float64) float64 {
	return math.Round(value*10) / 10
}
