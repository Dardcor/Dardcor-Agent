import readline from 'readline';
import http from 'http';
import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';
import { spawn, execSync, exec } from 'child_process';
import { initializeSystem } from './init_agent.js';

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
  bgBox:   '\x1b[48;5;16m',
  bgInput: '\x1b[48;5;234m',
};

const pos = (r, c) => process.stdout.write(`\x1b[${r};${c}H`);
const clear = () => process.stdout.write('\x1b[H\x1b[2J');

function httpCall(url, method = 'GET', body = null) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const options = {
      hostname: u.hostname, port: u.port, path: u.pathname, method,
      headers: { 'Content-Type': 'application/json' },
    };
    if (body) options.headers['Content-Length'] = Buffer.byteLength(body);
    const req = http.request(options, (res) => {
      let d = ''; res.on('data', c => d += c);
      res.on('end', () => { try { resolve(JSON.parse(d)); } catch { resolve(d); } });
    });
    req.on('error', reject);
    if (body) req.write(body);
    req.end();
  });
}

const logo = [
  '  ██████╗   █████╗  ██████╗  ██████╗   ██████╗ ██████╗  ██████╗ ',
  '  ██╔══██╗ ██╔══██╗ ██╔══██╗ ██╔══██╗ ██╔════╝██╔═══██╗ ██╔══██╗',
  '  ██║  ██║ ███████║ ██████╔╝ ██║  ██║ ██║     ██║   ██║ ██████╔╝',
  '  ██║  ██║ ██╔══██║ ██╔══██╗ ██║  ██║ ██║     ██║   ██║ ██╔══██╗',
  '  ██████╔╝ ██║  ██║ ██║  ██║ ██████╔╝ ╚██████╗╚██████╔╝ ██║  ██║',
  '  ╚═════╝  ╚═╝  ╚═╝ ╚═╝  ╚═╝ ╚═════╝   ╚═════╝ ╚═════╝  ╚═╝  ╚═╝'
];

