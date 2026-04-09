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
  hdr('=== HEALTH CHECK ===');
  const cfg = loadConfig();
  let issues = 0;
  let fixed = 0;

  if (fs.existsSync(CONFIG_FILE)) {
    ok(`Config: ${CONFIG_FILE}`);
  } else {
    wrn(`Config not found. Creating default...`);
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
    ok(`API Key: configured`);
  } else {
    issues++;
    err('API Key: missing');
  }

  if (reloadedCfg.port !== '25000') {
    wrn('Port mismatch. Forcing port 25000...');
    reloadedCfg.port = '25000';
    saveConfig(reloadedCfg);
    fixed++;
    ok('Port set to 25000.');
  } else {
    ok('Port: 25000');
  }

  try {
    execSync('go version', { stdio: 'pipe' });
    ok('Go runtime: available');
  } catch {
    issues++;
    err('Go runtime: not found');
  }

  const nodeVerStr = process.version;
  const major = parseInt(nodeVerStr.replace('v', '').split('.')[0]);
  if (major >= 18) {
    ok(`Node.js: ${nodeVerStr}`);
  } else {
    issues++;
    err(`Node.js: ${nodeVerStr} — requires v18+`);
  }

  const dataDir = path.join(rootDir, 'database');
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

  const serverUp = await checkServer('http://127.0.0.1:25000/api/system');
  if (serverUp) {
    ok('Gateway: online at http://127.0.0.1:25000');
  } else {
    wrn('Gateway: offline (run dardcor run)');
  }

  hdr('=== DIAGNOSIS COMPLETE ===');
  if (fixed > 0) ok(`${fixed} issue(s) auto-repaired.`);
  if (issues > 0) {
    err(`${issues} issue(s) require manual attention.`);
  } else {
    ok('All systems operational.');
  }
}
