import { sendToAgent, ensureEngine, inf, err, wrn, ok, printBox, confirmPrompt, gitExec, getStagedDiff, getFullStagedDiff, C } from '../shared.js';
import { execSync } from 'child_process';

export async function handleCommit(opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }

  let diff = getStagedDiff();
  if (!diff) {
    wrn('No staged changes found. Staging all...');
    try { execSync('git add -A', { stdio: 'inherit' }); } catch (e) { err('git add failed'); return; }
    diff = getStagedDiff();
  }

  if (!diff) { wrn('No changes to commit.'); return; }

  const fullDiff = getFullStagedDiff();
  const conventional = opts.conventional ? ' Use conventional commit format: type(scope): description. Types: feat|fix|docs|style|refactor|test|chore|perf|ci|build.' : '';
  inf('Analyzing changes...');

  const prompt = `Generate a concise git commit message for these changes.${conventional}\n\nStat:\n${diff}\n\nDiff:\n\`\`\`diff\n${fullDiff.slice(0, 8000)}\n\`\`\`\n\nReturn ONLY the commit message text, no explanation, no quotes, no markdown.`;
  const message = await sendToAgent(prompt);
  const cleanMsg = message.trim().replace(/^["'`]|["'`]$/g, '');

  printBox(cleanMsg, 'Suggested Commit');

  if (opts.dryRun) { inf('Dry run — not committing.'); return; }

  const confirmed = await confirmPrompt('Commit with this message?');
  if (confirmed) {
    try {
      execSync(`git commit -m "${cleanMsg.replace(/"/g, '\\"')}"`, { stdio: 'inherit' });
      ok('Committed!');
    } catch (e) { err('git commit failed'); }
  }
}
