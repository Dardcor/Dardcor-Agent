import { spawn, exec } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';
import http from 'http';
import net from 'net';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.join(__dirname, '..');

const C = {
    reset:    '\x1b[0m',
    bold:     '\x1b[1m',
    dim:      '\x1b[2m',
    italic:   '\x1b[3m',
    underline:'\x1b[4m',
    red:      '\x1b[31m',
    green:    '\x1b[32m',
    yellow:   '\x1b[33m',
    cyan:     '\x1b[36m',
    white:    '\x1b[97m',
    gray:     '\x1b[90m',
    purple:   '\x1b[38;5;93m',
    lavender: '\x1b[38;5;183m',
    bgPurple: '\x1b[48;5;54m',
    bgDark:   '\x1b[48;5;232m',
    bgBar:    '\x1b[48;5;235m',
    bgInput:  '\x1b[48;5;234m',
};

const LOGO = [
    '  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ',
    '  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗',
    '  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝',
    '  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗',
    '  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║',
    '  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝',
];

const LOGO_WIDTH = 65;

const gateUrl = 'http://127.0.0.1:25000';
const wsHost  = '127.0.0.1';
const wsPort  = 25000;
const wsPath  = '/ws';

const out = (s) => process.stdout.write(s);
const moveTo = (row, col) => out(`\x1b[${row};${col}H`);
const clearScreen = () => out('\x1b[2J\x1b[H');
const clearLine = () => out('\x1b[2K');
const saveCursor = () => out('\x1b[s');
const restoreCursor = () => out('\x1b[u');
const hideCursor = () => out('\x1b[?25l');
const showCursor = () => out('\x1b[?25h');
const altScreen = () => out('\x1b[?1049h');
const normalScreen = () => out('\x1b[?1049l');

