package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

type Benchmark struct {
	NsPerOp      int `json:"ns_per_op"`
	AllocsPerOp  int `json:"allocs_per_op"`
	BytesPerOp   int `json:"bytes_per_op"`
}

type ServiceBenchmarks map[string]Benchmark

type BenchmarkFile struct {
	Timestamp  string                       `json:"timestamp"`
	Benchmarks map[string]ServiceBenchmarks `json:"benchmarks"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run compare.go <baseline.json> <current.json>")
		os.Exit(1)
	}

	baseline, err := loadBenchmarkFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error loading baseline: %v\n", err)
		os.Exit(1)
	}

	current, err := loadBenchmarkFile(os.Args[2])
	if err != nil {
		fmt.Printf("Error loading current: %v\n", err)
		os.Exit(1)
	}

	violations := compareBenchmarks(baseline, current, 0.10) // ±10% tolerance

	if len(violations) > 0 {
		fmt.Println("Performance regressions detected:")
		for _, v := range violations {
			fmt.Println(v)
		}
		os.Exit(1)
	}

	fmt.Println("All benchmarks within ±10% tolerance")
}

func loadBenchmarkFile(path string) (*BenchmarkFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bf BenchmarkFile
	if err := json.Unmarshal(data, &bf); err != nil {
		return nil, err
	}

	return &bf, nil
}

func compareBenchmarks(baseline, current *BenchmarkFile, tolerance float64) []string {
	var violations []string

	for serviceName, baselineService := range baseline.Benchmarks {
		currentService, exists := current.Benchmarks[serviceName]
		if !exists {
			violations = append(violations, fmt.Sprintf("Service %s missing in current", serviceName))
			continue
		}

		for benchName, baselineBench := range baselineService {
			currentBench, exists := currentService[benchName]
			if !exists {
				violations = append(violations, fmt.Sprintf("%s.%s missing in current", serviceName, benchName))
				continue
			}

			// Check ns_per_op (lower is better)
			percentChange := float64(currentBench.NsPerOp-baselineBench.NsPerOp) / float64(baselineBench.NsPerOp)
			if math.Abs(percentChange) > tolerance {
				violations = append(violations, fmt.Sprintf(
					"%s.%s: ns_per_op changed by %.1f%% (baseline: %d, current: %d)",
					serviceName, benchName, percentChange*100, baselineBench.NsPerOp, currentBench.NsPerOp,
				))
			}

			// Check allocations (lower is better)
			allocChange := float64(currentBench.AllocsPerOp-baselineBench.AllocsPerOp) / float64(baselineBench.AllocsPerOp)
			if math.Abs(allocChange) > tolerance {
				violations = append(violations, fmt.Sprintf(
					"%s.%s: allocs_per_op changed by %.1f%% (baseline: %d, current: %d)",
					serviceName, benchName, allocChange*100, baselineBench.AllocsPerOp, currentBench.AllocsPerOp,
				))
			}
		}
	}

	return violations
}