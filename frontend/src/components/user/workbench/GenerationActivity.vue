<script setup lang="ts">
import { computed } from 'vue'
import type { WorkbenchGenerationStage, WorkbenchTask } from './types'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  tasks: WorkbenchTask[]
  active: boolean
  stage?: WorkbenchGenerationStage
}>()

const ariaLabel = computed(() => {
  if (props.stage === 'reasoning') return '模型正在组织思路'
  if (props.stage === 'answering') return '模型正在生成回答'
  return props.active ? '模型正在处理' : '任务进度'
})
</script>

<template>
  <div v-if="tasks.length || active" class="generation-activity" :class="{ 'generation-activity--active': active }" :aria-label="ariaLabel">
    <TransitionGroup name="task-step" tag="div" class="task-progress">
      <div v-for="task in tasks" :key="task.id" class="task-row" :class="`task-row--${task.status}`">
        <span class="task-indicator" :class="`task-indicator--${task.status}`">
          <Icon v-if="task.status === 'completed'" name="check" size="xs" />
          <Icon v-else-if="task.status === 'failed'" name="x" size="xs" />
          <span v-else-if="task.status === 'in_progress'" class="task-spinner" />
        </span>
        <span class="task-title">{{ task.title }}</span>
        <span v-if="task.status === 'in_progress'" class="task-dots" aria-hidden="true">
          <i /><i /><i />
        </span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.generation-activity {
  margin: 5px 0 10px;
}

.task-progress {
  position: relative;
  display: grid;
  min-width: 210px;
  max-width: 520px;
  gap: 7px;
  padding: 5px 0 5px 13px;
}

.task-progress::before {
  position: absolute;
  top: 3px;
  bottom: 3px;
  left: 0;
  width: 1px;
  background: rgba(79, 143, 131, 0.42);
  content: '';
}

.task-progress::after {
  position: absolute;
  top: 3px;
  left: -1px;
  width: 3px;
  height: 30px;
  border-radius: 4px;
  background: linear-gradient(180deg, transparent, #6aa89c, transparent);
  content: '';
  opacity: 0;
}

.generation-activity--active .task-progress::after {
  animation: activity-sweep 1.6s ease-in-out infinite;
  opacity: 1;
}

.task-row {
  display: grid;
  min-height: 18px;
  grid-template-columns: 15px auto minmax(18px, 1fr);
  align-items: center;
  gap: 7px;
  color: #7d756b;
  font-size: 11px;
  line-height: 1.35;
}

.task-row--in_progress { color: #365f57; }
.task-row--completed { color: #716a62; }
.task-row--failed { color: #b8534c; }

.task-indicator {
  display: inline-flex;
  width: 14px;
  height: 14px;
  align-items: center;
  justify-content: center;
  border: 1px solid #aaa198;
  border-radius: 50%;
  color: #fff;
}

.task-indicator--completed { border-color: #4f9672; background: #4f9672; }
.task-indicator--failed { border-color: #b8534c; background: #b8534c; }
.task-indicator--in_progress { border-color: rgba(79, 143, 131, 0.24); }

.task-spinner {
  width: 12px;
  height: 12px;
  border: 1.5px solid rgba(79, 143, 131, 0.22);
  border-top-color: #4f8f83;
  border-radius: 50%;
  animation: task-spin 0.72s linear infinite;
}

.task-title { white-space: nowrap; }

.task-dots {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.task-dots i {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #5d958a;
  animation: dot-wave 1s ease-in-out infinite;
}

.task-dots i:nth-child(2) { animation-delay: 0.14s; }
.task-dots i:nth-child(3) { animation-delay: 0.28s; }

.task-step-enter-active,
.task-step-leave-active { transition: opacity 220ms ease, transform 220ms ease; }
.task-step-enter-from { opacity: 0; transform: translateY(5px); }
.task-step-leave-to { opacity: 0; transform: translateY(-3px); }

:global(.dark .task-progress::before) { background: rgba(45, 212, 191, 0.34); }
:global(.dark .task-row) { color: #94a3b8; }
:global(.dark .task-row--in_progress) { color: #99f6e4; }

@keyframes task-spin { to { transform: rotate(360deg); } }
@keyframes dot-wave {
  0%, 60%, 100% { opacity: 0.25; transform: translateY(0); }
  30% { opacity: 1; transform: translateY(-2px); }
}
@keyframes activity-sweep {
  0% { transform: translateY(-34px); }
  100% { transform: translateY(calc(100% + 8px)); }
}

@media (prefers-reduced-motion: reduce) {
  .task-spinner,
  .task-dots i { animation: none; }
  .task-step-enter-active,
  .task-step-leave-active { transition: none; }
  .generation-activity--active .task-progress::after { animation: none; }
}
</style>
