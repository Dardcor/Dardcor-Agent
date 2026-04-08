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
};

export function printBanner(subtitle = '') {
  console.log(`
${C.magenta}${C.bold}
  ██████╗  █████╗ ██████╗ ██████╗  ██████╗ ██████╗ ██████╗
  ██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔═══██╗██╔══██╗
  ██║  ██║███████║██████╔╝██║  ██║██║     ██║   ██║██████╔╝
  ██║  ██║██╔══██║██╔══██╗██║  ██║██║     ██║   ██║██╔══██╗
  ██████╔╝██║  ██║██║  ██║██████╔╝╚██████╗╚██████╔╝██║  ██║
  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝
${C.reset}${C.dim}  Your Personal AI Agent — Any OS, Any Platform.${C.reset}${subtitle ? `\n  ${C.cyan}${subtitle}${C.reset}` : ''}
`);
}

export function printHelp() {
  printBanner();
  console.log(`${C.bold}USAGE:${C.reset}

  ${C.cyan}dardcor help${C.reset}              Show this help message
  ${C.cyan}dardcor doctor${C.reset}            Health check & auto-repair system
  ${C.cyan}dardcor run${C.reset}               Start Gateway + WebUI (Live Updates)
  ${C.cyan}dardcor cli${C.reset}               Interactive terminal coding agent

${C.bold}INSTALLATION:${C.reset}

  git clone https://github.com/dardcor/dardcor-agent.git
  cd dardcor-agent
  npm i -g dardcor-agent

${C.bold}EXAMPLES:${C.reset}

  dardcor run                    Launch Dashboard with live preview
  dardcor cli                    Headless coding agent
  dardcor doctor                 Fix system issues

${C.bold}DOCS:${C.reset}  https://github.com/dardcor/dardcor-agent
`);
}
