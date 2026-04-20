import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleScaffold(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Scaffolding: ${C.dim}${description.slice(0, 60)}...${C.reset}`);
  const prompt = `Generate production-ready boilerplate/scaffold for: ${description}\n\nCreate all necessary files with:\n- Proper project structure\n- Configuration files\n- Entry points\n- Example implementations\n- Package dependencies list\n\nUse [ACTION] write commands to create each file.`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
