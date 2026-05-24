package contrib

const (
	SortCommits    = "commits"
	SortFeatures   = "features"
	SortFixes      = "fixes"
	SortRefactors  = "refactors"
	SortAdded      = "added"
	SortRemoved    = "removed"
	SortNet        = "net"
	SortShrink     = "shrink"
	SortChurn      = "churn"
	SortFiles      = "files"
	SortActiveDays = "active-days"
)

// Options controls Git contribution analysis.
type Options struct {
	Root             string
	Since            string
	Until            string
	GoOnly           bool
	IncludeGenerated bool
	IncludeMocks     bool
}

// ParseOptions controls filtering while parsing Git numstat history.
type ParseOptions struct {
	GoOnly           bool
	IncludeGenerated bool
	IncludeMocks     bool
}

// RankOptions controls contributor ordering and truncation.
type RankOptions struct {
	By  string
	Top int
}

// Report contains contribution metrics derived from Git history.
type Report struct {
	Root         string        `json:"root"`
	Scope        string        `json:"scope"`
	Notes        []string      `json:"notes,omitempty"`
	Contributors []Contributor `json:"contributors"`
}

// Contributor contains contribution metrics for one Git author identity.
type Contributor struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Commits     int    `json:"commits"`
	Features    int    `json:"features"`
	Fixes       int    `json:"fixes"`
	Refactors   int    `json:"refactors"`
	Added       int    `json:"added"`
	Removed     int    `json:"removed"`
	Net         int    `json:"net"`
	Churn       int    `json:"churn"`
	Files       int    `json:"files"`
	ActiveDays  int    `json:"active_days"`
	FirstCommit string `json:"first_commit"`
	LastCommit  string `json:"last_commit"`

	filesByPath map[string]struct{}
	days        map[string]struct{}
}

type commit struct {
	name    string
	email   string
	date    string
	subject string
}

type pendingCommit struct {
	commit  commit
	added   int
	removed int
	files   map[string]struct{}
	kept    bool
}
