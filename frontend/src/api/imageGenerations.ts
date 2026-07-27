import { apiClient } from './client'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export interface ImageGenerationImage {
  index: number
  url: string
  mime_type: string
  size_bytes: number
  path?: string
}

export interface ImageGenerationRecord {
  id: number
  api_key_id?: number
  group_id?: number
  request_id: string
  upstream_request_id?: string
  model: string
  prompt: string
  size: string
  resolution_tier: '1K' | '2K' | '4K'
  aspect_ratio: '1:1' | '2:3' | '3:2'
  quality: string
  output_format: string
  n: number
  status: string
  storage_status: string
  image_count: number
  images: ImageGenerationImage[]
  file_size_bytes: number
  error_message?: string
  created_at: string
  completed_at?: string
}

export interface CreateImageGenerationRequest {
  api_key_id?: number
  api_key?: string
  model: string
  prompt: string
  source_image?: string
  size: string
  resolution_tier: '1K' | '2K' | '4K'
  aspect_ratio: '1:1' | '2:3' | '3:2'
  quality: string
  output_format: string
  n: number
}

export interface ImageGenerationPromptVersion {
  version: number
  prompt: string
  model: string
  size: string
  quality: string
  output_format: string
  created_at: string
}

export interface PaginatedImageGenerations {
  items: ImageGenerationRecord[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface ImageGenerationPricing {
  currency: 'USD'
  tiers: Record<'1K' | '2K' | '4K', number>
}

export interface ImageGenerationKeyCapability {
  api_key_id: number
  key_name: string
  group_id?: number
  group_name?: string
  status: string
  available: boolean
  allowed_tiers: Array<'1K' | '2K' | '4K'>
  unavailable_reason?: string
}

export interface ImageGenerationCapabilities {
  keys: ImageGenerationKeyCapability[]
  tiers: Record<'1K' | '2K' | '4K', number[]>
}

export const imageGenerationsAPI = {
  pricing(): Promise<ImageGenerationPricing> {
    return apiClient.get('/image-generations/pricing').then((res) => res.data)
  },

  capabilities(): Promise<ImageGenerationCapabilities> {
    return apiClient.get('/image-generations/capabilities').then((res) => res.data)
  },

  create(data: CreateImageGenerationRequest): Promise<ImageGenerationRecord> {
    return apiClient.post('/image-generations', data, { timeout: 300000 }).then((res) => res.data)
  },

  list(page = 1, pageSize = 20): Promise<PaginatedImageGenerations> {
    return apiClient
      .get('/image-generations', { params: { page, page_size: pageSize } })
      .then((res) => res.data)
  },

  get(id: number): Promise<ImageGenerationRecord> {
    return apiClient.get(`/image-generations/${id}`).then((res) => res.data)
  },

  delete(id: number): Promise<void> {
    return apiClient.delete(`/image-generations/${id}`).then(() => {})
  },

  promptVersions(id: number): Promise<{ items: ImageGenerationPromptVersion[] }> {
    return apiClient.get(`/image-generations/${id}/prompt-versions`).then((res) => res.data)
  },

  getImageUrl(id: number, index: number, mode: 'preview' | 'download' = 'preview'): string {
    return `${API_BASE_URL}/image-generations/${id}/images/${index}/${mode}`
  },

  getImageBlob(id: number, index: number, mode: 'preview' | 'download' = 'preview'): Promise<Blob> {
    return apiClient
      .get(`/image-generations/${id}/images/${index}/${mode}`, { responseType: 'blob' })
      .then((res) => res.data)
  }
}
