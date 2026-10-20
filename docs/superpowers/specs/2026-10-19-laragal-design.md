# LaraGal Architecture & Design Specification

- **Project Name:** LaraGal
- **Target OS:** Windows (Primary), with cross-platform architecture readiness for macOS and Linux
- **Date:** 2026-10-19
- **Status:** v1.0.0 — Released / Production Ready
- **Version:** 1.0.0 (First Official Release)

---

## 1. Executive Summary & Problem Statement

### 1.1 The Challenge
Local PHP/Laravel development environments on Windows face common drawbacks:
- **XAMPP:** Heavyweight, lacks automated virtual hosts, lacks multi-PHP support, and requires complex manual configuration for SSL.
- **Laragon:** Fast and beloved, but development has stagnated; SSL certificates often trigger browser warnings on modern Windows/Chrome versions, and it is Windows-only with a legacy architecture.
- **Laravel Herd:** Proprietary, features like multi-version switching and services require paid licenses on Windows.
- **LaraStack:** Emerging, but needs a robust, open, lightweight, and truly zero-config solution for local SSL domains.

### 1.2 The LaraGal Vision
LaraGal is an ultra-fast, lightweight, open-source local development suite built with Go and modern web technologies. It provides:
1. **Zero-Config Local SSL (`https://*.test`):** Powered by an embedded Caddy server with an automated local Root Certificate Authority (CA) installed directly into the Windows Trust Store and Firefox NSS database.
2. **True Process Isolation & Multi-PHP:** Spawns isolated FastCGI listeners for PHP 8.1, 8.2, 8.3, and 8.4 simultaneously, allowing different projects to use different PHP versions concurrently.
3. **Multi-Language & Reverse Proxy Engine:** Smart auto-detection and hosting for Go standalone web applications (with reverse-proxying and `/static/*` routing), Node.js, static HTML sites, and Python web projects alongside PHP/Laravel/WordPress.
4. **Automated Host Resolution:** Instant virtual host binding (`<project>.test -> 127.0.0.1`) via Windows `hosts` synchronization.
5. **All-in-One Portable Services:** Bundled MySQL/MariaDB, Redis, and Mailpit (local email testing) that run as supervised child processes with zero system pollution.
6. **1-Click Project Scaffolding:** Instant setup for Laravel and WordPress with automatic `.env` synchronization (database creation, Mailpit configuration, and trusted SSL URL).
7. **Modern, Bilingual, Responsive GUI:** Built with Vue 3 and Tailwind CSS, featuring full Arabic (RTL) and English (LTR) support, BiDi-compliant window controls on physical left in Arabic, Dark/Light modes, and a seamless System Tray integration.

---

## 2. Technology Stack & Architectural Decisions

### 2.1 Core Runtime & Desktop Layer
- **Core Engine (Backend):** `Go` (version 1.22+) using the `Wails` v2 framework.
  - *Why Go?* Native compiled binary, minimal idle memory consumption (~25–35 MB), fast cold start (< 1 second), cross-platform compatibility, and rock-solid concurrent process supervision (`goroutines`).
  - *Why Wails over Electron?* Avoids bundling Chromium and Node.js runtimes. Utilizes the native Windows `WebView2` runtime, resulting in an installer under 20 MB and negligible memory overhead.
- **Presentation Layer (Frontend):** `Vue 3` (Composition API) + `Vite` + `Tailwind CSS` + `Lucide Icons`.
  - Fully bilingual (`AR` RTL / `EN` LTR) with zero text leakage.
  - Dark and Light theme support conforming to WCAG AA accessibility standards.
  - Minimalistic, production-grade aesthetic (anti-AI-slop compliant: no excessive gradients, authentic layout).

### 2.2 Web Server & SSL Engine: Caddy
- **Web Server:** `Caddy Server` (portable binary managed by LaraGal).
- **SSL / TLS Engine:** Caddy’s internal automated certificate manager (`internal ACME / local CA`).
  - On initial setup, LaraGal prompts for elevation (UAC) once to install Caddy's Root CA (`root.crt`) into `Cert:\LocalMachine\Root` and Firefox NSS stores.
  - Caddy dynamically issues and renews valid TLS certificates for all `*.test` domains on port 443 with zero user intervention.
  - Supports `HTTP/2` and `HTTP/3` natively.
- **FastCGI Proxying:** Caddy dynamically proxies `.php` requests to the assigned PHP FastCGI port for each specific project.

### 2.3 Process Supervisor & Service Registry
LaraGal acts as a process supervisor for all managed dependencies:
- Spawns child processes using Go's `os/exec`.
- Attaches to Windows Job Objects to guarantee clean termination of all child processes when LaraGal exits (preventing orphaned `php-cgi.exe` or `mysqld.exe` processes).
- Periodic heartbeat health checks on assigned TCP ports.

---

## 3. System Architecture & Component Diagram

