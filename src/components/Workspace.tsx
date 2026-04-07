import React, { useState, useEffect } from 'react'

interface ProviderConfig {
  provider: string
  model: string
  apiKey: string
  port: string
}

const PROVIDERS = [
  { id: 'local',      label: 'Local (Built-in)',   models: ['dardcor-agent'],            needsKey: false },
  { id: 'openai',     label: 'OpenAI',              models: ['gpt-4o', 'gpt-4.1', 'o3'], needsKey: true },
  { id: 'anthropic',  label: 'Anthropic (Claude)',  models: ['claude-opus', 'claude-sonnet', 'claude-haiku'], needsKey: true },
  { id: 'gemini',     label: 'Google Gemini',       models: ['gemini-pro', 'gemini-flash'], needsKey: true },
  { id: 'deepseek',   label: 'DeepSeek',            models: ['deepseek-chat', 'deepseek-reasoner'], needsKey: true },
  { id: 'openrouter', label: 'OpenRouter',          models: ['anthropic/claude-opus', 'openai/gpt-4o'], needsKey: true },
  { id: 'ollama',     label: 'Ollama (Local LLM)',  models: ['llama3', 'qwen', 'mistral', 'codellama'], needsKey: false },
]

const Workspace: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'workspace' | 'provider' | 'about'>('workspace')
  const [workspacePath, setWorkspacePath] = useState('')
  const [wsStatus, setWsStatus] = useState<'idle' | 'success' | 'error'>('idle')

  const [providerCfg, setProviderCfg] = useState<ProviderConfig>({
    provider: 'local',
    model: 'dardcor-agent',
    apiKey: '',
    port: '25000',
  })
  const [providerStatus, setProviderStatus] = useState<'idle' | 'success' | 'error'>('idle')
  const [showKey, setShowKey] = useState(false)

  useEffect(() => {
    const saved = localStorage.getItem('dardcor_workspace_path')
    setWorkspacePath(saved || 'C:\\Users\\user\\Documents\\Dardcor-Workspace')

    const savedCfg = localStorage.getItem('dardcor_provider_cfg')
    if (savedCfg) {
      try { setProviderCfg(JSON.parse(savedCfg)) } catch { }
    }
  }, [])

  const selectedProvider = PROVIDERS.find(p => p.id === providerCfg.provider) || PROVIDERS[0]

  const handleSaveWorkspace = (e: React.FormEvent) => {
    e.preventDefault()
    if (workspacePath.trim()) {
      localStorage.setItem('dardcor_workspace_path', workspacePath.trim())
      setWsStatus('success')
      setTimeout(() => setWsStatus('idle'), 3000)
    } else {
      setWsStatus('error')
    }
  }

  const handleSaveProvider = (e: React.FormEvent) => {
    e.preventDefault()
    localStorage.setItem('dardcor_provider_cfg', JSON.stringify(providerCfg))
    fetch('/api/provider/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(providerCfg),
    }).catch(() => { })
    setProviderStatus('success')
    setTimeout(() => setProviderStatus('idle'), 3000)
  }

  const handleProviderChange = (providerId: string) => {
    const p = PROVIDERS.find(x => x.id === providerId)!
    setProviderCfg(prev => ({
      ...prev,
      provider: providerId,
      model: p.models[0],
      apiKey: p.needsKey ? prev.apiKey : '',
    }))
  }

  return (
    <div className="workspace-container">
      <div className="ws-tabs">
        {[
          { id: 'workspace', label: '💼 Workspace', },
          { id: 'provider',  label: '🤖 AI Provider', },
          { id: 'about',     label: 'ℹ️ About', },
        ].map(tab => (
          <button
            key={tab.id}
            id={`ws-tab-${tab.id}`}
            className={`ws-tab-btn ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id as typeof activeTab)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'workspace' && (
        <div className="workspace-card" style={{ animation: 'fadeIn 0.3s ease' }}>
          <div className="workspace-header">
            <div className="workspace-icon">🛠️</div>
            <div className="workspace-title">
              <h2>Workspace Configuration</h2>
              <p>Define the primary working directory for your Agent sessions.</p>
            </div>
          </div>

          <form className="workspace-form" onSubmit={handleSaveWorkspace}>
            <div className="form-group">
              <label htmlFor="workspace-path">Primary Directory Path</label>
              <div className="input-with-button">
                <input
                  type="text"
                  id="workspace-path"
                  value={workspacePath}
                  onChange={e => setWorkspacePath(e.target.value)}
                  placeholder="Path..."
                  spellCheck={false}
                />
                <button type="submit" className="save-btn" id="save-workspace-btn">Save</button>
              </div>
              <p className="hint">
                💡 Local path for terminal and explorer.
              </p>
            </div>
          </form>

          {wsStatus === 'success' && (
            <div className="workspace-alert success">✅ Workspace path saved!</div>
          )}
          {wsStatus === 'error' && (
            <div className="workspace-alert error">❌ Please enter a valid path.</div>
          )}

          <div className="workspace-guide">
            <h3>How workspace works</h3>
            <div className="guide-grid">
              <div className="guide-item">
                <h4>🔒 Security Boundary</h4>
                <p>Limits Agent scope to your project folder.</p>
              </div>
              <div className="guide-item">
                <h4>🚀 Quick Access</h4>
                <p>Terminal auto-navigates to this path.</p>
              </div>
              <div className="guide-item">
                <h4>📂 Organized Output</h4>
                <p>Agent files are centralized here.</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'provider' && (
        <div className="workspace-card" style={{ animation: 'fadeIn 0.3s ease' }}>
          <div className="workspace-header">
            <div className="workspace-icon">🤖</div>
            <div className="workspace-title">
              <h2>AI Provider Configuration</h2>
              <p>Configure your LLM provider.</p>
            </div>
          </div>

          <form className="workspace-form" onSubmit={handleSaveProvider}>
            <div className="form-group">
              <label>AI Provider</label>
              <div className="provider-grid">
                {PROVIDERS.map(p => (
                  <button
                    key={p.id}
                    type="button"
                    id={`provider-${p.id}`}
                    className={`provider-pill ${providerCfg.provider === p.id ? 'selected' : ''}`}
                    onClick={() => handleProviderChange(p.id)}
                  >
                    {!p.needsKey && <span className="free-badge">FREE</span>}
                    {p.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="model-select">Model</label>
              <select
                id="model-select"
                className="form-select"
                value={providerCfg.model}
                onChange={e => setProviderCfg(p => ({ ...p, model: e.target.value }))}
              >
                {selectedProvider.models.map(m => (
                  <option key={m} value={m}>{m}</option>
                ))}
              </select>
            </div>

            {selectedProvider.needsKey && (
              <div className="form-group">
                <label htmlFor="api-key">API Key</label>
                <div className="input-with-button">
                  <input
                    type={showKey ? 'text' : 'password'}
                    id="api-key"
                    value={providerCfg.apiKey}
                    onChange={e => setProviderCfg(p => ({ ...p, apiKey: e.target.value }))}
                    placeholder={`API Key...`}
                    spellCheck={false}
                    autoComplete="off"
                  />
                  <button
                    type="button"
                    className="save-btn"
                    onClick={() => setShowKey(s => !s)}
                    style={{ minWidth: 60 }}
                  >
                    {showKey ? 'Hide' : 'Show'}
                  </button>
                </div>
              </div>
            )}

            <div className="form-group">
              <label htmlFor="gateway-port">Gateway Port</label>
              <input
                type="number"
                id="gateway-port"
                className="form-input-sm"
                value={providerCfg.port}
                onChange={e => setProviderCfg(p => ({ ...p, port: e.target.value }))}
                min={1024}
                max={65535}
              />
            </div>

            <button type="submit" id="save-provider-btn" className="save-btn" style={{ width: '100%', padding: '12px' }}>
              💾 Save Provider Configuration
            </button>
          </form>

          {providerStatus === 'success' && (
            <div className="workspace-alert success">✅ Provider configuration saved!</div>
          )}

          <div className="provider-tip">
            <strong>💡 Tip:</strong> Run <code>dardcor run</code> to apply changes.
          </div>
        </div>
      )}

      {activeTab === 'about' && (
        <div className="workspace-card" style={{ animation: 'fadeIn 0.3s ease' }}>
          <div className="workspace-header">
            <div className="workspace-icon">⚡</div>
            <div className="workspace-title">
              <h2>Dardcor Agent</h2>
              <p>Your Personal AI — Any OS, Any Platform.</p>
            </div>
          </div>

          <div className="about-grid">
            <div className="about-item">
              <div className="about-icon">🚀</div>
              <div>
                <h4>dardcor run</h4>
                <p>Gateway + WebUI Dashboard.</p>
              </div>
            </div>
            <div className="about-item">
              <div className="about-icon">💻</div>
              <div>
                <h4>dardcor cli</h4>
                <p>Interactive TUI agent.</p>
              </div>
            </div>
            <div className="about-item">
              <div className="about-icon">🧙</div>
              <div>
                <h4>dardcor onboard</h4>
                <p>Interactive setup wizard.</p>
              </div>
            </div>
            <div className="about-item">
              <div className="about-icon">🩺</div>
              <div>
                <h4>dardcor doctor</h4>
                <p>Health check.</p>
              </div>
            </div>
            <div className="about-item">
              <div className="about-icon">⚡</div>
              <div>
                <h4>ultrawork</h4>
                <p>Autonomous agent execution.</p>
              </div>
            </div>
            <div className="about-item">
              <div className="about-icon">🔌</div>
              <div>
                <h4>Multi-Provider</h4>
                <p>OpenAI, Claude, Gemini, DeepSeek, etc.</p>
              </div>
            </div>
          </div>

          <div className="about-links">
            <a href="https://github.com/syahr/Dardcor-Agent" target="_blank" rel="noopener noreferrer" className="about-link-btn" id="github-link">
              📦 GitHub Repository
            </a>
          </div>
        </div>
      )}
    </div>
  )
}

export default Workspace
