<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d0017,40:1a0033,80:2d0055,100:3d006e&height=240&section=header&text=DARDCOR%20AGENT&fontSize=72&fontColor=d4b8ff&animation=fadeIn&fontAlignY=42&desc=The%20AI%20That%20Doesn%27t%20Just%20Talk%20%E2%80%94%20It%20Acts.&descAlignY=64&descSize=22&fontStyle=bold" width="100%"/>

<br/>
<img src="public/dardcor.png" width="120" height="auto" alt="Dardcor Logo" />
<br/>

[![NPM edition](https://img.shields.io/npm/v/dardcor-agent?color=%232d0055&label=npm&style=for-the-badge&logo=npm&logoColor=d4b8ff&labelColor=1a0033)](https://www.npmjs.com/package/dardcor-agent)
[![License: MIT](https://img.shields.io/badge/License-MIT-1a0033?style=for-the-badge&labelColor=0d0017&color=2d0055)](LICENSE)
[![Node.js](https://img.shields.io/badge/Node.js-Required-2d0055?style=for-the-badge&logo=node.js&logoColor=d4b8ff&labelColor=1a0033)](https://nodejs.org)
[![Go Backend](https://img.shields.io/badge/Backend-Go-3d006e?style=for-the-badge&logo=go&logoColor=d4b8ff&labelColor=1a0033)](https://golang.org)
[![TypeScript UI](https://img.shields.io/badge/UI-TypeScript-2d0055?style=for-the-badge&logo=typescript&logoColor=d4b8ff&labelColor=1a0033)](https://typescriptlang.org)

<br/>

<img src="https://readme-typing-svg.herokuapp.com?font=JetBrains+Mono&weight=700&size=18&duration=3000&pause=1000&color=9B59FF&center=true&vCenter=true&width=600&lines=%22Minimal+Input.+Maximum+Execution.%22;Autonomous.+Private.+Unstoppable.;Your+terminal+just+got+a+brain+upgrade." alt="Typing SVG" />

<br/>

**[🚀 Quick Start](#-quick-start)** · **[⚡ Features](#-features)** · **[📖 Usage](#-usage)** · **[🧠 Architecture](#-system-architecture)** · **[🗺️ Roadmap](#️-roadmap)**

</div>

---

## 🔥 Why Dardcor Agent?

Most AI assistants are glorified chatbots — they advise, suggest, and then leave you to do all the actual work.

**Dardcor Agent is different. It executes.**

It doesn't wait for your next prompt. It analyzes your system, makes decisions, runs code, detects errors, fixes itself — and keeps going. This is what true AI autonomy looks like: a relentless local agent that turns your intent into results with zero hand-holding.

Whether you're a solo developer, a power user, or someone who's sick of switching between 10 tools — Dardcor Agent collapses your workflow into a single, razor-sharp command.

<br/>

| Capability | Ordinary AI Chat | **Dardcor Agent** |
|:---|:---:|:---:|
| Answer questions | ✅ | ✅ |
| Autonomous code execution | ❌ | ✅ |
| Self-healing on errors | ❌ | ✅ |
| 100% local & private | ❌ | ✅ |
| Web UI + CLI in one tool | ❌ | ✅ |
| Ultra token-efficient output | ❌ | ✅ |
| Zero external data exposure | ❌ | ✅ |

---

## ✨ Features

### 🔥 Ultra Token Efficient
Every token counts. Dardcor Agent's prompt architecture is engineered to extract maximum logical output with minimum token usage — leaner responses, faster results, lower cost. No fluff, no filler, just pure execution.

### 🤖 Fully Autonomous Execution
This isn't an assistant that drafts code for you to copy-paste. Dardcor Agent **analyzes, decides, and executes** operations directly on your local system — end-to-end, without babysitting.

### 🛠️ Self-Healing System (Doctor Mode)
Broken config? Corrupted dependencies? Just run `dardcor doctor` and watch the agent diagnose, repair, and restore your system to peak condition — automatically, no manual debugging required.

### 🔒 100% Local & Private
Every operation stays within `127.0.0.1:25000`. Zero telemetry. Zero external calls. Your code, your secrets, your data — **never leave your machine**. Privacy isn't a feature here, it's the architecture.

### 🌐 Modern Web UI — Real-Time Dashboard
A sleek, interactive browser dashboard with live session history, real-time feedback, and a built-in cache-busting system. Beautiful and functional — no Electron bloat, just open `http://127.0.0.1:25000`.

### ⚡ Lightweight CLI — Terminal Native
A blazing-fast terminal interface with zero overhead. Perfect for automation pipelines, scripting, and developer workflows that demand speed over spectacle.

---

## 🚀 Quick Start

> **Prerequisite:** [Node.js](https://nodejs.org) must be installed. Use a terminal with admin access if required.

### Install via NPM (Recommended)

```bash
npm install -g dardcor-agent
```

Once installed, the `dardcor` command is immediately available globally. A local JSON-based database is initialized automatically on first run — no setup, no configuration wizard, no friction.

### Install from Source

```bash
git clone https://github.com/Dardcor/Dardcor-Agent.git
cd Dardcor-Agent
```

### Update

```bash
# Via NPM
npm install -g dardcor-agent@latest
```

### Uninstall

```bash
npm uninstall -g dardcor-agent
```

---

## 📖 Usage

### Show All Commands

```bash
dardcor help
```

---

### Launch Web UI

```bash
dardcor run
```

Open your browser and navigate to:

```
http://127.0.0.1:25000
```

**What you get in Web UI:**
- Real-time session dashboard
- Interactive conversation history
- Built-in cache-busting system
- Clean, responsive interface — no install, no app, just a browser tab

---

### 💻 Launch CLI Mode

```bash
dardcor cli
```

**Built for:**
- Rapid coding sessions directly in terminal
- Automation, CI pipelines, and scripting
- Lightweight workflows where UI overhead is unwelcome

---

### 🛠️ Run System Doctor

```bash
dardcor doctor
```

Trigger this whenever something breaks. The agent will:

1. **Scan** — detect broken configurations and failed dependencies
2. **Diagnose** — identify the root cause, not just symptoms
3. **Repair** — automatically restore the system to its optimal state

No error Googling. No Stack Overflow rabbit holes. Just one command.

---

## 🧠 System Architecture

```
Dardcor Agent
├── 🌐 Web Server (Go backend)      → 127.0.0.1:25000
├── 💻 CLI Interface (Node.js)      → Direct terminal access
├── 🗄️  Local Database              → /database (JSON file-based)
├── ⚙️  Config Store                → ~/.dardcor/config.json
└── 🔒 Network Scope               → Localhost only — zero external exposure
```

**Tech Stack:**

| Layer | Technology | Why It's Here |
|:---|:---|:---|
| Backend | **Go** | Native concurrency, high throughput, tiny binary |
| Frontend | **TypeScript + Vite** | Type-safe, fast-build, modern DX |
| CLI | **Node.js** | Cross-platform, lightweight, instant startup |
| Storage | **JSON File-based** | No database server required — just works |

The architecture is intentionally simple. No microservices. No Docker requirement. No ops overhead. A single agent that runs anywhere Node.js runs.

---

## 🗺️ Roadmap

- [x] Web UI with real-time dashboard
- [x] CLI mode
- [x] Self-healing system (`dardcor doctor`)
- [x] Local-only network (privacy-first by default)
- [x] Multi-agent collaboration system
- [x] Plugin ecosystem & extension API
- [x] Cloud sync (optional, privacy-first)
- [x] GUI config editor (no more JSON editing)

---

## 📄 License

Licensed under the [MIT License](LICENSE) — free to use, modify, and distribute.

---

<div align="center">

<br/>

**Built with 💜 by [Dardcor](https://github.com/Dardcor)**

*If this project saves you time, a ⭐ means everything.*

<br/>

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:3d006e,50:1a0033,100:0d0017&height=140&section=footer" width="100%"/>

</div>


