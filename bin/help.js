import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import path from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);


let VERSION = '1.0.11';
try {
  const pkg = JSON.parse(readFileSync(path.join(__dirname, '..', 'package.json'), 'utf8'));
  VERSION = pkg.version;
} catch { }


const C = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  italic: '\x1b[3m',
  under: '\x1b[4m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[97m',
  gray: '\x1b[90m',
  purple: '\x1b[38;5;93m',
  violet: '\x1b[38;5;135m',
  orange: '\x1b[38;5;208m',
  pink: '\x1b[38;5;213m',
};


const cols = () => process.stdout.columns || 100;
const hr = (ch = '─') => `${C.gray}${ch.repeat(Math.min(cols() - 2, 72))}${C.reset}`;

const tag = (label, color = C.cyan) => `${color}${C.bold}${label}${C.reset}`;
const flag = (f, desc, def = '') => {
  const left = `    ${C.green}${f.padEnd(28)}${C.reset}`;
  const right = `${C.dim}${desc}${def ? ` ${C.gray}[default: ${def}]${C.reset}` : ''}${C.reset}`;
  return `${left}${right}`;
};
const example = (code, comment = '') => {
  const ex = `    ${C.cyan}${code}${C.reset}`;
  const cmt = comment ? `  ${C.gray}# ${comment}${C.reset}` : '';
  return `${ex}${cmt}`;
};
const note = (msg) => `  ${C.yellow}ℹ${C.reset}  ${C.dim}${msg}${C.reset}`;


const banner = [
  '  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ',
  '  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗',
  '  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝',
  '  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗',
  '  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║',
  '  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝',
  '       S   U   P   R   E   M   E      A   G   E   N   T     '
];

export function printBanner(subtitle = '') {
  process.stdout.write('\n');
  banner.forEach((l, i) => {
    const col = i === 6 ? C.cyan : C.purple;
    console.log(`${col}${C.bold}${l}${C.reset}`);
  });
  const ver = `${C.gray}v${VERSION}${C.reset}`;
  const sub = subtitle
    ? `  ${C.cyan}❯${C.reset} ${C.bold}${subtitle}${C.reset}`
    : '';
  console.log(
    `\n  ${C.bold}${C.purple}DARDCOR AGENT${C.reset}  ${ver}  ${C.dim}Superior Autonomous System${C.reset}${sub ? '\n' + sub : ''}\n`
  );
}


