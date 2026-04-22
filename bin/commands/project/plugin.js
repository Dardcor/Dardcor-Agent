import { inf, ok, err, wrn, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';
import { execSync } from 'child_process';

const PLUGINS_DIR = path.join(os.homedir(), '.dardcor', 'plugins');
const PLUGINS_CONFIG = path.join(os.homedir(), '.dardcor', 'plugins.json');

function loadPlugins() {
  try { return JSON.parse(fs.readFileSync(PLUGINS_CONFIG, 'utf8')); } catch { return { plugins: [] }; }
}
function savePlugins(cfg) {
  const dir = path.dirname(PLUGINS_CONFIG);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(PLUGINS_CONFIG, JSON.stringify(cfg, null, 2));
}

export async function handlePlugin(action, arg) {
  if (!fs.existsSync(PLUGINS_DIR)) fs.mkdirSync(PLUGINS_DIR, { recursive: true });

  switch (action) {
    case 'list': {
      const cfg = loadPlugins();
      if (cfg.plugins.length === 0) { inf('No plugins installed.'); return; }
      console.log(`\n${C.bold}Installed Plugins:${C.reset}\n`);
      cfg.plugins.forEach(p => {
        console.log(`  ${C.cyan}${p.name}${C.reset}  ${C.dim}v${p.version || '?'} — ${p.description || ''}${C.reset}`);
      });
      console.log('');
      break;
    }
    case 'add':
    case 'install': {
      if (!arg) { err('Usage: dardcor plugin add <package-name>'); return; }
      inf(`Installing plugin: ${arg}`);
      try {
        execSync(`npm install ${arg}`, { cwd: PLUGINS_DIR, stdio: 'inherit' });
        const cfg = loadPlugins();
        if (!cfg.plugins.find(p => p.name === arg)) {
          cfg.plugins.push({ name: arg, installedAt: new Date().toISOString() });
          savePlugins(cfg);
        }
        ok(`Plugin installed: ${arg}`);
      } catch (e) { err(`Failed to install plugin: ${e.message}`); }
      break;
    }
    case 'remove':
    case 'uninstall': {
      if (!arg) { err('Usage: dardcor plugin remove <package-name>'); return; }
      const cfg = loadPlugins();
      cfg.plugins = cfg.plugins.filter(p => p.name !== arg);
      savePlugins(cfg);
      try {
        execSync(`npm uninstall ${arg}`, { cwd: PLUGINS_DIR, stdio: 'inherit' });
      } catch {}
      ok(`Plugin removed: ${arg}`);
      break;
    }
    case 'create': {
      if (!arg) { err('Usage: dardcor plugin create <name>'); return; }
      const pluginDir = path.join(PLUGINS_DIR, arg);
      fs.mkdirSync(pluginDir, { recursive: true });
      const template = {
        name: arg,
        version: '1.0.0',
        description: 'A Dardcor plugin',
        main: 'index.js',
        keywords: ['dardcor', 'plugin'],
      };
      fs.writeFileSync(path.join(pluginDir, 'package.json'), JSON.stringify(template, null, 2));
      fs.writeFileSync(path.join(pluginDir, 'index.js'), `// ${arg} plugin\nexport default {\n  name: '${arg}',\n  init(agent) {\n    // Add custom commands, hooks, or tools here\n  }\n};\n`);
      ok(`Plugin scaffold created: ${pluginDir}`);
      break;
    }
    default:
      err(`Unknown action: ${action}. Use: list, add, remove, create`);
  }
}
