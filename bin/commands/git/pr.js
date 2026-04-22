import { sendToAgent, ensureEngine, inf, err, wrn, printBox, gitExec, getCurrentBranch, C } from '../shared.js';

export async function handlePR(opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }

  const branch = getCurrentBranch();
  const base = opts.base || 'main';
  const log = gitExec(`git log ${base}..HEAD --oneline`) || '';
  const diff = gitExec(`git diff ${base}...HEAD --stat`) || '';
  const fullDiff = gitExec(`git diff ${base}...HEAD`) || '';

  if (!log && !diff) { wrn('No commits found compared to base branch.'); return; }

  inf('Generating PR description...');
  const prompt = `Generate a comprehensive GitHub Pull Request description.\n\nBranch: ${branch}\nBase: ${base}\n\nCommits:\n${log}\n\nFiles changed:\n${diff}\n\nDiff:\n\`\`\`diff\n${fullDiff.slice(0, 8000)}\n\`\`\`\n\nCreate a PR description with:\n## Summary\n(bullet points of what changed)\n\n## Changes\n(detailed breakdown)\n\n## Testing\n(how to test)\n\n## Notes\n(any important considerations)`;
  const res = await sendToAgent(prompt);
  printBox(res, 'PR Description');
  inf('Copy the above and paste into your GitHub PR.');
}
