<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, shallowRef } from 'vue'
import type { WorkbenchKey, WorkbenchModel } from '@/api/workbench'
import Icon from '@/components/icons/Icon.vue'
import AddWorkbenchModelsDialog from './AddWorkbenchModelsDialog.vue'
import { compareWorkbenchModels } from './workbenchModelSort'

const props = withDefaults(defineProps<{
  keys?: WorkbenchKey[]
  modelValue: string
  loading?: boolean
}>(), {
  keys: () => [],
  loading: false
})

const emit = defineEmits<{
  'update:modelValue': [bindingId: string]
  selectKey: [apiKeyId: number]
  addModels: [modelIds: string[], apiKeyId: number]
  removeModel: [bindingId: string]
}>()

const rootRef = shallowRef<HTMLElement | null>(null)
const openMenu = shallowRef<'key' | 'model' | ''>('')
const addDialogOpen = shallowRef(false)
const selectedKey = computed<WorkbenchKey | null>(() => (
  props.keys.find(key => key.models.some(model => model.binding_id === props.modelValue))
  ?? props.keys[0]
  ?? null
))
const currentModel = computed<WorkbenchModel | null>(() => (
  selectedKey.value?.models.find(model => model.binding_id === props.modelValue)
  ?? selectedKey.value?.models.find(model => model.id === selectedKey.value?.default_model_id)
  ?? null
))
const defaultModel = computed<WorkbenchModel | null>(() => (
  selectedKey.value?.models.find(model => model.id === selectedKey.value?.default_model_id) ?? null
))
const explicitModels = computed(() => (
  [...(selectedKey.value?.models.filter(model => model.id !== selectedKey.value?.default_model_id) ?? [])]
    .sort(compareWorkbenchModels)
))
const defaultSelected = computed(() => currentModel.value?.id === selectedKey.value?.default_model_id)

function toggleMenu(menu: 'key' | 'model'): void {
  if (props.loading) return
  openMenu.value = openMenu.value === menu ? '' : menu
}

function selectKey(key: WorkbenchKey): void {
  openMenu.value = ''
  emit('selectKey', key.api_key_id)
}

function selectModel(model: WorkbenchModel | null): void {
  if (!model?.binding_id || !model.available) return
  openMenu.value = ''
  emit('update:modelValue', model.binding_id)
}

function openAddDialog(): void {
  if (!selectedKey.value || props.loading) return
  openMenu.value = ''
  addDialogOpen.value = true
}

function addSelectedModels(modelIds: string[]): void {
  if (!selectedKey.value || modelIds.length === 0) return
  emit('addModels', modelIds, selectedKey.value.api_key_id)
  addDialogOpen.value = false
}

function removeModel(model: WorkbenchModel): void {
  if (!model.binding_id || props.loading) return
  openMenu.value = ''
  emit('removeModel', model.binding_id)
}

function onPointerDown(event: PointerEvent): void {
  if (!rootRef.value?.contains(event.target as Node)) openMenu.value = ''
}

onMounted(() => document.addEventListener('pointerdown', onPointerDown))
onBeforeUnmount(() => document.removeEventListener('pointerdown', onPointerDown))
</script>

