import { ref, readonly } from 'vue'
import type {
  AgentInfo,
  CreateServerResponse,
  DeleteServerResponse,
  ServerCatalog,
  ServerProfile,
  ServerTypePrice,
  ServicesResponse,
  StoredServer,
} from '~/lib/api-contract'
export type {
  AgentInfo,
  ServerCatalog,
  ServerProfile,
  ServerTypePrice,
  ServicesResponse,
  StoredServer,
} from '~/lib/api-contract'

export interface CreateServerInput {
  provider_id: number
  name: string
  location: string
  server_type: string
  profile_key: string
}

export type AgentStatusType = AgentInfo['status']

export function useServers() {
  const { apiFetch } = useApiClient()
  const servers = ref<StoredServer[]>([])
  const profiles = ref<ServerProfile[]>([])
  const catalog = ref<ServerCatalog | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')

  const fetchServers = async () => {
    loading.value = true
    error.value = ''
    try {
      servers.value = await apiFetch<StoredServer[]>('/servers')
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  const fetchCatalog = async (providerId: number) => {
    error.value = ''
    catalog.value = null
    profiles.value = []
    const body = await apiFetch<{ catalog: ServerCatalog; profiles: ServerProfile[] }>(
      `/servers/catalog?provider_id=${providerId}`,
    )
    catalog.value = body.catalog
    profiles.value = body.profiles
  }

  const createServer = async (payload: CreateServerInput): Promise<CreateServerResponse> => {
    saving.value = true
    error.value = ''
    try {
      return await apiFetch<CreateServerResponse>('/servers', {
        method: 'POST',
        body: payload,
      })
    } finally {
      saving.value = false
    }
  }

  const deleteServer = async (serverId: number): Promise<DeleteServerResponse> => {
    error.value = ''
    return await apiFetch<DeleteServerResponse>(`/servers/${serverId}`, {
      method: 'DELETE',
    })
  }

  const fetchServer = async (serverId: number): Promise<StoredServer> => {
    error.value = ''
    return await apiFetch<StoredServer>(`/servers/${serverId}`)
  }

  const fetchAgentStatus = async (serverId: number): Promise<AgentInfo> => {
    return await apiFetch<AgentInfo>(`/servers/${serverId}/agent-status`)
  }

  const fetchAllAgentStatus = async (): Promise<Record<number, AgentInfo>> => {
    return await apiFetch<Record<number, AgentInfo>>('/servers/agents')
  }

  const fetchServices = async (serverId: number): Promise<ServicesResponse> => {
    return await apiFetch<ServicesResponse>(`/servers/${serverId}/services`)
  }

  return {
    servers: readonly(servers),
    profiles: readonly(profiles),
    catalog: readonly(catalog),
    loading: readonly(loading),
    saving: readonly(saving),
    error: readonly(error),
    fetchServers,
    fetchServer,
    fetchCatalog,
    createServer,
    deleteServer,
    fetchAgentStatus,
    fetchAllAgentStatus,
    fetchServices,
  }
}
