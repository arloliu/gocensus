package contrib_test

import (
	"strings"
	"testing"

	"github.com/arloliu/gocensus/internal/contrib"
)

func TestParseLogAggregatesAuthorsFromGitNumstat(t *testing.T) {
	input := strings.Join([]string{
		"\x1eabc\x1fAlice\x1falice@example.com\x1f2026-01-03\x1ffeat: add who command",
		"10\t2\tinternal/cli/commands.go",
		"5\t0\tREADME.md",
		"\x1edef\x1fBob\x1fbob@example.com\x1f2026-01-04\x1ffix: close issue #12",
		"1\t20\tinternal/count/count.go",
		"\x1eghi\x1fAlice\x1falice@example.com\x1f2026-01-04\x1frefactor: split contribution parser",
		"3\t7\tinternal/contrib/contrib.go",
	}, "\n")

	report, err := contrib.ParseLog(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	alice := contributorByName(t, report.Contributors, "Alice")
	if alice.Commits != 2 {
		t.Fatalf("Alice commits = %d, want 2", alice.Commits)
	}
	if alice.Features != 1 || alice.Fixes != 0 || alice.Refactors != 1 {
		t.Fatalf("Alice categories = feat %d fix %d refactor %d, want 1/0/1", alice.Features, alice.Fixes, alice.Refactors)
	}
	if alice.Added != 18 || alice.Removed != 9 || alice.Net != 9 || alice.Churn != 27 {
		t.Fatalf("Alice lines = added %d removed %d net %d churn %d, want 18/9/9/27",
			alice.Added, alice.Removed, alice.Net, alice.Churn)
	}
	if alice.Files != 3 {
		t.Fatalf("Alice files = %d, want 3", alice.Files)
	}
	if alice.ActiveDays != 2 {
		t.Fatalf("Alice active days = %d, want 2", alice.ActiveDays)
	}
	if alice.FirstCommit != "2026-01-03" || alice.LastCommit != "2026-01-04" {
		t.Fatalf("Alice first/last = %s/%s, want 2026-01-03/2026-01-04", alice.FirstCommit, alice.LastCommit)
	}

	bob := contributorByName(t, report.Contributors, "Bob")
	if bob.Fixes != 1 {
		t.Fatalf("Bob fixes = %d, want 1", bob.Fixes)
	}
	if bob.Net != -19 {
		t.Fatalf("Bob net = %d, want -19", bob.Net)
	}
}

func TestParseLogSkipsBinaryNumstatLines(t *testing.T) {
	input := strings.Join([]string{
		"\x1eabc\x1fAlice\x1falice@example.com\x1f2026-01-03\x1ffeat: add image asset",
		"-\t-\tassets/logo.png",
	}, "\n")

	report, err := contrib.ParseLog(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	alice := contributorByName(t, report.Contributors, "Alice")
	if alice.Added != 0 || alice.Removed != 0 || alice.Files != 1 {
		t.Fatalf("Alice binary stats = added %d removed %d files %d, want 0/0/1", alice.Added, alice.Removed, alice.Files)
	}
}

func TestParseLogWithOptionsKeepsHumanAuthoredGoByDefault(t *testing.T) {
	input := strings.Join([]string{
		"\x1eaaa\x1fAlice\x1falice@example.com\x1f2026-01-03\x1ffeat: add main",
		"10\t2\tmain.go",
		"100\t0\tservice.pb.go",
		"\x1ebbb\x1fBob\x1fbob@example.com\x1f2026-01-04\x1ffeat: add mock",
		"20\t0\tmock_client.go",
		"\x1eccc\x1fCarol\x1fcarol@example.com\x1f2026-01-05\x1ffeat: add docs",
		"5\t0\tREADME.md",
	}, "\n")

	report, err := contrib.ParseLogWithOptions(strings.NewReader(input), contrib.ParseOptions{
		GoOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Contributors) != 1 {
		t.Fatalf("contributors = %#v, want only Alice", report.Contributors)
	}
	alice := contributorByName(t, report.Contributors, "Alice")
	if alice.Commits != 1 || alice.Added != 10 || alice.Removed != 2 || alice.Files != 1 {
		t.Fatalf("Alice = %#v, want one human-authored Go file only", alice)
	}
}

func TestParseLogWithOptionsCanIncludeGeneratedAndMocks(t *testing.T) {
	input := strings.Join([]string{
		"\x1eaaa\x1fAlice\x1falice@example.com\x1f2026-01-03\x1ffeat: add generated",
		"100\t0\tservice.pb.go",
		"\x1ebbb\x1fBob\x1fbob@example.com\x1f2026-01-04\x1ffeat: add mock",
		"20\t0\tmocks/client.go",
		"\x1eccc\x1fCarol\x1fcarol@example.com\x1f2026-01-05\x1ffeat: add docs",
		"5\t0\tREADME.md",
	}, "\n")

	report, err := contrib.ParseLogWithOptions(strings.NewReader(input), contrib.ParseOptions{
		GoOnly:           true,
		IncludeGenerated: true,
		IncludeMocks:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := names(report.Contributors); strings.Join(got, ",") != "Alice,Bob" {
		t.Fatalf("contributors = %v, want Alice,Bob", got)
	}
}

func TestParseLogWithOptionsCanExcludeGeneratedAndMocksAcrossAllPaths(t *testing.T) {
	input := strings.Join([]string{
		"\x1eaaa\x1fAlice\x1falice@example.com\x1f2026-01-03\x1ffeat: add generated docs",
		"100\t0\tdocs/generated/report.md",
		"\x1ebbb\x1fBob\x1fbob@example.com\x1f2026-01-04\x1ffeat: add mock fixture",
		"20\t0\tfixtures/mock_payload.json",
		"\x1eccc\x1fCarol\x1fcarol@example.com\x1f2026-01-05\x1ffeat: add docs",
		"5\t0\tREADME.md",
		"\x1eddd\x1fDan\x1fdan@example.com\x1f2026-01-06\x1ffeat: add source",
		"7\t1\tinternal/app/main.go",
	}, "\n")

	report, err := contrib.ParseLogWithOptions(strings.NewReader(input), contrib.ParseOptions{
		ExcludeGenerated: true,
		ExcludeMocks:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := names(report.Contributors); strings.Join(got, ",") != "Carol,Dan" {
		t.Fatalf("contributors = %v, want Carol,Dan", got)
	}
}

func TestRankSortsAndLimitsContributors(t *testing.T) {
	report := contrib.Report{
		Contributors: []contrib.Contributor{
			{Name: "Alice", Commits: 2, Churn: 30},
			{Name: "Bob", Commits: 5, Churn: 10},
			{Name: "Carol", Commits: 3, Churn: 20},
		},
	}

	ranked, err := contrib.Rank(report, contrib.RankOptions{By: contrib.SortCommits, Top: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := names(ranked); strings.Join(got, ",") != "Bob,Carol" {
		t.Fatalf("ranked by commits = %v, want Bob,Carol", got)
	}

	ranked, err = contrib.Rank(report, contrib.RankOptions{By: contrib.SortChurn})
	if err != nil {
		t.Fatal(err)
	}
	if got := names(ranked); strings.Join(got, ",") != "Alice,Carol,Bob" {
		t.Fatalf("ranked by churn = %v, want Alice,Carol,Bob", got)
	}
}

func TestRankRejectsUnknownSort(t *testing.T) {
	_, err := contrib.Rank(contrib.Report{}, contrib.RankOptions{By: "unknown"})
	if err == nil {
		t.Fatal("Rank error = nil, want error")
	}
}

func contributorByName(t *testing.T, contributors []contrib.Contributor, name string) contrib.Contributor {
	t.Helper()
	for _, contributor := range contributors {
		if contributor.Name == name {
			return contributor
		}
	}
	t.Fatalf("missing contributor %q in %#v", name, contributors)
	return contrib.Contributor{}
}

func names(contributors []contrib.Contributor) []string {
	out := make([]string, 0, len(contributors))
	for _, contributor := range contributors {
		out = append(out, contributor.Name)
	}
	return out
}
