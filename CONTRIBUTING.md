# Contributing to MetriyonAPI

Thanks for considering a contribution.

## Development

Requirements:

- Go 1.24 or newer;
- Node.js 20 or newer;
- npm;
- Wails v2.10.2 or a compatible Wails v2 release;
- Windows WebView2 Runtime for Windows desktop development.

Install frontend dependencies:

```bash
cd frontend
npm ci
cd ..
```

Run in development mode:

```bash
wails dev
```

Run tests:

```bash
go test ./...
```

Build the frontend:

```bash
cd frontend
npm run build
```

Windows helper `.bat` files used by maintainers are intentionally excluded from the public repository. For a local desktop build, use the standard Wails command:

```bash
wails build -platform windows/amd64
```

Official release signing will be handled by the release signing pipeline once the open-source signing provider is enabled.

## Pull requests

Keep changes focused, describe the behavior changed, and include tests when practical. Do not include credentials, generated databases, `node_modules`, local signing certificates or private keys.

By contributing, you agree that your contribution is licensed under the repository's MIT License.
