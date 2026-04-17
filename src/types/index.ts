export interface AgentRequest {
  message: string;
  conversation_id?: string;
}

export interface AgentResponse {
  id: string;
  conversation_id: string;
  role: string;
  content: string;
  actions?: Action[];
  timestamp: string;
  status: string;
}

export interface Action {
  type: string;
  description: string;
  parameters?: Record<string, unknown>;
  result?: unknown;
  status: string;
  error?: string;
  duration_ms?: number;
}

export interface Conversation {
  id: string;
  title: string;
  messages: Message[];
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  actions?: Action[];
  timestamp: string;
}

export interface FileInfo {
  name: string;
  path: string;
  size: number;
  is_dir: boolean;
  extension: string;
  modified_at: string;
  permission: string;
}

export interface FileContent {
  path: string;
  content: string;
  encoding: string;
  size: number;
}

export interface CommandRequest {
  command: string;
  working_dir?: string;
  timeout?: number;
}

export interface CommandResponse {
  id: string;
  command: string;
  output: string;
  error?: string;
  exit_code: number;
  duration_ms: number;
  started_at: string;
  finished_at: string;
}

export interface SystemInfo {
  os: OSInfo;
  cpu: CPUInfo;
  memory: MemoryInfo;
  disks: DiskInfo[];
  network: NetInfo[];
  uptime: number;
  hostname: string;
  collected_at: string;
}

export interface OSInfo {
  platform: string;
  family: string;
  edition: string;
  arch: string;
}

export interface CPUInfo {
  model_name: string;
  cores: number;
  threads: number;
  usage_percent: number;
  frequency_mhz: number;
}

export interface MemoryInfo {
  total: number;
  used: number;
  free: number;
  used_percent: number;
}

export interface DiskInfo {
  device: string;
  mount_point: string;
  fs_type: string;
  total: number;
  used: number;
  free: number;
  used_percent: number;
}

export interface NetInfo {
  name: string;
  mac_address: string;
  addresses: string[];
  bytes_sent: number;
  bytes_recv: number;
}

export interface ProcessInfo {
  pid: number;
  name: string;
  status: string;
  cpu_percent: number;
  mem_percent: number;
  memory_mb: number;
  username: string;
  create_time: number;
  cmdline: string;
}

export interface WSMessage {
  type: string;
  payload: unknown;
}

export interface WSCommandOutput {
  command_id: string;
  output: string;
  is_error: boolean;
  done: boolean;
}

export interface APIResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export type TabType = 'chat' | 'files' | 'terminal' | 'system' | 'workspace';

export interface Notification {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  duration?: number;
}

export interface AgentTurnEvent {
  conversation_id: string;
  message_id: string;
  content: string;
  is_final: boolean;
}

export interface ToolExecutionEvent {
  conversation_id: string;
  message_id: string;
  tool_name: string;
  status: 'start' | 'end';
  result?: unknown;
}

export interface SkillConfig {
  id: string;
  name: string;
  description: string;
  category?: string;
  enabled?: boolean;
  tags?: string[];
}

export interface ToolConfig {
  id: string;
  name: string;
  description: string;
  mcp?: boolean;
  last_used?: string;
  usage_count?: number;
  category?: string;
}

