import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleRefactor(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Refactoring: ${C.dim}${description.slice(0, 60)}...${C.reset}`);
  const prompt = `Refactor: ${description}\n\nApply these improvements:\n- Extract reusable functions/modules\n- Improve naming clarity\n- Reduce code duplication (DRY)\n- Improve error handling\n- Add TypeScript types where missing\n- Follow SOLID principles\n\nUse [ACTION] tags to read current files and write refactored versions.`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
