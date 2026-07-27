<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, shallowRef, watch } from 'vue'
import type { WorkbenchMessage } from '@/api/workbench'
import type { WorkbenchGenerationStage, WorkbenchTask } from './types'
import Icon from '@/components/icons/Icon.vue'
import ConversationRail from './ConversationRail.vue'
import GenerationActivity from './GenerationActivity.vue'
import MessageAttachments from './MessageAttachments.vue'
import ReasoningPanel from './ReasoningPanel.vue'
import StreamingMarkdownContent from './StreamingMarkdownContent.vue'

const props = defineProps<{
  messages: WorkbenchMessage[]
  tasksByMessage: Record<string, WorkbenchTask[]>
  stageByMessage: Record<string, WorkbenchGenerationStage>
  conversationId?: string
  loading?: boolean
}>()

const scrollRef = shallowRef<HTMLElement | null>(null)
const railVisible = shallowRef(false)
const railPositions = shallowRef<Record<string, number>>({})
const activeTurnId = shallowRef('')
const copiedMessageId = shallowRef('')
const showScrollToBottom = shallowRef(false)
const typingByMessage = shallowRef<Record<string, boolean>>({})
let followLatest = true
let lastPositionedConversationId = ''
let copiedTimer: number | undefined
let followScrollTimer: number | undefined
let resizeObserver: ResizeObserver | undefined

const railItems = computed(() => props.messages
  .map((message, index) => ({ message, index }))
  .filter(({ message }) => message.role === 'user')
  .map(({ message, index }) => {
    const assistant = props.messages.slice(index + 1).find(item => item.role === 'assistant')
    const prompt = compactText(message.content)
    const answer = compactText(assistant?.content ?? '')
    const attachments = message.attachments ?? []
    return {
      id: message.id,
      title: truncatePreview(prompt || attachments[0]?.name || '附件消息', 28),
      preview: truncatePreview(
        answer || (assistant && isGenerating(assistant) ? '维枢AI 正在处理这条消息' : prompt || '查看这轮对话'),
        160
      ),
      attachments: attachments.slice(0, 2).map(attachment => ({
        id: attachment.id,
        name: attachment.name
      })),
      attachmentOverflow: Math.max(0, attachments.length - 2),
      busy: Boolean(assistant && isGenerating(assistant))
    }
  }))

const positionedRailItems = computed(() => railItems.value.map(item => ({
  ...item,
  position: railPositions.value[item.id]
})))

function compactText(value: string): string {
  return value.trim().replace(/\s+/g, ' ')
}

function truncatePreview(value: string, limit: number): string {
  return value.length > limit ? `${value.slice(0, limit).trim()}...` : value
}

function tasksFor(messageId: string): WorkbenchTask[] {
  return props.tasksByMessage[messageId] ?? []
}

function stageFor(messageId: string): WorkbenchGenerationStage | undefined {
  return props.stageByMessage[messageId]
}

function reasoningVisible(message: WorkbenchMessage): boolean {
  const stage = stageFor(message.id)
  return Boolean(
    processingVisible(message) && (
      message.reasoning_summary || stage === 'reasoning' || stage === 'answering'
    )
  )
}

function isGenerating(message: WorkbenchMessage): boolean {
  return message.status === 'pending' || message.status === 'in_progress'
}

function isTyping(messageId: string): boolean {
  return Boolean(typingByMessage.value[messageId])
}

function processingVisible(message: WorkbenchMessage): boolean {
  return isGenerating(message) || isTyping(message.id)
}

function setTyping(messageId: string, typing: boolean): void {
  if (Boolean(typingByMessage.value[messageId]) === typing) return
  const next = { ...typingByMessage.value }
  if (typing) next[messageId] = true
  else delete next[messageId]
  typingByMessage.value = next
}

function updateRailVisibility(): void {
  const element = scrollRef.value
  if (element) {
    const total = Math.max(element.scrollHeight, element.clientHeight)
    const nextPositions: Record<string, number> = {}
    for (const item of railItems.value) {
      // Use dataset matching instead of CSS.escape so this also works in DOM test
      // environments that do not expose the optional CSS global.
      const messageElement = Array.from(element.querySelectorAll<HTMLElement>('[data-message-id]'))
        .find(node => node.dataset.messageId === item.id)
      if (messageElement) {
        nextPositions[item.id] = (messageElement.offsetTop + messageElement.offsetHeight / 2) / total
      }
    }
    railPositions.value = nextPositions
  }
  railVisible.value = Boolean(
    element && railItems.value.length > 3 && element.scrollHeight > element.clientHeight + 4
  )
}

