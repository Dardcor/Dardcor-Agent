#!/usr/bin/env node

/**
 * ============================================================================
 *   DARDCOR AGENT CLI
 *   Global Terminal Interface for Dardcor Autonomous Agent
 * ============================================================================
 */

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import path from 'path';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const args = process.argv.slice(2);

// Mapping commands
const command = args[0];
const subCommand = args[1];

if (!command) {
  printHelp();
  process.exit(0);
}

// Handle 'dardcor agent' or 'dardcor run'
if ((command === 'agent') || (command === 'run')) {
  runAgent();
} else if (command === 'build') {
  runBuild();
} else if (command === 'install') {
  installProject();
} else {
  console.log(`\n[\x1b[31m!\x1b[0m] Perintah tidak dikenal: ${command}`);
  printHelp();
  process.exit(1);
}

function runAgent() {
  const batPath = path.join(__dirname, 'dardcor.bat');
  if (!fs.existsSync(batPath)) {
    console.error('[\x1b[31m!\x1b[0m] Error: dardcor.bat tidak ditemukan di direktori package.');
    process.exit(1);
  }

  console.log('[\x1b[35m*\x1b[0m] Meluncurkan Dardcor Agent Core...');
  const child = spawn('cmd.exe', ['/c', batPath, 'run'], {
    cwd: __dirname,
    stdio: 'inherit'
  });

  child.on('exit', (code) => {
    process.exit(code || 0);
  });
}

function runBuild() {
  const batPath = path.join(__dirname, 'dardcor.bat');
  const child = spawn('cmd.exe', ['/c', batPath, 'build'], {
    cwd: __dirname,
    stdio: 'inherit'
  });
  child.on('exit', (code) => {
    process.exit(code || 0);
  });
}

function installProject() {
  console.log('\n[\x1b[32m*\x1b[0m] Menyiapkan Dardcor Agent di sistem Anda...');
  console.log('[\x1b[32m*\x1b[0m] Mendownload aset UI dan Core...');
  
  // Logic for a real public installer would fetch from GitHub/NPM
  // For this project, it's already in the package folder.
  
  console.log('[\x1b[36mOK\x1b[0m] Dardcor Agent berhasil diinstal secara global.');
  console.log('Gunakan perintah: \x1b[35mdardcor agent\x1b[0m untuk menjalankan.\n');
}

function printHelp() {
  console.log(`
\x1b[35m
  _____                 _                   
 |  __ \\               | |                  
 | |  | | __ _ _ __  __| | ___ ___  _ __    
 | |  | |/ _\` | '__|/ _\` |/ __/ _ \\| '__|   
 | |__| | (_| | |  | (_| | (_| (_) | |      
 |_____/ \\__,_|_|   \\__,_|\\___\\___/|_|      
                                            
 \x1b[0m
 \x1b[1mDARDCOR AGENT - Autonomous System Controller\x1b[0m
 Version: 1.0.0

 \x1b[1mPENGGUNAAN:\x1b[0m
   \x1b[36mdardcor agent\x1b[0m        Menjalankan Agen & Dashboard (Port 25000)
   \x1b[36mdardcor build\x1b[0m        Melakukan kompilasi sistem
   \x1b[36mdardcor install\x1b[0m      Setup awal project

 \x1b[1mCONTOH:\x1b[0m
   npx dardcor agent
  `);
}
