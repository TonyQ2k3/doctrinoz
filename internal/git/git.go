package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type State struct {
	LastCommit  string            `json:"last_commit"`
	GeneratedAt time.Time         `json:"generated_at"`
	PageCommits map[string]string `json:"page_commits"`
}

// ChanedFiles returns a list of files modified between fromCommit and HEAD
// If fromCommit is empty it will return all tracked files.
func ChangedFiles(repoPath, fromCommit string) ([]string, error) {
	repo, err := gogit.PlainOpenWithOptions(repoPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Error opening repo: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("Error reading HEAD: %w", err)
	}

	headCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("Error reading HEAD commit: %w", err)
	}
	if fromCommit == "" {
		return allTrackedFiles(headCommit)
	}

	fromHash := plumbing.NewHash(fromCommit)
	fromCommitObj, err := repo.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("Error reading base commit %s: %w", fromCommit, err)
	}

	return diffFiles(fromCommitObj, headCommit)
}

// HeadCommit returns the current HEAD commit hash
func HeadCommit(repoPath string) (string, error) {
	repo, err := gogit.PlainOpenWithOptions(repoPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("Error opening repo: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("Error reading HEAD: %w", err)
	}
	return head.Hash().String(), nil
}

// Top-level info about repo
type RepoMetadata struct {
	Name          string
	Description   string
	Languages     map[string]int
	TotalFiles    int
	RecentAuthors []string
}

// Does lightweight scan of repo
func GatherMetadata(repoPath string) (*RepoMetadata, error) {
	repo, err := gogit.PlainOpenWithOptions(repoPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Error opening repo: %w", err)
	}

	meta := &RepoMetadata{
		Name:      filepath.Base(repoPath),
		Languages: make(map[string]int),
	}

	head, err := repo.Head()
	if err != nil {
		return meta, nil
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return meta, nil
	}

	tree, err := commit.Tree()
	if err != nil {
		return meta, nil
	}

	// Count files by extensions and collect authors
	authorSet := map[string]struct{}{}
	tree.Files().ForEach(func(f *object.File) error {
		meta.TotalFiles++
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext != "" {
			meta.Languages[ext]++
		}
		return nil
	})

	logOpts := &gogit.LogOptions{From: head.Hash()}
	iter, err := repo.Log(logOpts)
	if err == nil {
		count := 0
		iter.ForEach(func(c *object.Commit) error {
			if count >= 20 {
				return fmt.Errorf("stop")
			}
			authorSet[c.Author.Name] = struct{}{}
			count++
			return nil
		})
		for a := range authorSet {
			meta.RecentAuthors = append(meta.RecentAuthors, a)
		}
	}
	// Read README
	readmeFile, _ := tree.File("README.md")
	if readmeFile == nil {
		readmeFile, _ = tree.File("README")
	}
	if readmeFile != nil {
		content, err := readmeFile.Contents()
		if err == nil {
			for _, line := range strings.Split(content, "\n") {
				line = strings.TrimSpace(strings.TrimLeft(line, "#"))
				if line != "" {
					meta.Description = line
					break
				}
			}
		}
	}

	return meta, nil
}

// Check if a directory is a repo
func IsGitRepo(path string) bool {
	_, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	return err == nil
}

// --------------- Helpers ------------------

func allTrackedFiles(commit *object.Commit) ([]string, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []string
	tree.Files().ForEach(func(f *object.File) error {
		files = append(files, f.Name)
		return nil
	})
	return files, nil
}

func diffFiles(from, to *object.Commit) ([]string, error) {
	fromTree, err := from.Tree()
	if err != nil {
		return nil, err
	}
	toTree, err := to.Tree()
	if err != nil {
		return nil, err
	}
	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	for _, change := range changes {
		if change.To.Name != "" {
			seen[change.To.Name] = struct{}{}
		}
		if change.From.Name != "" {
			seen[change.From.Name] = struct{}{}
		}
	}

	files := make([]string, 0, len(seen))
	for f := range seen {
		files = append(files, f)
	}
	return files, nil
}

// ReadFile read a file from repo working tree
func ReadFile(repoPath, filePath string) (string, error) {
	full := filepath.Join(repoPath, filePath)
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
