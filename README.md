# Mr. Code Fixer 🤖

An AI-powered GitHub bot that automatically analyzes and fixes issues in your repositories. Download, configure, and let it run autonomously to maintain your projects.

## What Does It Do?

Mr. Code Fixer is an autonomous bot that:
- 🔍 Monitors your GitHub repositories for open issues
- 🧠 Uses AI to understand what needs to be fixed
- 💡 Asks clarifying questions when uncertain
- ✏️ Creates fixes and opens pull requests
- ✅ Closes issues when confident about the solution

Think of it as a tireless contributor that works 24/7 to help maintain your codebase.

## Quick Start

### 1. Download

Get the latest release for your platform from the [releases page](https://github.com/pefman/Mr-Code-Fixer/releases):
- `mr-code-fixer_Windows_x86_64.zip` - Windows
- `mr-code-fixer_Linux_x86_64.tar.gz` - Linux (Intel/AMD)
- `mr-code-fixer_Linux_arm64.tar.gz` - Linux (ARM)
- `mr-code-fixer_Darwin_x86_64.tar.gz` - macOS (Intel)
- `mr-code-fixer_Darwin_arm64.tar.gz` - macOS (Apple Silicon)

### 2. Setup

Run the bot for the first time:

```bash
./mr-code-fixer
```

It will guide you through interactive setup:
- **GitHub Repository**: Which repo should the bot help fix?
- **GitHub Token**: Your personal access token (see below)
- **AI Service**: Choose ChatGPT, Grok, or local Ollama
- **Working Directory**: Where to clone repos (defaults to `~/.mr-code-fixer/workspace`)

Configuration is saved in `~/.mr-code-fixer.json` for future runs.

### 3. Run

The bot will:
1. Show all open issues
2. Let you select which one to fix
3. Analyze the issue with AI
4. Either ask questions (if uncertain) or create a PR (if confident)

## Use Cases

### Personal Projects
Run manually whenever you want help with issues:
```bash
./mr-code-fixer
```

### Automation
Set up the bot to run periodically:

**Windows (Task Scheduler):**
```powershell
# Run every 6 hours
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Hours 6)
$action = New-ScheduledTaskAction -Execute "C:\path\to\mr-code-fixer.exe"
Register-ScheduledTask -TaskName "Mr Code Fixer" -Trigger $trigger -Action $action
```

**Linux/macOS (cron):**
```bash
# Add to crontab (runs every 6 hours)
0 */6 * * * cd /path/to && ./mr-code-fixer
```

**Docker:**
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o mr-code-fixer
CMD ["./mr-code-fixer"]
```

### Dedicated Bot Account
For a true "bot experience":

1. Create a new GitHub account (e.g., `my-code-fixer-bot`)
2. Set up a nice bot profile (avatar, bio: "🤖 Automated code fixer")
3. Generate a PAT from the bot account
4. Invite the bot as a collaborator to your repos
5. Run Mr. Code Fixer with the bot's token

Now all PRs and comments will appear from the bot account!

## Configuration

### GitHub Personal Access Token

Create a token at https://github.com/settings/tokens with these permissions:

**Fine-grained token (recommended):**
- Repository access: Select repositories you want the bot to help with
- Permissions:
  - Contents: Read and write
  - Issues: Read and write
  - Pull requests: Read and write
  - Metadata: Read-only

**Classic token:**
- `repo` (full control)

### AI Services

Choose one of three AI providers:

#### 1. ChatGPT (OpenAI)
- **Get API Key**: https://platform.openai.com/api-keys
- **Models**: gpt-4o, gpt-4-turbo, gpt-3.5-turbo
- **Cost**: Pay-per-use (check OpenAI pricing)

#### 2. Grok (xAI)
- **Get API Key**: https://x.ai
- **Models**: grok-3, grok-4-fast-reasoning, grok-code-fast-1
- **Cost**: Pay-per-use (check xAI pricing)

#### 3. Ollama (Local)
- **Install**: https://ollama.ai
- **Models**: llama2, codellama, deepseek-coder (free, runs on your machine)
- **Cost**: Free, but uses your compute resources
- **Setup**: `ollama pull codellama` then select in the bot

## How The Bot Thinks

### Confidence-Based Decisions

The bot assesses its confidence before acting:

**High Confidence** ✅
- Creates PR with the fix
- Adds comment to issue
- Automatically closes the issue

**Medium/Low Confidence** ⚠️
- Creates PR with warnings
- Leaves issue open for human review

**Needs More Info** ❓
- Posts questions as issue comments
- Waits for human clarification
- Does NOT create a PR

### Branch Naming

The bot creates descriptive branches:
- `fix/1-app-crashes-on-startup`
- `fix/23-typo-in-documentation`
- `fix/5-where-is-documentation`

### Example Workflow

```
Issue #5: "App crashes on startup"
  ↓
Bot analyzes code + issue description
  ↓
High confidence? 
  ├─ Yes → Creates fix, opens PR, closes issue
  └─ No → Posts: "Could you provide the error logs?"
```

## Advanced Usage

### Configuration File

The bot saves settings in `~/.mr-code-fixer.json`. You can edit this directly:

```json
{
  "github_token": "ghp_xxxxx",
  "repo_owner": "yourusername",
  "repo_name": "yourrepo",
  "ai_service": "grok",
  "ai_api_key": "xai-xxxxx",
  "ai_model": "grok-code-fast-1",
  "work_dir": "/home/user/.mr-code-fixer/workspace"
}
```

### Multiple Repositories

To use the bot with multiple repos, either:
- Run it interactively and change the repo each time
- Create separate config files and use: `./mr-code-fixer --config /path/to/config.json`
- Set up separate bot instances with different working directories

## Building From Source

### Requirements
- Go 1.21 or higher
- Git

### Build

```bash
git clone https://github.com/pefman/Mr-Code-Fixer.git
cd Mr-Code-Fixer
go build -o mr-code-fixer
```

### Create Release

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
goreleaser release --clean
```

## Limitations & Considerations

- **Quality depends on AI**: The bot is only as good as the AI model you choose
- **Complex issues**: May require human intervention or clarification
- **No testing**: The bot cannot run tests (yet) - always review PRs before merging
- **API limits**: Respects GitHub API rate limits (5000 requests/hour for authenticated)
- **Cost awareness**: ChatGPT and Grok are paid services - monitor your usage

## Tips for Best Results

1. **Write clear issues**: The better the issue description, the better the fix
2. **Start with simple issues**: Test the bot on documentation or simple bugs first
3. **Review PRs carefully**: Always review before merging, especially for critical code
4. **Use labels**: Consider only letting the bot handle issues labeled "auto-fix" or "good-first-issue"
5. **Monitor costs**: If using paid AI services, track your API usage

## Contributing

Contributions welcome! Areas for improvement:
- Add automated testing before creating PRs
- Implement GitHub App for easier installation
- Add webhook listener for real-time responses
- Support for more AI providers
- Better code context understanding
- Issue priority/labeling system

## License

MIT License - See LICENSE file for details.

## Credits

Built with Go and AI. Zero external dependencies beyond the standard library.

---

**Need help?** Open an issue or check existing issues for common problems and solutions.
