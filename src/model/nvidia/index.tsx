import React from 'react'

const NvidiaView: React.FC = () => {
  const [isActive, setIsActive] = React.useState(false)
  const [apiKey, setApiKey] = React.useState('')
  const [model, setModel] = React.useState('minimaxai/minimax-m2.7')
  const [isSaving, setIsSaving] = React.useState(false)

  React.useEffect(() => {
    fetch('/api/model/active')
      .then(res => res.json())
      .then(data => {
        if (data.success && data.data) {
          if (data.data.nvidia !== undefined) setIsActive(data.data.nvidia)
          if (data.data.nvidia_api_key) setApiKey(data.data.nvidia_api_key)
          if (data.data.nvidia_model) setModel(data.data.nvidia_model)
        }
      })
      .catch(() => { })
  }, [])

  const toggleModel = async () => {
    const newState = !isActive
    setIsActive(newState)
    try {
      const resp = await fetch('/api/model/active')
      const data = await resp.json()
      const current = data.success ? data.data : {}
      await fetch('/api/model/active', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...current, nvidia: newState })
      })
    } catch { }
  }

  const saveConfig = async () => {
    setIsSaving(true)
    try {
      const resp = await fetch('/api/model/active')
      const data = await resp.json()
      const current = data.success ? data.data : {}

      await fetch('/api/model/active', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...current,
          nvidia_api_key: apiKey,
          nvidia_model: model
        })
      })
      alert('NVIDIA configuration saved successfully!')
    } catch (err) {
      alert('Failed to save configuration')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="model-config-view card-premium animate-fade" style={{ maxWidth: '800px' }}>
      <div style={{
        background: isActive ? 'linear-gradient(135deg, rgba(118,185,0,0.1) 0%, rgba(124,58,237,0.08) 100%)' : 'rgba(255,255,255,0.03)',
        border: isActive ? '1px solid rgba(118,185,0,0.25)' : '1px solid rgba(255,255,255,0.08)',
        borderRadius: '12px',
        padding: '14px 22px',
        display: 'flex',
        alignItems: 'center',
        gap: '15px',
        marginBottom: '24px',
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        position: 'relative',
        overflow: 'hidden'
      }}>
        {isActive && (
          <div style={{
            position: 'absolute', top: 0, left: 0, width: '4px', height: '100%', background: '#76b900'
          }}></div>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
          <div style={{
            width: '10px', height: '10px', borderRadius: '50%',
            background: isActive ? '#76b900' : '#4b5563',
            boxShadow: isActive ? '0 0 12px #76b900, 0 0 20px rgba(118,185,0,0.4)' : 'none',
            transition: 'all 0.3s ease'
          }}></div>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <strong style={{
              fontSize: '15px',
              color: isActive ? '#fff' : 'rgba(255,255,255,0.6)',
            }}>Aktifkan provider ini</strong>
            {isActive && (
              <span style={{ fontSize: '11px', color: 'rgba(118,185,0,0.7)', marginTop: '2px' }}>
                NVIDIA API is currently active
              </span>
            )}
          </div>
        </div>

        <div
          onClick={toggleModel}
          style={{
            width: '48px',
            height: '26px',
            borderRadius: '13px',
            background: isActive ? '#76b900' : 'rgba(255,255,255,0.1)',
            border: isActive ? '1px solid rgba(118,185,0,0.5)' : '1px solid rgba(255,255,255,0.1)',
            position: 'relative',
            cursor: 'pointer',
            transition: 'all 0.4s',
          }}
        >
          <div style={{
            width: '18px',
            height: '18px',
            borderRadius: '50%',
            background: '#fff',
            position: 'absolute',
            top: '3px',
            left: isActive ? '25px' : '4px',
            transition: 'all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
          }}></div>
        </div>
      </div>

      <div className="config-header">
        <h2>NVIDIA Configuration</h2>
        <p>Link your NVIDIA API key to start using specialized models like minimax-m2.7 via NVIDIA's catalog.</p>
      </div>

      <div className="config-form">
        <div className="form-group">
          <label>API Key</label>
          <input
            type="password"
            placeholder="nvapi-..."
            className="input-premium"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
          />
        </div>

        <div className="form-group">
          <label>Select Model</label>
          <select
            className="input-premium"
            value={model}
            onChange={(e) => setModel(e.target.value)}
          >

            <option value="minimaxai/minimax-m2.7">minimaxai/minimax-m2.7 ⭐ Recommended</option>

            <option value="nvidia/ising-calibration-1-35b-a3b">nvidia/ising-calibration-1-35b-a3b (calibration)</option>

            <option value="nvidia/llama-3.1-nemotron-70b-instruct">nvidia/llama-3.1-nemotron-70b-instruct</option>
            <option value="nvidia/llama-3.1-405b-instruct">nvidia/llama-3.1-405b-instruct</option>

            <option value="meta/llama-3.3-70b-instruct">meta/llama-3.3-70b-instruct</option>
            <option value="meta/llama-3.1-70b-instruct">meta/llama-3.1-70b-instruct</option>
            <option value="meta/llama-3.1-8b-instruct">meta/llama-3.1-8b-instruct</option>

            <option value="mistralai/mixtral-8x22b-instruct-v0.1">mistralai/mixtral-8x22b-instruct-v0.1</option>
            <option value="mistralai/mistral-large-2-instruct">mistralai/mistral-large-2-instruct</option>

            <option value="qwen/qwen2.5-72b-instruct">qwen/qwen2.5-72b-instruct</option>
            <option value="deepseek-ai/deepseek-r1-distill-llama-70b">deepseek-ai/deepseek-r1-distill-llama-70b</option>
            <option value="google/gemma-2-27b-it">google/gemma-2-27b-it</option>
            <option value="microsoft/phi-3.5-mini-instruct">microsoft/phi-3.5-mini-instruct</option>
          </select>
        </div>

        <button
          className="btn-primary-glow"
          onClick={saveConfig}
          disabled={isSaving}
        >
          {isSaving ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>
    </div>
  )
}

export default NvidiaView