async function copy(messageId: string, content: string): Promise<void> {
  if (!content) return
  await navigator.clipboard.writeText(content)
  copiedMessageId.value = messageId
  window.clearTimeout(copiedTimer)
  copiedTimer = window.setTimeout(() => {
    if (copiedMessageId.value === messageId) copiedMessageId.value = ''
  }, 1400)
}

async function scrollToBottom(behavior: 'auto' | 'smooth' = 'smooth'): Promise<void> {
  await nextTick()
  const element = scrollRef.value
  if (!element) return
  updateRailVisibility()
  followLatest = true
  showScrollToBottom.value = false
  element.scrollTo({ top: element.scrollHeight, behavior })
  updateActiveTurn()
}

function onScroll(): void {
  const element = scrollRef.value
  if (!element) return
  updateRailVisibility()
  const distanceFromBottom = element.scrollHeight - element.scrollTop - element.clientHeight
  followLatest = distanceFromBottom <= 72
  showScrollToBottom.value = distanceFromBottom > 72
  updateActiveTurn()
}

function followRenderedContent(): void {
  if (!followLatest || followScrollTimer !== undefined) return
  followScrollTimer = window.setTimeout(() => {
    followScrollTimer = undefined
    if (followLatest) void scrollToBottom('auto')
  }, 16)
}

function updateActiveTurn(): void {
  const element = scrollRef.value
  if (!element || railItems.value.length === 0) return
  const threshold = element.scrollTop + element.clientHeight * 0.34
  let active = railItems.value[0].id
  const messageElements = Array.from(element.querySelectorAll<HTMLElement>('[data-message-id]'))
  for (const item of railItems.value) {
    const messageElement = messageElements.find(node => node.dataset.messageId === item.id)
    if (!messageElement || messageElement.offsetTop > threshold) break
    active = item.id
  }
  activeTurnId.value = active
}

function scrollToTurn(messageId: string): void {
  const element = scrollRef.value
  if (!element) return
  const messageElement = Array.from(element.querySelectorAll<HTMLElement>('[data-message-id]'))
    .find(node => node.dataset.messageId === messageId)
  if (!messageElement) return
  followLatest = false
  activeTurnId.value = messageId
  element.scrollTo({ top: Math.max(0, messageElement.offsetTop - 22), behavior: 'smooth' })
}

watch(
  () => [props.conversationId, props.loading] as const,
  ([conversationId, loading]) => {
    if (loading || !conversationId || conversationId === lastPositionedConversationId) return
    lastPositionedConversationId = conversationId
    typingByMessage.value = {}
    activeTurnId.value = railItems.value.at(-1)?.id ?? ''
    void scrollToBottom('auto')
  },
  { immediate: true, flush: 'post' }
)

watch(
  () => {
    const latest = props.messages.at(-1)
    return [props.messages.length, latest?.id, latest?.content.length, latest?.reasoning_summary?.length] as const
  },
  (current, previous) => {
    const latestMessageChanged = Boolean(current[1] && current[1] !== previous?.[1])
    if (latestMessageChanged) followLatest = true
    if (followLatest) void scrollToBottom('auto')
    else void nextTick(() => {
      updateRailVisibility()
      updateActiveTurn()
    })
  },
  { flush: 'post' }
)

onMounted(() => {
  const element = scrollRef.value
  if (!element) return
  updateRailVisibility()
  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => updateRailVisibility())
    resizeObserver.observe(element)
  }
})

onBeforeUnmount(() => {
  window.clearTimeout(copiedTimer)
  window.clearTimeout(followScrollTimer)
  resizeObserver?.disconnect()
})
</script>

