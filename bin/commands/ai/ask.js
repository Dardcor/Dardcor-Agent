import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';

export async function handleAsk(prompt, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Asking agent: ${C.dim}${prompt.slice(0, 60)}...${C.reset}`);
  const res = await sendToAgent(prompt, opts.convId ? { conversation_id: opts.convId } : {});
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
