import React from 'react'

const DeepSeekView: React.FC = () => {
  const [isActive, setIsActive] = React.useState(false)
  const [apiKey, setApiKey] = React.useState('')
  const [model, setModel] = React.useState('deepseek-coder')
  const [isSaving, setIsSaving] = React.useState(false)

  React.useEffect(() => {
    fetch('/api/model/active')
      .then(res => res.json())
      .then(data => {
        if (data.success && data.data) {
          if (data.data.deepseek !== undefined) setIsActive(data.data.deepseek)
          if (data.data.deepseek_api_key) setApiKey(data.data.deepseek_api_key)
          if (data.data.deepseek_model) setModel(data.data.deepseek_model)
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
        body: JSON.stringify({ ...current, deepseek: newState })
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
          deepseek_api_key: apiKey,
          deepseek_model: model
        })
      })
      alert('DeepSeek configuration saved successfully!')
    } catch (err) {
      alert('Failed to save configuration')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="model-config-view card-premium animate-fade" style={{ maxWidth: '800px' }}>
      <div style={{
        background: isActive ? 'linear-gradient(135deg, rgba(61,102,245,0.1) 0%, rgba(124,58,237,0.08) 100%)' : 'rgba(255,255,255,0.03)',
        border: isActive ? '1px solid rgba(61,102,245,0.25)' : '1px solid rgba(255,255,255,0.08)',
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
             position: 'absolute', top: 0, left: 0, width: '4px', height: '100%', background: '#3d66f5'
          }}></div>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
           <div style={{
             width: '10px', height: '10px', borderRadius: '50%',
             background: isActive ? '#3d66f5' : '#4b5563',
             boxShadow: isActive ? '0 0 12px #3d66f5, 0 0 20px rgba(61,102,245,0.4)' : 'none',
             transition: 'all 0.3s ease'
           }}></div>
           <div style={{ display: 'flex', flexDirection: 'column' }}>
             <strong style={{
               fontSize: '15px',
               color: isActive ? '#fff' : 'rgba(255,255,255,0.6)',
             }}>Aktifkan provider ini</strong>
             {isActive && (
               <span style={{ fontSize: '11px', color: 'rgba(61,102,245,0.7)', marginTop: '2px' }}>
                 DeepSeek is currently active
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
            background: isActive ? '#3d66f5' : 'rgba(255,255,255,0.1)',
            border: isActive ? '1px solid rgba(61,102,245,0.5)' : '1px solid rgba(255,255,255,0.1)',
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
         <h2>DeepSeek Configuration</h2>
         <p>Link your DeepSeek API key to start using DeepSeek models with Dardcor Agent.</p>
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
               <option value="deepseek-chat">deepseek-chat</option>
               <option value="deepseek-coder">deepseek-coder</option>
               <option value="deepseek-v3">deepseek-v3</option>
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

export default DeepSeekView
