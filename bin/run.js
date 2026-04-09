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
  purple:  '\x1b[38;5;93m',
};

const ok  = (msg) => console.log(`${C.purple}[✓]${C.reset} ${msg}`);
const inf = (msg) => console.log(`${C.purple}[i]${C.reset} ${msg}`);
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
      execSync(`for /f "tokens=5" %p in ('netstat -ano ^| findstr :25000 ^| findstr LISTENING') do taskkill /F /PID %p /T >nul 2>&1`, { timeout: 2000, shell: true });
    }
  } catch {}
}

export async function run() {
  const cfg = loadConfig();
  
  console.clear();
  printBanner();

  if (!fs.existsSync(path.join(rootDir, 'node_modules'))) {
    inf('Optimizing Dependencies...');
    execSync('npm install', { cwd: rootDir, stdio: 'ignore' });
  }

  killOldInstance();

  const port = '25000';
  const devUrl = `http://127.0.0.1:${port}`;

  process.env.PORT = port;
  process.env.DARDCOR_AI_PROVIDER = cfg.provider || 'local';
  process.env.DARDCOR_AI_MODEL = cfg.model || 'dardcor-agent';
  process.env.DARDCOR_API_KEY = cfg.api_key || '';

  if (fs.existsSync(path.join(rootDir, 'dist'))) {
    fs.rmSync(path.join(rootDir, 'dist'), { recursive: true, force: true });
  }

  process.stdout.write(`${C.purple}[i]${C.reset} Compiling Hyper-Agent UI... `);
  try {
    execSync('npm run build', { cwd: rootDir, stdio: 'ignore' });
    process.stdout.write(`${C.green}COMPLETE${C.reset}\n`);
  } catch (e) {
    process.stdout.write(`${C.red}FAILED${C.reset}\n`);
    console.error(e);
  }

  console.log(`${C.purple}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C.reset}`);
  console.log(`${C.purple}  Status     →${C.reset} ${C.green}${C.bold}OPTIMIZED AGENT ACTIVE${C.reset}`);
  console.log(`${C.purple}  Dashboard  →${C.reset} ${C.bold}${devUrl}${C.reset}`);
  console.log(`${C.purple}  Provider   →${C.reset} ${C.dim}${process.env.DARDCOR_AI_PROVIDER} | ${process.env.DARDCOR_AI_MODEL}${C.reset}`);
  console.log(`${C.purple}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C.reset}`);

  if (fs.existsSync(path.join(rootDir, 'src'))) {
    spawn(process.platform === 'win32' ? 'npx.cmd' : 'npx', ['vite', 'build', '--watch', '--emptyOutDir', 'false'], {
      cwd: rootDir,
      stdio: 'ignore',
      shell: true
    });
  }

  const backend = spawn('go', ['run', 'main.go', 'run'], {
    cwd: rootDir,
    stdio: 'ignore',
    shell: true,
    env: process.env,
  });

  const cleanup = () => {
    try {
      if (process.platform === 'win32') {
        execSync(`taskkill /F /T /PID ${backend.pid} >nul 2>&1`, { shell: true });
      } else {
        backend.kill();
      }
    } catch (e) {}
  };

  process.on('SIGINT', () => {
    cleanup();
    process.exit();
  });

  process.on('SIGTERM', () => {
    cleanup();
    process.exit();
  });

  backend.on('exit', (code) => {
    if (code !== 0 && code !== null) {
      wrn(`Agent Workspace Engine stopped (Code: ${code})`);
    }
  });

  setTimeout(() => {
    const openCmd = process.platform === 'win32' ? `start "" "${devUrl}"` : `open "${devUrl}"`;
    exec(openCmd);
  }, 2000);
}
