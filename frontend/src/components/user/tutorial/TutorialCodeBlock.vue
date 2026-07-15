<script setup lang="ts">
import { useClipboard } from '@/composables/useClipboard'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  code: string
  label: string
  language: string
}>()

const { copied, copyToClipboard } = useClipboard()

function copyCode() {
  copyToClipboard(props.code, '配置已复制')
}
</script>

<template>
  <div class="tutorial-code">
    <div class="tutorial-code__header">
      <div>
        <strong>{{ label }}</strong>
        <span>{{ language }}</span>
      </div>
      <button type="button" class="tutorial-code__copy" @click="copyCode">
        <Icon :name="copied ? 'check' : 'copy'" size="xs" :stroke-width="2" />
        {{ copied ? '已复制' : '复制' }}
      </button>
    </div>
    <pre><code>{{ code }}</code></pre>
  </div>
</template>

<style scoped>
.tutorial-code {
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.18);
  border-radius: 10px;
  background: #17202e;
  box-shadow: 0 14px 30px rgba(23, 32, 46, 0.1);
}

.tutorial-code__header {
  display: flex;
  min-height: 38px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 7px 10px 7px 14px;
  color: #94a3b8;
  background: #111827;
}

.tutorial-code__header > div {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: 8px;
}

.tutorial-code__header strong {
  overflow: hidden;
  color: #e2e8f0;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-code__header span {
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 10px;
  text-transform: uppercase;
}

.tutorial-code__copy {
  display: inline-flex;
  min-width: 58px;
  height: 26px;
  align-items: center;
  justify-content: center;
  gap: 5px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 7px;
  color: #a7f3d0;
  background: rgba(255, 255, 255, 0.06);
  font-size: 11px;
  transition: background 0.16s ease, border-color 0.16s ease;
}

.tutorial-code__copy:hover {
  border-color: rgba(94, 234, 212, 0.3);
  background: rgba(94, 234, 212, 0.1);
}

.tutorial-code pre {
  max-width: 100%;
  margin: 0;
  overflow-x: auto;
  padding: 15px 16px 17px;
  color: #dbeafe;
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 12px;
  line-height: 1.75;
  scrollbar-color: #475569 #17202e;
}

.tutorial-code code {
  font: inherit;
  white-space: pre;
}

@media (max-width: 640px) {
  .tutorial-code pre {
    padding: 13px 14px 15px;
    font-size: 11px;
  }
}
</style>
