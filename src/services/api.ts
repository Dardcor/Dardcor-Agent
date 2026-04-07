import type {
  APIResponse,
  FileInfo,
  FileContent,
  CommandResponse,
  SystemInfo,
  ProcessInfo,
  MemoryInfo,
  AgentResponse,
} from '../types'

const API_BASE = '/api'

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 30000)

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
      },
      signal: controller.signal,
      ...options,
    })

    clearTimeout(timeoutId)

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const data: APIResponse<T> = await response.json()

    if (!data.success) {
      throw new Error(data.error || 'API request failed')
    }

    return data.data as T
  } catch (err) {
    clearTimeout(timeoutId)

    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new Error('Request timeout — server not responding')
    }

    if (err instanceof TypeError && err.message.includes('fetch')) {
      throw new Error('Connection failed — make sure Dardcor Agent is running on port 25000')
    }

    throw err
  }
}

export const agentAPI = {
  sendMessage: (message: string, conversationId?: string) =>
    fetchAPI<AgentResponse>('/agent', {
      method: 'POST',
      body: JSON.stringify({
        message,
        conversation_id: conversationId,
      }),
    }),
}

export const filesAPI = {
  listDirectory: (path: string) =>
    fetchAPI<FileInfo[]>(`/files?path=${encodeURIComponent(path)}`),

  readFile: (path: string) =>
    fetchAPI<FileContent>(`/files/read?path=${encodeURIComponent(path)}`),

  writeFile: (path: string, content: string) =>
    fetchAPI<void>('/files/write', {
      method: 'POST',
      body: JSON.stringify({ path, content }),
    }),

  deleteFile: (path: string) =>
    fetchAPI<void>(`/files?path=${encodeURIComponent(path)}`, {
      method: 'DELETE',
    }),

  searchFiles: (path: string, query: string, searchContent = false) =>
    fetchAPI<FileInfo[]>('/files/search', {
      method: 'POST',
      body: JSON.stringify({ path, query, search_content: searchContent }),
    }),

  getFileInfo: (path: string) =>
    fetchAPI<FileInfo>(`/files/info?path=${encodeURIComponent(path)}`),

  createDirectory: (path: string) =>
    fetchAPI<void>('/files/mkdir', {
      method: 'POST',
      body: JSON.stringify({ path }),
    }),

  getDrives: () => fetchAPI<string[]>('/files/drives'),

  moveFile: (source: string, destination: string) =>
    fetchAPI<void>('/files/move', {
      method: 'POST',
      body: JSON.stringify({ source, destination }),
    }),

  copyFile: (source: string, destination: string) =>
    fetchAPI<void>('/files/copy', {
      method: 'POST',
      body: JSON.stringify({ source, destination }),
    }),
}

export const commandAPI = {
  execute: (command: string, workingDir?: string, timeout?: number) =>
    fetchAPI<CommandResponse>('/command', {
      method: 'POST',
      body: JSON.stringify({
        command,
        working_dir: workingDir,
        timeout,
      }),
    }),

  getHistory: (limit = 50) =>
    fetchAPI<CommandResponse[]>(`/command/history?limit=${limit}`),

  getShellInfo: () => fetchAPI<Record<string, string>>('/command/info'),
}

export const systemAPI = {
  getSystemInfo: () => fetchAPI<SystemInfo>('/system'),

  getProcesses: (sort = 'cpu', limit = 50) =>
    fetchAPI<ProcessInfo[]>(`/system/processes?sort=${sort}&limit=${limit}`),

  killProcess: (pid: number) =>
    fetchAPI<void>('/system/kill', {
      method: 'POST',
      body: JSON.stringify({ pid }),
    }),

  getCPUUsage: () => fetchAPI<number[]>('/system/cpu'),

  getMemoryUsage: () => fetchAPI<MemoryInfo>('/system/memory'),
}
