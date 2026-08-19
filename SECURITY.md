# Security Policy

## Supported versions

Security fixes are applied to the latest maintained release of MetriyonAPI.

## Reporting a vulnerability

Please do not publish sensitive vulnerability details in a public GitHub issue before maintainers have had a reasonable opportunity to investigate.

For the moment, use the contact channels published at https://www.organizza.com.br and identify the report as a **MetriyonAPI security report**.

Include, when possible:

- affected version;
- steps to reproduce;
- impact;
- proof of concept that does not expose third-party secrets;
- suggested remediation, if known.

## Secrets

Never commit passwords, API tokens, OAuth client secrets, private keys, PFX/P12 files or signing credentials. The repository `.gitignore` excludes common local signing files, but contributors remain responsible for reviewing commits before pushing.

## Software updates

Update discovery is restricted to the official public GitHub Releases feed for `ricardow2005/MetriyonAPI`. The in-app download action only opens URLs under the canonical GitHub Releases path. MetriyonAPI does not silently replace its executable; users remain in control of downloading and installing a new release.
