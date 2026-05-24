package contrib

// ScopeDescription returns the human-readable contribution scope.
func ScopeDescription(opts ParseOptions) string {
	if !opts.GoOnly {
		switch {
		case opts.IncludeGenerated && opts.IncludeMocks:
			return "all Git-tracked files, including generated and mock paths"
		case opts.IncludeGenerated:
			return "all Git-tracked files, including generated paths and excluding mock paths"
		case opts.IncludeMocks:
			return "all Git-tracked files, excluding generated paths and including mock paths"
		}
		return "all Git-tracked files, generated and mock paths excluded"
	}
	excludeGenerated := excludeGenerated(opts)
	excludeMocks := excludeMocks(opts)
	if !excludeGenerated && !excludeMocks {
		return "all Go files, including generated and mock paths"
	}
	if !excludeGenerated {
		return "Go files, including generated paths and excluding mock paths"
	}
	if !excludeMocks {
		return "Go files, excluding generated paths and including mock paths"
	}
	return "human-authored Go files (*.go, generated and mock paths excluded)"
}

// ScopeNotes returns explanatory notes for contribution scope and metric meaning.
func ScopeNotes(opts ParseOptions) []string {
	notes := []string{
		"Feature, fix, and refactor counts are commit-message heuristics.",
		"Line, file, commit, and active-day counts come from git log --numstat.",
	}
	if opts.GoOnly {
		notes = append(notes, "Go-only filtering is path-based for historical safety; it does not inspect old file contents.")
	} else if excludeGenerated(opts) || excludeMocks(opts) {
		notes = append(notes, "Contribution path filtering is path-based for historical safety; it does not inspect old file contents.")
	}
	if excludeGenerated(opts) {
		notes = append(notes, "Generated paths are excluded by default; use --include-generated to include them.")
	}
	if excludeMocks(opts) {
		notes = append(notes, "Mock paths are excluded by default; use --include-mocks to include them.")
	}
	return notes
}
