<script setup lang="ts">
import type { WorkbenchKey, WorkbenchReasoningPreset } from '@/api/workbench'
import Icon from '@/components/icons/Icon.vue'
import ModelPicker from './ModelPicker.vue'

defineProps<{
  keys: WorkbenchKey[]
  bindingId: string
  reasoningPreset: WorkbenchReasoningPreset
  loading?: boolean
}>()

const emit = defineEmits<{
  toggleSidebar: []
  selectKey: [apiKeyId: number]
  selectModel: [bindingId: string]
  addModels: [modelIds: string[], apiKeyId: number]
  removeModel: [bindingId: string]
  selectReasoning: [preset: WorkbenchReasoningPreset]
}>()

const presets: Array<{ id: WorkbenchReasoningPreset; label: string }> = [
  { id: 'fast', label: '快速' },
  { id: 'standard', label: '标准' },
  { id: 'deep', label: '深入' }
]

</script>

<template>
  <header class="workbench-toolbar">
    <button type="button" class="toolbar-icon sidebar-toggle" title="会话列表" @click="emit('toggleSidebar')">
      <Icon name="menu" size="sm" />
    </button>
    <ModelPicker
      :keys="keys"
      :model-value="bindingId"
      :loading="loading"
      @select-key="emit('selectKey', $event)"
      @update:model-value="emit('selectModel', $event)"
      @add-models="(modelIds, apiKeyId) => emit('addModels', modelIds, apiKeyId)"
      @remove-model="emit('removeModel', $event)"
    />
    <div class="toolbar-separator" />
    <div class="reasoning-control" aria-label="思考程度">
      <button
        v-for="preset in presets"
        :key="preset.id"
        type="button"
        :class="{ active: reasoningPreset === preset.id }"
        :aria-pressed="reasoningPreset === preset.id"
        :disabled="loading"
        @click="emit('selectReasoning', preset.id)"
      >
        {{ preset.label }}
      </button>
    </div>
    <span class="toolbar-status">
      <span class="toolbar-status__dot" />
      已连接
    </span>
  </header>
</template>

<style scoped>
.workbench-toolbar {
  position: relative;
  z-index: 20;
  display: flex;
  min-height: 58px;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.08);
  background: rgba(251, 248, 242, 0.76);
  backdrop-filter: blur(12px);
}

.toolbar-icon {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.72);
  color: #5f584f;
}

.sidebar-toggle {
  display: none;
}

.toolbar-separator {
  width: 1px;
  height: 24px;
  margin: 0 2px;
  background: rgba(23, 20, 17, 0.08);
}

.reasoning-control {
  display: inline-grid;
  grid-template-columns: repeat(3, 1fr);
  padding: 2px;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.48);
}

.reasoning-control button {
  min-width: 49px;
  min-height: 29px;
  padding: 0 9px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #80766b;
  font-size: 11px;
  font-weight: 600;
}

.reasoning-control button.active {
  background: #171411;
  color: #fffdf8;
}

.toolbar-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  color: #8a8176;
  font-size: 10px;
}

.toolbar-status__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3e9c75;
  box-shadow: 0 0 0 3px rgba(62, 156, 117, 0.1);
}

:global(.dark .workbench-toolbar) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(11, 18, 32, 0.82);
}

:global(.dark .toolbar-icon),
:global(.dark .reasoning-control) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.78);
  color: #cbd5e1;
}

:global(.dark .reasoning-control button.active) {
  background: #14b8a6;
  color: #06211d;
}

@media (max-width: 900px) {
  .sidebar-toggle {
    display: inline-flex;
  }
}

@media (max-width: 600px) {
  .workbench-toolbar {
    flex-wrap: wrap;
    padding: 9px 10px;
  }

  .toolbar-separator,
  .toolbar-status {
    display: none;
  }

  .reasoning-control {
    order: 3;
    width: 100%;
  }
}
</style>
