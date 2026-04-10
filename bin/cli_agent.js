import readline from 'readline';
import http from 'http';
import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';

import { spawn } from 'child_process';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.join(__dirname, '..');

const C = {
  reset:   '\x1b[0m',
  bold:    '\x1b[1m',
  dim:     '\x1b[2m',
  italic:  '\x1b[3m',
  red:     '\x1b[31m',
  green:   '\x1b[32m',
  yellow:  '\x1b[33m',
  blue:    '\x1b[34m',
  magenta: '\x1b[35m',
  cyan:    '\x1b[36m',
  white:   '\x1b[37m',
  purple:  '\x1b[38;5;93m',
  bgGray:  '\x1b[48;5;235m',
};

const pos = (r, c) => process.stdout.write(`\x1b[${r};${c}H`);
const clear = () => process.stdout.write('\x1b[H\x1b[2J');

const CONFIG_DIR = path.join(os.homedir(), '.dardcor');
const CONFIG_FILE = path.join(CONFIG_DIR, 'config.json');

function loadConfig() {
  if (fs.existsSync(CONFIG_FILE)) {
    try { return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8')); } catch { return {}; }
  }
  return {};
}

function checkServer(url) {
  return new Promise((resolve) => {
    http.get(url, (res) => {
      resolve(res.statusCode === 200);
    }).on('error', () => resolve(false));
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
      res.on('end', () => { try { resolve(JSON.parse(data)); } catch { resolve(data); } });
    });
    req.on('error', reject);
    req.write(body);
    req.end();
  });
}

const banner = [
  '  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ',
  '  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗',
  '  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝',
  '  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗',
  '  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║',
  '  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝'
];

