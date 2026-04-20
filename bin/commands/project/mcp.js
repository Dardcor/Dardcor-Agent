import { httpGet, httpPost, ensureEngine, inf, ok, err, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

const MCP_CONFIG = path.join(os.homedir(), '.dardcor', 'mcp.json');

function loadMCPConfig() {
  try { return JSON.parse(fs.readFileSync(MCP_CONFIG, 'utf8')); } catch { return { servers: {} }; }
}
function saveMCPConfig(cfg) {
  const dir = path.dirname(MCP_CONFIG);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(MCP_CONFIG, JSON.stringify(cfg, null, 2));
}

export async function handleMCP(action, name, command, ...args) {
  switch (action) {
    case 'list': {
      const cfg = loadMCPConfig();
      const servers = Object.entries(cfg.servers || {});
      if (servers.length === 0) { inf('No MCP servers configured.'); return; }
      console.log(`\n${C.bold}MCP Servers:${C.reset}\n`);
      servers.forEach(([n, s]) => {
        console.log(`  ${C.cyan}${n}${C.reset}  ${C.dim}${s.command} ${(s.args||[]).join(' ')}${C.reset}`);
      });
      console.log('');
      break;
    }
    case 'add': {
      if (!name || !command) { err('Usage: dardcor mcp add <name> <command> [args...]'); return; }
      const cfg = loadMCPConfig();
      cfg.servers[name] = { command, args: args.flat() };
      saveMCPConfig(cfg);
      ok(`MCP server added: ${name}`);
      break;
    }
    case 'remove':
    case 'rm': {
      if (!name) { err('Usage: dardcor mcp remove <name>'); return; }
      const cfg = loadMCPConfig();
      delete cfg.servers[name];
      saveMCPConfig(cfg);
      ok(`MCP server removed: ${name}`);
      break;
    }
    case 'builtin': {
      const builtins = {
        'context7':  { command: 'npx', args: ['-y', '@upstash/context7-mcp@latest'], desc: 'Up-to-date library docs' },
        'grep-app':  { command: 'npx', args: ['-y', '@grep-app/mcp@latest'], desc: 'Web code search' },
        'websearch': { command: 'npx', args: ['-y', '@modelcontextprotocol/server-brave-search'], desc: 'Web search (Brave)' },
        'filesystem':{ command: 'npx', args: ['-y', '@modelcontextprotocol/server-filesystem', process.cwd()], desc: 'Local file system access' },
        'github':    { command: 'npx', args: ['-y', '@modelcontextprotocol/server-github'], desc: 'GitHub API integration' },
        'memory':    { command: 'npx', args: ['-y', '@modelcontextprotocol/server-memory'], desc: 'Persistent memory' },
      };
      console.log(`\n${C.bold}Built-in MCP Servers:${C.reset}\n`);
      for (const [n, s] of Object.entries(builtins)) {
        console.log(`  ${C.cyan}${n.padEnd(12)}${C.reset}  ${C.dim}${s.desc}${C.reset}`);
        console.log(`  ${C.dim}Install: dardcor mcp add ${n} ${s.command} ${s.args.join(' ')}${C.reset}\n`);
      }
      break;
    }
    default:
      err(`Unknown action: ${action}. Use: list, add, remove, builtin`);
  }
}
