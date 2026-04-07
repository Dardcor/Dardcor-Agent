#!/usr/bin/env node

import { spawn, execSync } from 'child_process';
import { fileURLToPath } from 'url';
import path from 'path';
import fs from 'fs';
import os from 'os';
import readline from 'readline';
import http from 'http';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const args = process.argv.slice(2);
const command = args[0];
const subArgs = args.slice(1);

const CONFIG_DIR = path.join(os.homedir(), '.dardcor');
const CONFIG_FILE = path.join(CONFIG_DIR, 'config.json');
const DEFAULT_PORT = '25000';
const GATEWAY_URL = `http://127.0.0.1:${DEFAULT_PORT}`;

function ensureConfigDir() {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true });
  }
}

function loadConfig() {
  ensureConfigDir();
  if (fs.existsSync(CONFIG_FILE)) {
    try {
      return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8'));
    } catch {
      return {};
    }
  }
  return {};
}

function saveConfig(cfg) {
  ensureConfigDir();
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(cfg, null, 2), 'utf8');
}

const C = {
  reset:   '\x1b[0m',
  bold:    '\x1b[1m',
  dim:     '\x1b[2m',
  red:     '\x1b[31m',
  green:   '\x1b[32m',
  yellow:  '\x1b[33m',
  blue:    '\x1b[34m',
  magenta: '\x1b[35m',
  cyan:    '\x1b[36m',
  white:   '\x1b[37m',
  bgMagenta: '\x1b[45m',
  bgBlue:    '\x1b[44m',
};

const ok  = (msg) => console.log(`${C.green}[✓]${C.reset} ${msg}`);
const err = (msg) => console.log(`${C.red}[✗]${C.reset} ${msg}`);
const inf = (msg) => console.log(`${C.cyan}[i]${C.reset} ${msg}`);
const wrn = (msg) => console.log(`${C.yellow}[!]${C.reset} ${msg}`);
const hdr = (msg) => console.log(`\n${C.magenta}${C.bold}${msg}${C.reset}`);

function printBanner(subtitle = '') {
  console.log(`
${C.magenta}${C.bold}
  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗
  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗
  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝
  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗
  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║
  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝
${C.reset}${C.dim}  Your Personal AI Agent — Any OS, Any Platform.${C.reset}${subtitle ? `\n  ${C.cyan}${subtitle}${C.reset}` : ''}
`);
}

function printHelp() {
  printBanner();
  console.log(`${C.bold}USAGE:${C.reset}

  ${C.cyan}dardcor help${C.reset}              Show this help message
  ${C.cyan}dardcor doctor${C.reset}            Health check & auto-repair system
  ${C.cyan}dardcor run${C.reset}               Start Gateway + WebUI (Live Updates)
  ${C.cyan}dardcor cli${C.reset}               Interactive terminal coding agent

${C.bold}INSTALLATION:${C.reset}

  git clone https://github.com/dardcor/dardcor-agent.git
  cd dardcor-agent
  npm i -g dardcor-agent

${C.bold}EXAMPLES:${C.reset}

  dardcor run                    Launch Dashboard with live preview
  dardcor cli                    Headless coding agent
  dardcor doctor                 Fix system issues

${C.bold}DOCS:${C.reset}  https://github.com/dardcor/dardcor-agent
`);
}

if (!command || command === 'help' || command === '--help' || command === '-h') {
  printHelp();
  process.exit(0);
}

switch (command) {
  case 'run':     runGateway(); break;
  case 'cli':     runCLI(); break;
  case 'doctor':  runDoctor(); break;
  default:
    err(`Unknown command: ${C.yellow}${command}${C.reset}`);
    printHelp();
    process.exit(1);
}

