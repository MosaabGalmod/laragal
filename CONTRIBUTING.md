# Contributing to LaraGal

Thank you for contributing to LaraGal. LaraGal is a local development and hosting suite for Windows 10 and 11 (64-bit), built with Go and Wails v2 (Vue 3 + Tailwind CSS).

## Development Setup

### Prerequisites
- Windows 10 or 11 (64-bit)
- Go 1.22 or higher
- Node.js 18 or higher with npm

### Setup Steps
1. **Clone the repository:**
   ```bash
   git clone https://github.com/MosaabGalmod/laragal.git
   cd laragal
   ```

2. **Frontend development and validation:**
   ```bash
   cd frontend
   npm install
   npm run lint
   npm test
   npm run build
   cd ..
   ```

3. **Backend verification:**
   ```bash
   go vet ./...
   go test ./...
   ```

4. **Compile the desktop binary:**
   ```bash
   go build -tags "desktop,production" -ldflags "-H windowsgui -s -w" -o laragal.exe .
   ```

## Contribution Rules

All contributions must follow these requirements, which are enforced by automated test suites in CI:

- **Bilingual UI Keys (ar.json & en.json):** All user-facing text in the frontend must use locale keys registered in both `frontend/src/locales/ar.json` and `frontend/src/locales/en.json`. Hardcoded text in Vue templates, script tags, or components is rejected by automated tests (`npm test`).
- **No Arabic in Backend Source Files:** Go source files (`*.go`) must contain no Arabic characters. All backend log entries, error messages, and internal strings must remain in English to maintain terminal compatibility and encoding consistency. This rule is enforced by an automated AST/string scan test in `go test ./...`.
- **Default Tailwind Shades:** Frontend styling must use standard, default Tailwind CSS color shades (such as `slate-700`, `slate-800`, `slate-900`). Non-standard, invented shade numbers are disallowed and verified by automated frontend tests.
- **Process Supervision Model:** Child process management uses `taskkill /F /T` and netstat-based port cleanup (`internal/winutil/cmd_windows.go`). Do not assume Windows Job Objects or persistent daemons.
- **Platform Scope:** Production code targets Windows 10/11 x64. Non-Windows source files are stubs intended solely to allow cross-compilation validation.
- **Test-Driven Changes:** New features or bug fixes must include corresponding tests (`*_test.go` in Go packages or Vitest tests in `frontend/test/`).
