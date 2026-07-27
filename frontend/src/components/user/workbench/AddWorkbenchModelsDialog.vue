<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { WorkbenchKey, WorkbenchModel } from '@/api/workbench'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { compareWorkbenchModels } from './workbenchModelSort'

const props = withDefaults(defineProps<{
  show: boolean
  apiKey: WorkbenchKey | null
  loading?: boolean
}>(), {
  loading: false
})

const emit = defineEmits<{
  close: []
  confirm: [modelIds: string[]]
}>()

const query = shallowRef('')
const selectedIds = shallowRef<string[]>([])

const sortedModels = computed<WorkbenchModel[]>(() => (
  [...(props.apiKey?.available_models ?? [])].sort(compareWorkbenchModels)
))
const filteredModels = computed<WorkbenchModel[]>(() => {
  const needle = query.value.trim().toLocaleLowerCase()
  if (!needle) return sortedModels.value
  return sortedModels.value.filter(model => (
    model.display_name.toLocaleLowerCase().includes(needle)
    || model.id.toLocaleLowerCase().includes(needle)
  ))
})
const addedCount = computed(() => sortedModels.value.filter(model => model.added).length)

watch(
  () => [props.show, props.apiKey?.api_key_id] as const,
  ([show]) => {
    if (!show) return
    query.value = ''
    selectedIds.value = []
  },
  { immediate: true }
)

function isChecked(model: WorkbenchModel): boolean {
  return model.added || selectedIds.value.includes(model.id)
}

function toggle(model: WorkbenchModel): void {
  if (model.added || props.loading) return
  selectedIds.value = selectedIds.value.includes(model.id)
    ? selectedIds.value.filter(id => id !== model.id)
    : [...selectedIds.value, model.id]
}

function confirm(): void {
  if (selectedIds.value.length === 0 || props.loading) return
  emit('confirm', selectedIds.value)
}
</script>

<template>
  <BaseDialog
    :show="show"
    title="添加模型"
    width="normal"
    :close-on-click-outside="true"
    :close-on-escape="!loading"
    :show-close-button="!loading"
    @close="emit('close')"
  >
    <div class="catalog-dialog">
      <div class="catalog-key">
        <span class="provider-dot" :data-provider="apiKey?.platform || 'none'" />
        <span class="catalog-key__copy">
          <strong>{{ apiKey?.key_name || '未选择密钥' }}</strong>
          <small>{{ apiKey?.provider_label }} · {{ apiKey?.group_name }}</small>
        </span>
        <span class="catalog-key__count">已添加 {{ addedCount }} / {{ sortedModels.length }}</span>
      </div>

      <label class="catalog-search">
        <Icon name="search" size="sm" />
        <input v-model="query" type="search" placeholder="搜索模型名称或 ID" autocomplete="off">
      </label>

      <div class="catalog-list" aria-label="当前密钥支持的对话模型">
        <label
          v-for="model in filteredModels"
          :key="model.id"
          class="catalog-row"
          :class="{
            'catalog-row--checked': isChecked(model),
            'catalog-row--added': model.added
          }"
        >
          <input
            type="checkbox"
            :checked="isChecked(model)"
            :disabled="model.added || loading"
            @change="toggle(model)"
          >
          <span class="catalog-check" aria-hidden="true">
            <Icon v-if="isChecked(model)" name="check" size="xs" :stroke-width="2.2" />
          </span>
          <span class="catalog-row__copy">
            <strong>{{ model.display_name }}</strong>
            <small>{{ model.id }}</small>
          </span>
          <span v-if="model.added" class="catalog-added">已添加</span>
        </label>

        <p v-if="filteredModels.length === 0" class="catalog-empty">
          {{ sortedModels.length === 0 ? '当前密钥暂无可用于对话的模型' : '没有匹配的模型' }}
        </p>
      </div>
    </div>

    <template #footer>
      <div class="catalog-actions">
        <button type="button" class="catalog-button catalog-button--secondary" :disabled="loading" @click="emit('close')">
          取消
        </button>
        <button
          type="button"
          class="catalog-button catalog-button--primary"
          :disabled="selectedIds.length === 0 || loading"
          @click="confirm"
        >
          <span v-if="loading" class="catalog-spinner" />
          {{ loading ? '正在添加' : selectedIds.length > 0 ? `添加 ${selectedIds.length} 个模型` : '确认添加' }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<style scoped>
.catalog-dialog {
  display: grid;
  gap: 12px;
}

.catalog-key {
  display: grid;
  min-height: 52px;
  grid-template-columns: 8px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid rgba(23, 20, 17, 0.09);
  border-radius: 6px;
  background: #f5f4ef;
}

.provider-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #8f867b;
}

