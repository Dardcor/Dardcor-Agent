import React, { useState, useEffect } from 'react'
import { Routes, Route, useNavigate, useLocation, Link, Navigate } from 'react-router-dom'
import ChatPanel from './components/ChatPanel'
import Terminal from './components/Terminal'
import FileExplorer from './components/FileExplorer'
import Workspace from './components/Workspace'
import SystemDashboard from './components/SystemDashboard'
import ModelSelector from './components/ModelSelector'
import ToolsPanel from './components/ToolsPanel'
import SkillsPanel from './components/SkillsPanel'
import wsService from './services/websocket'

const App: React.FC = () => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const navigate = useNavigate()
  const location = useLocation()

  const getActiveTab = () => {
    const path = location.pathname
    if (path.startsWith('/chat')) return 'chat'
    if (path.startsWith('/model')) return 'model'
    if (path === '/tools') return 'tools'
    if (path === '/skills') return 'skills'
    if (path === '/file-explorer') return 'explorer'
    if (path === '/terminal') return 'terminal'
    if (path === '/monitor') return 'system'
    if (path === '/workspace') return 'workspace'
    return 'chat'
  }

  const activeTab = getActiveTab()

  useEffect(() => {
    wsService.connect().catch(() => { })
    return () => {
      wsService.disconnect()
    }
  }, [])

  return (
    <div className="app-container">
      <aside className={`sidebar ${isSidebarOpen ? 'is-open' : ''}`}>
        <div className="sidebar-header">
          <div className="sidebar-top-row">
            <Link to="/chat" className="sidebar-logo" style={{ textDecoration: 'none' }}>
              <div className="sidebar-logo-img"></div>
              <div className="sidebar-logo-text">
                <h1>Dardcor</h1>
              </div>
            </Link>
            <button className="close-sidebar-btn" onClick={() => setIsSidebarOpen(false)}>×</button>
          </div>
        </div>
        
        <div className="sidebar-separator"></div>

        <div className="sidebar-action-group">
          <button className="action-btn history-btn" onClick={() => {
            if (location.pathname !== '/chat') navigate('/chat')
            setTimeout(() => document.dispatchEvent(new CustomEvent('toggle-history')), 50)
          }}>
            <span>🕒</span> Riwayat
          </button>
          <button className="action-btn new-chat-btn" onClick={() => {
            if (location.pathname !== '/chat') navigate('/chat')
            setTimeout(() => document.dispatchEvent(new CustomEvent('new-chat')), 50)
          }}>
            <span>+</span> Baru
          </button>
        </div>

        <nav className="sidebar-nav">
          {[
            { id: 'chat',        path: localStorage.getItem('last_conv_id') ? `/chat/${localStorage.getItem('last_conv_id')}` : '/chat', label: 'Agent Chat', icon: '💬' },
            { id: 'model',       path: '/model',         label: 'Model Settings', icon: '🤖' },
            { id: 'tools',       path: '/tools',         label: 'Tools',           icon: '🛠️' },
            { id: 'skills',      path: '/skills',        label: 'Skills',          icon: '🧠' },
            { id: 'explorer',    path: '/file-explorer', label: 'File Explorer',  icon: '📂' },
            { id: 'terminal',    path: '/terminal',      label: 'Terminal',       icon: '💻' },
            { id: 'system',      path: '/monitor',       label: 'System Monitor', icon: '📊' },
            { id: 'workspace',   path: '/workspace',     label: 'Workspace',      icon: '🏗️' },
          ].map(item => (
            <Link
              key={item.id}
              to={item.path}
              className={`sidebar-nav-item ${activeTab === item.id ? 'active' : ''}`}
              style={{ textDecoration: 'none' }}
            >
              <span className="nav-tab-icon">{item.icon}</span>
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="sidebar-footer">
          <div className="connection-status">
            <div className={`status-dot ${wsService.isConnected ? 'connected' : ''}`}></div>
            {wsService.isConnected ? 'Connected' : 'Disconnected'}
          </div>
        </div>
      </aside>
      
      <main className="main-content">
        <header className="main-header">
          <div className="header-title">
            {!isSidebarOpen && (
              <button className="hamburger-btn" onClick={() => setIsSidebarOpen(true)}>☰</button>
            )}
            <h2 style={{ textTransform: 'capitalize' }}>{activeTab === 'system' ? 'System Monitor' : activeTab}</h2>
          </div>

          <div className="header-actions">
            <Link 
              to="/chat"
              className={`header-btn ${activeTab === 'chat' ? 'active' : ''}`}
            >
              Chat
            </Link>
            <Link 
              to="/terminal"
              className={`header-btn ${activeTab === 'terminal' ? 'active' : ''}`}
            >
              Terminal
            </Link>
          </div>
        </header>
        
        <div className="tab-content">
          <Routes>
            <Route path="/" element={<Navigate to="/chat" replace />} />
            <Route path="/chat" element={<ChatPanel />} />
            <Route path="/chat/:id" element={<ChatPanel />} />
            <Route path="/model" element={<ModelSelector />} />
            <Route path="/model/:provider" element={<ModelSelector />} />
            <Route path="/tools" element={<ToolsPanel />} />
            <Route path="/skills" element={<SkillsPanel />} />
            <Route path="/file-explorer" element={<FileExplorer />} />
            <Route path="/terminal" element={<Terminal />} />
            <Route path="/monitor" element={<SystemDashboard />} />
            <Route path="/workspace" element={<Workspace />} />
            <Route path="*" element={<Navigate to="/chat" replace />} />
          </Routes>
        </div>
      </main>
    </div>
  )
}

export default App
