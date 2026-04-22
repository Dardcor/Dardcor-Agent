import { httpGet, ensureEngine, inf, err, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

export async function handleStats(opts = {}) {
  const statsFile = path.join(os.homedir(), '.dardcor', 'database', 'stats.json');
  let stats = { totalMessages: 0, totalTokens: 0, totalCost: 0, sessions: 0, providers: {} };

  if (fs.existsSync(statsFile)) {
    try { stats = { ...stats, ...JSON.parse(fs.readFileSync(statsFile, 'utf8')) }; } catch { }
  }


  const convDirs = [
    path.join(os.homedir(), '.dardcor', 'database', 'conversations-cli'),
    path.join(os.homedir(), '.dardcor', 'database', 'conversations-web'),
  ];

  let totalConvs = 0, totalMsgs = 0;
  for (const dir of convDirs) {
    if (!fs.existsSync(dir)) continue;
    const files = fs.readdirSync(dir).filter(f => f.endsWith('.json'));
    totalConvs += files.length;
    for (const f of files) {
      try {
        const data = JSON.parse(fs.readFileSync(path.join(dir, f), 'utf8'));
        totalMsgs += (data.messages || []).length;
      } catch { }
    }
  }

  console.log(`\n${C.bold}${C.purple}Dardcor Usage Statistics${C.reset}\n`);
  console.log(`  ${C.cyan}Sessions:${C.reset}       ${totalConvs}`);
  console.log(`  ${C.cyan}Messages:${C.reset}       ${totalMsgs}`);
  console.log(`  ${C.cyan}Total Tokens:${C.reset}   ${(stats.totalTokens || 0).toLocaleString()}`);
  console.log(`  ${C.cyan}Est. Cost:${C.reset}      $${((stats.totalCost || 0)).toFixed(4)}`);

  if (Object.keys(stats.providers || {}).length > 0) {
    console.log(`\n${C.bold}Provider Usage:${C.reset}`);
    for (const [p, s] of Object.entries(stats.providers)) {
      console.log(`  ${C.dim}•${C.reset} ${p}: ${s.messages || 0} messages`);
    }
  }
  console.log('');
}
