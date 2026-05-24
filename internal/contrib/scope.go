package contrib

// ScopeDescription returns the human-readable contribution scope.
func ScopeDescription(opts ParseOptions) string {
	if !opts.GoOnly {
		return "all Git-tracked files"
	}
	if opts.IncludeGenerated && opts.IncludeMocks {
		return "all Go files, including generated and mock paths"
	}
	if opts.IncludeGenerated {
		return "Go files, including generated paths and excluding mock paths"
	}
	if opts.IncludeMocks {
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
		if !opts.IncludeGenerated {
			notes = append(notes, "Generated Go paths are excluded by default; use --include-generated with --go-only to include them.")
		}
		if !opts.IncludeMocks {
			notes = append(notes, "Mock Go paths are excluded by default; use --include-mocks with --go-only to include them.")
		}
	}
	return notes
}