<template>
  <div ref="rootRef" class="model-picker">
    <div class="picker-control picker-control--key">
      <button
        type="button"
        class="picker-trigger key-trigger"
        :class="{ 'picker-trigger--open': openMenu === 'key' }"
        :disabled="loading"
        aria-haspopup="listbox"
        :aria-expanded="openMenu === 'key'"
        @click="toggleMenu('key')"
      >
        <span class="provider-dot" :data-provider="selectedKey?.platform || 'none'" />
        <span class="picker-trigger__text">
          <small>密钥</small>
          <strong>{{ selectedKey?.key_name || '选择密钥' }}</strong>
        </span>
        <Icon name="chevronDown" size="xs" />
      </button>

      <div v-if="openMenu === 'key'" class="picker-menu picker-menu--keys" role="listbox" aria-label="选择密钥">
        <div class="picker-menu__head">
          <strong>选择密钥</strong>
          <span>{{ keys.length }} 个可用</span>
        </div>
        <button
          v-for="key in keys"
          :key="key.api_key_id"
          type="button"
          class="picker-option picker-option--key"
          :class="{ 'picker-option--selected': key.api_key_id === selectedKey?.api_key_id }"
          role="option"
          :aria-selected="key.api_key_id === selectedKey?.api_key_id"
          @click="selectKey(key)"
        >
          <Icon v-if="key.api_key_id === selectedKey?.api_key_id" name="check" size="sm" />
          <span v-else class="picker-option__spacer" />
          <span class="provider-dot" :data-provider="key.platform" />
          <span class="picker-option__body">
            <strong>{{ key.key_name }}</strong>
            <small>{{ key.provider_label }} · {{ key.group_name }}</small>
          </span>
        </button>
        <p v-if="keys.length === 0" class="picker-empty">暂无可用密钥，请先在 API Key 中创建。</p>
      </div>
    </div>

    <div class="picker-control picker-control--model">
      <button
        type="button"
        class="picker-trigger model-trigger"
        :class="{ 'picker-trigger--open': openMenu === 'model' }"
        :disabled="loading || !selectedKey"
        aria-haspopup="listbox"
        :aria-expanded="openMenu === 'model'"
        @click="toggleMenu('model')"
      >
        <span class="picker-trigger__text">
          <small>模型</small>
          <strong>{{ defaultSelected ? `Default · ${currentModel?.display_name || ''}` : currentModel?.display_name || '选择模型' }}</strong>
        </span>
        <Icon name="chevronDown" size="xs" />
      </button>

      <div v-if="openMenu === 'model'" class="picker-menu picker-menu--models model-menu" role="listbox" aria-label="选择模型">
        <div class="picker-menu__head">
          <strong>选择模型</strong>
          <span>按发布时间排序</span>
        </div>
        <button
          v-if="defaultModel"
          type="button"
          class="picker-option model-option__select picker-option--default"
          :class="{ 'picker-option--selected': defaultSelected }"
          role="option"
          :aria-selected="defaultSelected"
          @click="selectModel(defaultModel)"
        >
          <Icon v-if="defaultSelected" name="check" size="sm" />
          <span v-else class="picker-option__spacer" />
          <span class="picker-option__body">
            <strong>Default</strong>
            <small>{{ selectedKey?.default_model_name }}</small>
          </span>
          <span class="default-badge">推荐</span>
        </button>
        <div class="picker-menu__divider" />
        <div
          v-for="model in explicitModels"
          :key="model.binding_id || model.id"
          class="model-option"
          :class="{ 'model-option--selected': model.binding_id === modelValue }"
          role="option"
          :aria-selected="model.binding_id === modelValue"
        >
          <button
            type="button"
            class="model-option__select"
            :disabled="!model.available"
            @click="selectModel(model)"
          >
            <Icon v-if="model.binding_id === modelValue" name="check" size="sm" />
            <span v-else class="picker-option__spacer" />
            <span class="picker-option__body">
              <strong>{{ model.display_name }}</strong>
              <small>{{ model.id }}</small>
            </span>
          </button>
          <button
            type="button"
            class="model-option__remove"
            title="移除模型"
            aria-label="移除模型"
            :disabled="loading || !model.binding_id"
            @click.stop="removeModel(model)"
          >
            <Icon name="x" size="xs" />
          </button>
        </div>
        <div class="model-menu__footer">
          <button type="button" class="model-menu__add" @click="openAddDialog">
            <Icon name="plus" size="sm" />
            <span>添加模型</span>
            <small>{{ selectedKey?.available_models?.length ?? 0 }} 个可用</small>
          </button>
        </div>
      </div>
    </div>
  </div>

  <AddWorkbenchModelsDialog
    :show="addDialogOpen"
    :api-key="selectedKey"
    :loading="loading"
    @close="addDialogOpen = false"
    @confirm="addSelectedModels"
  />
</template>

