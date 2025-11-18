# BrowserGit Release Guide

This guide documents the process for releasing new versions of BrowserGit.

## Pre-Release Checklist

Before starting the release process, ensure:

- [ ] All tests pass (`npm run test:all`)
- [ ] All benchmarks pass (`npm run bench:run`)
- [ ] Code is linted (`npm run lint`)
- [ ] Type checking passes (`npm run type-check`)
- [ ] Documentation is up to date
- [ ] CHANGELOG.md is updated with new version
- [ ] Examples work correctly
- [ ] Cross-browser tests pass (Chrome, Firefox, Safari)
- [ ] Security audit is complete
- [ ] No open critical bugs
- [ ] All planned features for the release are complete

## Release Process

### 1. Version Bump

Choose the appropriate version bump based on changes:

```bash
# For bug fixes (0.1.0 -> 0.1.1)
npm run version:patch

# For new features (0.1.0 -> 0.2.0)
npm run version:minor

# For breaking changes (0.1.0 -> 1.0.0)
npm run version:major
```

This will update the version in all package.json files across the workspace.

### 2. Sync Versions

Ensure all packages have consistent versions:

```bash
npm run version:sync
```

### 3. Update Changelog

Edit `CHANGELOG.md` to:
- Add release date for the new version
- Move items from `[Unreleased]` to the new version section
- Add release notes highlighting key changes
- Document any breaking changes
- List known issues

### 4. Update Documentation

- Update version numbers in documentation
- Update installation instructions if needed
- Update API documentation for any changes
- Review and update README.md

### 5. Build and Test

Run the full CI suite:

```bash
npm run ci:full
```

This runs:
- Linting
- Type checking
- All unit tests
- All browser tests
- All benchmarks
- Full build
- Bundle analysis

### 6. Prepare Release

Run the prepare script:

```bash
npm run release:prepare
```

This will:
- Sync versions
- Build all packages
- Prepare artifacts for publishing

### 7. Dry Run

Test the publish process without actually publishing:

```bash
npm run release:dry-run
```

Review the output to ensure:
- All packages are included
- Versions are correct
- Files to be published are correct
- No sensitive files are included

### 8. Git Tag and Commit

Commit the version changes:

```bash
git add .
git commit -m "chore: release v0.1.0"
git tag -a v0.1.0 -m "Release v0.1.0"
```

### 9. Publish to npm

Publish all packages to npm:

```bash
npm run release:publish
```

Enter your npm 2FA code when prompted.

### 10. Push to GitHub

Push the commits and tags:

```bash
git push origin main
git push origin v0.1.0
```

### 11. Create GitHub Release

Create a GitHub release:

1. Go to https://github.com/user/browser-git/releases/new
2. Select the tag (v0.1.0)
3. Set release title: "BrowserGit v0.1.0"
4. Copy content from CHANGELOG.md for this version
5. Add any additional notes
6. Attach build artifacts (WASM files, etc.)
7. Publish release

### 12. Update Documentation Site

If you have a documentation site:

```bash
cd docs
npm run deploy
```

### 13. Announce Release

Share the release on:

- Twitter/X
- Reddit (r/javascript, r/webdev, r/programming)
- Hacker News
- Dev.to
- LinkedIn
- Discord communities
- Slack communities

Use the template from `SOCIAL_MEDIA_TEMPLATES.md`.

## Post-Release Checklist

- [ ] npm packages are published
- [ ] GitHub release is created
- [ ] Documentation site is updated
- [ ] Announcement blog post is published
- [ ] Social media announcements are posted
- [ ] Monitor for issues in the first 24 hours
- [ ] Update project board/roadmap
- [ ] Close related GitHub milestones

## Rollback Procedure

If critical issues are discovered after release:

### Option 1: Quick Fix (Patch Release)

1. Fix the issue on main branch
2. Follow release process for patch version
3. Communicate the fix to users

### Option 2: Deprecate Version

```bash
npm deprecate @browser-git/browser-git@0.1.0 "Critical bug, please upgrade to 0.1.1"
```

### Option 3: Unpublish (Last Resort)

