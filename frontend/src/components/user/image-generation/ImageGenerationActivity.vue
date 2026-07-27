<script setup lang="ts">
import { shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ImageGenerationRecord } from '@/api/imageGenerations'
import type { ImageGenerationQueueItem } from './types'

defineProps<{
  queueItems: ImageGenerationQueueItem[]
  queueRunning: boolean
  history: ImageGenerationRecord[]
  loadingHistory: boolean
  currentRecordId?: number
  previewUrl: (recordId: number, imageIndex: number, fallback?: string) => string
}>()

const emit = defineEmits<{
  selectRecord: [record: ImageGenerationRecord]
  removeQueue: [index: number]
  versions: [record: ImageGenerationRecord]
  delete: [record: ImageGenerationRecord]
}>()

const { t } = useI18n()
const activeTab = shallowRef<'queue' | 'history'>('queue')
const recentGenerationLimit = 5

function formatDate(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <aside class="image-activity">
    <div class="activity-tabs" role="tablist">
      <button type="button" :class="{ active: activeTab === 'queue' }" @click="activeTab = 'queue'">
        {{ t('imageGeneration.queue') }}<span>{{ queueItems.length }}</span>
      </button>
      <button type="button" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">
        {{ t('imageGeneration.history') }}
      </button>
    </div>

    <div v-if="activeTab === 'queue'" class="activity-body">
      <div class="activity-heading">
        <strong>{{ queueRunning ? t('imageGeneration.processing') : t('imageGeneration.waitingQueue') }}</strong>
        <span>{{ queueItems.length }}</span>
      </div>
      <div v-if="queueItems.length" class="activity-list">
        <article v-for="(item, index) in queueItems" :key="item.id" class="queue-item">
          <div class="queue-row">
            <strong>{{ item.prompt }}</strong>
            <span :class="{ running: item.status === 'running' }">
              {{ item.status === 'running' ? t('imageGeneration.generating') : t('imageGeneration.waiting') }}
            </span>
          </div>
          <p>{{ item.resolution_tier }} · {{ item.aspect_ratio }} · {{ item.output_format.toUpperCase() }}</p>
          <div v-if="item.status === 'running'" class="queue-progress"><span /></div>
          <button v-else type="button" :title="t('imageGeneration.removeFromQueue')" @click="emit('removeQueue', index)">
            <Icon name="x" size="xs" />
          </button>
        </article>
      </div>
      <div v-else class="activity-empty">
        <Icon name="inbox" size="lg" />
        <p>{{ t('imageGeneration.emptyQueue') }}</p>
      </div>

      <div class="recent-label">{{ t('imageGeneration.recentGenerations') }}</div>
      <div class="history-list">
        <button
          v-for="record in history.slice(0, recentGenerationLimit)"
          :key="record.id"
          type="button"
          class="history-item"
          :class="{ selected: currentRecordId === record.id }"
          @click="emit('selectRecord', record)"
        >
          <span class="history-thumb">
            <img
              v-if="record.images?.[0]"
              :src="previewUrl(record.id, record.images[0].index, record.images[0].url)"
              :alt="record.prompt"
            />
            <Icon v-else name="sparkles" size="sm" />
          </span>
          <span class="history-copy">
            <strong>{{ record.prompt }}</strong>
            <small>{{ formatDate(record.created_at) }} · {{ record.model }}</small>
            <span><b>{{ record.resolution_tier || record.size }}</b><b>{{ record.aspect_ratio || '1:1' }}</b></span>
          </span>
        </button>
      </div>
    </div>

    <div v-else class="activity-body">
      <div class="activity-heading">
        <strong>{{ t('imageGeneration.history') }}</strong>
        <span>{{ history.length }}</span>
      </div>
      <div v-if="loadingHistory" class="activity-empty"><Icon name="refresh" size="md" class="spin" /></div>
      <div v-else-if="history.length" class="history-list history-list--full">
        <article v-for="record in history" :key="record.id" class="history-record" :class="{ selected: currentRecordId === record.id }">
          <button type="button" class="history-main" @click="emit('selectRecord', record)">
            <span class="history-thumb">
              <img v-if="record.images?.[0]" :src="previewUrl(record.id, record.images[0].index, record.images[0].url)" :alt="record.prompt" />
              <Icon v-else name="sparkles" size="sm" />
            </span>
            <span class="history-copy">
              <strong>{{ record.prompt }}</strong>
              <small>{{ formatDate(record.created_at) }} · {{ record.resolution_tier || record.size }}</small>
            </span>
          </button>
          <div class="history-actions">
            <button type="button" :title="t('imageGeneration.versions')" @click="emit('versions', record)"><Icon name="clock" size="xs" /></button>
            <button type="button" :title="t('common.delete')" @click="emit('delete', record)"><Icon name="trash" size="xs" /></button>
          </div>
        </article>
      </div>
      <div v-else class="activity-empty">
        <Icon name="clock" size="lg" />
        <p>{{ t('imageGeneration.emptyHistory') }}</p>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.image-activity { min-width: 0; min-height: 0; border-left: 1px solid var(--ig-line); background: var(--ig-panel-soft); overflow: hidden; display: flex; flex-direction: column; }
.activity-tabs { height: 54px; padding: 10px 11px 0; border-bottom: 1px solid var(--ig-line); display: grid; grid-template-columns: 1fr 1fr; gap: 4px; }
.activity-tabs button { height: 35px; border: 0; border-radius: 6px 6px 0 0; color: var(--ig-muted); background: transparent; font-size: 13px; }
.activity-tabs button.active { color: var(--ig-ink); background: var(--ig-panel); font-weight: 700; }
.activity-tabs span { margin-left: 4px; padding: 2px 5px; border-radius: 999px; color: var(--ig-teal); background: var(--ig-teal-soft); font-size: 11px; }
.activity-body { flex: 1; min-height: 0; padding: 12px; overflow-y: auto; }
.activity-heading { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 8px; }
.activity-heading strong { color: var(--ig-ink); font-size: 14px; }
.activity-heading span { color: var(--ig-muted); font-size: 12px; }
.activity-list, .history-list { display: grid; gap: 7px; }
.queue-item { position: relative; padding: 10px; border: 1px solid var(--ig-line); border-radius: 7px; background: var(--ig-panel); }
.queue-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 7px; }
.queue-row strong { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--ig-ink-soft); font-size: 13px; }
.queue-row > span { color: var(--ig-muted); font-size: 12px; }
.queue-row > span.running { color: #aa7730; }
.queue-item p { margin: 5px 24px 0 0; color: var(--ig-muted); font-size: 12px; }
.queue-item > button { position: absolute; right: 7px; bottom: 7px; width: 22px; height: 22px; border: 1px solid var(--ig-line); border-radius: 5px; color: var(--ig-muted); background: var(--ig-control); display: grid; place-items: center; }
.queue-progress { height: 3px; margin-top: 8px; overflow: hidden; border-radius: 3px; background: var(--ig-control); }
.queue-progress span { display: block; width: 64%; height: 100%; background: linear-gradient(90deg, var(--ig-teal), #70a8b7); animation: queue-progress 2.4s ease-in-out infinite; }
.activity-empty { min-height: 110px; border: 1px dashed var(--ig-line); border-radius: 7px; color: var(--ig-muted-light); display: flex; flex-direction: column; align-items: center; justify-content: center; }
.activity-empty p { margin: 8px 0 0; color: var(--ig-muted); font-size: 13px; }
.recent-label { margin: 15px 0 8px; color: var(--ig-muted); font-size: 13px; }
.history-item { width: 100%; min-width: 0; padding: 8px; border: 1px solid var(--ig-line); border-radius: 7px; color: inherit; background: var(--ig-panel); display: grid; grid-template-columns: 50px minmax(0, 1fr); gap: 9px; text-align: left; }
.history-item.selected, .history-record.selected { border-color: rgba(57, 125, 117, .38); background: var(--ig-teal-faint); }
.history-thumb { width: 50px; height: 50px; overflow: hidden; border-radius: 5px; color: var(--ig-muted); background: var(--ig-control); display: grid; place-items: center; }
.history-thumb img { width: 100%; height: 100%; object-fit: cover; }
.history-copy { min-width: 0; }
.history-copy > strong { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--ig-ink-soft); font-size: 13px; }
.history-copy > small { display: block; margin-top: 5px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--ig-muted); font-size: 12px; }
.history-copy > span { display: flex; gap: 4px; margin-top: 7px; }
.history-copy b { padding: 3px 5px; border-radius: 999px; color: var(--ig-teal); background: var(--ig-teal-soft); font-size: 11px; font-weight: 650; }
.history-list--full { gap: 8px; }
.history-record { position: relative; border: 1px solid var(--ig-line); border-radius: 7px; background: var(--ig-panel); }
.history-main { width: 100%; min-width: 0; padding: 8px 55px 8px 8px; border: 0; color: inherit; background: transparent; display: grid; grid-template-columns: 50px minmax(0, 1fr); gap: 9px; text-align: left; }
.history-actions { position: absolute; top: 8px; right: 8px; display: flex; gap: 4px; }
.history-actions button { width: 22px; height: 22px; border: 1px solid var(--ig-line); border-radius: 5px; color: var(--ig-muted); background: var(--ig-control); display: grid; place-items: center; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes queue-progress { 0% { transform: translateX(-70%); } 50% { transform: translateX(30%); } 100% { transform: translateX(120%); } }

@media (max-width: 1439px) and (min-width: 821px) {
  .image-activity { grid-column: 1 / -1; border-left: 0; border-top: 1px solid var(--ig-line); }
  .activity-body { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 12px; }
  .activity-heading, .recent-label { grid-column: 1 / -1; }
}

@media (max-width: 820px) {
  .image-activity { border-left: 0; border-top: 1px solid var(--ig-line); min-height: 420px; }
}
</style>
