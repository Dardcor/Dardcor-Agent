import { sendToAgent, ensureEngine, inf, err, ok, C } from '../shared.js';
import { gitExec } from '../shared.js';
import fs from 'fs';

export async function handleResolve(opts = {}) {
  if (!(await ensureEngine())) { err('Engine offline. Run: dardcor run'); process.exit(1); }

  const conflicts = gitExec('git diff --name-only --diff-filter=U') || '';
  if (!conflicts) { ok('No merge conflicts detected.'); return; }

  const conflictFiles = conflicts.split('\n').filter(Boolean);
  inf(`Found ${conflictFiles.length} conflict(s): ${conflictFiles.join(', ')}`);

  for (const file of conflictFiles) {
    if (!fs.existsSync(file)) continue;
    const content = fs.readFileSync(file, 'utf8');
    if (!content.includes('<<<<<<<')) continue;

    inf(`Resolving: ${file}`);
    const prompt = `Resolve this merge conflict in file: ${file}\n\nFile content with conflicts:\n\`\`\`\n${content}\n\`\`\`\n\nAnalyze both versions and produce the correct merged result. Return ONLY the resolved file content without any conflict markers. Choose the best resolution that preserves intent from both changes.`;
    const resolved = await sendToAgent(prompt);
    const clean = resolved.replace(/^```[a-z]*\n?/, '').replace(/```$/, '').trim();
    fs.writeFileSync(file, clean + '\n');
    ok(`Resolved: ${file}`);
  }

  inf('Review resolved files, then run: git add . && git commit');
}
