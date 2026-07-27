<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'

export interface ConversationRailAttachment {
  id: string
  name: string
}

export interface ConversationRailItem {
  id: string
  title: string
  preview: string
  attachments?: ConversationRailAttachment[]
  attachmentOverflow?: number
  busy?: boolean
  position?: number
}

const props = defineProps<{
  items: ConversationRailItem[]
  activeId?: string
}>()

const emit = defineEmits<{
  navigate: [id: string]
}>()

const activeIndex = computed(() => {
  const index = props.items.findIndex(item => item.id === props.activeId)
  return index >= 0 ? index : Math.max(0, props.items.length - 1)
})

function markerPosition(item: ConversationRailItem, index: number): string {
  if (typeof item.position === 'number') {
    return `${Math.min(98, Math.max(2, item.position * 100))}%`
  }
  if (props.items.length <= 1) return '2%'
  return `${2 + (index / (props.items.length - 1)) * 96}%`
}

function fileName(value: string): string {
  return value.split(/[\\/]/).pop() || value
}
</script>

<template>
  <nav v-if="items.length" class="conversation-rail" aria-label="对话定位">
    <button
      v-for="(item, index) in items"
      :key="item.id"
      type="button"
      class="conversation-rail__mark"
      :class="{
        'conversation-rail__mark--active': index === activeIndex,
        'conversation-rail__mark--busy': item.busy,
        'conversation-rail__mark--first': index === 0,
        'conversation-rail__mark--last': items.length > 1 && index === items.length - 1
      }"
      :style="{ top: markerPosition(item, index) }"
      :aria-label="`跳转到第 ${index + 1} 轮：${item.title}`"
      :aria-describedby="`conversation-rail-preview-${item.id}`"
      :aria-current="index === activeIndex ? 'step' : undefined"
      @click="emit('navigate', item.id)"
    >
      <span class="conversation-rail__line" aria-hidden="true" />
      <span
        :id="`conversation-rail-preview-${item.id}`"
        class="conversation-rail__preview"
        role="tooltip"
      >
        <strong>{{ item.title }}</strong>
        <span class="conversation-rail__preview-copy">{{ item.preview }}</span>
        <span v-if="item.attachments?.length" class="conversation-rail__files">
          <span
            v-for="attachment in item.attachments"
            :key="attachment.id"
            class="conversation-rail__file"
          >
            <Icon name="document" size="xs" />
            <span>{{ fileName(attachment.name) }}</span>
          </span>
          <span v-if="item.attachmentOverflow" class="conversation-rail__file-more">
            +{{ item.attachmentOverflow }}
          </span>
        </span>
      </span>
    </button>
  </nav>
</template>

<style scoped>
.conversation-rail {
  position: absolute;
  top: 24px;
  bottom: 24px;
  left: 9px;
  z-index: 6;
  width: 34px;
}

.conversation-rail__mark {
  position: absolute;
  left: 0;
  z-index: 1;
  width: 32px;
  height: 18px;
  border: 0;
  padding: 0;
  background: transparent;
  color: #847d74;
  transform: translateY(-50%);
}

.conversation-rail__mark:hover,
.conversation-rail__mark:focus-visible {
  z-index: 3;
}

.conversation-rail__mark:focus-visible {
  outline: 0;
}

.conversation-rail__line {
  position: absolute;
  top: 8px;
  left: 3px;
  width: 8px;
  height: 1px;
  border-radius: 1px;
  background: currentColor;
  opacity: 0.52;
  transform-origin: left center;
  transition: width 150ms ease, height 150ms ease, opacity 150ms ease, background-color 150ms ease;
}

.conversation-rail__mark:hover .conversation-rail__line,
.conversation-rail__mark:focus-visible .conversation-rail__line {
  width: 19px;
  height: 2px;
  background: #477f75;
  opacity: 0.9;
}

.conversation-rail__mark--active .conversation-rail__line {
  width: 19px;
  height: 2px;
  background: #286c60;
  box-shadow: 0 0 0 3px rgba(79, 143, 131, 0.09);
  opacity: 1;
}

.conversation-rail__mark--busy .conversation-rail__line {
  animation: rail-breathe 1.4s ease-in-out infinite;
}

.conversation-rail__preview {
  position: absolute;
  top: 50%;
  left: 34px;
  display: grid;
  width: min(320px, calc(100vw - 92px));
  max-width: 320px;
  gap: 6px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  padding: 12px 13px 11px;
  background: #292724;
  box-shadow: 0 16px 36px rgba(28, 25, 22, 0.22);
  color: #f8f5ef;
  opacity: 0;
  pointer-events: none;
  text-align: left;
  transform: translate(8px, -50%) scale(0.985);
  transform-origin: left center;
  transition: opacity 130ms ease, transform 160ms ease;
  visibility: hidden;
}

.conversation-rail__mark:hover .conversation-rail__preview,
.conversation-rail__mark:focus-visible .conversation-rail__preview {
  opacity: 1;
  transform: translate(0, -50%) scale(1);
  visibility: visible;
}

.conversation-rail__mark--first .conversation-rail__preview {
  top: -4px;
  transform: translate(8px, 0) scale(0.985);
  transform-origin: left top;
}

.conversation-rail__mark--first:hover .conversation-rail__preview,
.conversation-rail__mark--first:focus-visible .conversation-rail__preview {
  transform: translate(0, 0) scale(1);
}

.conversation-rail__mark--last .conversation-rail__preview {
  top: auto;
  bottom: -4px;
  transform: translate(8px, 0) scale(0.985);
  transform-origin: left bottom;
}

.conversation-rail__mark--last:hover .conversation-rail__preview,
.conversation-rail__mark--last:focus-visible .conversation-rail__preview {
  transform: translate(0, 0) scale(1);
}

.conversation-rail__preview strong {
  overflow: hidden;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-rail__preview-copy {
  display: -webkit-box;
  overflow: hidden;
  color: #bdb8b0;
  font-size: 11px;
  line-height: 1.65;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.conversation-rail__files {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
  padding-top: 1px;
  color: #b8b3ab;
}

.conversation-rail__file {
  display: inline-flex;
  min-width: 0;
  max-width: 108px;
  align-items: center;
  gap: 5px;
  font-size: 10px;
}

.conversation-rail__file span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-rail__file-more {
  flex: 0 0 auto;
  color: #918b83;
  font-size: 10px;
}

:global(.dark .conversation-rail__mark) {
  color: #707984;
}

:global(.dark .conversation-rail__mark:hover .conversation-rail__line),
:global(.dark .conversation-rail__mark:focus-visible .conversation-rail__line),
:global(.dark .conversation-rail__mark--active .conversation-rail__line) {
  background: #5eead4;
}

:global(.dark .conversation-rail__preview) {
  border-color: rgba(148, 163, 184, 0.12);
  background: #25292e;
  box-shadow: 0 18px 42px rgba(2, 6, 23, 0.36);
}

@keyframes rail-breathe {
  0%, 100% { opacity: 0.42; transform: scaleX(0.78); }
  50% { opacity: 1; transform: scaleX(1); }
}

@media (prefers-reduced-motion: reduce) {
  .conversation-rail__line,
  .conversation-rail__preview { transition: none; }
  .conversation-rail__mark--busy .conversation-rail__line { animation: none; }
}

@media (max-width: 760px) {
  .conversation-rail { display: none; }
}
</style>
