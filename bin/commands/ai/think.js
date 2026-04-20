import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleThink(prompt, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Deep thinking mode activated...`);
  const thinkPrompt = `[THOUGHT] Think step-by-step deeply about this problem before answering. Consider all angles, edge cases, potential issues, and alternative approaches.\n\n${prompt}\n\nProvide your complete reasoning chain, then your final answer.`;
  const res = await sendToAgent(thinkPrompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
