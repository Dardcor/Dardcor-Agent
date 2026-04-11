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
};

const inf = (msg) => console.log(`${C.purple}[i]${C.reset} ${msg}`);
const wrn = (msg) => console.log(`${C.yellow}[!]${C.reset} ${msg}`);

const CONFIG_DIR = path.join(os.homedir(), '.dardcor');
const CONFIG_FILE = path.join(CONFIG_DIR, 'config.json');

function loadConfig() {
  if (fs.existsSync(CONFIG_FILE)) {
    try { return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8')); } catch { return {}; }
  }
  return {};
}

function killOldInstance() {
  try {
    if (process.platform === 'win32') {
      execSync(`for /f "tokens=5" %p in ('netstat -ano ^| findstr :25000 ^| findstr LISTENING') do taskkill /F /PID %p /T >nul 2>&1`, { timeout: 2000, shell: true });
    }
  } catch { }
}

export async function run() {
  const cfg = loadConfig();
  console.clear();
  printBanner('DARDCOR');


  initializeSystem(rootDir);


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

  const distPath = path.join(rootDir, 'dist');
  const needsBuild = !fs.existsSync(path.join(distPath, 'index.html'));

  if (needsBuild) {
    process.stdout.write(`${C.purple}[i]${C.reset} Synchronizing System UI... `);
    try {
      execSync('npm run build', { cwd: rootDir, stdio: 'ignore' });
      process.stdout.write(`${C.green}READY${C.reset}\n`);
    } catch (error) {
      process.stdout.write(`${C.red}FAILED${C.reset}\n`);
      if (fs.existsSync(distPath)) {
        wrn('System build failed. Using legacy assets...');
      } else {
        console.error(`${C.red}[!]${C.reset} Critical: Build failed and no legacy assets found. Aborting.`);
        process.exit(1);
      }
    }
  }

  console.log(`${C.purple}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C.reset}`);
  console.log(`${C.purple}  Status     →${C.reset} ${C.green}${C.bold}DARDCOR ENGINE ACTIVE${C.reset}`);
  console.log(`${C.purple}  Interface  →${C.reset} ${C.bold}Dashboard: ${devUrl}${C.reset}`);
  console.log(`${C.purple}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C.reset}`);

  let backend;
  if (process.platform === 'win32') {
    const vbsPath = path.join(os.tmpdir(), "dardcor_dashboard_stealth.vbs");
    const batPath = path.join(os.tmpdir(), "dardcor_dashboard.bat");
    
    fs.writeFileSync(batPath, `cd /d "${rootDir}"\ngo run main.go`);
    fs.writeFileSync(vbsPath, `Set WshShell = CreateObject("WScript.Shell")\nWshShell.Run chr(34) & "${batPath}" & Chr(34), 0\nSet WshShell = Nothing`);
    
    backend = spawn('cscript', ['//nologo', vbsPath], {
      cwd: rootDir,
      stdio: 'ignore',
      shell: false,
      windowsHide: true,
      detached: true
    });
    backend.unref();
  } else {
    backend = spawn('go', ['run', 'main.go'], {
      cwd: rootDir,
      stdio: 'ignore',
      shell: false,
      detached: true,
      env: { ...process.env, GOMAXPROCS: '' + os.cpus().length },
    });
    backend.unref();
  }

  const cleanup = () => {
    try {
      if (process.platform === 'win32') {
        execSync(`taskkill /F /IM go.exe /T >nul 2>&1`, { shell: true });
      } else if (backend) {
        backend.kill();
      }
    } catch { }
  };

  process.on('SIGINT', cleanup);
  process.on('SIGTERM', cleanup);
  process.on('exit', cleanup);

  setTimeout(async () => {
    try {
      const openCmd = process.platform === 'win32' ? `start "" "${devUrl}"` : `open "${devUrl}"`;
      exec(openCmd);
    } catch { }
  }, 3000);
}
