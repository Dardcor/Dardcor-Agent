import React, { useState, useEffect } from 'react'
import ChatPanel from './components/ChatPanel'
import Terminal from './components/Terminal'
import FileExplorer from './components/FileExplorer'
import Workspace from './components/Workspace'
import SystemDashboard from './components/SystemDashboard'
import wsService from './services/websocket'

const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState('chat')
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)

  useEffect(() => {
    wsService.connect().catch(() => { })
    return () => {
      wsService.disconnect()
    }
  }, [])

  const renderContent = () => {
    switch (activeTab) {
      case 'chat':      return <ChatPanel />
      case 'terminal':  return <Terminal />
      case 'explorer':  return <FileExplorer />
      case 'workspace': return <Workspace />
      case 'system':    return <SystemDashboard />
      default:          return <ChatPanel />
    }
  }

  return (
    <div className={`app-container ${isSidebarOpen ? 'is-open' : ''}`}>
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-top-row">
            <div className="sidebar-logo">
              <div className="sidebar-logo-img"></div>
              <div className="sidebar-logo-text">
                <h1>Dardcor OK</h1>
              </div>
            </div>
            <button className="close-sidebar-btn" onClick={() => setIsSidebarOpen(false)}>×</button>
          </div>
          
          <div className="sidebar-separator"></div>

          <div className="sidebar-action-group">
            <button className="action-btn history-btn">
              <span>🕒</span> Riwayat
            </button>
            <button className="action-btn new-chat-btn" onClick={() => setActiveTab('chat')}>
              <span>+</span> Baru
            </button>
          </div>
        </div>

        <nav className="sidebar-nav">
          {[
            { id: 'chat',      label: 'Agent Chat',      icon: '💬' },
            { id: 'explorer',  label: 'File Explorer',  icon: '📂' },
            { id: 'terminal',  label: 'Terminal',       icon: '💻' },
            { id: 'system',    label: 'System Monitor', icon: '📊' },
            { id: 'workspace', label: 'Workspace',      icon: '🛠️' },
          ].map(tab => (
            <button
              key={tab.id}
              className={`sidebar-nav-item ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              <span className="nav-tab-icon">{tab.icon}</span>
              {tab.label}
            </button>
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
            <h2 style={{ textTransform: 'capitalize' }}>{activeTab}</h2>
          </div>
          <div className="header-actions">
            <button 
              className={`header-btn ${activeTab === 'chat' ? 'active' : ''}`}
              onClick={() => setActiveTab('chat')}
            >
              Chat
            </button>
            <button 
              className={`header-btn ${activeTab === 'terminal' ? 'active' : ''}`}
              onClick={() => setActiveTab('terminal')}
            >
              Terminal
            </button>
          </div>
        </header>
        
        <div className="tab-content" style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
          {renderContent()}
        </div>
      </main>
    </div>
  )
}

export default App
