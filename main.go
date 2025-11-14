package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Config struct {
	RepoOwner    string
	RepoName     string
	GithubToken  string
	AIService    string // "groq" or "ollama"
	AIAPIKey     string // for Groq
	AIModel      string
	OllamaURL    string
	IssueNumber  int  // specific issue to fix, 0 = all open issues
	MaxIssues    int  // max number of issues to process
	WorkDir      string
}

func main() {
	config := Config{}

	// Define CLI flags
	flag.StringVar(&config.RepoOwner, "owner", "", "GitHub repository owner (required)")
	flag.StringVar(&config.RepoName, "repo", "", "GitHub repository name (required)")
	flag.StringVar(&config.GithubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token")
	flag.StringVar(&config.AIService, "ai-service", "groq", "AI service to use: groq or ollama")
	flag.StringVar(&config.AIAPIKey, "ai-key", os.Getenv("GROQ_API_KEY"), "API key for AI service (Groq)")
	flag.StringVar(&config.AIModel, "ai-model", "llama-3.3-70b-versatile", "AI model to use")
	flag.StringVar(&config.OllamaURL, "ollama-url", "http://localhost:11434", "Ollama API URL")
	flag.IntVar(&config.IssueNumber, "issue", 0, "Specific issue number to fix (0 = all open issues)")
	flag.IntVar(&config.MaxIssues, "max-issues", 1, "Maximum number of issues to process")
	flag.StringVar(&config.WorkDir, "work-dir", "./workspace", "Working directory for cloning repos")

	flag.Parse()

	// Validate required flags
	if config.RepoOwner == "" || config.RepoName == "" {
		fmt.Println("Error: -owner and -repo flags are required")
		flag.Usage()
		os.Exit(1)
	}

	if config.GithubToken == "" {
		fmt.Println("Error: GitHub token required (set GITHUB_TOKEN env var or use -github-token flag)")
		os.Exit(1)
	}

	if config.AIService == "groq" && config.AIAPIKey == "" {
		fmt.Println("Error: Groq API key required (set GROQ_API_KEY env var or use -ai-key flag)")
		os.Exit(1)
	}

	// Run the fixer
	if err := run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(config Config) error {
	fmt.Printf("Mr. Code Fixer starting...\n")
	fmt.Printf("Repository: %s/%s\n", config.RepoOwner, config.RepoName)
	fmt.Printf("AI Service: %s (model: %s)\n", config.AIService, config.AIModel)

	// Initialize GitHub client
	ghClient := NewGitHubClient(config.GithubToken, config.RepoOwner, config.RepoName)

	// Initialize AI client
	var aiClient AIClient
	var err error
	if config.AIService == "groq" {
		aiClient = NewGroqClient(config.AIAPIKey, config.AIModel)
	} else {
		aiClient = NewOllamaClient(config.OllamaURL, config.AIModel)
	}

	// Fetch issues
	var issues []Issue
	if config.IssueNumber > 0 {
		issue, err := ghClient.GetIssue(config.IssueNumber)
		if err != nil {
			return fmt.Errorf("failed to fetch issue #%d: %w", config.IssueNumber, err)
		}
		issues = []Issue{*issue}
	} else {
		issues, err = ghClient.GetOpenIssues(config.MaxIssues)
		if err != nil {
			return fmt.Errorf("failed to fetch issues: %w", err)
		}
	}

	if len(issues) == 0 {
		fmt.Println("No issues found to process")
		return nil
	}

	fmt.Printf("Found %d issue(s) to process\n\n", len(issues))

	// Process each issue
	for _, issue := range issues {
		fmt.Printf("=== Processing Issue #%d: %s ===\n", issue.Number, issue.Title)
		
		if err := processIssue(config, ghClient, aiClient, issue); err != nil {
			fmt.Printf("Failed to process issue #%d: %v\n\n", issue.Number, err)
			continue
		}
		
		fmt.Printf("Successfully processed issue #%d\n\n", issue.Number)
	}

	return nil
}

func processIssue(config Config, ghClient *GitHubClient, aiClient AIClient, issue Issue) error {
	// Clone repository
	gitOps, err := NewGitOps(config.WorkDir, config.RepoOwner, config.RepoName, config.GithubToken)
	if err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}
	defer gitOps.Cleanup()

	if err := gitOps.Clone(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// Create a new branch
	branchName := fmt.Sprintf("fix-issue-%d", issue.Number)
	if err := gitOps.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Read relevant files from the repository
	repoContext, err := gitOps.GetRepoContext()
	if err != nil {
		return fmt.Errorf("failed to read repo context: %w", err)
	}

	// Ask AI to analyze and fix the issue
	fmt.Println("Analyzing issue with AI...")
	fix, err := aiClient.AnalyzeAndFix(issue, repoContext)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	if len(fix.FileChanges) == 0 {
		return fmt.Errorf("AI did not suggest any file changes")
	}

	// Apply the changes
	fmt.Printf("Applying %d file change(s)...\n", len(fix.FileChanges))
	for _, change := range fix.FileChanges {
		if err := gitOps.ApplyFileChange(change); err != nil {
			return fmt.Errorf("failed to apply changes to %s: %w", change.FilePath, err)
		}
		fmt.Printf("  ✓ Modified %s\n", change.FilePath)
	}

	// Commit changes
	commitMsg := fmt.Sprintf("Fix #%d: %s\n\n%s", issue.Number, issue.Title, fix.Explanation)
	if err := gitOps.CommitChanges(commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push branch
	if err := gitOps.Push(branchName); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	// Create pull request
	prTitle := fmt.Sprintf("Fix #%d: %s", issue.Number, issue.Title)
	prBody := fmt.Sprintf("This PR automatically fixes issue #%d\n\n## Changes\n%s\n\n---\n*Generated by Mr. Code Fixer*",
		issue.Number, fix.Explanation)
	
	prURL, err := ghClient.CreatePullRequest(prTitle, prBody, branchName, "main")
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	fmt.Printf("✓ Pull request created: %s\n", prURL)

	return nil
}
