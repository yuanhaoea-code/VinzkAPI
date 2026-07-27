<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useWorkbench } from '@/composables/useWorkbench'
import ConversationSidebar from './ConversationSidebar.vue'
import MessageThread from './MessageThread.vue'
import PromptComposer from './PromptComposer.vue'
import WorkbenchToolbar from './WorkbenchToolbar.vue'

const sidebarOpen = shallowRef(false)
const prompt = shallowRef('')
const deleteTarget = shallowRef('')

const {
  models,
  conversations,
  activeConversation,
  messages,
  tasksByMessage,
  stageByMessage,
  selectedBindingId,
  selectedModel,
  reasoningPreset,
  loading,
  loadingConversation,
  managingModels,
  creatingConversation,
  streamingMessageId,
  streamingConversationIds,
  activeConversationUpdating,
  canSend,
  initialize,
  newConversation,
  openConversation,
  deleteConversation,
  selectKey,
  selectModel,
  addModels,
  hideModel,
  selectReasoningPreset,
  sendMessage,
  stopGeneration
} = useWorkbench()

async function confirmDelete(): Promise<void> {
  const id = deleteTarget.value
  if (!id) return
  deleteTarget.value = ''
  await deleteConversation(id)
}

onMounted(() => void initialize())
</script>

<template>
  <section class="workbench-shell" aria-label="维枢AI 工作台">
    <ConversationSidebar
      :conversations="conversations"
      :active-id="activeConversation?.id"
      :open="sidebarOpen"
      :disabled="loading || creatingConversation"
      :busy-ids="streamingConversationIds"
      @select="openConversation"
      @create="newConversation"
      @delete="deleteTarget = $event"
      @close="sidebarOpen = false"
    />

    <div class="workbench-main">
      <WorkbenchToolbar
        :keys="models.keys ?? []"
        :binding-id="selectedBindingId"
        :reasoning-preset="reasoningPreset"
        :loading="loading || managingModels || activeConversationUpdating"
        @toggle-sidebar="sidebarOpen = true"
        @select-key="selectKey"
        @select-model="selectModel"
        @add-models="addModels"
        @remove-model="hideModel"
        @select-reasoning="selectReasoningPreset"
      />

      <MessageThread
        :messages="messages"
        :tasks-by-message="tasksByMessage"
        :stage-by-message="stageByMessage"
        :conversation-id="activeConversation?.id"
        :loading="loading || loadingConversation"
      />

      <PromptComposer
        v-model="prompt"
        :can-send="canSend"
        :streaming="Boolean(streamingMessageId)"
        :model-name="selectedModel?.display_name"
        @send="sendMessage"
        @stop="stopGeneration"
      />
    </div>
  </section>

  <ConfirmDialog
    :show="Boolean(deleteTarget)"
    title="删除对话"
    message="此操作会永久删除该对话及其中的消息。"
    confirm-text="删除"
    danger
    @confirm="confirmDelete"
    @cancel="deleteTarget = ''"
  />
</template>

<style scoped>
.workbench-shell {
  display: flex;
  width: 100%;
  min-height: 620px;
  height: calc(100vh - 88px);
  overflow: hidden;
  border: 1px solid rgba(23, 20, 17, 0.11);
  border-radius: 8px;
  background: rgba(255, 253, 248, 0.86);
  box-shadow: 0 18px 50px rgba(72, 57, 39, 0.08);
}

.workbench-main {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  background:
    linear-gradient(rgba(255, 253, 248, 0.82), rgba(255, 253, 248, 0.82)),
    repeating-linear-gradient(0deg, transparent, transparent 27px, rgba(76, 65, 54, 0.025) 28px);
}

:global(.dark .workbench-shell) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(11, 18, 32, 0.9);
  box-shadow: 0 18px 50px rgba(2, 6, 23, 0.24);
}

:global(.dark .workbench-main) {
  background:
    linear-gradient(rgba(10, 17, 29, 0.92), rgba(10, 17, 29, 0.92)),
    repeating-linear-gradient(0deg, transparent, transparent 27px, rgba(148, 163, 184, 0.04) 28px);
}

@media (max-width: 900px) {
  .workbench-shell {
    min-height: calc(100vh - 78px);
    height: calc(100vh - 78px);
    border-radius: 6px;
  }
}

@media (max-width: 640px) {
  .workbench-shell {
    width: calc(100% + 20px);
    min-height: calc(100dvh - 68px);
    height: calc(100dvh - 68px);
    margin: -6px -10px -10px;
    border-right: 0;
    border-bottom: 0;
    border-left: 0;
    border-radius: 0;
  }
}
</style>
