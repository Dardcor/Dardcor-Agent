import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import wsService from '../services/websocket'
import ThinkingConsole from './ThinkingConsole'
import { getAllModels, SuggestionModel } from '../services/modelSuggestions'

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
  
  // Model Modal State
  const [showModelPopup, setShowModelPopup] = useState(false)
  const [modelSearch, setModelSearch] = useState('')
  const allAvailableModels = useMemo(() => getAllModels(), [])

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

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

    return () => {
      unsubConn(); unsubResp(); unsubTyping(); unsubError(); unsubConvList(); unsubConvDetail()
    }
  }, [navigate])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isTyping])

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value
    setInput(val)

    // Trigger Popup if they type /model
    if (val === '/model' || val === '/model ') {
      setShowModelPopup(true)
      setInput('') // clear input automatically
    }
  }

  const selectModel = (modelId: string) => {
    wsService.send('agent_request', {
        content: `/model ${modelId}`,
        mode: agentMode
    })
    setShowModelPopup(false)
    setModelSearch('')
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    } else if (e.key === 'Tab') {
      e.preventDefault()
      setAgentMode(prev => prev === 'build' ? 'plan' : 'build')
    }
  }

  const handleSubmit = () => {
    if (!input.trim() || !isConnected) return

    if (input.trim() === '/model') {
       setShowModelPopup(true)
       setInput('')
       return
    }

    const userMessage: Message = {
      role: 'user',
      content: input,
      timestamp: new Date().toLocaleTimeString(),
      mode: agentMode,
    }

    setMessages(prev => [...prev, userMessage])
    setIsTyping(true)

    wsService.send('agent_request', {
      content: input,
      mode: agentMode,
      conversation_id: conversationId,
    })

    setInput('')
  }

  const filteredModels = allAvailableModels.filter(m => 
    m.id.toLowerCase().includes(modelSearch.toLowerCase()) || 
    m.name.toLowerCase().includes(modelSearch.toLowerCase()) || 
    m.provider.toLowerCase().includes(modelSearch.toLowerCase())
  )

  const lastAssistantMessage = messages.slice().reverse().find(m => m.role === 'assistant')?.content || ''

  return (
    <div className="flex flex-col h-full bg-[#0a0a0a] text-white overflow-hidden relative">
      {}
      <div className="border-b border-white/5 bg-black/20 px-6 py-3 flex items-center justify-between backdrop-blur-md">
        <div className="flex items-center gap-3">
          <div className="w-2 h-2 rounded-full bg-purple-500 animate-pulse"></div>
          <h2 className="text-sm font-semibold tracking-wider text-purple-100/80">DARDCOR AGENT</h2>
          <div className="px-2 py-0.5 rounded border border-white/10 text-[10px] font-bold text-white/40">
            {agentMode.toUpperCase()} MODE
          </div>
        </div>
        <div className="flex items-center gap-4">
          <button 
            onClick={() => {
              setLoadingHistory(true)
              wsService.send('get_conversations', {})
              setShowHistory(true)
            }}
            className="text-[11px] font-bold text-white/40 hover:text-white/80 transition-colors uppercase tracking-widest"
          >
            History
          </button>
          <button 
            onClick={startNewChat}
            className="text-[11px] font-bold text-purple-400 hover:text-purple-300 transition-colors uppercase tracking-widest"
          >
            New Chat
          </button>
        </div>
      </div>

      {}
      <div className="flex-1 overflow-y-auto px-6 py-8 space-y-8 scroll-smooth custom-scrollbar">
        {messages.length === 0 && !isTyping && (
          <div className="h-full flex flex-col items-center justify-center text-center space-y-4 opacity-20">
            <div className="w-12 h-12 rounded-2xl border-2 border-white/20 flex items-center justify-center">
              <span className="text-2xl">⚡</span>
            </div>
            <p className="text-sm font-medium">Ready for input.</p>
          </div>
        )}
        
        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'} animate-fade-in`}>
            <div className={`max-w-[85%] group`}>
              <div className="flex items-center gap-2 mb-2 opacity-30 group-hover:opacity-100 transition-opacity">
                <span className="text-[10px] font-bold uppercase tracking-tighter">{msg.role}</span>
                <span className="text-[10px]">{msg.timestamp}</span>
              </div>
              <div className={`
                px-5 py-3 rounded-2xl text-[14px] leading-relaxed whitespace-pre-wrap
                ${msg.role === 'user' 
                  ? 'bg-purple-600/10 border border-purple-500/20 text-purple-50' 
                  : 'bg-white/5 border border-white/5 text-white/90'}
              `}>
                {msg.content}
              </div>
            </div>
          </div>
        ))}
        {isTyping && <ThinkingConsole content={lastAssistantMessage} isStreaming={true} />}
        <div ref={messagesEndRef} />
      </div>

      {}
      <div className="p-6 bg-black/40 border-t border-white/5 backdrop-blur-xl relative">
        <div className="relative group max-w-5xl mx-auto">
          <textarea
            ref={inputRef}
            rows={1}
            value={input}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder="Ask anything... (type /model for model popup)"
            className="
              w-full bg-white/5 border border-white/10 rounded-2xl px-6 py-4 
              text-sm focus:outline-none focus:border-purple-500/50 transition-all
              placeholder:text-white/20 resize-none overflow-hidden
            "
          />
          <div className="absolute right-4 bottom-4 flex items-center gap-3">
            <div className="text-[10px] font-bold text-white/10 group-focus-within:text-white/30 transition-colors uppercase tracking-widest">
              {agentMode} mode
            </div>
          </div>
        </div>
      </div>

      {}
      {showModelPopup && (
        <div className="absolute inset-0 z-[200] animate-fade-in flex items-center justify-center p-6 bg-black/80 backdrop-blur-md">
          <div className="w-full max-w-lg bg-[#0f0f0f] border border-white/20 rounded-3xl overflow-hidden shadow-2xl flex flex-col max-h-[80%] relative">
            
            <div className="p-6 border-b border-white/5 bg-white/5">
              <h3 className="text-sm font-bold uppercase tracking-widest text-white/80 mb-4">Select AI Model</h3>
              <input 
                autoFocus
                type="text"
                placeholder="Search models... (e.g., gpt, claude)"
                value={modelSearch}
                onChange={e => setModelSearch(e.target.value)}
                className="w-full bg-black/40 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-purple-500/50 transition-all placeholder:text-white/30"
              />
            </div>
            
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
              {filteredModels.length === 0 ? (
                <div className="h-40 flex items-center justify-center text-white/20 text-sm">No models found.</div>
              ) : (
                <div className="grid gap-2">
                  {filteredModels.map(m => (
                    <button
                      key={m.id}
                      onClick={() => selectModel(m.id)}
                      className="p-4 flex items-center justify-between rounded-xl bg-white/5 border border-transparent hover:border-purple-500/50 hover:bg-white/10 text-left transition-all group"
                    >
                      <div className="flex flex-col">
                        <span className="text-[14px] font-medium text-white/90 group-hover:text-purple-300">
                          {m.name}
                        </span>
                        <span className="text-[10px] text-white/30 font-mono tracking-tighter mt-1">{m.id}</span>
                      </div>
                      <span className="text-[9px] px-2 py-1 rounded border border-white/10 text-white/40 group-hover:border-purple-500/30 group-hover:text-purple-400 uppercase tracking-widest">
                        {m.provider}
                      </span>
                    </button>
                  ))}
                </div>
              )}
            </div>

            <button 
              onClick={() => setShowModelPopup(false)}
              className="absolute top-4 right-4 w-8 h-8 rounded-full hover:bg-white/10 flex items-center justify-center text-white/40 hover:text-white transition-colors"
            >
              ✕
            </button>
          </div>
        </div>
      )}

      {}
      {showHistory && (
        <div className="absolute inset-0 z-[100] animate-fade-in flex items-center justify-center p-6 bg-black/60 backdrop-blur-md">
          <div className="w-full max-w-2xl bg-[#0f0f0f] border border-white/10 rounded-3xl overflow-hidden shadow-2xl flex flex-col max-h-[80%]">
            <div className="p-6 border-b border-white/5 flex items-center justify-between bg-white/5">
              <h3 className="text-sm font-bold uppercase tracking-widest text-white/60">Session History</h3>
              <button 
                onClick={() => setShowHistory(false)}
                className="w-8 h-8 rounded-full hover:bg-white/10 flex items-center justify-center transition-colors"
              >
                ✕
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
              {loadingHistory ? (
                <div className="h-40 flex items-center justify-center">
                  <div className="typing-dots"><span></span><span></span><span></span></div>
                </div>
              ) : conversations.length === 0 ? (
                <div className="h-40 flex items-center justify-center text-white/20 text-sm">No history found.</div>
              ) : (
                <div className="grid gap-2">
                  {conversations.map(conv => (
                    <button
                      key={conv.id}
                      onClick={() => {
                        wsService.send('get_conversation', { id: conv.id })
                        navigate(`/chat/${conv.id}`)
                      }}
                      className="p-4 rounded-2xl bg-white/5 border border-white/0 hover:border-white/10 hover:bg-white/10 text-left transition-all group"
                    >
                      <div className="text-sm font-medium mb-1 group-hover:text-purple-400 transition-colors">{conv.title || 'Untitled Session'}</div>
                      <div className="text-[10px] text-white/30">{new Date(conv.updated_at).toLocaleString()}</div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ChatPanel
