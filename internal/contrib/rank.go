package contrib

import (
	"fmt"
	"slices"
)

// Rank returns contributors sorted by the selected contribution metric.
func Rank(report Report, opts RankOptions) ([]Contributor, error) {
	sortBy := opts.By
	if sortBy == "" {
		sortBy = SortCommits
	}
	if !knownSort(sortBy) {
		return nil, fmt.Errorf("unknown contribution sort %q", sortBy)
	}

	ranked := slices.Clone(report.Contributors)
	slices.SortFunc(ranked, func(a, b Contributor) int {
		if cmp := compareContributor(a, b, sortBy); cmp != 0 {
			return cmp
		}
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		if a.Email < b.Email {
			return -1
		}
		if a.Email > b.Email {
			return 1
		}
		return 0
	})
	if opts.Top > 0 && len(ranked) > opts.Top {
		ranked = ranked[:opts.Top]
	}
	return ranked, nil
}

func knownSort(sortBy string) bool {
	switch sortBy {
	case SortCommits, SortFeatures, SortFixes, SortRefactors, SortAdded, SortRemoved, SortNet, SortShrink, SortChurn, SortFiles, SortActiveDays:
		return true
	default:
		return false
	}
}

func compareContributor(a Contributor, b Contributor, sortBy string) int {
	av := contributorValue(a, sortBy)
	bv := contributorValue(b, sortBy)
	if av > bv {
		return -1
	}
	if av < bv {
		return 1
	}
	return 0
}

func contributorValue(contributor Contributor, sortBy string) int {
	switch sortBy {
	case SortFeatures:
		return contributor.Features
	case SortFixes:
		return contributor.Fixes
	case SortRefactors:
		return contributor.Refactors
	case SortAdded:
		return contributor.Added
	case SortRemoved:
		return contributor.Removed
	case SortNet:
		return contributor.Net
	case SortShrink:
		return -contributor.Net
	case SortChurn:
		return contributor.Churn
	case SortFiles:
		return contributor.Files
	case SortActiveDays:
		return contributor.ActiveDays
	default:
		return contributor.Commits
	}
}
