# Code signing policy

**Free code signing provided by SignPath.io, certificate by SignPath Foundation.**

MetriyonAPI is an open-source project licensed under the MIT License and maintained by OrganizzaTech in the canonical repository `ricardow2005/MetriyonAPI`.

## Team roles

- **Committers and reviewers:** Ricardo Alberto Walter (`@ricardow2005`)
- **Approvers:** Ricardo Alberto Walter (`@ricardow2005`)

At the current project stage, the repository owner performs the committer, reviewer, release-maintainer, and signing-approver responsibilities. Contributions from other people are reviewed before they are accepted into the canonical repository.

## Build provenance

Official Windows binaries are built from tagged commits in this repository by `.github/workflows/release.yml`. The workflow performs the frontend build, Go tests, and the native Windows Wails build before producing release artifacts.

Only artifacts produced from the canonical repository's automated release process are eligible to be submitted for official signing. Local developer builds and binaries produced by forks are not official OrganizzaTech releases.

## Signing process

The project is applying for the SignPath Foundation open-source code-signing program. Until that integration is approved and enabled, Windows CI and draft-release artifacts are explicitly named **unsigned**.

After approval:

1. A release is built automatically from a tagged commit in the canonical repository.
2. The build provenance is verified through the configured trusted build system.
3. The release artifact is submitted to SignPath.
4. An authorized approver manually approves the signing request.
5. Only the resulting signed artifact is published as the official signed Windows release.

Every release requires explicit signing approval. Signing credentials and private keys are never stored in this repository.

## Product metadata

Signed Windows binaries use consistent product metadata:

- Product name: `MetriyonAPI`
- Company/project identity: `OrganizzaTech`
- Product page: `https://www.organizza.com.br/produtos/metriyonapi`
- Project website: `https://www.organizza.com.br`
- Source repository: `https://github.com/ricardow2005/MetriyonAPI`

## Privacy

See [`PRIVACY.md`](../PRIVACY.md).

MetriyonAPI does not transfer information to other networked systems unless specifically requested or configured by the user or the person operating the application. The application's purpose is to send REST/SOAP/API traffic to endpoints explicitly selected by the user. It does not intentionally send request contents, credentials, workspace data, or usage telemetry to OrganizzaTech.

## Release ownership

Official MetriyonAPI releases are published only from the canonical repository maintained by `ricardow2005` / OrganizzaTech. Forks may build and redistribute modified versions under the MIT License, but those builds are not official MetriyonAPI releases and are not eligible for this project's official signing process.
