import { httpPost, httpGet, ensureEngine, inf, ok, err, C } from '../shared.js';
import fs from 'fs';
import path from 'path';
import os from 'os';

const INDEX_FILE = path.join(os.homedir(), '.dardcor', 'database', 'code_index.json');

function buildIndex(rootPath, patterns = ['**/*.ts', '**/*.js', '**/*.go', '**/*.py', '**/*.tsx', '**/*.jsx']) {
  const excluded = new Set(['node_modules', '.git', 'dist', 'build', '__pycache__', 'vendor', '.next']);
  const index = { path: rootPath, files: [], builtAt: new Date().toISOString() };

  function walk(dir) {
    let entries;
    try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch { return; }
    for (const e of entries) {
      if (excluded.has(e.name)) continue;
      const fullPath = path.join(dir, e.name);
      if (e.isDirectory()) {
        walk(fullPath);
      } else if (e.isFile()) {
        const ext = path.extname(e.name);
        const textExts = new Set(['.ts', '.tsx', '.js', '.jsx', '.go', '.py', '.rs', '.java', '.c', '.cpp', '.h', '.md', '.json', '.yaml', '.yml', '.toml', '.env']);
        if (textExts.has(ext)) {
          try {
            const stat = fs.statSync(fullPath);
            if (stat.size < 500 * 1024) {
              const content = fs.readFileSync(fullPath, 'utf8');
              index.files.push({
                path: path.relative(rootPath, fullPath),
                size: stat.size,
                ext,
                lines: content.split('\n').length,
                preview: content.slice(0, 200),
              });
            }
          } catch { }
        }
      }
    }
  }

  walk(rootPath);
  return index;
}

export async function handleIndex(opts = {}) {
  if (opts.status) {
    if (fs.existsSync(INDEX_FILE)) {
      const idx = JSON.parse(fs.readFileSync(INDEX_FILE, 'utf8'));
      console.log(`\n${C.bold}Index Status:${C.reset}`);
      console.log(`  Path:     ${idx.path}`);
      console.log(`  Files:    ${idx.files.length}`);
      console.log(`  Built:    ${new Date(idx.builtAt).toLocaleString()}`);
    } else {
      inf('No index built yet. Run: dardcor index');
    }
    return;
  }

  const rootPath = process.cwd();
  inf(`Building code index for: ${rootPath}`);
  const include = opts.include;
  const exclude = opts.exclude;

  const index = buildIndex(rootPath);
  const dir = path.dirname(INDEX_FILE);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(INDEX_FILE, JSON.stringify(index, null, 2));
  ok(`Index built: ${index.files.length} files indexed`);


  try {
    if (await ensureEngine()) {
      await httpPost('/api/index/build', { path: rootPath });
    }
  } catch { }
}
