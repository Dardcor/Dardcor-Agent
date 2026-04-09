<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:1a0033,50:3d0066,100:6600cc&height=220&section=header&text=DARDCOR%20AGENT&fontSize=64&fontColor=e8d5ff&animation=fadeIn&fontAlignY=40&desc=Autonomous%20AI%20Programming%20Assistant&descAlignY=62&descSize=20" width="100%"/>

<br/>

[![NPM Version](https://img.shields.io/npm/v/dardcor-agent?color=%236600cc&label=npm&style=for-the-badge&logo=npm&logoColor=white)](https://www.npmjs.com/package/dardcor-agent)
[![License: MIT](https://img.shields.io/badge/License-MIT-4b0082?style=for-the-badge)](LICENSE)
[![Node.js](https://img.shields.io/badge/Node.js-Required-7b2fff?style=for-the-badge&logo=node.js&logoColor=white)](https://nodejs.org)
[![Go](https://img.shields.io/badge/Backend-Go-5c00b8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![TypeScript](https://img.shields.io/badge/UI-TypeScript-8a2be2?style=for-the-badge&logo=typescript&logoColor=white)](https://typescriptlang.org)

<br/>

> **"Minimal input. Maximum execution."**

**Dardcor Agent** adalah asisten pemrograman AI yang sepenuhnya otonom — menggabungkan kecerdasan model AI terkini dengan sistem eksekusi lokal yang cepat, aman, dan efisien, langsung dari terminal atau browser Anda.

<br/>

[🚀 Quick Start](#-quick-start) · [⚡ Features](#-features) · [📖 Usage](#-usage) · [🗺️ Roadmap](#️-roadmap)

</div>

---

## 🎯 Mengapa Dardcor Agent?

Kebanyakan asisten AI hanya bisa *berbicara*. **Dardcor Agent mengambil tindakan nyata.**

| Kemampuan | AI Chat Biasa | Dardcor Agent |
|---|:---:|:---:|
| Menjawab pertanyaan | ✅ | ✅ |
| Eksekusi kode otonom | ❌ | ✅ |
| Self-healing saat error | ❌ | ✅ |
| Berjalan 100% lokal & privat | ❌ | ✅ |
| Antarmuka Web UI + CLI | ❌ | ✅ |
| Output hemat token | ❌ | ✅ |

---

## ✨ Features

### 🔥 Ultra Token Efficient
Arsitektur prompt yang dioptimalkan menghasilkan output logis maksimum dengan penggunaan token minimal — lebih ramping, lebih cepat, lebih tajam.

### 🤖 Fully Autonomous Execution
Bukan sekadar saran — agen ini **menganalisis, memutuskan, dan mengeksekusi** operasi secara mandiri di sistem lokal Anda.

### 🛠️ Self-Healing System
Jalankan `dardcor doctor` dan biarkan agen mendeteksi konfigurasi yang rusak, memperbaiki dependensi, dan memulihkan sistem ke kondisi optimal secara otomatis.

### 🔒 100% Local & Private
Semua aktivitas dibatasi pada `127.0.0.1:25000`. Tidak ada data yang keluar dari mesin Anda — kode Anda tetap milik Anda.

### 🌐 Modern Web UI
Dashboard real-time dengan sistem cache-busting bawaan. Visual, interaktif, dan responsif — dapat diakses dari browser kapan saja.

### ⚡ Lightweight CLI
Mode terminal yang cepat dan ringan. Sempurna untuk otomasi, scripting, dan alur kerja developer yang tidak memerlukan UI.

---

## 🚀 Quick Start

> **Prasyarat:** Pastikan [Node.js](https://nodejs.org) sudah terinstal. Gunakan terminal dengan akses admin jika diperlukan.

### Install via NPM (Global)

```bash
npm install -g dardcor-agent
```

Setelah terinstal, perintah `dardcor` langsung tersedia di terminal Anda — database berbasis JSON lokal diinisialisasi secara otomatis.

### Clone Manual

```bash
git clone https://github.com/Dardcor/Dardcor-Agent.git
cd Dardcor-Agent
```

### Uninstall

```bash
npm uninstall -g dardcor-agent
```

### Update (via NPM)

```bash
npm install -g dardcor-agent@latest
```

### Update (via GitHub)

```bash
npm install -g github:Dardcor/Dardcor-Agent
```

---

## 📖 Usage

### Tampilkan Semua Perintah

```bash
dardcor help
```

---

### 🌐 Web UI Mode

```bash
dardcor run
```

Buka browser dan akses:

```
http://127.0.0.1:25000
```

Fitur yang tersedia di Web UI:
- Real-time dashboard
- Riwayat sesi interaktif
- Sistem cache-busting

---

### 💻 CLI Mode

```bash
dardcor cli
```

Cocok untuk:
- Coding cepat langsung dari terminal
- Otomasi & scripting
- Alur kerja ringan tanpa overhead UI

---

### 🛠️ Self-Repair (System Doctor)

```bash
dardcor doctor
```

Jalankan ini saat terjadi masalah. Agen akan:
1. Mendeteksi konfigurasi yang rusak
2. Memperbaiki dependensi yang bermasalah
3. Memulihkan sistem ke kondisi optimal

---

## 🧠 Arsitektur Sistem

```
Dardcor Agent
├── 🌐 Web Server (Go backend)     → Port 127.0.0.1:25000
├── 💻 CLI Interface (Node.js)     → Akses terminal langsung
├── 🗄️ Local Database              → /database (berbasis JSON)
├── ⚙️  Config                     → ~/.dardcor/config.json
└── 🔒 Network Scope               → Localhost only (zero external exposure)
```

**Tech Stack:**

| Layer | Teknologi | Keterangan |
|---|---|---|
| Backend | Go | Performa tinggi, konkurensi native |
| Frontend | TypeScript + Vite | UI modern dan type-safe |
| CLI | Node.js | Cross-platform dan ringan |
| Database | JSON File-based | Tanpa setup database eksternal |

---

## 🗺️ Roadmap

- [x] Web UI (real-time dashboard)
- [x] CLI mode
- [x] Self-healing system (doctor)
- [x] Local-only network (privacy-first)
- [ ] Sistem kolaborasi multi-agent
- [ ] Ekosistem plugin
- [ ] Cloud sync (opsional, privacy-first)
- [ ] Config GUI (tanpa edit JSON manual)

---

## 📄 Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE) — bebas digunakan, dimodifikasi, dan didistribusikan.

---

<div align="center">

**Built with 💜 by [Dardcor](https://github.com/Dardcor)**

Jika proyek ini membantu Anda, berikan ⭐ — itu sangat berarti!

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:6600cc,50:3d0066,100:1a0033&height=120&section=footer" width="100%"/>

</div>