Only within 72 hours of publish:

```bash
npm unpublish @browser-git/browser-git@0.1.0
```

⚠️ **Warning**: Unpublishing can break existing installations. Use only for critical security issues.

## Release Schedule

- **Patch releases**: As needed for bug fixes (typically weekly if bugs are found)
- **Minor releases**: Monthly with new features
- **Major releases**: Quarterly or when breaking changes are necessary

## Version Support

- **Current release**: Full support
- **Previous minor**: Security fixes only
- **Older versions**: No official support

## Automated Release (CI/CD)

For automated releases via GitHub Actions:

1. Create a release branch: `release/v0.1.0`
2. Push to GitHub
3. GitHub Actions will:
   - Run all tests
   - Build packages
   - Create draft release
4. Review and publish the draft release
5. GitHub Actions will automatically publish to npm

## Emergency Hotfix Process

For critical security issues:

1. Create hotfix branch from main: `hotfix/security-fix`
2. Implement fix
3. Run tests: `npm run test:all`
4. Bump patch version: `npm run version:patch`
5. Commit and tag: `git commit -m "security: fix critical vulnerability"`
6. Publish immediately: `npm run release:publish`
7. Create GitHub security advisory
8. Notify users via all channels
9. Merge hotfix back to main

## Package-Specific Releases

To release a specific package only:

```bash
cd packages/browser-git
npm version patch
npm publish --access public
```

⚠️ **Note**: This should be avoided. Prefer releasing all packages together to maintain consistency.

## Release Artifacts

Each release should include:

1. **npm Packages**
   - @browser-git/browser-git
   - @browser-git/storage-adapters
   - @browser-git/diff-engine
   - @browser-git/git-cli

2. **GitHub Release Assets**
   - Source code (zip/tar.gz)
   - WASM binaries
   - Standalone builds (if applicable)
   - Documentation PDF (optional)

3. **Documentation**
   - Updated API docs
   - Migration guide (for major versions)
   - Updated examples

## Communication Templates

### npm Package Description

```
BrowserGit - A complete Git implementation for browsers using Go + WebAssembly and TypeScript. Enables offline-first version control in web applications with support for all core Git operations including clone, commit, branch, merge, and remote operations. Multiple storage backends (IndexedDB, OPFS, LocalStorage). Security-first design with SSRF protection and input validation.
```

### GitHub Release Template

```markdown
## BrowserGit v0.1.0

[Brief description of the release]

### 🎉 Highlights

- Feature 1
- Feature 2
- Feature 3

### 🐛 Bug Fixes

- Fix 1
- Fix 2

### ⚠️ Breaking Changes

- Breaking change 1
- Breaking change 2

### 📦 Packages

- @browser-git/browser-git@0.1.0
- @browser-git/storage-adapters@0.1.0
- @browser-git/diff-engine@0.1.0
- @browser-git/git-cli@0.1.0

### 📚 Documentation

- [Getting Started](./docs/getting-started.md)
- [API Reference](./docs/api-reference/repository.md)
- [Migration Guide](./docs/MIGRATION.md) (if applicable)

### 🙏 Contributors

Thanks to all contributors who made this release possible!

**Full Changelog**: https://github.com/user/browser-git/compare/v0.0.1...v0.1.0
```

## Troubleshooting

### Publish Fails with 401

- Ensure you're logged in to npm: `npm whoami`
- Re-login if needed: `npm login`
- Check if you have publish rights to the @browser-git scope

### Publish Fails with 403

- Check if version already exists on npm
- Ensure you have 2FA enabled and working
- Verify you're a maintainer of the packages

### Tests Fail in CI

- Run tests locally first
- Check for environment-specific issues
- Review test logs in CI

### Version Mismatch

- Run `npm run version:sync` to fix
- Manually check all package.json files
- Ensure all internal dependencies use correct versions

## Resources

- [npm Documentation](https://docs.npmjs.com/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github)

## Support

For questions about the release process:
- Open an issue: https://github.com/user/browser-git/issues
- Contact maintainers: [list of maintainer contacts]
