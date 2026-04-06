import React, { useState, useRef, useEffect } from 'react'
import type { Message } from '../types'

interface ChatPanelProps {
  messages: Message[]
  isTyping: boolean
  onSendMessage: (text: string) => void
}

const ChatPanel: React.FC<ChatPanelProps> = ({ messages, isTyping, onSendMessage }) => {
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isTyping])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (input.trim()) {
      onSendMessage(input.trim())
      setInput('')
      if (inputRef.current) {
        inputRef.current.style.height = 'auto'
      }
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    e.target.style.height = 'auto'
    e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px'
  }

  const handleQuickAction = (action: string) => {
    onSendMessage(action)
  }

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  }

  const renderContent = (content: string) => {
    return content.split('\n').map((line, i) => {
      let processed = line.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      processed = processed.replace(/`([^`]+)`/g, '<code>$1</code>')
      if (processed.startsWith('• ') || processed.startsWith('- ')) {
        processed = '  ' + processed
      }
      return <span key={i} dangerouslySetInnerHTML={{ __html: processed + '\n' }} />
    })
  }

  if (messages.length === 0) {
    return (
      <div className="chat-container">
        <div className="chat-messages">
          <div className="chat-welcome">
            <h2>Dardcor Agent</h2>
            <p>
              AI Agent yang powerful untuk mengakses dan mengontrol seluruh komputer Anda.
              Melebihi semua AI Agent yang ada. Ketik perintah atau pilih aksi cepat di bawah.
            </p>
            <div className="quick-actions">
              <button className="quick-action-btn" onClick={() => handleQuickAction('sysinfo')} id="qa-sysinfo">
                <span className="icon">📊</span>
                <div className="title">System Info</div>
                <div className="desc">Lihat info sistem lengkap</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('processes')} id="qa-processes">
                <span className="icon">⚙️</span>
                <div className="title">Processes</div>
                <div className="desc">Daftar proses berjalan</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('drives')} id="qa-drives">
                <span className="icon">💽</span>
                <div className="title">Drives</div>
                <div className="desc">Lihat drive tersedia</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('list C:\\')} id="qa-files">
                <span className="icon">📂</span>
                <div className="title">Browse Files</div>
                <div className="desc">Jelajahi file di drive C</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('cpu')} id="qa-cpu">
                <span className="icon">🔧</span>
                <div className="title">CPU Info</div>
                <div className="desc">Detail prosesor</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('memory')} id="qa-memory">
                <span className="icon">🧠</span>
                <div className="title">Memory</div>
                <div className="desc">Penggunaan RAM</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('whoami')} id="qa-whoami">
                <span className="icon">🪪</span>
                <div className="title">Who Am I</div>
                <div className="desc">Info tentang agent</div>
              </button>
              <button className="quick-action-btn" onClick={() => handleQuickAction('help')} id="qa-help">
                <span className="icon">❓</span>
                <div className="title">Help</div>
                <div className="desc">Panduan perintah lengkap</div>
              </button>
            </div>
          </div>
        </div>
        <div className="chat-input-container">
          <form onSubmit={handleSubmit}>
            <div className="chat-input-wrapper">
              <textarea
                ref={inputRef}
                className="chat-input"
                value={input}
                onChange={handleInputChange}
                onKeyDown={handleKeyDown}
                placeholder="Ketik perintah... (contoh: sysinfo, list C:\, run dir, help)"
                rows={1}
                id="chat-input"
              />
              <button type="submit" className="chat-send-btn" disabled={!input.trim()} id="chat-send-btn">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="22" y1="2" x2="11" y2="13" />
                  <polygon points="22 2 15 22 11 13 2 9 22 2" />
                </svg>
              </button>
            </div>
          </form>
          <div className="chat-input-hint">Tekan Enter untuk kirim, Shift+Enter untuk baris baru</div>
        </div>
      </div>
    )
  }

  return (
    <div className="chat-container">
      <div className="chat-messages">
        {messages.map((msg) => (
          <div key={msg.id} className={`message ${msg.role}`}>
            <div className="message-avatar">
              {msg.role === 'user' ? '👤' : <div className="avatar-img" aria-label="Agent" />}
            </div>
            <div className="message-body">
              <div className="message-content">
                {renderContent(msg.content)}
              </div>
              {msg.actions && msg.actions.length > 0 && (
                <div className="message-actions">
                  {msg.actions.map((action, i) => (
                    <span key={i} className={`action-badge ${action.status}`}>
                      {action.status === 'completed' ? '✅' : action.status === 'error' ? '❌' : '⏳'}
                      {' '}{action.type}
                      {action.duration_ms ? ` (${action.duration_ms}ms)` : ''}
                    </span>
                  ))}
                </div>
              )}
              <div className="message-time">{formatTime(msg.timestamp)}</div>
            </div>
          </div>
        ))}

        {isTyping && (
          <div className="typing-indicator">
            <div className="message-avatar" style={{ background: 'transparent' }}>
              <div className="avatar-img" aria-label="Typing" />
            </div>
            <div className="typing-dots">
              <span /><span /><span />
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-container">
        <form onSubmit={handleSubmit}>
          <div className="chat-input-wrapper">
            <textarea
              ref={inputRef}
              className="chat-input"
              value={input}
              onChange={handleInputChange}
              onKeyDown={handleKeyDown}
              placeholder="Ketik perintah..."
              rows={1}
              id="chat-input-active"
            />
            <button
              type="submit"
              className="chat-send-btn"
              disabled={!input.trim() || isTyping}
              id="chat-send-active-btn"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="22" y1="2" x2="11" y2="13" />
                <polygon points="22 2 15 22 11 13 2 9 22 2" />
              </svg>
            </button>
          </div>
        </form>
        <div className="chat-input-hint">Tekan Enter untuk kirim, Shift+Enter untuk baris baru</div>
      </div>
    </div>
  )
}

export default ChatPanel
