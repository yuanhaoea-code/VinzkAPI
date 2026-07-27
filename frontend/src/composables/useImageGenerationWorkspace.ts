import { computed, effectScope, onMounted, reactive, ref, shallowRef, watch, type EffectScope } from 'vue'
import {
  imageGenerationsAPI,
  type ImageGenerationKeyCapability,
  type ImageGenerationImage,
  type ImageGenerationPromptVersion,
  type ImageGenerationRecord
} from '@/api/imageGenerations'
import { useAppStore } from '@/stores'
import type {
  ImageAspectRatio,
  ImageGenerationFormState,
  ImageGenerationMode,
  ImageGenerationQueueItem,
  ImageGenerationResultMetadata,
  ImageGenerationViewer,
  ImageKeyPool,
  ImageResolutionTier
} from '@/components/user/image-generation/types'

const tierSizes: Record<ImageResolutionTier, Record<ImageAspectRatio, string>> = {
  '1K': { '1:1': '1024x1024', '2:3': '1024x1536', '3:2': '1536x1024' },
  '2K': { '1:1': '2048x2048', '2:3': '1365x2048', '3:2': '2048x1365' },
  '4K': { '1:1': '2880x2880', '2:3': '1920x2880', '3:2': '2880x1920' }
}

const defaultTierPrices: Record<ImageResolutionTier, number> = {
  '1K': 0.06,
  '2K': 0.16,
  '4K': 0.2
}

const imagePreviewPlaceholder = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw=='

function safeInlinePreview(fallback?: string): string {
  const value = fallback?.trim() || ''
  return value.startsWith('data:image/') || value.startsWith('blob:')
    ? value
    : imagePreviewPlaceholder
}

function blobToDataURL(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('failed to read image blob'))
    reader.readAsDataURL(blob)
  })
}

function fileToDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('failed to read file'))
    reader.readAsDataURL(file)
  })
}

