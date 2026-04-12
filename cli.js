#!/usr/bin/env node
import { run } from './bin/run.js';
import { runCLI } from './bin/cli_agent.js';
import { runDoctor } from './bin/doctor.js';
import { printHelp } from './bin/help.js';

const [command] = process.argv.slice(2);

async function start() {
    try {
        if (command === 'run') {
            await run();
        } else if (command === 'cli') {
            await runCLI();
        } else if (command === 'doctor') {
            await runDoctor();
        } else {
            printHelp();
        }
    } catch (e) {
        console.error(`\n[!] DARDCOR CRITICAL ERROR: ${e.message}`);
        console.error(e.stack);
        process.exit(1);
    }
}

await start();
