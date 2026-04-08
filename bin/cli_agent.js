import readline from 'readline';
import http from 'http';
import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';
import { spawn } from 'child_process';
import { printBanner } from './help.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.join(__dirname, '..');

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
};

const ok  = (msg) => console.log(`${C.green}[✓]${C.reset} ${msg}`);
const err = (msg) => console.log(`${C.red}[✗]${C.reset} ${msg}`);
const inf = (msg) => console.log(`${C.cyan}[i]${C.reset} ${msg}`);
const wrn = (msg) => console.log(`${C.yellow}[!]${C.reset} ${msg}`);

const CONFIG_DIR = path.join(os.homedir(), '.dardcor');
const CONFIG_FILE = path.join(CONFIG_DIR, 'config.json');

function loadConfig() {
  if (fs.existsSync(CONFIG_FILE)) {
    try {
      return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8'));
    } catch {
      return {};
    }
  }
  return {};
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
  return new Promise((resolve) => {
    http.get(url, (res) => {
      resolve(res.statusCode === 200);
    }).on('error', () => resolve(false));
  });
}

export async function runCLI() {
  const cfg = loadConfig();
  printBanner('Interactive CLI Terminal Agent');

  const gateUrl = 'http://127.0.0.1:25000';
  
  const serverUp = await checkServer(gateUrl + '/api/system');
  if (!serverUp) {
    inf('Starting backend for CLI mode...');
    const env = { ...process.env, PORT: '25000', DARDCOR_AI_PROVIDER: cfg.provider || 'local' };
    spawn('go', ['run', 'main.go', 'run'], { cwd: rootDir, stdio: 'ignore', shell: true, env, detached: true }).unref();
    
    let connected = false;
    for (let i = 0; i < 10; i++) {
      if (await checkServer(gateUrl + '/api/system')) { connected = true; break; }
      await new Promise(r => setTimeout(r, 1000));
    }
    if (!connected) {
      err('Could not establish connection to Dardcor Agent.');
      process.exit(1);
    }
  }

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: true,
  });

  let convId = null;
  let mode = 'build';
  let history = [];
  let busy = false;

  function render() {
    console.clear();
    console.log(`${C.magenta}${C.bold}╔════════════════════════════════════════════════════════╗${C.reset}`);
    console.log(`${C.magenta}${C.bold}║    DARDCOR TERMINAL AGENT                              ║${C.reset}`);
    console.log(`${C.magenta}${C.bold}╚════════════════════════════════════════════════════════╝${C.reset}`);
    console.log(`  Mode: ${mode === 'build' ? C.green + 'BUILD' : C.blue + 'PLAN'}  | Provider: ${C.cyan}${cfg.provider || 'local'}${C.reset}`);
    console.log(`${C.dim}────────────────────────────────────────────────────────────${C.reset}\n`);
    history.slice(-10).forEach(e => {
      if (e.role === 'user') console.log(`  ${C.bold}${C.cyan}User:${C.reset} ${e.content}`);
      else console.log(`  ${C.bold}${C.magenta}Agent:${C.reset} ${e.content}\n`);
    });
  }

  async function ask() {
    if (busy) return;
    rl.question(`${C.magenta}dardcor${C.reset}:${C.cyan}${mode}${C.reset}> `, async (input) => {
      if (!input.trim()) { ask(); return; }
      if (input === '/exit' || input === 'exit') { rl.close(); process.exit(0); }
      if (input === '/clear') { history = []; render(); ask(); return; }
      if (input === '/mode') { mode = mode === 'build' ? 'plan' : 'build'; render(); ask(); return; }

      let msg = input;
      if (mode === 'plan') msg = `[PLAN MODE] ${msg}`;

      history.push({ role: 'user', content: input });
      busy = true;
      process.stdout.write(`  ${C.dim}Thinking...${C.reset}\r`);

      try {
        const res = await httpPost(gateUrl + '/api/agent', JSON.stringify({ message: msg, conversation_id: convId }));
        if (res.conversation_id) convId = res.conversation_id;
        const reply = res.content || 'No response.';
        history.push({ role: 'agent', content: reply });
        render();
      } catch (e) {
        err(`Communication error: ${e.message}`);
      } finally {
        busy = false;
        ask();
      }
    });
  }

  render();
  ask();
}
