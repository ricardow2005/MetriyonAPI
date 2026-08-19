# Changelog

All notable changes to MetriyonAPI will be documented here.

## [0.6.3] - 2026-08-19

### Added

- Automatic daily update check against the official MetriyonAPI GitHub Releases feed.
- New-version dialog showing installed/latest versions, release notes, and an official GitHub download action.
- Manual **Check for updates** action in Settings → About.
- Persistent 24-hour update-check throttle to avoid unnecessary GitHub API calls.

### Security

- Update downloads are opened only from the canonical `ricardow2005/MetriyonAPI` GitHub Releases URL.
- Update checks do not transmit saved API requests, credentials, environments, or workspace data.

## [0.6.2] - 2026-08-19

### Added

- Windows executable product metadata for OrganizzaTech and MetriyonAPI.
- Local Windows build automation.
- Open-source repository documentation and MIT license.
- GitHub Actions build and draft-release workflows.

### Changed

- Updated public Code signing policy and SignPath Foundation application documentation for the OSS signing process.
- Corrected stale README license/version/signing instructions.
- Local Authenticode signing is optional for developer builds while the public open-source signing pipeline is being established.

## [0.6.0] - 2026-08-19

### Added

- Advanced flow nodes, including loop, wait, switch and parallel execution.
- Live execution-path visualization in flows.