async function runGateway() {
  const cfg = loadConfig();
  const isFirstRun = !cfg.initialized;

  printBanner('Dardcor Gateway (Unified Mode)');

  if (isFirstRun) {
    wrn('Initial setup required. Guided configuration starting...');
    await runOnboardFlow();
  }

  // Pre-flight check: ensure dependencies
  if (!fs.existsSync(path.join(__dirname, 'node_modules'))) {
    inf('Fresh install detected. Running npm install...');
    execSync('npm install', { stdio: 'inherit' });
  }

  killOldInstance();

  const port = cfg.port || DEFAULT_PORT;
  const gatewayUrl = `http://127.0.0.1:${port}`;
  const devUrl = `http://127.0.0.1:25099`;

  process.env.DARDCOR_DEV_URL = devUrl;
  process.env.PORT = port;

  hdr('🚀 UNIFIED RUNTIME ACTIVE');
  console.log(`  ${C.blue}•${C.reset} Access UI:   ${C.green}${C.bold}${gatewayUrl}${C.reset}`);
  console.log(`  ${C.blue}•${C.reset} HMR Engine:  ${C.cyan}Vite Dev Server (Port 5173)${C.reset}`);
  console.log(`  ${C.blue}•${C.reset} Logic Port:  ${C.magenta}Go Backend (Port ${port})${C.reset}`);
  console.log(`  ${C.dim}${'─'.repeat(50)}${C.reset}\n`);

  inf('Igniting real-time preview engines...');
  
  // Start Vite (HMR)
  spawn('npx', ['vite', '--host', '127.0.0.1', '--no-open'], {
    cwd: __dirname,
    stdio: 'inherit',
    shell: true,
    env: process.env
  });

  // Start Go Backend (Proxy + API)
  startBackend('run');
}

async function runCLI() {
  const cfg = loadConfig();

  printBanner('Interactive CLI Mode');

  if (!cfg.provider) {
    inf('No provider configured. Starting setup wizard...');
    await runOnboardFlow(true);
  }

  console.log(`${C.cyan}[*]${C.reset} Starting Dardcor CLI Agent...`);
  console.log(`${C.cyan}[i]${C.reset} API: ${C.cyan}${GATEWAY_URL}${C.reset}`);
  console.log(`${C.dim}    Provider: ${cfg.provider || 'local'} | Model: ${cfg.model || 'auto'}${C.reset}`);
  console.log(`${C.yellow}[!]${C.reset} Press ${C.bold}Ctrl+C${C.reset} to stop.\n`);

  startBackend('cli');

  await sleep(2000);
  await startInteractiveTUI();
}

