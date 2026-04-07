import React, { useState, useEffect, useRef } from 'react'
import wsService from '../services/websocket'

interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  mode?: 'build' | 'plan'
}

const ChatPanel: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const [agentMode, setAgentMode] = useState<'build' | 'plan'>('build')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const unsub = wsService.on('connection', (msg: any) => {
      setIsConnected(msg.payload.status === 'connected')
    })

    const unsubMsg = wsService.on('agent_response', (msg: any) => {
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: msg.payload.content,
        timestamp: new Date().toLocaleTimeString()
      }])
    })

    setIsConnected(wsService.isConnected)

    return () => {
      unsub()
      unsubMsg()
    }
  }, [])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault()
    if (!input.trim() || !isConnected) return

    let processedInput = input
    if (input.toLowerCase().startsWith('ultrawork ') || input.toLowerCase().startsWith('ulw ')) {
      const task = input.replace(/^(ultrawork|ulw)\s+/i, '')
      processedInput = `[ULTRAWORK MODE] ${task}`
    }

    if (agentMode === 'plan') {
      processedInput = `[READ-ONLY ANALYSIS MODE] ${processedInput}`
    }

    const newMessage: Message = {
      role: 'user',
      content: input,
      timestamp: new Date().toLocaleTimeString(),
      mode: agentMode
    }

    setMessages(prev => [...prev, newMessage])
    wsService.send('agent_message', { message: processedInput })
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
    if (e.key === 'Tab') {
      e.preventDefault()
      setAgentMode(m => m === 'build' ? 'plan' : 'build')
    }
  }

  return (
    <div className="chat-panel">
      <div className="chat-header">
        <div className="chat-title-group">
          <h2>Agent Chat</h2>
          <div className="mode-switcher">
            <button 
              className={`mode-pill ${agentMode === 'build' ? 'active build' : ''}`}
              onClick={() => setAgentMode('build')}
              title="Full access (Tab)"
            >
              BUILD
            </button>
            <button 
              className={`mode-pill ${agentMode === 'plan' ? 'active plan' : ''}`}
              onClick={() => setAgentMode('plan')}
              title="Read-only (Tab)"
            >
              PLAN
            </button>
          </div>
        </div>
        <div className={`status-indicator ${isConnected ? 'online' : 'offline'}`}>
          {isConnected ? 'Online' : 'Offline'}
        </div>
      </div>

      <div className="messages-container">
        {messages.length === 0 && (
          <div className="welcome-screen">
            <h2 style={{ fontSize: '28px', color: 'var(--accent-primary)', marginBottom: '8px' }}>Dardcor Agent</h2>
            <p>Ready to help in <strong>{agentMode.toUpperCase()}</strong> mode.</p>
            <div className="mode-explain">
              {agentMode === 'build' ? 
                '🛠️ BUILD: I can execute commands and modify files.' : 
                '📄 PLAN: I only analyze and give suggestions.'}
            </div>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`message ${msg.role} ${msg.mode || ''}`}>
            <div className="message-header">
              <span className="role-tag">
                {msg.role === 'user' ? 'YOU' : 'DARDCOR'}
                {msg.mode && <span className={`mode-tag ${msg.mode}`}>{msg.mode}</span>}
              </span>
              <span className="timestamp">{msg.timestamp}</span>
            </div>
            <div className="message-content">
              {msg.content.split('\n').map((line, j) => (
                <p key={j}>{line}</p>
              ))}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <form className={`input-container ${agentMode}`} onSubmit={handleSubmit}>
        <textarea
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={agentMode === 'build' ? "Ask anything (BUILD mode)..." : "Ask for analysis (PLAN mode)..."}
          rows={1}
        />
        <button type="submit" disabled={!input.trim() || !isConnected}>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2.5">
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
        </button>
      </form>
      <div className="input-hint">
        <span>Tab to switch mode</span>
        {input.toLowerCase().startsWith('ulw') && <span className="ultrawork-tag">⚡ ULTRAWORK ACTIVE</span>}
      </div>
    </div>
  )
}

export default ChatPanel
