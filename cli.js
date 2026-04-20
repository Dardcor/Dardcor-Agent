#!/usr/bin/env node
import { run } from './bin/run.js';
import { runCLI } from './bin/cli_agent.js';
import { runDoctor } from './bin/doctor.js';
import { printHelp, printBanner, printCommandHelp } from './bin/help.js';

const args = process.argv.slice(2);
const [command, ...rest] = args;

// ─── Option parser ────────────────────────────────────────────────────────────
function parseOpts(args) {
  const opts = {};
  const positional = [];
  for (let i = 0; i < args.length; i++) {
    const a = args[i];
    if (a === '--help' || a === '-h') {
      opts.help = true;
    } else if (a === '--version' || a === '-v') {
      opts.version = true;
    } else if (a.startsWith('--')) {
      const key = a.slice(2);
      // Next arg is the value unless it starts with -- or missing
      if (args[i + 1] !== undefined && !args[i + 1].startsWith('--')) {
        opts[key] = args[++i];
      } else {
        opts[key] = true;
      }
    } else if (a.startsWith('-') && a.length === 2) {
      // short flag  -f  -o value
      const key = a.slice(1);
      if (args[i + 1] !== undefined && !args[i + 1].startsWith('-')) {
        opts[key] = args[++i];
      } else {
        opts[key] = true;
      }
    } else {
      positional.push(a);
    }
  }
  opts._positional = positional;
  return opts;
}

// ─── Print version ────────────────────────────────────────────────────────────
function printVersion() {
  import('./bin/help.js').then(({ printBanner }) => printBanner());
}

