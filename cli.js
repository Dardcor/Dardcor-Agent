#!/usr/bin/env node
import { run } from './bin/run.js';
import { runCLI } from './bin/cli_agent.js';
import { runDoctor } from './bin/doctor.js';
import { printHelp } from './bin/help.js';

const [command] = process.argv.slice(2);

switch (command) {
  case 'run':    run(); break;
  case 'cli':    runCLI(); break;
  case 'doctor': runDoctor(); break;
  default:       printHelp(); break;
}