const COMMAND_HELP = {


  run: {
    usage: 'dardcor run',
    short: 'Start the Dashboard UI and backend engine',
    desc: 'Launches the Go backend (port 25000) + Vite dev server when in source mode.\nAutomatically opens the browser. Press Ctrl+C to stop.',
    flags: [
      ['(no flags)', 'Runs with auto-detected mode (dev vs prod)'],
    ],
    examples: [
      ['dardcor run', 'Start dashboard'],
    ],
  },

  cli: {
    usage: 'dardcor cli',
    short: 'Interactive TUI terminal agent',
    desc: 'Launches a full-screen terminal interface powered by raw mode.\nType queries directly, see responses inline. Ctrl+C to exit.',
    flags: [
      ['(no flags)', 'Connects to engine on port 25000 (auto-starts if offline)'],
    ],
    examples: [
      ['dardcor cli', 'Open TUI agent'],
    ],
  },

  doctor: {
    usage: 'dardcor doctor',
    short: 'Diagnose and auto-fix system issues',
    desc: 'Checks Go installation, Node version, directory structure, config file,\nport availability, and engine connectivity. Auto-repairs what it can.',
    flags: [
      ['(no flags)', 'Runs full diagnostic and auto-repair'],
    ],
    examples: [
      ['dardcor doctor', 'Run full system check'],
    ],
  },


  ask: {
    usage: 'dardcor ask <prompt>',
    short: 'Send a one-shot question to the AI agent',
    desc: 'Sends your question to the active AI provider and prints the answer.\nDoes NOT modify files unless the agent decides to.',
    flags: [
      ['(no flags)', 'Uses default provider and model from config'],
    ],
    examples: [
      ['dardcor ask "what is a closure in JS?"', 'Quick concept question'],
      ['dardcor ask "list all TODO comments in src/"', 'Ask about your codebase'],
      ['dardcor ask "how do I center a div"', 'Frontend help'],
    ],
  },

  fix: {
    usage: 'dardcor fix <description>',
    short: 'Find and fix a bug in the codebase',
    desc: 'Instructs the agent to read relevant files, trace the root cause,\nthen apply a fix using [ACTION] commands. Full agentic loop.',
    flags: [
      ['(no flags)', 'Runs fix with full file read/write access'],
    ],
    examples: [
      ['dardcor fix "login endpoint returns 404"', 'Backend bug'],
      ['dardcor fix "TypeError: cannot read null"', 'Runtime error'],
      ['dardcor fix "build fails with TS2345"', 'TypeScript error'],
    ],
  },

  debug: {
    usage: 'dardcor debug <description>',
    short: 'Deep debugging assistant with root cause analysis',
    desc: 'Unlike `fix`, debug produces a full diagnostic report:\nroot cause, reproduction steps, proposed fix, and prevention tips.',
    flags: [
      ['(no flags)', 'Full diagnostic mode'],
    ],
    examples: [
      ['dardcor debug "memory leak in worker pool"', 'Performance issue'],
      ['dardcor debug "flaky test in CI"', 'CI issue'],
    ],
  },

  explain: {
    usage: 'dardcor explain <file|snippet>',
    short: 'Explain code in plain language',
    desc: 'Pass a file path or a code snippet. The agent reads the file\nand explains what it does, how it works, and notable patterns.',
    flags: [
      ['--depth shallow', 'Quick overview only'],
      ['--depth normal', 'Standard explanation (default)', 'normal'],
      ['--depth deep', 'Deep-dive: algorithms, edge cases, suggestions'],
    ],
    examples: [
      ['dardcor explain src/agent.go', 'Explain a file'],
      ['dardcor explain src/utils/retry.ts --depth deep', 'Deep dive'],
      ['dardcor explain "reduce((a,b) => a+b, 0)"', 'Inline snippet'],
    ],
  },

  doc: {
    usage: 'dardcor doc <file|target>',
    short: 'Generate documentation for code',
    desc: 'Reads the file and generates docs in the chosen format.\nOutput is printed to stdout — pipe or redirect as needed.',
    flags: [
      ['--readme', 'Generate a full README.md'],
      ['--type jsdoc', 'Add JSDoc/TSDoc comments to functions (default)', 'jsdoc'],
      ['--type readme', 'Generate README.md'],
      ['--type adr', 'Architecture Decision Record'],
      ['--type changelog', 'CHANGELOG.md from code structure'],
    ],
    examples: [
      ['dardcor doc src/api.ts', 'Generate JSDoc comments'],
      ['dardcor doc . --readme', 'Generate project README'],
      ['dardcor doc src/auth.go --type adr', 'Write ADR for auth module'],
      ['dardcor doc . --type changelog', 'Generate CHANGELOG'],
    ],
  },

  plan: {
    usage: 'dardcor plan <description>',
    short: 'Generate a structured execution plan',
    desc: 'Breaks down a task into a numbered, dependency-aware plan.\nUseful before starting a large feature or refactor.',
    flags: [
      ['(no flags)', 'Outputs a full plan with IDs, steps, and risk notes'],
    ],
    examples: [
      ['dardcor plan "add OAuth2 login to this app"', 'Feature plan'],
      ['dardcor plan "migrate from MongoDB to Postgres"', 'Migration plan'],
      ['dardcor plan "set up CI/CD with GitHub Actions"', 'DevOps plan'],
    ],
  },

  refactor: {
    usage: 'dardcor refactor <description>',
    short: 'Intelligently refactor code',
    desc: 'Agent reads the codebase, identifies target files, and applies\nclean refactoring: DRY, SOLID, naming, error handling, TypeScript types.',
    flags: [
      ['(no flags)', 'Full agentic refactor with read/write access'],
    ],
    examples: [
      ['dardcor refactor "extract duplicate API calls into a service"', 'Extract service'],
      ['dardcor refactor "add proper TypeScript types to handlers/"', 'Add types'],
      ['dardcor refactor "replace callbacks with async/await"', 'Modernize'],
    ],
  },

  scaffold: {
    usage: 'dardcor scaffold <description>',
    short: 'Generate project boilerplate and file structure',
    desc: 'Creates all necessary files, configs, and entry points for\na new component, module, or project scaffold.',
    flags: [
      ['(no flags)', 'Writes files directly to current directory'],
    ],
    examples: [
      ['dardcor scaffold "REST API endpoint for /users CRUD"', 'CRUD scaffold'],
      ['dardcor scaffold "React component UserProfile with tests"', 'Component'],
      ['dardcor scaffold "CLI tool with commander.js"', 'New CLI tool'],
    ],
  },

  test: {
    usage: 'dardcor test <description|file>',
    short: 'Generate comprehensive tests',
    desc: 'Pass a file path or description. Generates unit, integration,\nand edge case tests using the project\'s existing test framework.',
    flags: [
      ['--coverage', 'Focus on increasing coverage — adds boundary and error path tests'],
    ],
    examples: [
      ['dardcor test src/auth/jwt.ts', 'Generate tests for file'],
      ['dardcor test "user registration flow"', 'Describe what to test'],
      ['dardcor test src/api.ts --coverage', 'Max coverage mode'],
    ],
  },

  review: {
    usage: 'dardcor review',
    short: 'Run an AI code review on staged/changed code',
    desc: 'Analyzes your git diff for bugs, security issues, performance\nproblems, code quality, missing tests, and docs. Color-coded severity.',
    flags: [
      ['--all', 'Review last commit (git diff HEAD~1)'],
      ['--branch <name>', 'Review diff vs another branch'],
      ['--strict', 'Flag ALL issues including minor style suggestions'],
    ],
    examples: [
      ['dardcor review', 'Review staged changes'],
      ['dardcor review --all', 'Review last commit'],
      ['dardcor review --branch main', 'Compare against main'],
      ['dardcor review --strict', 'Strict review mode'],
    ],
  },

  optimize: {
    usage: 'dardcor optimize <file>',
    short: 'Optimize code for performance',
    desc: 'Reads the target file and suggests or applies improvements:\nBig-O complexity, memory, I/O, caching, parallelization, bundle size.',
    flags: [
      ['(no flags)', 'Outputs analysis + optimized version'],
    ],
    examples: [
      ['dardcor optimize src/indexer.ts', 'Optimize file'],
      ['dardcor optimize services/query.go', 'Optimize Go service'],
    ],
  },

  translate: {
    usage: 'dardcor translate <file>',
    short: 'Translate code to another programming language',
    desc: 'Converts the source file to the target language with idiomatic\npatterns, type annotations, and equivalent error handling.',
    flags: [
      ['--to <language>', 'Target language', 'TypeScript'],
    ],
    examples: [
      ['dardcor translate app.py --to TypeScript', 'Python → TS'],
      ['dardcor translate server.js --to Go', 'JS → Go'],
      ['dardcor translate api.go --to Python', 'Go → Python'],
    ],
  },

  sandbox: {
    usage: 'dardcor sandbox <description>',
    short: 'Run AI code in an isolated sandbox',
    desc: 'Creates a temp directory, writes + executes code inside it,\ncaptures all output, then cleans up. Safe isolated execution.',
    flags: [
      ['(no flags)', 'Creates temp dir, runs, captures output, cleans up'],
    ],
    examples: [
      ['dardcor sandbox "write and run a fibonacci generator in Python"', 'Python sandbox'],
      ['dardcor sandbox "test a regex pattern on sample strings"', 'Regex test'],
    ],
  },

  think: {
    usage: 'dardcor think <prompt>',
    short: 'Activate deep chain-of-thought reasoning mode',
    desc: 'Forces the agent to reason step-by-step before answering.\nIdeal for complex architecture decisions or algorithm design.',
    flags: [
      ['(no flags)', 'Prepends chain-of-thought instructions to your prompt'],
    ],
    examples: [
      ['dardcor think "design a distributed rate limiter"', 'Architecture'],
      ['dardcor think "what are the trade-offs of event sourcing"', 'Analysis'],
      ['dardcor think "optimize this recursive algorithm"', 'Algorithm'],
    ],
    note: 'Keywords like "think", "reason", "step by step" also auto-activate this in any command.',
  },


  commit: {
    usage: 'dardcor commit',
    short: 'Generate an AI-powered git commit message',
    desc: 'Reads staged diff (auto-stages all if nothing staged), sends to AI,\nshows suggested message, confirms before committing.',
    flags: [
      ['--conventional', 'Use conventional commit format: type(scope): description'],
      ['--dry-run', 'Show message only, do not commit'],
    ],
    examples: [
      ['dardcor commit', 'Stage + smart commit'],
      ['dardcor commit --conventional', 'feat(auth): add OAuth2 login'],
      ['dardcor commit --dry-run', 'Preview message without committing'],
    ],
  },

  branch: {
    usage: 'dardcor branch <description>',
    short: 'Create a git branch from a plain-text description',
    desc: 'AI generates a proper kebab-case branch name with prefix,\nthen confirms before creating and checking it out.',
    flags: [
      ['(no flags)', 'Generates name, confirms, then: git checkout -b <name>'],
    ],
    examples: [
      ['dardcor branch "add dark mode to settings page"', '→ feat/add-dark-mode-to-settings'],
      ['dardcor branch "fix null pointer in auth module"', '→ fix/null-pointer-in-auth-module'],
      ['dardcor branch "docs: update API reference"', '→ docs/update-api-reference'],
    ],
  },

  pr: {
    usage: 'dardcor pr',
    short: 'Generate a GitHub Pull Request description',
    desc: 'Compares current branch vs base, reads commits and diff,\ngenerates a formatted PR body with Summary, Changes, Testing sections.',
    flags: [
      ['--base <branch>', 'Base branch to compare against', 'main'],
    ],
    examples: [
      ['dardcor pr', 'Generate PR description'],
      ['dardcor pr --base dev', 'Compare against dev branch'],
    ],
  },

  resolve: {
    usage: 'dardcor resolve',
    short: 'Auto-resolve git merge conflicts with AI',
    desc: 'Detects all conflicted files, reads both sides of each conflict,\nand writes a clean merged version. Review before staging.',
    flags: [
      ['(no flags)', 'Processes all files with <<<<<<< markers'],
    ],
    examples: [
      ['dardcor resolve', 'Auto-resolve all conflict markers'],
    ],
    note: 'Always review resolved files before: git add . && git commit',
  },

  diff: {
    usage: 'dardcor diff [ref]',
    short: 'AI-enhanced diff analysis',
    desc: 'Gets git diff vs ref (default HEAD) and asks AI to summarize\nwhat changed, the impact, and any potential issues.',
    flags: [
      ['[ref]', 'Git ref to diff against (branch, commit, tag)', 'HEAD'],
      ['--risk', 'Include risk assessment (Low / Medium / High / Critical)'],
    ],
    examples: [
      ['dardcor diff', 'Diff vs HEAD'],
      ['dardcor diff main', 'Diff vs main branch'],
      ['dardcor diff abc1234', 'Diff vs specific commit'],
      ['dardcor diff --risk', 'With risk rating'],
    ],
  },


  init: {
    usage: 'dardcor init',
    short: 'Initialize .dardcorrules for your project',
    desc: 'Interactive wizard that creates a .dardcorrules file in the\ncurrent directory. Choose from templates or customize.',
    flags: [
      ['(no flags)', 'Interactive template picker'],
    ],
    templates: ['react', 'nextjs', 'node-api', 'python', 'flutter', 'rust', 'custom'],
    examples: [
      ['dardcor init', 'Start the project init wizard'],
    ],
    note: '.dardcorrules is auto-loaded and injected into the agent system prompt.',
  },

  index: {
    usage: 'dardcor index',
    short: 'Build a searchable code index of the workspace',
    desc: 'Walks all source files, stores a fast-lookup index at\n~/.dardcor/cache/code_index.json for AI-assisted search.',
    flags: [
      ['--status', 'Show current index info without rebuilding'],
      ['--include <pattern>', 'Include glob pattern (e.g. "**/*.go")'],
      ['--exclude <pattern>', 'Exclude glob pattern'],
    ],
    examples: [
      ['dardcor index', 'Build index for current directory'],
      ['dardcor index --status', 'Show index status'],
    ],
  },

  config: {
    usage: 'dardcor config',
    short: 'View or edit Dardcor configuration',
    desc: `Reads/writes ~/.dardcor/config.json.\nKeys: provider, model, api_key, base_url, port.`,
    flags: [
      ['--set <key=value>', 'Set a config key (e.g. --set provider=openai)'],
      ['--themes', 'List available UI themes'],
    ],
    examples: [
      ['dardcor config', 'Show current config'],
      ['dardcor config --set provider=openai', 'Set AI provider'],
      ['dardcor config --set model=gpt-4o', 'Set model'],
      ['dardcor config --set api_key=sk-...', 'Set API key'],
      ['dardcor config --themes', 'List themes'],
    ],
  },

  login: {
    usage: 'dardcor login',
    short: 'Authenticate with an AI provider interactively',
    desc: 'Guided setup wizard — select a provider, enter API key,\nchoose a default model. Saves to ~/.dardcor/config.json.',
    providers: ['antigravity (free)', 'openai', 'anthropic', 'gemini', 'groq', 'deepseek', 'openrouter', 'ollama', 'custom'],
    flags: [
      ['(no flags)', 'Interactive numbered menu'],
    ],
    examples: [
      ['dardcor login', 'Run auth wizard'],
    ],
  },

  models: {
    usage: 'dardcor models',
    short: 'Browse all AI models across every supported provider',
    desc: `Full model directory: 200+ models across 11 providers.\nEach entry shows name, context window, capability tags, and pricing.\n\nProviders: antigravity · openai · anthropic · gemini · groq · deepseek\n           xai · mistral · openrouter · together · fireworks · perplexity\n           cohere · ollama · custom`,
    flags: [
      ['--filter <tag>', 'Show only models with this tag'],
      ['--search <text>', 'Search model name or ID'],
      ['--provider <name>', 'Show only one provider'],
      ['--pricing', 'Show cost per 1M tokens (input↑ output↓)'],
      ['--count', 'Print total model count and exit'],
      ['--all', 'Include custom provider entry'],
    ],
    tags: [
      ['free', 'Zero cost (Antigravity OAuth, Ollama local)'],
      ['fast', 'Optimized for low latency'],
      ['cheap', 'Very low cost per token'],
      ['local', 'Runs on your machine via Ollama'],
      ['vision', 'Supports image/multimodal input'],
      ['code', 'Optimized for coding tasks'],
      ['reason', 'Extended reasoning / chain-of-thought'],
      ['long', 'Very large context window (1M+ tokens)'],
    ],
    examples: [
      ['dardcor models', 'All models'],
      ['dardcor models --filter free', 'Free models only'],
      ['dardcor models --filter code', 'Coding-optimized models'],
      ['dardcor models --filter reason', 'Reasoning models (R1, o3, Gemini thinking)'],
      ['dardcor models --filter local', 'Ollama local models only'],
      ['dardcor models --filter vision', 'Vision/multimodal models'],
      ['dardcor models --provider groq', 'Groq provider only'],
      ['dardcor models --provider ollama', 'All Ollama models'],
      ['dardcor models --search claude', 'Search "claude" in name/id'],
      ['dardcor models --search qwen3', 'Search Qwen3 models'],
      ['dardcor models --pricing', 'Show cost per 1M tokens'],
      ['dardcor models --count', 'Just print total count'],
    ],
  },

  stats: {
    usage: 'dardcor stats',
    short: 'Show usage statistics and estimated cost',
    desc: 'Reads local stats from ~/.dardcor/cache/stats.json.\nShows total requests, tokens, estimated USD cost, and per-provider breakdown.',
    flags: [
      ['(no flags)', 'Display stats summary'],
    ],
    examples: [
      ['dardcor stats', 'Show usage and cost'],
    ],
  },

  history: {
    usage: 'dardcor history',
    short: 'Browse past conversation sessions',
    desc: 'Lists all saved CLI and web sessions with title, date, and message count.\nFiles stored in ~/.dardcor/session/{cli,web}/',
    flags: [
      ['--clear', 'Delete all conversation history'],
      ['--export <id>', 'Export a session (first 8 chars of ID)'],
      ['--format markdown', 'Export format', 'markdown'],
      ['--format json', 'Export as JSON'],
    ],
    examples: [
      ['dardcor history', 'List all sessions'],
      ['dardcor history --clear', 'Delete all history'],
      ['dardcor history --export a1b2c3d4', 'Export session'],
    ],
  },

  resume: {
    usage: 'dardcor resume',
    short: 'Resume the last CLI session',
    desc: 'Loads the most recent CLI session and re-opens the TUI with\nthat conversation context pre-loaded.',
    flags: [
      ['--session <id>', 'Resume a specific session by ID prefix'],
    ],
    examples: [
      ['dardcor resume', 'Resume latest session'],
      ['dardcor resume --session a1b2c3d4', 'Resume specific session'],
    ],
  },

  mcp: {
    usage: 'dardcor mcp <action> [name] [command] [args...]',
    short: 'Manage Model Context Protocol (MCP) servers',
    desc: 'MCP servers extend the agent with external tools like web search,\ndoc lookup, file access, and more. Config at ~/.dardcor/mcp.json.',
    actions: {
      list: 'List all configured MCP servers',
      add: 'Add a server: dardcor mcp add <name> <command> [args...]',
      remove: 'Remove a server: dardcor mcp remove <name>',
      builtin: 'Show built-in MCP servers available to install',
    },
    flags: [
      ['list', 'Show all configured servers'],
      ['add <n> <cmd>', 'Register a new MCP server'],
      ['remove <name>', 'Remove a server by name'],
      ['builtin', 'List installable built-in servers'],
    ],
    examples: [
      ['dardcor mcp list', 'Show servers'],
      ['dardcor mcp builtin', 'Browse built-ins'],
      ['dardcor mcp add context7 npx -y @upstash/context7-mcp@latest', 'Add context7'],
      ['dardcor mcp add websearch npx -y @modelcontextprotocol/server-brave-search', 'Brave search'],
      ['dardcor mcp add fs npx -y @modelcontextprotocol/server-filesystem .', 'Filesystem'],
      ['dardcor mcp remove context7', 'Remove server'],
    ],
  },

  plugin: {
    usage: 'dardcor plugin <action> [name]',
    short: 'Manage Dardcor plugins',
    desc: 'Plugins are npm packages that extend Dardcor with new commands,\nhooks, or tools. Installed to ~/.dardcor/plugins/',
    flags: [
      ['list', 'List all installed plugins'],
      ['add <package>', 'Install an npm plugin package'],
      ['remove <package>', 'Uninstall a plugin'],
      ['create <name>', 'Scaffold a new plugin in ~/.dardcor/plugins/<name>'],
    ],
    examples: [
      ['dardcor plugin list', 'Show installed plugins'],
      ['dardcor plugin add dardcor-plugin-docker', 'Install Docker plugin'],
      ['dardcor plugin remove dardcor-plugin-docker', 'Remove plugin'],
      ['dardcor plugin create my-plugin', 'Create plugin scaffold'],
    ],
  },
};


