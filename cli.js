#!/usr/bin/env node

import { run as startRun } from './bin/run.js';
import { runCLI } from './bin/cli_agent.js';
import { runDoctor } from './bin/doctor.js';
import { printHelp } from './bin/help.js';

const args = process.argv.slice(2);
const command = args[0];

if (!command || command === 'help' || command === '--help' || command === '-h') {
  printHelp();
  process.exit(0);
}

switch (command) {
  case 'run':
    startRun();
    break;
  case 'cli':
    runCLI();
    break;
  case 'doctor':
    runDoctor();
    break;
  default:
    console.log(`Unknown command: ${command}`);
    printHelp();
    process.exit(1);
}
