# Changelog

All notable changes to MetriyonAPI will be documented here.

## [0.6.6] - 2026-08-19

- Corrigida a importação de WSDL por URL em serviços SOAP/IIS legados que encerram conexões HTTP de clientes genéricos.
- A leitura de WSDL remoto agora usa cabeçalhos compatíveis com navegador, desabilita keep-alive e força conexão curta para aumentar compatibilidade.
- Adicionado fallback HTTP/1.0 para endpoints HTTP antigos quando a tentativa HTTP padrão é encerrada pelo servidor remoto.
- Adicionada validação de URL, limite explícito de 8 MB e teste automatizado de importação WSDL por URL.

## [0.6.4] - 2026-08-19

### Fixed

- Fixed multiline cURL imports using Linux/macOS `\\` line continuations with LF and Windows CRLF line endings.
- Imported cURL endpoint URLs are now preserved correctly instead of falling back to the request placeholder.
- Query-string parameters are extracted into the Params editor during cURL import.
- `Authorization: Bearer` and `Authorization: Basic` headers are converted into the corresponding Auth configuration.
- Common API-key headers such as `X-API-Key` are imported into API Key authentication.
- Added support for common cURL forms including `--option=value`, `--json`, `--get`, `--data-urlencode`, `--user-agent`, and `--cookie`.

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