async function startInteractiveTUI() {
  const cfg = loadConfig();

  const serverUp = await waitForServer(GATEWAY_URL, 10);
  if (!serverUp) {
    err('Gateway not responding. Try restarting.');
    process.exit(1);
  }

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: true,
  });

  let conversationId = null;
  let agentMode = 'build';
  let sessionHistory = [];
  let isProcessing = false;

  function renderHeader() {
    console.clear();
    const modeLabel = agentMode === 'build' ? `${C.green}BUILD${C.reset}` : `${C.blue}PLAN${C.reset}`;
    const provider = cfg.provider || 'local';
    const model = cfg.model || 'auto';
    console.log(`${C.magenta}${C.bold}╔════════════════════════════════════════════════════════╗${C.reset}`);
    console.log(`${C.magenta}${C.bold}║    DARDCOR AGENT — CLI                               ║${C.reset}`);
    console.log(`${C.magenta}${C.bold}╚════════════════════════════════════════════════════════╝${C.reset}`);
    console.log(`  Mode: ${modeLabel}  |  Provider: ${C.cyan}${provider}${C.reset}  |  Model: ${C.cyan}${model}${C.reset}`);
    console.log(`  ${C.dim}Tab: switch mode | /help: commands | Ctrl+C: exit${C.reset}`);
    console.log(`${C.dim}${'─'.repeat(60)}${C.reset}\n`);
    if (sessionHistory.length > 0) {
      const start = Math.max(0, sessionHistory.length - 10);
      sessionHistory.slice(start).forEach(entry => {
        if (entry.role === 'user') {
          console.log(`  ${C.bold}${C.cyan}You${C.reset}  ${entry.content}`);
        } else {
          console.log(`  ${C.bold}${C.magenta}Dardcor${C.reset}  ${entry.content}`);
        }
      });
      console.log();
    }
  }

  function showTyping() {
    process.stdout.write(`  ${C.dim}Dardcor is thinking...${C.reset}\r`);
  }

  async function sendToAgent(message) {
    isProcessing = true;
    showTyping();
    try {
      const body = JSON.stringify({
        message,
        conversation_id: conversationId,
      });
      const result = await httpPost(`${GATEWAY_URL}/api/agent`, body);
      if (result) {
        if (result.conversation_id) conversationId = result.conversation_id;
        return result.content || result.message || JSON.stringify(result);
      }
    } catch (e) {
      return `Error: ${e.message}`;
    } finally {
      isProcessing = false;
    }
    return 'No response from agent.';
  }

  async function handleBuiltinCommand(input) {
    const cmd = input.trim().toLowerCase();
    if (cmd === '/help' || cmd === '/?') {
      return `${C.bold}Dardcor CLI Commands:${C.reset}
  /help            Show this help
  /clear           Clear screen
  /new             Start new conversation
  /history         Show session history
  /mode build      Switch to BUILD agent (full access)
  /mode plan       Switch to PLAN agent (read-only)
  /skills          List available skills
  /doctor          Run health check
  /exit            Exit CLI
  
${C.bold}Shorthand:${C.reset}
  Tab              Toggle build/plan mode
  ultrawork <task> Full autonomous agent mode
  ulw <task>       Shorthand for ultrawork`;
    }
    if (cmd === '/clear') { renderHeader(); return null; }
    if (cmd === '/new') { conversationId = null; sessionHistory = []; renderHeader(); return 'New conversation started.'; }
    if (cmd === '/exit' || cmd === 'exit') { console.log('\nGoodbye!'); rl.close(); process.exit(0); }
    if (cmd === '/history') {
      return sessionHistory.length
        ? sessionHistory.map((e, i) => `[${i+1}] ${e.role}: ${e.content.substring(0, 80)}`).join('\n')
        : 'No history yet.';
    }
    if (cmd.startsWith('/mode ')) {
      const mode = cmd.split(' ')[1];
      if (mode === 'build' || mode === 'plan') {
        agentMode = mode;
        renderHeader();
        return `Switched to ${mode.toUpperCase()} mode.`;
      }
      return 'Unknown mode. Use: /mode build | /mode plan';
    }
    if (cmd === '/skills') {
      return await sendToAgent('list skills');
    }
    if (cmd === '/doctor') {
      await runDoctorCheck();
      return null;
    }
    if (cmd === 'tab') {
      agentMode = agentMode === 'build' ? 'plan' : 'build';
      renderHeader();
      return null;
    }
    return undefined;
  }

  renderHeader();

  function promptUser() {
    if (isProcessing) return;
    const modeIndicator = agentMode === 'build' ? `${C.green}build${C.reset}` : `${C.blue}plan${C.reset}`;
    rl.question(`${C.magenta}dardcor${C.reset}:${modeIndicator}${C.cyan}>${C.reset} `, async (input) => {
      if (!input.trim()) { promptUser(); return; }

      if (input === '\t') {
        agentMode = agentMode === 'build' ? 'plan' : 'build';
        renderHeader();
        promptUser();
        return;
      }

      let processedInput = input;
      if (input.toLowerCase().startsWith('ultrawork ') || input.toLowerCase().startsWith('ulw ')) {
        const task = input.replace(/^(ultrawork|ulw)\s+/i, '');
        processedInput = `[ULTRAWORK MODE] Please complete this task autonomously using all available tools and sub-operations: ${task}`;
        inf('ULTRAWORK mode activated — running full agent...');
      }

      const builtinResult = await handleBuiltinCommand(input);
      if (builtinResult !== undefined) {
        if (builtinResult) {
          console.log(`\n${builtinResult}\n`);
        }
        promptUser();
        return;
      }

      if (agentMode === 'plan') {
        processedInput = `[READ-ONLY ANALYSIS MODE - do not execute commands or modify files] ${processedInput}`;
      }

      sessionHistory.push({ role: 'user', content: input });
      const response = await sendToAgent(processedInput);
      process.stdout.write('\r' + ' '.repeat(50) + '\r');

      if (response) {
        sessionHistory.push({ role: 'assistant', content: response });
        console.log(`\n  ${C.bold}${C.magenta}Dardcor:${C.reset}\n`);
        response.split('\n').forEach(line => console.log(`  ${line}`));
        console.log();
      }
      promptUser();
    });
  }

  promptUser();

  if (process.stdin.isTTY) {
    process.stdin.setRawMode(false);
  }
}