<style scoped>
.model-picker {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.picker-control {
  position: relative;
  min-width: 0;
}

.picker-trigger {
  display: grid;
  width: 188px;
  min-height: 38px;
  grid-template-columns: minmax(0, 1fr) 14px;
  align-items: center;
  gap: 8px;
  padding: 5px 10px;
  border: 1px solid rgba(23, 20, 17, 0.12);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.78);
  color: #171411;
  transition: border-color 160ms ease, box-shadow 160ms ease, background 160ms ease;
}

.key-trigger {
  grid-template-columns: 8px minmax(0, 1fr) 14px;
  width: 178px;
}

.picker-trigger:hover:not(:disabled) {
  background: #fff;
  border-color: rgba(64, 114, 104, 0.42);
}

.picker-trigger--open,
.picker-trigger:focus-visible {
  border-color: #729b92;
  outline: 0;
  box-shadow: 0 0 0 3px rgba(114, 155, 146, 0.13);
}

.picker-trigger:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.picker-trigger__text {
  display: grid;
  min-width: 0;
  gap: 1px;
  text-align: left;
}

.picker-trigger__text small {
  color: #958b80;
  font-size: 8px;
  font-weight: 650;
  line-height: 1;
}

.picker-trigger__text strong {
  overflow: hidden;
  font-size: 11px;
  font-weight: 680;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-dot {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: #8f867b;
}

.provider-dot[data-provider='openai'] { background: #3f9483; }
.provider-dot[data-provider='anthropic'],
.provider-dot[data-provider='antigravity'] { background: #bb6847; }
.provider-dot[data-provider='gemini'] { background: #4f79bc; }
.provider-dot[data-provider='grok'] { background: #826aa4; }

.picker-menu {
  position: absolute;
  top: calc(100% + 7px);
  left: 0;
  z-index: 40;
  width: 258px;
  max-height: min(330px, calc(100vh - 150px));
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
  border: 1px solid rgba(23, 20, 17, 0.11);
  border-radius: 7px;
  background: #fffdf9;
  box-shadow: 0 16px 34px rgba(54, 45, 35, 0.15);
  animation: picker-menu-in 150ms ease-out both;
}

.picker-menu--models {
  width: 286px;
}

.picker-menu__head {
  position: sticky;
  top: 0;
  z-index: 1;
  display: flex;
  min-height: 39px;
  align-items: center;
  justify-content: space-between;
  padding: 0 11px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.07);
  background: rgba(255, 253, 249, 0.96);
  backdrop-filter: blur(8px);
}

.picker-menu__head strong { font-size: 11px; }
.picker-menu__head span { color: #968c80; font-size: 9px; }

.picker-option {
  display: grid;
  width: calc(100% - 10px);
  min-height: 48px;
  grid-template-columns: 16px minmax(0, 1fr);
  align-items: center;
  gap: 7px;
  margin: 3px 5px;
  padding: 6px 8px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: #211e1a;
  text-align: left;
  transition: background 130ms ease, transform 130ms ease;
}

.picker-option--key {
  grid-template-columns: 16px 7px minmax(0, 1fr);
}

.picker-option--default {
  grid-template-columns: 16px minmax(0, 1fr) auto;
}

.picker-option:hover:not(:disabled) {
  background: #f1f4f0;
  transform: translateX(1px);
}

.picker-option--selected {
  background: #e4efec;
  color: #173e37;
}

.picker-option:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.model-option {
  display: grid;
  width: calc(100% - 10px);
  min-height: 48px;
  grid-template-columns: minmax(0, 1fr) 30px;
  align-items: stretch;
  margin: 3px 5px;
  border-radius: 5px;
  transition: background 130ms ease, transform 130ms ease;
}

.model-option:hover {
  background: #f1f4f0;
  transform: translateX(1px);
}

.model-option--selected {
  background: #e4efec;
  color: #173e37;
}

.model-option__select {
  display: grid;
  min-width: 0;
  grid-template-columns: 16px minmax(0, 1fr);
  align-items: center;
  gap: 7px;
  padding: 6px 4px 6px 8px;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
}

.model-option__select:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.model-option__remove {
  display: grid;
  width: 26px;
  height: 26px;
  place-items: center;
  align-self: center;
  justify-self: center;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #a39a8f;
  opacity: 0.54;
  transition: color 120ms ease, background 120ms ease, opacity 120ms ease;
}

.model-option:hover .model-option__remove,
.model-option__remove:focus-visible {
  opacity: 1;
}

.model-option__remove:hover {
  background: rgba(181, 72, 61, 0.1);
  color: #a64238;
}

.model-option__remove:focus-visible {
  outline: 2px solid rgba(166, 66, 56, 0.28);
  outline-offset: 1px;
}

.picker-option__spacer { width: 14px; }

.picker-option__body {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.picker-option__body strong,
.picker-option__body small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.picker-option__body strong { font-size: 11px; font-weight: 680; }
.picker-option__body small { color: #8c8277; font-size: 9px; }

.default-badge {
  padding: 2px 5px;
  border: 1px solid rgba(63, 148, 131, 0.2);
  border-radius: 4px;
  color: #47786e;
  font-size: 8px;
}

.picker-menu__divider {
  height: 1px;
  margin: 4px 10px;
  background: rgba(23, 20, 17, 0.07);
}

.picker-empty {
  margin: 0;
  padding: 24px 14px;
  color: #8d8378;
  font-size: 10px;
  line-height: 1.6;
  text-align: center;
}

.model-menu__footer {
  position: sticky;
  bottom: 0;
  padding: 6px;
  border-top: 1px solid rgba(23, 20, 17, 0.07);
  background: rgba(255, 253, 249, 0.97);
  backdrop-filter: blur(8px);
}

.model-menu__add {
  display: grid;
  width: 100%;
  min-height: 36px;
  grid-template-columns: 16px minmax(0, 1fr) auto;
  align-items: center;
  gap: 7px;
  padding: 0 8px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: #315e56;
  text-align: left;
  transition: background 130ms ease, color 130ms ease;
}

.model-menu__add:hover {
  background: #e9f0ed;
  color: #173e37;
}

.model-menu__add span { font-size: 11px; font-weight: 680; }
.model-menu__add small { color: #8c8277; font-size: 9px; }

@keyframes picker-menu-in {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

:global(.dark .picker-trigger),
:global(.dark .picker-menu) {
  border-color: rgba(148, 163, 184, 0.18);
  background: #111827;
  color: #f8fafc;
}

:global(.dark .picker-menu__head) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(17, 24, 39, 0.96);
}

:global(.dark .model-menu__footer) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(17, 24, 39, 0.97);
}

:global(.dark .model-menu__add) { color: #99f6e4; }
:global(.dark .model-menu__add:hover) { background: rgba(20, 184, 166, 0.12); }

:global(.dark .picker-option) { color: #f8fafc; }
:global(.dark .picker-option:hover:not(:disabled)) { background: rgba(148, 163, 184, 0.09); }
:global(.dark .picker-option--selected) { background: rgba(20, 184, 166, 0.13); color: #ccfbf1; }
:global(.dark .model-option:hover) { background: rgba(148, 163, 184, 0.09); }
:global(.dark .model-option--selected) { background: rgba(20, 184, 166, 0.13); color: #ccfbf1; }
:global(.dark .model-option__remove) { color: #94a3b8; }
:global(.dark .model-option__remove:hover) { background: rgba(248, 113, 113, 0.12); color: #fca5a5; }

@media (max-width: 740px) {
  .model-picker {
    display: grid;
    min-width: 0;
    flex: 1;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.12fr);
  }

  .picker-trigger,
  .key-trigger {
    width: 100%;
  }

  .picker-menu {
    width: min(286px, calc(100vw - 24px));
  }

  .picker-menu--models {
    right: 0;
    left: auto;
  }
}

@media (prefers-reduced-motion: reduce) {
  .picker-menu { animation: none; }
  .picker-trigger,
  .picker-option,
  .model-option,
  .model-option__remove,
  .model-menu__add { transition: none; }
}
</style>
