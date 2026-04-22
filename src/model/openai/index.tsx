import React from 'react'

const OpenAIView: React.FC = () => {
  const [isActive, setIsActive] = React.useState(false)
  const [apiKey, setApiKey] = React.useState('')
  const [model, setModel] = React.useState('gpt-4o')
  const [isSaving, setIsSaving] = React.useState(false)

  React.useEffect(() => {
    fetch('/api/model/active')
      .then(res => res.json())
      .then(data => {
        if (data.success && data.data) {
          if (data.data.openai !== undefined) setIsActive(data.data.openai)
          if (data.data.openai_api_key) setApiKey(data.data.openai_api_key)
          if (data.data.openai_model) setModel(data.data.openai_model)
        }
      })
      .catch(() => {})
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
        body: JSON.stringify({ ...current, openai: newState })
      })
    } catch {}
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
          openai_api_key: apiKey,
          openai_model: model
        })
      })
      alert('OpenAI configuration saved successfully!')
    } catch (err) {
      alert('Failed to save configuration')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="model-config-view card-premium animate-fade" style={{ maxWidth: '800px' }}>
      <div style={{
        background: isActive ? 'linear-gradient(135deg, rgba(16,163,127,0.1) 0%, rgba(124,58,237,0.08) 100%)' : 'rgba(255,255,255,0.03)',
        border: isActive ? '1px solid rgba(16,163,127,0.25)' : '1px solid rgba(255,255,255,0.08)',
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
             position: 'absolute', top: 0, left: 0, width: '4px', height: '100%', background: '#10a37f'
          }}></div>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
           <div style={{
             width: '10px', height: '10px', borderRadius: '50%',
             background: isActive ? '#10a37f' : '#4b5563',
             boxShadow: isActive ? '0 0 12px #10a37f, 0 0 20px rgba(16,163,127,0.4)' : 'none',
             transition: 'all 0.3s ease'
           }}></div>
           <div style={{ display: 'flex', flexDirection: 'column' }}>
             <strong style={{
               fontSize: '15px',
               color: isActive ? '#fff' : 'rgba(255,255,255,0.6)',
             }}>Aktifkan provider ini</strong>
             {isActive && (
               <span style={{ fontSize: '11px', color: 'rgba(16,163,127,0.7)', marginTop: '2px' }}>
                 OpenAI is currently active
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
            background: isActive ? '#10a37f' : 'rgba(255,255,255,0.1)',
            border: isActive ? '1px solid rgba(16,163,127,0.5)' : '1px solid rgba(255,255,255,0.1)',
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
         <h2>OpenAI Configuration</h2>
         <p>Link your OpenAI API key to start using GPT models with Dardcor Agent.</p>
      </div>

      <div className="config-form">
         <div className="form-group">
            <label>API Key</label>
            <input 
              type="password" 
              placeholder="sk-..." 
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
               <option value="gpt-4o">GPT-4o</option>
               <option value="gpt-4-turbo">GPT-4 Turbo</option>
               <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
               <option value="o1-preview">o1-preview</option>
               <option value="o1-mini">o1-mini</option>
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

export default OpenAIView
