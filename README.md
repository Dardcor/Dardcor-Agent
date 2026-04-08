# Dardcor Agent

Dardcor Agent is a superior, autonomous AI programming assistant designed for token efficiency and high-level, completely autonomous system execution.

It operates seamlessly as a Gateway Server and an Interactive Terminal Command Line Interface.

## Features
- **Highly Token Efficient:** Designed to utilize minimal tokens while producing maximal execution logic.
- **Autonomous Execution:** Capable of resolving and running logical operations locally on your system.
- **Unified Interface:** Works both as a beautiful Web UI and a headless CLI.
- **Auto-Healing System:** Built-in system doctor to repair broken configurations automatically.

## Installation

You can install Dardcor Agent via Manual Git Clone:

```bash
git clone https://github.com/dardcor/dardcor-agent.git
cd dardcor-agent
npm install -g dardcor-agent
```

*Note: The `npm install -g dardcor-agent` command will automatically set up the system globally and initialize your internal local JSON database.*

To uninstall the agent at any point:
```bash
npm uninstall -g dardcor-agent
```

## Usage

Dardcor Agent has a streamlined set of commands to prevent collision and ensure perfect execution.

### View all commands:
```bash
dardcor help
```

### Self-Repair System
If the system ever has trouble starting, the automatic doctor will fix configurations and dependencies:
```bash
dardcor doctor
```

### Launch Web Server (UI Mode)
To launch the primary autonomous system with a realtime, cache-busting Web UI:
```bash
dardcor run
```
*Note: This command will display the Dardcor UI banner and start serving the dashboard on `http://127.0.0.1:25000` exclusively.*

### Launch Terminal Agent (CLI Mode)
To launch the interactive system directly in your terminal for highly efficient coding flows:
```bash
dardcor cli
```
*Note: This command will display the Dardcor CLI banner and drop you immediately into the interactive mode.*

---

## Architecture
- All configurations are securely stored inside `~/.dardcor/config.json`.
- The database is completely local and stored within the `database` folder.
- Network activity is rigorously restricted absolutely to `127.0.0.1:25000` to prevent any port collisions.
