import { spawn, execSync, exec } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';
import { printBanner } from './help.js';
import { initializeSystem } from './init_agent.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const cwd = process.cwd();
const isLocalProject = fs.existsSync(path.join(cwd, 'package.json')) &&
  fs.readFileSync(path.join(cwd, 'package.json'), 'utf8').includes('dardcor-agent');

const rootDir = isLocalProject ? cwd : path.join(__dirname, '..');

const C = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  purple: '\x1b[38;5;93m',
  dim: '\x1b[2m',
};

const inf = (msg) => console.log(`${C.purple}[i]${C.reset} ${msg}`);
const errLog = (msg) => console.log(`${C.red}[!] ERROR: ${msg}${C.reset}`);

function killPort25000() {
  try {
    if (process.platform === 'win32') {
      execSync(`for /f "tokens=5" %p in ('netstat -ano ^| findstr :25000 ^| findstr LISTENING') do taskkill /F /PID %p /T >nul 2>&1`, { timeout: 3000, shell: true });
    }
  } catch { }
}

function killPort5173() {
    try {
      if (process.platform === 'win32') {
        execSync(`for /f "tokens=5" %p in ('netstat -ano ^| findstr :5173 ^| findstr LISTENING') do taskkill /F /PID %p /T >nul 2>&1`, { timeout: 3000, shell: true });
      }
    } catch { }
}

export async function run() {
  try {
    console.clear();
    printBanner('DARDCOR');

    inf('Initializing Agent...');
    initializeSystem(rootDir);
    killPort25000();

    const devMode = isLocalProject;
    if (devMode) {
      inf(`${C.yellow}Development Mode Detected (Real-time UI Enabled)${C.reset}`);
      killPort5173();
      spawn('npm', ['run', 'dev'], {
        cwd: rootDir,
        stdio: 'ignore',
        shell: true,
        detached: false
      });
      process.env.DARDCOR_DEV = "true";
    }

    const port = '25000';
    const devUrl = `http://127.0.0.1:${port}`;
    process.env.PORT = port;

    if (!devMode && !fs.existsSync(path.join(rootDir, 'dist'))) {
      inf('Compiling UI Assets...');
      execSync('npm run build', { cwd: rootDir, stdio: 'inherit' });
    }

    inf('Igniting Engine...');
    const backend = spawn('go', ['run', 'main.go'], {
      cwd: rootDir,
      stdio: 'inherit',
      shell: true,
      env: { ...process.env, GOMAXPROCS: '' + os.cpus().length }
    });

    backend.on('exit', (code) => {
      if (code !== 0 && code !== null) {
        errLog(`Engine stalled unexpectedly (Exit Code: ${code})`);
      }
    });

    const cleanup = () => {
      try {
        if (process.platform === 'win32') {
          execSync(`taskkill /F /IM go.exe /T >nul 2>&1`, { shell: true });
          killPort25000();
          if (devMode) killPort5173();
        }
      } catch { }
    };

    process.on('SIGINT', () => { cleanup(); process.exit(); });
    process.on('SIGTERM', () => { cleanup(); process.exit(); });

    setTimeout(() => {
      try {
        const openCmd = process.platform === 'win32' ? `start "" "${devUrl}"` : `open "${devUrl}"`;
        exec(openCmd);
      } catch { }
    }, devMode ? 6000 : 3000);

    console.log(`\n${C.purple}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C.reset}`);
    console.log(`${C.purple}  Status     →${C.reset} ${C.green}${C.bold}DARDCOR ENGINE ACTIVE${C.reset}`);
    console.log(`${C.purple}  Interface  →${C.reset} ${C.bold}Dashboard: ${devUrl}${C.reset}`);
    console.log(`${C.purple}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C.reset}`);
    console.log(`${C.reset}  Press ${C.bold}Ctrl+C${C.reset} to stop the engine\n`);

    process.stdin.resume();
    await new Promise(() => {});
  } catch (err) {
    errLog(`Runtime Error: ${err.message}`);
    process.exit(1);
  }
}
