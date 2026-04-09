import { spawn, execSync, exec } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';
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

function killOldInstance() {
  try {
    if (process.platform === 'win32') {
      execSync('taskkill /F /IM dardcor-agent.exe /T >nul 2>&1', { timeout: 2000 });
      execSync(`for /f "tokens=5" %p in ('netstat -ano ^| findstr :25000 ^| findstr LISTENING') do taskkill /F /PID %p /T >nul 2>&1`, { timeout: 2000 });
    }
  } catch {}
}

export async function run() {
  const cfg = loadConfig();
  printBanner('Dardcor Agent Runtime (Port 25000)');

  if (!fs.existsSync(path.join(rootDir, 'node_modules'))) {
    inf('Installing dependencies...');
    execSync('npm install', { cwd: rootDir, stdio: 'inherit' });
  }

  killOldInstance();

  const port = '25000';
  const devUrl = `http://127.0.0.1:${port}`;

  process.env.PORT = port;
  process.env.DARDCOR_AI_PROVIDER = cfg.provider || 'local';
  process.env.DARDCOR_AI_MODEL = cfg.model || 'dardcor-agent';
  process.env.DARDCOR_API_KEY = cfg.api_key || '';

  console.clear();
  inf('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  inf('           DARDCOR UNIFIED RUNTIME           ');
  inf('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  inf(`Dashboard  → ${C.bold}${devUrl}${C.reset}`);
  inf(`Status     → ${C.green}OPTIMIZED AGENT ACTIVE${C.reset}`);
  inf('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  if (fs.existsSync(path.join(rootDir, 'dist'))) {
    fs.rmSync(path.join(rootDir, 'dist'), { recursive: true, force: true });
  }
  inf('Building UI...');
  execSync('npm run build', { cwd: rootDir, stdio: 'inherit' });

  const exePath = path.join(rootDir, 'dardcor-agent.exe');
  let cmd, args;

  if (fs.existsSync(exePath)) {
    cmd = exePath;
    args = ['run'];
  } else {
    if (fs.existsSync(path.join(rootDir, 'src'))) {
      inf('Real-Time UI Builder Active...');
      spawn(process.platform === 'win32' ? 'npx.cmd' : 'npx', ['vite', 'build', '--watch', '--emptyOutDir', 'false'], {
        cwd: rootDir,
        stdio: 'ignore',
        shell: false
      });
    }
    cmd = 'go';
    args = ['run', 'main.go', 'run'];
  }

  const backend = spawn(cmd, args, {
    cwd: rootDir,
    stdio: 'inherit',
    shell: false,
    env: process.env,
  });

  backend.on('exit', (code) => {
    if (code !== 0 && code !== null) {
      wrn(`Backend exited with code ${code}.`);
    }
  });

  setTimeout(() => {
    const openCmd = process.platform === 'win32' ? `start "" "${devUrl}"` : `open "${devUrl}"`;
    exec(openCmd);
  }, 3000);
}
