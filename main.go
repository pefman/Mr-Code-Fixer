package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	RepoOwner    string `json:"repo_owner"`
	RepoName     string `json:"repo_name"`
	RepoURL      string `json:"repo_url"`
	GithubToken  string `json:"github_token"`
	AIService    string `json:"ai_service"`
	AIAPIKey     string `json:"ai_api_key"`
	AIModel      string `json:"ai_model"`
	OllamaURL    string `json:"ollama_url"`
	WorkDir      string `json:"work_dir"`
}

func parseRepoURL(url string) (owner, repo string, err error) {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")
	
	// Handle various GitHub URL formats:
	// https://github.com/owner/repo
	// git@github.com:owner/repo
	// github.com/owner/repo
	
	if strings.Contains(url, "github.com") {
		// Extract owner/repo part
		parts := strings.Split(url, "github.com")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format")
		}
		
		path := strings.TrimPrefix(parts[1], "/")
		path = strings.TrimPrefix(path, ":")
		
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			return "", "", fmt.Errorf("invalid repository path")
		}
		
		return pathParts[0], pathParts[1], nil
	}
	
	return "", "", fmt.Errorf("only GitHub repositories are supported")
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".mr-code-fixer.json")
}

func getDefaultWorkDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./workspace"
	}
	return filepath.Join(homeDir, ".mr-code-fixer", "workspace")
}

func loadConfig() Config {
	config := Config{
		AIService: "groq",
		AIModel:   "llama-3.3-70b-versatile",
		OllamaURL: "http://localhost:11434",
		WorkDir:   getDefaultWorkDir(),
	}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	return config
}

func saveConfig(config Config) error {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

func prompt(label string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s: ", label)
	}
	
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		return defaultValue
	}
	return input
}

func promptSecret(label string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [****]: ", label)
	} else {
		fmt.Printf("%s: ", label)
	}
	
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

func promptWithOptions(label string, options []string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	
	// Find default in options
	defaultIndex := 0
	for i, opt := range options {
		if opt == defaultValue {
			defaultIndex = i + 1
			break
		}
	}
	
	if defaultIndex > 0 {
		fmt.Printf("%s (1-%d) [%d]: ", label, len(options), defaultIndex)
	} else {
		fmt.Printf("%s (1-%d) [1]: ", label, len(options))
	}
	
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		if defaultIndex > 0 {
			return options[defaultIndex-1]
		}
		return options[0]
	}
	
	// Try to parse as number
	num, err := strconv.Atoi(input)
	if err == nil && num > 0 && num <= len(options) {
		return options[num-1]
	}
	
	// Or match by name
	for _, opt := range options {
		if strings.EqualFold(opt, input) {
			return opt
		}
	}
	
	return options[0]
}

func selectIssue(issues []Issue) *Issue {
	if len(issues) == 0 {
		return nil
	}

	fmt.Println("\n=== Available Issues ===")
	for i, issue := range issues {
		fmt.Printf("%d. #%d - %s\n", i+1, issue.Number, issue.Title)
		if len(issue.Body) > 100 {
			fmt.Printf("   %s...\n", issue.Body[:100])
		} else if issue.Body != "" {
			fmt.Printf("   %s\n", issue.Body)
		}
		fmt.Println()
	}

	for {
		choice := prompt(fmt.Sprintf("Select issue to fix (1-%d, or 0 to fix all)", len(issues)), "1")
		
		num, err := strconv.Atoi(choice)
		if err != nil || num < 0 || num > len(issues) {
			fmt.Println("Invalid selection. Please try again.")
			continue
		}

		if num == 0 {
			return nil // Fix all
		}

		return &issues[num-1]
	}
}

