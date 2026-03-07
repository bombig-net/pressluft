import type {
  JobKind,
  JobStatus,
  NodeStatus,
  ServerStatus,
  SetupState,
  SupportLevel,
} from '~/lib/platform-contract.generated'

export interface AuthActor {
  id: string
  type: string
  email: string
  role: string
  authenticated: boolean
  auth_source?: string
}

export interface ProviderType {
  type: string
  name: string
  docs_url: string
}

export interface StoredProvider {
  id: number
  type: string
  name: string
  status: string
  created_at: string
  updated_at: string
}

export interface ValidationResult {
  valid: boolean
  read_write: boolean
  message: string
  project_name?: string
}

export interface Activity {
  id: number
  event_type: string
  category: string
  level: 'info' | 'success' | 'warning' | 'error' | string
  resource_type?: string
  resource_id?: number
  parent_resource_type?: string
  parent_resource_id?: number
  actor_type: string
  actor_id?: string
  title: string
  message?: string
  payload?: string
  requires_attention: boolean
  read_at?: string
  created_at: string
}

export interface ActivityListResponse {
  data: Activity[]
  next_cursor?: string
}

export interface ServerProfile {
  key: string
  name: string
  description: string
  artifact_path: string
  support_level: SupportLevel
  configure_guarantee: string
  support_reason?: string
}

export interface ServerLocation {
  name: string
  description: string
  country?: string
  city?: string
  network_zone?: string
}

export interface ServerTypePrice {
  location_name: string
  hourly_gross: string
  monthly_gross: string
  currency: string
}

export interface ServerTypeOption {
  name: string
  description: string
  cores: number
  memory_gb: number
  disk_gb: number
  architecture: string
  available_at: string[]
  prices: ServerTypePrice[]
}

export interface ServerCatalog {
  locations: ServerLocation[]
  server_types: ServerTypeOption[]
}

export interface StoredServer {
  id: number
  provider_id: number
  provider_type: string
  provider_server_id?: string
  name: string
  location: string
  server_type: string
  image: string
  profile_key: string
  status: ServerStatus
  setup_state: SetupState
  setup_last_error?: string
  action_id?: string
  action_status?: string
  node_status?: NodeStatus
  node_last_seen?: string
  node_version?: string
  created_at: string
  updated_at: string
}

export interface AgentInfo {
  connected: boolean
  status: NodeStatus
  last_seen?: string
  version?: string
  cpu_percent?: number
  mem_used_mb?: number
  mem_total_mb?: number
}

export interface Service {
  name: string
  description: string
  active_state: string
  load_state: string
}

export interface ServicesResponse {
  server_id: number
  agent_connected: boolean
  services: Service[]
}

export interface CreateServerResponse {
  server_id: number
  job_id: number
  status: ServerStatus
}

export interface DeleteServerResponse {
  server_id: number
  job_id: number
  status: ServerStatus
  job_status: JobStatus
  async: boolean
  description: string
}

export interface Job {
  id: number
  server_id?: number
  kind: JobKind
  status: JobStatus
  current_step: string
  retry_count: number
  last_error?: string
  payload?: string
  created_at: string
  updated_at: string
}

export interface JobEvent {
  job_id: number
  seq: number
  event_type: string
  level: string
  step_key?: string
  status?: string
  message: string
  payload?: string
  occurred_at: string
}
