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

const Version = "v1.3.5"

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

	for i, issue := range issues {
		fmt.Printf("  \033[1;36m%d.\033[0m \033[1m#%d\033[0m - %s\n", i+1, issue.Number, issue.Title)
		if len(issue.Body) > 80 {
			fmt.Printf("     \033[90m%s...\033[0m\n", issue.Body[:80])
		} else if issue.Body != "" {
			fmt.Printf("     \033[90m%s\033[0m\n", issue.Body)
		}
		fmt.Println()
	}

	for {
		fmt.Printf("\n\033[1m→\033[0m Select issue to fix (\033[36m1-%d\033[0m, or \033[33m0\033[0m to fix all) [\033[32m1\033[0m]: ", len(issues))
		choice := prompt("", "1")
		
		num, err := strconv.Atoi(choice)
		if err != nil || num < 0 || num > len(issues) {
			fmt.Println("\033[31m✗\033[0m Invalid selection. Please try again.")
			continue
		}

		if num == 0 {
			return nil // Fix all
		}

		return &issues[num-1]
	}
}

func selectIssueWithSettings(issues []Issue, config *Config, analytics *SessionAnalytics) *Issue {
	if len(issues) == 0 {
		return nil
	}

	fmt.Println()
	for i, issue := range issues {
		fmt.Printf("  \033[1;36m%d.\033[0m \033[1m#%d\033[0m - %s\n", i+1, issue.Number, issue.Title)
		if len(issue.Body) > 80 {
			fmt.Printf("     \033[90m%s...\033[0m\n", issue.Body[:80])
		} else if issue.Body != "" {
			fmt.Printf("     \033[90m%s\033[0m\n", issue.Body)
		}
		fmt.Println()
	}

	for {
		fmt.Printf("\n\033[1m→\033[0m Select issue (\033[36m1-%d\033[0m, \033[33m0\033[0m=fix all, \033[35mS\033[0m=settings, \033[90mQ\033[0m=quit) [\033[32m1\033[0m]: ", len(issues))
		choice := strings.ToLower(strings.TrimSpace(prompt("", "1")))
		
		// Handle special commands
		if choice == "s" {
			*config = interactiveSetup()
			fmt.Println("\n\033[32m✓\033[0m Settings updated. Please restart the application.")
			return nil
		}
		
		if choice == "q" {
			fmt.Println("Exiting...")
			return nil
		}
		
		num, err := strconv.Atoi(choice)
		if err != nil || num < 0 || num > len(issues) {
			fmt.Println("\033[31m✗\033[0m Invalid selection. Please try again.")
			continue
		}

		if num == 0 {
			// Return a special marker for "fix all"
			return &Issue{Number: -1}
		}

		return &issues[num-1]
	}
}

