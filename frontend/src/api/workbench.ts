import { apiClient, buildApiUrl } from './client'
import { refreshToken } from './auth'

export type WorkbenchReasoningPreset = 'fast' | 'standard' | 'deep'
export type WorkbenchMessageStatus = 'pending' | 'in_progress' | 'completed' | 'failed' | 'canceled'

export interface WorkbenchEligibleKey {
  api_key_id: number
  key_name: string
  group_id: number
  group_name: string
  platform: string
  rate: number
}

export interface WorkbenchModel {
  id: string
  display_name: string
  provider: string
  provider_label: string
  default_visible: boolean
  added: boolean
  available: boolean
  unavailable_reason?: string
  binding_id?: string
  api_key_id?: number
  key_name?: string
  group_name?: string
  sort_order: number
  reasoning_presets: WorkbenchReasoningPreset[]
  eligible_keys?: WorkbenchEligibleKey[]
}

export interface WorkbenchKey {
  api_key_id: number
  key_name: string
  group_id: number
  group_name: string
  platform: string
  provider_label: string
  rate: number
  default_model_id: string
  default_model_name: string
  models: WorkbenchModel[]
  available_models?: WorkbenchModel[]
}

export interface WorkbenchModels {
	keys?: WorkbenchKey[]
  selected: WorkbenchModel[]
  available: WorkbenchModel[]
}

export interface WorkbenchMessage {
  id: string
  conversation_id: string
  sequence: number
  role: 'user' | 'assistant'
  content: string
  status: WorkbenchMessageStatus
  model_binding_id?: string
  reasoning_preset?: WorkbenchReasoningPreset
  reasoning_summary?: string
  error_message?: string
  request_id?: string
  attachments?: WorkbenchAttachment[]
  created_at: string
  updated_at: string
}

export interface WorkbenchAttachment {
  id: string
  name: string
  mime_type: string
  size_bytes: number
  data_url: string
}

export interface WorkbenchConversation {
  id: string
  model_binding_id?: string
  title: string
  reasoning_preset: WorkbenchReasoningPreset
  created_at: string
  updated_at: string
  messages?: WorkbenchMessage[]
}

export interface WorkbenchTurn {
  conversation: WorkbenchConversation
  user_message: WorkbenchMessage
  assistant_message: WorkbenchMessage
}

export interface WorkbenchStreamEvent {
  name: string
  data: Record<string, unknown>
}

let streamTokenRefresh: Promise<string> | null = null

export async function getWorkbenchModels(): Promise<WorkbenchModels> {
  const { data } = await apiClient.get<WorkbenchModels>('/workbench/models')
  return data
}

export async function addWorkbenchModels(modelIds: string[], apiKeyId: number): Promise<void> {
  await apiClient.post('/workbench/models', { model_ids: modelIds, api_key_id: apiKeyId })
}

export async function addWorkbenchModel(modelId: string, apiKeyId: number): Promise<void> {
  await addWorkbenchModels([modelId], apiKeyId)
}

export async function hideWorkbenchModel(bindingId: string): Promise<void> {
  await apiClient.delete(`/workbench/models/${encodeURIComponent(bindingId)}`)
}

export async function listWorkbenchConversations(): Promise<WorkbenchConversation[]> {
  const { data } = await apiClient.get<{ items: WorkbenchConversation[] }>('/workbench/conversations')
  return data.items ?? []
}

export async function createWorkbenchConversation(input: {
  model_binding_id?: string
  reasoning_preset?: WorkbenchReasoningPreset
}): Promise<WorkbenchConversation> {
  const { data } = await apiClient.post<WorkbenchConversation>('/workbench/conversations', input)
  return data
}

export async function getWorkbenchConversation(id: string): Promise<WorkbenchConversation> {
  const { data } = await apiClient.get<WorkbenchConversation>(`/workbench/conversations/${encodeURIComponent(id)}`)
  return data
}

export async function updateWorkbenchConversation(
  id: string,
  input: Partial<Pick<WorkbenchConversation, 'title' | 'model_binding_id' | 'reasoning_preset'>>
): Promise<WorkbenchConversation> {
  const { data } = await apiClient.patch<WorkbenchConversation>(`/workbench/conversations/${encodeURIComponent(id)}`, input)
  return data
}

