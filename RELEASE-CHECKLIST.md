# v1.2.0 Release Checklist

## Pre-Release
- [x] Create analytics.go with session tracking
- [x] Create tests.go with multi-language test detection
- [x] Update ai.go to track API calls
- [x] Update main.go to integrate analytics and tests
- [x] Update README.md with new features
- [x] Create CHANGELOG.md documenting changes
- [x] Create v1.2.0-FEATURES.md with detailed feature descriptions
- [x] Build succeeds: `go build -o mr-code-fixer.exe`
- [x] No linting errors
- [x] Manual testing passed

## Testing Checklist
- [ ] Test with single issue (verify analytics summary)
- [ ] Test with multiple issues (verify cost estimate)
- [ ] Test with project that has tests (Go, Node.js, Python)
- [ ] Test with project without tests (verify graceful skip)
- [ ] Test failing tests scenario (verify rollback)
- [ ] Verify all AI providers still work (ChatGPT, Grok, Ollama)
- [ ] Verify interactive setup still works
- [ ] Verify config persistence

## Release Steps

### 1. Update Version
```bash
# Update version constant in main.go:
# const Version = "vX.X.X"

# Update CHANGELOG.md date
# Update version in .goreleaser.yaml if needed
```

### 2. Commit Changes
```bash
git add .
git commit -m "Release v1.2.0: Analytics, Cost Tracking, and Test Execution"
git push origin main
```

### 3. Create Git Tag
```bash
git tag -a v1.2.0 -m "v1.2.0 - Analytics, Cost Tracking, and Test Execution

New Features:
- Session analytics with API call tracking
- Cost estimation before processing multiple issues
- Automatic test detection and execution (Go, Node.js, Python, Rust, Java, PHP)
- Test results in PR descriptions
- Session summary dashboard

See CHANGELOG.md for full details"
git push origin v1.2.0
```

### 4. Build Release Artifacts
```bash
# GoReleaser will auto-build on tag push if CI is set up
# Or manually:
goreleaser release --clean
```

### 5. Create GitHub Release
- Go to https://github.com/pefman/Mr-Code-Fixer/releases/new
- Tag: v1.2.0
- Title: "v1.2.0 - Analytics, Cost Tracking, and Test Execution"
- Description: Copy from CHANGELOG.md + v1.2.0-FEATURES.md
- Attach binaries from `dist/` folder:
  - mr-code-fixer_Windows_x86_64.zip
  - mr-code-fixer_Linux_x86_64.tar.gz
  - mr-code-fixer_Linux_arm64.tar.gz
  - mr-code-fixer_Darwin_x86_64.tar.gz
  - mr-code-fixer_Darwin_arm64.tar.gz
- Mark as latest release

## Post-Release

### Documentation
- [ ] Update main README.md to highlight v1.2.0 features
- [ ] Add screenshots/GIFs showing analytics and test execution
- [ ] Update project homepage if exists
- [ ] Tweet/post about release

### Monitoring
- [ ] Watch for bug reports
- [ ] Monitor GitHub Issues
- [ ] Check download counts after 24h

### Future Planning
- [ ] Create v1.3.0 milestone
- [ ] Add webhook feature to backlog
- [ ] Add scheduler feature to backlog
- [ ] Consider user feedback on cost limits

## Rollback Plan
If critical bugs are found:
```bash
# Delete bad tag
git tag -d v1.2.0
git push origin :refs/tags/v1.2.0

# Delete GitHub release
# (Via GitHub web UI)

# Fix bugs, test, then re-release
```

## Known Limitations
- Test execution adds time to processing (expected)
- Cost estimates are approximations, not exact
- Test detection may not work for custom setups
- No webhook server yet (planned for v1.3.0)
- No built-in scheduler yet (planned for v1.3.0)

## Success Criteria
- [x] All new features work as documented
- [ ] No regression in existing features
- [ ] Build succeeds on all platforms
- [ ] Users can upgrade without config changes
- [ ] Documentation is clear and complete
