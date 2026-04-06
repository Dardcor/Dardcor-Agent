# 🤖 Dardcor Agent: AI-Powered Autonomous System Controller

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=flat&logo=react&logoColor=61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?style=flat&logo=typescript&logoColor=white)

**Dardcor Agent** is a sophisticated, AI-driven autonomous agent designed to control and monitor your local system through a beautiful, modern web interface. Built with high-performance **Go** and a responsive **React/TypeScript** frontend, it bridges the gap between AI intelligence and direct operating system control.

---

## ✨ Key Features

### 🧠 Intelligent System Control
*   **Autonomous OS Interaction**: Ask the AI to manage files, run commands, or monitor system health.
*   **Contextual Memory**: Keeps track of conversation history for better task execution.

### 📁 Advanced File Management
*   **Full CRUD Operations**: Create, Read, Update, and Delete files/directories.
*   **Safe File Manipulation**: Search, move, and copy files across your drives with built-in safety checks.
*   **Drive Mapping**: View all available drives and their status.

### 💻 Remote Command Execution
*   **Live Shell Access**: Execute any shell command directly from the web interface.
*   **Real-time Output**: Integrated WebSocket Support for streaming terminal logs and command outputs.
*   **Shell History**: Keeps track of previous commands for quick access.

### 📊 System Monitoring & Analytics
*   **Real-time Resource Usage**: Monitor CPU and Memory usage dynamically.
*   **Process Management**: View all running processes and terminate unresponsive ones if necessary.

---

## 🛠️ Technology Stack

| Role | Technology |
| :--- | :--- |
| **Backend Core** | Go (Golang) |
| **API Framework** | Gorilla Mux |
| **Frontend UI** | React 19 + TypeScript |
| **Build Tool** | Vite 8 |
| **Styling** | Modern Vanilla CSS (Nocturnal Purple Theme) |
| **Real-time** | WebSockets (gorilla/websocket) |

---

## 🚀 Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) (version 1.21+)
- [Node.js](https://nodejs.org/) (version 20+)
- Windows OS (Optimized for Windows)

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/dardcor-agent.git
   cd dardcor-agent
   ```
2. Install frontend dependencies:
   ```bash
   npm install
   ```

### 🌐 Public Installation & Usage

Dardcor Agent can be installed globally or run instantly using **NPX**, just like other professional system agents.

#### ⚡ Method 1: Instant Run (Recommended)
You don't need to download the source code manually. Simply run this from any terminal:
```cmd
npx dardcor agent
```

#### 🛠️ Method 2: Global Installation
Install the agent permanently on your machine:
```cmd
npm install -g dardcor-agent
```
Then use the global command from any directory:
```cmd
dardcor agent
```

#### 📦 Method 3: Local Installation
If you have the source code folder:
1.  Install dependencies: `npm install`
2.  Link locally: `npm link`
3.  Run from anywhere: `dardcor agent`

---

## 🎮 Available Commands

| Command | Action |
| :--- | :--- |
| **`dardcor agent`** | Launch the Autonomous Agent & Dashboard (Port 25000) |
| **`dardcor build`** | Compile the Go Core and React UI |
| **`dardcor install`** | Setup project assets |

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
**Dardcor Agent** - *Turning AI into Action.*
