# LaraGal

<div align="center">

![LaraGal Banner](https://raw.githubusercontent.com/MosaabGalmod/laragal/main/docs/assets/banner.png)

**Lightweight, zero-config local development and hosting suite for PHP, Go, Rust, .NET, Python, Node.js, Ruby, Java, and static sites on Windows.**

[![Version](https://img.shields.io/badge/Version-1.1.0-blue.svg)](https://github.com/MosaabGalmod/laragal/releases)
[![License: BSL 1.1](https://img.shields.io/badge/License-BSL_1.1-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Frontend-Vue_3_+_Tailwind_v4-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Multi-Runtime](https://img.shields.io/badge/Runtimes-PHP_%7C_Go_%7C_Rust_%7C_.NET_%7C_Python_%7C_Node_%7C_Ruby_%7C_Java-orange.svg)](#multi-language--runtime-support)
[![Platform](https://img.shields.io/badge/Platform-Windows_10%2F11_x64_Only-0078D6?logo=windows&logoColor=white)](#)

[English](README.md) • [العربية](#نظرة-عامة-بالعربية) • [Documentation](docs/) • [Architecture Spec](docs/superpowers/specs/2026-10-19-laragal-design.md)

</div>

---

## Overview

**LaraGal** is a local server and development suite built with **Go** and **Wails v2 (Vue 3 + Tailwind CSS)** for **Windows 10 and 11 (64-bit)** only. (Non-Windows builds compile stub implementations for build validation only). It provides local site serving, SSL termination, and service management for web development:

1. **Zero-Warning Local SSL (`https://*.test`):** Embedded internal Certificate Authority (CA) powered by Caddy that installs into the Windows Trust Store and Firefox NSS database with zero browser security warnings.
2. **Multi-Language & Reverse Proxy Engine:** Native serving for PHP (Laravel, WordPress, Bedrock) via FastCGI, and reverse proxy routing for external runtimes (Go, Node.js, Python, Rust, .NET, Ruby, Java) detected on the system `PATH`.
3. **Simultaneous Multi-PHP Isolation:** Runs multiple PHP FastCGI versions concurrently (`8.1`, `8.2`, `8.3`, `8.4`, `8.5`) on isolated local loopback ports (`:9081..:9085`). PHP 8.1 is end-of-life (flagged in the UI); PHP 8.2 receives security fixes only until 2026-12-31.
4. **Resource Footprint:** Single Go executable (≈ 16 MB) using the native Windows WebView2 runtime, with an idle private memory footprint of approximately 57 MB.
5. **Portable Supervised Services:** Bundled MySQL/MariaDB, Redis, and Mailpit (local email catcher) running in portable directories. Process supervision uses `taskkill /F /T` and netstat-based port cleanup (`internal/winutil/cmd_windows.go`) to prevent orphaned processes.
6. **Bilingual BiDi-Compliant Desktop UI:** Ergonomic Arabic (RTL) and English (LTR) desktop interface with dark and light theme support. Window controls (`✕`, `▢`, `—`) dynamically position on the physical left in Arabic mode and physical right in English mode.
7. **Built-in Design Tools:** Includes a responsive multi-screen viewport simulator, a WCAG AA/AAA color contrast analyzer, and curated design resources for frontend development.

---

## Multi-Language & Runtime Support

LaraGal detects project structures via `internal/projects/detector.go` and dynamically configures Caddy (`internal/services/caddy/caddy.go`). PHP FastCGI runtimes are managed directly by LaraGal.

> **Important:** LaraGal does **not** install Node.js, Python, Go, Rust, .NET, Ruby, or Java. It detects existing runtime binaries on the system `PATH` and reverse-proxies their local development servers to local `.test` domains.

| Project Type | Detection Signatures | Serving Architecture | Capabilities |
|---|---|---|---|
| **PHP & Laravel** | `artisan`, `composer.json`, `index.php` | PHP FastCGI (`:9081..:9085`) + Caddy | Multi-version FastCGI (`8.1`, `8.2`, `8.3`, `8.4`, `8.5`), project scaffolding, automatic `.env` synchronization. PHP 8.1 flagged as EOL; 8.2 security-only until 2026-12-31 |
| **WordPress & Bedrock** | `wp-config.php`, `wp-load.php`, Bedrock | PHP FastCGI + Caddy | Isolated docroot, database linking, local SSL |
| **Go Applications** | `go.mod` + (`web/`, `cmd/server/`, `templates/`) | Reverse Proxy (`127.0.0.1:8080`) + Static File Server | Caddy serves `/static/*` directly; routes application traffic to internal port with trusted `https://<app>.test` (requires Go on `PATH`) |
| **Rust Applications** | `Cargo.toml` (Actix-web, Axum, Rocket, Warp) | Reverse Proxy (`127.0.0.1:8080`) + Process Runner | Proxies application port to Caddy (requires Cargo/Rust on `PATH`) |
| **.NET / C#** | `*.csproj`, `Program.cs` (ASP.NET Core) | Reverse Proxy (`127.0.0.1:5000`) + Static `/wwwroot/*` | Static assets served from `wwwroot/`, reverse proxy to local port (requires .NET SDK on `PATH`) |
| **Python Web** | `manage.py`, `requirements.txt`, `wsgi.py`, `asgi.py` | Reverse Proxy (`127.0.0.1:8000`) + Static File Server | Detects Django, FastAPI, and Flask; proxies application port with direct `/static/*` serving (requires Python on `PATH`) |
| **Node.js & Frontend** | `package.json`, Vite, Next.js, Nuxt, Astro, Svelte | Caddy Static File Server / Reverse Proxy | Serves static build directories (`dist/`, `build/`, `out/`) or proxies local dev server (requires Node.js on `PATH`) |
| **Ruby on Rails** | `Gemfile`, `config.ru` | Reverse Proxy (`127.0.0.1:3000`) + Static `/public/*` | Routes Rails & Sinatra apps with static asset delivery from `public/` (requires Ruby on `PATH`) |
| **Java & Kotlin** | `pom.xml`, `build.gradle`, `build.gradle.kts` | Reverse Proxy (`127.0.0.1:8080`) + Caddy Auto-TLS | Detects Spring Boot, Quarkus, and Ktor backends with local domain routing (requires JDK on `PATH`) |
| **Flutter Web & Static** | `build/web/index.html`, `index.html` | Caddy `file_server` | Static file serving with HTTP/2, HTTP/3, and gzip/zstd compression |

---

## Comparison Matrix

> **Note on Comparisons:** Competitor memory usage and package sizes vary widely depending on installed extensions, modules, and operating system configuration. The values for LaraGal below represent measured figures from testing; competitor figures are listed qualitatively without unverified numeric claims.

| Feature | LaraGal | Laragon | Laravel Herd (Win) | XAMPP | Docker / Sail |
|---|---|---|---|---|---|
| **Idle Memory (Private Bytes)** | **≈ 57 MB** (measured) | Moderate (varies by modules) | Moderate to High (Electron) | Moderate (Apache + MySQL) | High (VM & container runtime) |
| **Binary / Footprint** | **≈ 16 MB** (single executable) | Installer package | Installer package | Installer package | Multi-gigabyte image storage |
| **Local SSL (`*.test`)** | **Automated via Caddy CA** | Unreliable / manual certs | Automated | Manual configuration | Manual / mkcert required |
| **Multi-Language Hosting** | **PHP + Proxied Go/Rust/.NET/Python/Node** | PHP primarily | PHP (Node via CLI) | PHP & Perl | Multi-container setups |
| **Concurrent Multi-PHP** | **Yes (ports 9081–9085)** | No (global switch only) | Paid (Herd Pro) | No | Yes (separate containers) |
| **Process Supervision** | **`taskkill /F /T` + port cleanup** | Windows Service / tray | Background service | Control panel | Docker daemon |
| **Managed DB & Mailpit** | **Included** | Partial (HeidiSQL) | Paid (Herd Pro) | phpMyAdmin only | Manual Compose configuration |
| **BiDi & Arabic UI** | **Native RTL/LTR Compliant** | English only | English only | English only | N/A |
| **License** | **BSL 1.1 (Free for Dev)** | Freeware | Commercial / Freemium | GPL | Open Source |

---

## Architecture & Core Subsystems

```
+-------------------------------------------------------------------------+
|                           LaraGal Desktop App                           |
|  +--------------------+-----------------------------------------------+ |
|  |  Wails (Go Backend)|  Frontend UI (Vue 3 + Tailwind + BiDi Core)   | |
|  |  - Process Supervisor (taskkill /F /T + netstat port cleanup)      | |
|  |  - Windows Hosts Manager (C:\Windows\System32\drivers\etc\hosts)   | |
|  |  - Windows Trust Store Injector (Cert:\LocalMachine\Root)          | |
|  |  - Dynamic Caddyfile Generator (FastCGI & Reverse Proxy)           | |
|  |  - Smart Project Detector (PHP, Go, Node.js, Static, Python)       | |
|  |  - On-Demand Runtime Downloader & Extractor (SHA-256 verified)    | |
|  |  - Project Scaffolder & .env Synchronizer                          | |
|  +--------------------+-----------------------------------------------+ |
+----------------------------+--------------------------------------------+
                             | Spawns & Supervises
                             v
+-------------------------------------------------------------------------+
|                     Isolated Portable Child Services                    |
|                                                                         |
|  +-------------------+  +-------------------+  +---------------------+  |
|  |   Caddy Server    |  |   PHP FastCGI     |  |    MySQL / MariaDB  |  |
|  |   Ports: 80, 443  |  |   8.1 -> :9081(EOL)| |    Port: 3306       |  |
|  |   (Auto-TLS CA &  |  |   8.2 -> :9082    |  |    Data: data/mysql |  |
|  |   Reverse Proxy)  |  |   8.3 -> :9083    |  |                     |  |
|  |                   |  |   8.4 -> :9084    |  |                     |  |
|  |                   |  |   8.5 -> :9085    |  |                     |  |
|  +---------+---------+  +---------+---------+  +----------+----------+  |
|            |                      |                       |             |
|  +---------+---------+  +---------+---------+             |             |
|  |      Redis        |  |      Mailpit      |             |             |
|  |   Port: 6379 (EOL)|  |   SMTP: :1025     |             |             |
|  |                   |  |   Web UI: :8025   |             |             |
|  +-------------------+  +-------------------+             |             |
+-----------------------------------------------------------+-------------+
```

---

## Design Tools

LaraGal includes built-in design utility panels accessible directly from the desktop interface:
- **Responsive Viewport Simulator:** Preview active sites across preset device viewports (mobile, tablet, laptop, desktop) with orientation toggling and custom dimension scaling. The preview checks framing headers (`X-Frame-Options` and CSP `frame-ancestors`) via `CheckFrameable` and explains when a project forbids iframe embedding, providing an external browser fallback.
- **WCAG Contrast Checker:** Real-time color contrast evaluation between foreground and background colors with AA and AAA compliance ratios under WCAG 2.1 standards.
- **Design Resources:** Curated quick-access links to typography, color palette tools, and accessibility guidelines.

---

## Security

LaraGal is built strictly for local development environments:
- **Loopback Binding Only:** Caddy site blocks and reverse proxy routes bind exclusively to `127.0.0.1` and `::1`. Local projects are not exposed to the local network by default.
- **Download Integrity Verification:** All downloaded runtime packages (Caddy, MariaDB, Mailpit, phpMyAdmin, Adminer, Redis, and PHP builds) are verified against hardcoded SHA-256 checksums before extraction. Live PHP patch upgrades verify archives against the hash published in official `windows.php.net` `releases.json` metadata.
- **Local Database Access:** Bundled database managers (phpMyAdmin and Adminer) are configured for automatic local login using the passwordless MySQL root account. These interfaces are accessible only over loopback hostnames (`phpmyadmin.test` and `adminer.test`) and must never be exposed to public networks.

---

## Updates

- **Live Official Sources & Verified Checksums:** Every managed runtime now queries a live official upstream source before downloading; nothing is downloaded without checksum verification:
  - **PHP:** Queries `windows.php.net/downloads/releases/releases.json` using the SHA-256 digest in the feed. New minors newer than the newest catalog minor (e.g., `8.6`, `9.0`) are offered automatically; older unlisted minors (`7.4`, `8.0`) are never offered. Loopback port is assigned deterministically as `9000 + major*10 + minor`.
  - **Caddy, Mailpit, Adminer, & Redis:** Monitored via the GitHub Releases API, where the checksum is the SHA-256 "digest" GitHub publishes per release asset. Results are cached for 1 hour to respect unauthenticated rate limits (60 requests/hour); if unreachable, falls back to the pinned catalog.
  - **MariaDB:** Queries the `downloads.mariadb.org` REST API (SHA-256). Patch updates within the installed series install normally. Newer Stable LTS series (currently `11.4`, `11.8`, `12.3`) are offered as major upgrades requiring confirmation, followed by `mariadb-upgrade` on next start. Preview, RC, and Rolling series are never offered.
  - **phpMyAdmin:** Checked against `phpmyadmin.net` `version.json` with `.sha256` fetched from `files.phpmyadmin.net`. Shows a compatibility note when the serving PHP version is outside phpMyAdmin's supported range (currently: 5.2.3 supports PHP <8.4, while LaraGal serves it with PHP 8.4).
- **Major-Version & Provider Changes:** Upgrades involving a different major version or a provider change (Adminer `4.8.1` → `6.x`, Redis `tporadowski 5.0` → `redis-windows 8.x`, or a new MariaDB LTS series) require explicit confirmation in a dialog.
- **Atomic Installs & Rollback Safety:** Installations are atomic: extracted into a staging folder, config files (`php.ini`, `my.ini`, `redis.conf`) preserved, old folder swapped out and restored on failure, and database files are never deleted by cleanup.
- **Honest Security Note:** A checksum fetched from the same vendor as the file detects corruption and tampering in transit, but does not protect against a compromised vendor account. Only Caddy additionally publishes Sigstore signatures, which LaraGal does not verify yet.
- **Application Releases:** LaraGal itself still has no auto-update or update notification. New versions are published on [GitHub Releases](https://github.com/MosaabGalmod/laragal/releases) and installed manually by replacing the executable.

---

## Quickstart & Building From Source

### Prerequisites
- Windows 10 or 11 (64-bit)
- [Go 1.22+](https://go.dev/)
- [Node.js 18+](https://nodejs.org/) & npm

### Automated Build Script (Recommended)
Run the bundled build script from the project root:
```bat
build.bat
```

### Manual Build Instructions

1. **Clone the repository:**
   ```bash
   git clone https://github.com/MosaabGalmod/laragal.git
   cd laragal
   ```

2. **Build the Vue 3 frontend:**
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

3. **Run verification commands:**
   Frontend checks:
   ```bash
   cd frontend
   npm run lint
   npm test
   cd ..
   ```
   Backend checks:
   ```bash
   go vet ./...
   go test ./...
   ```
   Continuous Integration (CI) executes these frontend and backend checks automatically on tag push.

4. **Compile the production executable:**
   ```bash
   go build -tags "desktop,production" -ldflags "-H windowsgui -s -w" -o laragal.exe .
   ```

5. **Run LaraGal:**
   ```powershell
   .\laragal.exe
   ```

---

## نظرة عامة بالعربية

<div dir="rtl" lang="ar">

برنامج `LaraGal` هو بيئة تطوير واستضافة محلية لمطوري تطبيقات الويب على نظام التشغيل `Windows 10` و `Windows 11` بمعمارية `(64-bit)` فقط (بناء البرنامج على الأنظمة الأخرى ينتج هياكل شكلية مخصصة للتحقق من التجميع فقط). يجمع البرنامج بين الواجهة الخلفية بلغة `Go` وخادم الويب `Caddy` والواجهة المكتبية المبنية بـ `Vue 3` و `Tailwind CSS` عبر إطار `Wails v2`.

### الخصائص الأساسية:

* شهادات أمان محلية موثوقة لنطاقات `https://*.test`: تعتمد على هيئة الشهادات الداخلية المدمجة في خادم `Caddy` مع تثبيت تلقائي في متجر شهادات ويندوز ومتجر شهادات متصفح فايرفوكس لتفادي تحذيرات الأمان.
* إدارة ومراقبة العمليات: يعتمد المشرف على العمليات على الأمر `taskkill /F /T` مع تنظيف منافذ الاتصال عبر `netstat` في ملف `internal/winutil/cmd_windows.go` لإغلاق الخدمات التابعة دون ترك عمليات معلقة. لا يستخدم البرنامج كائنات مهام ويندوز `(Windows Job Objects)`.
* تشغيل متزامن لإصدارات `PHP FastCGI`: يدعم تشغيل الإصدارات من `8.1` إلى `8.5` على المنافذ من `9081` إلى `9085`. إصدار `PHP 8.1` منتهي الدعم رسميًا `(EOL)` وتظهر شارة توضح ذلك في الواجهة، بينما إصدار `PHP 8.2` يقتصر على التحديثات الأمنية حتى تاريخ `2026-12-31`.
* توجيه بيئات التشغيل الخارجية: لا يقوم برنامج `LaraGal` بتثبيت `Node.js` أو `Python` أو `Go` أو `Rust` أو `.NET` أو `Ruby` أو `Java`، بل يكتشف وجودها ضمن مسار النظام `(PATH)` ويوجه الطلبات إلى خوادم التطوير المحلية الخاصة بها عبر البروكسي العكسي في `Caddy`.
* استهلاك الموارد وحجم الملف: يبلغ حجم الملف التنفيذي المترجم نحو `16 MB`، ويبلغ استهلاك الذاكرة الخاصة في وضع الخمول قرابة `57 MB` بالاعتماد على محرك `WebView2`.
* خدمات التطوير المرفقة: يوفر البرنامج حزمًا محمولة لكل من خادم قواعد البيانات `MySQL/MariaDB` وخادم `Redis` وأداة اختبار البريد `Mailpit` مع واجهة ويب على المنفذ `8025`.
* واجهة مكتبية ثنائية اللغة: تدعم اللغتين العربية `(RTL)` والإنجليزية `(LTR)` مع الوضعين الداكن والفاتح، وتتموضع أزرار التحكم بالنافذة على اليسار الفعلي في الوضع العربي وعلى اليمين الفعلي في الوضع الإنجليزي.
* أدوات التصميم المدمجة: تتضمن نافذة لمحاكاة العرض على مقاسات الشاشات المختلفة مع فحص ترويسات التأطير `(X-Frame-Options)` وسياسة أمان المحتوى `(CSP frame-ancestors)` وتوضيح سبب تعذر العرض عند حظر المشروع للتضمين، وأداة لفحص تباين الألوان وفق معايير إمكانية الوصول `(WCAG AA/AAA)`، وروابط سريعة لأدلة التصميم.

### الأمان:

* قصر الاستماع على المضيف المحلي: تقتصر كتل المواقع وقواعد التوجيه في خادم `Caddy` على العناوين `127.0.0.1` و `::1` ولا تُفتح للشبكة الخارجية افتراضيًا.
* التحقق من سلامة التنزيلات: تُفحص كافة الحزم المنزلة بمطابقة بصمة `SHA-256` المحددة مسبقًا في الكود. وتُفحص ترقيات `PHP` بمطابقة البصمة المنشورة رسميًا في ملف `releases.json` على موقع `windows.php.net`.
* واجهات إدارة قواعد البيانات: يُسجل كل من `phpMyAdmin` و `Adminer` الدخول تلقائيًا بحساب الجذر `root` دون كلمة مرور لأغراض التطوير المحلي فقط، ويحظر كشفهما على أي شبكة عامة.

### التحديثات:

* فحص المصادر الرسمية الحية لجميع بيئات التشغيل: تفحص كافة الخدمات المدارة مصادر إصداراتها الرسمية عبر الإنترنت، ولا يجري تنزيل أي حزمة دون التحقق من بصمة التجزئة `(SHA-256)`:
  * حزم `PHP`: يجلب البرنامج التحديثات من ملف `releases.json` على موقع `windows.php.net` بالاعتماد على بصمة `(SHA-256)` المرفقة بالتغذية. تُعرض الإصدارات الفرعية الأحدث من أعلى إصدار مسجل في الكتالوج (مثل `8.6` و `9.0`) تلقائيًا، بينما تُحجب الإصدارات الأقدم غير المدرجة (مثل `7.4` و `8.0`). ويُحسب منفذ الاستماع تلقائيًا وفق المعادلة `(9000 + major*10 + minor)`.
  * خدمات `Caddy` و `Mailpit` و `Adminer` و `Redis`: تُفحص عبر واجهة `GitHub Releases API` بالاعتماد على بصمة `(digest)` المنشورة لكل أصل برمجي من المنبع. تُحفظ النتائج مؤقتًا لمدة ساعة واحدة لتفادي تجاوز الحد المسموح للاستعلام غير الموثق وهو 60 طلبًا في الساعة، مع الرجوع التلقائي إلى الكتالوج المثبت عند تعذر الوصول إلى الشبكة.
  * خادم `MariaDB`: يُفحص عبر واجهة `downloads.mariadb.org REST API` بمطابقة بصمة `(SHA-256)`. تُثبت التحديثات التصحيحية ضمن السلسلة الحالية بصورة اعتيادية، بينما تُعرض السلاسل المستقرة طويلة الدعم `(Stable LTS)` الأحدث (حاليًا `11.4` و `11.8` و `12.3`) كترقيات رئيسية تتطلب تأكيدًا يدويًا، ويتبعها تشغيل أداة `mariadb-upgrade` عند الإقلاع التالي للخدمة. وتُحجب إصدارات المعاينة والتطوير `(Preview/RC/Rolling)` بصورة دائمة.
  * أداة `phpMyAdmin`: تُفحص عبر ملف `version.json` على `phpmyadmin.net` مع ملف `.sha256` من `files.phpmyadmin.net`، مع إظهار ملاحظة توافقية عندما يكون إصدار `PHP` المشغِّل خارج نطاق الدعم الرسمي للمكتبة (يدعم إصدار `5.2.3` حاليًا `PHP` أقل من `8.4`، بينما يشغله البرنامج عبر `PHP 8.4`).
* تأكيد الترقيات الرئيسية وتغيير المزود: تتطلب التغييرات في الإصدار الرئيسي أو الانتقال لمزود مختلف (مثل الانتقال في `Adminer` من `4.8.1` إلى `6.x`، أو في `Redis` من `tporadowski 5.0` إلى `redis-windows 8.x`، أو الانتقال لسلسلة `MariaDB` جديدة) موافقة صريحة من المستخدم في نافذة حوارية قبل البدء.
* التثبيت الذري وسلامة الاسترجاع: تجري عمليات التثبيت بصورة ذرية؛ إذ تُفك الحزم داخل مجلد مؤقت `(staging)`، وتُحفظ ملفات الإعداد القائمة `(php.ini)` و `(my.ini)` و `(redis.conf)`، ويُستبدل المجلد القديم مع إمكانية استرجاعه تلقائيًا عند أي تعثر؛ ولا تحذف عمليات التنظيف ملفات قواعد البيانات أبدًا.
* إشعار أمني بشأن سلامة البصمات: تكتشف بصمات التجزئة المجلوبة من نفس مزود الملفات أي تلف أو تلاعب يطرأ أثناء النقل عبر الشبكة، لكنها لا تحمي من مخاطر اختراق حساب المزود نفسه. وينفرد خادم `Caddy` بنشر تواقيع `Sigstore` إضافية لا يتحقق منها البرنامج حاليًا.
* تحديث برنامج `LaraGal`: لا يحتوي البرنامج نفسه على خاصية التحديث التلقائي أو إشعارات التحديث حتى الآن، وتُثبت الإصدارات الجديدة يدويًا عبر تنزيل الملف التنفيذي واستبداله من صفحة الإصدارات في `GitHub Releases`.

</div>

---

## License & Commercial Rights

LaraGal is licensed under the **Business Source License 1.1 (BSL 1.1)**:
- **Free for all developers, students, and organizations** for internal development, local testing, and non-competing use.
- Commercial redistribution, white-labeling, or packaging as a competing cloud-hosted developer service requires a commercial agreement.
- See the full [LICENSE](LICENSE) file for details.
