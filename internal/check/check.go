package check

import (
	"fmt"

	"github.com/arloliu/gocensus"
)

// Options controls repository policy checks.
type Options struct {
	MinTestRatio float64
}

// Report contains the result of repository policy checks.
type Report struct {
	Root       string  `json:"root"`
	ModulePath string  `json:"module_path"`
	Scope      string  `json:"scope"`
	Passed     bool    `json:"passed"`
	Checks     []Check `json:"checks"`
}

// Check contains one policy check result.
type Check struct {
	Name      string  `json:"name"`
	Label     string  `json:"label"`
	Passed    bool    `json:"passed"`
	Actual    float64 `json:"actual"`
	Threshold float64 `json:"threshold"`
	Message   string  `json:"message"`
}

// Evaluate checks census metrics against the requested policy thresholds.
func Evaluate(result gocensus.Result, opts Options) Report {
	checks := []Check{
		minimumTestRatio(result.Ratios.TestToProductionEffective, opts.MinTestRatio),
	}
	report := Report{
		Root:       result.Root,
		ModulePath: result.ModulePath,
		Scope:      result.Scope,
		Passed:     true,
		Checks:     checks,
	}
	for _, check := range checks {
		if !check.Passed {
			report.Passed = false
			break
		}
	}
	return report
}

func minimumTestRatio(actual float64, threshold float64) Check {
	passed := actual >= threshold
	message := fmt.Sprintf("actual %s meets minimum %s", formatRatio(actual), formatRatio(threshold))
	if !passed {
		message = fmt.Sprintf("actual %s below minimum %s", formatRatio(actual), formatRatio(threshold))
	}
	return Check{
		Name:      "minimum_test_ratio",
		Label:     "Minimum Test Ratio",
		Passed:    passed,
		Actual:    actual,
		Threshold: threshold,
		Message:   message,
	}
}

func formatRatio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
