package contrib

// ScopeDescription returns the human-readable contribution scope.
func ScopeDescription(opts ParseOptions) string {
	if !opts.GoOnly {
		switch {
		case opts.ExcludeGenerated && opts.ExcludeMocks:
			return "all Git-tracked files, generated and mock paths excluded"
		case opts.ExcludeGenerated:
			return "all Git-tracked files, generated paths excluded"
		case opts.ExcludeMocks:
			return "all Git-tracked files, mock paths excluded"
		}
		return "all Git-tracked files"
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
		if opts.ExcludeGenerated {
			notes = append(notes, "Generated Go paths are excluded.")
		} else if excludeGenerated(opts) {
			notes = append(notes, "Generated Go paths are excluded by default; use --include-generated with --go-only to include them.")
		}
		if opts.ExcludeMocks {
			notes = append(notes, "Mock Go paths are excluded.")
		} else if excludeMocks(opts) {
			notes = append(notes, "Mock Go paths are excluded by default; use --include-mocks with --go-only to include them.")
		}
	} else if opts.ExcludeGenerated || opts.ExcludeMocks {
		notes = append(notes, "Contribution path filtering is path-based for historical safety; it does not inspect old file contents.")
	}
	return notes
}
