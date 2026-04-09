<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=200&section=header&text=DARDCOR%20AGENT&fontSize=60&fontColor=fff&animation=fadeIn&fontAlignY=38&desc=Autonomous%20AI%20Programming%20Assistant&descAlignY=60&descSize=18" width="100%"/>

<br/>

[![NPM Version](https://img.shields.io/npm/v/dardcor-agent?color=%23FF6B6B&label=npm&style=for-the-badge)](https://www.npmjs.com/package/dardcor-agent)
[![License: MIT](https://img.shields.io/badge/License-MIT-4ECDC4?style=for-the-badge)](LICENSE)
[![Node.js](https://img.shields.io/badge/Node.js-Required-FFE66D?style=for-the-badge&logo=node.js)](https://nodejs.org)
[![Go](https://img.shields.io/badge/Backend-Go-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![TypeScript](https://img.shields.io/badge/UI-TypeScript-3178C6?style=for-the-badge&logo=typescript)](https://typescriptlang.org)

<br/>

> **"Minimal input. Maximum execution."**

**Dardcor Agent** is a fully autonomous AI programming assistant that combines the intelligence of the latest AI models with a fast, secure, and efficient local execution system — right from your terminal or browser.

<br/>

[🚀 Quick Start](#-quick-start) · [⚡ Features](#-features) · [📖 Usage](#-usage) · [🗺️ Roadmap](#️-roadmap)

</div>

---

## 🎯 Why Dardcor Agent?

Most AI assistants can only *talk*. **Dardcor Agent takes action.**

| Capability | Regular AI Chat | Dardcor Agent |
|-----------|:---:|:---:|
| Answer questions | ✅ | ✅ |
| Autonomous code execution | ❌ | ✅ |
| Self-healing on errors | ❌ | ✅ |
| Runs 100% locally & privately | ❌ | ✅ |
| Web UI + CLI interface | ❌ | ✅ |
| Token-efficient output | ❌ | ✅ |

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔥 Ultra Token Efficient
Optimized prompt architecture delivers maximum logical output with minimal token usage — leaner, faster, sharper.

</td>
<td width="50%">

### 🤖 Fully Autonomous Execution
Not just suggestions — the agent **analyzes, decides, and executes** operations independently on your local system.

</td>
</tr>
<tr>
<td width="50%">

### 🛠️ Self-Healing System
Run `dardcor doctor` and let the agent detect broken configs, repair dependencies, and restore the system to optimal condition automatically.

</td>
<td width="50%">

### 🔒 100% Local & Private
All activity is restricted to `127.0.0.1:25000`. No data leaves your machine — your code stays yours.

</td>
</tr>
<tr>
<td width="50%">

### 🌐 Modern Web UI
Real-time dashboard with a built-in cache-busting system. Visual, interactive, and responsive — accessible from your browser anytime.

</td>
<td width="50%">

### ⚡ Lightweight CLI
A fast and lean terminal mode. Perfect for automation, scripting, and developer workflows that don't need a UI.

</td>
</tr>
</table>

---

## 🚀 Quick Start

### Install via NPM (Global)

```bash
npm install -g dardcor-agent
```

Once installed, the `dardcor` command is immediately available in your terminal — a local JSON-based database is initialized automatically.

### Or Clone Manually

```bash
git clone https://github.com/Dardcor/Dardcor-Agent.git
cd Dardcor-Agent
```

### Uninstall

```bash
npm uninstall -g dardcor-agent
```

> **Prerequisites:** Make sure [Node.js](https://nodejs.org) is installed. Use a terminal with admin access if required.

---

## 📖 Usage

### Show All Commands

```bash
dardcor help
```

---

### 🌐 Web UI Mode

```bash
dardcor run
```

Open your browser and go to:

```
http://127.0.0.1:25000
```

Available features in Web UI:
- Real-time dashboard
- Interactive session history
- Cache-busting system

---

### 💻 CLI Mode

```bash
dardcor cli
```

Best suited for:
- Fast coding directly from the terminal
- Automation & scripting
- Lightweight workflows without UI overhead

---

### 🛠️ Self-Repair (System Doctor)

```bash
dardcor doctor
```

Run this when something goes wrong. The agent will:
1. Detect broken configuration
2. Repair problematic dependencies
3. Restore the system to optimal condition

---

## 🧠 System Architecture

```
Dardcor Agent
├── 🌐 Web Server (Go backend)     → Port 127.0.0.1:25000
├── 💻 CLI Interface (Node.js)     → Direct terminal access
├── 🗄️ Local Database              → /database (JSON-based)
├── ⚙️  Config                     → ~/.dardcor/config.json
└── 🔒 Network Scope               → Localhost only (zero external exposure)
```

**Tech stack:**
- **Backend:** Go — high performance, native concurrency
- **Frontend:** TypeScript + Vite — modern and type-safe UI
- **CLI:** Node.js — cross-platform and lightweight
- **Database:** File-based JSON — no external database setup required

---

## 🗺️ Roadmap

- [x] Web UI (real-time dashboard)
- [x] CLI mode
- [x] Self-healing system (doctor)
- [x] Local-only network (privacy-first)
- [ ] Multi-agent collaboration system
- [ ] Plugin ecosystem
- [ ] Cloud sync (optional, privacy-first)
- [ ] Config GUI (no manual JSON editing)

---

## 📄 License

This project is licensed under the [MIT License](LICENSE) — free to use, modify, and distribute.

---

<div align="center">

**Built with ❤️ by [Dardcor](https://github.com/Dardcor)**

If this project helps you, drop a ⭐ — it means a lot!

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=100&section=footer" width="100%"/>

</div>