export function printCommandHelp(cmdName) {
  const h = COMMAND_HELP[cmdName];
  if (!h) {
    console.log(`\n  ${C.yellow}Unknown command:${C.reset} ${cmdName}\n  Run ${C.cyan}dardcor help${C.reset} for the full list.\n`);
    return;
  }

  console.log('');
  console.log(`${C.bold}${C.purple}${h.usage}${C.reset}`);
  console.log(`  ${C.dim}${h.short}${C.reset}`);
  console.log('');
  console.log(hr());


  console.log(`\n${C.bold}DESCRIPTION${C.reset}`);
  h.desc.split('\n').forEach(l => console.log(`  ${l}`));


  if (h.flags && h.flags.length > 0) {
    console.log(`\n${C.bold}OPTIONS${C.reset}`);
    h.flags.forEach(([f, d, def]) => console.log(flag(f, d, def)));
  }


  if (h.actions) {
    console.log(`\n${C.bold}ACTIONS${C.reset}`);
    Object.entries(h.actions).forEach(([a, d]) => {
      console.log(`  ${C.cyan}${a.padEnd(12)}${C.reset}  ${C.dim}${d}${C.reset}`);
    });
  }


  if (h.templates) {
    console.log(`\n${C.bold}TEMPLATES${C.reset}  ${C.dim}(choose during init)${C.reset}`);
    h.templates.forEach(t => console.log(`  ${C.dim}•${C.reset} ${t}`));
  }


  if (h.tags) {
    console.log(`\n${C.bold}TAGS${C.reset}`);
    h.tags.forEach(([t, d]) => console.log(`  ${C.cyan}${t.padEnd(10)}${C.reset}  ${C.dim}${d}${C.reset}`));
  }


  if (h.providers) {
    console.log(`\n${C.bold}PROVIDERS${C.reset}`);
    h.providers.forEach(p => console.log(`  ${C.dim}•${C.reset} ${p}`));
  }


  if (h.examples && h.examples.length > 0) {
    console.log(`\n${C.bold}EXAMPLES${C.reset}`);
    h.examples.forEach(([code, cmt]) => console.log(example(code, cmt)));
  }


  if (h.note) {
    console.log('');
    console.log(note(h.note));
  }

  console.log('');
}


