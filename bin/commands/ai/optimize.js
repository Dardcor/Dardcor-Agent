import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';
import fs from 'fs';

export async function handleOptimize(target, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Optimizing: ${C.dim}${target}${C.reset}`);

  let content = target;
  if (fs.existsSync(target)) {
    try { content = fs.readFileSync(target, 'utf8'); } catch {}
  }

  const prompt = `Optimize the performance of: ${target}\n\n${fs.existsSync(target) ? `Current code:\n\`\`\`\n${content}\n\`\`\`` : ''}\n\nFocus on:\n1. Algorithm complexity (Big-O improvements)\n2. Memory usage reduction\n3. I/O bottlenecks\n4. Caching opportunities\n5. Parallelization potential\n6. Database query optimization\n7. Bundle size (if frontend)\n\nProvide before/after comparisons with complexity analysis.`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
