# Changelog

All notable changes to Mr. Code Fixer will be documented in this file.

## [v1.3.5] - 2025-01-14

### Added
- Version display in welcome banner
- Immediate issue loading on startup (no manual trigger needed)
- Settings integrated into issue selection menu (S/Q options)
- Smart issue filtering - checks if bot's comment is the last one
- Repository title displayed above issue list
- Error handling with settings retry option
- Currency changed from USD to SEK (kr)

### Changed
- Comment-only responses now automatically close issues
- Cleaner UI without duplicate boxes
- "Loading issues" message with improved spacing
- Bot reprocesses issues when users reply after bot's comment

### Fixed
- Issues now load immediately without waiting
- Better handling of issues with new user responses

## [v1.2.0] - 2024-01-XX

### Added
- **Session Analytics**: Track API calls, costs, PRs created, issues handled, and questions asked
- **Cost Estimation**: Shows estimated cost before processing multiple issues (warns if > $0.10)
- **Test Execution**: Automatically detects and runs tests before creating PRs
  - Supports: Node.js (npm test), Go (go test), Python (pytest), Rust (cargo test), Java (Maven/Gradle), PHP (PHPUnit)
  - Only creates PR if tests pass
  - Rolls back changes if tests fail
  - Notes test status in PR descriptions
- **Session Summary**: Displays detailed analytics at end of each session (duration, API calls, costs, metrics)

### Changed
- AI clients now track analytics for each API call
- PR descriptions now include test results section
- Analytics tracking integrated throughout issue processing flow

### Technical
- New `analytics.go`: SessionAnalytics struct with mutex-safe tracking
- New `tests.go`: TestRunner for multi-language test detection and execution
- Updated `ai.go`: Added SetAnalytics() method to all AI clients
- Updated `main.go`: Integrated analytics and test execution into workflow

## [v1.1.0] - 2024-01-XX

### Added
- Vague issue detection with automatic request for more details
- Filtering of pull requests from issue list
- Duplicate comment prevention (bot won't respond to same issue twice)
- Early filtering of already-handled issues

### Changed
- Improved issue selection menu shows only new issues
- Better user feedback when skipping handled issues

## [v1.0.0] - 2024-01-XX

### Added
- Initial release
- GitHub issue fetching and processing
- Three AI providers: OpenAI (ChatGPT), xAI (Grok), Ollama (local)
- Interactive setup with config persistence
- Smart file selection (top 30 relevant files by keyword/mention scoring)
- Confidence-based decision making (high/medium/low)
- Automatic question asking when uncertain
- Beautiful colored UI with emojis and progress indicators
- Technical PR descriptions with detailed analysis
- User-friendly issue comments
- Automatic issue closing for high-confidence fixes
- Sanitized branch naming from issue titles
- Multi-platform builds (Windows, Linux x86/ARM64, macOS Intel/Apple Silicon)
- GoReleaser configuration for automated releases
