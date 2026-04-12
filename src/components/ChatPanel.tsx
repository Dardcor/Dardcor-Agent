import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import wsService from '../services/websocket'

interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  mode?: 'build' | 'plan'
}

interface ConversationSummary {
  id: string
  title: string
  created_at: string
  updated_at: string
}

const ChatPanel: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const [isTyping, setIsTyping] = useState(false)
  const [agentMode, setAgentMode] = useState<'build' | 'plan'>('build')
  const [conversationId, setConversationId] = useState<string | undefined>(undefined)
  const [showHistory, setShowHistory] = useState(false)
  const [conversations, setConversations] = useState<ConversationSummary[]>([])
  const [loadingHistory, setLoadingHistory] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const startNewChat = useCallback(() => {
    setMessages([])
    setConversationId(undefined)
    localStorage.removeItem('last_conv_id')
    setShowHistory(false)
    navigate('/chat')
  }, [navigate])

  useEffect(() => {
    wsService.connect().catch(() => {})

    const unsubConn = wsService.on('connection', (msg: any) => {
      setIsConnected(msg.payload.status === 'connected')
    })

    const unsubResp = wsService.on('agent_response', (msg: any) => {
      setIsTyping(false)
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: msg.payload?.content || '',
        timestamp: new Date().toLocaleTimeString(),
      }])

      if (msg.payload?.conversation_id) {
        const newId = msg.payload.conversation_id
        setConversationId(newId)
        localStorage.setItem('last_conv_id', newId)

        if (window.location.pathname === '/chat') {
          navigate(`/chat/${newId}`, { replace: true })
        }
      }
    })

    const unsubTyping = wsService.on('typing', (msg: any) => {
      setIsTyping(msg.payload?.typing === true)
    })

    const unsubError = wsService.on('error', (msg: any) => {
      setIsTyping(false)
      const errText = msg.payload?.error || 'Unknown error'
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: `⚠️ ${errText}`,
        timestamp: new Date().toLocaleTimeString(),
      }])
    })

    const unsubConvList = wsService.on('conversations_list', (msg: any) => {
      setLoadingHistory(false)
      const list: ConversationSummary[] = Array.isArray(msg.payload) ? msg.payload : []
      setConversations(list.sort((a, b) =>
        new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
      ))
    })

    const unsubConvDetail = wsService.on('conversation_detail', (msg: any) => {
      const conv = msg.payload
      if (!conv) return
      setConversationId(conv.id)
      localStorage.setItem('last_conv_id', conv.id)
      setShowHistory(false)
      const loaded: Message[] = (conv.messages || []).map((m: any) => ({
        role: m.role,
        content: m.content,
        timestamp: m.timestamp ? new Date(m.timestamp).toLocaleTimeString() : '',
      }))
      setMessages(loaded)
    })

    const handleExternalToggle = () => setShowHistory(prev => !prev)
    const handleNewChat = () => startNewChat()

    document.addEventListener('toggle-history', handleExternalToggle)
    document.addEventListener('new-chat', handleNewChat)

    if (id && isConnected) {
      wsService.getConversation(id)
      localStorage.setItem('last_conv_id', id)
    }

    setIsConnected(wsService.isConnected)

    return () => {
      unsubConn(); unsubResp(); unsubTyping()
      unsubError(); unsubConvList(); unsubConvDetail()
      document.removeEventListener('toggle-history', handleExternalToggle)
      document.removeEventListener('new-chat', handleNewChat)
    }
  }, [startNewChat, id, isConnected, navigate])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isTyping])

  useEffect(() => {
    if (showHistory) {
      setLoadingHistory(true)
      wsService.getConversations()
    }
  }, [showHistory])

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault()
    if (!input.trim() || !isConnected || isTyping) return

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
      mode: agentMode,
    }
    setMessages(prev => [...prev, newMessage])
    wsService.send('agent_message', {
      message: processedInput,
      conversation_id: conversationId,
    })
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSubmit() }
    if (e.key === 'Tab') { e.preventDefault(); setAgentMode(m => m === 'build' ? 'plan' : 'build') }
  }

  const loadConversation = (id: string) => {
    setConversationId(id)
    localStorage.setItem('last_conv_id', id)
    navigate(`/chat/${id}`)
    wsService.getConversation(id)
  }

  const formatDate = (iso: string) => {
    try {
      const d = new Date(iso)
      const today = new Date()
      if (d.toDateString() === today.toDateString()) return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      return d.toLocaleDateString([], { day: '2-digit', month: 'short' })
    } catch { return '' }
  }

  const renderContent = (content: string) => {
    const thinkingMatch = content.match(/^> \[Thinking\]\n([\s\S]*?)\n\n/);
    let thinkingPrompt = '';
    let mainContent = content;

    if (thinkingMatch) {
      thinkingPrompt = thinkingMatch[1];
      mainContent = content.replace(thinkingMatch[0], '');
    }

    const parts = mainContent.split(/(```[\s\S]*?```)/g);
    
    return (
      <>
        {thinkingPrompt && (
          <div style={{
            background: 'rgba(124,58,237,0.05)',
            borderLeft: '3px solid #7c3aed',
            padding: '10px 14px',
            marginBottom: '15px',
            borderRadius: '0 8px 8px 0',
            fontSize: '13px',
            color: 'rgba(255,255,255,0.7)',
            fontStyle: 'italic',
            lineHeight: '1.5',
            fontFamily: 'system-ui'
          }}>
            <div style={{ fontSize: '10px', fontWeight: 800, color: '#a78bfa', textTransform: 'uppercase', marginBottom: '4px', letterSpacing: '0.5px' }}>Strategic Thought</div>
            {thinkingPrompt}
          </div>
        )}
        {parts.map((part, i) => {
          if (part.startsWith('```')) {
            const code = part.replace(/^```[^\n]*\n?/, '').replace(/```$/, '')
            return (
              <pre key={i} style={{
                background: 'rgba(0,0,0,0.5)', padding: '15px', borderRadius: '10px',
                overflowX: 'auto', fontSize: '12px', margin: '12px 0',
                border: '1px solid rgba(124,58,237,0.15)', fontFamily: 'Fira Code, monospace',
                boxShadow: 'inset 0 0 20px rgba(0,0,0,0.2)'
              }}><code>{code}</code></pre>
            )
          }
          return <span key={i} style={{ whiteSpace: 'pre-wrap' }}>{part}</span>
        })}
      </>
    )
  }

  return (
    <div className="chat-container">
      {showHistory && (
        <div style={{
          position: 'fixed', inset: 0, zIndex: 9999,
          background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center'
        }}>
          <div style={{
            width: '600px', height: '70vh', background: 'var(--bg-secondary)',
            border: '1px solid var(--border-subtle)', borderRadius: '12px',
            display: 'flex', flexDirection: 'column',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)'
          }}>
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border-subtle)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontSize: '16px', fontWeight: 700, color: '#a78bfa' }}>🕒 Riwayat</span>
              <div style={{ display: 'flex', gap: '10px' }}>
                <button onClick={startNewChat} style={{
                  background: 'var(--accent-primary)', border: 'none', color: '#fff',
                  borderRadius: '6px', padding: '6px 12px', fontSize: '13px', cursor: 'pointer',
                  fontWeight: 600
                }}>+ Baru</button>
                <button onClick={() => setShowHistory(false)} style={{
                  background: 'none', border: 'none', color: '#94a3b8', cursor: 'pointer', fontSize: '24px', lineHeight: 1
                }}>×</button>
              </div>
            </div>

            <div style={{ flex: 1, overflowY: 'auto', padding: '16px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {loadingHistory ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '40px' }}>
                  <div className="typing-dots"><span /><span /><span /></div>
                </div>
              ) : conversations.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '40px', color: '#64748b', fontSize: '14px' }}>
                  Belum ada riwayat percakapan.
                </div>
              ) : (
                conversations.map(conv => (
                  <div
                    key={conv.id}
                    onClick={() => loadConversation(conv.id)}
                    style={{
                      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                      background: conversationId === conv.id ? 'rgba(124,58,237,0.15)' : 'var(--bg-tertiary)',
                      border: conversationId === conv.id ? '1px solid var(--accent-primary)' : '1px solid var(--border-subtle)',
                      borderRadius: '8px', padding: '12px 16px', cursor: 'pointer',
                      transition: 'all 0.2s',
                    }}
                  >
                    <div style={{ flex: 1, overflow: 'hidden' }}>
                      <div style={{ fontSize: '14px', color: conversationId === conv.id ? '#c4b5fd' : '#f8fafc', fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {conv.title || 'Percakapan'}
                      </div>
                      <div style={{ fontSize: '12px', color: '#64748b', marginTop: '4px' }}>
                        {formatDate(conv.updated_at)}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}

      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0, width: '100%' }}>
        <div className="chat-messages">
          {messages.length === 0 && (
            <div className="chat-welcome">
              <h2>Dardcor Agent</h2>
              <p>Ready in <strong>{agentMode.toUpperCase()}</strong> mode.</p>
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
                {msg.role === 'assistant' ? <div className="avatar-img" /> : 'U'}
              </div>
              <div className="message-body">
                <div className="message-meta">
                  <span className="message-sender">{msg.role === 'assistant' ? 'Dardcor Agent' : 'You'}</span>
                  <span className="message-time">{msg.timestamp}</span>
                </div>
                <div className="message-content">
                  {renderContent(msg.content)}
                </div>
              </div>
            </div>
          ))}

          {isTyping && (
            <div className="message assistant">
              <div className="message-avatar"><div className="avatar-img" /></div>
              <div className="message-body">
                <div className="message-meta">
                  <span className="message-sender">Dardcor Agent</span>
                </div>
                <div className="message-content">
                  <div className="typing-dots"><span /><span /><span /></div>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        <div className="chat-input-container">
          <form className="chat-input-wrapper" onSubmit={handleSubmit}>
            <textarea
              className="chat-input"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={agentMode === 'build' ? 'Ask anything (BUILD mode)...' : 'Ask for analysis (PLAN mode)...'}
              rows={1}
              disabled={isTyping}
            />
            <button className="chat-send-btn" type="submit" disabled={!input.trim() || !isConnected || isTyping}>
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2.5">
                <line x1="22" y1="2" x2="11" y2="13" />
                <polygon points="22 2 15 22 11 13 2 9 22 2" />
              </svg>
            </button>
          </form>
          <div className="chat-input-hint">
            Tab to switch mode • {agentMode.toUpperCase()} MODE
            {!isConnected && <span style={{ color: '#ef4444', marginLeft: '10px' }}>⚠ Disconnected</span>}
            {input.toLowerCase().startsWith('ulw') && <span style={{ color: 'var(--accent-primary)', marginLeft: '10px' }}>⚡ ULTRAWORK</span>}
          </div>
        </div>
      </div>

      <style dangerouslySetInnerHTML={{ __html: `
        .chat-container { height: 100%; display: flex; flex-direction: column; overflow: hidden; position: relative; }
        .chat-messages { 
          flex: 1; overflow-y: scroll !important; padding: 20px 40px 20px 20px; 
          display: flex; flex-direction: column; gap: 20px;
          scrollbar-gutter: stable;
        }
        .chat-messages::-webkit-scrollbar { width: 8px !important; }
        .chat-messages::-webkit-scrollbar-track { background: transparent !important; }
        .chat-messages::-webkit-scrollbar-thumb { 
          background: var(--accent-primary) !important; 
          border-radius: 10px !important;
          border: 2px solid var(--bg-primary) !important;
        }
        .chat-input-container { padding: 20px; max-width: 1000px; margin: 0 auto; width: 100%; }
        .chat-input-wrapper { 
          display: flex; align-items: center; background: var(--bg-tertiary); 
          border: 1.5px solid var(--border-subtle); border-radius: 20px; padding: 8px 12px;
          gap: 10px;
        }
        .chat-input { flex: 1; background: transparent; border: none; color: white; resize: none; font-size: 14px; }
        .chat-send-btn { 
          background: var(--accent-primary); color: white; border: none; 
          width: 36px; height: 36px; border-radius: 50%; cursor: pointer;
          display: flex; align-items: center; justify-content: center;
          transition: var(--transition-fast);
        }
        .chat-send-btn:hover:not(:disabled) { transform: scale(1.05); box-shadow: 0 0 15px var(--accent-glow-sm); }
        .chat-send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
      `}} />
    </div>
  )
}

export default ChatPanel
