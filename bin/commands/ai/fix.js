import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleFix(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Analyzing bug: ${C.dim}${description.slice(0, 60)}...${C.reset}`);
  const prompt = `Fix this bug/issue: ${description}\n\nAnalyze the codebase, find the root cause, and implement a fix. Read relevant files, trace the error path, understand the issue, then fix it. Use [ACTION] tags to execute file reads/writes.`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