async function runOnboardFlow(cliOnly = false) {
  hdr('=== DARDCOR SETUP WIZARD ===');
  console.log(`${C.dim}Setup AI provider, model, and preferences.${C.reset}\n`);

  const cfg = loadConfig();

  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
  const ask = (q) => new Promise(resolve => rl.question(q, resolve));

  console.log(`${C.bold}[1/4] Select AI Provider:${C.reset}`);
  console.log(`  1) ${C.cyan}OpenAI${C.reset}          (GPT-4o, GPT-4.1, o3)            [API Key]`);
  console.log(`  2) ${C.cyan}Anthropic${C.reset}       (Claude Opus, Sonnet, Haiku)      [API Key]`);
  console.log(`  3) ${C.cyan}Google Gemini${C.reset}   (Gemini 2.5 Pro/Flash)            [API Key]`);
  console.log(`  4) ${C.cyan}DeepSeek${C.reset}        (DeepSeek-V3/R1)                  [API Key]`);
  console.log(`  5) ${C.cyan}OpenRouter${C.reset}      (Multi-provider router)            [API Key]`);
  console.log(`  6) ${C.cyan}Ollama${C.reset}          (Local models — free)              [local]`);
  console.log(`  7) ${C.cyan}Local${C.reset}           (Built-in Dardcor Agent)           [no key]`);

  const providerChoice = await ask(`\n  Select [1-7, default: 7]: `);
  const providerMap = {
    '1': { name: 'openai',    base: 'https://api.openai.com/v1',        defaultModel: 'gpt-4o' },
    '2': { name: 'anthropic', base: 'https://api.anthropic.com',         defaultModel: 'claude-sonnet-4-20250514' },
    '3': { name: 'gemini',    base: 'https://generativelanguage.googleapis.com', defaultModel: 'gemini-2.5-pro' },
    '4': { name: 'deepseek',  base: 'https://api.deepseek.com/v1',      defaultModel: 'deepseek-chat' },
    '5': { name: 'openrouter',base: 'https://openrouter.ai/api/v1',     defaultModel: 'anthropic/claude-sonnet-4-20250514' },
    '6': { name: 'ollama',    base: 'http://localhost:11434/v1',         defaultModel: 'llama3.2' },
    '7': { name: 'local',     base: '',                                  defaultModel: 'dardcor-agent' },
  };
  const chosen = providerMap[providerChoice.trim() || '7'];
  cfg.provider = chosen.name;
  cfg.provider_base_url = chosen.base;

  if (chosen.name !== 'local' && chosen.name !== 'ollama') {
    console.log(`\n${C.bold}[2/4] API Key for ${chosen.name}:${C.reset}`);
    const existing = cfg.api_key ? `${C.dim}(press Enter to keep existing)${C.reset}` : '';
    const apiKey = await ask(`  API Key ${existing}: `);
    if (apiKey.trim()) cfg.api_key = apiKey.trim();
  } else {
    cfg.api_key = '';
  }

  console.log(`\n${C.bold}[3/4] Model (default: ${chosen.defaultModel}):${C.reset}`);
  const modelInput = await ask(`  Model: `);
  cfg.model = modelInput.trim() || chosen.defaultModel;

  console.log(`\n${C.bold}[4/4] Gateway Port (default: 25000):${C.reset}`);
  const portInput = await ask(`  Port: `);
  cfg.port = portInput.trim() || '25000';

  cfg.initialized = true;
  cfg.setup_date = new Date().toISOString();

  saveConfig(cfg);
  rl.close();

  console.log(`\n${C.green}${C.bold}Configuration saved: ${CONFIG_FILE}${C.reset}`);
  console.log(`${C.dim}  Provider: ${cfg.provider} | Model: ${cfg.model} | Port: ${cfg.port}${C.reset}\n`);

  process.env.DARDCOR_AI_PROVIDER = cfg.provider;
  process.env.DARDCOR_AI_MODEL = cfg.model;
  process.env.DARDCOR_API_KEY = cfg.api_key || '';
  process.env.PORT = cfg.port;
}

