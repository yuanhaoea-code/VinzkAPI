<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  summary?: string
  streaming: boolean
  answerStarted: boolean
  visible: boolean
}>()

const open = shallowRef(false)

watch(
  () => [props.streaming, props.summary, props.answerStarted, props.visible] as const,
  ([streaming, summary, answerStarted, visible], previous) => {
    if (!visible) {
      open.value = false
      return
    }
    if (answerStarted && !previous?.[2]) {
      open.value = false
      return
    }
    if (streaming && !answerStarted && (!previous || summary !== previous[1] || visible !== previous[3])) {
      open.value = true
    }
  },
  { immediate: true }
)
</script>

<template>
  <section v-if="visible || summary" class="reasoning-panel">
    <button
      type="button"
      class="reasoning-trigger"
      :aria-expanded="open"
      @click="open = !open"
    >
      <Icon name="brain" size="sm" />
      <span>思考过程</span>
      <span v-if="streaming && !answerStarted" class="reasoning-wait" aria-label="正在接收思考摘要">
        <i /><i /><i />
      </span>
      <Icon name="chevronDown" size="xs" class="reasoning-chevron" :class="{ open }" />
    </button>
    <Transition name="reasoning-expand">
      <div v-if="open" class="reasoning-content">
        <p v-if="summary">
          <span>{{ summary }}</span>
          <i v-if="streaming && !answerStarted" class="reasoning-cursor" aria-hidden="true" />
        </p>
        <p v-else class="reasoning-placeholder">
          正在接收模型可展示的思考摘要
          <span v-if="streaming && !answerStarted" class="reasoning-inline-dots" aria-hidden="true">
            <i /><i /><i />
          </span>
        </p>
      </div>
    </Transition>
  </section>
</template>

<style scoped>
.reasoning-panel {
  max-width: 640px;
  margin: 3px 0 11px;
  border-left: 2px solid rgba(81, 125, 116, 0.34);
  background: rgba(239, 245, 242, 0.46);
}

.reasoning-trigger {
  display: inline-grid;
  min-height: 32px;
  grid-template-columns: 16px auto minmax(0, 1fr) 14px;
  align-items: center;
  gap: 7px;
  border: 0;
  padding: 0 10px;
  background: transparent;
  color: #416d64;
  font-size: 11px;
  font-weight: 650;
}

.reasoning-chevron { transition: transform 160ms ease; }
.reasoning-chevron.open { transform: rotate(180deg); }

.reasoning-wait {
  display: inline-flex;
  gap: 3px;
}

.reasoning-wait i {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #6c9b91;
  animation: reasoning-dot 1s ease-in-out infinite;
}

.reasoning-wait i:nth-child(2) { animation-delay: 0.12s; }
.reasoning-wait i:nth-child(3) { animation-delay: 0.24s; }

.reasoning-content {
  padding: 0 12px 10px 33px;
  color: #70685e;
  font-size: 11px;
  line-height: 1.65;
}

.reasoning-content p { margin: 0; white-space: pre-wrap; }
.reasoning-placeholder { color: #958b80; }

.reasoning-cursor {
  display: inline-block;
  width: 5px;
  height: 12px;
  margin-left: 4px;
  vertical-align: -2px;
  background: #5f9187;
  animation: reasoning-cursor 0.8s steps(2, jump-none) infinite;
}

.reasoning-inline-dots {
  display: inline-flex;
  gap: 3px;
  margin-left: 5px;
}

.reasoning-inline-dots i {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #6c9b91;
  animation: reasoning-dot 1s ease-in-out infinite;
}

.reasoning-inline-dots i:nth-child(2) { animation-delay: 0.12s; }
.reasoning-inline-dots i:nth-child(3) { animation-delay: 0.24s; }

:global(.dark .reasoning-panel) { background: rgba(20, 184, 166, 0.07); }
:global(.dark .reasoning-trigger) { color: #99f6e4; }
:global(.dark .reasoning-content) { color: #94a3b8; }

.reasoning-expand-enter-active,
.reasoning-expand-leave-active {
  overflow: hidden;
  transition: max-height 240ms ease, opacity 180ms ease, transform 180ms ease;
}

.reasoning-expand-enter-active { max-height: 360px; }
.reasoning-expand-enter-from,
.reasoning-expand-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-4px);
}

.reasoning-expand-enter-to,
.reasoning-expand-leave-from { max-height: 360px; opacity: 1; transform: translateY(0); }

@keyframes reasoning-dot {
  0%, 60%, 100% { opacity: 0.3; transform: translateY(0); }
  30% { opacity: 1; transform: translateY(-2px); }
}

@keyframes reasoning-cursor {
  0%, 45% { opacity: 1; }
  46%, 100% { opacity: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .reasoning-wait i,
  .reasoning-inline-dots i,
  .reasoning-cursor { animation: none; }
  .reasoning-chevron,
  .reasoning-expand-enter-active,
  .reasoning-expand-leave-active { transition: none; }
}
</style>