export async function deleteWorkbenchConversation(id: string): Promise<void> {
  await apiClient.delete(`/workbench/conversations/${encodeURIComponent(id)}`)
}

export async function createWorkbenchMessage(
  conversationId: string,
  input: {
    content: string
    model_binding_id: string
    reasoning_preset: WorkbenchReasoningPreset
    attachments?: WorkbenchAttachment[]
  }
): Promise<WorkbenchTurn> {
  const { data } = await apiClient.post<WorkbenchTurn>(
    `/workbench/conversations/${encodeURIComponent(conversationId)}/messages`,
    input
  )
  return data
}

export async function cancelWorkbenchGeneration(messageId: string): Promise<void> {
  await apiClient.post(`/workbench/generations/${encodeURIComponent(messageId)}/cancel`)
}

export async function streamWorkbenchGeneration(
  messageId: string,
  onEvent: (event: WorkbenchStreamEvent) => void,
  signal?: AbortSignal
): Promise<void> {
  let token = localStorage.getItem('auth_token')
  let response = await fetchWorkbenchGeneration(messageId, token, signal)
  if (response.status === 401 && localStorage.getItem('refresh_token')) {
    await response.body?.cancel()
    token = await refreshWorkbenchStreamToken()
    response = await fetchWorkbenchGeneration(messageId, token, signal)
  }
  if (!response.ok || !response.body) {
    const body = await response.text()
    throw new Error(readStreamHTTPError(body, response.statusText))
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    buffer += decoder.decode(value, { stream: !done })
    const blocks = buffer.split(/\r?\n\r?\n/)
    buffer = blocks.pop() ?? ''
    for (const block of blocks) emitSSEBlock(block, onEvent)
    if (done) break
  }
  if (buffer.trim()) emitSSEBlock(buffer, onEvent)
}

function fetchWorkbenchGeneration(
  messageId: string,
  token: string | null,
  signal?: AbortSignal
): Promise<Response> {
  return fetch(buildApiUrl(`/workbench/generations/${encodeURIComponent(messageId)}/stream`), {
    method: 'POST',
    headers: {
      Accept: 'text/event-stream',
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    signal
  })
}

function refreshWorkbenchStreamToken(): Promise<string> {
  if (!streamTokenRefresh) {
    streamTokenRefresh = refreshToken()
      .then(tokens => tokens.access_token)
      .finally(() => {
        streamTokenRefresh = null
      })
  }
  return streamTokenRefresh
}

function emitSSEBlock(block: string, onEvent: (event: WorkbenchStreamEvent) => void): void {
  let eventName = ''
  const dataLines: string[] = []
  for (const line of block.split(/\r?\n/)) {
    if (line.startsWith('event:')) eventName = line.slice(6).trim()
    if (line.startsWith('data:')) dataLines.push(line.slice(5).trimStart())
  }
  const raw = dataLines.join('\n').trim()
  if (!raw || raw === '[DONE]') return
  try {
    const data = JSON.parse(raw) as Record<string, unknown>
    onEvent({ name: eventName || String(data.type || 'message'), data })
  } catch {
    // Ignore keepalive and non-JSON compatibility frames.
  }
}

function readStreamHTTPError(body: string, fallback: string): string {
  try {
    const parsed = JSON.parse(body) as { message?: string; error?: { message?: string } }
    return parsed.error?.message || parsed.message || fallback
  } catch {
    return body.trim() || fallback
  }
}

export const workbenchAPI = {
  getModels: getWorkbenchModels,
  addModel: addWorkbenchModel,
  addModels: addWorkbenchModels,
  hideModel: hideWorkbenchModel,
  listConversations: listWorkbenchConversations,
  createConversation: createWorkbenchConversation,
  getConversation: getWorkbenchConversation,
  updateConversation: updateWorkbenchConversation,
  deleteConversation: deleteWorkbenchConversation,
  createMessage: createWorkbenchMessage,
  cancelGeneration: cancelWorkbenchGeneration,
  streamGeneration: streamWorkbenchGeneration
}

export default workbenchAPI
