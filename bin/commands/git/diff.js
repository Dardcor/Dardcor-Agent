import { sendToAgent, ensureEngine, inf, err, wrn, C } from '../shared.js';
import { gitExec } from '../shared.js';

export async function handleDiff(ref, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }

  const target = ref || 'HEAD';
  const diff = gitExec(`git diff ${target}`) || gitExec(`git diff ${target} --stat`) || '';
  if (!diff) { wrn('No diff found.'); return; }

  inf(`AI diff analysis vs ${target}...`);
  const riskNote = opts.risk ? '\n\nAlso provide a RISK ASSESSMENT: rate the overall risk (Low/Medium/High/Critical) and explain what could break.' : '';
  const prompt = `Analyze this git diff and provide an intelligent summary:\n\n\`\`\`diff\n${diff.slice(0, 12000)}\n\`\`\`\n\nProvide:\n1. Summary of what changed\n2. Impact analysis (what areas affected)\n3. Potential issues or concerns${riskNote}`;
  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
_