import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import wsService from '../services/websocket'
import ThinkingConsole from './ThinkingConsole'

interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  mode?: 'build' | 'plan'
  isNew?: boolean
}

interface ConversationSummary {
  id: string
  title: string
  created_at: string
  updated_at: string
}

const TypewriterText: React.FC<{ text: string; onType?: () => void }> = ({ text, onType }) => {
  const [displayed, setDisplayed] = useState('')
  useEffect(() => {
    let index = 0
    const timer = setInterval(() => {
      setDisplayed(text.substring(0, index))
      index += 2
      if (onType) onType()
      if (index > text.length + 1) clearInterval(timer)
    }, 15)
    return () => clearInterval(timer)
  }, [text, onType])
  return <span style={{ whiteSpace: 'pre-wrap' }}>{displayed}</span>
}

const ChatPanel: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const [isTyping, setIsTyping] = useState(() => wsService.isAgentTyping)
  const [agentMode, setAgentMode] = useState<'build' | 'plan'>('build')
  const [conversationId, setConversationId] = useState<string | undefined>(undefined)
  const [showHistory, setShowHistory] = useState(false)
  const [conversations, setConversations] = useState<ConversationSummary[]>([])
  const [loadingHistory, setLoadingHistory] = useState(false)
  const [isCreatingChat, setIsCreatingChat] = useState(false)
  const [activeMenuId, setActiveMenuId] = useState<string | null>(null)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [isGeneratingImage, setIsGeneratingImage] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const lastCreateTime = useRef<number>(0)

  const startNewChat = useCallback((title?: string) => {
    if (id && messages.length === 0) {
      return
    }

    const lastConv = conversations[0]
    if (lastConv && (!lastConv.updated_at || new Date(lastConv.updated_at).getTime() === new Date(lastConv.created_at).getTime())) {
    }

    const now = Date.now()
    if (isCreatingChat || (now - lastCreateTime.current < 1500)) return
    lastCreateTime.current = now

    setIsCreatingChat(true)
    wsService.createConversation(title || 'Percakapan Baru')
  }, [id, messages.length, isCreatingChat, conversations])

  useEffect(() => {
    wsService.connect().catch(() => { })

    const unsubConn = wsService.on('connection', (msg: any) => {
      setIsConnected(msg.payload.status === 'connected')
    })

    const unsubResp = wsService.on('agent_response', (msg: any) => {
      setIsTyping(false)
      setIsGeneratingImage(false)
      setMessages(prev => {
        const lastMsg = prev[prev.length - 1]
        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.isNew) {
          return [...prev.slice(0, -1), {
            ...lastMsg,
            content: msg.payload?.content || lastMsg.content,
            isNew: false
          }]
        }
        return [...prev, {
          role: 'assistant',
          content: msg.payload?.content || '',
          timestamp: new Date().toLocaleTimeString(),
          isNew: false
        }]
      })

      if (msg.payload?.conversation_id) {
        const newId = msg.payload.conversation_id
        setConversationId(newId)
        localStorage.setItem('last_conv_id', newId)

        if (window.location.pathname === '/chat') {
          navigate(`/chat/${newId}`, { replace: true })
        }
      }
    })

    const unsubTurn = wsService.onAgentTurn((payload) => {
      if (payload.status === 'stream_chunk') {
        const chunk = payload.content || ''
        if (chunk.includes('[COMPLETE]')) {
          setIsTyping(false)
        }
        setMessages(prev => {
          const newMsgs = [...prev]
          const lastMsg = newMsgs[newMsgs.length - 1]
          if (lastMsg && lastMsg.role === 'assistant' && lastMsg.isNew) {
            lastMsg.content += chunk
            return newMsgs
          } else {
            return [...newMsgs, {
              role: 'assistant',
              content: chunk,
              timestamp: new Date().toLocaleTimeString(),
              isNew: true
            }]
          }
        })
      } else if (payload.status === 'processing' && payload.content) {
        if (payload.content.includes('[COMPLETE]')) {
          setIsTyping(false)
        }
        setMessages(prev => {
          const newMsgs = [...prev]
          const lastMsg = newMsgs[newMsgs.length - 1]
          if (lastMsg && lastMsg.role === 'assistant' && lastMsg.isNew) {
            lastMsg.content = payload.content
            return newMsgs
          }
          return prev
        })
      }
    })

    const unsubTyping = wsService.on('typing', (msg: any) => {
      setIsTyping(msg.payload?.typing === true)
    })

    const unsubError = wsService.on('error', (msg: any) => {
      setIsTyping(false)
      setIsGeneratingImage(false)
      setIsCreatingChat(false)
      const errText = msg.payload?.error || 'Unknown error'

      if (errText.includes('Session not found')) {
        localStorage.removeItem('last_conv_id')
        setConversationId(undefined)
        navigate('/chat', { replace: true })
        return
      }

      setMessages(prev => [...prev, {
        role: 'assistant',
        content: `⚠️ ${errText}`,
        timestamp: new Date().toLocaleTimeString(),
        isNew: true
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
      const loaded: Message[] = (conv.messages || []).map((m: any) => ({
        role: m.role,
        content: m.content,
        timestamp: m.timestamp ? new Date(m.timestamp).toLocaleTimeString() : '',
        isNew: false
      }))
      setMessages(loaded)
    })

    const unsubConvCreated = wsService.on('conversation_created', (msg: any) => {
      const conv = msg.payload
      if (!conv) return
      setIsCreatingChat(false)
      setMessages([])
      setConversationId(conv.id)
      localStorage.setItem('last_conv_id', conv.id)
      setShowHistory(false)
      navigate(`/chat/${conv.id}`)
    })

    const handleExternalToggle = () => setShowHistory(prev => !prev)
    const handleNewChatEvent = (e: Event) => {
      const customEvent = e as CustomEvent;
      startNewChat(customEvent.detail?.title);
    };

    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setActiveMenuId(null)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('toggle-history', handleExternalToggle)
    document.addEventListener('new-chat', handleNewChatEvent)

    if (id && isConnected) {
      wsService.getConversation(id)
      localStorage.setItem('last_conv_id', id)
    }

    setIsConnected(wsService.isConnected)

    return () => {
      unsubConn(); unsubResp(); unsubTyping(); unsubTurn();
      unsubError(); unsubConvList(); unsubConvDetail(); unsubConvCreated()
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('toggle-history', handleExternalToggle)
      document.removeEventListener('new-chat', handleNewChatEvent)
    }
  }, [id, isConnected, navigate, startNewChat])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isTyping])

  const scrollToBottomFast = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'auto' })
  }, [])

  useEffect(() => {
    if (showHistory) {
      setLoadingHistory(true)
      wsService.getConversations()
    }
  }, [showHistory])

  useEffect(() => {
    if (!id && isConnected && !isCreatingChat) {
      const lastId = localStorage.getItem('last_conv_id')
      if (lastId) {
        navigate(`/chat/${lastId}`, { replace: true })
      } else {
        startNewChat()
      }
    }
  }, [id, isConnected, isCreatingChat, startNewChat, navigate])

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault()
    if (!input.trim() || !isConnected || isTyping) return

    let processedInput = input
    const isImagePrompt = /\b(buatkan|bikin|generate|create|render)\b.*\b(gambar|image|photo|foto|lukisan|ilustrasi|visual)\b/i.test(input)

    if (isImagePrompt) {
      setIsGeneratingImage(true)
      processedInput = `[IMAGE_GEN_MODE] ${input}`
    } else if (input.toLowerCase().startsWith('ultrawork ') || input.toLowerCase().startsWith('ulw ')) {
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
      isNew: true
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

  const handleStop = () => {
    wsService.stopAgent()
    setIsTyping(false)
    setIsGeneratingImage(false)
  }

  const loadConversation = (id: string) => {
    setConversationId(id)
    localStorage.setItem('last_conv_id', id)
    navigate(`/chat/${id}`)
    wsService.getConversation(id)
    setShowHistory(false)
  }

  const handleDelete = (e: React.MouseEvent, convId: string) => {
    e.stopPropagation()
    if (window.confirm('Hapus percakapan ini?')) {
      wsService.deleteConversation(convId)
      setActiveMenuId(null)
      if (conversationId === convId) {
        setMessages([])
        setConversationId(undefined)
        navigate('/chat')
      }
    }
  }

  const startRename = (e: React.MouseEvent, conv: ConversationSummary) => {
    e.stopPropagation()
    setRenamingId(conv.id)
    setNewName(conv.title)
    setActiveMenuId(null)
  }

  const handleRenameSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (renamingId && newName.trim()) {
      wsService.renameConversation(renamingId, newName.trim())
      setRenamingId(null)
      setNewName('')
    }
  }

  const formatDate = (iso: string) => {
    try {
      const d = new Date(iso)
      const today = new Date()
      if (d.toDateString() === today.toDateString()) return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      return d.toLocaleDateString([], { day: '2-digit', month: 'short' })
    } catch { return '' }
  }

  const renderContent = (content: string, isLastMsg: boolean, role: string, isNew?: boolean) => {
    const screenshotMatch = content.match(/Screenshot saved to (screenshot_\d+\.png)/i)
    const screenshotUrl = screenshotMatch ? `http://127.0.0.1:25000/storage/screenshots/${screenshotMatch[1]}` : null

    const parts = content.split(/(```[\s\S]*?```)/g)

    return (
      <>
        <ThinkingConsole content={content} />
        {screenshotUrl && (
          <div style={{ margin: '15px 0' }}>
            <img
              src={screenshotUrl}
              alt="Browser Screenshot"
              style={{
                maxWidth: '100%',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)',
                boxShadow: '0 10px 25px rgba(0,0,0,0.3)',
                cursor: 'zoom-in'
              }}
              onClick={() => window.open(screenshotUrl, '_blank')}
            />
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

          const cleanText = part.replace(/\[THOUGHT\][\s\S]*?(\[PLAN\]|\[ACTION\]|$)/i, '$1').trim()
          if (!cleanText) return null

          const segments = cleanText.split(/(\[ACTION\][\s\S]*?(?:\[\/ACTION\]|$)|\[PLAN\][\s\S]*?(?:\[\/PLAN\]|$))/g)

          return (
            <React.Fragment key={i}>
              {segments.map((segment, idx) => {
                if (segment.startsWith('[ACTION]')) {
                  const hasEndTag = segment.endsWith('[/ACTION]')
                  const inner = hasEndTag ? segment.substring(8, segment.length - 9).trim() : segment.substring(8).trim()

                  let cmdName = '...'
                  let args = inner

                  try {
                    // Try to parse as JSON if it looks like JSON
                    if (inner.startsWith('{')) {
                      const json = JSON.parse(hasEndTag ? inner : inner + '}')
                      cmdName = json.tool || json.name || 'tool'
                      args = JSON.stringify(json, null, 2)
                    } else {
                      const firstSpace = inner.indexOf(' ')
                      cmdName = firstSpace > -1 ? inner.substring(0, firstSpace) : inner
                      args = firstSpace > -1 ? inner.substring(firstSpace + 1) : '{ }'
                    }
                  } catch (e) {
                    // Fallback to text parsing
                    const firstSpace = inner.indexOf(' ')
                    cmdName = firstSpace > -1 ? inner.substring(0, firstSpace) : inner
                    args = firstSpace > -1 ? inner.substring(firstSpace + 1) : inner
                  }

                  return (
                    <div key={idx} className="tool-execution-block">
                      <div className="tool-header">
                        <div className="tool-icon">🛠️</div>
                        <span className="tool-title">Executing Tool</span>
                        <span className="tool-badge">{cmdName}</span>
                      </div>
                      <pre className="tool-args">{args}</pre>
                    </div>
                  )
                }
                if (segment.startsWith('[PLAN]')) {
                  const hasEndTag = segment.endsWith('[/PLAN]')
                  const inner = hasEndTag ? segment.substring(6, segment.length - 7).trim() : segment.substring(6).trim()
                  return (
                    <div key={idx} className="plan-execution-block">
                      <div className="plan-header">
                        <div className="tool-icon">📄</div>
                        <span className="tool-title">Agent Plan</span>
                      </div>
                      <div className="plan-content">{inner}</div>
                    </div>
                  )
                }
                if (!segment.trim()) return null

                if (isLastMsg && role === 'assistant' && isNew) {
                  return <TypewriterText key={idx} text={segment} onType={scrollToBottomFast} />
                }
                return <span key={idx} style={{ whiteSpace: 'pre-wrap' }}>{segment}</span>
              })}
            </React.Fragment>
          )
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
                <button
                  onClick={() => startNewChat()}
                  disabled={isCreatingChat}
                  style={{
                    background: isCreatingChat ? '#4b5563' : 'var(--accent-primary)',
                    border: 'none', color: '#fff',
                    borderRadius: '6px', padding: '6px 12px', fontSize: '13px', cursor: isCreatingChat ? 'not-allowed' : 'pointer',
                    fontWeight: 600
                  }}>
                  {isCreatingChat ? 'Membuat...' : '+ Baru'}
                </button>
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
                      transition: 'all 0.2s', position: 'relative'
                    }}
                  >
                    <div style={{ flex: 1, overflow: 'hidden' }}>
                      {renamingId === conv.id ? (
                        <form onSubmit={handleRenameSubmit} onClick={e => e.stopPropagation()} style={{ display: 'flex', gap: '8px' }}>
                          <input
                            autoFocus
                            value={newName}
                            onChange={e => setNewName(e.target.value)}
                            onBlur={() => setRenamingId(null)}
                            style={{
                              flex: 1, background: 'rgba(0,0,0,0.3)', border: '1px solid var(--accent-primary)',
                              borderRadius: '4px', color: '#fff', padding: '2px 8px', fontSize: '13px'
                            }}
                          />
                        </form>
                      ) : (
                        <>
                          <div style={{ fontSize: '14px', color: conversationId === conv.id ? '#c4b5fd' : '#f8fafc', fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {conv.title || 'Percakapan'}
                          </div>
                          <div style={{ fontSize: '12px', color: '#64748b', marginTop: '4px' }}>
                            {formatDate(conv.updated_at)}
                          </div>
                        </>
                      )}
                    </div>

                    <div style={{ position: 'relative', marginLeft: '8px' }} ref={activeMenuId === conv.id ? menuRef : null}>
                      <button
                        onClick={(e) => { e.stopPropagation(); setActiveMenuId(prev => prev === conv.id ? null : conv.id) }}
                        style={{
                          background: 'none', border: 'none', color: '#64748b', cursor: 'pointer',
                          padding: '4px 8px', borderRadius: '4px', transition: 'background 0.2s',
                          fontSize: '18px', fontWeight: 'bold', display: 'flex', alignItems: 'center'
                        }}
                        onMouseEnter={e => e.currentTarget.style.background = 'rgba(255,255,255,0.05)'}
                        onMouseLeave={e => e.currentTarget.style.background = 'none'}
                      >
                        ⋮
                      </button>

                      {activeMenuId === conv.id && (
                        <div style={{
                          position: 'absolute', top: '100%', right: 0, zIndex: 100,
                          background: 'var(--bg-secondary)', border: '1px solid var(--border-subtle)',
                          borderRadius: '8px', boxShadow: '0 4px 12px rgba(0,0,0,0.5)',
                          minWidth: '120px', padding: '4px', marginTop: '4px'
                        }}>
                          <button
                            onClick={(e) => startRename(e, conv)}
                            style={{
                              width: '100%', textAlign: 'left', background: 'none', border: 'none',
                              color: '#f8fafc', padding: '8px 12px', borderRadius: '4px', cursor: 'pointer',
                              fontSize: '13px', display: 'flex', alignItems: 'center', gap: '8px'
                            }}
                            onMouseEnter={e => e.currentTarget.style.background = 'rgba(124,58,237,0.1)'}
                            onMouseLeave={e => e.currentTarget.style.background = 'none'}
                          >
                            ✏️ Rename
                          </button>
                          <button
                            onClick={(e) => handleDelete(e, conv.id)}
                            style={{
                              width: '100%', textAlign: 'left', background: 'none', border: 'none',
                              color: '#ef4444', padding: '8px 12px', borderRadius: '4px', cursor: 'pointer',
                              fontSize: '13px', display: 'flex', alignItems: 'center', gap: '8px'
                            }}
                            onMouseEnter={e => e.currentTarget.style.background = 'rgba(239,68,68,0.1)'}
                            onMouseLeave={e => e.currentTarget.style.background = 'none'}
                          >
                            🗑️ Delete
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

          {messages.map((msg, i) => {
            const isLastMessage = i === messages.length - 1
            return (
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
                    {renderContent(msg.content, isLastMessage, msg.role, msg.isNew)}
                  </div>
                </div>
              </div>
            )
          })}

          {isGeneratingImage && (
            <div className="message assistant">
              <div className="message-avatar" style={{ background: 'transparent', display: 'flex', justifyContent: 'center', alignItems: 'center', boxShadow: 'none' }}>
                <div style={{ width: '28px', height: '28px', borderRadius: '50%', border: '3px solid rgba(59, 130, 246, 0.2)', borderTopColor: '#3b82f6', animation: 'spin 1s linear infinite' }} />
              </div>
              <div className="message-body">
                <div className="message-content" style={{ minWidth: '400px', padding: '0', background: 'transparent', border: 'none', boxShadow: 'none' }}>
                  <div style={{
                    width: '100%', height: '300px', background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.05) 0%, rgba(37, 99, 235, 0.02) 100%)',
                    borderRadius: '20px', border: '2px dashed rgba(59, 130, 246, 0.25)',
                    display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
                    gap: '20px', position: 'relative', overflow: 'hidden',
                    boxShadow: 'inset 0 0 30px rgba(59, 130, 246, 0.05)'
                  }}>
                    <div className="image-loading-ripple"></div>
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px', zIndex: 1 }}>
                      <span style={{
                        color: '#60a5fa', fontSize: '15px', fontWeight: 700,
                        letterSpacing: '1px', textTransform: 'uppercase',
                        textShadow: '0 0 10px rgba(96, 165, 250, 0.5)'
                      }}>Generating Visual Intelligence</span>
                      <span style={{ color: 'rgba(255,255,255,0.4)', fontSize: '12px' }}>Dardcor Image Engine v2.4 (Active)</span>
                      <button
                        onClick={handleStop}
                        style={{
                          marginTop: '15px',
                          background: 'rgba(239, 68, 68, 0.2)', border: '1px solid rgba(239, 68, 68, 0.4)',
                          color: '#ef4444', borderRadius: '8px', padding: '6px 16px', fontSize: '12px',
                          fontWeight: 700, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '8px',
                          transition: 'all 0.2s', zIndex: 10
                        }}
                        onMouseEnter={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.35)'}
                        onMouseLeave={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.2)'}
                      >
                        <div style={{ width: '10px', height: '10px', background: '#ef4444', borderRadius: '2px' }} />
                        CANCEL GENERATION
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {isTyping && !isGeneratingImage && (
            <div className="message assistant">
              <div className="message-avatar" style={{ background: 'transparent', display: 'flex', justifyContent: 'center', alignItems: 'center', boxShadow: 'none' }}>
                <div style={{ width: '28px', height: '28px', borderRadius: '50%', border: '3px solid rgba(168, 85, 247, 0.2)', borderTopColor: '#a855f7', animation: 'spin 1s linear infinite' }} />
              </div>
              <div className="message-body" style={{ justifyContent: 'center' }}>
                <div className="message-content" style={{ display: 'flex', alignItems: 'center', gap: '15px', background: 'rgba(124,58,237,0.05)', border: '1px solid rgba(124,58,237,0.2)', padding: '8px 15px', borderRadius: '12px' }}>
                  <span style={{ color: '#a855f7', fontSize: '13px', fontWeight: 600, letterSpacing: '0.5px' }}>Dardcor Agent Thinking...</span>
                  <button
                    onClick={handleStop}
                    style={{
                      background: 'rgba(239, 68, 68, 0.15)', border: '1px solid rgba(239, 68, 68, 0.3)',
                      color: '#ef4444', borderRadius: '6px', padding: '2px 8px', fontSize: '11px',
                      fontWeight: 700, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px',
                      transition: 'all 0.2s'
                    }}
                    onMouseEnter={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.25)'}
                    onMouseLeave={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.15)'}
                  >
                    <div style={{ width: '8px', height: '8px', background: '#ef4444', borderRadius: '1px' }} />
                    STOP
                  </button>
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
            {isTyping ? (
              <button className="chat-send-btn" type="button" onClick={handleStop} style={{ background: '#ef4444' }}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <rect x="6" y="6" width="12" height="12" />
                </svg>
              </button>
            ) : (
              <button className="chat-send-btn" type="submit" disabled={!input.trim() || !isConnected}>
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <line x1="22" y1="2" x2="11" y2="13" />
                  <polygon points="22 2 15 22 11 13 2 9 22 2" />
                </svg>
              </button>
            )}
          </form>
          <div className="chat-input-hint">
            Tab to switch mode • {agentMode.toUpperCase()} MODE
            {!isConnected && <span style={{ color: '#ef4444', marginLeft: '10px' }}>⚠ Disconnected</span>}
            {input.toLowerCase().startsWith('ulw') && <span style={{ color: 'var(--accent-primary)', marginLeft: '10px' }}>⚡ ULTRAWORK</span>}
          </div>
        </div>
      </div>

      <style dangerouslySetInnerHTML={{
        __html: `
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
        .chat-container { flex: 1; height: 100%; display: flex; flex-direction: column; overflow: hidden; position: relative; min-height: 0; }
        .chat-messages { 
          flex: 1; overflow-y: auto !important; padding: 20px 24px; 
          display: flex; flex-direction: column; gap: 16px;
          scrollbar-gutter: stable;
          min-height: 0;
          overscroll-behavior: contain;
        }
        .chat-messages::-webkit-scrollbar { width: 10px !important; }
        .chat-messages::-webkit-scrollbar-track { background: rgba(0,0,0,0.1) !important; }
        .chat-messages::-webkit-scrollbar-thumb { 
          background: var(--accent-primary) !important; 
          border-radius: 10px !important;
          border: 2px solid var(--bg-primary) !important;
        }
        .chat-input-container { 
          padding: 10px 20px 30px 20px; 
          max-width: 1000px; 
          margin: 0 auto; 
          width: 100%; 
          background: transparent;
          backdrop-filter: none;
          border-top: none;
          flex-shrink: 0;
          z-index: 10;
        }
        .chat-input-hint { text-align: center; margin-top: 8px; font-size: 11px; color: var(--text-muted); }
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
        .image-loading-ripple {
          width: 48px; height: 48px; border-radius: 50%;
          border: 3px solid rgba(96, 165, 250, 0.3);
          position: relative; animation: ripple 1.5s infinite;
        }
        .image-loading-ripple::after {
          content: ""; position: absolute; inset: -3px; border-radius: 50%;
          border: 3px solid #60a5fa; animation: ripple-expand 1.5s infinite;
          opacity: 0;
        }
        @keyframes ripple-expand {
          0% { transform: scale(0.8); opacity: 0.8; }
          100% { transform: scale(2.2); opacity: 0; }
        }


        .tool-execution-block {
           background: rgba(0, 0, 0, 0.4); border-radius: 10px; margin: 12px 0;
           border: 1px solid rgba(139, 92, 246, 0.3); padding: 12px;
           box-shadow: inset 0 0 15px rgba(139, 92, 246, 0.05);
        }
        .tool-header {
           display: flex; align-items: center; gap: 8px; margin-bottom: 8px;
        }
        .tool-icon {
           display: flex; align-items: center; justify-content: center;
           width: 24px; height: 24px; background: rgba(139, 92, 246, 0.2);
           border-radius: 6px; font-size: 14px;
        }
        .tool-title { font-weight: 600; color: #c4b5fd; font-size: 13px; }
        .tool-badge {
           background: rgba(139, 92, 246, 0.2); color: #ddd6fe;
           padding: 2px 8px; border-radius: 4px; font-size: 11px; font-family: monospace;
           border: 1px solid rgba(139, 92, 246, 0.4);
        }
        .tool-args {
           background: rgba(0, 0, 0, 0.6); padding: 10px; border-radius: 6px;
           color: #a78bfa; font-size: 12px; font-family: monospace;
           white-space: pre-wrap; word-break: break-all; margin: 0;
           border: 1px dashed rgba(139, 92, 246, 0.2);
        }

        .plan-execution-block {
           background: rgba(16, 185, 129, 0.05); border-radius: 10px; margin: 12px 0;
           border: 1px solid rgba(16, 185, 129, 0.2); padding: 12px;
        }
        .plan-header {
           display: flex; align-items: center; gap: 8px; margin-bottom: 8px;
        }
        .plan-header .tool-icon { background: rgba(16, 185, 129, 0.15); }
        .plan-header .tool-title { font-weight: 600; color: #6ee7b7; font-size: 13px; }
        .plan-content {
           color: #a7f3d0; font-size: 13px; line-height: 1.5;
           white-space: pre-wrap;
        }
      `}} />
    </div>
  )
}

export default ChatPanel
