import { useState, useEffect, useCallback, useRef } from 'react'
import { wsService } from '../services/websocket'
import { agentAPI, systemAPI, filesAPI, commandAPI } from '../services/api'
import type {
  Message,
  Conversation,
  SystemInfo,
  FileInfo,
  ProcessInfo,
  CommandResponse,
  WSMessage,
  AgentResponse,
} from '../types'

export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    const unsubscribe = wsService.on('connection', (msg: WSMessage) => {
      const payload = msg.payload as { status: string }
      setIsConnected(payload.status === 'connected')
    })

    wsService.connect().catch(() => {})

    return () => {
      unsubscribe()
    }
  }, [])

  return { isConnected, ws: wsService }
}

export function useChat() {
  const [messages, setMessages] = useState<Message[]>([])
  const [isTyping, setIsTyping] = useState(false)
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])

  useEffect(() => {
    const unsubTyping = wsService.on('typing', (msg: WSMessage) => {
      const payload = msg.payload as { typing: boolean }
      setIsTyping(payload.typing)
    })

    const unsubResponse = wsService.on('agent_response', (msg: WSMessage) => {
      const response = msg.payload as AgentResponse
      setConversationId(response.conversation_id)

      const newMessage: Message = {
        id: response.id,
        role: 'assistant',
        content: response.content,
        actions: response.actions,
        timestamp: response.timestamp,
      }

      setMessages((prev) => [...prev, newMessage])
      setIsTyping(false)
      wsService.getConversations()
    })

    const unsubConvList = wsService.on('conversations_list', (msg: WSMessage) => {
      const list = msg.payload as Conversation[] | null
      setConversations(list || [])
    })

    const unsubConvDetail = wsService.on('conversation_detail', (msg: WSMessage) => {
      const conv = msg.payload as Conversation
      setConversationId(conv.id)
      setMessages(conv.messages || [])
    })

    const unsubError = wsService.on('error', (msg: WSMessage) => {
      const payload = msg.payload as { error: string }
      const errorMessage: Message = {
        id: Date.now().toString(),
        role: 'assistant',
        content: `Error: ${payload.error}`,
        timestamp: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, errorMessage])
      setIsTyping(false)
    })

    return () => {
      unsubTyping()
      unsubResponse()
      unsubConvList()
      unsubConvDetail()
      unsubError()
    }
  }, [])

  const sendMessage = useCallback(
    async (text: string) => {
      const userMessage: Message = {
        id: Date.now().toString(),
        role: 'user',
        content: text,
        timestamp: new Date().toISOString(),
      }

      setMessages((prev) => [...prev, userMessage])
      setIsTyping(true)

      if (wsService.isConnected) {
        wsService.sendAgentMessage(text, conversationId || undefined)
      } else {
        try {
          const response = await agentAPI.sendMessage(text, conversationId || undefined)
          setConversationId(response.conversation_id)

          const assistantMessage: Message = {
            id: response.id,
            role: 'assistant',
            content: response.content,
            actions: response.actions,
            timestamp: response.timestamp,
          }

          setMessages((prev) => [...prev, assistantMessage])
        } catch (err) {
          const errorMessage: Message = {
            id: Date.now().toString(),
            role: 'assistant',
            content: `Error: ${err instanceof Error ? err.message : 'Connection failed. Make sure the server is running.'}`,
            timestamp: new Date().toISOString(),
          }
          setMessages((prev) => [...prev, errorMessage])
        }
        setIsTyping(false)
      }
    },
    [conversationId],
  )

  const loadConversation = useCallback((id: string) => {
    wsService.getConversation(id)
  }, [])

  const newConversation = useCallback(() => {
    setMessages([])
    setConversationId(null)
  }, [])

  const loadConversations = useCallback(() => {
    wsService.getConversations()
  }, [])

  const deleteConversation = useCallback((id: string) => {
    wsService.deleteConversation(id)
  }, [])

  const renameConversation = useCallback((id: string, title: string) => {
    wsService.renameConversation(id, title)
  }, [])

  return {
    messages,
    isTyping,
    conversationId,
    conversations,
    sendMessage,
    loadConversation,
    newConversation,
    loadConversations,
    deleteConversation,
    renameConversation,
  }
}

