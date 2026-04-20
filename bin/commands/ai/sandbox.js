import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleSandbox(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Running in sandbox: ${C.dim}${description.slice(0, 60)}...${C.reset}`);
  const prompt = `Run this in sandbox isolation: ${description}\n\n[THOUGHT] Create a temporary workspace, execute the task safely, capture all output, then clean up.\n\nSteps:\n1. Create a temp directory\n2. Write any needed files there\n3. Execute the code/commands\n4. Capture and display all output\n5. Clean up temp files\n\nUse [ACTION] tags for all file/command operations.`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
