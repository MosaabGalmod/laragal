# LaraGal 🚀

<div align="center">

![LaraGal Banner](https://raw.githubusercontent.com/MosaabGalmod/laragal/main/docs/assets/banner.png)

**The ultra-fast, lightweight, zero-config trusted SSL multi-runtime local development and hosting suite for PHP, Go, Rust, .NET, Python, Node.js, Ruby, Java, and Flutter Web.**

[![Version](https://img.shields.io/badge/Release-v1.0.0_(First_Official)-blue.svg)](https://github.com/MosaabGalmod/laragal/releases)
[![License: BSL 1.1](https://img.shields.io/badge/License-BSL_1.1-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Frontend-Vue_3_+_Tailwind_v4-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Multi-Runtime](https://img.shields.io/badge/Runtimes-PHP_%7C_Go_%7C_Rust_%7C_.NET_%7C_Python_%7C_Node_%7C_Ruby_%7C_Java-orange.svg)](#-multi-language--runtime-support)
[![Platform](https://img.shields.io/badge/Platform-Windows_(macOS_%26_Linux_Ready)-0078D6?logo=windows&logoColor=white)](#)

[English](README.md) • [العربية](#-نظرة-عامة-بالعربية-الإصدار-100) • [Documentation](docs/) • [Architecture Spec](docs/superpowers/specs/2026-10-19-laragal-design.md)

</div>

---

> **Release Note (v1.0.0):**  
> Version `1.0.0` marks the first official, fully functional, and production-tested release of LaraGal. All earlier pre-release drafts and experimental builds were incomplete and are deprecated in favor of this release.

---

## ⚡ Overview

**LaraGal** is a modern, blazing-fast local server and development suite built with **Go** and **Wails v2 (Vue 3 + Tailwind CSS)**. Designed as an open, ultra-efficient alternative to legacy tools like *Laragon*, *XAMPP*, and commercial suites like *Laravel Herd*, LaraGal delivers zero-friction local hosting on Windows:

1. **Zero-Warning Local SSL (`https://*.test`):** Embedded internal Certificate Authority (CA) powered by Caddy that installs into the Windows Trust Store and Firefox NSS database with zero browser security warnings.
2. **Multi-Language & Reverse Proxy Engine:** Native support for PHP (Laravel, WordPress, Bedrock), standalone Go web applications with built-in reverse proxy routing, Node.js frontends, static HTML sites, and Python web assets.
3. **Simultaneous Multi-PHP Isolation:** Runs multiple PHP FastCGI versions concurrently (`8.1`, `8.2`, `8.3`, `8.4`) on isolated local loopback ports (`:9081..:9084`).
4. **Ultra-Low Memory Footprint:** Native Go engine and Windows WebView2 runtime—uses only **25–35 MB RAM** at idle (compared to 250 MB+ for Electron and 2–4 GB for Docker/WSL2).
5. **Portable Supervised Services:** Bundled MySQL/MariaDB, Redis, and Mailpit (local email catcher) running cleanly in portable directories with no Windows registry pollution.
6. **Bilingual BiDi-Compliant Desktop UI:** Ergonomic Arabic (RTL) and English (LTR) desktop interface. Window controls (`✕`, `▢`, `—`) dynamically position on the physical left in Arabic mode and physical right in English mode.

---

## 🌐 Multi-Language & Runtime Support

LaraGal features an intelligent project detector (`internal/projects/detector.go`), process supervisor (`app.go`), and dynamic Caddy orchestration (`internal/services/caddy/caddy.go`) to automatically detect, configure, and serve projects across 9 different runtime ecosystems:

| Project Type | Detection Signatures | Serving Architecture | Capabilities |
|---|---|---|---|
| **PHP & Laravel** | `artisan`, `composer.json`, `index.php` | PHP FastCGI (`:9081..:9084`) + Caddy | Multi-version FastCGI (`8.1`, `8.2`, `8.3`, `8.4`), 1-click scaffolding, automatic `.env` synchronization |
| **WordPress & Bedrock** | `wp-config.php`, `wp-load.php`, Bedrock | PHP FastCGI + Caddy | Isolated docroot, database linking, instant SSL |
| **Go Applications** | `go.mod` + (`web/`, `cmd/server/`, `templates/`) | Reverse Proxy (`127.0.0.1:8080`) + Static File Server | Caddy serves `/static/*` directly; routes application traffic to internal port with trusted `https://<app>.test` |
| **Rust Applications** | `Cargo.toml` (Actix-web, Axum, Rocket, Warp) | Reverse Proxy (`127.0.0.1:8080`) + Process Runner | Automated `cargo run --release` / `<app>.exe` supervisor and instant Caddy proxying |
| **.NET / C#** | `*.csproj`, `Program.cs` (ASP.NET Core) | Reverse Proxy (`127.0.0.1:5000`) + Static `/wwwroot/*` | Direct static assets serving from `wwwroot/`, `dotnet run` supervisor, and trusted SSL |
| **Python Web** | `manage.py`, `requirements.txt`, `wsgi.py`, `asgi.py` | Reverse Proxy (`127.0.0.1:8000`) + Static File Server | Detects Django, FastAPI, and Flask; proxies application port with direct `/static/*` asset serving |
| **Node.js & Frontend** | `package.json`, Vite, Next.js, Nuxt, Astro, Svelte | Caddy Static File Server / Reverse Proxy | Auto-detects `dist/`, `build/`, `public/`, or `out/` build directories or dev server |
| **Ruby on Rails** | `Gemfile`, `config.ru` | Reverse Proxy (`127.0.0.1:3000`) + Static `/public/*` | Routes Rails & Sinatra apps with direct static assets delivery from `public/` |
| **Java & Kotlin** | `pom.xml`, `build.gradle`, `build.gradle.kts` | Reverse Proxy (`127.0.0.1:8080`) + Caddy Auto-TLS | Detects Spring Boot, Quarkus, and Ktor backends with zero-config local domain routing |
| **Flutter Web & Static** | `build/web/index.html`, `index.html` | Caddy `file_server` | Ultra-fast HTTP/2, HTTP/3, and `gzip`/`zstd` compression with zero backend overhead |

---

## 📊 Comparison Matrix

| Feature | LaraGal (v1.0.0) | Laragon | Laravel Herd (Win) | XAMPP | Docker / Sail |
|---|---|---|---|---|---|
| **Idle Memory (RAM)** | **~30 MB** | ~50 MB | ~150 MB | ~140 MB | **2000–4000 MB+** |
| **Local SSL (`*.test`)** | **100% Trusted & Automated** | Unreliable / Red Warnings | Automated | Painful / Manual | Complex / Manual |
| **Multi-Language Hosting** | **PHP, Go, Rust, .NET, Python, Node, Ruby, Java, Flutter** | PHP only | PHP only (Node via CLI) | PHP & Perl | Multi-container overhead |
| **Concurrent Multi-PHP** | **Yes (Per-site simultaneous)** | No (Global switch only) | **Paid (Herd Pro)** | No | Yes (Container overhead) |
| **Process Supervision** | **Built-in Multi-Runtime Runner** | No | No | No | Compose dependent |
| **Binary Size** | **~7.3 MB** | ~180 MB | ~120 MB | ~160 MB | Multi-Gigabyte |
| **Managed DB & Mailpit** | **Included Free** | Partial (HeidiSQL) | **Paid (Herd Pro)** | phpMyAdmin only | Manual Compose |
| **BiDi & Arabic UI** | **Native RTL/LTR Compliant** | English only | English only | English only | N/A |
| **License** | **BSL 1.1 (Free for Dev)** | Freeware (Stagnant) | Freemium ($99-$199/yr) | GPL | Open Source |

---

## 🛠️ Architecture & Core Subsystems

```
+-------------------------------------------------------------------------+
|                           LaraGal Desktop App                           |
|  +--------------------+-----------------------------------------------+ |
|  |  Wails (Go Backend)|  Frontend UI (Vue 3 + Tailwind + BiDi Core)   | |
|  |  - Process Supervisor (Windows Job Objects)                        | |
|  |  - Windows Hosts Manager (C:\Windows\System32\drivers\etc\hosts)   | |
|  |  - Windows Trust Store Injector (Cert:\LocalMachine\Root)          | |
|  |  - Dynamic Caddyfile Generator (FastCGI & Reverse Proxy)           | |
|  |  - Smart Project Detector (PHP, Go, Node.js, Static, Python)       | |
|  |  - On-Demand Runtime Downloader & Extractor                        | |
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
|  |   Ports: 80, 443  |  |   8.1 -> :9081    |  |    Port: 3306       |  |
|  |   (Auto-TLS CA &  |  |   8.2 -> :9082    |  |    Data: data/mysql |  |
|  |   Reverse Proxy)  |  |   8.3 -> :9083    |  |                     |  |
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

## 🚀 Quickstart & Building From Source

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

3. **Run unit & integration tests:**
   ```bash
   go test ./...
   ```

4. **Compile the production executable:**
   ```bash
   go build -tags "desktop,production" -ldflags "-H windowsgui -s -w" -o laragal.exe .
   ```

5. **Run LaraGal:**
   ```powershell
   .\laragal.exe
   ```

---

## 🇸🇦 نظرة عامة بالعربية (الإصدار 1.0.0)

برنامج **`LaraGal`** هو بيئة خادم محلي خفيفة وعصرية لمطوري تطبيقات الويب على نظام ويندوز. يجمع بين سرعة لغة `(Go)` ومرونة خادم `(Caddy)` وواجهة مكتبية أنيقة مبنية بـ `(Vue 3)` و`(Tailwind CSS)`.

> **تنويه الإصدار الأول `(v1.0.0)`**:  
> يمثل هذا الإصدار أول انطلاقة رسمية ومستقرة للبرنامج بعد اكتمال كافة الفحوصات ومعالجة مشاكل الإصدارات التجريبية السابقة غير العاملة.

### أبرز مميزات الإصدار 1.0.0:

1. **دعم استضافة شامل لكافة اللغات وبيئات التشغيل `(Multi-Runtime Architecture)`**:
   - **تطبيقات `(PHP / Laravel / WordPress)`**: تشغيل متزامن وعزل تام لإصدارات `PHP 8.1` و `8.2` و `8.3` و `8.4` على منافذ مستقلة، مع معالج مشاريع ينشئ المشروع ويربط قاعدة البيانات وملف `.env` بنقرة واحدة.
   - **تطبيقات لغة `(Go)` المستقلة**: كشف تلقائي عبر `go.mod` وتوجيه الملفات الثابتة `/static/*` مباشرة، مع تمرير الطلبات البرمجية عبر وسيط عكسي `(Reverse Proxy)` إلى المنفذ الداخلي (مثل `127.0.0.1:8080`) مع نطاق `https://*.test` موثوق.
   - **تطبيقات لغة رَست `(Rust)`**: كشف مشاريع `Cargo.toml` (مثل أطر `Actix-web` و `Axum` و `Rocket`) مع تشغيل تلقائي وموجه بروكسي إلى المنفذ `8080`.
   - **تطبيقات دوت نت `(.NET / C#)`**: كشف مشاريع `*.csproj` وتطبيقات `ASP.NET Core` مع تقديم الأصول الثابتة من مجلد `wwwroot/` وتوجيه الطلبات للمنفذ `5000`.
   - **تطبيقات لغة بايثون `(Python Web)`**: كشف أطر `Django` و `Flask` و `FastAPI` ومجلدات الأصول الثابتة وربطها بالبروكسي الداخلي على المنفذ `8000`.
   - **واجهات وتطبيقات نود `(Node.js)`**: دعم تلقائي لأطر `Vite` و `Next.js` و `Nuxt` و `React` و `Vue` و `Astro` وتقديم مخرجات البناء الجاهزة.
   - **تطبيقات روبي `(Ruby on Rails)`**: كشف ملفات `Gemfile` و `config.ru` مع تقديم الأصول من `public/` وتوجيه البروكسي للمنفذ `3000`.
   - **تطبيقات جافا وكوتلن `(Java / Kotlin)`**: كشف مشاريع `Maven` و `Gradle` (مثل `Spring Boot` و `Quarkus`) وتوجيهها للمنفذ `8080`.
   - **تطبيقات فلاتر ويب والمواقع الثابتة `(Flutter Web & Static)`**: كشف مجلدات `build/web` ومواقع `HTML/CSS/JS` وتقديمها مباشرة عبر خادم `Caddy` فائق السرعة مع دعم `HTTP/3` وضغط `gzip` و `zstd`.

2. **حل مشكلة شهادات `(SSL)` المحلية نهائياً**:
   - توليد شهادات أمان موثوقة بنسبة 100% لنطاقات `*.test` وحقن شهادة الجذر `(Root CA)` تلقائياً في متجر شهادات ويندوز ومتصفح فايرفوكس دون أي تحذيرات أمنية حمراء.

3. **خفيف على موارد النظام**:
   - ملف تنفيذي وحيد بحجم تقريبي `7.3MB` واستهلاك ذاكرة في وضع الخمول يتراوح بين `25-35MB` فقط (مقارنة بأكثر من `250MB` لتطبيقات Electron وأكثر من `2GB` لبيئات Docker).

4. **خدمات مدمجة ومحمولة**:
   - خوادم `MySQL` و `Redis` وأداة اختبار واعتراض البريد `Mailpit` مع واجهة ويب على المنفذ `8025`، تعمل جميعها في مجلدات محمولة دون تلويث لسجل النظام `(Registry)`.

5. **واجهة استخدام مكتبية متوافقة مع معايير `(BiDi)`**:
   - شريط تحكم علوي موحد ومصمم بعناية؛ حيث تقع أزرار التحكم بالنافذة (`✕`، `▢`، `—`) على اليسار الفعلي في الوضع العربي تماشياً مع المعايير الأصيلة، وعلى اليمين الفعلي في الوضع الإنجليزي، مع دعم الوضعين الداكن والفاتح.

---

## 📄 License & Commercial Rights

LaraGal is licensed under the **Business Source License 1.1 (BSL 1.1)**:
- **Free for all developers, students, and organizations** for internal development, local testing, and non-competing use.
- Commercial redistribution, white-labeling, or packaging as a competing cloud-hosted developer service requires a commercial agreement.
- See the full [LICENSE](LICENSE) file for details.