```
+-------------------------------------------------------------------------+
|                           LaraGal Desktop App                           |
|  +--------------------+-----------------------------------------------+ |
|  |  Wails (Go Backend)|  Frontend UI (Vue 3 + Tailwind + BiDi Core)   | |
|  |  - Process Supervisor                                              | |
|  |  - Windows Hosts Manager                                           | |
|  |  - Windows Trust Store Injector (CA)                               | |
|  |  - Caddy Dynamic Config Builder                                    | |
|  |  - Project Scanner & .env Sync                                     | |
|  +--------------------+-----------------------------------------------+ |
+----------------------------+--------------------------------------------+
                             | Spawns & Supervises
                             v
+-------------------------------------------------------------------------+
|                     Isolated Portable Child Services                    |
|                                                                         |
|  +-------------------+  +-------------------+  +---------------------+  |
|  |   Caddy Server    |  |   PHP FastCGI     |  |    MySQL / MariaDB  |  |
|  |   Ports: 80, 443  |  |   8.1 -> :9081    |  |    Port: 3306       |  |
|  |   (Auto-TLS CA)   |  |   8.2 -> :9082    |  |    Data: data/mysql |  |
|  |                   |  |   8.3 -> :9083    |  |                     |  |
|  |                   |  |   8.4 -> :9084    |  |                     |  |
|  +---------+---------+  +---------+---------+  +----------+----------+  |
|            |                      |                       |             |
|  +---------+---------+  +---------+---------+             |             |
|  |      Redis        |  |      Mailpit      |             |             |
|  |   Port: 6379      |  |   SMTP: :1025     |             |             |
|  |                   |  |   Web UI: :8025   |             |             |
|  +-------------------+  +-------------------+             |             |
+-----------------------------------------------------------+-------------+
```

---

## 4. Detailed Subsystem Specifications

### 4.1 SSL & Local Certificate Management (`internal/certs`)
1. **Root CA Generation:**
   - Caddy automatically initializes its internal CA upon first launch and saves `root.crt` to `data/caddy/pki/authorities/local/root.crt`.
2. **Windows Trust Store Injection:**
   - LaraGal executes PowerShell/CertUtil with elevated rights:
     `certutil -addstore -f "Root" "path\to\root.crt"`
   - Verifies fingerprint existence in `Cert:\LocalMachine\Root` before triggering UAC.
3. **Firefox NSS Trust Injection:**
   - Detects Firefox profiles (`%APPDATA%\Mozilla\Firefox\Profiles\*`) and uses `certutil` or direct SQLite certificate database update (`cert9.db`) to trust the local CA.
4. **Certificate Issuance:**
   - Automatic wildcard for `*.test` or per-domain SANs in Caddyfile:
     ```caddyfile
     *.test {
         tls internal
         ...
     }
     ```

### 4.2 Windows Hosts Manager (`internal/hosts`)
- **File Location:** `C:\Windows\System32\drivers\etc\hosts`.
- **LaraGal Managed Block:** All entries are bounded within clear markers to avoid touching user custom entries:
  ```hosts
  # --- BEGIN LARAGAL MANAGED DOMAINS ---
  127.0.0.1  myproject.test
  127.0.0.1  shop.test
  # --- END LARAGAL MANAGED DOMAINS ---
  ```
- **Sync Routine:** When a site is created, renamed, or deleted, LaraGal updates this block.
- **Alternative / Optional Mode:** Built-in loopback DNS proxy on `127.0.0.1:53` responding `127.0.0.1` to any query ending in `.test`.

### 4.3 Multi-PHP Manager (`internal/services/php`)
- **Structure:**
  Each PHP version lives in `bin/php/php-<version>-nts-Win32/`:
  - Contains `php.exe`, `php-cgi.exe`, and a pre-configured `php.ini`.
  - Common extensions pre-enabled: `curl`, `mbstring`, `openssl`, `pdo_mysql`, `pdo_sqlite`, `pdo_pgsql`, `fileinfo`, `gd`, `zip`, `intl`, `redis`.
- **FastCGI Daemonization:**
  - Each configured PHP version runs via:
    `php-cgi.exe -b 127.0.0.1:908x -c bin/php/php-<version>/php.ini`
  - Environment variables (`PHP_FCGI_MAX_REQUESTS=0`) configured to prevent worker shutdown.
- **Version Mapping:**
  - Projects can define their version via `.php-version` file (e.g. `8.3`) or via LaraGal GUI project settings.

### 4.4 Caddy Configuration Generator (`internal/services/caddy`)
- Automatically compiles a `Caddyfile` from active projects:
  ```caddyfile
  {
      local_certs
      auto_https disable_redirects
  }

  myproject.test {
      root * "D:/projects/myproject/public"
      encode gzip zstd
      php_fastcgi 127.0.0.1:9083
      file_server
      tls internal
  }

  wordpress.test {
      root * "D:/projects/wordpress"
      encode gzip zstd
      php_fastcgi 127.0.0.1:9082
      file_server
      tls internal
  }
  ```
- Triggered dynamically via Caddy's admin API (`POST http://localhost:2019/load`) for zero-downtime configuration reload.

