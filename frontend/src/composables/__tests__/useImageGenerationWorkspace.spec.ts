import { defineComponent, h, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ApiKey, Group } from '@/types'
import type { ImageGenerationRecord } from '@/api/imageGenerations'
import { useImageGenerationWorkspace } from '../useImageGenerationWorkspace'

const { listKeys, getPricing, createImage, listImages, showError, showSuccess } = vi.hoisted(() => ({
  listKeys: vi.fn(),
  getPricing: vi.fn(),
  createImage: vi.fn(),
  listImages: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/keys', () => ({
  keysAPI: { list: listKeys }
}))

vi.mock('@/api/imageGenerations', () => ({
  imageGenerationsAPI: {
    pricing: getPricing,
    create: createImage,
    list: listImages,
    getImageUrl: vi.fn((id: number, index: number) => `/preview/${id}/${index}`),
    getImageBlob: vi.fn(),
    delete: vi.fn(),
    promptVersions: vi.fn()
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

function group(id: number, tiers: Array<'1K' | '2K' | '4K'>): Group {
  return {
    id,
    name: `group-${id}`,
    description: null,
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: true,
    image_allowed_tiers: tiers,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: 0.02,
    image_price_2k: 0.05,
    image_price_4k: 0.1,
    peak_rate_enabled: false,
    peak_start: '',
    peak_end: '',
    peak_rate_multiplier: 1,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  }
}

function apiKey(id: number, tiers: Array<'1K' | '2K' | '4K'>): ApiKey {
  return {
    id,
    user_id: 1,
    key: `sk-test-${id}`,
    name: `key-${id}`,
    group_id: id,
    group: group(id, tiers),
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null
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

async function mountWorkspace(keys: ApiKey[]) {
  listKeys.mockResolvedValue({ items: keys, total: keys.length, page: 1, page_size: 100, pages: 1 })
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
    vi.clearAllMocks()
  })

  it('keeps standard, 2K, and 4K keys in their permitted pools', async () => {
    const { workspace } = await mountWorkspace([
      apiKey(1, ['1K']),
      apiKey(2, ['2K']),
      apiKey(3, ['4K']),
      apiKey(4, ['2K', '4K']),
      apiKey(5, ['1K', '2K', '4K'])
    ])

    expect(workspace.standardKeys.value.map((key) => key.id)).toEqual([1])
    workspace.form.resolution_tier = '2K'
    await nextTick()
    expect(workspace.hdKeys.value.map((key) => key.id)).toEqual([2, 4, 5])
    workspace.form.resolution_tier = '4K'
    await nextTick()
    expect(workspace.hdKeys.value.map((key) => key.id)).toEqual([3, 4, 5])
  })

  it('uses the backend fixed price list for estimates', async () => {
    const { workspace } = await mountWorkspace([apiKey(2, ['2K'])])
    workspace.form.resolution_tier = '2K'
    workspace.form.n = 3
    await nextTick()

    expect(workspace.tierPrices.value).toEqual({ '1K': 0.06, '2K': 0.16, '4K': 0.2 })
    expect(workspace.estimatedUnitPrice.value).toBe(0.16)
    expect(workspace.estimatedTotalPrice.value).toBeCloseTo(0.48)
  })

  it('switches to the matching key pool and disables unsupported HD generation', async () => {
    const { workspace } = await mountWorkspace([apiKey(1, ['1K'])])
    workspace.form.prompt = 'a quiet coast'
    expect(workspace.canGenerate.value).toBe(true)

    workspace.form.resolution_tier = '4K'
    await nextTick()
    expect(workspace.hdKeys.value).toEqual([])
    expect(workspace.canGenerate.value).toBe(false)
  })

  it('sends the selected tier, aspect ratio, dimensions, and key id', async () => {
    createImage.mockResolvedValue(record())
    const { workspace } = await mountWorkspace([apiKey(2, ['2K'])])
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

  it('uses a manually entered existing key instead of the selected key id', async () => {
    createImage.mockResolvedValue(record({ api_key_id: 2 }))
    const { workspace } = await mountWorkspace([apiKey(1, ['1K']), apiKey(2, ['1K'])])
    workspace.form.prompt = 'coast at sunset'
    workspace.manualApiKey.value = 'sk-test-2'

    await workspace.generate()

    expect(createImage).toHaveBeenCalledWith(expect.objectContaining({
      api_key: 'sk-test-2',
      api_key_id: undefined
    }))
  })

  it('keeps queue snapshots independent from later form changes', async () => {
    const { workspace } = await mountWorkspace([apiKey(2, ['2K']), apiKey(3, ['4K'])])
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
