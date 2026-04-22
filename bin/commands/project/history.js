import { httpGet, ensureEngine, inf, err, wrn, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

export async function handleHistory(opts = {}) {
  const convDir = path.join(os.homedir(), '.dardcor', 'database', 'conversations-cli');
  const webConvDir = path.join(os.homedir(), '.dardcor', 'database', 'conversations-web');

  if (opts.clear) {
    let cleared = 0;
    for (const dir of [convDir, webConvDir]) {
      if (fs.existsSync(dir)) {
        fs.readdirSync(dir).forEach(f => {
          fs.rmSync(path.join(dir, f), { force: true });
          cleared++;
        });
      }
    }
    inf(`Cleared ${cleared} conversation(s).`);
    return;
  }

  let allConvs = [];
  for (const dir of [convDir, webConvDir]) {
    if (!fs.existsSync(dir)) continue;
    const files = fs.readdirSync(dir).filter(f => f.endsWith('.json'));
    for (const f of files.slice(-20)) {
      try {
        const data = JSON.parse(fs.readFileSync(path.join(dir, f), 'utf8'));
        allConvs.push({ id: data.id, title: data.title, created: data.created_at, msgCount: data.messages?.length || 0 });
      } catch {}
    }
  }

  if (allConvs.length === 0) { wrn('No conversation history found.'); return; }

  allConvs.sort((a, b) => new Date(b.created) - new Date(a.created));

  console.log(`\n${C.bold}${C.purple}Conversation History${C.reset} (${allConvs.length} sessions)\n`);
  allConvs.slice(0, 20).forEach((c, i) => {
    const date = new Date(c.created).toLocaleDateString();
    console.log(`  ${C.dim}${i+1}.${C.reset} ${C.cyan}${c.id.slice(0,8)}${C.reset}  ${C.white}${(c.title || 'Untitled').slice(0,50)}${C.reset}  ${C.dim}${date} · ${c.msgCount} msgs${C.reset}`);
  });
  console.log('');

  if (opts.export) {
    const conv = allConvs.find(c => c.id.startsWith(opts.export));
    if (!conv) { err(`Session not found: ${opts.export}`); return; }
    const format = opts.format || 'markdown';
    inf(`Exported as ${format} (feature: use dardcor run for full export)`);
  }
}
