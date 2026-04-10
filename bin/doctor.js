import fs from 'fs';
import path from 'path';
import os from 'os';
import { execSync } from 'child_process';
import http from 'http';
import { fileURLToPath } from 'url';

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
const hdr = (msg) => console.log(`\n${C.magenta}${C.bold}${msg}${C.reset}`);

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

function saveConfig(cfg) {
  if (!fs.existsSync(CONFIG_DIR)) fs.mkdirSync(CONFIG_DIR, { recursive: true });
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(cfg, null, 2), 'utf8');
}

function checkServer(url) {
  return new Promise((resolve) => {
    http.get(url, (res) => {
      resolve(res.statusCode === 200);
    }).on('error', () => resolve(false));
  });
}

export async function runDoctor() {
  hdr('=== DARDCOR SYSTEM DOCTOR ===');
  const cfg = loadConfig();
  let issues = 0;
  let fixed = 0;

  if (fs.existsSync(CONFIG_FILE)) {
    ok(`Configuration: FOUND (${CONFIG_FILE})`);
  } else {
    wrn(`Configuration: NOT FOUND. Initializing...`);
    saveConfig({ provider: 'local', model: 'dardcor-agent', port: '25000', initialized: true });
    fixed++;
    ok('Configuration: CREATED with defaults');
  }

  const reloadedCfg = loadConfig();

  if (reloadedCfg.port !== '25000') {
    wrn('Port Conflict: Forcing 25000...');
    reloadedCfg.port = '25000';
    saveConfig(reloadedCfg);
    fixed++;
    ok('Port: RESOLVED to 25000');
  } else {
    ok('Port: 25000');
  }

  try {
    execSync('go version', { stdio: 'pipe' });
    ok('Go Kernel: AVAILABLE');
  } catch {
    issues++;
    err('Go Kernel: MISSING');
  }

  const nodeVer = process.version;
  const majorNode = parseInt(nodeVer.substring(1).split('.')[0]);
  if (majorNode >= 18) {
    ok(`Node Runtime: ${nodeVer}`);
  } else {
    issues++;
    err(`Node Runtime: ${nodeVer} (Requires 18+)`);
  }

  const dataDir = path.join(rootDir, 'database');
  const requiredPaths = [
    dataDir,
    path.join(dataDir, 'conversations'),
    path.join(dataDir, 'commands'),
    path.join(dataDir, 'settings'),
    path.join(dataDir, 'model', 'antigravity'),
  ];
  for (const p of requiredPaths) {
    if (!fs.existsSync(p)) {
      wrn(`Structure: ${path.basename(p)} missing — Creating...`);
      fs.mkdirSync(p, { recursive: true });
      fixed++;
      ok(`Structure: ${path.basename(p)} CREATED`);
    }
  }

  const engineUp = await checkServer('http://127.0.0.1:25000/api/system');
  if (engineUp) {
    ok('Dardcor Engine: ONLINE');
  } else {
    wrn('Dardcor Engine: OFFLINE (use dardcor run)');
  }

  hdr('=== DIAGNOSIS RESULT ===');
  if (fixed > 0) ok(`${fixed} issue(s) resolved automatically.`);
  if (issues > 0) {
    err(`${issues} issue(s) need manual intervention.`);
  } else {
    ok('System Perfect. No issues found.');
  }
}




