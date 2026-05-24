package gitrepo

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Root returns the top-level directory for the Git repository containing root.
func Root(ctx context.Context, root string) (string, error) {
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", "-C", absRoot, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", gitError("git rev-parse", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Archive extracts ref into a temporary directory without changing the worktree.
func Archive(ctx context.Context, root string, ref string) (string, func(), error) {
	if ref == "" {
		return "", nil, errors.New("git ref is required")
	}
	repoRoot, err := Root(ctx, root)
	if err != nil {
		return "", nil, err
	}
	out, err := os.MkdirTemp("", "gocensus-archive-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(out)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "archive", "--format=tar", ref)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if err := cmd.Start(); err != nil {
		cleanup()
		return "", nil, gitError("git archive", err)
	}
	if err := extractTar(pipe, out); err != nil {
		_ = cmd.Wait()
		cleanup()
		return "", nil, err
	}
	if err := cmd.Wait(); err != nil {
		cleanup()
		return "", nil, gitError("git archive", err)
	}
	return out, cleanup, nil
}

// Numstat returns git log --numstat output for the repository containing root.
func Numstat(ctx context.Context, root string, since string, until string) ([]byte, error) {
	repoRoot, err := Root(ctx, root)
	if err != nil {
		return nil, err
	}
	args := []string{"-C", repoRoot, "log", "--numstat", "--date=short", "--format=format:%x1e%H%x1f%ad"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	args = append(args, "--", ".")

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, gitError("git log", err)
	}
	return output, nil
}

func extractTar(r io.Reader, out string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(out, filepath.Clean(header.Name))
		if !strings.HasPrefix(target, out+string(os.PathSeparator)) && target != out {
			return fmt.Errorf("archive path escapes output directory: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tr); err != nil {
				_ = file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
	}
}

func gitError(action string, err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("%s: %s", action, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return fmt.Errorf("%s: %w", action, err)
}