export async function runCLI() {
  const cfg = loadConfig();
  const gateUrl = 'http://127.0.0.1:25000';
  
  const serverUp = await checkServer(gateUrl + '/api/system');
  if (!serverUp) {
    const env = { ...process.env, PORT: '25000', DARDCOR_AI_PROVIDER: cfg.provider || 'local' };
    

    const child = spawn('go', ['run', 'main.go', 'run'], { 
      cwd: rootDir, 
      stdio: 'ignore', 
      shell: true, 
      env, 
      detached: true,
      windowsHide: true,
    });
    child.unref();


    clear();
    const cols = process.stdout.columns;
    const bannerX = Math.floor((cols - 59) / 2);
    banner.forEach((line, i) => { pos(5 + i, bannerX); process.stdout.write(`${C.purple}${C.bold}${line}${C.reset}`); });
    
    let connected = false;
    for (let i = 0; i < 30; i++) {
        const dots = '.'.repeat((i % 3) + 1).padEnd(3);
        pos(13, Math.floor(cols/2) - 15);
        process.stdout.write(`${C.cyan}Initialize Hyper-Engine ${C.bold}${dots}${C.reset}  ${C.dim}[${i}/30s]${C.reset}`);
        if (await checkServer(gateUrl + '/api/system')) { connected = true; break; }
        await new Promise(r => setTimeout(r, 1000));
    }
    if (!connected) {
        process.stdout.write(`\n\n  ${C.red}System failure: Could not reach engine gateway.${C.reset}\n`);
        process.exit(1);
    }
  }

  let convId = null;
  let history = [];
  let busy = false;
  let lastResponse = '';

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: true,
  });

  function drawUI(history) {
    clear();
    const cols = process.stdout.columns;
    const rows = process.stdout.rows;

    const bannerX = Math.floor((cols - 59) / 2);
    banner.forEach((line, i) => {
      pos(2 + i, bannerX);
      process.stdout.write(`${C.purple}${C.bold}${line}${C.reset}`);
    });

    const dividerY = 9;
    pos(dividerY, 5);
    process.stdout.write(`${C.purple}${C.bold}DARDCOR SYSTEM${C.reset}  ${C.dim}────────────────────────────────────────────────────────────────────────${C.reset}`);
    
    pos(dividerY + 1, 5);
    process.stdout.write(`${C.cyan}${C.bold}Sisyphus (Ultraworker)${C.reset}  ${C.white}Claude Opus 4.5 Thinking (Antigravity)${C.reset}  ${C.white}${C.dim}Google${C.reset}`);

    const historyY = dividerY + 3;
    const maxHistory = rows - historyY - 8;
    const displayHistory = history.slice(-Math.max(1, maxHistory));

    displayHistory.forEach((msg, i) => {
      pos(historyY + i*2, 5);
      if (msg.role === 'user') {
        process.stdout.write(`${C.blue}${C.bold}USER:${C.reset} ${msg.content.slice(0, cols - 15)}`);
      } else {
        process.stdout.write(`${C.magenta}${C.bold}DARDCOR:${C.reset} ${C.white}${msg.content.slice(0, cols - 15)}${C.reset}`);
      }
    });

    const boxY = rows - 6;
    const boxW = Math.min(cols - 10, 100);
    const boxX = Math.floor((cols - boxW) / 2);

    for (let i = 0; i < 3; i++) {
        pos(boxY + i, boxX);
        process.stdout.write(`${C.bgGray}${' '.repeat(boxW)}${C.reset}`);
    }

    pos(boxY, boxX);
    process.stdout.write(`${C.cyan}┃${C.reset}`);
    pos(boxY+1, boxX);
    process.stdout.write(`${C.cyan}┃${C.reset}  ${C.white}${C.bgGray}${C.bold}❯${C.reset} ${C.white}${C.bgGray}Ask Dardcor... ${C.dim}(Type /exit to quit)${C.reset}`);
    pos(boxY+2, boxX);
    process.stdout.write(`${C.cyan}┃${C.reset}`);

    const statusRow = rows;
    const gitStatus = (() => {
      try {
        const branch = require('child_process').execSync('git branch --show-current', { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] }).trim();
        return `${C.green}${C.bold}git:${C.reset}${C.bold}${branch}${C.reset}`;
      } catch { return `${C.dim}no-git${C.reset}`; }
    })();
    
    pos(statusRow, 1);
    const statusLine = ` ${gitStatus}  ${C.blue}${C.bold}engine:${C.reset}latest  ${C.magenta}${C.bold}model:${C.reset}${cfg.model || 'auto'}  ${C.yellow}${C.bold}mcp:${C.reset}local `;
    process.stdout.write(statusLine);
    
    pos(statusRow, cols - 12);
    process.stdout.write(`${C.dim}${C.bold}DARDCOR AGENT${C.reset} `);

    pos(boxY+1, boxX + 6);
  }

  async function ask() {
    if (busy) return;
    drawUI(history);
    rl.question('', async (input) => {
      if (!input.trim()) { ask(); return; }
      if (input === '/exit' || input === 'exit') { rl.close(); process.exit(0); }
      if (input === '/clear') { history = []; lastResponse = ''; drawUI(history); ask(); return; }
      
      history.push({ role: 'user', content: input });
      busy = true;


      const cols = process.stdout.columns;
      const boxY = process.stdout.rows - 6;
      const boxW = Math.min(cols - 10, 100);
      const boxX = Math.floor((cols - boxW) / 2);
      pos(boxY + 1, boxX + 5);
      process.stdout.write(`${C.bgGray}${C.magenta}${C.bold}❯${C.reset} ${C.bgGray}${C.italic}Thinking...${C.reset}${' '.repeat(20)}`);

      try {
        const res = await httpPost(gateUrl + '/api/agent', JSON.stringify({ message: input, conversation_id: convId }));
        if (res.conversation_id) convId = res.conversation_id;
        history.push({ role: 'agent', content: res.content || 'No response.' });
      } catch (e) {
        history.push({ role: 'agent', content: `Communication error: ${e.message}` });
      } finally {
        busy = false;
        ask();
      }
    });
  }

  process.stdout.on('resize', () => drawUI(history));
  ask();
}




