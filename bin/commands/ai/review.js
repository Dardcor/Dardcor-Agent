import { sendToAgent, ensureEngine, inf, err, wrn, C } from '../shared.js';
import { gitExec } from '../shared.js';

export async function handleReview(opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }

  let diff = '';
  if (opts.branch) {
    diff = gitExec(`git diff ${opts.branch}...HEAD`) || '';
  } else if (opts.all) {
    diff = gitExec('git diff HEAD~1') || '';
  } else {
    diff = gitExec('git diff --cached') || gitExec('git diff') || '';
  }

  if (!diff) {
    wrn('No changes detected to review.');
    return;
  }

  inf('Performing AI code review...');
  const prompt = `Perform a thorough code review on these changes:\n\n\`\`\`diff\n${diff.slice(0, 15000)}\n\`\`\`\n\nReview for:\n1. **Bugs** — Logic errors, null pointer risks, off-by-one errors\n2. **Security** — Injection, XSS, auth issues, exposed secrets\n3. **Performance** — N+1 queries, inefficient loops, memory leaks\n4. **Code Quality** — SOLID, DRY, readability, naming\n5. **Tests** — Missing test coverage for new code\n6. **Documentation** — Missing or incorrect docs\n\nRate severity: 🔴 Critical | 🟡 Warning | 🟢 Suggestion\n\n${opts.strict ? 'STRICT MODE: Flag ALL issues including minor ones.' : ''}`;

  const res = await sendToAgent(prompt);
  console.log(`\n${C.white}${res}${C.reset}\n`);
}
