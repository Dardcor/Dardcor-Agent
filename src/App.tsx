import React, { useState, useEffect } from 'react'
import ChatPanel from './components/ChatPanel'
import FileExplorer from './components/FileExplorer'
import Terminal from './components/Terminal'
import SystemDashboard from './components/SystemDashboard'
import Workspace from './components/Workspace'
import { useWebSocket, useChat } from './hooks/useAgent'
import type { TabType } from './types'

const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabType>('chat')
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [isHistoryOpen, setIsHistoryOpen] = useState(false)
  const [editingConvId, setEditingConvId] = useState<string | null>(null)
  const [editingTitle, setEditingTitle] = useState('')
  const [activeMenuId, setActiveMenuId] = useState<string | null>(null)
  const { isConnected } = useWebSocket()
  const chat = useChat()

  useEffect(() => {
    if (isConnected) {
      chat.loadConversations()
    }
  }, [isConnected])

  const handleRefresh = () => {
    chat.loadConversations()
    // Trigger re-fetch for other tabs too if applicable
  }

  const tabs: { id: TabType; label: string; icon: string }[] = [
    { id: 'chat', label: 'Agent Chat', icon: '🤖' },
    { id: 'files', label: 'File Explorer', icon: '📁' },
    { id: 'terminal', label: 'Terminal', icon: '💻' },
    { id: 'system', label: 'System Monitor', icon: '📊' },
    { id: 'workspace', label: 'Workspace', icon: '🛠️' },
  ]

  const handleSelectConversation = (id: string) => {
    chat.loadConversation(id)
    setIsHistoryOpen(false)
    setActiveTab('chat')
  }

  const handleStartRename = (e: React.MouseEvent, id: string, title: string) => {
    e.stopPropagation()
    setEditingConvId(id)
    setEditingTitle(title.split(' (')[0]) // Strip (N messages) if present
    setActiveMenuId(null)
  }

  const handleRenameConfirm = (e: React.FormEvent) => {
    e.preventDefault()
    if (editingConvId && editingTitle.trim()) {
      chat.renameConversation(editingConvId, editingTitle.trim())
      setEditingConvId(null)
    }
  }

  const handleDeleteConversation = (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    if (window.confirm('Hapus percakapan ini secara permanen?')) {
      chat.deleteConversation(id)
      setActiveMenuId(null)
    }
  }

  const renderContent = () => {
    switch (activeTab) {
      case 'chat':
        return (
          <ChatPanel
            messages={chat.messages}
            isTyping={chat.isTyping}
            onSendMessage={chat.sendMessage}
          />
        )
      case 'files':
        return <FileExplorer />
      case 'terminal':
        return <Terminal />
      case 'system':
        return <SystemDashboard />
      case 'workspace':
        return <Workspace />
      default:
        return null
    }
  }

  return (
    <div className="app-container">
      {/* Sidebar */}
      <aside className={`sidebar ${isSidebarOpen ? 'is-open' : ''}`}>
        <div className="sidebar-header">
          <div className="sidebar-top-row">
            <div className="sidebar-logo">
              <div className="sidebar-logo-img" aria-label="Dardcor Logo" />
              <div className="sidebar-logo-text">
                <h1>Dardcor Agent</h1>
              </div>
            </div>
            <button className="close-sidebar-btn" onClick={() => setIsSidebarOpen(false)}>×</button>
          </div>
          
          <div className="sidebar-action-group">
            <button
              className="action-btn history-btn"
              onClick={() => setIsHistoryOpen(true)}
              id="history-btn"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M12 8v4l3 3" /><circle cx="12" cy="12" r="9" />
              </svg>
              Riwayat
            </button>
            <button
              className="action-btn new-chat-btn"
              onClick={chat.newConversation}
              id="new-chat-btn"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
              </svg>
              Baru
            </button>
          </div>
        </div>

        <nav className="sidebar-nav">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              className={`sidebar-nav-item ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
              id={`nav-${tab.id}`}
            >
              <span style={{ fontSize: '18px' }}>{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </nav>

        {/* Sidebar Footer */}
        <div className="sidebar-footer">
          <div className="connection-status">
            <div className={`status-dot ${isConnected ? 'connected' : ''}`} />
            <span>{isConnected ? 'Terhubung' : 'Terputus'}</span>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="main-content">
        <header className="main-header">
          <div className="header-title">
            <button className="hamburger-btn" onClick={() => setIsSidebarOpen(true)}>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="3" y1="12" x2="21" y2="12"></line>
                <line x1="3" y1="6" x2="21" y2="6"></line>
                <line x1="3" y1="18" x2="21" y2="18"></line>
              </svg>
            </button>
          </div>
          <div className="header-actions">
            <button className="header-btn" id="refresh-btn" onClick={handleRefresh}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <polyline points="23 4 23 10 17 10" /><polyline points="1 20 1 14 7 14" />
                <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15" />
              </svg>
              Refresh
            </button>
          </div>
        </header>

        {renderContent()}
      </main>

      {/* History Modal */}
      {isHistoryOpen && (
        <div className="modal-overlay" onClick={() => setIsHistoryOpen(false)}>
          <div className="modal-content history-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Riwayat Percakapan</h3>
              <button className="modal-close" onClick={() => setIsHistoryOpen(false)}>×</button>
            </div>
            <div className="modal-body history-list">
              {chat.conversations.length === 0 ? (
                <div className="empty-state">
                  <div className="empty-state-icon">💬</div>
                  <p>Belum ada riwayat percakapan</p>
                </div>
              ) : (
                chat.conversations.map((conv) => (
                  <div
                    key={conv.id}
                    className={`history-item ${chat.conversationId === conv.id ? 'active' : ''}`}
                    onClick={() => handleSelectConversation(conv.id)}
                  >
                    <div className="history-item-icon">💬</div>
                    <div className="history-item-info">
                      {editingConvId === conv.id ? (
                        <form onSubmit={handleRenameConfirm} onClick={(e) => e.stopPropagation()}>
                          <input
                            autoFocus
                            className="rename-input"
                            value={editingTitle}
                            onChange={(e) => setEditingTitle(e.target.value)}
                            onBlur={() => setEditingConvId(null)}
                          />
                        </form>
                      ) : (
                        <div className="history-item-title">{conv.title || 'Percakapan Tanpa Judul'}</div>
                      )}
                      <div className="history-item-time">
                        {new Date(conv.updated_at).toLocaleString('id-ID', {
                          day: 'numeric',
                          month: 'short',
                          year: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    </div>
                    <div className="history-item-actions">
                      <button 
                        className="menu-btn" 
                        onClick={(e) => {
                          e.stopPropagation();
                          setActiveMenuId(activeMenuId === conv.id ? null : conv.id);
                        }}
                      >
                        ⋮
                      </button>
                      {activeMenuId === conv.id && (
                        <div className="dropdown-menu">
                          <button onClick={(e) => handleStartRename(e, conv.id, conv.title)}>
                            Rename
                          </button>
                          <button 
                            className="delete-item" 
                            onClick={(e) => handleDeleteConversation(e, conv.id)}
                          >
                            Delete
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default App