func interactiveSetup() Config {
	fmt.Println("=== Mr. Code Fixer - Interactive Setup ===")
	fmt.Println()
	
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
		// Check if config exists
		configPath := getConfigPath()
		if _, err := os.Stat(configPath); err == nil {
			// Config exists - just load it
			config = loadConfig()
		} else {
			// No config - run full setup
			config = interactiveSetup()
		}
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
	// Show welcome banner
	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║         🤖 Mr. Code Fixer - Ready to Help! %-19s║\n", Version)
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n📦 Repository: \033[1m%s/%s\033[0m", config.RepoOwner, config.RepoName)
	fmt.Printf("\n🧠 AI Service: \033[1m%s\033[0m (model: \033[36m%s\033[0m)\n\n", config.AIService, config.AIModel)

	// Initialize analytics
	analytics := NewSessionAnalytics()

	// Initialize GitHub client
	ghClient := NewGitHubClient(config.GithubToken, config.RepoOwner, config.RepoName)

	// Initialize AI client with analytics
	var aiClient AIClient
	if config.AIService == "chatgpt" || config.AIService == "openai" {
		client := NewOpenAIClient(config.AIAPIKey, config.AIModel)
		client.SetAnalytics(analytics)
		aiClient = client
	} else if config.AIService == "grok" {
		client := NewXAIClient(config.AIAPIKey, config.AIModel)
		client.SetAnalytics(analytics)
		aiClient = client
	} else {
		client := NewOllamaClient(config.OllamaURL, config.AIModel)
		client.SetAnalytics(analytics)
		aiClient = client
	}

	// Fetch all open issues
	fmt.Print("🔍 Fetching open issues")
	for i := 0; i < 3; i++ {
		fmt.Print(".")
	}
	fmt.Println()
	issues, err := ghClient.GetOpenIssues(100) // Get up to 100 issues
	if err != nil {
		fmt.Printf("\n\033[31m✗ Error fetching issues:\033[0m %v\n\n", err)
		
		// Offer to review settings
		fmt.Println("This might be due to incorrect configuration.")
		response := prompt("Would you like to review settings? (yes/no)", "yes")
		if strings.ToLower(response) == "yes" || strings.ToLower(response) == "y" {
			config = interactiveSetup()
			// Retry with new config
			return run(config)
		}
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No open issues found.")
		return nil
	}

	// Filter out issues the bot has already responded to
	fmt.Print("📝 Loading issues")
	for i := 0; i < 3; i++ {
		fmt.Print(".")
	}
	fmt.Println("\n")
	
	var unhandledIssues []Issue
	for _, issue := range issues {
		comments, err := ghClient.GetIssueComments(issue.Number)
		if err != nil {
			// If we can't check, include it to be safe
			unhandledIssues = append(unhandledIssues, issue)
			continue
		}
		
		// Check if bot's comment is the last one
		// If there are new comments after bot's response, process the issue again
		needsProcessing := true
		lastBotCommentIndex := -1
		
		// Find the last bot comment
		for i, comment := range comments {
			if strings.Contains(comment.Body, "Mr. Code Fixer") || 
			   strings.Contains(comment.Body, "🤖") {
				lastBotCommentIndex = i
			}
		}
		
		// If bot commented and it's still the last comment, skip
		if lastBotCommentIndex != -1 && lastBotCommentIndex == len(comments)-1 {
			needsProcessing = false
		}
		
		if needsProcessing {
			unhandledIssues = append(unhandledIssues, issue)
		}
	}
	
	if len(unhandledIssues) == 0 {
		fmt.Println("\n✓ All open issues have already been handled by the bot!")
		return nil
	}
	
	if len(issues) != len(unhandledIssues) {
		fmt.Printf("✓ Found %d new issue(s) (skipped %d already handled)\n", 
			len(unhandledIssues), len(issues)-len(unhandledIssues))
	}

	fmt.Printf("\n\033[1m📦 %s/%s\033[0m\n", config.RepoOwner, config.RepoName)

	// Let user select which issue(s) to fix (with settings option)
	selectedIssue := selectIssueWithSettings(unhandledIssues, &config, analytics)
	
	// If user chose settings, the config has been updated and we should restart
	if selectedIssue == nil {
		return nil // User chose to exit or settings were changed
	}

	var issuesToProcess []Issue
	if selectedIssue.Number == -1 {
		// Special case: user chose to fix all
		analytics.PrintCostEstimate(len(unhandledIssues), config.AIService)
		
		confirm := prompt(fmt.Sprintf("Fix all %d issues? (yes/no)", len(unhandledIssues)), "no")
		if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
		issuesToProcess = unhandledIssues
	} else {
		issuesToProcess = []Issue{*selectedIssue}
	}

	// Process each issue
	fmt.Println("\n" + strings.Repeat("─", 66))
	for _, issue := range issuesToProcess {
		fmt.Printf("\n\n🔧 Processing Issue #%d: \033[1m%s\033[0m\n", issue.Number, issue.Title)
		fmt.Println(strings.Repeat("─", 66))
		
		if err := processIssue(config, ghClient, aiClient, issue, analytics); err != nil {
			fmt.Printf("Failed to process issue #%d: %v\n\n", issue.Number, err)
			
			if len(issuesToProcess) > 1 {
				cont := prompt("Continue with next issue? (yes/no)", "yes")
				if strings.ToLower(cont) != "yes" && strings.ToLower(cont) != "y" {
					analytics.PrintSummary()
					return fmt.Errorf("stopped processing issues")
				}
			}
			continue
		}
		
		fmt.Printf("✓ Successfully processed issue #%d\n", issue.Number)
	}

	// Print session summary
	fmt.Println("\n" + strings.Repeat("═", 66))
	analytics.PrintSummary()

	return nil
}

func processIssue(config Config, ghClient *GitHubClient, aiClient AIClient, issue Issue, analytics *SessionAnalytics) error {
	// Check if issue is too vague before processing
	if isIssueTooVague(issue) {
		fmt.Println("\n⚠ Issue description is too vague to fix automatically.")
		fmt.Println("Posting request for more details...")
		
		questionComment := `## ❓ Need More Information

Hi! I'd love to help fix this issue, but I need more details to understand what's wrong.

Please provide:

1. **What's the expected behavior?** What should happen?
2. **What's the actual behavior?** What's currently happening instead?
3. **Steps to reproduce:** How can I see this problem?
4. **Any error messages?** Copy-paste any errors from console/logs
5. **Which file(s) are affected?** (e.g., src/main.js or components/Login.tsx)

The more details you provide, the better I can help! 🙏

---

<sub>🤖 Mr. Code Fixer - I need clear information to create good fixes</sub>`
		
		if err := ghClient.AddIssueComment(issue.Number, questionComment); err != nil {
			return fmt.Errorf("failed to post comment: %w", err)
		}
		
		analytics.RecordQuestionAsked()
		fmt.Printf("✓ Posted request for more information on issue #%d\n", issue.Number)
		return nil
	}

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
		
		analytics.RecordQuestionAsked()
		fmt.Printf("✓ Posted %d question(s) to issue #%d\n", len(fix.Questions), issue.Number)
		return nil
	}

	// Check if AI determined this is not a code fix (e.g., question, discussion, etc.)
	if len(fix.FileChanges) == 0 {
		fmt.Println("\n💬 This issue doesn't require code changes.")
		
		responseComment := fmt.Sprintf(`## 💬 Response

%s

This issue appears to be a question or discussion rather than a bug or feature requiring code changes. If you need specific code modifications, please provide more details about what changes you'd like to see.

---

<sub>🤖 Mr. Code Fixer</sub>`, fix.Explanation)
		
		if err := ghClient.AddIssueComment(issue.Number, responseComment); err != nil {
			return fmt.Errorf("failed to post response: %w", err)
		}
		
		// Close the issue since we've responded
		if err := ghClient.CloseIssue(issue.Number); err != nil {
			fmt.Printf("Warning: Could not close issue: %v\n", err)
		} else {
			fmt.Printf("✓ Issue #%d closed\n", issue.Number)
		}
		
		analytics.RecordIssueHandled()
		fmt.Printf("✓ Posted response explaining no code changes needed\n")
		return nil
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

	// Run tests if available
	fmt.Println("\n🧪 Checking for tests...")
	testRunner := NewTestRunner(gitOps.repoPath)
	testResult := testRunner.Execute()
	
	if testResult.Command != "" {
		fmt.Printf("Found test command: %s\n", testResult.Command)
		
		if !testResult.Passed {
			fmt.Println("\n❌ Tests failed! Not creating PR.")
			fmt.Println("Test output:")
			fmt.Println(testResult.Output)
			
			// Rollback by not proceeding - cleanup will happen via defer
			return fmt.Errorf("tests failed after applying changes")
		}
		fmt.Println("✓ All tests passed!")
	} else {
		fmt.Println("No tests detected - proceeding without test validation")
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

	// Create pull request with detailed technical description
	prTitle := fmt.Sprintf("Fix #%d: %s", issue.Number, issue.Title)
	confidenceNote := ""
	if fix.Confidence == "high" {
		confidenceNote = "✅ **High confidence** - This fix should resolve the issue."
	} else if fix.Confidence == "medium" {
		confidenceNote = "⚠️ **Medium confidence** - Please review carefully."
	} else {
		confidenceNote = "⚠️ **Low confidence** - This is a best attempt, please review thoroughly."
	}
	
	// Build detailed file changes list
	fileChangesList := ""
	for _, change := range fix.FileChanges {
		fileChangesList += fmt.Sprintf("- `%s`\n", change.FilePath)
	}
	
	// Add test results to PR body
	testSection := ""
	if testResult.Command != "" {
		if testResult.Passed {
			testSection = "\n### ✅ Tests Passed\n\nAll existing tests passed after applying the changes.\n"
		}
	}
	
	prBody := fmt.Sprintf(`## 🔧 Automated Fix

Fixes #%d

**Confidence Level:** %s

### 📋 Analysis

%s

### 🔨 Technical Details

This PR addresses the issue by making targeted changes to the codebase. The modifications were determined through analysis of the repository structure, issue description, and relevant code context.

**Modified Files:**
%s
**Approach:**
The fix was generated by analyzing the issue requirements and applying best practices for the detected programming language and framework. All changes maintain backward compatibility where possible and follow the existing code style.
%s
**Testing Recommendations:**
- Verify the fix addresses the reported issue
- Check for any unintended side effects
- Run existing test suite if available
- Test edge cases related to the changes

---

<sub>🤖 This PR was automatically generated by [Mr. Code Fixer](https://github.com/pefman/Mr-Code-Fixer) - an AI-powered issue resolution bot</sub>`,
		issue.Number, confidenceNote, fix.Explanation, fileChangesList, testSection)
	
	prURL, err := ghClient.CreatePullRequest(prTitle, prBody, branchName, gitOps.DefaultBranch)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	analytics.RecordPRCreated()
	analytics.RecordIssueHandled()
	fmt.Printf("✓ Pull request created: %s\n", prURL)

	// If high confidence, close the issue with a detailed comment
	if fix.Confidence == "high" {
		fmt.Println("Closing issue (high confidence fix)...")
		
		// Create user-friendly explanation
		fileList := ""
		for i, change := range fix.FileChanges {
			if i < 3 { // Show first 3 files
				fileList += fmt.Sprintf("`%s`", change.FilePath)
				if i < len(fix.FileChanges)-1 && i < 2 {
					fileList += ", "
				}
			}
		}
		if len(fix.FileChanges) > 3 {
			fileList += fmt.Sprintf(" and %d more", len(fix.FileChanges)-3)
		}
		
		closeComment := fmt.Sprintf(`## ✅ Issue Resolved!

Great news! I've analyzed this issue and created a fix that should resolve the problem.

**What I did:**
%s

**Files modified:** %s

**Next steps:**
I've created a pull request with the changes: %s

Please review the PR to make sure everything looks good. The fix has been implemented with high confidence, but it's always good to double-check before merging. If you notice any issues or have questions about the approach, feel free to comment on the PR!

---

<sub>🤖 Fixed automatically by Mr. Code Fixer</sub>`,
			fix.Explanation, fileList, prURL)
		
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

// isIssueTooVague checks if an issue lacks sufficient detail to fix
func isIssueTooVague(issue Issue) bool {
	combined := strings.ToLower(issue.Title + " " + issue.Body)
	
	// Vague phrases that indicate lack of detail
	vaguePhrases := []string{
		"something is wrong",
		"something broken",
		"doesn't work",
		"not working",
		"broken",
		"fix this",
		"fix it",
		"help",
		"issue",
		"problem",
	}
	
	// If title is very short and vague
	if len(issue.Title) < 20 {
		for _, phrase := range vaguePhrases {
			if strings.Contains(combined, phrase) {
				// Check if there's substantial detail in body
				if len(issue.Body) < 50 { // Less than 50 chars in body
					return true
				}
			}
		}
	}
	
	// If no file mentions and very short description
	hasFileMention := strings.Contains(combined, "/") || 
					 strings.Contains(combined, ".js") ||
					 strings.Contains(combined, ".py") ||
					 strings.Contains(combined, ".go") ||
					 strings.Contains(combined, ".php") ||
					 strings.Contains(combined, ".java")
	
	if !hasFileMention && len(combined) < 30 {
		return true
	}
	
	return false
}
