# Dardcor Agent

> **Autonomous AI System Controller**  
> Premium Control Dashboard & AI Agent for Any Platform.

---

## Quick Start

### Installation

```bash
git clone https://github.com/dardcor/dardcor-agent.git
cd dardcor-agent
npm i -g dardcor-agent
```

### Commands

```bash
dardcor help              Show all commands
dardcor doctor            Health check & auto-repair
dardcor run               Start Gateway + WebUI Dashboard
dardcor cli               Interactive terminal coding agent
```

---

## Modes

### Dashboard Mode (`dardcor run`)
Full system with graphical interface in your browser.
- Starts the backend, auto-builds UI, and opens the Dashboard at `http://127.0.0.1:25000`.
- Features: Real-time System Monitor, File Explorer, Terminal, Chat AI.

### CLI Mode (`dardcor cli`)
Interactive terminal coding agent — headless, no browser needed.
- Full autonomous agent with BUILD and PLAN modes.
- Ultrawork mode for complex autonomous tasks.
- Multi-provider LLM support (OpenAI, Anthropic, Gemini, DeepSeek, Ollama).

### Doctor (`dardcor doctor`)
Health check and auto-repair system.
- Verifies configuration, backend binary, Node.js, Go runtime.
- Automatically fixes missing config, directories, and defaults.

---

## Multi-Provider AI

Supports multiple AI providers out of the box:

| Provider | Models | Auth |
|----------|--------|------|
| OpenAI | GPT-4o, GPT-4.1, o3 | API Key |
| Anthropic | Claude Opus, Sonnet, Haiku | API Key |
| Google Gemini | Gemini 2.5 Pro/Flash | API Key |
| DeepSeek | DeepSeek-V3/R1 | API Key |
| OpenRouter | Multi-provider router | API Key |
| Ollama | Local models | Free |
| Local | Built-in Dardcor Agent | No key |

---

## System Requirements

- **Node.js** >= 18
- **Go** (latest recommended)
- **Windows / Linux / macOS**

---

## Design

- Deep Purple & Electric Violet interface
- Modern typography (Outfit) & JetBrains Mono for Terminal
- Real-time System Monitor, File Explorer, Reactive Terminal, Chat AI

---

**Dardcor Team** | *Powering Autonomous Control.*