function strWidth(s) {
    return s.replace(/\x1b\[[0-9;]*m/g, '').length;
}

function padRight(s, w) {
    const vis = strWidth(s);
    return vis < w ? s + ' '.repeat(w - vis) : s;
}

function wrapText(text, maxWidth) {
    const lines = [];
    const paragraphs = text.split('\n');
    for (const para of paragraphs) {
        if (para.length === 0) { lines.push(''); continue; }
        let remaining = para;
        while (remaining.length > 0) {
            if (remaining.length <= maxWidth) { lines.push(remaining); break; }
            let cut = maxWidth;
            const spaceIdx = remaining.lastIndexOf(' ', maxWidth);
            if (spaceIdx > 0) cut = spaceIdx + 1;
            lines.push(remaining.slice(0, cut).trimEnd());
            remaining = remaining.slice(cut);
        }
    }
    return lines;
}

function renderMarkdown(text, maxWidth) {
    const rawLines = text.split('\n');
    const result = [];
    let inCode = false;
    let codeLang = '';

    for (let i = 0; i < rawLines.length; i++) {
        const line = rawLines[i];

        if (line.startsWith('```')) {
            inCode = !inCode;
            if (inCode) {
                codeLang = line.slice(3).trim();
                const label = codeLang ? ` ${codeLang} ` : ' code ';
                result.push(`${C.bgBar}${C.lavender}${C.bold}╭─${label}${'─'.repeat(Math.max(0, maxWidth - label.length - 3))}╮${C.reset}`);
            } else {
                result.push(`${C.bgBar}${C.lavender}╰${'─'.repeat(maxWidth - 1)}╯${C.reset}`);
                codeLang = '';
            }
            continue;
        }

        if (inCode) {
            const padded = padRight(line, maxWidth - 2);
            result.push(`${C.bgBar}${C.lavender}│${C.reset}${C.bgDark} ${C.white}${padded}${C.reset}${C.bgBar}${C.lavender}│${C.reset}`);
            continue;
        }

        if (line.startsWith('### ')) {
            result.push(`${C.cyan}${C.bold}${line.slice(4)}${C.reset}`);
            continue;
        }
        if (line.startsWith('## ')) {
            result.push(`${C.purple}${C.bold}${line.slice(3)}${C.reset}`);
            continue;
        }
        if (line.startsWith('# ')) {
            result.push(`${C.lavender}${C.bold}${line.slice(2)}${C.reset}`);
            continue;
        }

        if (line.startsWith('- ') || line.startsWith('* ')) {
            const content = formatInline(line.slice(2), maxWidth - 4);
            result.push(`  ${C.purple}▸${C.reset} ${content}`);
            continue;
        }

        const numMatch = line.match(/^(\d+)\. (.+)/);
        if (numMatch) {
            const content = formatInline(numMatch[2], maxWidth - 5);
            result.push(`  ${C.cyan}${numMatch[1]}.${C.reset} ${content}`);
            continue;
        }

        if (line.startsWith('> ')) {
            result.push(`${C.gray}│${C.reset} ${C.italic}${C.dim}${line.slice(2)}${C.reset}`);
            continue;
        }

        if (line.match(/^[-─═]{3,}$/)) {
            result.push(`${C.gray}${'─'.repeat(maxWidth)}${C.reset}`);
            continue;
        }

        if (line.trim() === '') {
            result.push('');
            continue;
        }

        const wrapped = wrapText(formatInline(line, maxWidth), maxWidth);
        for (const wl of wrapped) result.push(wl);
    }

    return result;
}

function formatInline(text, maxWidth) {
    let s = text;
    s = s.replace(/`([^`]+)`/g, `${C.bgDark}${C.lavender}$1${C.reset}`);
    s = s.replace(/\*\*([^*]+)\*\*/g, `${C.bold}$1${C.reset}`);
    s = s.replace(/\*([^*]+)\*/g, `${C.italic}$1${C.reset}`);
    s = s.replace(/__([^_]+)__/g, `${C.bold}$1${C.reset}`);
    s = s.replace(/_([^_]+)_/g, `${C.italic}$1${C.reset}`);
    return s;
}

function httpGet(url) {
    return new Promise((resolve, reject) => {
        const u = new URL(url);
        const req = http.get({ hostname: u.hostname, port: u.port, path: u.pathname + u.search }, res => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)); } catch { resolve({}); }
            });
        });
        req.on('error', reject);
        req.setTimeout(2000, () => { req.destroy(); reject(new Error('timeout')); });
    });
}

function httpPost(url, body) {
    return new Promise((resolve, reject) => {
        const u = new URL(url);
        const data = JSON.stringify(body);
        const req = http.request({
            hostname: u.hostname, port: u.port, path: u.pathname + u.search,
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) }
        }, res => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)); } catch { resolve({}); }
            });
        });
        req.on('error', reject);
        req.setTimeout(5000, () => { req.destroy(); reject(new Error('timeout')); });
        req.write(data);
        req.end();
    });
}

class NativeWebSocket {
    constructor(host, port, path) {
        this._host = host;
        this._port = port;
        this._path = path;
        this._socket = null;
        this._connected = false;
        this._buffer = Buffer.alloc(0);
        this._key = Buffer.from(Math.random().toString(36)).toString('base64');
        this.onopen = null;
        this.onmessage = null;
        this.onclose = null;
        this.onerror = null;
    }

    connect() {
        this._socket = net.createConnection({ host: this._host, port: this._port });
        this._socket.on('connect', () => {
            const handshake = [
                `GET ${this._path} HTTP/1.1`,
                `Host: ${this._host}:${this._port}`,
                `Upgrade: websocket`,
                `Connection: Upgrade`,
                `Sec-WebSocket-Key: ${this._key}`,
                `Sec-WebSocket-Version: 13`,
                `\r\n`
            ].join('\r\n');
            this._socket.write(handshake);
        });

        this._socket.on('data', (chunk) => {
            this._buffer = Buffer.concat([this._buffer, chunk]);
            if (!this._connected) {
                const headerEnd = this._buffer.indexOf('\r\n\r\n');
                if (headerEnd !== -1) {
                    const header = this._buffer.slice(0, headerEnd).toString();
                    if (header.includes('101')) {
                        this._connected = true;
                        this._buffer = this._buffer.slice(headerEnd + 4);
                        if (this.onopen) this.onopen();
                    }
                    if (this._buffer.length > 0) this._processFrames();
                }
            } else {
                this._processFrames();
            }
        });

        this._socket.on('close', () => {
            this._connected = false;
            if (this.onclose) this.onclose();
        });

        this._socket.on('error', (err) => {
            if (this.onerror) this.onerror(err);
        });
    }

    _processFrames() {
        while (this._buffer.length >= 2) {
            const b0 = this._buffer[0];
            const b1 = this._buffer[1];
            const opcode = b0 & 0x0F;
            const masked = (b1 & 0x80) !== 0;
            let payloadLen = b1 & 0x7F;
            let offset = 2;

            if (payloadLen === 126) {
                if (this._buffer.length < offset + 2) break;
                payloadLen = this._buffer.readUInt16BE(offset);
                offset += 2;
            } else if (payloadLen === 127) {
                if (this._buffer.length < offset + 8) break;
                payloadLen = Number(this._buffer.readBigUInt64BE(offset));
                offset += 8;
            }

            const maskLen = masked ? 4 : 0;
            if (this._buffer.length < offset + maskLen + payloadLen) break;

            let mask;
            if (masked) { mask = this._buffer.slice(offset, offset + 4); offset += 4; }

            const payload = this._buffer.slice(offset, offset + payloadLen);
            offset += payloadLen;

            if (masked) {
                for (let i = 0; i < payload.length; i++) payload[i] ^= mask[i % 4];
            }

            this._buffer = this._buffer.slice(offset);

            if (opcode === 8) { this._socket.destroy(); break; }
            if (opcode === 9) { this._sendPong(payload); continue; }
            if (opcode === 1 || opcode === 2) {
                const msg = payload.toString('utf8');
                if (this.onmessage) this.onmessage(msg);
            }
        }
    }

    _sendPong(payload) {
        if (!this._socket || !this._connected) return;
        const frame = Buffer.alloc(2 + payload.length);
        frame[0] = 0x8A;
        frame[1] = payload.length;
        payload.copy(frame, 2);
        this._socket.write(frame);
    }

    send(data) {
        if (!this._socket || !this._connected) return;
        const payload = Buffer.from(data, 'utf8');
        const mask = Buffer.from([
            Math.floor(Math.random() * 256),
            Math.floor(Math.random() * 256),
            Math.floor(Math.random() * 256),
            Math.floor(Math.random() * 256)
        ]);
        const len = payload.length;
        let headerLen = 2;
        if (len > 65535) headerLen += 8;
        else if (len > 125) headerLen += 2;
        headerLen += 4;

        const frame = Buffer.alloc(headerLen + len);
        frame[0] = 0x81;
        let offset = 1;
        if (len <= 125) { frame[offset++] = 0x80 | len; }
        else if (len <= 65535) { frame[offset++] = 0x80 | 126; frame.writeUInt16BE(len, offset); offset += 2; }
        else { frame[offset++] = 0x80 | 127; frame.writeBigUInt64BE(BigInt(len), offset); offset += 8; }

        mask.copy(frame, offset); offset += 4;
        for (let i = 0; i < len; i++) frame[offset + i] = payload[i] ^ mask[i % 4];
        this._socket.write(frame);
    }

    close() {
        if (this._socket) {
            try {
                const frame = Buffer.from([0x88, 0x80, 0x00, 0x00, 0x00, 0x00]);
                this._socket.write(frame);
                this._socket.destroy();
            } catch {}
        }
    }
}

const SPINNERS = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];

export async function runCLI() {
    async function ensureEngine() {
        try {
            const ok = await new Promise(r => {
                const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
                req.setTimeout(500);
            });
            if (ok) return true;
        } catch {}

        normalScreen();
        showCursor();
        process.stdout.write(`\n  ${C.purple}${C.bold}DARDCOR${C.reset}  ${C.dim}─ Initializing Source Engine...${C.reset}\n\n`);

        if (process.platform === 'win32') {
            const vbsPath = path.join(os.tmpdir(), 'dardcor_stealth.vbs');
            const batPath = path.join(os.tmpdir(), 'dardcor_run.bat');
            fs.writeFileSync(batPath, `@echo off\ncd /d "${rootDir}"\ngo run main.go`);
            fs.writeFileSync(vbsPath, `Set WshShell = CreateObject("WScript.Shell")\nWshShell.Run chr(34) & "${batPath}" & Chr(34), 0\nSet WshShell = Nothing`);
            exec(`cscript /nologo "${vbsPath}"`);
        } else {
            spawn('go', ['run', 'main.go'], { cwd: rootDir, stdio: 'ignore', shell: false, detached: true }).unref();
        }

        process.stdout.write(`  ${C.dim}Waiting for engine`);
        for (let i = 0; i < 40; i++) {
            const ok = await new Promise(r => {
                const req = http.get(gateUrl + '/api/system', res => r(res.statusCode === 200)).on('error', () => r(false));
                req.setTimeout(1000);
            });
            if (ok) { process.stdout.write(`${C.reset}\n`); return true; }
            process.stdout.write('.');
            await new Promise(r => setTimeout(r, 1000));
        }
        process.stdout.write(`${C.reset}\n`);
        return false;
    }

    if (!(await ensureEngine())) {
        process.stderr.write(`\n  ${C.red}${C.bold}ENGINE FAILURE${C.reset}${C.red}: Link timeout. Run \`dardcor run\` first.${C.reset}\n\n`);
        process.exit(1);
    }

    altScreen();
    hideCursor();

    let activeModel = 'Dardcor Agent';
    let activeProvider = 'Antigravity';
    let connStatus = 'CONNECTING';
    let sessionId = null;
    let sessionLabel = 'new';
    let sessions = [];

    const messages = [];
    let inputLines = [''];
    let inputCursorLine = 0;
    let inputCursorCol = 0;
    let scrollOffset = 0;
    let busy = false;
    let streamBuffer = '';
    let spinnerFrame = 0;
    let spinnerInterval = null;
    let actionLines = [];
    let actionTimer = null;
    let wsInstance = null;
    let pendingMsgIdx = -1;
    let toolProgress = [];

    const dims = () => ({ cols: process.stdout.columns || 80, rows: process.stdout.rows || 24 });

    function getContentWidth(cols) { return Math.max(20, cols - 4); }
    function getHistoryHeight(rows) { return Math.max(3, rows - LOGO.length - 3 - 4 - 3); }
    function getInputHeight() { return Math.min(inputLines.length, 4) + 2; }

    async function fetchModel() {
        try {
            const sys = await httpGet(`${gateUrl}/api/system`);
            if (sys?.provider) activeProvider = sys.provider;
            if (sys?.model) activeModel = sys.model;
            const accs = await httpGet(`${gateUrl}/api/antigravity/accounts`);
            if (Array.isArray(accs) && accs.length) {
                const act = accs.find(a => a.is_active) || accs[0];
                activeModel = act.name || act.email || activeModel;
            }
        } catch {}
    }

    async function fetchSessions() {
        try {
            const data = await httpGet(`${gateUrl}/api/sessions`);
            if (Array.isArray(data)) sessions = data;
            else if (Array.isArray(data?.data)) sessions = data.data;
        } catch {}
    }

    await fetchModel();
    await fetchSessions();

    function renderBanner(cols) {
        const xOff = Math.max(1, Math.floor((cols - LOGO_WIDTH) / 2));
        const lines = [];
        for (const l of LOGO) {
            lines.push(' '.repeat(xOff - 1) + C.purple + C.bold + l + C.reset);
        }
        const sub = 'DARDCOR CLI  ─  Terminal Interface';
        const subX = Math.max(0, Math.floor((cols - sub.length) / 2));
        lines.push(' '.repeat(subX) + C.dim + sub + C.reset);
        return lines;
    }

    function renderSeparator(cols, char = '─') {
        return C.gray + char.repeat(cols) + C.reset;
    }

    function renderMessage(msg, cols) {
        const cw = getContentWidth(cols);
        const result = [];

        if (msg.role === 'user') {
            const prefix = `${C.cyan}${C.bold}❯ You${C.reset}`;
            result.push(prefix);
            const wrapped = wrapText(msg.content, cw - 2);
            for (const wl of wrapped) result.push(`  ${C.white}${wl}${C.reset}`);
        } else if (msg.role === 'agent') {
            const prefix = `${C.purple}${C.bold}● Dardcor${C.reset}`;
            result.push(prefix);
            if (msg.streaming && msg.content === '') {
                result.push(`  ${C.dim}${SPINNERS[spinnerFrame]} Generating...${C.reset}`);
            } else {
                const rendered = renderMarkdown(msg.content, cw - 2);
                for (const l of rendered) result.push(`  ${l}`);
                if (msg.streaming) result.push(`  ${C.dim}${SPINNERS[spinnerFrame]}${C.reset}`);
            }
        } else if (msg.role === 'tool') {
            const icon = msg.done ? `${C.green}✅` : `${C.yellow}${SPINNERS[spinnerFrame]}`;
            const timing = msg.ms ? `${C.gray} (${msg.ms}ms)${C.reset}` : '';
            result.push(`  ${icon}${C.reset} ${C.dim}${C.yellow}⚡ ${msg.content}${C.reset}${timing}`);
        } else if (msg.role === 'system') {
            result.push(`  ${C.gray}${C.italic}${msg.content}${C.reset}`);
        }

        result.push('');
        return result;
    }

    function renderHistory(startRow, histRows, cols) {
        const cw = getContentWidth(cols);
        const allLines = [];

        for (const msg of messages) {
            const rendered = renderMessage(msg, cols);
            for (const l of rendered) allLines.push(l);
        }

        const totalLines = allLines.length;
        const maxScroll = Math.max(0, totalLines - histRows);
        if (scrollOffset > maxScroll) scrollOffset = maxScroll;
        const visibleStart = scrollOffset === 0 ? maxScroll : maxScroll - scrollOffset;
        const start = Math.max(0, visibleStart);
        const visible = allLines.slice(start, start + histRows);

        for (let i = 0; i < histRows; i++) {
            moveTo(startRow + i, 1);
            clearLine();
            if (i < visible.length) {
                const line = visible[i];
                const truncated = strWidth(line) > cols - 2 ? line.slice(0, cols - 2) : line;
                out(' ' + truncated);
            }
        }
    }

    function renderInputArea(startRow, cols) {
        const inputH = getInputHeight();
        const cw = cols - 6;

        moveTo(startRow, 1);
        clearLine();
        out(`${C.bgInput}${C.purple}${C.bold} ╭${'─'.repeat(cols - 4)}╮ ${C.reset}`);

        for (let i = 0; i < inputH - 2; i++) {
            moveTo(startRow + 1 + i, 1);
            clearLine();
            const lineContent = inputLines[i] || '';
            const prefix = i === 0 ? `${C.cyan}❯ ${C.reset}` : `  `;
            const display = lineContent.length > cw ? lineContent.slice(-cw) : lineContent;
            out(`${C.bgInput} ${prefix}${C.white}${display}${C.reset}${' '.repeat(Math.max(0, cols - strWidth(display) - 5))}${C.bgInput} ${C.reset}`);
        }

        moveTo(startRow + inputH - 2, 1);
        clearLine();
        const hints = busy
            ? `${C.dim}${C.yellow}${SPINNERS[spinnerFrame]} Processing...${C.reset}`
            : `${C.dim}Enter to send  Shift+Enter newline  Ctrl+N new session  Ctrl+L clear  Ctrl+C exit${C.reset}`;
        out(`${C.bgInput} ${hints}${' '.repeat(Math.max(0, cols - strWidth(hints.replace(/\x1b\[[0-9;]*m/g, '')) - 3))} ${C.reset}`);

        moveTo(startRow + inputH - 1, 1);
        clearLine();
        out(`${C.bgInput}${C.purple}${C.bold} ╰${'─'.repeat(cols - 4)}╯ ${C.reset}`);
    }

    function renderStatusBar(row, cols) {
        moveTo(row, 1);
        clearLine();

        const connIcon = connStatus === 'CONNECTED' ? `${C.green}●${C.reset}` : connStatus === 'CONNECTING' ? `${C.yellow}●${C.reset}` : `${C.red}●${C.reset}`;
        const connText = ` ${connIcon} ${connStatus === 'CONNECTED' ? C.green : connStatus === 'CONNECTING' ? C.yellow : C.red}${connStatus}${C.reset}`;
        const modelText = `${C.dim} │ ${C.reset}${C.lavender}${activeModel}${C.reset}`;
        const sessText = `${C.dim} │ ${C.reset}${C.cyan}Session: ${sessionLabel}${C.reset}`;
        const scrollText = scrollOffset > 0 ? `${C.dim} │ ↑ scrolled ${C.reset}` : '';
        const left = `${C.bgBar}${connText}${modelText}${sessText}${scrollText}`;
        const right = `${C.gray} DARDCOR CLI ${C.reset}`;
        const leftVis = strWidth(connText.replace(/\x1b\[[0-9;]*m/g, '') + modelText.replace(/\x1b\[[0-9;]*m/g, '') + sessText.replace(/\x1b\[[0-9;]*m/g, '') + scrollText.replace(/\x1b\[[0-9;]*m/g, ''));
        const gap = Math.max(0, cols - leftVis - 13);

        out(`${C.bgBar}${left}${' '.repeat(gap)}${right}${C.reset}`);
    }

    function redraw() {
        const { cols, rows } = dims();
        const banner = renderBanner(cols);
        const bannerH = banner.length;
        const statusBarRow = rows;
        const inputH = getInputHeight();
        const inputStart = statusBarRow - inputH - 1;
        const sepRow1 = bannerH + 1;
        const sepRow2 = inputStart;
        const histStart = bannerH + 2;
        const histRows = Math.max(1, sepRow2 - histStart);

        out('\x1b[?25l');

        for (let i = 0; i < bannerH; i++) {
            moveTo(i + 1, 1);
            clearLine();
            out(banner[i]);
        }

        moveTo(sepRow1, 1);
        clearLine();
        out(renderSeparator(cols));

        renderHistory(histStart, histRows, cols);

        moveTo(sepRow2, 1);
        clearLine();
        out(renderSeparator(cols));

        renderInputArea(sepRow2 + 1, cols);
        renderStatusBar(statusBarRow, cols);

        const cursorLine = inputCursorLine;
        const cursorCol = inputCursorCol;
        const inputContentRow = sepRow2 + 2 + cursorLine;
        const inputContentCol = Math.min(cursorCol, getContentWidth(cols)) + 4;
        moveTo(inputContentRow, inputContentCol);
        out('\x1b[?25h');
    }

    function startSpinner() {
        if (spinnerInterval) return;
        spinnerInterval = setInterval(() => {
            spinnerFrame = (spinnerFrame + 1) % SPINNERS.length;
            redraw();
        }, 80);
    }

    function stopSpinner() {
        if (spinnerInterval) { clearInterval(spinnerInterval); spinnerInterval = null; }
    }

    function addMessage(role, content, extra = {}) {
        messages.push({ role, content, ...extra });
        scrollOffset = 0;
        redraw();
    }

    function connectWebSocket() {
        const ws = new NativeWebSocket(wsHost, wsPort, wsPath);
        wsInstance = ws;

        ws.onopen = () => {
            connStatus = 'CONNECTED';
            redraw();
        };

        ws.onmessage = (raw) => {
            let data;
            try { data = JSON.parse(raw); } catch { return; }

            const type = data.type;
            const payload = data.payload || data;

            if (type === 'connected') {
                connStatus = 'CONNECTED';
                redraw();
                return;
            }

            if (type === 'typing') {
                if (pendingMsgIdx === -1) {
                    messages.push({ role: 'agent', content: '', streaming: true });
                    pendingMsgIdx = messages.length - 1;
                    startSpinner();
                }
                const chunk = payload.chunk || payload.content || payload.text || '';
                if (typeof chunk === 'string' && chunk.length > 0) {
                    messages[pendingMsgIdx].content += chunk;
                    scrollOffset = 0;
                }
                redraw();
                return;
            }

            if (type === 'agent_turn' || type === 'agent_response') {
                const content = payload.content || payload.message || payload.response || '';
                if (pendingMsgIdx !== -1) {
                    messages[pendingMsgIdx].content = content || messages[pendingMsgIdx].content;
                    messages[pendingMsgIdx].streaming = false;
                    pendingMsgIdx = -1;
                } else {
                    messages.push({ role: 'agent', content });
                }
                if (payload.conversation_id) { sessionId = payload.conversation_id; sessionLabel = sessionId.slice(0, 8); }
                busy = false;
                stopSpinner();
                scrollOffset = 0;
                redraw();
                fetchModel();
                return;
            }

            if (type === 'tool_start' || type === 'tool_call') {
                const toolName = payload.tool || payload.name || 'action';
                const toolArgs = payload.args || payload.arguments || '';
                const argsStr = typeof toolArgs === 'string' ? toolArgs : JSON.stringify(toolArgs);
                const label = `Executing: ${toolName}${argsStr ? ' ' + argsStr.slice(0, 40) : ''}`;
                const tIdx = messages.push({ role: 'tool', content: label, done: false, startTime: Date.now() }) - 1;
                toolProgress.push({ idx: tIdx, tool: toolName });
                scrollOffset = 0;
                redraw();
                return;
            }

            if (type === 'tool_end' || type === 'tool_result') {
                const toolName = payload.tool || payload.name || '';
                const entry = toolProgress.find(t => t.tool === toolName);
                if (entry) {
                    const msg = messages[entry.idx];
                    if (msg) {
                        msg.done = true;
                        msg.ms = Date.now() - (msg.startTime || Date.now());
                    }
                    toolProgress = toolProgress.filter(t => t !== entry);
                }
                scrollOffset = 0;
                redraw();
                return;
            }

            if (type === 'error') {
                const errMsg = payload.message || payload.error || 'Unknown error';
                if (pendingMsgIdx !== -1) {
                    messages[pendingMsgIdx].content = `Error: ${errMsg}`;
                    messages[pendingMsgIdx].streaming = false;
                    pendingMsgIdx = -1;
                } else {
                    messages.push({ role: 'system', content: `Error: ${errMsg}` });
                }
                busy = false;
                stopSpinner();
                redraw();
                return;
            }
        };

        ws.onclose = () => {
            connStatus = 'DISCONNECTED';
            wsInstance = null;
            redraw();
            setTimeout(() => { connectWebSocket(); }, 3000);
        };

        ws.onerror = () => {
            connStatus = 'ERROR';
            redraw();
        };

        ws.connect();
    }

    async function sendMessage() {
        const text = inputLines.join('\n').trim();
        if (!text || busy) return;

        inputLines = [''];
        inputCursorLine = 0;
        inputCursorCol = 0;
        busy = true;
        pendingMsgIdx = -1;
        toolProgress = [];

        addMessage('user', text);
        startSpinner();

        try {
            if (wsInstance && wsInstance._connected) {
                wsInstance.send(JSON.stringify({
                    type: 'agent_message',
                    payload: { message: text, conversation_id: sessionId, source: 'cli' }
                }));
            } else {
                const res = await httpPost(`${gateUrl}/api/agent`, { message: text, conversation_id: sessionId, source: 'cli' });
                const content = res?.content || res?.data?.content || res?.message || 'No response.';
                if (res?.conversation_id) { sessionId = res.conversation_id; sessionLabel = sessionId.slice(0, 8); }
                addMessage('agent', content);
                busy = false;
                stopSpinner();
                fetchModel();
            }
        } catch (e) {
            addMessage('system', `Signal error: ${e.message}`);
            busy = false;
            stopSpinner();
        }

        redraw();
    }

    async function newSession() {
        sessionId = null;
        sessionLabel = 'new';
        messages.length = 0;
        pendingMsgIdx = -1;
        toolProgress = [];
        inputLines = [''];
        inputCursorLine = 0;
        inputCursorCol = 0;
        scrollOffset = 0;
        busy = false;
        stopSpinner();
        addMessage('system', 'New session started.');
        await fetchSessions();
        redraw();
    }

    async function listSessions() {
        await fetchSessions();
        if (!sessions.length) {
            addMessage('system', 'No sessions found.');
            return;
        }
        const lines = sessions.slice(0, 10).map((s, i) => `${i + 1}. ${s.id?.slice(0, 8) || s.slice(0, 8)}`).join('  ');
        addMessage('system', `Sessions: ${lines}`);
    }

    function handleInput(key) {
        if (key === '\u0003') {
            cleanup();
            process.exit(0);
        }

        if (key === '\f') {
            messages.length = 0;
            scrollOffset = 0;
            redraw();
            return;
        }

        if (key === '\x0e') {
            newSession();
            return;
        }

        if (key === '\r' || key === '\n') {
            if (!busy) sendMessage();
            return;
        }

        if (key === '\r\n') {
            if (!busy) sendMessage();
            return;
        }

        if (key === '\x1b[A') {
            if (scrollOffset < 9999) { scrollOffset += 3; redraw(); }
            return;
        }

        if (key === '\x1b[B') {
            if (scrollOffset > 0) { scrollOffset -= 3; if (scrollOffset < 0) scrollOffset = 0; redraw(); }
            return;
        }

        if (key === '\x1b[5~') {
            scrollOffset += 10; redraw();
            return;
        }

        if (key === '\x1b[6~') {
            scrollOffset = Math.max(0, scrollOffset - 10); redraw();
            return;
        }

        if (busy) return;

        if (key === '\x1b\r' || key === '\x1b\n' || key === '\r\x1b') {
            const currentLine = inputLines[inputCursorLine] || '';
            const before = currentLine.slice(0, inputCursorCol);
            const after = currentLine.slice(inputCursorCol);
            inputLines[inputCursorLine] = before;
            inputLines.splice(inputCursorLine + 1, 0, after);
            inputCursorLine++;
            inputCursorCol = 0;
            redraw();
            return;
        }

        if (key === '\x1b[D') {
            if (inputCursorCol > 0) { inputCursorCol--; redraw(); }
            else if (inputCursorLine > 0) { inputCursorLine--; inputCursorCol = inputLines[inputCursorLine].length; redraw(); }
            return;
        }

        if (key === '\x1b[C') {
            const lineLen = (inputLines[inputCursorLine] || '').length;
            if (inputCursorCol < lineLen) { inputCursorCol++; redraw(); }
            else if (inputCursorLine < inputLines.length - 1) { inputCursorLine++; inputCursorCol = 0; redraw(); }
            return;
        }

        if (key === '\x1b[H' || key === '\x01') { inputCursorCol = 0; redraw(); return; }
        if (key === '\x1b[F' || key === '\x05') { inputCursorCol = (inputLines[inputCursorLine] || '').length; redraw(); return; }

        if (key === '\u007f' || key === '\b') {
            if (inputCursorCol > 0) {
                const line = inputLines[inputCursorLine];
                inputLines[inputCursorLine] = line.slice(0, inputCursorCol - 1) + line.slice(inputCursorCol);
                inputCursorCol--;
            } else if (inputCursorLine > 0) {
                const prev = inputLines[inputCursorLine - 1];
                const curr = inputLines[inputCursorLine];
                inputCursorCol = prev.length;
                inputLines[inputCursorLine - 1] = prev + curr;
                inputLines.splice(inputCursorLine, 1);
                inputCursorLine--;
            }
            redraw();
            return;
        }

        if (key === '\x1b[3~') {
            const line = inputLines[inputCursorLine] || '';
            if (inputCursorCol < line.length) {
                inputLines[inputCursorLine] = line.slice(0, inputCursorCol) + line.slice(inputCursorCol + 1);
            } else if (inputCursorLine < inputLines.length - 1) {
                inputLines[inputCursorLine] = line + inputLines[inputCursorLine + 1];
                inputLines.splice(inputCursorLine + 1, 1);
            }
            redraw();
            return;
        }

        if (key.length === 1 && key.charCodeAt(0) >= 32) {
            const line = inputLines[inputCursorLine] || '';
            inputLines[inputCursorLine] = line.slice(0, inputCursorCol) + key + line.slice(inputCursorCol);
            inputCursorCol++;
            redraw();
            return;
        }

        if (key.length > 1 && !key.startsWith('\x1b')) {
            for (const ch of key) {
                if (ch.charCodeAt(0) >= 32) {
                    const line = inputLines[inputCursorLine] || '';
                    inputLines[inputCursorLine] = line.slice(0, inputCursorCol) + ch + line.slice(inputCursorCol);
                    inputCursorCol++;
                }
            }
            redraw();
        }
    }

    function cleanup() {
        stopSpinner();
        if (wsInstance) { try { wsInstance.close(); } catch {} }
        normalScreen();
        showCursor();
    }

    process.stdin.setRawMode(true);
    process.stdin.resume();
    process.stdin.setEncoding('utf8');

    process.stdin.on('data', (key) => { handleInput(key); });
    process.stdout.on('resize', () => { redraw(); });

    process.on('exit', cleanup);
    process.on('SIGINT', () => { cleanup(); process.exit(0); });
    process.on('SIGTERM', () => { cleanup(); process.exit(0); });

    addMessage('system', 'Welcome to DARDCOR CLI. Type a message and press Enter to send.');

    connectWebSocket();
    redraw();
}
