import React, { useState, useEffect } from 'react'

const Workspace: React.FC = () => {
  const [workspacePath, setWorkspacePath] = useState('')
  const [wsStatus, setWsStatus] = useState<'idle' | 'success' | 'error'>('idle')

  useEffect(() => {

    fetch('/api/workspace/config')
      .then(res => res.json())
      .then(data => {
        if (data.success && data.data && data.data.path) {
          setWorkspacePath(data.data.path)
          localStorage.setItem('dardcor_workspace_path', data.data.path)
        } else {

          const savedLocal = localStorage.getItem('dardcor_workspace_path')
          if (savedLocal) {
            setWorkspacePath(savedLocal)

            saveToBackend(savedLocal)
          } else {

            fetch('/api/files/workspace/default')
              .then(res => res.json())
              .then(d => {
                if (d.success && d.data) {
                  setWorkspacePath(d.data)
                  localStorage.setItem('dardcor_workspace_path', d.data)
                  saveToBackend(d.data)
                }
              })
          }
        }
      })
      .catch(() => {
        setWorkspacePath('C:\\Dardcor-Workspace')
      })
  }, [])

  const saveToBackend = async (path: string) => {
    try {
      const res = await fetch('/api/workspace/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: path.trim() })
      })
      return res.ok
    } catch (err) {
      console.error('Failed to sync workspace to backend:', err)
      return false
    }
  }

  const handleSaveWorkspace = async (e: React.FormEvent) => {
    e.preventDefault()
    const trimmedPath = workspacePath.trim()
    if (trimmedPath) {
      // Create folder if it doesn't exist
      try {
        await fetch('/api/files/mkdir', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ path: trimmedPath })
        })
      } catch (e) { console.error(e) }

      const success = await saveToBackend(trimmedPath)
      if (success) {
        localStorage.setItem('dardcor_workspace_path', trimmedPath)
        setWsStatus('success')
        setTimeout(() => setWsStatus('idle'), 3000)
      } else {
        setWsStatus('error')
        setTimeout(() => setWsStatus('idle'), 3000)
      }
    } else {
      setWsStatus('error')
      setTimeout(() => setWsStatus('idle'), 3000)
    }
  }

  const handleOpenWorkspace = () => {
    if (!workspacePath.trim()) return

    let path = workspacePath.trim().replace(/\//g, '\\')

    // Remove trailing \ for better explorer compatibility
    if (path.length > 3 && path.endsWith('\\')) path = path.slice(0, -1)
    if (/^[a-zA-Z]:$/.test(path)) path += '\\'

    const cmd = window.navigator.platform.includes('Win')
      ? `explorer "${path}"`
      : `open "${path}"`

    fetch('/api/command', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: cmd })
    })
  }

  return (
    <div className="workspace-container">
      <div className="workspace-card" style={{ animation: 'fadeIn 0.3s ease', marginTop: '20px' }}>
        <div className="workspace-header">
          <div className="workspace-icon">🛠️</div>
          <div className="workspace-title">
            <h2>Workspace Configuration</h2>
            <p>Define the primary working directory for your Agent sessions.</p>
          </div>
        </div>

        <form className="workspace-form" onSubmit={handleSaveWorkspace}>
          <div className="form-group">
            <label htmlFor="workspace-path">PRIMARY DIRECTORY PATH</label>
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
              <button type="button" className="open-btn" onClick={handleOpenWorkspace}>Open</button>
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
    </div>
  )
}

export default Workspace




