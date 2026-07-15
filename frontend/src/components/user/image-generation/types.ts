import type { ImageGenerationRecord } from '@/api/imageGenerations'

export type ImageGenerationMode = 'text' | 'image'
export type ImageResolutionTier = '1K' | '2K' | '4K'
export type ImageAspectRatio = '1:1' | '2:3' | '3:2'
export type ImageKeyPool = 'standard' | 'hd'

export interface ImageGenerationFormState {
  model: string
  prompt: string
  resolution_tier: ImageResolutionTier
  aspect_ratio: ImageAspectRatio
  quality: string
  output_format: string
  n: number
  source_image: string
}

export interface ImageGenerationQueueItem extends ImageGenerationFormState {
  id: string
  mode: ImageGenerationMode
  size: string
  api_key_id?: number
  api_key_name?: string
  manual_api_key?: string
  source_image_name?: string
  status: 'waiting' | 'running'
}

export interface ImageGenerationTierOption {
  value: ImageResolutionTier
  label: string
  description: string
}

export interface ImageGenerationViewer {
  record: ImageGenerationRecord
  imageIndex: number
}