function createImageGenerationWorkspace() {
  const appStore = useAppStore()

  const form = reactive<ImageGenerationFormState>({
    model: 'gpt-image-2',
    prompt: '',
    resolution_tier: '1K',
    aspect_ratio: '1:1',
    quality: 'auto',
    output_format: 'png',
    n: 1,
    source_image: ''
  })
  const activeMode = shallowRef<ImageGenerationMode>('text')
  const generating = shallowRef(false)
  const loadingHistory = shallowRef(false)
  const loadingApiKeys = shallowRef(false)
  const errorMessage = shallowRef('')
  const manualApiKey = shallowRef('')
  const sourceImageName = shallowRef('')
  const currentRecord = shallowRef<ImageGenerationRecord | null>(null)
  const activeGeneration = shallowRef<ImageGenerationResultMetadata | null>(null)
  const history = ref<ImageGenerationRecord[]>([])
  const availableApiKeys = ref<ImageGenerationKeyCapability[]>([])
  const tierPrices = ref<Record<ImageResolutionTier, number>>({ ...defaultTierPrices })
  const selectedKeyByPool = reactive<Record<ImageKeyPool, string>>({ standard: '', hd: '' })
  const imageDataUrls = ref<Record<string, string>>({})
  const queueItems = ref<ImageGenerationQueueItem[]>([])
  const queueRunning = shallowRef(false)
  const queueAutoRunRequested = shallowRef(false)
  const viewer = shallowRef<ImageGenerationViewer | null>(null)
  const viewerUrl = shallowRef('')
  const showVersions = shallowRef(false)
  const versionLoading = shallowRef(false)
  const versionItems = ref<ImageGenerationPromptVersion[]>([])
  const deleteTarget = shallowRef<ImageGenerationRecord | null>(null)

  const activeKeys = computed(() => availableApiKeys.value.filter((key) => key.available))
  const keysByTier = computed<Record<ImageResolutionTier, ImageGenerationKeyCapability[]>>(() => ({
    '1K': activeKeys.value.filter((key) => key.allowed_tiers.includes('1K')),
    '2K': activeKeys.value.filter((key) => key.allowed_tiers.includes('2K')),
    '4K': activeKeys.value.filter((key) => key.allowed_tiers.includes('4K'))
  }))
  const standardKeys = computed(() => keysByTier.value['1K'])
  const hdKeys = computed(() => (
    form.resolution_tier === '4K' ? keysByTier.value['4K'] : keysByTier.value['2K']
  ))
  const currentPool = computed<ImageKeyPool>(() => form.resolution_tier === '1K' ? 'standard' : 'hd')
  const currentPoolKeys = computed(() => currentPool.value === 'standard' ? standardKeys.value : hdKeys.value)
  const selectedApiKeyId = computed({
    get: () => selectedKeyByPool[currentPool.value],
    set: (value: string) => { selectedKeyByPool[currentPool.value] = value }
  })
  const selectedApiKey = computed(() => {
    const selected = selectedApiKeyId.value
    return currentPoolKeys.value.find((key) => String(key.api_key_id) === selected) || currentPoolKeys.value[0]
  })
  const resolvedSize = computed(() => tierSizes[form.resolution_tier][form.aspect_ratio])
  const hasCurrentImages = computed(() => (
    !!currentRecord.value && Array.isArray(currentRecord.value.images) && currentRecord.value.images.length > 0
  ))
  const canGenerate = computed(() => {
    const hasPrompt = form.prompt.trim().length > 0
    const hasSource = activeMode.value !== 'image' || form.source_image.trim().length > 0
    const hasKey = manualApiKey.value.trim().length > 0 || currentPoolKeys.value.length > 0
    return hasPrompt && hasSource && hasKey && !loadingApiKeys.value && !generating.value
  })
  const estimatedUnitPrice = computed(() => tierPrices.value[form.resolution_tier])
  const estimatedTotalPrice = computed(() => (
    estimatedUnitPrice.value * Math.max(1, Number(form.n) || 1)
  ))

  function ensureKeySelections() {
    if (!standardKeys.value.some((key) => String(key.api_key_id) === selectedKeyByPool.standard)) {
      selectedKeyByPool.standard = standardKeys.value[0] ? String(standardKeys.value[0].api_key_id) : ''
    }
    if (!hdKeys.value.some((key) => String(key.api_key_id) === selectedKeyByPool.hd)) {
      selectedKeyByPool.hd = hdKeys.value[0] ? String(hdKeys.value[0].api_key_id) : ''
    }
  }

  watch([standardKeys, hdKeys], ensureKeySelections, { immediate: true })

  function imageKey(recordId: number, imageIndex: number): string {
    return `${recordId}:${imageIndex}`
  }

  function previewUrl(recordId: number, imageIndex: number, fallback?: string): string {
    return imageDataUrls.value[imageKey(recordId, imageIndex)] || safeInlinePreview(fallback)
  }

  async function ensurePreview(recordId: number, imageIndex: number): Promise<string> {
    const key = imageKey(recordId, imageIndex)
    if (imageDataUrls.value[key]) return imageDataUrls.value[key]
    try {
      const blob = await imageGenerationsAPI.getImageBlob(recordId, imageIndex, 'preview')
      const dataUrl = await blobToDataURL(blob)
      imageDataUrls.value = { ...imageDataUrls.value, [key]: dataUrl }
      return dataUrl
    } catch {
      imageDataUrls.value = { ...imageDataUrls.value, [key]: imagePreviewPlaceholder }
      return imagePreviewPlaceholder
    }
  }

  async function ensureRecordPreviews(record: ImageGenerationRecord, firstOnly = false) {
    const images = Array.isArray(record.images) ? (firstOnly ? record.images.slice(0, 1) : record.images) : []
    await Promise.all(images.map((image) => ensurePreview(record.id, image.index)))
  }

  async function loadApiKeys() {
    loadingApiKeys.value = true
    try {
      const result = await imageGenerationsAPI.capabilities()
      availableApiKeys.value = result.keys || []
      ensureKeySelections()
    } catch (error: unknown) {
      availableApiKeys.value = []
      appStore.showError(error instanceof Error ? error.message : '加载 API Key 失败')
    } finally {
      loadingApiKeys.value = false
    }
  }

  async function loadPricing() {
    try {
      const pricing = await imageGenerationsAPI.pricing()
      const next = { ...defaultTierPrices }
      for (const tier of ['1K', '2K', '4K'] as const) {
        const value = pricing.tiers?.[tier]
        if (typeof value === 'number' && Number.isFinite(value) && value >= 0) {
          next[tier] = value
        }
      }
      tierPrices.value = next
    } catch {
      tierPrices.value = { ...defaultTierPrices }
    }
  }

  async function loadHistory() {
    loadingHistory.value = true
    try {
      const result = await imageGenerationsAPI.list(1, 24)
      history.value = result.items || []
      await Promise.all(history.value.slice(0, 12).map((record) => (
        record.images?.length ? ensureRecordPreviews(record, true) : Promise.resolve()
      )))
      if (!currentRecord.value && history.value[0]) {
        currentRecord.value = history.value[0]
        await ensureRecordPreviews(history.value[0])
      }
    } catch (error: unknown) {
      appStore.showError(error instanceof Error ? error.message : '加载生图历史失败')
    } finally {
      loadingHistory.value = false
    }
  }

  function currentSnapshot(): ImageGenerationQueueItem {
    const manualKey = manualApiKey.value.trim()
    const key = manualKey ? undefined : selectedApiKey.value
    return {
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      mode: activeMode.value,
      model: form.model,
      prompt: form.prompt.trim(),
      resolution_tier: form.resolution_tier,
      aspect_ratio: form.aspect_ratio,
      size: resolvedSize.value,
      quality: form.quality,
      output_format: form.output_format,
      n: Math.max(1, Number(form.n) || 1),
      source_image: activeMode.value === 'image' ? form.source_image : '',
      source_image_name: sourceImageName.value,
      api_key_id: key?.api_key_id,
      api_key_name: key?.key_name,
      manual_api_key: manualKey || undefined,
      status: 'waiting'
    }
  }

  function validateSnapshot(item: ImageGenerationQueueItem): boolean {
    if (!item.prompt) return false
    if (item.mode === 'image' && !item.source_image) {
      appStore.showError('请先选择一张原图')
      return false
    }
    if (!item.api_key_id && !item.manual_api_key) {
      appStore.showError(item.resolution_tier === '1K' ? '当前没有可用的标准生图密钥' : '当前没有可用的高清生图密钥')
      return false
    }
    return true
  }

  async function executeGeneration(item: ImageGenerationQueueItem) {
    activeGeneration.value = {
      model: item.model,
      resolution_tier: item.resolution_tier,
      size: item.size
    }
    generating.value = true
    errorMessage.value = ''
    try {
      const record = await imageGenerationsAPI.create({
        model: item.model,
        prompt: item.prompt,
        size: item.size,
        resolution_tier: item.resolution_tier,
        aspect_ratio: item.aspect_ratio,
        quality: item.quality,
        output_format: item.output_format,
        n: item.n,
        api_key_id: item.api_key_id,
        api_key: item.manual_api_key,
        source_image: item.mode === 'image' ? item.source_image : undefined
      })
      currentRecord.value = record
      if (record.status !== 'success' || !record.images?.length) {
        const message = record.error_message || '生成失败，暂时没有返回可用图片'
        errorMessage.value = message
        appStore.showError(message)
        await loadHistory()
        return
      }
      await ensureRecordPreviews(record)
      appStore.showSuccess('图片生成完成')
      await loadHistory()
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : '生成失败，请稍后重试'
      errorMessage.value = message
      appStore.showError(message)
    } finally {
      generating.value = false
    }
  }

  async function generate() {
    if (generating.value || queueRunning.value) return
    const snapshot = currentSnapshot()
    if (!validateSnapshot(snapshot)) return
    await executeGeneration(snapshot)
    if (queueAutoRunRequested.value && !queueRunning.value && queueItems.value.length > 0) {
      queueAutoRunRequested.value = false
      await processQueue()
    }
  }

  function enqueueCurrent() {
    const snapshot = currentSnapshot()
    if (!validateSnapshot(snapshot)) return
    queueItems.value = [...queueItems.value, snapshot]
    appStore.showSuccess('已加入生成队列')
  }

  async function runQueue() {
    if (queueItems.value.length === 0) return
    if (generating.value || queueRunning.value) {
      queueAutoRunRequested.value = true
      return
    }
    await processQueue()
  }

  async function processQueue() {
    if (queueItems.value.length === 0) return
    queueRunning.value = true
    try {
      while (queueItems.value.length > 0) {
        const item = queueItems.value[0]
        if (!item) break
        item.status = 'running'
        await executeGeneration(item)
        queueItems.value = queueItems.value.slice(1)
      }
    } finally {
      queueRunning.value = false
      queueAutoRunRequested.value = false
    }
  }

  function removeQueueItem(index: number) {
    if (queueItems.value[index]?.status === 'running') return
    queueItems.value = queueItems.value.filter((_, itemIndex) => itemIndex !== index)
  }

  function resetCurrentForm() {
    form.model = 'gpt-image-2'
    form.prompt = ''
    form.resolution_tier = '1K'
    form.aspect_ratio = '1:1'
    form.quality = 'auto'
    form.output_format = 'png'
    form.n = 1
    form.source_image = ''
    activeMode.value = 'text'
    sourceImageName.value = ''
    manualApiKey.value = ''
  }

  async function setSourceImage(file?: File) {
    if (!file) {
      form.source_image = ''
      sourceImageName.value = ''
      return
    }
    sourceImageName.value = file.name
    form.source_image = await fileToDataURL(file)
  }

  function firstImage(record: ImageGenerationRecord): ImageGenerationImage | null {
    return record.images?.[0] || null
  }

  async function selectRecord(record: ImageGenerationRecord) {
    errorMessage.value = ''
    currentRecord.value = record
    await ensureRecordPreviews(record)
  }

  async function openViewer(record: ImageGenerationRecord, imageIndex: number) {
    viewer.value = { record, imageIndex }
    viewerUrl.value = await ensurePreview(record.id, imageIndex)
  }

  function closeViewer() {
    viewer.value = null
    viewerUrl.value = ''
  }

  async function downloadImage(record: ImageGenerationRecord, imageIndex: number) {
    try {
      const blob = await imageGenerationsAPI.getImageBlob(record.id, imageIndex, 'download')
      const url = URL.createObjectURL(blob)
      const image = record.images.find((item) => item.index === imageIndex)
      const ext = image?.mime_type?.split('/')[1] || 'png'
      const link = document.createElement('a')
      link.href = url
      link.download = `vinzk-image-${record.id}-${imageIndex}.${ext}`
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
    } catch (error: unknown) {
      appStore.showError(error instanceof Error ? error.message : '下载图片失败')
    }
  }

  async function openVersions(record: ImageGenerationRecord) {
    showVersions.value = true
    versionLoading.value = true
    versionItems.value = []
    try {
      const result = await imageGenerationsAPI.promptVersions(record.id)
      versionItems.value = result.items || []
    } catch (error: unknown) {
      appStore.showError(error instanceof Error ? error.message : '加载提示词版本失败')
    } finally {
      versionLoading.value = false
    }
  }

  function closeVersions() {
    showVersions.value = false
    versionItems.value = []
  }

  function confirmDelete(record: ImageGenerationRecord) {
    deleteTarget.value = record
  }

  async function deleteConfirmed() {
    const target = deleteTarget.value
    if (!target) return
    try {
      await imageGenerationsAPI.delete(target.id)
      if (currentRecord.value?.id === target.id) currentRecord.value = null
      deleteTarget.value = null
      appStore.showSuccess('生图记录已删除')
      await loadHistory()
    } catch (error: unknown) {
      appStore.showError(error instanceof Error ? error.message : '删除失败')
    }
  }

  return {
    form,
    activeMode,
    generating,
    loadingHistory,
    loadingApiKeys,
    errorMessage,
    manualApiKey,
    sourceImageName,
    currentRecord,
    activeGeneration,
    history,
    standardKeys,
    hdKeys,
    currentPool,
    selectedKeyByPool,
    selectedApiKey,
    resolvedSize,
    hasCurrentImages,
    canGenerate,
    estimatedUnitPrice,
    estimatedTotalPrice,
    tierPrices,
    queueItems,
    queueRunning,
    viewer,
    viewerUrl,
    showVersions,
    versionLoading,
    versionItems,
    deleteTarget,
    previewUrl,
    firstImage,
    loadApiKeys,
    loadPricing,
    loadHistory,
    generate,
    enqueueCurrent,
    runQueue,
    removeQueueItem,
    resetCurrentForm,
    setSourceImage,
    selectRecord,
    openViewer,
    closeViewer,
    downloadImage,
    openVersions,
    closeVersions,
    confirmDelete,
    deleteConfirmed
  }
}

export type ImageGenerationWorkspace = ReturnType<typeof createImageGenerationWorkspace>

let persistentWorkspace: ImageGenerationWorkspace | null = null
let persistentWorkspaceScope: EffectScope | null = null

export function useImageGenerationWorkspace(): ImageGenerationWorkspace {
  if (!persistentWorkspace) {
    persistentWorkspaceScope = effectScope(true)
    const workspace = persistentWorkspaceScope.run(createImageGenerationWorkspace)
    if (!workspace) {
      persistentWorkspaceScope.stop()
      persistentWorkspaceScope = null
      throw new Error('Failed to initialize image generation workspace')
    }
    persistentWorkspace = workspace
  }

  const workspace = persistentWorkspace
  onMounted(() => {
    void workspace.loadApiKeys()
    void workspace.loadPricing()
    void workspace.loadHistory()
  })
  return workspace
}

export function resetImageGenerationWorkspace(): void {
  persistentWorkspaceScope?.stop()
  persistentWorkspaceScope = null
  persistentWorkspace = null
}