func interactiveSetup() Config {
	fmt.Println("=== Mr. Code Fixer - Interactive Setup ===\n")
	
	config := loadConfig()

	fmt.Println("GitHub Repository:")
	repoInput := prompt("Repository URL or owner/repo", config.RepoURL)
	
	// Try to parse as URL first, then fall back to owner/repo format
	if strings.Contains(repoInput, "github.com") || strings.Contains(repoInput, "/") {
		if strings.Contains(repoInput, "github.com") {
			// It's a URL
			config.RepoURL = repoInput
			owner, repo, err := parseRepoURL(repoInput)
			if err != nil {
				fmt.Printf("Warning: Could not parse URL: %v\n", err)
				config.RepoOwner = prompt("Repository Owner", config.RepoOwner)
				config.RepoName = prompt("Repository Name", config.RepoName)
			} else {
				config.RepoOwner = owner
				config.RepoName = repo
			}
		} else {
			// It's owner/repo format
			parts := strings.Split(repoInput, "/")
			if len(parts) == 2 {
				config.RepoOwner = parts[0]
				config.RepoName = parts[1]
				config.RepoURL = fmt.Sprintf("https://github.com/%s/%s", parts[0], parts[1])
			} else {
				config.RepoOwner = prompt("Repository Owner", config.RepoOwner)
				config.RepoName = prompt("Repository Name", config.RepoName)
			}
		}
	} else {
		config.RepoOwner = prompt("Repository Owner", config.RepoOwner)
		config.RepoName = prompt("Repository Name", config.RepoName)
	}
	
	config.GithubToken = promptSecret("GitHub Token", config.GithubToken)

	fmt.Println("\nAI Service Settings:")
	config.AIService = prompt("AI Service (chatgpt/grok/ollama)", config.AIService)
	
	if config.AIService == "chatgpt" || config.AIService == "openai" {
		config.AIAPIKey = promptSecret("OpenAI API Key", config.AIAPIKey)
		
		// Fetch available models
		fmt.Println("Fetching available models...")
		client := NewOpenAIClient(config.AIAPIKey, "")
		models, err := client.GetAvailableModels()
		if err == nil && len(models) > 0 {
			fmt.Println("Available models:")
			for i, model := range models {
				fmt.Printf("  %d. %s\n", i+1, model)
			}
			config.AIModel = promptWithOptions("Select model", models, config.AIModel)
		} else {
			config.AIModel = prompt("AI Model", config.AIModel)
		}
	} else if config.AIService == "grok" {
		config.AIAPIKey = promptSecret("Grok API Key", config.AIAPIKey)
		
		// Fetch available models
		fmt.Println("Fetching available models...")
		client := NewXAIClient(config.AIAPIKey, "")
		models, err := client.GetAvailableModels()
		if err == nil && len(models) > 0 {
			fmt.Println("Available models:")
			for i, model := range models {
				fmt.Printf("  %d. %s\n", i+1, model)
			}
			config.AIModel = promptWithOptions("Select model", models, config.AIModel)
		} else {
			config.AIModel = prompt("AI Model (grok-beta)", "grok-beta")
		}
	} else {
		config.OllamaURL = prompt("Ollama URL", config.OllamaURL)
		
		// Fetch available models
		fmt.Println("Fetching available local models...")
		client := NewOllamaClient(config.OllamaURL, "")
		models, err := client.GetAvailableModels()
		if err == nil && len(models) > 0 {
			fmt.Println("Available models:")
			for i, model := range models {
				fmt.Printf("  %d. %s\n", i+1, model)
			}
			config.AIModel = promptWithOptions("Select model", models, config.AIModel)
		} else {
			config.AIModel = prompt("AI Model", config.AIModel)
		}
	}

	fmt.Println("\nWorking Directory:")
	fmt.Printf("  (Repos will be cloned to: %s/<owner>/<repo>)\n", config.WorkDir)
	config.WorkDir = prompt("Work Directory", config.WorkDir)

	// Save config for next time
	if err := saveConfig(config); err != nil {
		fmt.Printf("Warning: Could not save config: %v\n", err)
	} else {
		fmt.Printf("\nConfiguration saved to: %s\n", getConfigPath())
	}

	return config
}

