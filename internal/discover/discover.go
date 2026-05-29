package discover

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Options controls Go source file discovery.
type Options struct {
	Root            string
	UseGitignore    bool
	ExtraExcludes   []string
	IncludeTestdata bool
}

// GoFiles returns Go source files under Root after applying built-in excludes
// and optional gitignore-style patterns.
func GoFiles(ctx context.Context, opts Options) ([]string, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	matcher, err := buildMatcher(root, opts)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(root, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if filePath == root {
			return nil
		}

		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if entry.IsDir() && isHardExcludedDir(entry.Name(), opts.IncludeTestdata) {
			return filepath.SkipDir
		}
		if matcher.match(rel, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		files = append(files, filePath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(files)
	return files, nil
}

type matcher struct {
	patterns []ignorePattern
}

type ignorePattern struct {
	domain  string
	pattern string
	dirOnly bool
}

func buildMatcher(root string, opts Options) (matcher, error) {
	var patterns []ignorePattern
	if opts.UseGitignore {
		err := filepath.WalkDir(root, func(filePath string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() && filePath != root && isHardExcludedDir(entry.Name(), opts.IncludeTestdata) {
				return filepath.SkipDir
			}
			if entry.IsDir() || entry.Name() != ".gitignore" {
				return nil
			}

			relDir, err := filepath.Rel(root, filepath.Dir(filePath))
			if err != nil {
				return err
			}
			loaded, err := readPatterns(filePath, cleanDomain(relDir))
			if err != nil {
				return err
			}
			patterns = append(patterns, loaded...)
			return nil
		})
		if err != nil {
			return matcher{}, err
		}
	}
	for _, pattern := range opts.ExtraExcludes {
		if parsed, ok := parsePattern(pattern, ""); ok {
			patterns = append(patterns, parsed)
		}
	}
	return matcher{patterns: patterns}, nil
}

func readPatterns(filePath string, domain string) ([]ignorePattern, error) {
	file, err := os.Open(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var patterns []ignorePattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if parsed, ok := parsePattern(scanner.Text(), domain); ok {
			patterns = append(patterns, parsed)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

func parsePattern(raw string, domain string) (ignorePattern, bool) {
	pattern := strings.TrimSpace(raw)
	if pattern == "" || strings.HasPrefix(pattern, "#") || strings.HasPrefix(pattern, "!") {
		return ignorePattern{}, false
	}
	dirOnly := strings.HasSuffix(pattern, "/")
	pattern = strings.TrimPrefix(strings.TrimSuffix(pattern, "/"), "/")
	if pattern == "" {
		return ignorePattern{}, false
	}
	return ignorePattern{
		domain:  domain,
		pattern: filepath.ToSlash(pattern),
		dirOnly: dirOnly,
	}, true
}

func (m matcher) match(rel string, isDir bool) bool {
	for _, pattern := range m.patterns {
		if pattern.match(rel, isDir) {
			return true
		}
	}
	return false
}

func (p ignorePattern) match(rel string, isDir bool) bool {
	target, ok := p.target(rel)
	if !ok {
		return false
	}
	if p.dirOnly {
		return isDir && target == p.pattern || strings.HasPrefix(target, p.pattern+"/")
	}
	if !strings.Contains(p.pattern, "/") {
		return path.Base(target) == p.pattern
	}
	matched, err := path.Match(p.pattern, target)
	if err != nil {
		return false
	}
	return matched
}

func (p ignorePattern) target(rel string) (string, bool) {
	if p.domain == "" {
		return rel, true
	}
	if rel == p.domain {
		return "", true
	}
	prefix := p.domain + "/"
	if !strings.HasPrefix(rel, prefix) {
		return "", false
	}
	return strings.TrimPrefix(rel, prefix), true
}

func cleanDomain(relDir string) string {
	relDir = filepath.ToSlash(relDir)
	if relDir == "." {
		return ""
	}
	return relDir
}

func isHardExcludedDir(name string, includeTestdata bool) bool {
	if name == ".git" || name == "vendor" || name == "node_modules" {
		return true
	}
	if name == "testdata" {
		return !includeTestdata
	}
	return strings.HasPrefix(name, ".")
}
