import { spawn, execSync, exec } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';
import http from 'http';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.join(__dirname, '..');

const C = {
    reset: '\x1b[0m',
    bold: '\x1b[1m',
    dim: '\x1b[2m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    cyan: '\x1b[36m',
    purple: '\x1b[38;5;93m',
    bgInput: '\x1b[48;5;234m',
    white: '\x1b[37m'
};

const logo = [
    '  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ',
    '  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗',
    '  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝',
    '  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗',
    '  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║',
    '  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝'
];

const pos = (y, x) => process.stdout.write(`\x1b[${y};${x}H`);
const clear = () => process.stdout.write('\x1b[2J\x1b[H');

const gateUrl = "http://127.0.0.1:25000";

async function httpCall(url, method = 'GET', body = null) {
    return new Promise((resolve, reject) => {
        const u = new URL(url);
        const req = http.request({
            hostname: u.hostname,
            port: u.port,
            path: u.pathname + u.search,
            method,
            headers: body ? { 'Content-Type': 'application/json' } : {}
        }, res => {
            let d = '';
            res.on('data', chunk => d += chunk);
            res.on('end', () => {
                try {
                    const j = JSON.parse(d);
                    resolve(j.data || j);
                } catch { resolve(d); }
            });
        });
        req.on('error', reject);
        if (body) req.write(body);
        req.end();
    });
}

export async function runCLI() {
    async function ensureEngine() {
        try {
            const checkInitial = await new Promise(r => {
                const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
                req.setTimeout(500);
            });
            if (checkInitial) return true;
        } catch {}

        clear();
        console.log(`\n  ${C.purple}${C.bold}DARDCOR${C.reset}  ${C.dim}─ Initializing Source Engine...${C.reset}\n`);

        if (process.platform === 'win32') {
            const vbsPath = path.join(os.tmpdir(), "dardcor_stealth.vbs");
            const batPath = path.join(os.tmpdir(), "dardcor_run.bat");

            fs.writeFileSync(batPath, `@echo off\ncd /d "${rootDir}"\ngo run main.go`);
            fs.writeFileSync(vbsPath, `Set WshShell = CreateObject("WScript.Shell")\nWshShell.Run chr(34) & "${batPath}" & Chr(34), 0\nSet WshShell = Nothing`);

            const nologo = "/nologo";
            exec(`cscript ${nologo} "${vbsPath}"`);
        } else {
            spawn('go', ['run', 'main.go'], { cwd: rootDir, stdio: 'ignore', shell: false, detached: true }).unref();
        }

        for (let i = 0; i < 40; i++) {
            const check = await new Promise(r => {
                const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
                req.setTimeout(1000);
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
                activeModel = `${act.name || 'Agent'} (${act.email})`;
            }
        } catch {}
    };
    await getModel();

    let convId = null, history = [], busy = false, cur = "";

    const drawUI = () => {
        clear();
        const cols = process.stdout.columns || 80, rows = process.stdout.rows || 24;

        const logoY = Math.max(1, Math.floor(rows * 0.15));
        const logoX = Math.max(1, Math.floor((cols - 65) / 2));
        logo.forEach((line, i) => {
            pos(logoY + i, logoX);
            process.stdout.write(`${C.purple}${C.bold}${line}${C.reset}`);
        });

        const boxW = Math.min(cols - 10, 80);
        const boxX = Math.floor((cols - boxW) / 2);
        const boxY = logoY + 8;

        for (let i = 0; i < 4; i++) {
            pos(boxY + i, boxX);
            process.stdout.write(`${C.bgInput}${' '.repeat(boxW)}${C.reset}`);
        }
        for (let i = 0; i < 4; i++) {
            pos(boxY + i, boxX);
            process.stdout.write(`${C.cyan}█${C.reset}`);
        }

        pos(boxY + 1, boxX + 4);
        if (!cur && !busy) {
            process.stdout.write(`${C.dim}Ask Dardcor...${C.reset}`);
        } else {
            const displayCur = cur.slice(-(boxW - 10));
            process.stdout.write(`${C.white}${C.bold}${displayCur}${C.reset}${busy ? C.dim + ' | Thinking...' : ''}`);
        }

        pos(boxY + 3, boxX + 4);
        process.stdout.write(`${C.cyan}${C.bold}Autonomous${C.reset}  ${C.dim}${activeModel}${C.reset}`);

        if (history.length) {
            const startHistoryY = 2;
            const historyToDisplay = history.slice(-Math.floor(logoY));
            historyToDisplay.forEach((m, i) => {
                pos(startHistoryY + (i), boxX);
                const color = m.role === 'user' ? C.cyan : C.purple;
                const icon = m.role === 'user' ? '❯' : '●';
                process.stdout.write(`${color}${icon}${C.reset} ${C.white}${m.content.slice(0, boxW - 5)}`);
            });
        }

        pos(rows, 2);
        process.stdout.write(`${C.dim}SYSTEM: ${C.reset}${C.green}CONNECTED${C.reset}`);
        pos(boxY + 1, Math.min(boxX + 4 + cur.length, boxX + boxW - 2));
    };

    process.stdin.setRawMode(true);
    process.stdin.resume();
    process.stdin.setEncoding('utf8');

    process.stdout.write('\x1b[?1000h');
    process.on('exit', () => process.stdout.write('\x1b[?1000l'));

    process.stdin.on('data', async (k) => {
        if (k.startsWith('\x1b[M')) {
            drawUI();
            return;
        }

        if (k === '\u0003') process.exit();
        if (busy) return;
        
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
        if (k === '\u007f' || k === '\b') {
            cur = cur.slice(0, -1);
        } else if (k.length === 1 && k.charCodeAt(0) >= 32) {
            cur += k;
        }
        drawUI();
    });

    process.stdout.on('resize', () => drawUI());
    drawUI();
}
