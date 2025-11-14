package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitOps struct {
	workDir   string
	repoPath  string
	owner     string
	repo      string
	token     string
}

func NewGitOps(workDir, owner, repo, token string) (*GitOps, error) {
	repoPath := filepath.Join(workDir, repo)
	
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	return &GitOps{
		workDir:  workDir,
		repoPath: repoPath,
		owner:    owner,
		repo:     repo,
		token:    token,
	}, nil
}

func (g *GitOps) Clone() error {
	// Remove existing directory if it exists
	if _, err := os.Stat(g.repoPath); err == nil {
		if err := os.RemoveAll(g.repoPath); err != nil {
			return fmt.Errorf("failed to remove existing repo: %w", err)
		}
	}

	// Clone with token authentication
	cloneURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", g.token, g.owner, g.repo)
	
	cmd := exec.Command("git", "clone", cloneURL, g.repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Configure git user for commits
	g.runGitCommand("config", "user.name", "Mr. Code Fixer")
	g.runGitCommand("config", "user.email", "code-fixer@automated.bot")

	return nil
}

func (g *GitOps) CreateBranch(branchName string) error {
	if err := g.runGitCommand("checkout", "-b", branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func (g *GitOps) CommitChanges(message string) error {
	// Add all changes
	if err := g.runGitCommand("add", "."); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit
	if err := g.runGitCommand("commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func (g *GitOps) Push(branchName string) error {
	if err := g.runGitCommand("push", "-u", "origin", branchName); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

func (g *GitOps) runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (g *GitOps) Cleanup() {
	// Optional: clean up the cloned repo
	// os.RemoveAll(g.repoPath)
}

type RepoContext struct {
	Structure string
	Files     map[string]string // path -> content
}

func (g *GitOps) GetRepoContext() (*RepoContext, error) {
	ctx := &RepoContext{
		Files: make(map[string]string),
	}

	// Get directory structure
	structure, err := g.getDirectoryStructure()
	if err != nil {
		return nil, err
	}
	ctx.Structure = structure

	// Read important files (limit to reasonable size)
	importantFiles := []string{
		"README.md",
		"package.json",
		"go.mod",
		"requirements.txt",
		"Cargo.toml",
		"pom.xml",
		"build.gradle",
	}

	for _, file := range importantFiles {
		filePath := filepath.Join(g.repoPath, file)
		if content, err := os.ReadFile(filePath); err == nil {
			ctx.Files[file] = string(content)
		}
	}

	// Read source files (limited depth and size)
	err = filepath.Walk(g.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || 
			   name == "vendor" || name == "target" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only read source code files up to 100KB
		if info.Size() > 100*1024 {
			return nil
		}

		ext := filepath.Ext(path)
		if isSourceFile(ext) {
			relPath, _ := filepath.Rel(g.repoPath, path)
			if content, err := os.ReadFile(path); err == nil {
				ctx.Files[relPath] = string(content)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (g *GitOps) getDirectoryStructure() (string, error) {
	var structure strings.Builder
	
	err := filepath.Walk(g.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(g.repoPath, path)
		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		
		if info.IsDir() {
			structure.WriteString(fmt.Sprintf("%s%s/\n", indent, info.Name()))
		} else {
			structure.WriteString(fmt.Sprintf("%s%s\n", indent, info.Name()))
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return structure.String(), nil
}

func isSourceFile(ext string) bool {
	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".java": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".rs": true, ".rb": true, ".php": true, ".cs": true, ".swift": true,
		".kt": true, ".scala": true, ".sh": true, ".bash": true,
		".html": true, ".css": true, ".scss": true, ".vue": true,
	}
	return sourceExts[ext]
}

type FileChange struct {
	FilePath string
	Content  string
}

func (g *GitOps) ApplyFileChange(change FileChange) error {
	fullPath := filepath.Join(g.repoPath, change.FilePath)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, []byte(change.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
