import { loadConfig } from '../shared.js';

const C = {
  reset: '\x1b[0m', bold: '\x1b[1m', dim: '\x1b[2m',
  red: '\x1b[31m', green: '\x1b[32m', yellow: '\x1b[33m',
  cyan: '\x1b[36m', white: '\x1b[97m', gray: '\x1b[90m',
  purple: '\x1b[38;5;93m', orange: '\x1b[38;5;208m',
  pink: '\x1b[38;5;213m', teal: '\x1b[38;5;43m',
  blue: '\x1b[34m',
};



const MODELS = {


  antigravity: {
    label: 'Antigravity  (Free Google AI via OAuth)',
    color: C.green,
    badge: '✨ FREE',
    url: 'aistudio.google.com',
    note: 'No API key needed — use dardcor login → antigravity',
    models: [
      { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', ctx: '1M', out: '65K', tags: ['free', 'vision', 'reason', 'code', 'long'], $in: 0, $out: 0 },
      { id: 'gemini-2.5-flash', name: 'Gemini 2.5 Flash', ctx: '1M', out: '65K', tags: ['free', 'fast', 'vision', 'code', 'long'], $in: 0, $out: 0 },
      { id: 'gemini-2.5-flash-lite', name: 'Gemini 2.5 Flash Lite', ctx: '1M', out: '65K', tags: ['free', 'fast', 'cheap', 'long'], $in: 0, $out: 0 },
      { id: 'gemini-2.0-flash', name: 'Gemini 2.0 Flash', ctx: '1M', out: '8K', tags: ['free', 'fast', 'vision', 'code'], $in: 0, $out: 0 },
      { id: 'gemini-2.0-flash-lite', name: 'Gemini 2.0 Flash Lite', ctx: '1M', out: '8K', tags: ['free', 'fast', 'cheap'], $in: 0, $out: 0 },
      { id: 'gemini-1.5-pro', name: 'Gemini 1.5 Pro', ctx: '2M', out: '8K', tags: ['free', 'vision', 'long'], $in: 0, $out: 0 },
      { id: 'gemini-1.5-flash', name: 'Gemini 1.5 Flash', ctx: '1M', out: '8K', tags: ['free', 'fast', 'vision'], $in: 0, $out: 0 },
    ],
  },


  openai: {
    label: 'OpenAI',
    color: C.teal,
    badge: '',
    url: 'platform.openai.com',
    note: 'Set key: dardcor config --set api_key=sk-...',
    models: [

      { id: 'gpt-4.1', name: 'GPT-4.1', ctx: '1M', out: '32K', tags: ['vision', 'code'], $in: 2.00, $out: 8.00 },
      { id: 'gpt-4.1-mini', name: 'GPT-4.1 Mini', ctx: '1M', out: '32K', tags: ['fast', 'cheap', 'vision'], $in: 0.40, $out: 1.60 },
      { id: 'gpt-4.1-nano', name: 'GPT-4.1 Nano', ctx: '1M', out: '32K', tags: ['fast', 'cheap'], $in: 0.10, $out: 0.40 },

      { id: 'gpt-4o', name: 'GPT-4o', ctx: '128K', out: '16K', tags: ['vision', 'code'], $in: 2.50, $out: 10.00 },
      { id: 'gpt-4o-mini', name: 'GPT-4o Mini', ctx: '128K', out: '16K', tags: ['fast', 'cheap', 'vision'], $in: 0.15, $out: 0.60 },
      { id: 'gpt-4o-audio-preview', name: 'GPT-4o Audio', ctx: '128K', out: '16K', tags: ['vision'], $in: 2.50, $out: 10.00 },

      { id: 'gpt-4-turbo', name: 'GPT-4 Turbo', ctx: '128K', out: '4K', tags: ['vision'], $in: 10.00, $out: 30.00 },
      { id: 'gpt-4', name: 'GPT-4', ctx: '8K', out: '8K', tags: [], $in: 30.00, $out: 60.00 },

      { id: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo', ctx: '16K', out: '4K', tags: ['cheap'], $in: 0.50, $out: 1.50 },

      { id: 'o4-mini', name: 'o4-mini', ctx: '200K', out: '100K', tags: ['reason', 'code'], $in: 1.10, $out: 4.40 },
      { id: 'o3', name: 'o3', ctx: '200K', out: '100K', tags: ['reason', 'code'], $in: 10.00, $out: 40.00 },
      { id: 'o3-mini', name: 'o3-mini', ctx: '200K', out: '100K', tags: ['reason', 'code', 'cheap'], $in: 1.10, $out: 4.40 },
      { id: 'o1', name: 'o1', ctx: '200K', out: '100K', tags: ['reason'], $in: 15.00, $out: 60.00 },
      { id: 'o1-mini', name: 'o1-mini', ctx: '128K', out: '65K', tags: ['reason', 'cheap'], $in: 3.00, $out: 12.00 },
      { id: 'o1-preview', name: 'o1-preview', ctx: '128K', out: '32K', tags: ['reason'], $in: 15.00, $out: 60.00 },

      { id: 'codex-mini-latest', name: 'Codex Mini', ctx: '200K', out: '100K', tags: ['code', 'reason'], $in: 1.50, $out: 6.00 },
    ],
  },


  anthropic: {
    label: 'Anthropic  (Claude)',
    color: C.orange,
    badge: '',
    url: 'console.anthropic.com',
    note: 'Set key: dardcor config --set api_key=sk-ant-...',
    models: [

      { id: 'claude-opus-4-6', name: 'Claude Opus 4.6', ctx: '1M', out: '128K', tags: ['reason', 'code', 'vision'], $in: 15.00, $out: 75.00 },
      { id: 'claude-sonnet-4-6', name: 'Claude Sonnet 4.6', ctx: '1M', out: '64K', tags: ['code', 'vision'], $in: 3.00, $out: 15.00 },
      { id: 'claude-opus-4-5-20250929', name: 'Claude Opus 4.5', ctx: '1M', out: '128K', tags: ['reason', 'code', 'vision'], $in: 15.00, $out: 75.00 },
      { id: 'claude-sonnet-4-5-20250929', name: 'Claude Sonnet 4.5', ctx: '1M', out: '64K', tags: ['code', 'vision'], $in: 3.00, $out: 15.00 },
      { id: 'claude-haiku-4-5-20251001', name: 'Claude Haiku 4.5', ctx: '200K', out: '64K', tags: ['fast', 'cheap', 'code'], $in: 0.25, $out: 1.25 },

      { id: 'claude-opus-4-20250514', name: 'Claude Opus 4', ctx: '200K', out: '32K', tags: ['reason', 'code', 'vision'], $in: 15.00, $out: 75.00 },
      { id: 'claude-sonnet-4-20250514', name: 'Claude Sonnet 4', ctx: '1M', out: '64K', tags: ['code', 'vision'], $in: 3.00, $out: 15.00 },

      { id: 'claude-3-7-sonnet-20250219', name: 'Claude 3.7 Sonnet', ctx: '200K', out: '128K', tags: ['reason', 'code', 'vision'], $in: 3.00, $out: 15.00 },

      { id: 'claude-3-5-sonnet-20241022', name: 'Claude 3.5 Sonnet', ctx: '200K', out: '8K', tags: ['code', 'vision'], $in: 3.00, $out: 15.00 },
      { id: 'claude-3-5-haiku-20241022', name: 'Claude 3.5 Haiku', ctx: '200K', out: '8K', tags: ['fast', 'cheap', 'code'], $in: 0.80, $out: 4.00 },

      { id: 'claude-3-opus-20240229', name: 'Claude 3 Opus', ctx: '200K', out: '4K', tags: ['reason'], $in: 15.00, $out: 75.00 },
      { id: 'claude-3-sonnet-20240229', name: 'Claude 3 Sonnet', ctx: '200K', out: '4K', tags: [], $in: 3.00, $out: 15.00 },
      { id: 'claude-3-haiku-20240307', name: 'Claude 3 Haiku', ctx: '200K', out: '4K', tags: ['fast', 'cheap'], $in: 0.25, $out: 1.25 },
    ],
  },


  gemini: {
    label: 'Google Gemini  (API Key)',
    color: C.blue,
    badge: '',
    url: 'aistudio.google.com',
    note: 'Set key: dardcor config --set api_key=AIza...',
    models: [
      { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', ctx: '1M', out: '65K', tags: ['vision', 'reason', 'code', 'long'], $in: 1.25, $out: 10.00 },
      { id: 'gemini-2.5-flash', name: 'Gemini 2.5 Flash', ctx: '1M', out: '65K', tags: ['fast', 'vision', 'code', 'long'], $in: 0.15, $out: 0.60 },
      { id: 'gemini-2.5-flash-lite', name: 'Gemini 2.5 Flash Lite', ctx: '1M', out: '65K', tags: ['fast', 'cheap', 'long'], $in: 0.075, $out: 0.30 },
      { id: 'gemini-2.0-flash', name: 'Gemini 2.0 Flash', ctx: '1M', out: '8K', tags: ['fast', 'vision'], $in: 0.10, $out: 0.40 },
      { id: 'gemini-2.0-flash-lite', name: 'Gemini 2.0 Flash Lite', ctx: '1M', out: '8K', tags: ['fast', 'cheap'], $in: 0.075, $out: 0.30 },
      { id: 'gemini-2.0-flash-thinking', name: 'Gemini 2.0 Flash Think', ctx: '1M', out: '8K', tags: ['reason'], $in: 0.15, $out: 0.60 },
      { id: 'gemini-1.5-pro', name: 'Gemini 1.5 Pro', ctx: '2M', out: '8K', tags: ['vision', 'long'], $in: 1.25, $out: 5.00 },
      { id: 'gemini-1.5-flash', name: 'Gemini 1.5 Flash', ctx: '1M', out: '8K', tags: ['fast', 'vision'], $in: 0.075, $out: 0.30 },
      { id: 'gemini-1.5-flash-8b', name: 'Gemini 1.5 Flash 8B', ctx: '1M', out: '8K', tags: ['fast', 'cheap'], $in: 0.0375, $out: 0.15 },
    ],
  },


  groq: {
    label: 'Groq  (Ultra-fast Inference)',
    color: C.pink,
    badge: '⚡ FAST',
    url: 'console.groq.com',
    note: 'Set key: dardcor config --set api_key=gsk_...',
    models: [

      { id: 'meta-llama/llama-4-scout-17b-16e-instruct', name: 'Llama 4 Scout 17B', ctx: '128K', out: '8K', tags: ['fast', 'vision'], $in: 0.11, $out: 0.34 },
      { id: 'meta-llama/llama-4-maverick-17b-128e-instruct', name: 'Llama 4 Maverick 17B', ctx: '128K', out: '8K', tags: ['fast', 'vision'], $in: 0.20, $out: 0.60 },

      { id: 'llama-3.3-70b-versatile', name: 'Llama 3.3 70B', ctx: '128K', out: '32K', tags: ['fast', 'code'], $in: 0.59, $out: 0.79 },
      { id: 'llama-3.3-70b-specdec', name: 'Llama 3.3 70B SpecDec', ctx: '8K', out: '8K', tags: ['fast'], $in: 0.59, $out: 0.99 },
      { id: 'llama-3.1-70b-versatile', name: 'Llama 3.1 70B', ctx: '128K', out: '8K', tags: ['fast'], $in: 0.59, $out: 0.79 },
      { id: 'llama-3.1-8b-instant', name: 'Llama 3.1 8B Instant', ctx: '128K', out: '8K', tags: ['fast', 'cheap'], $in: 0.05, $out: 0.08 },
      { id: 'llama3-70b-8192', name: 'Llama 3 70B', ctx: '8K', out: '8K', tags: ['fast'], $in: 0.59, $out: 0.79 },
      { id: 'llama3-8b-8192', name: 'Llama 3 8B', ctx: '8K', out: '8K', tags: ['fast', 'cheap'], $in: 0.05, $out: 0.08 },

      { id: 'moonshotai/kimi-k2-instruct', name: 'Kimi K2', ctx: '128K', out: '16K', tags: ['code', 'reason'], $in: 1.00, $out: 3.00 },

      { id: 'qwen/qwen3-32b', name: 'Qwen3 32B', ctx: '128K', out: '8K', tags: ['code', 'reason'], $in: 0.29, $out: 0.59 },

      { id: 'groq/compound', name: 'Groq Compound', ctx: '128K', out: '32K', tags: ['fast', 'code', 'reason'], $in: 0.00, $out: 0.00 },
      { id: 'groq/compound-mini', name: 'Groq Compound Mini', ctx: '128K', out: '32K', tags: ['fast', 'cheap'], $in: 0.00, $out: 0.00 },

      { id: 'mixtral-8x7b-32768', name: 'Mixtral 8x7B', ctx: '32K', out: '32K', tags: ['fast'], $in: 0.24, $out: 0.24 },
      { id: 'gemma2-9b-it', name: 'Gemma 2 9B', ctx: '8K', out: '8K', tags: ['fast', 'cheap'], $in: 0.20, $out: 0.20 },
      { id: 'gemma-7b-it', name: 'Gemma 7B', ctx: '8K', out: '8K', tags: ['fast', 'cheap'], $in: 0.07, $out: 0.07 },
    ],
  },


  deepseek: {
    label: 'DeepSeek',
    color: C.cyan,
    badge: '💰 CHEAP',
    url: 'platform.deepseek.com',
    note: 'Set key: dardcor config --set api_key=sk-...',
    models: [
      { id: 'deepseek-chat', name: 'DeepSeek V3 Chat', ctx: '64K', out: '8K', tags: ['code', 'cheap'], $in: 0.27, $out: 1.10 },
      { id: 'deepseek-reasoner', name: 'DeepSeek R1 Reasoner', ctx: '64K', out: '8K', tags: ['reason', 'code', 'cheap'], $in: 0.55, $out: 2.19 },
      { id: 'deepseek-coder', name: 'DeepSeek Coder V2', ctx: '128K', out: '8K', tags: ['code', 'cheap'], $in: 0.14, $out: 0.28 },
    ],
  },


  xai: {
    label: 'xAI  (Grok)',
    color: C.white,
    badge: '',
    url: 'console.x.ai',
    note: 'Set key: dardcor config --set api_key=xai-...',
    models: [
      { id: 'grok-3', name: 'Grok 3', ctx: '128K', out: '128K', tags: ['reason', 'code'], $in: 3.00, $out: 15.00 },
      { id: 'grok-3-mini', name: 'Grok 3 Mini', ctx: '128K', out: '128K', tags: ['reason', 'cheap'], $in: 0.30, $out: 0.50 },
      { id: 'grok-3-fast', name: 'Grok 3 Fast', ctx: '128K', out: '128K', tags: ['fast'], $in: 5.00, $out: 25.00 },
      { id: 'grok-3-mini-fast', name: 'Grok 3 Mini Fast', ctx: '128K', out: '128K', tags: ['fast', 'cheap'], $in: 0.60, $out: 4.00 },
      { id: 'grok-2-1212', name: 'Grok 2', ctx: '128K', out: '128K', tags: [], $in: 2.00, $out: 10.00 },
      { id: 'grok-vision-beta', name: 'Grok Vision Beta', ctx: '8K', out: '8K', tags: ['vision'], $in: 5.00, $out: 15.00 },
    ],
  },


  mistral: {
    label: 'Mistral AI',
    color: C.yellow,
    badge: '',
    url: 'console.mistral.ai',
    note: 'base_url: https://api.mistral.ai/v1',
    models: [

      { id: 'mistral-large-latest', name: 'Mistral Large', ctx: '128K', out: '128K', tags: ['code'], $in: 2.00, $out: 6.00 },
      { id: 'mistral-medium-latest', name: 'Mistral Medium', ctx: '128K', out: '128K', tags: [], $in: 0.40, $out: 2.00 },
      { id: 'mistral-small-latest', name: 'Mistral Small', ctx: '32K', out: '32K', tags: ['cheap'], $in: 0.10, $out: 0.30 },
      { id: 'open-mistral-nemo', name: 'Mistral Nemo', ctx: '128K', out: '128K', tags: ['cheap', 'fast'], $in: 0.15, $out: 0.15 },
      { id: 'mistral-tiny-latest', name: 'Mistral Tiny', ctx: '32K', out: '32K', tags: ['cheap'], $in: 0.025, $out: 0.025 },

      { id: 'codestral-latest', name: 'Codestral', ctx: '256K', out: '256K', tags: ['code'], $in: 0.30, $out: 0.90 },
      { id: 'codestral-mamba-latest', name: 'Codestral Mamba', ctx: '256K', out: '256K', tags: ['code', 'fast'], $in: 0.25, $out: 0.25 },

      { id: 'pixtral-large-latest', name: 'Pixtral Large', ctx: '128K', out: '128K', tags: ['vision'], $in: 2.00, $out: 6.00 },
      { id: 'pixtral-12b-2409', name: 'Pixtral 12B', ctx: '128K', out: '128K', tags: ['vision', 'cheap'], $in: 0.15, $out: 0.15 },

      { id: 'ministral-8b-latest', name: 'Ministral 8B', ctx: '128K', out: '128K', tags: ['fast', 'cheap'], $in: 0.10, $out: 0.10 },
      { id: 'ministral-3b-latest', name: 'Ministral 3B', ctx: '128K', out: '128K', tags: ['fast', 'cheap'], $in: 0.04, $out: 0.04 },

      { id: 'magistral-medium-latest', name: 'Magistral Medium', ctx: '32K', out: '16K', tags: ['reason'], $in: 2.00, $out: 5.00 },
      { id: 'magistral-small-latest', name: 'Magistral Small', ctx: '32K', out: '16K', tags: ['reason', 'cheap'], $in: 0.50, $out: 1.50 },
    ],
  },


  openrouter: {
    label: 'OpenRouter  (100+ models via one API)',
    color: C.purple,
    badge: '',
    url: 'openrouter.ai',
    note: 'base_url: https://openrouter.ai/api/v1',
    models: [

      { id: 'anthropic/claude-opus-4', name: 'Claude Opus 4', ctx: '200K', out: '32K', tags: ['reason', 'code'], $in: 15.00, $out: 75.00 },
      { id: 'anthropic/claude-sonnet-4', name: 'Claude Sonnet 4', ctx: '200K', out: '64K', tags: ['code'], $in: 3.00, $out: 15.00 },
      { id: 'anthropic/claude-3.7-sonnet', name: 'Claude 3.7 Sonnet', ctx: '200K', out: '128K', tags: ['reason', 'code'], $in: 3.00, $out: 15.00 },
      { id: 'openai/gpt-4.1', name: 'GPT-4.1', ctx: '1M', out: '32K', tags: ['vision'], $in: 2.00, $out: 8.00 },
      { id: 'openai/o3', name: 'o3', ctx: '200K', out: '100K', tags: ['reason'], $in: 10.00, $out: 40.00 },
      { id: 'openai/o4-mini', name: 'o4-mini', ctx: '200K', out: '100K', tags: ['reason', 'cheap'], $in: 1.10, $out: 4.40 },
      { id: 'google/gemini-2.5-pro', name: 'Gemini 2.5 Pro', ctx: '1M', out: '65K', tags: ['vision', 'reason', 'long'], $in: 1.25, $out: 10.00 },
      { id: 'google/gemini-2.5-flash', name: 'Gemini 2.5 Flash', ctx: '1M', out: '65K', tags: ['fast', 'vision'], $in: 0.15, $out: 0.60 },
      { id: 'x-ai/grok-3', name: 'Grok 3', ctx: '131K', out: '128K', tags: ['reason'], $in: 3.00, $out: 15.00 },
      { id: 'x-ai/grok-3-mini', name: 'Grok 3 Mini', ctx: '131K', out: '128K', tags: ['reason', 'cheap'], $in: 0.30, $out: 0.50 },

      { id: 'meta-llama/llama-4-maverick', name: 'Llama 4 Maverick', ctx: '524K', out: '128K', tags: ['vision'], $in: 0.18, $out: 0.60 },
      { id: 'meta-llama/llama-3.3-70b-instruct', name: 'Llama 3.3 70B', ctx: '128K', out: '32K', tags: ['code'], $in: 0.12, $out: 0.40 },
      { id: 'meta-llama/llama-3.1-405b-instruct', name: 'Llama 3.1 405B', ctx: '128K', out: '32K', tags: [], $in: 2.70, $out: 2.70 },
      { id: 'qwen/qwen3-235b-a22b', name: 'Qwen3 235B', ctx: '32K', out: '32K', tags: ['code', 'reason'], $in: 0.13, $out: 0.60 },
      { id: 'qwen/qwen3-30b-a3b', name: 'Qwen3 30B', ctx: '32K', out: '32K', tags: ['code', 'cheap'], $in: 0.05, $out: 0.28 },
      { id: 'moonshotai/kimi-k2', name: 'Kimi K2', ctx: '128K', out: '16K', tags: ['code'], $in: 0.50, $out: 2.50 },
      { id: 'deepseek/deepseek-r1', name: 'DeepSeek R1', ctx: '163K', out: '65K', tags: ['reason', 'code'], $in: 0.50, $out: 2.19 },
      { id: 'deepseek/deepseek-chat-v3-0324', name: 'DeepSeek V3', ctx: '163K', out: '65K', tags: ['code'], $in: 0.27, $out: 1.10 },
      { id: 'mistralai/mistral-large', name: 'Mistral Large', ctx: '128K', out: '128K', tags: [], $in: 2.00, $out: 6.00 },
      { id: 'mistralai/codestral-2501', name: 'Codestral 2501', ctx: '256K', out: '256K', tags: ['code'], $in: 0.30, $out: 0.90 },
      { id: 'microsoft/phi-4', name: 'Phi-4', ctx: '16K', out: '16K', tags: ['cheap'], $in: 0.07, $out: 0.14 },
      { id: 'cohere/command-r-plus', name: 'Command R+', ctx: '128K', out: '4K', tags: [], $in: 2.50, $out: 10.00 },
      { id: 'perplexity/sonar-pro', name: 'Perplexity Sonar Pro', ctx: '127K', out: '8K', tags: [], $in: 3.00, $out: 15.00 },
    ],
  },


  together: {
    label: 'Together AI',
    color: C.green,
    badge: '💰 CHEAP',
    url: 'api.together.ai',
    note: 'base_url: https://api.together.xyz/v1',
    models: [
      { id: 'meta-llama/Llama-4-Scout-17B-16E-Instruct', name: 'Llama 4 Scout 17B', ctx: '328K', out: '16K', tags: ['fast', 'vision'], $in: 0.12, $out: 0.30 },
      { id: 'meta-llama/Llama-3.3-70B-Instruct-Turbo', name: 'Llama 3.3 70B Turbo', ctx: '128K', out: '4K', tags: ['fast'], $in: 0.88, $out: 0.88 },
      { id: 'meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo', name: 'Llama 3.1 405B', ctx: '128K', out: '4K', tags: [], $in: 3.50, $out: 3.50 },
      { id: 'Qwen/Qwen3-235B-A22B-fp8-tput', name: 'Qwen3 235B', ctx: '32K', out: '32K', tags: ['reason', 'code'], $in: 0.13, $out: 0.60 },
      { id: 'deepseek-ai/DeepSeek-R1', name: 'DeepSeek R1', ctx: '163K', out: '32K', tags: ['reason', 'code'], $in: 3.00, $out: 7.00 },
      { id: 'deepseek-ai/DeepSeek-V3', name: 'DeepSeek V3', ctx: '128K', out: '32K', tags: ['code'], $in: 0.50, $out: 1.25 },
      { id: 'mistralai/Mixtral-8x22B-Instruct-v0.1', name: 'Mixtral 8x22B', ctx: '65K', out: '16K', tags: [], $in: 1.20, $out: 1.20 },
      { id: 'google/gemma-2-27b-it', name: 'Gemma 2 27B', ctx: '8K', out: '8K', tags: ['cheap'], $in: 0.80, $out: 0.80 },
      { id: 'microsoft/WizardLM-2-8x22B', name: 'WizardLM 2 8x22B', ctx: '65K', out: '16K', tags: ['code'], $in: 1.20, $out: 1.20 },
    ],
  },


  fireworks: {
    label: 'Fireworks AI  (Ultra-fast)',
    color: C.orange,
    badge: '⚡ FAST',
    url: 'app.fireworks.ai',
    note: 'base_url: https://api.fireworks.ai/inference/v1',
    models: [
      { id: 'accounts/fireworks/models/llama-v3p3-70b-instruct', name: 'Llama 3.3 70B', ctx: '128K', out: '16K', tags: ['fast'], $in: 0.90, $out: 0.90 },
      { id: 'accounts/fireworks/models/llama-v3p1-405b-instruct', name: 'Llama 3.1 405B', ctx: '131K', out: '16K', tags: [], $in: 3.00, $out: 3.00 },
      { id: 'accounts/fireworks/models/deepseek-r1', name: 'DeepSeek R1', ctx: '160K', out: '32K', tags: ['reason', 'code'], $in: 3.00, $out: 8.00 },
      { id: 'accounts/fireworks/models/deepseek-v3', name: 'DeepSeek V3', ctx: '64K', out: '8K', tags: ['code'], $in: 0.90, $out: 0.90 },
      { id: 'accounts/fireworks/models/qwen3-235b-a22b', name: 'Qwen3 235B', ctx: '32K', out: '32K', tags: ['code'], $in: 0.22, $out: 0.88 },
      { id: 'accounts/fireworks/models/kimi-k2-instruct', name: 'Kimi K2', ctx: '128K', out: '16K', tags: ['code'], $in: 0.89, $out: 2.48 },
      { id: 'accounts/fireworks/models/mixtral-8x22b-instruct', name: 'Mixtral 8x22B', ctx: '65K', out: '16K', tags: [], $in: 1.20, $out: 1.20 },
    ],
  },


  perplexity: {
    label: 'Perplexity AI  (Web-search capable)',
    color: C.teal,
    badge: '🔍 SEARCH',
    url: 'www.perplexity.ai/settings/api',
    note: 'base_url: https://api.perplexity.ai',
    models: [
      { id: 'sonar-pro', name: 'Sonar Pro', ctx: '200K', out: '8K', tags: ['reason'], $in: 3.00, $out: 15.00 },
      { id: 'sonar', name: 'Sonar', ctx: '127K', out: '8K', tags: ['cheap'], $in: 1.00, $out: 1.00 },
      { id: 'sonar-reasoning-pro', name: 'Sonar Reasoning Pro', ctx: '127K', out: '8K', tags: ['reason'], $in: 2.00, $out: 8.00 },
      { id: 'sonar-reasoning', name: 'Sonar Reasoning', ctx: '127K', out: '8K', tags: ['reason', 'cheap'], $in: 1.00, $out: 5.00 },
      { id: 'sonar-deep-research', name: 'Sonar Deep Research', ctx: '127K', out: '8K', tags: ['reason'], $in: 2.00, $out: 8.00 },
      { id: 'r1-1776', name: 'R1-1776', ctx: '128K', out: '8K', tags: ['reason'], $in: 2.00, $out: 8.00 },
    ],
  },


  cohere: {
    label: 'Cohere',
    color: C.blue,
    badge: '',
    url: 'dashboard.cohere.ai',
    note: 'base_url: https://api.cohere.ai/v1',
    models: [
      { id: 'command-r-plus-08-2024', name: 'Command R+ (Aug 24)', ctx: '128K', out: '4K', tags: ['long'], $in: 2.50, $out: 10.00 },
      { id: 'command-r-08-2024', name: 'Command R (Aug 24)', ctx: '128K', out: '4K', tags: ['cheap'], $in: 0.15, $out: 0.60 },
      { id: 'command-a-03-2025', name: 'Command A', ctx: '256K', out: '8K', tags: ['long'], $in: 2.50, $out: 10.00 },
      { id: 'command-r7b-12-2024', name: 'Command R 7B', ctx: '128K', out: '4K', tags: ['cheap', 'fast'], $in: 0.0375, $out: 0.15 },
    ],
  },


  ollama: {
    label: 'Ollama  (Local — no API key)',
    color: C.yellow,
    badge: '🏠 LOCAL',
    url: 'ollama.com/library',
    note: 'Install: ollama.com  |  base_url: http://localhost:11434',
    models: [

      { id: 'qwen2.5-coder:32b', name: 'Qwen2.5 Coder 32B', ctx: '128K', out: '8K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'qwen2.5-coder:14b', name: 'Qwen2.5 Coder 14B', ctx: '128K', out: '8K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'qwen2.5-coder:7b', name: 'Qwen2.5 Coder 7B', ctx: '128K', out: '8K', tags: ['local', 'code', 'fast'], $in: 0, $out: 0 },
      { id: 'codellama:70b', name: 'CodeLlama 70B', ctx: '100K', out: '16K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'codellama:34b', name: 'CodeLlama 34B', ctx: '100K', out: '16K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'codellama:13b', name: 'CodeLlama 13B', ctx: '100K', out: '16K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'codellama:7b', name: 'CodeLlama 7B', ctx: '100K', out: '16K', tags: ['local', 'code', 'fast'], $in: 0, $out: 0 },
      { id: 'deepseek-coder-v2:16b', name: 'DeepSeek Coder V2 16B', ctx: '128K', out: '8K', tags: ['local', 'code'], $in: 0, $out: 0 },

      { id: 'llama3.3:70b', name: 'Llama 3.3 70B', ctx: '128K', out: '8K', tags: ['local'], $in: 0, $out: 0 },
      { id: 'llama3.2:3b', name: 'Llama 3.2 3B', ctx: '128K', out: '8K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'llama3.1:70b', name: 'Llama 3.1 70B', ctx: '128K', out: '8K', tags: ['local'], $in: 0, $out: 0 },
      { id: 'llama3.1:8b', name: 'Llama 3.1 8B', ctx: '128K', out: '8K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'qwen2.5:72b', name: 'Qwen2.5 72B', ctx: '128K', out: '8K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'qwen2.5:32b', name: 'Qwen2.5 32B', ctx: '128K', out: '8K', tags: ['local', 'code'], $in: 0, $out: 0 },
      { id: 'qwen2.5:14b', name: 'Qwen2.5 14B', ctx: '128K', out: '8K', tags: ['local', 'code', 'fast'], $in: 0, $out: 0 },
      { id: 'gemma3:27b', name: 'Gemma 3 27B', ctx: '128K', out: '8K', tags: ['local'], $in: 0, $out: 0 },
      { id: 'gemma3:12b', name: 'Gemma 3 12B', ctx: '128K', out: '8K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'gemma3:4b', name: 'Gemma 3 4B', ctx: '128K', out: '8K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'phi4', name: 'Phi-4 14B', ctx: '16K', out: '4K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'phi4-mini', name: 'Phi-4 Mini 3.8B', ctx: '16K', out: '4K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'mistral:7b', name: 'Mistral 7B', ctx: '32K', out: '8K', tags: ['local', 'fast'], $in: 0, $out: 0 },
      { id: 'mixtral:8x7b', name: 'Mixtral 8x7B', ctx: '32K', out: '8K', tags: ['local'], $in: 0, $out: 0 },

      { id: 'deepseek-r1:70b', name: 'DeepSeek R1 70B', ctx: '64K', out: '8K', tags: ['local', 'reason'], $in: 0, $out: 0 },
      { id: 'deepseek-r1:32b', name: 'DeepSeek R1 32B', ctx: '64K', out: '8K', tags: ['local', 'reason'], $in: 0, $out: 0 },
      { id: 'deepseek-r1:14b', name: 'DeepSeek R1 14B', ctx: '64K', out: '8K', tags: ['local', 'reason', 'fast'], $in: 0, $out: 0 },
      { id: 'deepseek-r1:8b', name: 'DeepSeek R1 8B', ctx: '64K', out: '8K', tags: ['local', 'reason', 'fast'], $in: 0, $out: 0 },
      { id: 'qwq:32b', name: 'QwQ 32B', ctx: '32K', out: '8K', tags: ['local', 'reason'], $in: 0, $out: 0 },
      { id: 'marco-o1:7b', name: 'Marco-O1 7B', ctx: '128K', out: '8K', tags: ['local', 'reason'], $in: 0, $out: 0 },

      { id: 'llava:34b', name: 'LLaVA 34B', ctx: '4K', out: '4K', tags: ['local', 'vision'], $in: 0, $out: 0 },
      { id: 'minicpm-v:8b', name: 'MiniCPM-V 8B', ctx: '16K', out: '4K', tags: ['local', 'vision'], $in: 0, $out: 0 },
    ],
  },


  custom: {
    label: 'Custom  (Any OpenAI-compatible endpoint)',
    color: C.gray,
    badge: '',
    url: '',
    note: 'dardcor config --set provider=custom --set base_url=http://... --set model=<id>',
    models: [],
  },
};


const TAG_COLORS = {
  free: `${C.green}free${C.reset}`,
  fast: `${C.cyan}fast${C.reset}`,
  cheap: `${C.yellow}cheap${C.reset}`,
  local: `${C.yellow}local${C.reset}`,
  vision: `${C.blue}vision${C.reset}`,
  code: `${C.purple}code${C.reset}`,
  reason: `${C.orange}reason${C.reset}`,
  long: `${C.teal}long-ctx${C.reset}`,
};

function renderTags(tags) {
  return tags.map(t => TAG_COLORS[t] || t).join(' ');
}

function renderPrice(m) {
  if (m.tags.includes('free') || m.tags.includes('local')) return `${C.green}$0.00${C.reset}`;
  if (!m.$in && !m.$out) return '';
  return `${C.dim}$${m.$in.toFixed(2)}↑ $${m.$out.toFixed(2)}↓/1M${C.reset}`;
}

function hr(char = '─') {
  const w = Math.min((process.stdout.columns || 100) - 2, 90);
  return `${C.dim}${char.repeat(w)}${C.reset}`;
}


function totalModels() {
  return Object.values(MODELS).reduce((n, p) => n + p.models.length, 0);
}


export async function handleModels(opts = {}) {
  const cfg = loadConfig();
  const activeProvider = cfg.provider || 'local';
  const filter = opts.filter || opts.f || '';
  const search = opts.search || opts.s || '';
  const provOpt = opts.provider || opts.p || '';
  const pricing = opts.pricing !== undefined;
  const listAll = opts.all !== undefined;
  const count = opts.count !== undefined;

  if (count) {
    console.log(`\n  ${C.bold}${totalModels()}${C.reset} models across ${C.bold}${Object.keys(MODELS).length}${C.reset} providers\n`);
    return;
  }


  console.log('');
  console.log(`${C.bold}${C.purple}  Model Directory${C.reset}  ${C.dim}${totalModels()} models across ${Object.keys(MODELS).length} providers${C.reset}  ${C.gray}[active: ${activeProvider}]${C.reset}`);
  if (filter) console.log(`  ${C.dim}Filtered by tag:${C.reset} ${C.cyan}${filter}${C.reset}`);
  if (search) console.log(`  ${C.dim}Search:${C.reset} ${C.cyan}${search}${C.reset}`);
  if (provOpt) console.log(`  ${C.dim}Provider:${C.reset} ${C.cyan}${provOpt}${C.reset}`);
  console.log('');

  const lowerSearch = search.toLowerCase();
  const lowerFilter = filter.toLowerCase();
  let shown = 0;

  for (const [key, provider] of Object.entries(MODELS)) {

    if (provOpt && !key.startsWith(provOpt.toLowerCase())) continue;
    if (key === 'custom' && !listAll) continue;


    let models = provider.models;
    if (lowerFilter) models = models.filter(m => m.tags.includes(lowerFilter));
    if (lowerSearch) models = models.filter(m =>
      m.id.toLowerCase().includes(lowerSearch) ||
      m.name.toLowerCase().includes(lowerSearch)
    );
    if (models.length === 0 && (lowerFilter || lowerSearch)) continue;

    const isActive = key === activeProvider;
    const activeMark = isActive ? ` ${C.green}${C.bold}← ACTIVE${C.reset}` : '';
    const badge = provider.badge ? ` ${C.yellow}${provider.badge}${C.reset}` : '';

    console.log(`${provider.color}${C.bold}${provider.label}${C.reset}${badge}${activeMark}`);
    if (provider.note) console.log(`  ${C.dim}${provider.note}${C.reset}`);
    console.log(hr());

    if (models.length === 0) {
      console.log(`  ${C.dim}(no models — run: ollama pull <model>)${C.reset}`);
    } else {
      models.forEach(m => {
        const name = `${C.white}${m.name}${C.reset}`;
        const id = `${C.dim}${m.id}${C.reset}`;
        const ctx = `${C.gray}${m.ctx}${C.reset}`;
        const tags = renderTags(m.tags);
        const price = pricing ? `  ${renderPrice(m)}` : '';


        const nameWidth = 28;
        const namePadded = m.name.padEnd(nameWidth).slice(0, nameWidth);
        const ctxPadded = ('ctx:' + m.ctx).padEnd(10);

        console.log(
          `  ${C.white}${C.bold}${namePadded}${C.reset}` +
          `  ${C.gray}${ctxPadded}${C.reset}` +
          `  ${tags}` +
          `${price}`
        );


        console.log(`  ${C.dim}  ${m.id}${C.reset}`);
      });
    }

    console.log('');
    shown += models.length;
  }


  console.log(hr('═'));
  console.log(`  ${C.bold}${shown}${C.reset} ${C.dim}models shown${C.reset}`);
  console.log('');


  console.log(`${C.bold}Tags:${C.reset}  ` +
    Object.entries(TAG_COLORS).map(([k, v]) => `${v}=${k}`).join('  ')
  );
  console.log('');


  console.log(`${C.bold}Quick Commands:${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --filter code${C.reset}       ${C.dim}show only coding models${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --filter reason${C.reset}     ${C.dim}show only reasoning models${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --filter free${C.reset}       ${C.dim}show only free models${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --filter local${C.reset}      ${C.dim}show only local models (Ollama)${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --search gpt${C.reset}        ${C.dim}search by name/id${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --provider groq${C.reset}     ${C.dim}show one provider only${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --pricing${C.reset}           ${C.dim}show cost per 1M tokens${C.reset}`);
  console.log(`  ${C.cyan}dardcor models --count${C.reset}             ${C.dim}show total count only${C.reset}`);
  console.log('');
  console.log(`${C.bold}Set model:${C.reset}`);
  console.log(`  ${C.cyan}dardcor config --set provider=groq${C.reset}`);
  console.log(`  ${C.cyan}dardcor config --set model=llama-3.3-70b-versatile${C.reset}`);
  console.log('');
}