### 4.5 Project Scaffolder & Auto-Configuration (`internal/projects`)
When creating a new project:
1. **Laravel:**
   - Runs `composer create-project laravel/laravel <name>` using the selected PHP version and bundled Composer.
   - Automatically edits `.env`:
     - `APP_URL=https://<name>.test`
     - `DB_CONNECTION=mysql`
     - `DB_HOST=127.0.0.1`
     - `DB_PORT=3306`
     - `DB_DATABASE=<name>`
     - `DB_USERNAME=root`
     - `DB_PASSWORD=`
     - `MAIL_MAILER=smtp`
     - `MAIL_HOST=127.0.0.1`
     - `MAIL_PORT=1025`
   - Executes `CREATE DATABASE IF NOT EXISTS <name>;` via MySQL client.
2. **WordPress:**
   - Downloads latest WordPress core, creates `<name>` database in MySQL, generates `wp-config.php` with appropriate salts and DB settings.

---

## 5. Directory Layout

```
d:\projects\laragal/
├── app/                        # Go (Wails) Backend
│   ├── cmd/                    # CLI & Desktop bootstrap
│   ├── internal/
│   │   ├── certs/              # CA creation & OS Trust injection
│   │   ├── config/             # LaraGal state & user preferences
│   │   ├── hosts/              # Windows hosts file read/write
│   │   ├── process/            # Process lifecycle & Windows Job Objects
│   │   ├── projects/           # Project scanner, templates & .env writer
│   │   └── services/           # Caddy, PHP, MySQL, Redis, Mailpit controllers
│   ├── app.go                  # Wails bridge methods exposed to UI
│   ├── main.go                 # App entry point
│   ├── go.mod
│   └── go.sum
├── frontend/                   # Vue 3 Desktop User Interface
│   ├── src/
│   │   ├── assets/             # SVGs, fonts, base styles
│   │   ├── components/         # ServiceCards, SiteRow, WizardModal, Settings
│   │   ├── composables/        # useI18n, useTheme, useServices
│   │   ├── locales/            # ar.json, en.json (Zero text leakage)
│   │   ├── stores/             # Pinia stores for services and sites
│   │   ├── App.vue
│   │   └── main.ts
│   ├── index.html
│   ├── package.json
│   ├── tailwind.config.ts
│   └── vite.config.ts
├── bin/                        # Portable runtimes (auto-downloaded or packaged)
│   ├── caddy/
│   ├── php/
│   ├── mysql/
│   ├── redis/
│   └── mailpit/
├── data/                       # User data & databases
│   ├── caddy/
│   ├── mysql/
│   └── logs/
├── docs/                       # Specifications and developer documentation
└── wails.json                  # Wails project definition
```

---

## 6. Security, Permissions & Error Handling

1. **UAC Elevation Strategy:**
   - Modifying `C:\Windows\System32\drivers\etc\hosts` and injecting into the `LocalMachine` Root certificate store requires administrative privileges.
   - *Design Decision:* LaraGal provides a lightweight helper binary (`laragal-privilege-helper.exe`) invoked only when hosts or Root CA changes occur, avoiding the need to run the entire GUI app as Administrator continuously.
2. **Port Conflict Detection:**
   - Before launching any service (e.g. MySQL on 3306, Caddy on 80/443), LaraGal checks TCP port availability (`net.DialTimeout`).
   - If a port is occupied by another service (e.g. existing IIS or Skype on port 80), the UI clearly displays the conflicting Process ID (PID) and process name with an option to terminate or rebind.
3. **Graceful Shutdown:**
   - Windows Job Objects ensure that even if the LaraGal UI crashes or is terminated forcefully, the Windows kernel automatically terminates all child processes (PHP, MySQL, Caddy).

---

## 7. Implementation Roadmap & Milestones

- **Phase 1: Core Foundation & Supervisor**
  - Setup Wails v2 with Go backend + Vue 3 frontend.
  - Implement process supervisor with Windows Job Objects.
  - Portable Caddy integration + automated Root CA installation to Windows Certificate Store.
- **Phase 2: Host Resolution & Multi-PHP FastCGI**
  - Windows hosts file synchronization for `*.test`.
  - Multi-version PHP runner (`php-cgi.exe` listeners for 8.1, 8.2, 8.3, 8.4).
  - Dynamic Caddy reverse proxy generation for sites.
- **Phase 3: Database, Redis & Mail Services**
  - Portable MySQL integration + data directory initialization.
  - Redis portable integration.
  - Mailpit integration (SMTP & Web UI embed/link).
- **Phase 4: Scaffolding, UI Polish & Packaging**
  - Laravel & WordPress one-click creation wizard with automatic `.env` and DB creation.
  - System Tray icon with quick actions (Start/Stop All, Open Sites, Switch PHP).
  - Bilingual interface (Arabic RTL / English LTR) and Dark/Light themes.
  - Portable ZIP and NSIS installer packaging.