async function runDoctor() {
  printBanner('Doctor — Health Check & Auto-Repair');
  await runDoctorCheck();
}

async function runDoctorCheck() {
  hdr('=== HEALTH CHECK ===');
  const cfg = loadConfig();
  let issues = 0;
  let fixed = 0;

  if (fs.existsSync(CONFIG_FILE)) {
    ok(`Config: ${CONFIG_FILE}`);
  } else {
    wrn(`Config not found. Creating default...`);
    ensureConfigDir();
    saveConfig({ provider: 'local', model: 'dardcor-agent', port: '25000', initialized: true });
    fixed++;
    ok('Config created with default settings.');
  }

  const reloadedCfg = loadConfig();

  if (reloadedCfg.provider) {
    ok(`Provider: ${reloadedCfg.provider} | Model: ${reloadedCfg.model || 'auto'}`);
  } else {
    wrn('Provider not configured. Setting to local...');
    reloadedCfg.provider = 'local';
    reloadedCfg.model = 'dardcor-agent';
    reloadedCfg.initialized = true;
    saveConfig(reloadedCfg);
    fixed++;
    ok('Provider set to local.');
  }

  if (reloadedCfg.provider === 'local' || reloadedCfg.provider === 'ollama') {
    ok(`API Key: not required (${reloadedCfg.provider})`);
  } else if (reloadedCfg.api_key) {
    ok(`API Key: ***${reloadedCfg.api_key.slice(-6)} (configured)`);
  } else {
    issues++;
    err('API Key: missing — run dardcor cli to configure');
  }

  if (!reloadedCfg.port) {
    wrn('Port not set. Setting to 25000...');
    reloadedCfg.port = '25000';
    saveConfig(reloadedCfg);
    fixed++;
    ok('Port set to 25000.');
  }

  const batPath = path.join(__dirname, 'dardcor.bat');
  const exePath = path.join(__dirname, 'dardcor-agent.exe');
  if (fs.existsSync(exePath)) {
    ok(`Backend binary: ${exePath}`);
  } else {
    wrn(`Backend binary not found. Go runtime will be used.`);
    try {
      execSync('go version', { timeout: 5000, stdio: 'pipe' });
      ok('Go runtime: available');
    } catch {
      issues++;
      err('Go runtime: not found — install Go from https://go.dev');
    }
  }
  if (fs.existsSync(batPath)) {
    ok(`Launcher script: ${batPath}`);
  } else {
    issues++;
    err(`dardcor.bat not found!`);
  }

  inf('Checking gateway...');
  const serverUp = await checkServer(GATEWAY_URL);
  if (serverUp) {
    ok(`Gateway: online at ${GATEWAY_URL}`);
    try {
      const info = await httpGet(`${GATEWAY_URL}/api/system`);
      ok(`System: ${info.hostname || 'ok'} | OS: ${info.os?.platform || 'ok'}`);
    } catch {
      ok('Gateway: responding (system info not available)');
    }
  } else {
    wrn(`Gateway: offline — run 'dardcor run' to start`);
  }

  const nodeVerStr = process.version;
  const major = parseInt(nodeVerStr.replace('v', '').split('.')[0]);
  if (major >= 18) {
    ok(`Node.js: ${nodeVerStr}`);
  } else {
    issues++;
    err(`Node.js: ${nodeVerStr} — requires v18+`);
  }

  const port = reloadedCfg.port || DEFAULT_PORT;
  inf(`Port: ${port}`);

  const dataDir = path.join(__dirname, 'data');
  const requiredDirs = [
    dataDir,
    path.join(dataDir, 'conversations'),
    path.join(dataDir, 'commands'),
    path.join(dataDir, 'settings'),
  ];
  for (const dir of requiredDirs) {
    if (!fs.existsSync(dir)) {
      wrn(`Directory missing: ${dir} — creating...`);
      fs.mkdirSync(dir, { recursive: true });
      fixed++;
      ok(`Created: ${dir}`);
    }
  }

  console.log(`\n${C.bold}=== DIAGNOSIS COMPLETE ===${C.reset}`);
  if (fixed > 0) {
    ok(`${fixed} issue(s) auto-repaired.`);
  }
  if (issues > 0) {
    err(`${issues} issue(s) require manual attention.`);
  } else {
    ok('All systems operational.');
  }
  console.log();
}

