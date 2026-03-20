// Package gitcontrol provides safe, deterministic Git operations by
// shelling out to the system git binary.
//
// Design trade-off: shelling out to git (rather than using go-git) keeps
// the dependency tree small and leverages the user's installed git version.
// All commands use exec.CommandContext for timeout safety.  Output is
// captured and returned as structured errors on failure.
//
// The executor never operates on the user's working copy directly; callers
// should point it at a dedicated repo path (or worktree) to avoid
// corrupting in-flight work.
package gitcontrol

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Executor runs git commands against a local repository.
type Executor struct {
	repoPath string
	logger   *slog.Logger
}

// NewExecutor creates an Executor for the repository at repoPath.
// The directory must already exist (it may be empty for init).
func NewExecutor(repoPath string) *Executor {
	return &Executor{
		repoPath: repoPath,
		logger:   slog.Default().With("component", "gitcontrol"),
	}
}

// RepoPath returns the configured repository path.
func (e *Executor) RepoPath() string {
	return e.repoPath
}

// ---------------------------------------------------------------------------
// Low-level command runner
// ---------------------------------------------------------------------------

// GitResult holds the output of a git command.
type GitResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes a git command with the given arguments.
func (e *Executor) Run(ctx context.Context, args ...string) (*GitResult, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e.logger.Debug("git exec", "args", args, "dir", e.repoPath)

	err := cmd.Run()
	result := &GitResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("git %s failed (exit %d): %s",
			strings.Join(args, " "), result.ExitCode, result.Stderr)
	}
	if err != nil {
		return result, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Repository lifecycle
// ---------------------------------------------------------------------------

// Init initialises a new git repository at the executor's repoPath.
func (e *Executor) Init(ctx context.Context) error {
	if err := os.MkdirAll(e.repoPath, 0o755); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}
	_, err := e.Run(ctx, "init")
	return err
}

// Clone clones a remote repository into the executor's repoPath.
func (e *Executor) Clone(ctx context.Context, remoteURL string) error {
	parent := filepath.Dir(e.repoPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "git", "clone", remoteURL, e.repoPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %s", stderr.String())
	}
	return nil
}

// IsRepo returns true if the repoPath contains a git repository.
func (e *Executor) IsRepo(ctx context.Context) bool {
	_, err := e.Run(ctx, "rev-parse", "--git-dir")
	return err == nil
}

// ---------------------------------------------------------------------------
// Branch operations
// ---------------------------------------------------------------------------

// Fetch runs git fetch for the given remote.
func (e *Executor) Fetch(ctx context.Context, remote string) error {
	_, err := e.Run(ctx, "fetch", remote)
	return err
}

// CheckoutBranch checks out an existing branch. Creates it from startPoint
// if it doesn't exist locally.
func (e *Executor) CheckoutBranch(ctx context.Context, branch string) error {
	_, err := e.Run(ctx, "checkout", branch)
	if err != nil {
		_, err = e.Run(ctx, "checkout", "-b", branch)
	}
	return err
}

// BranchExists returns true if the named branch exists locally.
func (e *Executor) BranchExists(ctx context.Context, branch string) bool {
	_, err := e.Run(ctx, "rev-parse", "--verify", branch)
	return err == nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func (e *Executor) CurrentBranch(ctx context.Context) (string, error) {
	r, err := e.Run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return r.Stdout, nil
}

// HeadSHA returns the SHA of the current HEAD.
func (e *Executor) HeadSHA(ctx context.Context) (string, error) {
	r, err := e.Run(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return r.Stdout, nil
}

// ---------------------------------------------------------------------------
// Commit operations
// ---------------------------------------------------------------------------

// WriteFile writes content to a file in the working tree (relative path).
func (e *Executor) WriteFile(filePath, content string) error {
	absPath := filepath.Join(e.repoPath, filePath)
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directories for %q: %w", filePath, err)
	}
	return os.WriteFile(absPath, []byte(content), 0o644)
}

// StageFiles stages the given file paths.
func (e *Executor) StageFiles(ctx context.Context, paths []string) error {
	args := append([]string{"add", "--"}, paths...)
	_, err := e.Run(ctx, args...)
	return err
}

// Commit creates a commit with the given message. Returns the new SHA.
func (e *Executor) Commit(ctx context.Context, message string) (string, error) {
	_, err := e.Run(ctx, "commit", "-m", message)
	if err != nil {
		return "", err
	}
	return e.HeadSHA(ctx)
}

// CommitWithAuthor creates a commit with explicit author info.
func (e *Executor) CommitWithAuthor(ctx context.Context, message, authorName, authorEmail string) (string, error) {
	author := fmt.Sprintf("%s <%s>", authorName, authorEmail)
	_, err := e.Run(ctx, "commit", "-m", message, "--author", author)
	if err != nil {
		return "", err
	}
	return e.HeadSHA(ctx)
}

// ---------------------------------------------------------------------------
// Push
// ---------------------------------------------------------------------------

// Push pushes the current branch to the given remote.
func (e *Executor) Push(ctx context.Context, remote, branch string) error {
	_, err := e.Run(ctx, "push", remote, branch)
	return err
}

// PushWithRetry attempts to push with simple retry logic.
func (e *Executor) PushWithRetry(ctx context.Context, remote, branch string, maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		lastErr = e.Push(ctx, remote, branch)
		if lastErr == nil {
			return nil
		}
		e.logger.Warn("push attempt failed, retrying",
			"attempt", i+1, "max", maxRetries+1, "error", lastErr)
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	return fmt.Errorf("push failed after %d attempts: %w", maxRetries+1, lastErr)
}

// ---------------------------------------------------------------------------
// File read
// ---------------------------------------------------------------------------

// ReadFile reads a file from the working tree.
func (e *Executor) ReadFile(filePath string) (string, error) {
	absPath := filepath.Join(e.repoPath, filePath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ShowFile reads a file at a specific commit/ref (git show ref:path).
func (e *Executor) ShowFile(ctx context.Context, ref, filePath string) (string, error) {
	r, err := e.Run(ctx, "show", ref+":"+filePath)
	if err != nil {
		return "", err
	}
	return r.Stdout, nil
}
