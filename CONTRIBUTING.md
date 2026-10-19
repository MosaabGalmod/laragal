# Contributing to LaraGal

Thank you for your interest in contributing to LaraGal! We welcome contributions to make local PHP and Laravel development fast, reliable, and accessible for everyone.

## Development Setup

1. **Clone the repo:**
   ```bash
   git clone https://github.com/MosaabGalmod/laragal.git
   cd laragal
   ```

2. **Install frontend dependencies & build assets:**
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

3. **Run Go tests:**
   ```bash
   go test ./... -v
   ```

4. **Compile executable:**
   ```bash
   go build -o laragal.exe .
   ```

## Code Guidelines

- **Architecture:** Maintain strict separation between Go supervisor packages in `internal/` and the presentation layer in `frontend/`.
- **Test-Driven:** All new packages or bug fixes in `internal/` must include corresponding `*_test.go` unit tests.
- **BiDi & Localization:** All UI additions in `frontend/src/App.vue` must include translations in both `locales/ar.json` and `locales/en.json` conforming to RTL and LTR rules.
