import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { ImageGenerationRecord } from '@/api/imageGenerations'
import ImageGenerationResult from '../ImageGenerationResult.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

function record(overrides: Partial<ImageGenerationRecord> = {}): ImageGenerationRecord {
  return {
    id: 3,
    request_id: 'request-3',
    model: 'gpt-image-2',
    prompt: 'poster',
    size: '2880x2880',
    resolution_tier: '4K',
    aspect_ratio: '1:1',
    quality: 'auto',
    output_format: 'png',
    n: 1,
    status: 'success',
    storage_status: 'saved',
    image_count: 0,
    images: [],
    file_size_bytes: 0,
    created_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

describe('ImageGenerationResult', () => {
  it('shows the active request tier instead of stale metadata from the previous result', () => {
    const wrapper = mount(ImageGenerationResult, {
      props: {
        record: record(),
        activeGeneration: {
          model: 'gpt-image-2',
          resolution_tier: '1K',
          size: '1024x1024'
        },
        generating: true,
        errorMessage: '',
        previewUrl: vi.fn()
      },
      global: { stubs: { Icon: true } }
    })

    const tags = wrapper.get('.result-tags').text()
    expect(tags).toContain('1K')
    expect(tags).toContain('1024x1024')
    expect(tags).not.toContain('4K')
    expect(tags).not.toContain('2880x2880')
  })

  it('shows the stored result metadata after generation completes', () => {
    const wrapper = mount(ImageGenerationResult, {
      props: {
        record: record(),
        activeGeneration: {
          model: 'gpt-image-2',
          resolution_tier: '1K',
          size: '1024x1024'
        },
        generating: false,
        errorMessage: '',
        previewUrl: vi.fn()
      },
      global: { stubs: { Icon: true } }
    })

    const tags = wrapper.get('.result-tags').text()
    expect(tags).toContain('4K')
    expect(tags).toContain('2880x2880')
  })
})
