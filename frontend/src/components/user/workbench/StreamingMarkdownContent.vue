<script setup lang="ts">
import { onBeforeUnmount, shallowRef, watch } from 'vue'
import MarkdownContent from './MarkdownContent.vue'

const props = defineProps<{
  content: string
  streaming: boolean
}>()

const emit = defineEmits<{
  rendered: []
  typingChange: [typing: boolean]
}>()

const displayedContent = shallowRef('')
const typing = shallowRef(false)
const prefersReducedMotion = typeof window !== 'undefined'
  && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
let targetCharacters: string[] = []
let cursor = 0
let timer: number | undefined
let initialized = false

function setTyping(value: boolean): void {
  if (typing.value === value) return
  typing.value = value
  emit('typingChange', value)
}

function clearTimer(): void {
  window.clearTimeout(timer)
  timer = undefined
}

function flushTarget(): void {
  clearTimer()
  cursor = targetCharacters.length
  displayedContent.value = targetCharacters.join('')
  setTyping(false)
  emit('rendered')
}

function scheduleNextCharacter(): void {
  if (timer !== undefined || cursor >= targetCharacters.length) return
  timer = window.setTimeout(revealNextCharacter, 16)
}

function revealNextCharacter(): void {
  timer = undefined
  if (cursor >= targetCharacters.length) {
    setTyping(false)
    return
  }
  displayedContent.value += targetCharacters[cursor]
  cursor += 1
  emit('rendered')
  if (cursor < targetCharacters.length) scheduleNextCharacter()
  else setTyping(false)
}

watch(
  () => [props.content, props.streaming] as const,
  ([content, streaming]) => {
    const initialRender = !initialized
    initialized = true
    targetCharacters = Array.from(content)

    if (!content.startsWith(displayedContent.value)) {
      clearTimer()
      displayedContent.value = ''
      cursor = 0
    } else {
      cursor = Array.from(displayedContent.value).length
    }

    if (prefersReducedMotion || (initialRender && !streaming)) {
      flushTarget()
      return
    }
    if (cursor < targetCharacters.length) {
      setTyping(true)
      scheduleNextCharacter()
    } else {
      setTyping(false)
    }
  },
  { immediate: true }
)

onBeforeUnmount(clearTimer)
</script>

<template>
  <div
    class="streaming-markdown"
    :class="{ 'streaming-markdown--active': streaming || typing }"
    aria-live="off"
  >
    <MarkdownContent v-if="displayedContent" :content="displayedContent" />
  </div>
</template>

<style scoped>
.streaming-markdown {
  min-width: 0;
}

.streaming-markdown--active :deep(.markdown-content > :last-child)::after {
  display: inline-block;
  width: 5px;
  height: 15px;
  margin-left: 4px;
  vertical-align: -2px;
  background: #4f8f83;
  content: '';
  animation: stream-caret 0.8s steps(2, jump-none) infinite;
}

@keyframes stream-caret {
  0%, 45% { opacity: 1; }
  46%, 100% { opacity: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .streaming-markdown--active :deep(.markdown-content > :last-child)::after {
    animation: none;
  }
}
</style>