export async function runCLI() {
  initializeSystem(rootDir);
  const gateUrl = 'http://127.0.0.1:25000';
  
  async function ensureEngine() {
    // Kill old "go run" processes to prevent terminal residue
    if (process.platform === 'win32') {
        try { execSync('taskkill /F /IM go.exe /T', { stdio: 'ignore' }); } catch {}
    }

    clear();
    console.log(`\n  ${C.purple}${C.bold}DARDCOR${C.reset}  ${C.dim}─ Initializing Source Engine...${C.reset}\n`);

    if (process.platform === 'win32') {
        // Ultimate stealth using VBScript (No .exe needed, completely hidden)
        const vbsPath = path.join(os.tmpdir(), "dardcor_stealth.vbs");
        const batPath = path.join(os.tmpdir(), "dardcor_run.bat");
        
        fs.writeFileSync(batPath, `cd /d "${rootDir}"\ngo run main.go`);
        fs.writeFileSync(vbsPath, `Set WshShell = CreateObject("WScript.Shell")\nWshShell.Run chr(34) & "${batPath}" & Chr(34), 0\nSet WshShell = Nothing`);
        
        exec(`cscript //nologo "${vbsPath}"`);
    } else {
        spawn('go', ['run', 'main.go'], { cwd: rootDir, stdio: 'ignore', shell: false, detached: true }).unref();
    }
    
    for (let i = 0; i < 40; i++) {
        const check = await new Promise(r => {
            const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
            req.setTimeout(400);
        });
        if (check) return true;
        await new Promise(r => setTimeout(r, 1000));
    }
    return false;
  }

  if (!(await ensureEngine())) {
      console.error(`\n  ${C.red}ENGINE FAILURE: Link timeout.${C.reset}`);
      process.exit(1);
  }

  let activeModel = "Claude Opus (Antigravity)";
  const getModel = async () => {
    try {
        const accs = await httpCall(`${gateUrl}/api/antigravity/accounts`);
        if (accs?.length) {
            const act = accs.find(a => a.is_active) || accs[0];
            activeModel = `${act.name} (${act.email})`;
        }
    } catch {}
  };
  await getModel();

  let convId = null, history = [], busy = false, cur = "";

  const drawUI = () => {
    clear();
    const cols = process.stdout.columns, rows = process.stdout.rows;

    const logoY = Math.floor(rows * 0.25);
    const logoX = Math.floor((cols - 65) / 2);
    logo.forEach((line, i) => {
        pos(logoY + i, Math.max(1, logoX));
        process.stdout.write(`${C.purple}${C.bold}${line}${C.reset}`);
    });

    const boxW = Math.min(cols - 10, 80);
    const boxX = Math.floor((cols - boxW) / 2);
    const boxY = logoY + 8;

    // BOX (Opencode Parity)
    for (let i = 0; i < 4; i++) {
        pos(boxY + i, boxX);
        process.stdout.write(`${C.bgInput}${' '.repeat(boxW)}${C.reset}`);
    }
    for (let i = 0; i < 4; i++) {
        pos(boxY + i, boxX);
        process.stdout.write(`${C.cyan}█${C.reset}`);
    }

    // Centered Input Positioning (CRITICAL FIX)
    pos(boxY + 1, boxX + 4);
    if (!cur && !busy) {
        process.stdout.write(`${C.dim}Ask Dardcor...${C.reset}`);
    } else {
        process.stdout.write(`${C.white}${C.bold}${cur}${C.reset}${busy ? C.dim + ' | Thinking...' : ''}`);
    }

    pos(boxY + 3, boxX + 4);
    process.stdout.write(`${C.cyan}${C.bold}Autonomous${C.reset}  ${C.dim}${activeModel}${C.reset}`);

    if (history.length) {
        history.slice(-Math.floor(logoY/2)).forEach((m, i) => {
            pos(2 + (i*2), boxX);
            process.stdout.write(m.role==='user' ? `${C.cyan}❯${C.reset} ${C.white}${m.content.slice(0, boxW-5)}` : `${C.purple}●${C.reset} ${C.dim}${m.content.slice(0, boxW-5)}`);
        });
    }

    pos(rows, 2);
    process.stdout.write(`${C.dim}SYSTEM: ${C.reset}${C.green}CONNECTED${C.reset}  ${C.dim}v1.0.11${C.reset}`);
    pos(boxY + 1, boxX + 4 + cur.length);
  };

  process.stdin.setRawMode(true);
  process.stdin.resume();
  process.stdin.setEncoding('utf8');

  // Enable mouse tracking
  process.stdout.write('\x1b[?1000h'); 
  process.on('exit', () => process.stdout.write('\x1b[?1000l'));

  process.stdin.on('data', async (k) => {
    // Basic mouse click detection to redraw/focus
    if (k.startsWith('\x1b[M')) {
      drawUI();
      return;
    }

    if (busy) return;
    if (k === '\u0003') process.exit();
    if (k === '\r') {
        if (!cur.trim()) return;
        const msg = cur; cur = "";
        history.push({ role: 'user', content: msg });
        busy = true; drawUI();
        try {
            const res = await httpCall(`${gateUrl}/api/agent`, 'POST', JSON.stringify({ 
              message: msg, 
              conversation_id: convId,
              source: "cli"
            }));
            if (res.conversation_id) convId = res.conversation_id;
            history.push({ role: 'agent', content: res.content || 'CORE: No response.' });
        } catch (e) {
            history.push({ role: 'agent', content: 'SIGNAL ERROR.' });
        } finally {
            busy = false; await getModel(); drawUI();
        }
        return;
    }
    if (k === '\u007f' || k === '\b') cur = cur.slice(0, -1);
    else if (k.length === 1 && k.charCodeAt(0) >= 32) cur += k;
    drawUI();
  });

  process.stdout.on('resize', () => drawUI());
  drawUI();
}




