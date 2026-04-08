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
    <div className="chat-container">
      <div className="chat-messages">
        {messages.length === 0 && (
          <div className="chat-welcome">
            <h2>Dardcor Agent</h2>
            <p>Ready to help in <strong>{agentMode.toUpperCase()}</strong> mode.</p>
            <div className="mode-explain">
              {agentMode === 'build' ? 
                '🛠️ BUILD: I can execute commands and modify files.' : 
                '📄 PLAN: I only analyze and give suggestions.'}
            </div>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`message ${msg.role}`}>
            <div className="message-avatar">
              {msg.role === 'assistant' ? <div className="avatar-img"></div> : 'U'}
            </div>
            <div className="message-body">
              <div className="message-meta">
                <span className="message-sender">{msg.role === 'assistant' ? 'Dardcor Agent' : 'You'}</span>
                <span className="message-time">{msg.timestamp}</span>
              </div>
              <div className="message-content">
                {msg.content}
              </div>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-container">
        <form className="chat-input-wrapper" onSubmit={handleSubmit}>
          <textarea
            className="chat-input"
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={agentMode === 'build' ? "Ask anything (BUILD mode)..." : "Ask for analysis (PLAN mode)..."}
            rows={1}
          />
          <button className="chat-send-btn" type="submit" disabled={!input.trim() || !isConnected}>
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2.5">
              <line x1="22" y1="2" x2="11" y2="13" />
              <polygon points="22 2 15 22 11 13 2 9 22 2" />
            </svg>
          </button>
        </form>
        <div className="chat-input-hint">
          Tab to switch mode • {agentMode.toUpperCase()} MODE
          {input.toLowerCase().startsWith('ulw') && <span style={{color: 'var(--accent-primary)', marginLeft: '10px'}}>⚡ ULTRAWORK</span>}
        </div>
      </div>
    </div>
  )
}

export default ChatPanel

