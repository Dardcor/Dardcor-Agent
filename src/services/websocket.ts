import type { WSMessage } from '../types'

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
          console.log('🔌 WebSocket connected to Dardcor Agent')
          this.reconnectAttempts = 0
          this.isConnecting = false
          this.emit('connection', { type: 'connection', payload: { status: 'connected' } })
          this.startPing()
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data)
            this.emit(message.type, message)
            this.emit('*', message)
          } catch {
            console.error('Failed to parse WS message')
          }
        }

        this.ws.onclose = (event) => {
          console.log(`🔌 WebSocket disconnected: ${event.code}`)
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
      console.log('Max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1), 30000)
    console.log(`Reconnecting in ${Math.round(delay)}ms (attempt ${this.reconnectAttempts})`)

    setTimeout(() => {
      this.connect().catch(() => {
        // Will retry again from onclose handler
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
      console.warn('WebSocket not connected, attempting to connect...')
      this.connect()
        .then(() => {
          this.doSend(type, payload)
        })
        .catch(() => {
          console.error('Cannot send message: WebSocket not available')
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
          console.error(`WS handler error for ${type}:`, err)
        }
      })
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  // Convenience methods
  sendAgentMessage(message: string, conversationId?: string) {
    this.send('agent_message', { message, conversation_id: conversationId })
  }

  sendCommand(command: string, workingDir?: string) {
    this.send('execute_command', { command, working_dir: workingDir })
  }

  getConversations() {
    this.send('get_conversations', {})
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
}

export const wsService = new WebSocketService()
export default wsService
