import { sendToAgent, ensureEngine, inf, err, C, printBox } from '../shared.js';

export async function handlePlan(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Generating plan: ${C.dim}${description.slice(0, 60)}...${C.reset}`);
  const prompt = `[PLAN] Create a detailed execution plan for: ${description}\n\nProvide:\n1. Goal breakdown\n2. Step-by-step tasks with IDs\n3. Dependencies between tasks\n4. Risk assessment\n5. Estimated complexity\n\nFormat as a structured plan with checkboxes.`;
  const res = await sendToAgent(prompt);
  printBox(res, 'Execution Plan');
}
