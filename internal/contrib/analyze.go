package contrib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Analyze returns author contribution metrics for Git-tracked history under Root.
func Analyze(ctx context.Context, opts Options) (Report, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}

	args := []string{
		"-C", absRoot,
		"log",
		"--numstat",
		"--date=short",
		"--format=format:%x1e%H%x1f%an%x1f%ae%x1f%ad%x1f%s",
	}
	if opts.Since != "" {
		args = append(args, "--since="+opts.Since)
	}
	if opts.Until != "" {
		args = append(args, "--until="+opts.Until)
	}
	args = append(args, "--", ".")

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Report{}, fmt.Errorf("git log: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return Report{}, fmt.Errorf("git log: %w", err)
	}

	parseOpts := ParseOptions{
		GoOnly:           opts.GoOnly,
		IncludeGenerated: opts.IncludeGenerated,
		IncludeMocks:     opts.IncludeMocks,
		IncludeTestdata:  opts.IncludeTestdata,
	}
	report, err := ParseLogWithOptions(bytes.NewReader(output), parseOpts)
	if err != nil {
		return Report{}, err
	}
	report.Root = absRoot
	report.Scope = ScopeDescription(parseOpts)
	report.Notes = ScopeNotes(parseOpts)
	return report, nil
}