<template>
  <section class="message-thread" aria-live="polite">
    <ConversationRail
      v-if="railVisible"
      :items="positionedRailItems"
      :active-id="activeTurnId"
      @navigate="scrollToTurn"
    />

    <div ref="scrollRef" class="message-thread__scroller" @scroll.passive="onScroll">
      <div v-if="loading" class="thread-state">
        <span class="loading-ring" />
      </div>

      <div v-else-if="messages.length === 0" class="thread-empty">
        <div class="ai-mark"><Icon name="sparkles" size="lg" /></div>
        <h2>今天想完成什么？</h2>
      </div>

      <div v-else class="message-list">
        <article
          v-for="message in messages"
          :key="message.id"
          class="message"
          :class="`message--${message.role}`"
          :data-message-id="message.id"
        >
          <template v-if="message.role === 'assistant'">
            <div class="assistant-avatar"><Icon name="sparkles" size="sm" /></div>
            <div class="assistant-column">
              <div class="assistant-meta">
                <strong>维枢AI</strong>
              </div>

              <Transition name="processing-collapse">
                <div v-if="processingVisible(message)" class="assistant-processing">
                  <GenerationActivity
                    :tasks="tasksFor(message.id)"
                    :active="isGenerating(message)"
                    :stage="stageFor(message.id)"
                  />
                  <ReasoningPanel
                    :summary="message.reasoning_summary"
                    :streaming="isGenerating(message)"
                    :answer-started="Boolean(message.content)"
                    :visible="reasoningVisible(message)"
                  />
                </div>
              </Transition>

              <div v-if="message.content" class="assistant-content">
                <StreamingMarkdownContent
                  :content="message.content"
                  :streaming="isGenerating(message)"
                  @rendered="followRenderedContent"
                  @typing-change="setTyping(message.id, $event)"
                />
              </div>
              <p v-if="message.status === 'failed' || message.status === 'canceled'" class="message-error">
                {{ message.error_message || (message.status === 'canceled' ? '生成已停止' : '生成失败') }}
              </p>
              <div v-if="message.content" class="message-actions message-actions--assistant">
                <button
                  type="button"
                  class="message-copy"
                  :title="copiedMessageId === message.id ? '已复制' : '复制回答'"
                  @click="copy(message.id, message.content)"
                >
                  <Icon name="copy" size="xs" />
                </button>
              </div>
            </div>
          </template>

          <template v-else>
            <div class="user-column">
              <MessageAttachments :attachments="message.attachments ?? []" />
              <div v-if="message.content" class="user-bubble">
                <p v-if="message.content">{{ message.content }}</p>
              </div>
              <div v-if="message.content" class="message-actions message-actions--user">
                <button
                  type="button"
                  class="message-copy"
                  :title="copiedMessageId === message.id ? '已复制' : '复制消息'"
                  @click="copy(message.id, message.content)"
                >
                  <Icon name="copy" size="xs" />
                </button>
              </div>
            </div>
          </template>
        </article>
      </div>
    </div>

    <Transition name="scroll-button">
      <button
        v-if="showScrollToBottom"
        type="button"
        class="scroll-to-bottom"
        title="回到底部"
        aria-label="回到底部"
        @click="scrollToBottom('smooth')"
      >
        <Icon name="arrowDown" size="sm" :stroke-width="2" />
      </button>
    </Transition>
  </section>
</template>

<style scoped>
.message-thread {
  position: relative;
  min-height: 0;
  flex: 1;
  overflow: hidden;
}

.message-thread__scroller {
  height: 100%;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-color: rgba(124, 114, 103, 0.28) transparent;
}

.message-list {
  width: min(820px, calc(100% - 36px));
  margin: 0 auto;
  padding: 30px 0 28px;
}

.message {
  display: flex;
  width: 100%;
  margin-bottom: 28px;
}

.message--assistant {
  align-items: flex-start;
  gap: 11px;
}

.message--user {
  justify-content: flex-end;
}

.user-column {
  display: grid;
  max-width: min(72%, 620px);
  justify-items: end;
}

.assistant-avatar,
.ai-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #285d54;
  background: #dcebe7;
}

.assistant-avatar {
  width: 30px;
  height: 30px;
  flex: 0 0 30px;
  border: 1px solid rgba(48, 105, 94, 0.12);
  border-radius: 6px;
}

.assistant-column {
  position: relative;
  min-width: 0;
  max-width: calc(100% - 42px);
  padding-top: 1px;
}

.assistant-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 23px;
}

.assistant-meta strong {
  color: #2e2a25;
  font-size: 12px;
}

.assistant-content {
  max-width: 100%;
  color: #28241f;
  animation: answer-arrive 220ms ease-out both;
}

