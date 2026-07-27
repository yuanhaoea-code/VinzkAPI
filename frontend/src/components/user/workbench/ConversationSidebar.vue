<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import type { WorkbenchConversation } from '@/api/workbench'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  conversations: WorkbenchConversation[]
  activeId?: string
  open: boolean
  disabled?: boolean
  busyIds?: string[]
}>()

const emit = defineEmits<{
  select: [id: string]
  create: []
  delete: [id: string]
  close: []
}>()

const query = shallowRef('')
const filteredConversations = computed(() => {
  const normalized = query.value.trim().toLowerCase()
  if (!normalized) return props.conversations
  return props.conversations.filter(item => item.title.toLowerCase().includes(normalized))
})

function selectConversation(id: string): void {
  emit('select', id)
  emit('close')
}
</script>

<template>
  <button
    v-if="open"
    type="button"
    class="conversation-backdrop"
    aria-label="关闭会话列表"
    @click="emit('close')"
  />
  <aside class="conversation-sidebar" :class="{ 'conversation-sidebar--open': open }">
    <div class="conversation-sidebar__head">
      <div>
        <span class="conversation-sidebar__eyebrow">WORKBENCH</span>
        <strong>维枢AI</strong>
      </div>
      <button type="button" class="icon-button mobile-close" title="关闭" @click="emit('close')">
        <Icon name="x" size="sm" />
      </button>
    </div>

    <button
      type="button"
      class="new-conversation"
      :disabled="disabled"
      @click="emit('create')"
    >
      <Icon name="plus" size="sm" />
      <span>新对话</span>
    </button>

    <label class="conversation-search">
      <Icon name="search" size="sm" />
      <input v-model="query" type="search" placeholder="搜索对话" />
    </label>

    <div class="conversation-list">
      <div
        v-for="conversation in filteredConversations"
        :key="conversation.id"
        class="conversation-item"
        :class="{ 'conversation-item--active': conversation.id === activeId }"
      >
        <button
          type="button"
          class="conversation-item__select"
          @click="selectConversation(conversation.id)"
        >
          <span
            v-if="busyIds?.includes(conversation.id)"
            class="conversation-item__activity"
            aria-label="正在生成"
          />
          <Icon v-else name="chat" size="sm" />
          <span class="conversation-item__title">{{ conversation.title }}</span>
        </button>
        <button
          type="button"
          class="conversation-item__delete"
          title="删除对话"
          :disabled="disabled || busyIds?.includes(conversation.id)"
          @click.stop="emit('delete', conversation.id)"
        >
          <Icon name="trash" size="xs" />
        </button>
      </div>
      <p v-if="filteredConversations.length === 0" class="conversation-empty">暂无对话</p>
    </div>
  </aside>
</template>

<style scoped>
.conversation-sidebar {
  width: 248px;
  flex: 0 0 248px;
  display: flex;
  min-height: 0;
  flex-direction: column;
  border-right: 1px solid rgba(23, 20, 17, 0.08);
  background: rgba(250, 247, 240, 0.72);
}

.conversation-sidebar__head {
  display: flex;
  min-height: 70px;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px 12px;
}

.conversation-sidebar__head div {
  display: grid;
  gap: 2px;
}

.conversation-sidebar__head strong {
  color: #171411;
  font-size: 17px;
  letter-spacing: 0;
}

.conversation-sidebar__eyebrow {
  color: #887e72;
  font-size: 10px;
  letter-spacing: 0.16em;
}

.new-conversation {
  display: flex;
  min-height: 38px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 0 12px 12px;
  border: 1px solid #171411;
  border-radius: 6px;
  background: #171411;
  color: #fffdf8;
  font-size: 13px;
  font-weight: 650;
}

.new-conversation:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.conversation-search {
  display: flex;
  min-height: 34px;
  align-items: center;
  gap: 7px;
  margin: 0 12px 10px;
  padding: 0 10px;
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.58);
  color: #8b8277;
}

.conversation-search input {
  min-width: 0;
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  color: #171411;
  font-size: 12px;
}

.conversation-list {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding: 0 8px 14px;
}

.conversation-item {
  position: relative;
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) 22px;
  align-items: center;
  gap: 2px;
  min-height: 42px;
  margin: 1px 0;
  padding: 0 6px 0 0;
  border-radius: 6px;
  background: transparent;
  color: #6f675e;
}

.conversation-item:hover,
.conversation-item--active {
  background: rgba(255, 255, 255, 0.76);
  color: #171411;
}

.conversation-item--active {
  box-shadow: inset 2px 0 #75a49b;
}

.conversation-item__select {
  display: grid;
  min-width: 0;
  min-height: 42px;
  grid-template-columns: 16px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  padding: 7px 8px;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
}

.conversation-item__title {
  overflow: hidden;
  font-size: 12px;
  font-weight: 550;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-item__activity {
  width: 9px;
  height: 9px;
  margin-left: 3px;
  border: 1px solid rgba(79, 143, 131, 0.28);
  border-top-color: #4f8f83;
  border-radius: 50%;
  animation: conversation-spin 0.85s linear infinite;
}

.conversation-item__delete {
  display: none;
  width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #958b80;
}

.conversation-item:hover .conversation-item__delete,
.conversation-item--active .conversation-item__delete {
  display: flex;
}

.conversation-empty {
  margin: 22px 10px;
  color: #958b80;
  font-size: 12px;
  text-align: center;
}

.icon-button {
  display: inline-flex;
  width: 30px;
  height: 30px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: #71695f;
}

.mobile-close,
.conversation-backdrop {
  display: none;
}

:global(.dark .conversation-sidebar) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(9, 17, 30, 0.88);
}

:global(.dark .conversation-sidebar__head strong),
:global(.dark .conversation-search input) {
  color: #f8fafc;
}

:global(.dark .conversation-sidebar__eyebrow),
:global(.dark .conversation-item),
:global(.dark .conversation-search) {
  color: #94a3b8;
}

:global(.dark .conversation-search) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.7);
}

:global(.dark .new-conversation) {
  border-color: #2dd4bf;
  background: #14b8a6;
  color: #06211d;
}

:global(.dark .conversation-item:hover),
:global(.dark .conversation-item--active) {
  background: rgba(20, 184, 166, 0.1);
  color: #f8fafc;
}

@keyframes conversation-spin { to { transform: rotate(360deg); } }

@media (prefers-reduced-motion: reduce) {
  .conversation-item__activity { animation: none; }
}

@media (max-width: 900px) {
  .conversation-sidebar {
    position: fixed;
    inset: 56px auto 0 0;
    z-index: 45;
    background: #faf7f0;
    transform: translateX(-100%);
    transition: transform 180ms ease;
    box-shadow: 18px 0 38px rgba(23, 20, 17, 0.14);
  }

  :global(.dark .conversation-sidebar) {
    background: #09111e;
  }

  .conversation-sidebar--open {
    transform: translateX(0);
  }

  .mobile-close {
    display: inline-flex;
  }

  .conversation-backdrop {
    position: fixed;
    inset: 56px 0 0;
    z-index: 44;
    display: block;
    border: 0;
    background: rgba(23, 20, 17, 0.2);
  }
}
</style>
