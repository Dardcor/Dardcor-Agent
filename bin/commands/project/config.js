import { loadConfig, saveConfig, inf, ok, err, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

export async function handleConfig(opts = {}) {
  const cfg = loadConfig();
  const configDir = path.join(os.homedir(), '.dardcor');

  if (opts.set) {
    const [key, ...valueParts] = opts.set.split('=');
    const value = valueParts.join('=');
    cfg[key] = value;
    saveConfig(cfg);
    ok(`Set ${key} = ${value}`);
    return;
  }

  if (opts.themes) {
    const themesDir = path.join(path.dirname(path.dirname(path.dirname(import.meta.url.replace('file:///', '')))), 'themes');
    try {
      const themes = fs.readdirSync(themesDir.replace(/^\//, '')).filter(f => f.endsWith('.json'));
      console.log(`\n${C.bold}Available Themes:${C.reset}\n`);
      themes.forEach(t => console.log(`  ${C.cyan}${t.replace('.json','')}${C.reset}`));
    } catch {
      console.log(`${C.dim}No themes directory found${C.reset}`);
    }
    return;
  }

  console.log(`\n${C.bold}${C.purple}Dardcor Configuration${C.reset}`);
  console.log(`${C.dim}Config: ${configDir}/config.json${C.reset}\n`);

  if (Object.keys(cfg).length === 0) {
    console.log(`${C.dim}No configuration set. Use: dardcor config --set key=value${C.reset}`);
    return;
  }

  for (const [k, v] of Object.entries(cfg)) {
    const display = k.toLowerCase().includes('key') || k.toLowerCase().includes('secret') || k.toLowerCase().includes('token')
      ? (String(v).slice(0, 8) + '***')
      : v;
    console.log(`  ${C.cyan}${k}${C.reset} = ${C.white}${display}${C.reset}`);
  }
  console.log('');
}
