import http from 'http';
import fs from 'fs';
import path from 'path';
import os from 'os';
import { execSync } from 'child_process';

export const C = {
  reset:   '\x1b[0m',
  bold:    '\x1b[1m',
  dim:     '\x1b[2m',
  red:     '\x1b[31m',
  green:   '\x1b[32m',
  yellow:  '\x1b[33m',
  blue:    '\x1b[34m',
  cyan:    '\x1b[36m',
  purple:  '\x1b[38;5;93m',
  white:   '\x1b[37m',
};

export const GATE = 'http://127.0.0.1:25000';
export const BASE_DIR = path.join(os.homedir(), '.dardcor');
export const CONFIG_FILE = path.join(BASE_DIR, 'config.json');
export const ACCOUNT_FILE = path.join(BASE_DIR, 'account.json');
export const SESSION_DIR = path.join(BASE_DIR, 'session');
export const CACHE_DIR = path.join(BASE_DIR, 'cache');

export function ok(msg)  { console.log(`${C.green}[✓]${C.reset} ${msg}`); }
export function err(msg) { console.log(`${C.red}[✗]${C.reset} ${msg}`); }
export function inf(msg) { console.log(`${C.cyan}[i]${C.reset} ${msg}`); }
export function wrn(msg) { console.log(`${C.yellow}[!]${C.reset} ${msg}`); }

export function loadConfig() {
  try {
    if (fs.existsSync(CONFIG_FILE)) return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8'));
  } catch {}
  return {};
}

export function saveConfig(cfg) {
  if (!fs.existsSync(BASE_DIR)) fs.mkdirSync(BASE_DIR, { recursive: true });
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(cfg, null, 2));
}

export function loadAccount() {
  try {
    if (fs.existsSync(ACCOUNT_FILE)) return JSON.parse(fs.readFileSync(ACCOUNT_FILE, 'utf8'));
  } catch {}
  return {};
}

export function saveAccount(acc) {
  if (!fs.existsSync(BASE_DIR)) fs.mkdirSync(BASE_DIR, { recursive: true });
  fs.writeFileSync(ACCOUNT_FILE, JSON.stringify(acc, null, 2));
}

export async function httpPost(endpoint, body) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify(body);
    const u = new URL(GATE + endpoint);
    const req = http.request({
      hostname: u.hostname, port: u.port, path: u.pathname,
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) }
    }, res => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => {
        try { resolve(JSON.parse(d)); } catch { resolve(d); }
      });
    });
    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

export async function httpGet(endpoint) {
  return new Promise((resolve, reject) => {
    const u = new URL(GATE + endpoint);
    const req = http.request({
      hostname: u.hostname, port: u.port, path: u.pathname + u.search,
      method: 'GET'
    }, res => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => {
        try { resolve(JSON.parse(d)); } catch { resolve(d); }
      });
    });
    req.on('error', reject);
    req.end();
  });
}

export async function ensureEngine() {
  try {
    const check = await new Promise(r => {
      const req = http.get(GATE + '/health', res => r(res.statusCode === 200)).on('error', () => r(false));
      req.setTimeout(2000);
    });
    return check;
  } catch { return false; }
}

export async function sendToAgent(prompt, opts = {}) {
  const body = { message: prompt, source: 'cli', ...opts };
  const res = await httpPost('/api/agent', body);
  return res.content || res.data?.content || res;
}

export function gitExec(cmd) {
  try {
    return execSync(cmd, { encoding: 'utf8', stdio: ['pipe','pipe','pipe'] }).trim();
  } catch (e) {
    return null;
  }
}

export function getStagedDiff() {
  return gitExec('git diff --cached --stat') || '';
}

export function getFullStagedDiff() {
  return gitExec('git diff --cached') || '';
}

export function getCurrentBranch() {
  return gitExec('git branch --show-current') || 'main';
}

export function printBox(content, title = '') {
  const lines = content.split('\n');
  const width = Math.min(Math.max(...lines.map(l => l.length), title.length) + 4, 100);
  const line = '─'.repeat(width);
  console.log(`\n${C.purple}┌${line}┐${C.reset}`);
  if (title) {
    const pad = Math.floor((width - title.length) / 2);
    console.log(`${C.purple}│${C.reset}${' '.repeat(pad)}${C.bold}${title}${C.reset}${' '.repeat(width - pad - title.length)}${C.purple}│${C.reset}`);
    console.log(`${C.purple}├${line}┤${C.reset}`);
  }
  for (const l of lines.slice(0, 30)) {
    const padded = l.slice(0, width - 2).padEnd(width - 2);
    console.log(`${C.purple}│${C.reset} ${padded} ${C.purple}│${C.reset}`);
  }
  console.log(`${C.purple}└${line}┘${C.reset}\n`);
}

export async function confirmPrompt(question) {
  process.stdout.write(`${C.yellow}${question} (y/N): ${C.reset}`);
  return new Promise(resolve => {
    process.stdin.setRawMode(true);
    process.stdin.resume();
    process.stdin.setEncoding('utf8');
    process.stdin.once('data', key => {
      process.stdin.setRawMode(false);
      process.stdin.pause();
      console.log(key.trim());
      resolve(key.trim().toLowerCase() === 'y');
    });
  });
}
