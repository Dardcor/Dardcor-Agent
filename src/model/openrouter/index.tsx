import React from 'react'

const OpenRouterView: React.FC = () => {
  const [isActive, setIsActive] = React.useState(false)

  React.useEffect(() => {
    fetch('/api/model/active')
      .then(res => res.json())
      .then(data => {
        if (data.success && data.data && data.data.openrouter !== undefined) {
          setIsActive(data.data.openrouter)
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
        body: JSON.stringify({ ...current, openrouter: newState })
      })
    } catch { }
  }

  return (
    <div className="model-config-view card-premium animate-fade" style={{ maxWidth: '800px' }}>

      <div style={{
        background: isActive ? 'linear-gradient(135deg, rgba(255,107,0,0.1) 0%, rgba(124,58,237,0.08) 100%)' : 'rgba(255,255,255,0.03)',
        border: isActive ? '1px solid rgba(255,107,0,0.25)' : '1px solid rgba(255,255,255,0.08)',
        borderRadius: '12px',
        padding: '14px 22px',
        display: 'flex',
        alignItems: 'center',
        gap: '15px',
        marginBottom: '24px',
        boxShadow: isActive ? '0 0 25px rgba(255,107,0,0.05)' : 'none',
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        position: 'relative',
        overflow: 'hidden'
      }}>
        {isActive && (
          <div style={{
            position: 'absolute', top: 0, left: 0, width: '4px', height: '100%', background: '#ff6b00'
          }}></div>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
          <div style={{
            width: '10px', height: '10px', borderRadius: '50%',
            background: isActive ? '#ff6b00' : '#4b5563',
            boxShadow: isActive ? '0 0 12px #ff6b00, 0 0 20px rgba(255,107,0,0.4)' : 'none',
            flexShrink: 0,
            transition: 'all 0.3s ease'
          }}></div>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <strong style={{
              fontSize: '15px',
              letterSpacing: '0.2px',
              color: isActive ? '#fff' : 'rgba(255,255,255,0.6)',
              transition: 'all 0.3s ease'
            }}>Aktifkan model ini</strong>
            {isActive && (
              <span style={{ fontSize: '11px', color: 'rgba(255,107,0,0.7)', marginTop: '2px' }}>
                OpenRouter is currently active
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
            background: isActive ? '#ff6b00' : 'rgba(255,255,255,0.1)',
            border: isActive ? '1px solid rgba(255,107,0,0.5)' : '1px solid rgba(255,255,255,0.1)',
            position: 'relative',
            cursor: 'pointer',
            transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
            boxShadow: isActive ? '0 0 15px rgba(255,107,0,0.3)' : 'none'
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
            boxShadow: '0 2px 5px rgba(0,0,0,0.4)'
          }}></div>
        </div>
      </div>

      <div className="config-header">
        <h2>OpenRouter Configuration</h2>
        <p>Link your API key to start using OpenRouter with Dardcor Agent.</p>
      </div>

      <div className="config-form">
        <div className="form-group">
          <label>API Key</label>
          <input type="password" placeholder="Enter OpenRouter API Key..." className="input-premium" />
        </div>

        <div className="form-group">
          <label>Select Model</label>
          <select className="input-premium" defaultValue="Select a model...">
            <option>Select a model...</option>
            <option>GPT-4o</option>
            <option>Claude 3.5 Sonnet</option>
            <option>Llama 3 70B</option>
          </select>
        </div>

        <button className="btn-primary-glow">Save Configuration</button>
      </div>
    </div>
  )
}

export default OpenRouterView