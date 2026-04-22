import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';
import fs from 'fs';

export async function handleExplain(target, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Explaining: ${C.dim}${target}${C.reset}`);

  let content = target;
  if (fs.existsSync(target)) {
    try { content = fs.readFileSync(target, 'utf8'); } catch {}
  }

  const depth = opts.depth || 'normal';
  const depthInstructions = {
    shallow: 'Give a concise high-level overview only.',
    normal: 'Explain what it does, how it works, and any notable patterns.',
    deep: 'Provide a comprehensive deep-dive: architecture, patterns, algorithms, edge cases, performance, and improvement suggestions.',
  };

  const prompt = `Explain this code/concept: ${content}\n\nDepth: ${depthInstructions[depth] || depthInstructions.normal}`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
