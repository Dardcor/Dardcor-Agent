import { httpPost, ensureEngine, inf, err, ok, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

export async function handleResume(opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }

  const convDir = path.join(os.homedir(), '.dardcor', 'database', 'conversations-cli');
  if (!fs.existsSync(convDir)) { err('No CLI sessions found.'); return; }

  let sessionId = opts.session;

  if (!sessionId) {
    const files = fs.readdirSync(convDir)
      .filter(f => f.endsWith('.json'))
      .sort((a, b) => fs.statSync(path.join(convDir, b)).mtime - fs.statSync(path.join(convDir, a)).mtime);

    if (files.length === 0) { err('No sessions found.'); return; }
    const latest = JSON.parse(fs.readFileSync(path.join(convDir, files[0]), 'utf8'));
    sessionId = latest.id;
    inf(`Resuming latest session: ${C.cyan}${sessionId.slice(0,8)}${C.reset} — "${latest.title || 'Untitled'}"`);
  }

  inf(`Session ${sessionId.slice(0,8)} loaded. Continuing with 'dardcor cli'...`);
  process.env.DARDCOR_RESUME_SESSION = sessionId;
  const { runCLI } = await import('../../cli_agent.js');
  await runCLI();
}