func parseFlags(config *Config) {
	var repoURL string
	flag.StringVar(&repoURL, "repo-url", "", "GitHub repository URL (e.g., https://github.com/owner/repo)")
	flag.StringVar(&config.RepoOwner, "owner", config.RepoOwner, "GitHub repository owner")
	flag.StringVar(&config.RepoName, "repo", config.RepoName, "GitHub repository name")
	flag.StringVar(&config.GithubToken, "github-token", config.GithubToken, "GitHub personal access token")
	flag.StringVar(&config.AIService, "ai-service", config.AIService, "AI service to use: chatgpt/grok/ollama")
	flag.StringVar(&config.AIAPIKey, "ai-key", config.AIAPIKey, "API key for AI service")
	flag.StringVar(&config.AIModel, "ai-model", config.AIModel, "AI model to use")
	flag.StringVar(&config.OllamaURL, "ollama-url", config.OllamaURL, "Ollama API URL")
	flag.StringVar(&config.WorkDir, "work-dir", config.WorkDir, "Working directory for cloning repos")

	flag.Parse()

	// If repo URL provided, parse it
	if repoURL != "" {
		config.RepoURL = repoURL
		owner, repo, err := parseRepoURL(repoURL)
		if err == nil {
			config.RepoOwner = owner
			config.RepoName = repo
		}
	}

	// Override from env vars if not set via flags
	if config.GithubToken == "" {
		config.GithubToken = os.Getenv("GITHUB_TOKEN")
	}
	if config.AIAPIKey == "" {
		config.AIAPIKey = os.Getenv("GROQ_API_KEY")
	}
}

func validateConfig(config Config) error {
	if config.RepoOwner == "" || config.RepoName == "" {
		return fmt.Errorf("repository owner and name are required")
	}
	if config.GithubToken == "" {
		return fmt.Errorf("GitHub token is required")
	}
	if (config.AIService == "chatgpt" || config.AIService == "openai" || config.AIService == "grok") && config.AIAPIKey == "" {
		return fmt.Errorf("%s API key is required", config.AIService)
	}
	return nil
}

