# Mr. Code Fixer 🤖

An automated CLI tool that uses AI to fix GitHub issues and create pull requests automatically. Perfect for automating simple bug fixes and maintenance tasks.

## Features

- 🔍 Fetches open issues from GitHub repositories
- 🤖 Uses free AI services (Groq or Ollama) to analyze and fix issues
- 🔧 Automatically creates branches, commits, and pull requests
- 🚀 Easy to automate with simple CLI flags
- 💰 Uses free AI APIs (no expensive API costs!)

## Prerequisites

- Go 1.21 or higher
- Git installed and configured
- GitHub personal access token with `repo` permissions
- Groq API key (free tier: https://console.groq.com) OR Ollama installed locally

## Installation

### Option 1: Build from source

```bash
git clone https://github.com/pefman/mr-code-fixer.git
cd mr-code-fixer
go build -o mr-code-fixer
```

### Option 2: Install directly

```bash
go install github.com/pefman/mr-code-fixer@latest
```

## Setup

### 1. Get a GitHub Token

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select scope: `repo` (Full control of private repositories)
4. Copy the token

### 2. Get a Groq API Key (Free!)

1. Go to https://console.groq.com
2. Sign up for a free account
3. Navigate to API Keys
4. Create a new API key

### 3. Set Environment Variables

```bash
# Windows PowerShell
$env:GITHUB_TOKEN="your_github_token"
$env:GROQ_API_KEY="your_groq_api_key"

# Linux/Mac
export GITHUB_TOKEN="your_github_token"
export GROQ_API_KEY="your_groq_api_key"
```

## Usage

### Basic Usage

Fix all open issues (up to max-issues):

```bash
./mr-code-fixer -owner username -repo repository-name
```

### Fix a Specific Issue

```bash
./mr-code-fixer -owner username -repo repository-name -issue 42
```

### Using Ollama (Local AI)

```bash
# Make sure Ollama is running: ollama serve
# Pull a model: ollama pull llama2

./mr-code-fixer -owner username -repo repository-name \
  -ai-service ollama \
  -ai-model llama2
```

### All CLI Options

```bash
./mr-code-fixer [flags]

Flags:
  -owner string
        GitHub repository owner (required)
  -repo string
        GitHub repository name (required)
  -github-token string
        GitHub personal access token (default: $GITHUB_TOKEN)
  -ai-service string
        AI service to use: groq or ollama (default "groq")
  -ai-key string
        API key for AI service (default: $GROQ_API_KEY)
  -ai-model string
        AI model to use (default "llama-3.3-70b-versatile")
  -ollama-url string
        Ollama API URL (default "http://localhost:11434")
  -issue int
        Specific issue number to fix (0 = all open issues) (default 0)
  -max-issues int
        Maximum number of issues to process (default 1)
  -work-dir string
        Working directory for cloning repos (default "./workspace")
```

## Examples

### Fix Issue #15 in your own repo

```bash
./mr-code-fixer -owner pefman -repo my-project -issue 15
```

### Process up to 5 issues at once

```bash
./mr-code-fixer -owner pefman -repo my-project -max-issues 5
```

### Use with different AI model

```bash
./mr-code-fixer -owner pefman -repo my-project \
  -ai-model llama-3.1-70b-versatile
```

### Automate with a script

```powershell
# Windows PowerShell - process-issues.ps1
$repos = @("repo1", "repo2", "repo3")

foreach ($repo in $repos) {
    Write-Host "Processing $repo..."
    ./mr-code-fixer -owner pefman -repo $repo -max-issues 3
}
```

```bash
# Linux/Mac - process-issues.sh
#!/bin/bash
for repo in repo1 repo2 repo3; do
    echo "Processing $repo..."
    ./mr-code-fixer -owner username -repo $repo -max-issues 3
done
```

## How It Works

1. **Fetch Issues**: Retrieves open issues from the specified GitHub repository
2. **Clone Repo**: Clones the repository to a local workspace
3. **Analyze**: Sends issue details and repo context to AI for analysis
4. **Apply Fix**: AI suggests code changes, which are applied to the local repo
5. **Create PR**: Creates a new branch, commits changes, and opens a pull request

## AI Services

### Groq (Recommended for Cloud)

- **Free tier**: 14,400 requests/day
- **Fast**: Optimized for speed
- **Models**: Llama 3.3 70B, Mixtral, Gemma, etc.
- **Sign up**: https://console.groq.com

### Ollama (Recommended for Local/Privacy)

- **Completely free**: Runs locally on your machine
- **Private**: No data leaves your computer
- **Models**: Llama 2, Llama 3, CodeLlama, Mistral, etc.
- **Install**: https://ollama.com

## Limitations

- Best suited for simple to medium complexity issues
- AI may not always produce perfect fixes (review PRs before merging!)
- Large repositories may take time to process
- Requires clear issue descriptions for best results

## Tips for Best Results

1. **Clear Issue Descriptions**: Write detailed issues with steps to reproduce
2. **Start Small**: Test with simple issues first
3. **Review PRs**: Always review generated pull requests before merging
4. **Good Repo Structure**: Well-organized code helps AI understand context
5. **Include Tests**: Mention test files in issues for AI to reference

## Troubleshooting

### "GitHub token required"
Set the `GITHUB_TOKEN` environment variable or use the `-github-token` flag.

### "Groq API key required"
Set the `GROQ_API_KEY` environment variable or use the `-ai-key` flag.

### "git clone failed"
Ensure Git is installed and you have permission to access the repository.

### "failed to parse AI response"
The AI output was malformed. Try running again or use a different model.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

Created by pefman

---

**Note**: This tool is experimental. Always review generated pull requests before merging. The AI may not always produce correct solutions.
