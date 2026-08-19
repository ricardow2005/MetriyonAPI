# Code signing policy

MetriyonAPI is an open-source project licensed under the MIT License and maintained by OrganizzaTech.

## Release ownership

Official MetriyonAPI releases are published only from the canonical repository maintained by `ricardow2005` / OrganizzaTech. Forks may build and distribute modified versions under the terms of the MIT License, but those builds are not official MetriyonAPI releases.

## Build provenance

Official Windows binaries are built from tagged commits in this repository using the workflow under `.github/workflows/release.yml`. The release workflow creates a draft GitHub Release so a maintainer remains responsible for reviewing and publishing each release.

## Signing

The project intends to use the SignPath Foundation open-source code-signing program for official Windows release binaries after the project is accepted and the SignPath project configuration is connected.

Until that integration is active, CI artifacts and draft-release binaries are explicitly treated as **unsigned**. Local developer builds may optionally be signed with a developer-controlled Authenticode certificate, but local signatures do not make a binary an official project release.

When SignPath signing is enabled, only artifacts produced from the canonical repository's approved release process will be submitted for official signing.

## Roles

- **Committers:** contributors whose commits are accepted into the canonical repository through maintainer review or direct maintainer changes.
- **Reviewers:** maintainers who review changes before they are included in release branches/tags.
- **Approvers:** maintainers authorized to approve an official signing/release operation.

At the current project stage, repository owner `ricardow2005` performs the maintainer/reviewer/approver responsibilities.

## Privacy

See [`PRIVACY.md`](../PRIVACY.md). MetriyonAPI communicates with network destinations configured or requested by the user and does not intentionally send project/request data to an OrganizzaTech telemetry service.
