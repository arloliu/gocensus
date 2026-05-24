package contrib

import "slices"

import "strings"

func isFeature(subject string) bool {
	typ, text := commitTypeAndText(subject)
	if typ == "feat" || typ == "feature" {
		return true
	}
	return hasWord(text, "feature", "add", "adds", "added", "implement", "implements", "implemented", "support", "introduce", "introduces", "create", "creates", "enable", "enables")
}

func isFix(subject string) bool {
	typ, text := commitTypeAndText(subject)
	if typ == "fix" || typ == "bugfix" || typ == "hotfix" {
		return true
	}
	return hasWord(text, "fix", "fixes", "fixed", "bugfix", "hotfix", "resolve", "resolves", "resolved", "close", "closes", "closed")
}

func isRefactor(subject string) bool {
	typ, text := commitTypeAndText(subject)
	if typ == "refactor" {
		return true
	}
	return hasWord(text, "refactor", "refactors", "refactored", "rework", "reworks", "reworked")
}

func commitTypeAndText(subject string) (string, string) {
	text := strings.ToLower(strings.TrimSpace(subject))
	if idx := strings.Index(text, ":"); idx > 0 {
		prefix := text[:idx]
		if scope := strings.Index(prefix, "("); scope > 0 {
			prefix = prefix[:scope]
		}
		return strings.TrimSpace(prefix), text[idx+1:]
	}
	return "", text
}

func hasWord(text string, words ...string) bool {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	for _, field := range fields {
		if slices.Contains(words, field) {
			return true
		}
	}
	return false
}
