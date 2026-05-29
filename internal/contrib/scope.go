package contrib

// ScopeDescription returns the human-readable contribution scope.
func ScopeDescription(opts ParseOptions) string {
	testdataScope := "testdata paths excluded"
	if opts.IncludeTestdata {
		testdataScope = "testdata paths included"
	}
	if !opts.GoOnly {
		switch {
		case opts.IncludeGenerated && opts.IncludeMocks:
			return "all Git-tracked files, including generated and mock paths; " + testdataScope
		case opts.IncludeGenerated:
			return "all Git-tracked files, including generated paths and excluding mock paths; " + testdataScope
		case opts.IncludeMocks:
			return "all Git-tracked files, excluding generated paths and including mock paths; " + testdataScope
		}
		return "all Git-tracked files, generated and mock paths excluded; " + testdataScope
	}
	excludeGenerated := excludeGenerated(opts)
	excludeMocks := excludeMocks(opts)
	if !excludeGenerated && !excludeMocks {
		return "all Go files, including generated and mock paths; " + testdataScope
	}
	if !excludeGenerated {
		return "Go files, including generated paths and excluding mock paths; " + testdataScope
	}
	if !excludeMocks {
		return "Go files, excluding generated paths and including mock paths; " + testdataScope
	}
	return "human-authored Go files (*.go, generated and mock paths excluded; " + testdataScope + ")"
}

// ScopeNotes returns explanatory notes for contribution scope and metric meaning.
func ScopeNotes(opts ParseOptions) []string {
	notes := []string{
		"Feature, fix, and refactor counts are commit-message heuristics.",
		"Line, file, commit, and active-day counts come from git log --numstat.",
	}
	if opts.GoOnly {
		notes = append(notes, "Go-only filtering is path-based for historical safety; it does not inspect old file contents.")
	} else if excludeGenerated(opts) || excludeMocks(opts) || excludeTestdata(opts) {
		notes = append(notes, "Contribution path filtering is path-based for historical safety; it does not inspect old file contents.")
	}
	if excludeGenerated(opts) {
		notes = append(notes, "Generated paths are excluded by default; use --include-generated to include them.")
	}
	if excludeMocks(opts) {
		notes = append(notes, "Mock paths are excluded by default; use --include-mocks to include them.")
	}
	if excludeTestdata(opts) {
		notes = append(notes, "Testdata paths are excluded by default; use --include-testdata to include them.")
	}
	return notes
}
