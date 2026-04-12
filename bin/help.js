const C = {
  reset:   '\x1b[0m',
  bold:    '\x1b[1m',
  dim:     '\x1b[2m',
  red:     '\x1b[31m',
  green:   '\x1b[32m',
  yellow:  '\x1b[33m',
  blue:    '\x1b[34m',
  magenta: '\x1b[35m',
  cyan:    '\x1b[36m',
  white:   '\x1b[37m',
  bgMagenta: '\x1b[45m',
  bgBlue:    '\x1b[44m',
  purple:  '\x1b[38;5;93m',
};

const banner = [
  '  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ',
  '  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗',
  '  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝',
  '  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗',
  '  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║',
  '  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝'
];

export function printBanner(subtitle = '') {
  console.log(`${C.purple}${C.bold}`);
  banner.forEach(l => console.log(l));
  console.log(`${C.reset}${C.bold}${C.purple}  DARDCOR AGENT${C.reset} — ${C.dim}Superior Autonomous System${C.reset}${subtitle ? `\n  ${C.purple}→ ${C.bold}${subtitle}${C.reset}` : ''}\n`);
}

export function printHelp() {
  printBanner();
  console.log(`${C.bold}USAGE:${C.reset}

  ${C.cyan}dardcor run${C.reset}               Menjalankan Dashboard UI (Auto Real-time)
  ${C.cyan}dardcor cli${C.reset}               Menjalankan Terminal Agent Dardcor
  ${C.cyan}dardcor doctor${C.reset}            Memperbaiki system secara otomatis
  ${C.cyan}dardcor help${C.reset}              Menampilkan bantuan system

${C.bold}COMMANDS:${C.reset}

  dardcor run     (Port: 25000, Auto-Detects Dev/Source Mode)
  dardcor cli     (Port: 25000, Terminal Interface)
  dardcor doctor  (Repair system)
  dardcor help    (Show this help)

${C.bold}INSTALLATION:${C.reset}

  npm link        (Menghubungkan project ke command 'dardcor' sistem)
`);
}
