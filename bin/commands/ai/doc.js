import { sendToAgent, ensureEngine, inf, err, C } from '../shared.js';
import fs from 'fs';

export async function handleDoc(target, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf(`Generating docs for: ${C.dim}${target}${C.reset}`);

  let content = target;
  if (fs.existsSync(target)) {
    try { content = fs.readFileSync(target, 'utf8'); } catch {}
  }

  const type = opts.type || 'jsdoc';
  const typeInstructions = {
    jsdoc: 'Generate comprehensive JSDoc/TSDoc comments for all functions and classes.',
    readme: 'Generate a complete README.md with installation, usage, API reference, and examples.',
    adr: 'Generate an Architecture Decision Record (ADR) documenting design decisions.',
    changelog: 'Generate a CHANGELOG.md based on the code structure and features.',
  };

  const prompt = opts.readme
    ? `Generate a comprehensive README.md for: ${content}\n\nInclude: title, description, features, installation, usage, API reference, examples, contributing guide, and license.`
    : `${typeInstructions[type] || typeInstructions.jsdoc}\n\nCode:\n${content}`;

  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
