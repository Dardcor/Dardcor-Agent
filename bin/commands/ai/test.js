import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';
import fs from 'fs';

export async function handleTest(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Generating tests for: ${C.dim}${description.slice(0, 60)}...${C.reset}`);

  let context = '';
  if (fs.existsSync(description)) {
    try { context = `\n\nSource file:\n${fs.readFileSync(description, 'utf8')}`; } catch {}
  }

  const coverageNote = opts.coverage ? '\n\nFocus on increasing code coverage. Add edge cases, boundary conditions, and error path tests.' : '';
  const prompt = `Generate comprehensive tests for: ${description}${context}${coverageNote}\n\nInclude:\n- Unit tests for each function\n- Integration tests\n- Edge cases\n- Error handling tests\n- Mock/stub examples where needed\n\nUse the appropriate testing framework for the detected language.`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