function startBackend(mode) {
  const cfg = loadConfig();
  const env = {
    ...process.env,
    DARDCOR_AI_PROVIDER: cfg.provider || 'local',
    DARDCOR_AI_MODEL:    cfg.model    || 'dardcor-agent',
    DARDCOR_API_KEY:     cfg.api_key  || '',
    PORT:                cfg.port     || DEFAULT_PORT,
    DARDCOR_DEV_URL:     process.env.DARDCOR_DEV_URL || '',
  };

  const exePath = path.join(__dirname, 'dardcor-agent.exe');
  let cmd, args;

  if (fs.existsSync(exePath)) {
    inf(`Starting production binary: ${exePath}`);
    cmd = exePath;
    args = [mode];
  } else {
    inf('Production binary not found. Using "go run main.go"...');
    cmd = 'go';
    args = ['run', 'main.go', mode];
  }

  const child = spawn(cmd, args, {
    cwd: __dirname,
    stdio: 'inherit',
    shell: true,
    env,
  });

  child.on('error', (err) => {
    err(`Failed to start backend: ${err.message}`);
  });

  child.on('exit', (code) => {
    if (code !== 0 && code !== null) {
      wrn(`Backend exited with code ${code}. Restarting in 3s...`);
      setTimeout(() => startBackend(mode), 3000);
    }
  });

  process.on('SIGINT', () => {
    child.kill('SIGINT');
    process.exit(0);
  });
}

function killOldInstance() {
  try {
    execSync('taskkill /F /IM dardcor-agent.exe /T >nul 2>&1', { timeout: 3000 });
  } catch { }
}

function httpGet(url) {
  return new Promise((resolve, reject) => {
    http.get(url, (res) => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => {
        try { resolve(JSON.parse(data)); } catch { resolve(data); }
      });
    }).on('error', reject);
  });
}

function httpPost(url, body) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const options = {
      hostname: u.hostname,
      port:     u.port || 80,
      path:     u.pathname,
      method:   'POST',
      headers:  { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) },
    };
    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => {
        try { resolve(JSON.parse(data)); } catch { resolve(data); }
      });
    });
    req.on('error', reject);
    req.write(body);
    req.end();
  });
}

async function checkServer(url) {
  try {
    await httpGet(`${url}/api/system`);
    return true;
  } catch {
    return false;
  }
}

async function waitForServer(url, maxTries = 10) {
  for (let i = 0; i < maxTries; i++) {
    if (await checkServer(url)) return true;
    await sleep(1000);
    process.stdout.write(`\r${C.dim}Waiting for gateway... (${i+1}/${maxTries})${C.reset}`);
  }
  process.stdout.write('\n');
  return false;
}

function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}