.provider-dot[data-provider='openai'] { background: #3f9483; }
.provider-dot[data-provider='anthropic'],
.provider-dot[data-provider='antigravity'] { background: #bb6847; }
.provider-dot[data-provider='gemini'] { background: #4f79bc; }
.provider-dot[data-provider='grok'] { background: #826aa4; }

.catalog-key__copy,
.catalog-row__copy {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.catalog-key__copy strong { color: #211e1a; font-size: 12px; }
.catalog-key__copy small,
.catalog-row__copy small { color: #8d8378; font-size: 10px; }
.catalog-key__count { color: #776e64; font-size: 10px; white-space: nowrap; }

.catalog-search {
  display: grid;
  height: 38px;
  grid-template-columns: 16px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  padding: 0 11px;
  border: 1px solid rgba(23, 20, 17, 0.12);
  border-radius: 6px;
  background: #fffdf9;
  color: #8d8378;
  transition: border-color 150ms ease, box-shadow 150ms ease;
}

.catalog-search:focus-within {
  border-color: #729b92;
  box-shadow: 0 0 0 3px rgba(114, 155, 146, 0.13);
}

.catalog-search input {
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #211e1a;
  font-size: 12px;
}

.catalog-search input::placeholder { color: #aaa096; }

.catalog-list {
  max-height: min(390px, calc(100vh - 300px));
  min-height: 176px;
  overflow-y: auto;
  overscroll-behavior: contain;
  border: 1px solid rgba(23, 20, 17, 0.09);
  border-radius: 6px;
  background: #fffdf9;
}

.catalog-row {
  position: relative;
  display: grid;
  min-height: 53px;
  grid-template-columns: 18px minmax(0, 1fr) auto;
  align-items: center;
  gap: 9px;
  padding: 7px 10px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.06);
  cursor: pointer;
  transition: background 130ms ease;
}

.catalog-row:last-child { border-bottom: 0; }
.catalog-row:hover { background: #f3f5f1; }
.catalog-row--checked { background: #eaf1ef; }
.catalog-row--added { cursor: default; }
.catalog-row--added:hover { background: #eaf1ef; }
.catalog-row input { position: absolute; width: 1px; height: 1px; opacity: 0; }

.catalog-check {
  display: inline-flex;
  width: 17px;
  height: 17px;
  align-items: center;
  justify-content: center;
  border: 1px solid #b8afa5;
  border-radius: 4px;
  color: #fff;
}

.catalog-row--checked .catalog-check {
  border-color: #397b6e;
  background: #397b6e;
}

.catalog-row__copy strong {
  overflow: hidden;
  color: #211e1a;
  font-size: 12px;
  font-weight: 680;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.catalog-row__copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.catalog-added {
  padding: 2px 6px;
  border: 1px solid rgba(57, 123, 110, 0.2);
  border-radius: 4px;
  color: #397b6e;
  font-size: 9px;
}

.catalog-empty {
  display: grid;
  min-height: 176px;
  margin: 0;
  place-items: center;
  color: #8d8378;
  font-size: 11px;
}

.catalog-actions { display: flex; justify-content: flex-end; gap: 8px; }

.catalog-button {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 0 14px;
  border: 1px solid rgba(23, 20, 17, 0.13);
  border-radius: 5px;
  font-size: 11px;
  font-weight: 650;
}

.catalog-button--secondary { background: #fffdf9; color: #5f584f; }
.catalog-button--primary { border-color: #171411; background: #171411; color: #fffdf8; }
.catalog-button:disabled { cursor: not-allowed; opacity: 0.48; }

.catalog-spinner {
  width: 12px;
  height: 12px;
  border: 1.5px solid rgba(255, 255, 255, 0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: catalog-spin 700ms linear infinite;
}

@keyframes catalog-spin { to { transform: rotate(360deg); } }

:global(.dark .catalog-key),
:global(.dark .catalog-list),
:global(.dark .catalog-search) {
  border-color: rgba(148, 163, 184, 0.16);
  background: #111827;
}

:global(.dark .catalog-key__copy strong),
:global(.dark .catalog-row__copy strong),
:global(.dark .catalog-search input) { color: #f8fafc; }
:global(.dark .catalog-row) { border-color: rgba(148, 163, 184, 0.1); }
:global(.dark .catalog-row:hover) { background: rgba(148, 163, 184, 0.08); }
:global(.dark .catalog-row--checked) { background: rgba(20, 184, 166, 0.12); }
:global(.dark .catalog-button--secondary) { border-color: rgba(148, 163, 184, 0.18); background: #111827; color: #cbd5e1; }
:global(.dark .catalog-button--primary) { border-color: #14b8a6; background: #14b8a6; color: #06211d; }

@media (max-width: 540px) {
  .catalog-key { grid-template-columns: 8px minmax(0, 1fr); }
  .catalog-key__count { display: none; }
  .catalog-list { max-height: calc(100dvh - 330px); }
  .catalog-actions { display: grid; width: 100%; grid-template-columns: 1fr 1fr; }
}

@media (prefers-reduced-motion: reduce) {
  .catalog-row,
  .catalog-search { transition: none; }
  .catalog-spinner { animation-duration: 1400ms; }
}
</style>