// ─── Main router ─────────────────────────────────────────────────────────────
async function start() {
  try {
    const opts = parseOpts(rest);

    // Global --version / -v
    if (opts.version && !command) { printVersion(); return; }

    // ─── Global help: dardcor  |  dardcor help  |  dardcor --help ───
    if (command === '--help' || command === '-h') { printHelp(); return; }
    if (command === 'help') {
      const sub = rest[0];
      if (sub && !sub.startsWith('-')) { printCommandHelp(sub); }
      else { printHelp(); }
      return;
    }

    // Default: run interactive CLI if no command
    if (!command) { await runCLI(); return; }

    // Per-command --help  (dardcor commit --help)
    if (opts.help) { printCommandHelp(command); return; }

    // ─── Core ────────────────────────────────────────────────────────────────
    if (command === 'run')    { await run();       return; }
    if (command === 'cli')    { await runCLI();    return; }
    if (command === 'doctor') { await runDoctor(); return; }

    // dardcor auto [yolo|safe]
    if (command === 'auto') {
      const mode = rest[0] || 'auto';
      await runCLI(mode);
      return;
    }

    // dardcor auth login | logout
    if (command === 'auth') {
      const sub = rest[0];
      if (sub === 'login') {
        const { handleLogin } = await import('./bin/commands/project/login.js');
        await handleLogin(opts);
      } else if (sub === 'logout') {
        // Implement logout if needed, otherwise just message
        console.log(`\n  ${C.purple}DARDCOR${C.reset} Logged out successfully.\n`);
      } else {
        console.log(`\n  ${C.red}${C.bold}Error:${C.reset} Unknown auth command: ${sub || 'none'}\n  Use: dardcor auth login | logout\n`);
      }
      return;
    }

    // ─── AI Commands ─────────────────────────────────────────────────────────
    if (command === 'ask') {
      const prompt = rest.filter(r => !r.startsWith('-')).join(' ');
      if (!prompt) { printCommandHelp('ask'); return; }
      const { handleAsk } = await import('./bin/commands/ai/ask.js');
      await handleAsk(prompt, opts); return;
    }

    if (command === 'fix') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('fix'); return; }
      const { handleFix } = await import('./bin/commands/ai/fix.js');
      await handleFix(desc, opts); return;
    }

    if (command === 'debug') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('debug'); return; }
      const { handleDebug } = await import('./bin/commands/ai/debug.js');
      await handleDebug(desc, opts); return;
    }

    if (command === 'explain') {
      const target = opts._positional[0];
      if (!target) { printCommandHelp('explain'); return; }
      const { handleExplain } = await import('./bin/commands/ai/explain.js');
      await handleExplain(target, opts); return;
    }

    if (command === 'doc') {
      const target = opts._positional[0] || '.';
      const { handleDoc } = await import('./bin/commands/ai/doc.js');
      await handleDoc(target, opts); return;
    }

    if (command === 'plan') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('plan'); return; }
      const { handlePlan } = await import('./bin/commands/ai/plan.js');
      await handlePlan(desc, opts); return;
    }

    if (command === 'refactor') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('refactor'); return; }
      const { handleRefactor } = await import('./bin/commands/ai/refactor.js');
      await handleRefactor(desc, opts); return;
    }

    if (command === 'scaffold') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('scaffold'); return; }
      const { handleScaffold } = await import('./bin/commands/ai/scaffold.js');
      await handleScaffold(desc, opts); return;
    }

    if (command === 'test') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('test'); return; }
      const { handleTest } = await import('./bin/commands/ai/test.js');
      await handleTest(desc, opts); return;
    }

    if (command === 'review') {
      const { handleReview } = await import('./bin/commands/ai/review.js');
      await handleReview(opts); return;
    }

    if (command === 'optimize') {
      const target = opts._positional[0];
      if (!target) { printCommandHelp('optimize'); return; }
      const { handleOptimize } = await import('./bin/commands/ai/optimize.js');
      await handleOptimize(target, opts); return;
    }

    if (command === 'translate') {
      const file = opts._positional[0];
      if (!file) { printCommandHelp('translate'); return; }
      const { handleTranslate } = await import('./bin/commands/ai/translate.js');
      await handleTranslate(file, opts); return;
    }

    if (command === 'sandbox') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('sandbox'); return; }
      const { handleSandbox } = await import('./bin/commands/ai/sandbox.js');
      await handleSandbox(desc, opts); return;
    }

    if (command === 'think') {
      const prompt = opts._positional.join(' ');
      if (!prompt) { printCommandHelp('think'); return; }
      const { handleThink } = await import('./bin/commands/ai/think.js');
      await handleThink(prompt, opts); return;
    }

    // ─── Git Commands ─────────────────────────────────────────────────────────
    if (command === 'commit') {
      const { handleCommit } = await import('./bin/commands/git/commit.js');
      await handleCommit(opts); return;
    }

    if (command === 'branch') {
      const desc = opts._positional.join(' ');
      if (!desc) { printCommandHelp('branch'); return; }
      const { handleBranch } = await import('./bin/commands/git/branch.js');
      await handleBranch(desc, opts); return;
    }

    if (command === 'pr') {
      const { handlePR } = await import('./bin/commands/git/pr.js');
      await handlePR(opts); return;
    }

    if (command === 'resolve') {
      const { handleResolve } = await import('./bin/commands/git/resolve.js');
      await handleResolve(opts); return;
    }

    if (command === 'diff') {
      const ref = opts._positional[0];
      const { handleDiff } = await import('./bin/commands/git/diff.js');
      await handleDiff(ref, opts); return;
    }

    // ─── Project / System Commands ────────────────────────────────────────────
    if (command === 'config') {
      const { handleConfig } = await import('./bin/commands/project/config.js');
      await handleConfig(opts); return;
    }

    if (command === 'history') {
      const { handleHistory } = await import('./bin/commands/project/history.js');
      await handleHistory(opts); return;
    }

    if (command === 'resume') {
      const { handleResume } = await import('./bin/commands/project/resume.js');
      await handleResume(opts); return;
    }

    if (command === 'models') {
      const { handleModels } = await import('./bin/commands/project/models.js');
      // Pass all flags through: --filter, --search, --provider, --pricing, --all, --count
      await handleModels(opts); return;
    }

    if (command === 'stats') {
      const { handleStats } = await import('./bin/commands/project/stats.js');
      await handleStats(opts); return;
    }

    if (command === 'login') {
      const { handleLogin } = await import('./bin/commands/project/login.js');
      await handleLogin(opts); return;
    }

    if (command === 'mcp') {
      if (rest[0] === '--help' || rest[0] === '-h' || !rest[0]) {
        printCommandHelp('mcp'); return;
      }
      const { handleMCP } = await import('./bin/commands/project/mcp.js');
      const [action, name, cmd, ...mcpArgs] = rest;
      await handleMCP(action, name, cmd, ...mcpArgs); return;
    }

    if (command === 'plugin') {
      if (rest[0] === '--help' || rest[0] === '-h' || !rest[0]) {
        printCommandHelp('plugin'); return;
      }
      const { handlePlugin } = await import('./bin/commands/project/plugin.js');
      const [action, arg] = rest;
      await handlePlugin(action, arg); return;
    }

    if (command === 'index') {
      const { handleIndex } = await import('./bin/commands/project/index.js');
      await handleIndex(opts); return;
    }

    if (command === 'init') {
      const { handleInitProject } = await import('./bin/commands/project/init_project.js');
      await handleInitProject(opts); return;
    }

    // ─── Unknown command ──────────────────────────────────────────────────────
    const C = {
      reset: '\x1b[0m', bold: '\x1b[1m', dim: '\x1b[2m',
      red: '\x1b[31m', cyan: '\x1b[36m', yellow: '\x1b[33m',
    };
    console.log(`\n  ${C.red}${C.bold}Unknown command:${C.reset} ${C.yellow}${command}${C.reset}`);
    console.log(`\n  ${C.dim}Run${C.reset} ${C.cyan}dardcor help${C.reset}${C.dim} to see all commands.${C.reset}`);
    console.log(`  ${C.dim}Run${C.reset} ${C.cyan}dardcor <command> --help${C.reset}${C.dim} for per-command usage.${C.reset}\n`);
    process.exit(1);

  } catch (e) {
    const C = { reset: '\x1b[0m', red: '\x1b[31m', bold: '\x1b[1m', dim: '\x1b[2m' };
    console.error(`\n${C.red}${C.bold}[!] DARDCOR ERROR${C.reset}  ${e.message}`);
    if (process.env.DEBUG) console.error(e.stack);
    else console.error(`${C.dim}    Set DEBUG=1 for stack trace${C.reset}`);
    process.exit(1);
  }
}

start();
