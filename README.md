# Mr. Code Fixer 🤖

An AI-powered GitHub bot that automatically analyzes and fixes issues in your repositories. Download, configure, and let it run autonomously to maintain your projects.

## What Does It Do?

Mr. Code Fixer is an autonomous bot that:
- 🔍 Monitors your GitHub repositories for open issues
- 🧠 Uses AI to understand what needs to be fixed
- 💡 Asks clarifying questions when uncertain
- ✏️ Creates fixes and opens pull requests
- 🧪 Runs tests before creating PRs
- 📊 Tracks API costs and session metrics
- ✅ Closes issues when confident about the solution

Think of it as a tireless contributor that works 24/7 to help maintain your codebase.

## Features

- **Smart Issue Processing**: Filters out vague issues, duplicates, and PRs automatically
- **Multi-AI Support**: Choose between ChatGPT (OpenAI), Grok (xAI), or Ollama (local)
- **Confidence-Based Decisions**: High confidence fixes auto-close issues; uncertain ones ask questions
- **Test Execution**: Automatically detects and runs tests (Go, Node.js, Python, Rust, Java, PHP)
- **Cost Tracking**: Shows estimated costs before processing multiple issues
- **Session Analytics**: Tracks API calls, costs, PRs created, and questions asked
- **Beautiful UI**: Colored output with emojis and progress indicators

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
AI analyzes and creates fix
  ↓
Bot detects test command (e.g., "go test")
  ↓
Runs tests after applying changes
  ├─ Tests pass → Creates PR, closes issue
  └─ Tests fail → Rolls back, reports error
```

### Session Analytics

After processing issues, the bot shows a summary:

```
╔══════════════════════════════════════════════════════════════╗
║                     📊 Session Summary                        ║
╚══════════════════════════════════════════════════════════════╝

⏱  Duration: 2m 34s
🤖 AI Calls: 5
💰 Estimated Cost: $0.005
📝 Issues Handled: 3
🔀 PRs Created: 2
❓ Questions Asked: 1
```

When fixing multiple issues, you'll see cost estimates first:

```
💰 Estimated cost for 10 issues: ~$0.015 (15 AI calls @ $0.001 each)
⚠️  Note: Processing multiple issues will incur API costs
Fix all 10 issues? (yes/no) [no]:
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

### Supported Test Frameworks

Mr. Code Fixer automatically detects and runs tests for:

| Language/Framework | Detection | Command |
|-------------------|-----------|---------|
| **Node.js** | package.json | `npm test` |
| **Go** | go.mod | `go test ./...` |
| **Python** | requirements.txt, setup.py | `python -m pytest` |
| **Rust** | Cargo.toml | `cargo test` |
| **Java (Maven)** | pom.xml | `mvn test` |
| **Java (Gradle)** | build.gradle | `gradle test` |
| **PHP** | composer.json | `php vendor/bin/phpunit` |

If no tests are found, the bot proceeds without test validation and notes this in the PR.

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
- **File limit**: Analyzes up to 30 most relevant files per issue to keep AI context manageable

## Writing Issues the Bot Understands

The bot works best with well-written issues. Here are examples:

### ✅ Perfect Issue Example

```markdown
Title: Login button not working in src/components/Auth/LoginForm.tsx

Description:
The login button on the login page doesn't submit the form when clicked.

File: src/components/Auth/LoginForm.tsx
Line: Around line 45

Expected: Clicking "Login" should call handleSubmit() and send credentials to API
Actual: Nothing happens when clicking the button

Error in console:
TypeError: Cannot read property 'submit' of undefined
```

**Why this works:**
- ✅ Mentions exact file path: `src/components/Auth/LoginForm.tsx`
- ✅ Describes expected vs actual behavior
- ✅ Includes error message
- ✅ Specifies approximate location (line 45)

### ✅ Good Issue Example

```markdown
Title: Database connection failing in api/db.go

The app crashes on startup with "connection refused" error.
I think the issue is in api/db.go where we initialize the database pool.
```

**Why this works:**
- ✅ Mentions file: `api/db.go`
- ✅ Clear error description
- ✅ Hints at the problem area

### ⚠️ Okay Issue Example

```markdown
Title: Typo in documentation

README.md has wrong command - it says `npm start` but should be `npm run dev`
```

**Why this works:**
- ✅ Mentions file: `README.md`
- ✅ Clear what needs to change
- ⚠️ Simple fix, bot will handle confidently

### ❌ Poor Issue Example

```markdown
Title: App doesn't work

Something is broken. Please fix it.
```

**Why this fails:**
- ❌ No file mentioned
- ❌ No error description
- ❌ No context about what "doesn't work" means
- **Bot response:** Will ask clarifying questions in the issue

### Best Practices for Issues

1. **Mention files explicitly**: Use backticks for file paths: `src/utils/helper.js`
2. **Include error messages**: Copy-paste actual errors from console/logs
3. **Describe expected behavior**: What should happen vs what actually happens
4. **Add context**: Environment, steps to reproduce, related files
5. **Use keywords**: Words like "login", "database", "api" help the bot find relevant files

### How the Bot Finds Files

The bot uses smart file selection:
- **Explicit mentions**: Files mentioned in the issue get highest priority
- **Keyword matching**: Finds files with issue keywords in their path
- **Relevance scoring**: Ranks files by how likely they are related
- **Limit**: Analyzes top 30 most relevant files (not entire codebase)

**Example:** Issue mentions "login problem" → Bot prioritizes:
1. Files explicitly mentioned: `auth/login.js`
2. Files with "login" in path: `components/LoginForm.tsx`, `services/loginService.js`
3. Files with "auth" in path: `middleware/auth.js`
4. Common entry points: `index.js`, `main.go`, `app.py`

## Tips for Best Results

1. **Write clear issues**: Mention file paths and include error messages
2. **Start with simple issues**: Test the bot on documentation or simple bugs first
3. **Review PRs carefully**: Always review before merging, especially for critical code
4. **Use labels**: Consider only letting the bot handle issues labeled "auto-fix" or "good-first-issue"
5. **Monitor costs**: If using paid AI services, track your API usage
6. **One issue per problem**: Don't bundle multiple unrelated problems in one issue

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
