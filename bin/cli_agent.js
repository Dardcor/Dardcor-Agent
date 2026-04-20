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
    white: '\x1b[97m',
    bgDark: '\x1b[48;5;234m',
    divider: '\x1b[90m'
};

const logo = [
    '  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ',
    '  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗',
    '  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝',
    '  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗',
    '  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║',
    '  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝',
    '   D  A  R  D  C  O  R   —   S  U  P  R  E  M  E   A  G  E  N  T  '
];

const gateUrl = "http://127.0.0.1:25000";
const logError = (err) => {
    const logPath = path.join(rootDir, 'debug.log');
    const entry = `[${new Date().toISOString()}] CLI_ERROR: ${err.stack || err}\n`;
    try { fs.appendFileSync(logPath, entry); } catch (e) {}
};


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

const pos = (y, x) => process.stdout.write(`\x1b[${y};${x}H`);
const clearScreen = () => process.stdout.write('\x1b[2J');
const clearLine = () => process.stdout.write('\x1b[2K');

export async function runCLI(initialMode = 'auto') {
    let mode = initialMode; // auto, yolo, safe
    async function ensureEngine() {
        try {
            const checkInitial = await new Promise(r => {
                const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
                req.setTimeout(500);
            });
            if (checkInitial) return true;
        } catch {}

        console.log(`\n  ${C.purple}${C.bold}DARDCOR${C.reset}  ${C.dim}─ Starting Engine...${C.reset}\n`);

        if (process.platform === 'win32') {
            const vbsPath = path.join(os.tmpdir(), "dardcor_stealth.vbs");
            const batPath = path.join(os.tmpdir(), "dardcor_run.bat");
            fs.writeFileSync(batPath, `@echo off\ncd /d "${rootDir}"\ngo run main.go`);
            fs.writeFileSync(vbsPath, `Set WshShell = CreateObject("WScript.Shell")\nWshShell.Run chr(34) & "${batPath}" & Chr(34), 0\nSet WshShell = Nothing`);
            exec(`cscript /nologo "${vbsPath}"`);
        } else {
            spawn('go', ['run', 'main.go'], { cwd: rootDir, stdio: 'ignore', shell: false, detached: true }).unref();
        }

        for (let i = 0; i < 40; i++) {
            const check = await new Promise(r => {
                const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
                req.setTimeout(1000);
            });
            if (check) return true;
            await new Promise(r => setTimeout(r, 800));
        }
        return false;
    }

    if (!(await ensureEngine())) {
        console.error(`\n  ${C.red}ENGINE FAILURE: Link timeout.${C.reset}`);
        process.exit(1);
    }

    let activeModel = "Disconnected";
    const getActiveProviderName = (config) => {
        if (!config) return null;
        const providers = ['antigravity', 'openai', 'anthropic', 'gemini', 'groq', 'deepseek', 'openrouter', 'ollama', 'nvidia'];
        return providers.find(p => config[p] === true);
    };

    const getModel = async () => {
        try {
            const config = await httpCall(`${gateUrl}/api/model/active`);
            const active = getActiveProviderName(config);
            if (active) {
                const modelName = config[`${active}_model`] || 'default';
                activeModel = `${active.toUpperCase()}: ${modelName}`;
                return;
            }
            
            const accs = await httpCall(`${gateUrl}/api/antigravity/accounts`);
            if (accs?.length) {
                const act = accs.find(a => a.is_active) || accs[0];
                activeModel = `ANTIGRAVITY: ${act.name || 'Account'}`;
            } else {
                activeModel = "NO ACTIVE PROVIDER";
            }
        } catch (err) {
            logError(err);
            activeModel = "CONNECTION ERROR (See debug.log)";
        }
    };
    await getModel();

    let convId = null, history = [], busy = false, currentInput = "";
    
    const drawHeader = () => {
        const cols = process.stdout.columns || 80;
        const logoX = Math.max(1, Math.floor((cols - 60) / 2));
        logo.forEach((line, i) => {
            pos(i + 2, logoX);
            process.stdout.write(`${C.purple}${C.bold}${line}${C.reset}`);
        });
        pos(9, 1);
        process.stdout.write(`${C.divider}━`.repeat(cols) + C.reset);
    };

    const drawStatus = () => {
        const cols = process.stdout.columns || 80;
        const rows = process.stdout.rows || 24;
        pos(rows - 1, 1);
        process.stdout.write(`${C.divider}━`.repeat(cols) + C.reset);
        pos(rows, 2);
        
        const modeColor = mode === 'yolo' ? C.yellow : (mode === 'safe' ? C.green : C.cyan);
        const modeDisplay = `${modeColor}${C.bold}${mode.toUpperCase()}${C.reset}`;
        process.stdout.write(`${C.bold}${C.purple}DARDCOR${C.reset} ${C.dim}v1.0.11${C.reset}  ${C.divider}│${C.reset}  ${C.bold}${C.cyan}MODEL:${C.reset} ${C.dim}${activeModel.substring(0, Math.floor(cols * 0.4))}${C.reset}  ${C.divider}│${C.reset}  ${modeDisplay}  ${C.divider}│${C.reset}  ${C.green}● ONLINE${C.reset}`);
    }

    const drawInput = (isFocus = true) => {
        process.stdout.write('\x1b[?25l');
        const rows = process.stdout.rows || 24;
        const cols = process.stdout.columns || 80;
        pos(rows - 3, 1);
        process.stdout.write(`${C.divider}━`.repeat(cols) + C.reset);
        pos(rows - 2, 1);
        clearLine();
        pos(rows - 2, 2);
        if (busy) {
            process.stdout.write(`${C.bold}${C.yellow}●${C.reset} ${C.white}Thinking...${C.reset}`);
        } else {
            process.stdout.write(`${C.bold}${C.cyan}❯${C.reset} ${C.white}${currentInput}${C.reset}${currentInput ? "" : C.dim + "Ask anything..." + C.reset}`);
            if (isFocus) pos(rows - 2, currentInput.length + 4);
        }
        process.stdout.write('\x1b[?25h');
    };

    const drawHistory = () => {
        const rows = process.stdout.rows || 24;
        const cols = process.stdout.columns || 80;
        const chatAreaHeight = rows - 13;
        const maxWidth = cols - 12;
        
        for (let i = 10; i < rows - 3; i++) {
            pos(i, 1); clearLine();
        }

        const wrap = (text, width) => {
            const lines = [];
            // Filter out think tags content for UI
            const cleanText = text.replace(/<think>[\s\S]*?<\/think>/g, '').trim();
            if (!cleanText) return [];

            cleanText.split('\n').filter(p => p.trim()).forEach(p => {
                let current = p;
                while (current.length > width) {
                    let spaceIdx = current.lastIndexOf(' ', width);
                    if (spaceIdx === -1) spaceIdx = width;
                    lines.push(current.substring(0, spaceIdx));
                    current = current.substring(spaceIdx).trim();
                }
                if (current) lines.push(current);
            });
            return lines;
        };

        let allLines = [];
        history.forEach(m => {
            const name = m.role === 'user' ? `${C.cyan}${C.bold}YOU${C.reset}` : `${C.purple}${C.bold}DARDCOR${C.reset}`;
            const wrappedContent = wrap(m.content, maxWidth);
            if (wrappedContent.length > 0) {
                allLines.push({ type: 'first', icon: name, text: wrappedContent[0] });
                for (let i = 1; i < wrappedContent.length; i++) {
                    allLines.push({ type: 'extra', text: wrappedContent[i] });
                }
                allLines.push({ type: 'spacer' });
            }
        });

        const toRender = allLines.slice(-chatAreaHeight);
        let currentY = rows - 4 - toRender.length + 1;
        if (currentY < 10) currentY = 10;

        toRender.forEach(line => {
            if (line.type === 'spacer') { currentY++; return; }
            pos(currentY, 4);
            if (line.type === 'first') {
                process.stdout.write(`${line.icon} ${C.divider}─${C.reset} ${C.white}${line.text}${C.reset}`);
            } else {
                process.stdout.write(`    ${C.divider}│${C.reset}  ${C.white}${line.text}${C.reset}`);
            }
            currentY++;
        });
    };

    const refreshAll = (isFull = false) => {
        process.stdout.write('\x1b[?25l');
        if (isFull) {
            clearScreen();
            drawHeader();
        }
        drawHistory();
        drawStatus();
        drawInput(true);
        process.stdout.write('\x1b[?25h');
    };

    const enterAltBuffer = () => process.stdout.write('\x1b[?1049h');
    const exitAltBuffer = () => process.stdout.write('\x1b[?1049l');

    const cleanup = () => {
        exitAltBuffer();
        process.stdout.write('\x1b[?25h'); 
        process.stdin.setRawMode(false);
        process.stdin.pause();
    };

    process.stdin.setRawMode(true);
    process.stdin.resume();
    process.stdin.setEncoding('utf8');

    process.on('SIGINT', () => { cleanup(); process.exit(0); });
    process.on('SIGTERM', () => { cleanup(); process.exit(0); });
    process.on('exit', () => {
        try {
            process.stdout.write('\x1b[?1049l'); // exit alt buffer
            process.stdout.write('\x1b[?25h');   // show cursor
            if (process.stdin.setRawMode) process.stdin.setRawMode(false);
        } catch {}
    });

    enterAltBuffer();
    refreshAll(true);
    
    let ctrlCCount = 0;
    let ctrlCTimer = null;

    process.stdin.on('data', async (k) => {
        if (k === '\u0003') {
            ctrlCCount++;
            if (ctrlCTimer) clearTimeout(ctrlCTimer);
            if (ctrlCCount >= 2) {
                cleanup();
                process.exit(0);
            }
            // Show hint on first press
            const rows = process.stdout.rows || 24;
            const cols = process.stdout.columns || 80;
            pos(rows - 2, 1);
            clearLine();
            pos(rows - 2, 2);
            process.stdout.write(`${C.bold}${C.yellow}●${C.reset} ${C.dim}Press Ctrl+C again to exit${C.reset}`);
            ctrlCTimer = setTimeout(() => {
                ctrlCCount = 0;
                drawInput();
            }, 2000);
            return;
        }
        ctrlCCount = 0;
        if (ctrlCTimer) { clearTimeout(ctrlCTimer); ctrlCTimer = null; }
        if (busy) return;

        if (k === '\r' || k === '\n') {
            if (!currentInput.trim()) return;
            const msg = currentInput.trim(); 
            currentInput = "";
            
            // ─── Slash Commands ─────────────────────────────────────────────
            if (msg.startsWith('/')) {
                const parts = msg.slice(1).split(' ');
                const cmd = parts[0].toLowerCase();
                
                if (cmd === 'exit' || cmd === 'quit') {
                    cleanup(); process.exit();
                }
                
                if (cmd === 'clear' || cmd === 'cls') {
                    history = []; convId = null;
                    refreshAll();
                    return;
                }
                
                if (cmd === 'help') {
                    history.push({ role: 'agent', content: "AVAILABLE COMMANDS:\n/model - Show active AI provider/model\n/mode  - Switch mode (auto, yolo, safe)\n/clear - Reset conversation history\n/exit  - Close application\n/help  - Show this message" });
                    refreshAll();
                    return;
                }

                if (cmd === 'mode') {
                    const newMode = parts[1]?.toLowerCase();
                    if (['auto', 'yolo', 'safe'].includes(newMode)) {
                        mode = newMode;
                        history.push({ role: 'agent', content: `Execution mode switched to: ${mode.toUpperCase()}` });
                    } else {
                        history.push({ role: 'agent', content: `Current mode: ${mode.toUpperCase()}\nUsage: /mode [auto | yolo | safe]` });
                    }
                    refreshAll();
                    return;
                }
                
                if (cmd === 'model') {
                    try {
                        const cfg = await httpCall(`${gateUrl}/api/model/active`);
                        const active = getActiveProviderName(cfg);
                        const modelName = active ? (cfg[active + '_model'] || 'default') : null;
                        const info = active
                            ? `Provider : ${active.toUpperCase()}\nModel    : ${modelName}`
                            : "No active provider configured.\nUse the dashboard to set one.";
                        history.push({ role: 'agent', content: info });
                    } catch {
                        history.push({ role: 'agent', content: "Failed to fetch model info." });
                    }
                    refreshAll();
                    return;
                }
                
                history.push({ role: 'agent', content: `Unknown command: /${cmd}. Type /help for list.` });
                refreshAll();
                return;
            }

            history.push({ role: 'user', content: msg });
            busy = true; 
            refreshAll();

            try {
                const res = await httpCall(`${gateUrl}/api/agent`, 'POST', JSON.stringify({
                    message: msg,
                    conversation_id: convId,
                    source: "cli"
                }));
                if (res.conversation_id) convId = res.conversation_id;
                history.push({ role: 'agent', content: res.content || 'No response.' });
            } catch (e) {
                history.push({ role: 'agent', content: 'Connection Error.' });
            } finally {
                busy = false; 
                await getModel();
                refreshAll();
            }
            return;
        }

        if (k === '\u007f' || k === '\b') {
            currentInput = currentInput.slice(0, -1);
            drawInput(); 
            return;
        } else if (k.length === 1 && k.charCodeAt(0) >= 32) {
            currentInput += k;
            drawInput(); 
            return;
        }
        
        if (k.length > 1) refreshAll();
    });

    process.stdout.on('resize', () => refreshAll(true));
}
