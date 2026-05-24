package contrib

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// ParseLog parses output from git log --numstat with the format used by Analyze.
func ParseLog(r io.Reader) (Report, error) {
	return ParseLogWithOptions(r, ParseOptions{})
}

// ParseLogWithOptions parses output from git log --numstat with path filtering.
func ParseLogWithOptions(r io.Reader, opts ParseOptions) (Report, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	byAuthor := map[string]*Contributor{}
	var pending *pendingCommit
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "\x1e") {
			flushCommit(byAuthor, pending)
			parsed, err := parseCommit(line)
			if err != nil {
				return Report{}, err
			}
			pending = &pendingCommit{
				commit: parsed,
				files:  map[string]struct{}{},
			}
			continue
		}
		if pending == nil {
			continue
		}
		added, removed, path, ok, err := parseNumstat(line)
		if err != nil {
			return Report{}, err
		}
		if !ok {
			continue
		}
		if !keepPath(path, opts) {
			continue
		}
		pending.kept = true
		pending.added += added
		pending.removed += removed
		if path != "" {
			pending.files[path] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return Report{}, err
	}
	flushCommit(byAuthor, pending)

	report := Report{Contributors: make([]Contributor, 0, len(byAuthor))}
	for _, contributor := range byAuthor {
		contributor.Files = len(contributor.filesByPath)
		contributor.ActiveDays = len(contributor.days)
		contributor.filesByPath = nil
		contributor.days = nil
		report.Contributors = append(report.Contributors, *contributor)
	}
	slices.SortFunc(report.Contributors, func(a, b Contributor) int {
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
	return report, nil
}

func parseCommit(line string) (commit, error) {
	fields := strings.Split(strings.TrimPrefix(line, "\x1e"), "\x1f")
	if len(fields) != 5 {
		return commit{}, fmt.Errorf("parse commit header: got %d fields, want 5", len(fields))
	}
	return commit{
		name:    fields[1],
		email:   fields[2],
		date:    fields[3],
		subject: fields[4],
	}, nil
}

func ensureContributor(contributors map[string]*Contributor, parsed commit) *Contributor {
	key := parsed.name + "\x00" + parsed.email
	contributor := contributors[key]
	if contributor == nil {
		contributor = &Contributor{
			Name:        parsed.name,
			Email:       parsed.email,
			filesByPath: map[string]struct{}{},
			days:        map[string]struct{}{},
		}
		contributors[key] = contributor
	}
	return contributor
}

func flushCommit(contributors map[string]*Contributor, pending *pendingCommit) {
	if pending == nil || !pending.kept {
		return
	}
	contributor := ensureContributor(contributors, pending.commit)
	contributor.Commits++
	if isFeature(pending.commit.subject) {
		contributor.Features++
	}
	if isFix(pending.commit.subject) {
		contributor.Fixes++
	}
	if isRefactor(pending.commit.subject) {
		contributor.Refactors++
	}
	if pending.commit.date != "" {
		contributor.days[pending.commit.date] = struct{}{}
		if contributor.FirstCommit == "" || pending.commit.date < contributor.FirstCommit {
			contributor.FirstCommit = pending.commit.date
		}
		if pending.commit.date > contributor.LastCommit {
			contributor.LastCommit = pending.commit.date
		}
	}
	contributor.Added += pending.added
	contributor.Removed += pending.removed
	contributor.Net += pending.added - pending.removed
	contributor.Churn += pending.added + pending.removed
	for path := range pending.files {
		contributor.filesByPath[path] = struct{}{}
	}
}

func parseNumstat(line string) (int, int, string, bool, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 3 {
		return 0, 0, "", false, nil
	}
	path := fields[2]
	if fields[0] == "-" || fields[1] == "-" {
		return 0, 0, path, true, nil
	}
	added, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, "", false, fmt.Errorf("parse added lines %q: %w", fields[0], err)
	}
	removed, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, "", false, fmt.Errorf("parse removed lines %q: %w", fields[1], err)
	}
	return added, removed, path, true, nil
}
