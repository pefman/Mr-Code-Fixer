package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitOps struct {
	workDir       string
	repoPath      string
	owner         string
	repo          string
	token         string
	DefaultBranch string
}

func NewGitOps(workDir, owner, repo, token string) (*GitOps, error) {
	// Create a unique directory path for this repo
	repoPath := filepath.Join(workDir, owner, repo)
	
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
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

	// Detect default branch
	cmd = exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = g.repoPath
	output, err := cmd.CombinedOutput()
	if err == nil {
		// Output format: refs/remotes/origin/branch-name
		branch := strings.TrimSpace(string(output))
		parts := strings.Split(branch, "/")
		if len(parts) > 0 {
			g.DefaultBranch = parts[len(parts)-1]
		}
	}
	if g.DefaultBranch == "" {
		// Fallback to main
		g.DefaultBranch = "main"
	}

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
	FileCount int               // Total files analyzed
}

type fileScore struct {
	path  string
	score int
}

func (g *GitOps) GetRepoContext(issueTitle, issueBody string) (*RepoContext, error) {
	ctx := &RepoContext{
		Files: make(map[string]string),
	}

	// Get directory structure
	structure, err := g.getDirectoryStructure()
	if err != nil {
		return nil, err
	}
	ctx.Structure = structure

	// Extract file mentions and keywords from issue
	mentionedFiles := extractFileMentions(issueTitle + " " + issueBody)
	keywords := extractKeywords(issueTitle + " " + issueBody)

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

	// Collect all source files with relevance scores
	var scoredFiles []fileScore

	err = filepath.Walk(g.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || 
			   name == "vendor" || name == "target" || name == "dist" || name == "build" ||
			   name == "test" || name == "tests" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only consider source code files up to 100KB
		if info.Size() > 100*1024 {
			return nil
		}

		ext := filepath.Ext(path)
		if isSourceFile(ext) {
			relPath, _ := filepath.Rel(g.repoPath, path)
			
			// Calculate relevance score
			score := calculateRelevance(relPath, mentionedFiles, keywords)
			if score > 0 {
				scoredFiles = append(scoredFiles, fileScore{relPath, score})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by relevance and take top 30 files
	sortFilesByScore(scoredFiles)
	maxFiles := 30
	if len(scoredFiles) > maxFiles {
		scoredFiles = scoredFiles[:maxFiles]
	}

	// Read the selected files
	for _, sf := range scoredFiles {
		filePath := filepath.Join(g.repoPath, sf.path)
		if content, err := os.ReadFile(filePath); err == nil {
			ctx.Files[sf.path] = string(content)
		}
	}

	ctx.FileCount = len(ctx.Files)
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

// extractFileMentions finds file paths mentioned in the issue text
func extractFileMentions(text string) []string {
	var files []string
	text = strings.ToLower(text)
	
	// Simple pattern: words with file extensions
	words := strings.Fields(text)
	for _, word := range words {
		word = strings.Trim(word, "`,\"'()[]")
		if strings.Contains(word, "/") || strings.Contains(word, "\\") {
			// Looks like a path
			for _, ext := range []string{".go", ".js", ".ts", ".py", ".java", ".rb", ".php", ".tsx", ".jsx"} {
				if strings.HasSuffix(word, ext) {
					files = append(files, word)
					break
				}
			}
		}
	}
	
	return files
}

// extractKeywords pulls important words from the issue
func extractKeywords(text string) []string {
	text = strings.ToLower(text)
	
	// Remove common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "is": true, "are": true, "was": true, "were": true, "been": true,
		"be": true, "have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "should": true, "could": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "please": true,
		"help": true, "need": true, "want": true, "issue": true, "problem": true,
	}
	
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})
	
	var keywords []string
	for _, word := range words {
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// calculateRelevance scores a file based on mentions and keywords
func calculateRelevance(filePath string, mentionedFiles, keywords []string) int {
	score := 0
	lowerPath := strings.ToLower(filePath)
	
	// Exact file mention = very high score
	for _, mentioned := range mentionedFiles {
		if strings.Contains(lowerPath, strings.ToLower(mentioned)) {
			score += 100
		}
	}
	
	// Keyword in path = medium score
	for _, keyword := range keywords {
		if strings.Contains(lowerPath, keyword) {
			score += 10
		}
	}
	
	// If no matches yet, give small score to recently modified or common entry points
	if score == 0 {
		// Favor main entry points
		if strings.Contains(lowerPath, "main") || strings.Contains(lowerPath, "index") ||
		   strings.Contains(lowerPath, "app") || strings.Contains(lowerPath, "server") {
			score += 5
		}
		// Give base score to all source files
		score += 1
	}
	
	return score
}

// sortFilesByScore sorts files by relevance score (highest first)
func sortFilesByScore(files []fileScore) {
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].score > files[i].score {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
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
