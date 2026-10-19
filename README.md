# LaraGal 🚀

<div align="center">

![LaraGal Banner](https://raw.githubusercontent.com/MosaabGalmod/laragal/main/docs/assets/banner.png)

**The ultra-fast, lightweight, zero-config trusted SSL local development environment for PHP, Laravel & WordPress.**

[![License: BSL 1.1](https://img.shields.io/badge/License-BSL_1.1-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Frontend-Vue_3_+_Tailwind-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows_(macOS_%26_Linux_Ready)-0078D6?logo=windows&logoColor=white)](#)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[English](README.md) • [العربية](#-نظرة-عامة-بالعربية) • [Documentation](docs/) • [Architecture Spec](docs/superpowers/specs/2026-10-19-laragal-design.md)

</div>

---

## ⚡ Overview

**LaraGal** is a modern, blazing-fast local server and development suite built from the ground up with **Go** and **Wails v2 (Vue 3 + Tailwind CSS)**. Designed as a superior, open alternative to legacy tools like *Laragon*, *XAMPP*, and commercial freemium suites like *Laravel Herd on Windows*, LaraGal solves the most frustrating pain points of Windows local development:

1. **Zero-Warning Local SSL (`https://*.test`):** Embedded automated internal Certificate Authority (CA) powered by Caddy that installs into the Windows Trust Store and Firefox NSS databases with zero browser security warnings.
2. **True Multi-PHP Isolation:** Runs multiple PHP FastCGI versions simultaneously (`8.1`, `8.2`, `8.3`, `8.4`) on isolated local loopback ports (`:9081..:9084`).
3. **Ultra-Low Memory Footprint:** Built on native Go and Windows WebView2—uses only **25–35 MB RAM** at idle (compared to 250MB+ for Electron apps and 2–4GB for Docker/WSL2).
4. **Isolated Portable Services:** Bundled MySQL/MariaDB, Redis, and Mailpit (local email catcher) running cleanly in portable directories with no Windows registry bloat.
5. **1-Click Project Scaffolding:** Instant setup for Laravel & WordPress with automatic `.env` synchronization (DB credentials, Mailpit SMTP, and trusted SSL URL).
6. **Bilingual Arabic/English UI:** Native BiDi RTL support and Dark/Light themes built to high UX standards.

---

## 📊 Comparison Matrix

| Feature | LaraGal | Laragon | Laravel Herd (Win) | XAMPP | Docker / Sail |
|---|---|---|---|---|---|
| **Idle Memory (RAM)** | **~30 MB** | ~50 MB | ~150 MB | ~140 MB | **2000–4000 MB+** |
| **Local SSL (`*.test`)** | **100% Trusted & Automated** | Unreliable / Red Warnings | Automated | Painful / Manual | Complex / Manual |
| **Concurrent Multi-PHP** | **Yes (Per-site simultaneous)** | No (Global switch only) | **Paid (Herd Pro)** | No | Yes (Container overhead) |
| **Binary Size** | **~7.3 MB** | ~180 MB | ~120 MB | ~160 MB | Multi-Gigabyte |
| **Managed DB & Mailpit** | **Included Free** | Partial (HeidiSQL) | **Paid (Herd Pro)** | phpMyAdmin only | Manual Compose |
| **Language & BiDi** | **Arabic & English (RTL/LTR)** | English only | English only | English only | N/A |
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
|  |  - Dynamic Caddyfile Generator & FastCGI Proxy                     | |
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

## 🚀 Quickstart & Building From Source

### Prerequisites
- Windows 10/11 (64-bit)
- [Go 1.22+](https://go.dev/)
- [Node.js 18+](https://nodejs.org/) & npm

### Build Instructions

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

3. **Run the Go test suite:**
   ```bash
   go test ./...
   ```

4. **Compile the standalone executable:**
   ```bash
   go build -o laragal.exe .
   ```

5. **Run LaraGal:**
   ```powershell
   .\laragal.exe
   ```

---

## 🇸🇦 نظرة عامة بالعربية

برنامج **`LaraGal`** هو بيئة خادم محلي خفيفة وعصرية لمطوري `PHP` وإطار عمل `Laravel` على نظام ويندوز (مع جاهزية معمارية لأنظمة ماك ولينكس).

### أبرز المميزات:
* **حل مشكلة شهادات `SSL` المحلية تلقائياً**: توليد شهادات موثوقة بالكامل لنطاقات `*.test` وحقن شهادة الجذر في متجر ويندوز دون أي شاشات حمراء أو تحذيرات أمنية في المتصفحات.
* **عزل وتعدد إصدارات `PHP`**: تشغيل `PHP 8.1` و `8.2` و `8.3` و `8.4` في نفس الوقت على منافذ منفصلة.
* **خفيف وسريع جداً**: ملف تنفيذي واحد بحجم `7.3MB` فقط، واستهلاك ذاكرة خمول أقل من `35MB` مقارنة بـ `2GB+` لـ `Docker`.
* **خدمات مدمجة مجانية**: خوادم `MySQL` و `Redis` وأداة صيد البريد `Mailpit` مع واجهة ويب على المنفذ `8025`.
* **معالج المشاريع وضبط `.env` التلقائي**: إنشاء مشاريع `Laravel` و `WordPress` بنقرة واحدة وضبط الاتصال بقاعدة البيانات والرابط المشفر تلقائياً.
* **واجهة عربية أصيلة**: دعم كامل لاتجاه اليمين لليسار `(RTL)` وقواميس معزولة والوضعين الداكن والفاتح.

---

## 📄 License & Commercial Rights

LaraGal is licensed under the **Business Source License 1.1 (BSL 1.1)**:
- **Free for all developers, students, and organizations** for internal development, testing, and non-competing use.
- Commercial redistribution, resale, or packaging as a competing hosted developer service requires an explicit commercial agreement from the licensor.
- See the full [LICENSE](LICENSE) terms.
