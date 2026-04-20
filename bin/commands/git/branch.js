import { sendToAgent, ensureEngine, inf, err, ok, printBox, confirmPrompt, C } from '../shared.js';
import { execSync } from 'child_process';

export async function handleBranch(description, opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }
  inf('Generating branch name...');

  const prompt = `Generate a git branch name for: "${description}"\n\nRules:\n- lowercase-kebab-case\n- max 50 chars\n- descriptive but concise\n- use prefix: feat/, fix/, docs/, refactor/, test/, chore/, hotfix/\n\nReturn ONLY the branch name, nothing else.`;
  const branchName = (await sendToAgent(prompt)).trim().replace(/\s+/g, '-').toLowerCase();

  printBox(branchName, 'Branch Name');

  const confirmed = await confirmPrompt('Create and checkout this branch?');
  if (confirmed) {
    try {
      execSync(`git checkout -b "${branchName}"`, { stdio: 'inherit' });
      ok(`Branch created: ${branchName}`);
    } catch (e) { err('git checkout failed'); }
  }
}
