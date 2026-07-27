import { defineComponent, h, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ImageGenerationKeyCapability, ImageGenerationRecord } from '@/api/imageGenerations'
import {
  resetImageGenerationWorkspace,
  useImageGenerationWorkspace
} from '../useImageGenerationWorkspace'

const { getCapabilities, getPricing, createImage, listImages, getImageBlob, showError, showSuccess } = vi.hoisted(() => ({
  getCapabilities: vi.fn(),
  getPricing: vi.fn(),
  createImage: vi.fn(),
  listImages: vi.fn(),
  getImageBlob: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/imageGenerations', () => ({
  imageGenerationsAPI: {
    capabilities: getCapabilities,
    pricing: getPricing,
    create: createImage,
    list: listImages,
    getImageUrl: vi.fn((id: number, index: number) => `/preview/${id}/${index}`),
    getImageBlob,
    delete: vi.fn(),
    promptVersions: vi.fn()
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

function capability(id: number, tiers: Array<'1K' | '2K' | '4K'>): ImageGenerationKeyCapability {
  const allowedTiers = tiers.includes('2K') || tiers.includes('4K')
    ? tiers.filter((tier) => tier !== '1K')
    : tiers
  return {
    api_key_id: id,
    key_name: `key-${id}`,
    group_id: id,
    group_name: `group-${id}`,
    status: 'active',
    available: true,
    allowed_tiers: allowedTiers
  }
}

function record(overrides: Partial<ImageGenerationRecord> = {}): ImageGenerationRecord {
  return {
    id: 99,
    api_key_id: 2,
    group_id: 2,
    request_id: 'request-99',
    model: 'gpt-image-2',
    prompt: 'coast at sunset',
    size: '1365x2048',
    resolution_tier: '2K',
    aspect_ratio: '2:3',
    quality: 'auto',
    output_format: 'png',
    n: 1,
    status: 'success',
    storage_status: 'stored',
    image_count: 1,
    images: [],
    file_size_bytes: 0,
    created_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

async function mountWorkspace(keys: ImageGenerationKeyCapability[]) {
  getCapabilities.mockResolvedValue({
    keys,
    tiers: {
      '1K': keys.filter((key) => key.allowed_tiers.includes('1K')).map((key) => key.api_key_id),
      '2K': keys.filter((key) => key.allowed_tiers.includes('2K')).map((key) => key.api_key_id),
      '4K': keys.filter((key) => key.allowed_tiers.includes('4K')).map((key) => key.api_key_id)
    }
  })
  getPricing.mockResolvedValue({ currency: 'USD', tiers: { '1K': 0.06, '2K': 0.16, '4K': 0.2 } })
  listImages.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 24, pages: 0 })

  let workspace!: ReturnType<typeof useImageGenerationWorkspace>
  const Harness = defineComponent({
    setup() {
      workspace = useImageGenerationWorkspace()
      return () => h('div')
    }
  })
  const wrapper = mount(Harness)
  await flushPromises()
  return { wrapper, workspace }
}

describe('useImageGenerationWorkspace', () => {
  beforeEach(() => {
    resetImageGenerationWorkspace()
    vi.clearAllMocks()
    getImageBlob.mockReset()
  })

  it('keeps standard, 2K, and 4K keys in their permitted pools', async () => {
    const { workspace } = await mountWorkspace([
      capability(1, ['1K']),
      capability(2, ['2K']),
      capability(3, ['4K']),
      capability(4, ['2K', '4K']),
      capability(5, ['1K', '2K', '4K'])
    ])

    expect(workspace.standardKeys.value.map((key) => key.api_key_id)).toEqual([1])
    workspace.form.resolution_tier = '2K'
    await nextTick()
    expect(workspace.hdKeys.value.map((key) => key.api_key_id)).toEqual([2, 4, 5])
    workspace.form.resolution_tier = '4K'
    await nextTick()
    expect(workspace.hdKeys.value.map((key) => key.api_key_id)).toEqual([3, 4, 5])
  })

  it('auto-selects an eligible image key and preserves a manual key change', async () => {
    const { workspace } = await mountWorkspace([
      capability(2, ['2K', '4K']),
      capability(3, ['2K', '4K'])
    ])
    workspace.form.resolution_tier = '2K'
    await nextTick()

    expect(workspace.selectedKeyByPool.hd).toBe('2')
    expect(workspace.selectedApiKey.value?.key_name).toBe('key-2')

    workspace.selectedKeyByPool.hd = '3'
    await nextTick()
    expect(workspace.selectedApiKey.value?.key_name).toBe('key-3')
  })

  it('uses the backend fixed price list for estimates', async () => {
    const { workspace } = await mountWorkspace([capability(2, ['2K'])])
    workspace.form.resolution_tier = '2K'
    workspace.form.n = 3
    await nextTick()

    expect(workspace.tierPrices.value).toEqual({ '1K': 0.06, '2K': 0.16, '4K': 0.2 })
    expect(workspace.estimatedUnitPrice.value).toBe(0.16)
    expect(workspace.estimatedTotalPrice.value).toBeCloseTo(0.48)
  })

  it('switches to the matching key pool and disables unsupported HD generation', async () => {
    const { workspace } = await mountWorkspace([capability(1, ['1K'])])
    workspace.form.prompt = 'a quiet coast'
    expect(workspace.canGenerate.value).toBe(true)

    workspace.form.resolution_tier = '4K'
    await nextTick()
    expect(workspace.hdKeys.value).toEqual([])
    expect(workspace.canGenerate.value).toBe(false)
  })

  it('sends the selected tier, aspect ratio, dimensions, and key id', async () => {
    createImage.mockResolvedValue(record())
    const { workspace } = await mountWorkspace([capability(2, ['2K'])])
    workspace.form.prompt = 'coast at sunset'
    workspace.form.resolution_tier = '2K'
    workspace.form.aspect_ratio = '2:3'
    await nextTick()

    await workspace.generate()

    expect(createImage).toHaveBeenCalledWith(expect.objectContaining({
      api_key_id: 2,
      resolution_tier: '2K',
      aspect_ratio: '2:3',
      size: '1365x2048'
    }))
  })

  it('keeps active result metadata fixed to the request snapshot while generating', async () => {
    let finishGeneration!: (value: ImageGenerationRecord) => void
    createImage.mockImplementation(() => new Promise((resolve) => {
      finishGeneration = resolve
    }))
    const { workspace } = await mountWorkspace([capability(1, ['1K'])])
    workspace.form.prompt = 'coast at sunset'

    const generation = workspace.generate()
    expect(workspace.activeGeneration.value).toEqual({
      model: 'gpt-image-2',
      resolution_tier: '1K',
      size: '1024x1024'
    })

    workspace.form.resolution_tier = '4K'
    await nextTick()
    expect(workspace.activeGeneration.value?.resolution_tier).toBe('1K')

    finishGeneration(record({
      resolution_tier: '1K',
      size: '1024x1024',
      status: 'failed',
      error_message: 'test failure'
    }))
    await generation
  })

  it('keeps an in-flight generation synchronized after the page remounts', async () => {
    let finishGeneration!: (value: ImageGenerationRecord) => void
    createImage.mockImplementation(() => new Promise((resolve) => {
      finishGeneration = resolve
    }))
    getImageBlob.mockResolvedValue(new Blob(['image'], { type: 'image/png' }))

    const firstMount = await mountWorkspace([capability(1, ['1K'])])
    firstMount.workspace.form.prompt = 'coast at sunset'
    const generation = firstMount.workspace.generate()

    expect(firstMount.workspace.generating.value).toBe(true)
    expect(firstMount.workspace.canGenerate.value).toBe(false)
    expect(createImage).toHaveBeenCalledTimes(1)

    firstMount.wrapper.unmount()
    const secondMount = await mountWorkspace([capability(1, ['1K'])])

    expect(secondMount.workspace).toBe(firstMount.workspace)
    expect(secondMount.workspace.generating.value).toBe(true)
    expect(secondMount.workspace.canGenerate.value).toBe(false)

    await secondMount.workspace.generate()
    expect(createImage).toHaveBeenCalledTimes(1)

    finishGeneration(record({
      resolution_tier: '1K',
      size: '1024x1024',
      images: [{
        index: 0,
        url: '/api/v1/image-generations/99/images/0/preview',
        mime_type: 'image/png',
        size_bytes: 5
      }]
    }))
    await generation
    await flushPromises()

    expect(secondMount.workspace.generating.value).toBe(false)
    expect(secondMount.workspace.currentRecord.value?.id).toBe(99)
    expect(secondMount.workspace.previewUrl(99, 0)).toMatch(/^data:image\/png;base64,/)
  })

  it('never exposes an authenticated preview endpoint directly to image elements', async () => {
    getImageBlob.mockRejectedValue(new Error('preview unavailable'))
    const { workspace } = await mountWorkspace([capability(1, ['1K'])])
    const imageRecord = record({
      resolution_tier: '1K',
      size: '1024x1024',
      images: [{
        index: 0,
        url: '/api/v1/image-generations/99/images/0/preview',
        mime_type: 'image/png',
        size_bytes: 1024
      }]
    })

    expect(workspace.previewUrl(imageRecord.id, 0, imageRecord.images[0]?.url)).toMatch(/^data:image\/gif;base64,/)

    await workspace.selectRecord(imageRecord)

    expect(getImageBlob).toHaveBeenCalledWith(imageRecord.id, 0, 'preview')
    expect(workspace.previewUrl(imageRecord.id, 0, imageRecord.images[0]?.url)).toMatch(/^data:image\/gif;base64,/)
    expect(workspace.previewUrl(imageRecord.id, 0, imageRecord.images[0]?.url)).not.toContain('/image-generations/')
  })

  it('uses a manually entered existing key instead of the selected key id', async () => {
    createImage.mockResolvedValue(record({ api_key_id: 2 }))
    const { workspace } = await mountWorkspace([capability(1, ['1K']), capability(2, ['1K'])])
    workspace.form.prompt = 'coast at sunset'
    workspace.manualApiKey.value = 'sk-test-2'

    await workspace.generate()

    expect(createImage).toHaveBeenCalledWith(expect.objectContaining({
      api_key: 'sk-test-2',
      api_key_id: undefined
    }))
  })

  it('keeps queue snapshots independent from later form changes', async () => {
    const { workspace } = await mountWorkspace([capability(2, ['2K']), capability(3, ['4K'])])
    workspace.form.prompt = 'first prompt'
    workspace.form.resolution_tier = '2K'
    await nextTick()
    workspace.enqueueCurrent()

    workspace.form.prompt = 'second prompt'
    workspace.form.resolution_tier = '4K'
    await nextTick()

    expect(workspace.queueItems.value).toHaveLength(1)
    expect(workspace.queueItems.value[0]).toMatchObject({
      prompt: 'first prompt',
      resolution_tier: '2K',
      api_key_id: 2,
      status: 'waiting'
    })
  })
})