export function printHelp() {
  printBanner();

  const w = Math.min(cols(), 90);


  console.log(`${C.bold}USAGE${C.reset}`);
  console.log(`  ${C.cyan}dardcor${C.reset} ${C.dim}<command> [options]${C.reset}`);
  console.log(`  ${C.cyan}dardcor${C.reset} ${C.dim}<command> --help${C.reset}       ${C.gray}# detailed help for any command${C.reset}`);
  console.log('');

  console.log(`  ${C.purple}${C.bold}Core:${C.reset}`);
  console.log(`    ${C.cyan}cli${C.reset}        Start interactive agent mode`);
  console.log(`    ${C.cyan}auto${C.reset}       Start interactive mode (subcommands: ${C.dim}yolo, safe${C.reset})`);
  console.log(`    ${C.cyan}auth${C.reset}       Manage authentication (subcommands: ${C.dim}login, logout${C.reset})`);
  console.log(`    ${C.cyan}run${C.reset}        Run the web dashboard server`);
  console.log(`    ${C.cyan}doctor${C.reset}     Diagnose installation issues`);
  console.log('');


  const sec = (title, emoji) =>
    `\n${C.bold}${C.purple}${emoji}  ${title}${C.reset}\n${hr()}`;

  const row = (name, flags, desc) => {
    const namePart = `  ${C.cyan}${C.bold}${name}${C.reset}`;
    const flagPart = flags ? ` ${C.gray}${flags}${C.reset}` : '';
    const left = `${namePart}${flagPart}`;
    const leftLen = name.length + (flags ? flags.length + 1 : 0) + 2;
    const pad = Math.max(1, 42 - leftLen);
    const descPart = `${C.dim}${desc}${C.reset}`;
    return `${left}${' '.repeat(pad)}${descPart}`;
  };


  console.log(sec('Core', '⚙️ '));
  console.log(row('run', '', 'Start Dashboard UI + Go backend engine (port 25000)'));
  console.log(row('cli', '', 'Interactive full-screen TUI agent'));
  console.log(row('doctor', '', 'Auto-diagnose & repair system issues'));
  console.log(row('help', '[command]', 'Show this help (or per-command help)'));


  console.log(sec('AI Commands', '🤖'));
  console.log(row('ask', '<prompt>', 'One-shot AI question'));
  console.log(row('fix', '<description>', 'Find and fix a bug (full agentic loop)'));
  console.log(row('debug', '<description>', 'Deep debug: root cause + fix + prevention'));
  console.log(row('explain', '<file|snippet>', 'Explain code  [--depth shallow|normal|deep]'));
  console.log(row('doc', '<file>', 'Generate docs [--readme | --type jsdoc|adr]'));
  console.log(row('plan', '<description>', 'Create a step-by-step execution plan'));
  console.log(row('refactor', '<description>', 'Intelligent code refactoring'));
  console.log(row('scaffold', '<description>', 'Generate boilerplate + project structure'));
  console.log(row('test', '<file|desc>', 'Generate tests  [--coverage]'));
  console.log(row('review', '', 'AI code review  [--all | --branch | --strict]'));
  console.log(row('optimize', '<file>', 'Performance optimization analysis'));
  console.log(row('translate', '<file>', 'Translate to another language [--to <lang>]'));
  console.log(row('sandbox', '<description>', 'Isolated temp-dir execution'));
  console.log(row('think', '<prompt>', 'Force chain-of-thought reasoning'));


  console.log(sec('Git Commands', '🌿'));
  console.log(row('commit', '', 'AI commit message  [--conventional | --dry-run]'));
  console.log(row('branch', '<description>', 'Create branch from description'));
  console.log(row('pr', '', 'Generate GitHub PR description  [--base <branch>]'));
  console.log(row('resolve', '', 'Auto-resolve all merge conflicts'));
  console.log(row('diff', '[ref]', 'AI diff analysis  [--risk]'));


  console.log(sec('Project Commands', '📦'));
  console.log(row('init', '', 'Create .dardcorrules  (wizard)'));
  console.log(row('index', '', 'Build code index  [--status]'));
  console.log(row('config', '', 'View/edit config  [--set key=val | --themes]'));
  console.log(row('login', '', 'Authenticate with an AI provider (wizard)'));
  console.log(row('models', '', 'List all models per provider'));
  console.log(row('stats', '', 'Usage statistics & estimated cost'));
  console.log(row('history', '', 'Browse past sessions  [--clear]'));
  console.log(row('resume', '', 'Resume last session  [--session <id>]'));


  console.log(sec('Extensibility', '🧩'));
  console.log(row('mcp', '<action>', 'MCP servers  list|add|remove|builtin'));
  console.log(row('plugin', '<action>', 'Plugins       list|add|remove|create'));


  console.log(`\n${C.bold}${C.purple}🔑  Providers & Models${C.reset}\n${hr()}`);
  const providerRows = [
    ['antigravity', '✨ FREE', 'Gemini 2.5 Pro/Flash/Lite, 2.0 Flash (7 models)', 'OAuth only — no key needed'],
    ['openai', '', 'GPT-4.1/4o, o4-mini, o3, o1, gpt-3.5 (16 models)', 'platform.openai.com'],
    ['anthropic', '', 'Claude Opus/Sonnet/Haiku 4.6/4.5/4/3.7/3.5 (13 mod)', 'console.anthropic.com'],
    ['gemini', '', 'Gemini 2.5 Pro/Flash/Lite, 2.0, 1.5 (9 models)', 'aistudio.google.com'],
    ['groq', '⚡ FAST', 'Llama4, Llama3.x, Kimi K2, Qwen3, Mixtral (16 mod)', 'console.groq.com'],
    ['deepseek', '💰 CHEAP', 'V3 Chat, R1 Reasoner, Coder (3 models)', 'platform.deepseek.com'],
    ['xai', '', 'Grok 3/3-mini/3-fast, Grok 2 Vision (6 models)', 'console.x.ai'],
    ['mistral', '', 'Large/Medium/Small, Codestral, Ministral (13 mod)', 'console.mistral.ai'],
    ['openrouter', '', '24+ frontier + open-source via one API key', 'openrouter.ai'],
    ['together', '💰 CHEAP', 'Llama4, Qwen3 235B, DeepSeek R1/V3 (9 models)', 'api.together.ai'],
    ['fireworks', '⚡ FAST', 'Llama3.x, DeepSeek, Kimi K2, Qwen3 (7 models)', 'app.fireworks.ai'],
    ['perplexity', '🔍 SEARCH', 'Sonar Pro/Reasoning with web search (6 models)', 'perplexity.ai/settings/api'],
    ['cohere', '', 'Command R+, Command A, Command R 7B (4 models)', 'dashboard.cohere.ai'],
    ['ollama', '🏠 LOCAL', 'Qwen2.5-coder, Llama3, DeepSeek-R1, Gemma3 (30+)', 'ollama.com — runs locally'],
    ['custom', '', 'Any OpenAI-compatible endpoint', 'set base_url manually'],
  ];
  providerRows.forEach(([name, badge, models, url]) => {
    const b = badge ? `${C.yellow}${badge.padEnd(10)}${C.reset}` : ' '.repeat(10);
    const n = `${C.cyan}${name.padEnd(12)}${C.reset}`;
    const m = `${C.dim}${models}${C.reset}`;
    console.log(`  ${n} ${b} ${m}`);
    console.log(`  ${' '.repeat(12)}  ${C.gray}${url}${C.reset}`);
  });
  console.log(`\n  ${C.dim}Full list with pricing: ${C.reset}${C.cyan}dardcor models${C.reset}  ${C.dim}| filter: ${C.reset}${C.cyan}dardcor models --filter free${C.reset}`);


  console.log(`\n${C.bold}${C.purple}🏷️   Model Tags${C.reset}\n${hr()}`);
  const tagRows = [
    ['free', C.green, 'Zero cost — Antigravity (Google OAuth) or Ollama local'],
    ['fast', C.cyan, 'Low latency, optimized for speed (Groq, Fireworks)'],
    ['cheap', C.yellow, 'Very low cost per token'],
    ['local', C.yellow, 'Runs entirely on your machine via Ollama'],
    ['vision', C.blue, 'Accepts image/multimodal input'],
    ['code', '\x1b[38;5;93m', 'Fine-tuned / excels at coding tasks'],
    ['reason', '\x1b[38;5;208m', 'Extended reasoning / chain-of-thought (o3, R1, Gemini think)'],
    ['long', '\x1b[38;5;43m', 'Very large context window (1M+ tokens)'],
  ];
  tagRows.forEach(([t, col, desc]) =>
    console.log(`  ${col}${t.padEnd(8)}${C.reset}  ${C.dim}${desc}${C.reset}`)
  );


  console.log(`\n${C.bold}${C.purple}📋  Project Rules (.dardcorrules)${C.reset}\n${hr()}`);
  console.log(`  ${C.dim}Create a${C.reset} ${C.cyan}.dardcorrules${C.reset} ${C.dim}file to customize agent behavior per project.${C.reset}`);
  console.log(`  ${C.dim}Also recognized:${C.reset} ${C.gray}AGENTS.md  CLAUDE.md  .cursorrules${C.reset}`);
  console.log('');
  console.log(`  ${C.bold}Templates available:${C.reset}  react · nextjs · node-api · python · flutter · rust`);
  console.log(`  ${C.cyan}dardcor init${C.reset}  ${C.dim}→ interactive template wizard${C.reset}`);
  console.log('');
  console.log(`  ${C.dim}Example .dardcorrules:${C.reset}`);
  console.log(`  ${C.gray}## Tech Stack${C.reset}`);
  console.log(`  ${C.gray}Framework: Next.js 14, TypeScript, Tailwind CSS${C.reset}`);
  console.log(`  ${C.gray}## Conventions${C.reset}`);
  console.log(`  ${C.gray}NEVER use any type — use unknown and narrow${C.reset}`);
  console.log(`  ${C.gray}All API routes must validate with Zod${C.reset}`);


  console.log(`\n${C.bold}${C.purple}🔌  MCP Servers (Model Context Protocol)${C.reset}\n${hr()}`);
  console.log(`  ${C.dim}MCP extends the agent with external tools — web search, docs, GitHub, etc.${C.reset}`);
  console.log('');
  const mcpBuiltins = [
    ['context7', 'npx -y @upstash/context7-mcp@latest', 'Up-to-date library docs'],
    ['grep-app', 'npx -y @grep-app/mcp@latest', 'Web code search'],
    ['websearch', 'npx -y @modelcontextprotocol/server-brave-search', 'Web search (Brave API)'],
    ['filesystem', 'npx -y @modelcontextprotocol/server-filesystem .', 'Local file access'],
    ['github', 'npx -y @modelcontextprotocol/server-github', 'GitHub API'],
    ['memory', 'npx -y @modelcontextprotocol/server-memory', 'Persistent agent memory'],
  ];
  mcpBuiltins.forEach(([name, cmd, desc]) => {
    console.log(`  ${C.cyan}${name.padEnd(12)}${C.reset}  ${C.dim}${desc}${C.reset}`);
    console.log(`  ${C.gray}             dardcor mcp add ${name} ${cmd}${C.reset}`);
  });
  console.log(`\n  ${C.dim}Show all:${C.reset} ${C.cyan}dardcor mcp builtin${C.reset}   ${C.dim}Manage:${C.reset} ${C.cyan}dardcor mcp list${C.reset}`);


  console.log(`\n${C.bold}${C.purple}🎨  UI Themes${C.reset}\n${hr()}`);
  const themeList = [
    ['dracula', '#bd93f9 purple  · vampiric dark'],
    ['catppuccin', '#cba6f7 mocha   · soft pastel dark'],
    ['tokyo-night', '#7aa2f7 blue    · neon city night'],
    ['nord', '#88c0d0 ice     · arctic minimal'],
    ['gruvbox', '#b8bb26 gold    · retro warm'],
    ['monokai', '#f92672 pink    · classic editor dark'],
  ];
  themeList.forEach(([name, desc]) =>
    console.log(`  ${C.cyan}${name.padEnd(14)}${C.reset}  ${C.dim}${desc}${C.reset}`)
  );
  console.log(`\n  ${C.dim}Set theme:${C.reset} ${C.cyan}dardcor config --set theme=dracula${C.reset}`);


  console.log(`\n${C.bold}${C.purple}🌐  Environment & Paths${C.reset}\n${hr()}`);
  const envRows = [
    ['Dashboard', 'http://127.0.0.1:25000', C.cyan],
    ['Config', '~/.dardcor/config.json', C.dim],
    ['Account', '~/.dardcor/account.json', C.dim],
    ['Sessions', '~/.dardcor/session/', C.dim],
    ['Cache/Stats', '~/.dardcor/cache/', C.dim],
    ['Plugins', '~/.dardcor/plugins/', C.dim],
    ['MCP config', '~/.dardcor/mcp.json', C.dim],
    ['Project Config', '.dardcor.json', C.cyan],
    ['Rules file', '.dardcorrules / AGENTS.md / CLAUDE.md', C.dim],
    ['Port', '25000  (Go backend)', C.gray],
  ];
  envRows.forEach(([label, val, col]) =>
    console.log(`  ${C.dim}${label.padEnd(16)}→${C.reset}  ${col}${val}${C.reset}`)
  );


  console.log(`\n${C.bold}${C.purple}⚙️   Environment Variables${C.reset}\n${hr()}`);
  const envVars = [
    ['DARDCOR_AI_PROVIDER', 'Override provider (openai, anthropic, groq, …)'],
    ['DARDCOR_AI_MODEL', 'Override model ID'],
    ['DARDCOR_API_KEY', 'Override API key'],
    ['DARDCOR_BASE_URL', 'Override provider base URL'],
    ['DARDCOR_DATA_DIR', 'Override database directory'],
    ['DARDCOR_DEV', 'Set to "true" for Vite dev proxy mode'],
    ['PORT', 'Override server port (default: 25000)'],
    ['DEBUG', 'Set to 1 to print stack traces on error'],
  ];
  envVars.forEach(([v, d]) =>
    console.log(`  ${C.green}${v.padEnd(24)}${C.reset}  ${C.dim}${d}${C.reset}`)
  );


  console.log(`\n${C.bold}${C.purple}🚀  Quick Start Guide${C.reset}\n${hr()}`);
  const qs = [
    ['npm install -g dardcor-agent', '1. Install globally'],
    ['dardcor doctor', '2. Verify system requirements'],
    ['dardcor login', '3. Choose & authenticate a provider'],
    ['dardcor run', '4. Launch dashboard (http://localhost:25000)'],
    ['dardcor ask "what can you do?"', '5. First question via CLI'],
    ['dardcor init', '6. Set up project rules in your repo'],
  ];
  qs.forEach(([code, cmt]) => console.log(example(code, cmt)));


  console.log(`\n${C.bold}${C.purple}💡  Examples by Use Case${C.reset}\n${hr()}`);

  const groups = [
    ['Coding & Debugging', [
      ['dardcor fix "TypeError: cannot read property of undefined"', 'Fix runtime error'],
      ['dardcor debug "memory leak in worker pool"', 'Root cause analysis'],
      ['dardcor optimize src/query.go', 'Speed up slow code'],
      ['dardcor refactor "extract service layer from controllers"', 'Refactor codebase'],
    ]],
    ['Git & Code Review', [
      ['dardcor commit --conventional', 'AI commit message'],
      ['dardcor review --strict', 'Full code review'],
      ['dardcor pr --base main', 'Generate PR description'],
      ['dardcor branch "add rate limiting to API"', 'Smart branch name'],
      ['dardcor resolve', 'Auto-resolve conflicts'],
    ]],
    ['Scaffolding & Documentation', [
      ['dardcor scaffold "React hook for infinite scroll"', 'Generate boilerplate'],
      ['dardcor doc src/api.ts --readme', 'Write README'],
      ['dardcor test src/auth.go --coverage', 'Write unit tests'],
      ['dardcor translate server.py --to TypeScript', 'Convert language'],
    ]],
    ['AI & Models', [
      ['dardcor models', 'Browse all 200+ models'],
      ['dardcor models --filter free', 'Free models'],
      ['dardcor models --filter reason --pricing', 'Reasoning models + cost'],
      ['dardcor think "design a distributed lock system"', 'Deep reasoning mode'],
      ['dardcor login', 'Switch AI provider'],
    ]],
    ['Project Setup', [
      ['dardcor init', 'Create .dardcorrules'],
      ['dardcor index', 'Build codebase index'],
      ['dardcor mcp add context7 npx -y @upstash/context7-mcp', 'Add MCP server'],
      ['dardcor index && dardcor ask "find all auth endpoints"', 'Index + smart search'],
    ]],
  ];

  groups.forEach(([title, cmds]) => {
    console.log(`  ${C.bold}${title}${C.reset}`);
    cmds.forEach(([code, cmt]) => console.log(example(code, cmt)));
    console.log('');
  });


  console.log(hr());
  console.log(`${C.bold}${C.purple}ℹ️   Per-command help:${C.reset}  ${C.cyan}dardcor <command> --help${C.reset}`);
  console.log(`${C.dim}    e.g.  dardcor fix --help   dardcor commit --help   dardcor models --help${C.reset}`);
  console.log(`${C.bold}    Version:${C.reset} ${C.dim}v${VERSION}${C.reset}`);
  console.log('');
}
