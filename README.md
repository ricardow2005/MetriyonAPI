# MetriyonAPI

**Open-source API development, testing and workflow automation client by OrganizzaTech.**

- Product: https://www.organizza.com.br/produtos/metriyonapi
- OrganizzaTech: https://www.organizza.com.br
- License: [MIT](LICENSE)
- Privacy: [PRIVACY.md](PRIVACY.md)
- Security: [SECURITY.md](SECURITY.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Code signing policy: [docs/CODE_SIGNING.md](docs/CODE_SIGNING.md)
- First GitHub publication: [docs/PUBLISH_GITHUB.md](docs/PUBLISH_GITHUB.md)

## Open-source releases

The canonical repository is intended to be hosted at `github.com/ricardow2005/metriyonapi`. Source code is public under the MIT License, while official releases remain controlled by the repository maintainer. Forks are permitted by the MIT License but are not official OrganizzaTech releases.

Windows CI builds are currently marked **unsigned** until the SignPath Foundation open-source signing integration is approved and connected. The tag workflow creates a **draft release**, so the maintainer reviews and publishes each official release manually.

## Local Windows build

Run:

```bat
wails build -platform windows/amd64
```

Maintainer-specific Windows helper scripts and local signing configuration are intentionally kept outside the public repository.

---

# Metriyon API

Metriyon API is a local-first desktop client for building, sending, automating, and inspecting REST and SOAP requests. It uses a Go transport engine, a Wails desktop shell, a lightweight TypeScript interface, Tailwind CSS, and SQLite persistence. No account or external application server is required.

Current version: **0.6.7**

## Features in this delivery

- Local workspaces, collections, folders, requests, environments, open tabs, settings, and history.
- REST methods: GET, POST, PUT, PATCH, DELETE, HEAD, and OPTIONS.
- SOAP 1.1 and SOAP 1.2, manual XML envelopes, SOAPAction handling, SOAP Fault detection, and basic WSDL import.
- Query parameter, header, JSON, XML, text, HTML, form, and multipart editors.
- No Auth, Basic Auth, Bearer Token, API Key, and OAuth 2.0. OAuth token acquisition supports Client Credentials, Password, and Refresh Token grants, with client credentials in the request body or Basic Auth header.
- Automatic OAuth 2.0 acquisition and renewal before regular requests, Flows, and Load Tests. Missing tokens and tokens within 30 seconds of expiration are refreshed automatically; the manual token button remains available.
- Persisted request duplication with an independent UUID and a complete copy of request data.
- Sidebar collection management with contextual create/rename/delete actions and drag-and-drop request moves between folders.
- Persisted collection/folder collapse state, whole-folder moves between collections, and a concurrent Load Test workspace with latency percentiles, throughput, per-call records, cancellation, and SQLite run history.
- Response copying: formatted JSON/XML from Pretty view or the exact payload from Raw view.
- Collapsible response tray and Load Test chart mode with selectable historical baseline, overlaid per-call latency, and regression/improvement deltas for average, P95, P99, and throughput.
- Per-request Load Test scenario reset with confirmation and persistent removal of saved runs, results, graphs, and comparisons.
- Tab context menu for closing one tab, all other tabs, tabs to the right, or every tab, with a single unsaved-changes decision for batch closes. Pressing Enter saves an edited request name and confirms collection/folder/request renames.
- Visual Flows editor with draggable nodes, handle-to-handle connections, persisted layouts, branch-aware execution, cycle validation, per-step results, and response extraction. Every node now exposes a `+` shortcut to choose the next saved request or Flow function without returning to the toolbar.
- Flow functions: **If** with independent `true`/`false` output handles, **Filter** to continue only when conditions match, and **Extract Data** to create reusable variables from the previous response. Conditions support response paths such as `status`, `body.customer.id`, headers, `{{variables}}`, ALL/ANY matching, type conversion, numeric/string operators, existence checks, and regex. JSON paths such as `client.nrCliente` or `items[0].id` can still export variables directly from request nodes.
- Flow node positions accept sub-pixel coordinates produced by browser/WebView pointer events, preventing save/run failures after dragging a node.
- Variables with the priority `Global < Environment < Collection < Request` (the rightmost scope wins).
- Go `net/http` execution with cancellation, timeouts, SSL verification control, redirects, response-size limits, and `httptrace` timings.
- Pretty JSON/XML, raw body, isolated HTML preview, image preview, response headers, cookies, redirect/timing details, and binary response saving.
- cURL import/export, keyboard shortcuts, dark/light/system themes, and persisted panel layout.
- AES-GCM encryption at rest for environment variables marked as secret. The random local master key is stored separately with user-only filesystem permissions where supported.

OAuth Authorization Code/PKCE, OpenAPI import, scripting, mTLS certificate management, proxy UI, cookie-jar scoping, workspace archives, and cloud synchronization are intentionally not exposed as functional controls in this release. The protocol handler and transport boundaries leave room for these additions.

## Architecture

```text
TypeScript + Tailwind UI
        |
        | Wails bindings (models only)
        v
      app/App
        |
        +-- variable/auth/request builders (REST + SOAP)
        +-- common net/http transport engine
        +-- cURL and WSDL importers
        +-- SQLite repositories and migrations
        +-- local secret protector
```

The frontend never calls remote APIs directly. REST and SOAP share the same transport engine, so timeout, TLS, redirect, cancellation, tracing, and response-limit behavior remain consistent.

## Project layout

```text
app/                 Wails-facing application API
internal/httpclient/ shared request builder and transport engine
internal/variables/  variable precedence and substitution
internal/curl/       cURL import/export
internal/wsdl/       basic WSDL reader and envelope generation
internal/soap/       SOAP response processing
internal/storage/    SQLite repositories and migration runner
internal/security/   secret encryption and log sanitization
internal/models/     UI/engine data contracts
frontend/            Vite + TypeScript + Tailwind UI
migrations/          ordered SQLite schema migrations
version/             centralized build/version information
build/               Wails platform build resources
```

## Prerequisites

- Go 1.24 or newer.
- Node.js 20 or newer (frontend build only; it is not an application runtime).
- Wails CLI v2.10 or newer.
- Windows: WebView2 (normally included with current Windows versions).
- Linux: WebKit2GTK development packages required by Wails.
- macOS: Xcode command-line tools.

Install Wails if needed:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2
```

## Development

```bash
cd frontend
npm install
cd ..
wails dev
```

The development database is created in the operating-system user configuration directory under `ForgeAPIClient/forge.db`. The application runs migrations automatically and uses SQLite WAL mode.

## Build rapido no Windows

Para compilar localmente no Windows, instale as dependencias do frontend e use o Wails CLI diretamente:

No Prompt de Comando ou PowerShell:

```bat
wails build -platform windows/amd64
```

Tambem e possivel dar duplo clique no arquivo. O resultado e gerado em:

```text
build\bin\MetriyonAPI.exe
```

Para executar sem o `pause` final, por exemplo em scripts:

```bat
wails build -platform windows/amd64
```

## Tests

Tests do not rely on external services. Transport tests use `httptest.Server`.

```bash
go test ./...
cd frontend && npm run build
```

Coverage focuses on variable precedence, REST/SOAP request construction, authentication and header/query handling, real local HTTP execution, cURL conversion, WSDL parsing, SOAP Faults, SQLite persistence, encryption, and redaction.

## Builds

### Windows

```powershell
wails build -platform windows/amd64 -clean `
  -ldflags "-X forge-api-client/version.Version=0.6.7 -X forge-api-client/version.Commit=$((git rev-parse --short HEAD)) -X forge-api-client/version.BuildDate=$((Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ'))"
```

Output: `build/bin/MetriyonAPI.exe`.

### Linux

Build on a Linux host with the Wails WebKit dependencies installed:

```bash
wails build -platform linux/amd64 -clean \
  -ldflags "-X forge-api-client/version.Version=0.6.7 -X forge-api-client/version.Commit=$(git rev-parse --short HEAD) -X forge-api-client/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

### macOS

Build on macOS:

```bash
wails build -platform darwin/universal -clean \
  -ldflags "-X forge-api-client/version.Version=0.6.7 -X forge-api-client/version.Commit=$(git rev-parse --short HEAD) -X forge-api-client/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Wails platform builds should be performed on the target operating system for the most reliable native packaging.


## Flow functions (0.6.0)

Visual flows support Request, If, Filter, Extract Data, Loop Over Items, Wait, Switch and Parallel nodes. Loop has `loop` and `done` outputs and a fixed execution counter; nodes connected to the loop branch can return to the Loop input. Wait accepts milliseconds, seconds or minutes. Switch supports numbered outputs selected by routing rules or by a numeric expression. Parallel starts all configured numbered branches concurrently.

While a flow is running, Wails progress events update the canvas in real time: the currently executing node and edge are highlighted, completed paths stay green, and inactive paths remain unchanged.


## Update notifications

MetriyonAPI checks the official GitHub Releases feed at most once every 24 hours. When a newer public release is available, the desktop app displays **“Há uma nova versão disponível!”** with the installed and latest versions, release notes, and a button that opens the official GitHub Release for download.

The application never replaces its own executable silently. Users can also run a manual check from **Settings → About → Check for updates**. Automatic checks only contact the public GitHub Releases API and do not transmit saved requests, environments, secrets, or workspace data.

## SQLite and persistence

The database is never created through ad-hoc table setup. The migration runner applies ordered files embedded from `migrations/` and records them in `schema_migrations`. UUIDs are used for synchronizable entities instead of local auto-increment identifiers.

At restart, the application restores the active workspace/environment, saved requests, collections, folders, environments, history, open saved tabs, theme, and split-pane preference.

## REST and SOAP behavior

The request pipeline is:

```text
Request definition -> scope merge -> variable resolution -> auth -> REST/SOAP builder
-> shared Go HTTP transport -> response processing -> history -> UI
```

SOAP 1.1 adds `text/xml; charset=utf-8` and a quoted `SOAPAction` when supplied. SOAP 1.2 uses `application/soap+xml; charset=utf-8`. Manually entered headers override generated defaults.

The WSDL importer accepts a URL or local file, locates services, SOAP addresses, binding operations, and SOAPAction values, and generates an editable starter envelope for every discovered operation.

## Security

- Remote traffic is executed only in Go.
- Secret environment values are encrypted before SQLite writes and masked in the UI.
- Authorization, passwords, tokens, client secrets, and private-key-like values are handled by centralized redaction helpers.
- SSL certificate verification defaults to on, and the request view shows a persistent warning when disabled.
- HTML response previews use a sandboxed iframe without script or privileged application access.
- Response previews have a configurable size limit; full binary data can be saved without being rendered as text.
- The workspace contains no hardcoded credentials.

The local master key is not a substitute for a locked operating-system account. A later Windows-specific release can wrap that key with DPAPI and use Keychain/Secret Service equivalents on macOS/Linux without changing stored secret records.

## Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| Ctrl/Cmd + Enter | Send request |
| Ctrl/Cmd + S | Save request |
| Ctrl/Cmd + N | New REST request |
| Ctrl/Cmd + W | Close tab |
| Ctrl/Cmd + K | Global search |
| Ctrl/Cmd + L | Focus URL |
| Ctrl/Cmd + Shift + F | Format body |

## License

MetriyonAPI is licensed under the [MIT License](LICENSE).

## Windows publisher metadata and code signing

Windows builds embed the following product metadata:

- Company: `OrganizzaTech`
- Product: `MetriyonAPI`
- Product page: `https://www.organizza.com.br/produtos/metriyonapi`
- Main website: `https://www.organizza.com.br`

These metadata fields identify the product but do not establish Authenticode trust by themselves. Official Windows releases are intended to use the SignPath Foundation open-source code-signing program after project approval. Until that integration is active, downloadable binaries are explicitly identified as **unsigned**.

## Code signing policy

**Free code signing provided by SignPath.io, certificate by SignPath Foundation.**

- **Committers and reviewers:** Ricardo Alberto Walter (`@ricardow2005`)
- **Approvers:** Ricardo Alberto Walter (`@ricardow2005`)
- **Privacy policy:** [PRIVACY.md](PRIVACY.md)
- **Full code signing policy:** [docs/CODE_SIGNING.md](docs/CODE_SIGNING.md)

MetriyonAPI does not transfer information to other networked systems unless the destination or operation is specifically requested or configured by the user. Requests executed by the application are sent only to endpoints selected by the user.

Official signed binaries will only be submitted from automated builds of tagged commits in this canonical repository. Every official signing request remains subject to maintainer approval.