func main() {
	// Check if running in interactive mode
	interactive := len(os.Args) == 1

	var config Config
	
	if interactive {
		config = interactiveSetup()
	} else {
		// Load saved config as defaults
		config = loadConfig()
		
		// Parse command line flags to override config
		parseFlags(&config)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Run the fixer
	if err := run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(config Config) error {
	fmt.Printf("\n=== Mr. Code Fixer ===\n")
	fmt.Printf("Repository: %s/%s\n", config.RepoOwner, config.RepoName)
	fmt.Printf("AI Service: %s (model: %s)\n\n", config.AIService, config.AIModel)

	// Initialize GitHub client
	ghClient := NewGitHubClient(config.GithubToken, config.RepoOwner, config.RepoName)

	// Initialize AI client
	var aiClient AIClient
	if config.AIService == "chatgpt" || config.AIService == "openai" {
		aiClient = NewOpenAIClient(config.AIAPIKey, config.AIModel)
	} else if config.AIService == "grok" {
		aiClient = NewXAIClient(config.AIAPIKey, config.AIModel)
	} else {
		aiClient = NewOllamaClient(config.OllamaURL, config.AIModel)
	}

	// Fetch all open issues
	fmt.Println("Fetching open issues...")
	issues, err := ghClient.GetOpenIssues(100) // Get up to 100 issues
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No open issues found.")
		return nil
	}

	// Let user select which issue(s) to fix
	selectedIssue := selectIssue(issues)

	var issuesToProcess []Issue
	if selectedIssue == nil {
		// User chose to fix all
		confirm := prompt(fmt.Sprintf("Fix all %d issues? (yes/no)", len(issues)), "no")
		if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
		issuesToProcess = issues
	} else {
		issuesToProcess = []Issue{*selectedIssue}
	}

	// Process each issue
	for _, issue := range issuesToProcess {
		fmt.Printf("\n=== Processing Issue #%d: %s ===\n", issue.Number, issue.Title)
		
		if err := processIssue(config, ghClient, aiClient, issue); err != nil {
			fmt.Printf("Failed to process issue #%d: %v\n\n", issue.Number, err)
			
			if len(issuesToProcess) > 1 {
				cont := prompt("Continue with next issue? (yes/no)", "yes")
				if strings.ToLower(cont) != "yes" && strings.ToLower(cont) != "y" {
					return fmt.Errorf("stopped processing issues")
				}
			}
			continue
		}
		
		fmt.Printf("✓ Successfully processed issue #%d\n", issue.Number)
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

	// Read relevant files from the repository
	repoContext, err := gitOps.GetRepoContext(issue.Title, issue.Body)
	if err != nil {
		return fmt.Errorf("failed to read repo context: %w", err)
	}
	
	fmt.Printf("Analyzed %d relevant files from repository\n", repoContext.FileCount)

	// Ask AI to analyze and fix the issue
	fmt.Println("Analyzing issue with AI...")
	fix, err := aiClient.AnalyzeAndFix(issue, repoContext)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	// Check if AI needs more information
	if fix.NeedsMoreInfo && len(fix.Questions) > 0 {
		fmt.Println("\n⚠ AI needs more information to fix this issue.")
		fmt.Println("Posting questions to the issue...")
		
		questionComment := "I need some clarification to fix this issue:\n\n"
		for i, q := range fix.Questions {
			questionComment += fmt.Sprintf("%d. %s\n", i+1, q)
		}
		questionComment += "\nPlease provide more details so I can create a proper fix.\n\n---\n*Asked by Mr. Code Fixer*"
		
		if err := ghClient.AddIssueComment(issue.Number, questionComment); err != nil {
			return fmt.Errorf("failed to post questions: %w", err)
		}
		
		fmt.Printf("✓ Posted %d question(s) to issue #%d\n", len(fix.Questions), issue.Number)
		return nil
	}

	if len(fix.FileChanges) == 0 {
		return fmt.Errorf("AI did not suggest any file changes")
	}

	// Create a branch with sanitized issue title
	branchName := createBranchName(issue)
	if err := gitOps.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
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
	confidenceNote := ""
	if fix.Confidence == "high" {
		confidenceNote = "✅ **High confidence** - This fix should resolve the issue."
	} else if fix.Confidence == "medium" {
		confidenceNote = "⚠️ **Medium confidence** - Please review carefully."
	} else {
		confidenceNote = "⚠️ **Low confidence** - This is a best attempt, please review thoroughly."
	}
	
	prBody := fmt.Sprintf(`Fixes #%d

%s

## Changes
%s

---
*Generated by Mr. Code Fixer*`,
		issue.Number, confidenceNote, fix.Explanation)
	
	prURL, err := ghClient.CreatePullRequest(prTitle, prBody, branchName, gitOps.DefaultBranch)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	fmt.Printf("✓ Pull request created: %s\n", prURL)

	// If high confidence, close the issue with a comment
	if fix.Confidence == "high" {
		fmt.Println("Closing issue (high confidence fix)...")
		closeComment := fmt.Sprintf("Fixed in %s\n\n✅ A pull request has been created with a high-confidence fix for this issue.", prURL)
		
		if err := ghClient.AddIssueComment(issue.Number, closeComment); err != nil {
			fmt.Printf("Warning: Could not add closing comment: %v\n", err)
		}
		
		if err := ghClient.CloseIssue(issue.Number); err != nil {
			fmt.Printf("Warning: Could not close issue: %v\n", err)
		} else {
			fmt.Printf("✓ Issue #%d closed\n", issue.Number)
		}
	}

	return nil
}

func createBranchName(issue Issue) string {
	// Sanitize issue title for branch name
	title := strings.ToLower(issue.Title)
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ReplaceAll(title, "?", "")
	title = strings.ReplaceAll(title, "!", "")
	title = strings.ReplaceAll(title, ".", "")
	title = strings.ReplaceAll(title, ",", "")
	title = strings.ReplaceAll(title, "'", "")
	title = strings.ReplaceAll(title, "\"", "")
	
	// Limit length
	if len(title) > 40 {
		title = title[:40]
	}
	
	return fmt.Sprintf("fix/%d-%s", issue.Number, title)
}
