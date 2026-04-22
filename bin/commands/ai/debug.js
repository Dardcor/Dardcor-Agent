import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleDebug(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Starting debug session: ${C.dim}${description.slice(0, 60)}...${C.reset}`);
  const prompt = `Debug this issue: ${description}\n\n[THOUGHT] Analyze the codebase deeply. Read relevant files, trace execution paths, identify the root cause.\n\nProvide:\n1. Root cause analysis\n2. Reproduction steps\n3. Proposed fix with code\n4. Prevention recommendations`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
