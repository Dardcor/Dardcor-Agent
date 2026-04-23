import type { WSMessage, AgentTurnEvent, ToolExecutionEvent } from '../types'

type MessageHandler = (message: WSMessage) => void

class WebSocketService {
  private ws: WebSocket | null = null
  private handlers: Map<string, MessageHandler[]> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 15
  private reconnectDelay = 1000
  private url: string
  private isConnecting = false
  private pingInterval: ReturnType<typeof setInterval> | null = null
  public isAgentTyping: boolean = false

  constructor() {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    this.url = `${wsProtocol}//${host}/ws`
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      if (this.isConnecting) {
        resolve()
        return
      }

      this.isConnecting = true

      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          this.reconnectAttempts = 0
          this.isConnecting = false
          this.emit('connection', { type: 'connection', payload: { status: 'connected' } })
          this.startPing()
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data)
            if (message.type === 'typing') {
              this.isAgentTyping = (message.payload as any)?.typing === true
            } else if (message.type === 'agent_response' || message.type === 'error') {
              this.isAgentTyping = false
            }
            this.emit(message.type, message)
            this.emit('*', message)
          } catch { }
        }

        this.ws.onclose = (event) => {
          this.isConnecting = false
          this.stopPing()
          this.emit('connection', { type: 'connection', payload: { status: 'disconnected' } })
          this.attemptReconnect()
        }

        this.ws.onerror = () => {
          this.isConnecting = false
          this.emit('connection', { type: 'connection', payload: { status: 'error' } })
          reject(new Error('WebSocket connection failed'))
        }
      } catch (err) {
        this.isConnecting = false
        reject(err)
      }
    })
  }

  private startPing() {
    this.stopPing()
    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.doSend('ping', {})
      }
    }, 30000)
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1), 30000)

    setTimeout(() => {
      this.connect().catch(() => {
      })
    }, delay)
  }

  disconnect() {
    this.stopPing()
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
      this.ws = null
    }
  }

  send(type: string, payload: unknown) {
    if (this.ws?.readyState !== WebSocket.OPEN) {
      this.connect()
        .then(() => {
          this.doSend(type, payload)
        })
        .catch(() => {
        })
      return
    }

    this.doSend(type, payload)
  }

  private doSend(type: string, payload: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message: WSMessage = { type, payload }
      this.ws.send(JSON.stringify(message))
    }
  }

  on(type: string, handler: MessageHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, [])
    }
    this.handlers.get(type)!.push(handler)

    return () => {
      const handlers = this.handlers.get(type)
      if (handlers) {
        const index = handlers.indexOf(handler)
        if (index !== -1) {
          handlers.splice(index, 1)
        }
      }
    }
  }

  private emit(type: string, message: WSMessage) {
    const handlers = this.handlers.get(type)
    if (handlers) {
      handlers.forEach((handler) => {
        try {
          handler(message)
        } catch (err) {
        }
      })
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  sendAgentMessage(message: string, conversationId?: string) {
    this.send('agent_message', { message, conversation_id: conversationId })
  }

  sendCommand(command: string, workingDir?: string) {
    this.send('execute_command', { command, working_dir: workingDir })
  }

  getConversations() {
    this.send('get_conversations', {})
  }

  createConversation(title?: string) {
    this.send('create_conversation', { title })
  }

  getConversation(id: string) {
    this.send('get_conversation', { id })
  }

  deleteConversation(id: string) {
    this.send('delete_conversation', { id })
  }

  renameConversation(id: string, title: string) {
    this.send('rename_conversation', { id, title })
  }

  stopAgent() {
    this.send('stop_agent', {})
  }

  onAgentTurn(handler: (payload: AgentTurnEvent) => void): () => void {
    return this.on('agent_turn', (msg) => handler(msg.payload as AgentTurnEvent))
  }

  onToolProgress(handler: (payload: ToolExecutionEvent) => void): () => void {
    const unsubStart = this.on('tool_start', (msg) => handler({ ...(msg.payload as ToolExecutionEvent), status: 'start' }))
    const unsubEnd = this.on('tool_end', (msg) => handler({ ...(msg.payload as ToolExecutionEvent), status: 'end' }))
    return () => {
      unsubStart()
      unsubEnd()
    }
  }

  onAgentStatus(handler: (payload: any) => void): () => void {
    return this.on('agent_status', (msg) => handler(msg.payload))
  }
}

export const wsService = new WebSocketService()
export default wsService




