import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';
import fs from 'fs';

export async function handleTranslate(file, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  const to = opts.to || 'TypeScript';
  inf(`Translating ${C.dim}${file}${C.reset} → ${C.bold}${to}${C.reset}`);

  let content = file;
  if (fs.existsSync(file)) {
    try { content = fs.readFileSync(file, 'utf8'); } catch {}
  }

  const prompt = `Translate this code to ${to}:\n\n\`\`\`\n${content}\n\`\`\`\n\nRequirements:\n- Preserve all functionality exactly\n- Use idiomatic ${to} patterns\n- Add appropriate type annotations\n- Maintain comments and documentation\n- Handle language-specific differences (error handling, async patterns, etc.)\n- Note any features that don't have direct equivalents`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