export function useSystemInfo(pollingInterval = 5000) {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchInfo = useCallback(async () => {
    try {
      const info = await systemAPI.getSystemInfo()
      setSystemInfo(info)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch system info')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchInfo()
    intervalRef.current = setInterval(fetchInfo, pollingInterval)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [fetchInfo, pollingInterval])

  return { systemInfo, loading, error, refresh: fetchInfo }
}

export function useFileExplorer(initialPath = 'C:\\') {
  const [currentPath, setCurrentPath] = useState(initialPath)
  const [files, setFiles] = useState<FileInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [drives, setDrives] = useState<string[]>([])

  const loadDirectory = useCallback(async (path: string) => {
    setLoading(true)
    setError(null)
    try {
      const result = await filesAPI.listDirectory(path)
      setFiles(result || [])
      setCurrentPath(path)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load directory')
      setFiles([])
    } finally {
      setLoading(false)
    }
  }, [])

  const loadDrives = useCallback(async () => {
    try {
      const result = await filesAPI.getDrives()
      setDrives(result || [])
    } catch {
    }
  }, [])

  const goUp = useCallback(() => {
    const parts = currentPath.replace(/[/\\]+$/, '').split(/[/\\]/)
    if (parts.length > 1) {
      parts.pop()
      const parentPath = parts.join('\\') || parts[0] + '\\'
      loadDirectory(parentPath)
    }
  }, [currentPath, loadDirectory])

  useEffect(() => {
    loadDirectory(initialPath)
    loadDrives()
  }, [])

  return {
    currentPath,
    files,
    loading,
    error,
    drives,
    loadDirectory,
    goUp,
    refresh: () => loadDirectory(currentPath),
  }
}

export function useProcesses(pollingInterval = 5000) {
  const [processes, setProcesses] = useState<ProcessInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [sortBy, setSortBy] = useState('cpu')
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchProcesses = useCallback(async () => {
    try {
      const procs = await systemAPI.getProcesses(sortBy, 50)
      setProcesses(procs || [])
    } catch {
    } finally {
      setLoading(false)
    }
  }, [sortBy])

  useEffect(() => {
    fetchProcesses()
    intervalRef.current = setInterval(fetchProcesses, pollingInterval)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [fetchProcesses, pollingInterval])

  const killProcess = useCallback(
    async (pid: number) => {
      try {
        await systemAPI.killProcess(pid)
        await fetchProcesses()
      } catch (err) {
        console.error('Failed to kill process:', err)
      }
    },
    [fetchProcesses],
  )

  return { processes, loading, sortBy, setSortBy, killProcess, refresh: fetchProcesses }
}

export function useTerminal() {
  const [history, setHistory] = useState<CommandResponse[]>([])
  const [output, setOutput] = useState<string[]>([])
  const [isRunning, setIsRunning] = useState(false)
  const [workingDir, setWorkingDir] = useState('C:\\')

  useEffect(() => {
    commandAPI
      .getHistory(20)
      .then((h) => {
        setHistory(h || [])
      })
      .catch(() => {})

    const unsubOutput = wsService.on('command_output', (msg: WSMessage) => {
      const data = msg.payload as { output: string; is_error: boolean; done: boolean }
      setOutput((prev) => [...prev, data.output])
      if (data.done) setIsRunning(false)
    })

    const unsubComplete = wsService.on('command_complete', (msg: WSMessage) => {
      const result = msg.payload as CommandResponse
      setHistory((prev) => [...prev, result])
      setIsRunning(false)
    })

    return () => {
      unsubOutput()
      unsubComplete()
    }
  }, [])

  const executeCommand = useCallback(
    async (command: string) => {
      setIsRunning(true)
      setOutput((prev) => [...prev, `$ ${command}\n`])

      if (wsService.isConnected) {
        wsService.sendCommand(command, workingDir)
      } else {
        try {
          const result = await commandAPI.execute(command, workingDir)
          setOutput((prev) => [...prev, result.output || ''])
          if (result.error) {
            setOutput((prev) => [...prev, `Error: ${result.error}\n`])
          }
          setHistory((prev) => [...prev, result])
        } catch (err) {
          setOutput((prev) => [
            ...prev,
            `Error: ${err instanceof Error ? err.message : 'Unknown error'}\n`,
          ])
        }
        setIsRunning(false)
      }
    },
    [workingDir],
  )

  const clearOutput = useCallback(() => {
    setOutput([])
  }, [])

  const cancelCommand = useCallback(() => {
    if (isRunning) {
      setIsRunning(false)
      setOutput(prev => [...prev, '^C\n'])
    }
  }, [isRunning])

  return {
    history,
    output,
    isRunning,
    workingDir,
    setWorkingDir,
    executeCommand,
    clearOutput,
    cancelCommand,
  }
}