.assistant-processing {
  max-height: 520px;
  overflow: hidden;
}

.scroll-to-bottom {
  position: absolute;
  bottom: 16px;
  left: 50%;
  z-index: 12;
  display: inline-flex;
  width: 38px;
  height: 38px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 6px 18px rgba(55, 47, 38, 0.14);
  color: #25211d;
  transform: translateX(-50%);
  backdrop-filter: blur(8px);
}

.scroll-to-bottom:hover,
.scroll-to-bottom:focus-visible {
  border-color: rgba(64, 114, 104, 0.38);
  outline: 0;
  background: #fff;
  color: #356f64;
}

.user-bubble {
  max-width: 100%;
  border: 1px solid rgba(92, 125, 148, 0.08);
  border-radius: 15px 15px 4px 15px;
  padding: 10px 14px;
  background: #edf4fa;
  color: #26333c;
}

.user-bubble p {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-size: 14px;
  line-height: 1.62;
}

.message-error {
  margin: 8px 0 0;
  color: #b4534b;
  font-size: 11px;
}

.message-actions {
  display: flex;
  min-height: 28px;
  align-items: center;
  opacity: 0;
  pointer-events: none;
  transition: opacity 140ms ease;
}

.message-actions--assistant { justify-content: flex-start; }
.message-actions--user { justify-content: flex-end; }

.message:hover .message-actions,
.message:focus-within .message-actions {
  opacity: 1;
  pointer-events: auto;
}

.message-copy {
  display: inline-flex;
  width: 26px;
  height: 26px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #9a9085;
}

.message-copy:hover {
  background: rgba(23, 20, 17, 0.06);
  color: #554e45;
}

.thread-state,
.thread-empty {
  display: flex;
  height: 100%;
  min-height: 320px;
  align-items: center;
  justify-content: center;
}

.thread-empty {
  flex-direction: column;
  gap: 14px;
  color: #3d3832;
}

.thread-empty h2 {
  margin: 0;
  font-size: 19px;
  font-weight: 650;
  letter-spacing: 0;
}

.ai-mark {
  width: 44px;
  height: 44px;
  border-radius: 7px;
}

.loading-ring {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(85, 155, 141, 0.2);
  border-top-color: #559b8d;
  border-radius: 50%;
  animation: task-spin 0.8s linear infinite;
}

:global(.dark .assistant-avatar),
:global(.dark .ai-mark) {
  border-color: rgba(45, 212, 191, 0.2);
  background: rgba(20, 184, 166, 0.13);
  color: #5eead4;
}

:global(.dark .assistant-meta strong),
:global(.dark .assistant-content),
:global(.dark .thread-empty) {
  color: #f1f5f9;
}

:global(.dark .user-bubble) {
  border-color: rgba(125, 211, 252, 0.12);
  background: #263746;
  color: #edf7ff;
}

:global(.dark .scroll-to-bottom) {
  border-color: rgba(148, 163, 184, 0.2);
  background: rgba(17, 24, 39, 0.96);
  box-shadow: 0 8px 20px rgba(2, 6, 23, 0.34);
  color: #f8fafc;
}

.processing-collapse-enter-active,
.processing-collapse-leave-active {
  transition: max-height 220ms ease, opacity 170ms ease, transform 190ms ease;
}

.processing-collapse-enter-from,
.processing-collapse-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-5px);
}

.scroll-button-enter-active,
.scroll-button-leave-active {
  transition: opacity 160ms ease, transform 180ms ease;
}

.scroll-button-enter-from,
.scroll-button-leave-to {
  opacity: 0;
  transform: translate(-50%, 7px) scale(0.92);
}

@keyframes task-spin { to { transform: rotate(360deg); } }
@keyframes answer-arrive {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
@media (prefers-reduced-motion: reduce) {
  .loading-ring { animation: none; }
  .assistant-content { animation: none; }
  .message-actions { transition: none; }
  .processing-collapse-enter-active,
  .processing-collapse-leave-active,
  .scroll-button-enter-active,
  .scroll-button-leave-active { transition: none; }
}

@media (max-width: 640px) {
  .message-list {
    width: calc(100% - 22px);
    padding-top: 22px;
  }

  .message {
    margin-bottom: 23px;
  }

  .user-column { max-width: 88%; }

  .assistant-column {
    max-width: calc(100% - 40px);
  }
}
</style>
