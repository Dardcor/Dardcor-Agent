import { loadConfig, saveConfig, inf, ok, err, C } from '../shared.js';
import readline from 'readline';

function askQuestion(question) {
  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
  return new Promise(resolve => {
    rl.question(question, ans => { rl.close(); resolve(ans.trim()); });
  });
}

const PROVIDERS = {
  antigravity: { name: 'Antigravity (Free Google AI)', fields: [], note: 'Uses OAuth via browser' },
  openai: { name: 'OpenAI (GPT-4, o1, o3)', fields: ['api_key'], note: 'Get key at platform.openai.com' },
  anthropic: { name: 'Anthropic (Claude)', fields: ['api_key'], note: 'Get key at console.anthropic.com' },
  gemini: { name: 'Google Gemini', fields: ['api_key'], note: 'Get key at aistudio.google.com' },
  groq: { name: 'Groq (Fast inference)', fields: ['api_key'], note: 'Get key at console.groq.com' },
  deepseek: { name: 'DeepSeek', fields: ['api_key'], note: 'Get key at platform.deepseek.com' },
  openrouter: { name: 'OpenRouter (Multi-provider)', fields: ['api_key'], note: 'Get key at openrouter.ai' },
  ollama: { name: 'Ollama (Local AI)', fields: ['base_url'], note: 'Runs locally on port 11434' },
  nvidia: { name: 'NVIDIA NIM (AI Catalog)', fields: ['api_key'], note: 'Get key at build.nvidia.com', baseURL: 'https://integrate.api.nvidia.com/v1' },
  custom: { name: 'Custom OpenAI-compatible', fields: ['api_key', 'base_url', 'model'], note: 'Any OpenAI-compatible API' },
};

const NVIDIA_MODELS = [
  'minimaxai/minimax-m2.7',
  'nvidia/llama-3.1-nemotron-70b-instruct',
  'nvidia/llama-3.1-405b-instruct',
  'meta/llama-3.1-70b-instruct',
  'meta/llama-3.3-70b-instruct',
  'meta/llama-3.1-8b-instruct',
  'mistralai/mixtral-8x22b-instruct-v0.1',
  'mistralai/mistral-large-2-instruct',
  'google/gemma-2-27b-it',
  'microsoft/phi-3.5-mini-instruct',
  'qwen/qwen2.5-72b-instruct',
  'deepseek-ai/deepseek-r1-distill-llama-70b',
];

export async function handleLogin(opts = {}) {
  console.log(`\n${C.bold}${C.purple}Provider Authentication${C.reset}\n`);

  let i = 1;
  for (const [k, v] of Object.entries(PROVIDERS)) {
    console.log(`  ${C.cyan}${i++}.${C.reset} ${v.name}  ${C.dim}— ${v.note}${C.reset}`);
  }
  console.log('');

  const choice = await askQuestion(`${C.yellow}Select provider (1-${Object.keys(PROVIDERS).length}): ${C.reset}`);
  const idx = parseInt(choice) - 1;
  const entries = Object.entries(PROVIDERS);
  if (idx < 0 || idx >= entries.length) { err('Invalid choice'); return; }

  const [providerKey, provider] = entries[idx];
  const cfg = loadConfig();
  cfg.provider = providerKey;

  if (providerKey === 'antigravity') {
    inf('Antigravity uses Google OAuth. Open the dashboard to authenticate:');
    inf('Run: dardcor run  →  then visit Settings → Model → Antigravity');
    saveConfig(cfg);
    ok(`Provider set to: ${providerKey}`);
    return;
  }

  if (providerKey === 'nvidia') {
    const apiKey = await askQuestion(`API Key (nvapi-...): `);
    cfg.api_key = apiKey;
    cfg.base_url = 'https://integrate.api.nvidia.com/v1';

    console.log(`\n${C.bold}Available NVIDIA Models:${C.reset}`);
    NVIDIA_MODELS.forEach((m, i) => console.log(`  ${C.cyan}${i + 1}.${C.reset} ${m}`));
    console.log('');

    const modelChoice = await askQuestion(`Select model (1-${NVIDIA_MODELS.length}) or type custom: `);
    const modelIdx = parseInt(modelChoice) - 1;
    if (modelIdx >= 0 && modelIdx < NVIDIA_MODELS.length) {
      cfg.model = NVIDIA_MODELS[modelIdx];
    } else if (isNaN(parseInt(modelChoice)) && modelChoice.trim()) {
      cfg.model = modelChoice.trim();
    } else {
      cfg.model = NVIDIA_MODELS[0];
    }
  } else if (providerKey === 'ollama') {
    const url = await askQuestion(`Base URL [http://localhost:11434]: `);
    cfg.base_url = url || 'http://localhost:11434';
    const model = await askQuestion(`Default model [qwen2.5-coder:7b]: `);
    cfg.model = model || 'qwen2.5-coder:7b';
  } else {
    for (const field of provider.fields) {
      const val = await askQuestion(`${field.replace('_', ' ')} [hidden]: `);
      cfg[field === 'api_key' ? 'api_key' : field] = val;
    }
    if (providerKey !== 'custom') {
      const model = await askQuestion(`Default model (press Enter to use default): `);
      if (model) cfg.model = model;
    }
  }

  saveConfig(cfg);
  ok(`Authenticated with ${provider.name}`);
  inf('Restart engine for changes to take effect: dardcor run');
}
