import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { ImageGenerationRecord } from '@/api/imageGenerations'
import ImageGenerationActivity from '../ImageGenerationActivity.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

function record(id: number): ImageGenerationRecord {
  return {
    id,
    request_id: `request-${id}`,
    model: 'gpt-image-2',
    prompt: `poster-${id}`,
    size: '1024x1024',
    resolution_tier: '1K',
    aspect_ratio: '1:1',
    quality: 'auto',
    output_format: 'png',
    n: 1,
    status: 'success',
    storage_status: 'saved',
    image_count: 0,
    images: [],
    file_size_bytes: 0,
    created_at: `2026-01-0${id}T00:00:00Z`
  }
}

describe('ImageGenerationActivity', () => {
  it('shows the five most recent generations and keeps the fifth selectable', async () => {
    const history = Array.from({ length: 6 }, (_, index) => record(index + 1))
    const wrapper = mount(ImageGenerationActivity, {
      props: {
        queueItems: [],
        queueRunning: false,
        history,
        loadingHistory: false,
        previewUrl: vi.fn()
      },
      global: { stubs: { Icon: true } }
    })

    for (const id of [1, 2, 3, 4, 5]) {
      expect(wrapper.text()).toContain(`poster-${id}`)
    }
    expect(wrapper.text()).not.toContain('poster-6')

    const fifthRecord = wrapper.findAll('button').find((button) => button.text().includes('poster-5'))
    expect(fifthRecord).toBeDefined()
    await fifthRecord?.trigger('click')

    expect(wrapper.emitted('selectRecord')?.[0]).toEqual([history[4]])
  })
})
