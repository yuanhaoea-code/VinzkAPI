<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ImageGenerationRecord } from '@/api/imageGenerations'
import type { ImageGenerationResultMetadata } from './types'

const props = defineProps<{
  record: ImageGenerationRecord | null
  activeGeneration: ImageGenerationResultMetadata | null
  generating: boolean
  errorMessage: string
  previewUrl: (recordId: number, imageIndex: number, fallback?: string) => string
}>()

const emit = defineEmits<{
  view: [record: ImageGenerationRecord, imageIndex: number]
  download: [record: ImageGenerationRecord, imageIndex: number]
  delete: [record: ImageGenerationRecord]
}>()

const { t } = useI18n()
const primaryImage = computed(() => props.record?.images?.[0] || null)
const secondaryImages = computed(() => props.record?.images?.slice(1) || [])
const displayMetadata = computed(() => {
  if ((props.generating || props.errorMessage) && props.activeGeneration) {
    return props.activeGeneration
  }
  if (!props.record) return null
  return {
    model: props.record.model,
    resolution_tier: props.record.resolution_tier || props.record.size,
    size: props.record.size
  }
})

function formatBytes(bytes: number): string {
  if (!bytes) return '0 KB'
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

function formatDate(value?: string): string {
  if (!value) return ''
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <section class="image-result">
    <header class="result-heading">
      <div>
        <h2>{{ t('imageGeneration.result') }}</h2>
        <p v-if="record?.status === 'success'">{{ t('imageGeneration.completedCount', { count: record.image_count }) }}</p>
        <p v-else>{{ t('imageGeneration.resultWaiting') }}</p>
      </div>
      <div v-if="displayMetadata" class="result-tags">
        <span>{{ displayMetadata.model }}</span>
        <span class="accent">{{ displayMetadata.resolution_tier }} · {{ displayMetadata.size }}</span>
      </div>
    </header>

    <div class="result-stage">
      <div v-if="generating" class="result-state">
        <span class="generation-spinner"><Icon name="sparkles" size="md" /></span>
        <strong>{{ t('imageGeneration.generating') }}</strong>
        <p>{{ t('imageGeneration.generatingHint') }}</p>
      </div>

      <div v-else-if="record?.status === 'failed' || errorMessage" class="result-state result-state--error">
        <Icon name="exclamationTriangle" size="lg" />
        <strong>{{ t('imageGeneration.failed') }}</strong>
        <p>{{ record?.error_message || errorMessage }}</p>
      </div>

      <div v-else-if="record && primaryImage" class="result-canvas">
        <img
          :src="previewUrl(record.id, primaryImage.index, primaryImage.url)"
          :alt="record.prompt"
          @click="emit('view', record, primaryImage.index)"
        />
        <div class="canvas-badges">
          <span>{{ t('imageGeneration.completed') }}</span>
          <span class="dark">{{ record.resolution_tier || record.size }}</span>
        </div>
        <div class="canvas-actions">
          <button type="button" :title="t('imageGeneration.preview')" @click="emit('view', record, primaryImage.index)"><Icon name="eye" size="sm" /></button>
          <button type="button" :title="t('imageGeneration.download')" @click="emit('download', record, primaryImage.index)"><Icon name="download" size="sm" /></button>
          <button type="button" :title="t('common.delete')" @click="emit('delete', record)"><Icon name="trash" size="sm" /></button>
        </div>
        <p class="canvas-prompt">{{ record.prompt }}</p>
      </div>

      <div v-else class="result-state">
        <span class="empty-symbol"><Icon name="sparkles" size="lg" /></span>
        <strong>{{ t('imageGeneration.waitingPrompt') }}</strong>
        <p>{{ t('imageGeneration.waitingPromptHint') }}</p>
      </div>

      <div v-if="record && primaryImage" class="result-meta">
        <span>{{ record.output_format.toUpperCase() }} · {{ formatBytes(record.file_size_bytes) }} · {{ formatDate(record.completed_at || record.created_at) }}</span>
        <strong>{{ record.image_count }} {{ t('imageGeneration.imageUnit') }}</strong>
      </div>

      <div v-if="record && secondaryImages.length" class="secondary-results">
        <button
          v-for="image in secondaryImages"
          :key="image.index"
          type="button"
          @click="emit('view', record, image.index)"
        >
          <img :src="previewUrl(record.id, image.index, image.url)" :alt="`${record.prompt} ${image.index + 1}`" />
          <span>{{ image.index + 1 }}</span>
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.image-result { min-width: 0; min-height: 0; display: flex; flex-direction: column; background: var(--ig-panel); }
.result-heading { min-height: 61px; padding: 14px 17px 10px; border-bottom: 1px solid var(--ig-line); display: flex; align-items: flex-start; justify-content: space-between; gap: 14px; }
.result-heading h2 { margin: 0; color: var(--ig-ink); font-size: 16px; font-weight: 650; }
.result-heading p { margin: 3px 0 0; color: var(--ig-muted); font-size: 13px; }
.result-tags { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 5px; }
.result-tags span { padding: 4px 8px; border: 1px solid var(--ig-line); border-radius: 999px; color: var(--ig-muted); background: var(--ig-panel-soft); font-size: 12px; }
.result-tags .accent { color: #2d655c; border-color: rgba(57, 125, 117, .22); background: var(--ig-teal-soft); }
.result-stage { flex: 1; min-height: 0; padding: 16px; display: flex; flex-direction: column; }
.result-canvas { position: relative; flex: 1; min-height: 440px; overflow: hidden; border: 1px solid var(--ig-line); border-radius: 7px; background: var(--ig-control); }
.result-canvas > img { width: 100%; height: 100%; object-fit: contain; display: block; cursor: zoom-in; }
.result-canvas::after { content: ''; position: absolute; inset: 55% 0 0; pointer-events: none; background: linear-gradient(180deg, transparent, rgba(12, 20, 20, .58)); }
.canvas-badges { position: absolute; z-index: 2; top: 10px; left: 10px; display: flex; gap: 6px; }
.canvas-badges span { padding: 5px 8px; border-radius: 999px; color: #fff; background: rgba(32, 80, 71, .86); backdrop-filter: blur(8px); font-size: 12px; font-weight: 700; }
.canvas-badges .dark { background: rgba(20, 23, 22, .62); }
.canvas-actions { position: absolute; z-index: 3; top: 10px; right: 10px; display: flex; gap: 5px; }
.canvas-actions button { width: 31px; height: 31px; padding: 0; border: 1px solid rgba(255, 255, 255, .38); border-radius: 6px; color: #fff; background: rgba(20, 23, 22, .56); display: grid; place-items: center; }
.canvas-prompt { position: absolute; z-index: 2; left: 14px; right: 14px; bottom: 12px; margin: 0; color: rgba(255, 255, 255, .96); font-size: 13px; line-height: 1.55; }
.result-state { flex: 1; min-height: 440px; border: 1px dashed var(--ig-line-strong); border-radius: 7px; background: var(--ig-control-faint); display: flex; flex-direction: column; align-items: center; justify-content: center; color: var(--ig-muted); text-align: center; }
.result-state strong { margin-top: 12px; color: var(--ig-ink); font-size: 15px; }
.result-state p { max-width: 400px; margin: 6px 20px 0; font-size: 14px; line-height: 1.6; }
.result-state--error { color: #b65d50; border-color: rgba(182, 93, 80, .3); background: rgba(182, 93, 80, .06); }
.generation-spinner, .empty-symbol { width: 44px; height: 44px; border: 1px solid var(--ig-line); border-radius: 50%; display: grid; place-items: center; color: var(--ig-teal); background: var(--ig-panel); }
.generation-spinner { animation: pulse 1.6s ease-in-out infinite; }
.result-meta { min-height: 41px; padding-top: 11px; display: flex; align-items: center; justify-content: space-between; gap: 12px; color: var(--ig-muted); font-size: 13px; }
.result-meta strong { color: var(--ig-ink-soft); }
.secondary-results { display: grid; grid-template-columns: repeat(4, 72px); gap: 7px; padding-top: 8px; border-top: 1px solid var(--ig-line); }
.secondary-results button { position: relative; width: 72px; height: 58px; overflow: hidden; padding: 0; border: 1px solid var(--ig-line); border-radius: 6px; background: var(--ig-control); }
.secondary-results img { width: 100%; height: 100%; object-fit: cover; }
.secondary-results span { position: absolute; right: 4px; bottom: 4px; width: 21px; height: 21px; border-radius: 50%; color: #fff; background: rgba(0, 0, 0, .58); display: grid; place-items: center; font-size: 11px; }
@keyframes pulse { 0%, 100% { transform: scale(1); opacity: .72; } 50% { transform: scale(1.08); opacity: 1; } }

@media (max-width: 820px) {
  .result-canvas, .result-state { min-height: 360px; }
  .result-heading { flex-direction: column; }
  .result-tags { justify-content: flex-start; }
}
</style>
