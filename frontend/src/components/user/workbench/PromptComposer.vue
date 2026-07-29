<script setup lang="ts">
import { computed, nextTick, ref, shallowRef, watch } from 'vue'
import type { WorkbenchAttachment } from '@/api/workbench'
import Icon from '@/components/icons/Icon.vue'
import {
  useWorkbenchAttachments,
  WORKBENCH_ATTACHMENT_ACCEPT,
  type ComposerAttachment
} from '@/composables/useWorkbenchAttachments'

const props = defineProps<{
  canSend: boolean
  streaming: boolean
  modelName?: string
}>()

const emit = defineEmits<{
  send: [content: string, attachments: WorkbenchAttachment[]]
  stop: []
}>()

const value = defineModel<string>({ default: '' })
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const folderInputRef = ref<HTMLInputElement | null>(null)
const attachmentMenuOpen = shallowRef(false)
const dragging = shallowRef(false)
const {
  items: attachments,
  errorMessage: attachmentError,
  uploading: attachmentsUploading,
  addFiles,
  removeAttachment,
  attachmentsForSend,
  clearAfterSend
} = useWorkbenchAttachments()

const canSubmit = computed(() => (
  props.canSend && !props.streaming && !attachmentsUploading.value && Boolean(value.value.trim() || attachments.value.length)
))

function resize(): void {
  const element = textareaRef.value
  if (!element) return
  element.style.height = '0'
  element.style.height = `${Math.min(176, Math.max(46, element.scrollHeight))}px`
}

function submit(): void {
  if (!canSubmit.value) return
  emit('send', value.value.trim(), attachmentsForSend())
  value.value = ''
  clearAfterSend()
  attachmentMenuOpen.value = false
  void nextTick(resize)
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return
  event.preventDefault()
  submit()
}

function toggleAttachmentMenu(): void {
  if (!props.modelName || props.streaming || attachmentsUploading.value) return
  attachmentMenuOpen.value = !attachmentMenuOpen.value
}

function openFilePicker(type: 'file' | 'folder'): void {
  attachmentMenuOpen.value = false
  if (type === 'folder') folderInputRef.value?.click()
  else fileInputRef.value?.click()
}

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  void addFiles(Array.from(input.files ?? []))
  input.value = ''
}

function onPaste(event: ClipboardEvent): void {
  const files = Array.from(event.clipboardData?.items ?? [])
    .filter(item => item.kind === 'file')
    .map(item => item.getAsFile())
    .filter((file): file is File => Boolean(file))
  if (files.length === 0) return
  event.preventDefault()
  void addFiles(files)
}

function onDrop(event: DragEvent): void {
  dragging.value = false
  void addFiles(Array.from(event.dataTransfer?.files ?? []))
}

function isImage(attachment: ComposerAttachment): boolean {
  return attachment.mimeType.startsWith('image/')
}

function extension(name: string): string {
  const value = name.split('.').pop()?.toUpperCase()
  return value && value.length <= 6 ? value : 'FILE'
}

watch(value, () => void nextTick(resize))
</script>

<template>
  <footer class="composer-wrap">
    <div
      class="composer"
      :class="{ 'composer--streaming': streaming, 'composer--dragging': dragging }"
      @dragenter.prevent="dragging = true"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop.prevent="onDrop"
    >
      <div v-if="attachments.length" class="attachment-strip" aria-label="待发送附件">
        <figure
          v-for="attachment in attachments"
          :key="attachment.localId"
          class="attachment-preview"
          :class="{
            'attachment-preview--file': !isImage(attachment),
            'attachment-preview--uploading': attachment.status === 'uploading'
          }"
        >
          <img v-if="isImage(attachment)" :src="attachment.previewURL" :alt="attachment.name" />
          <template v-else>
            <Icon name="document" size="md" />
            <strong>{{ extension(attachment.name) }}</strong>
            <span>{{ attachment.name.split('/').pop() }}</span>
          </template>
          <span v-if="attachment.status === 'uploading'" class="attachment-spinner" aria-label="正在上传" />
          <button type="button" :title="`移除 ${attachment.name}`" @click="removeAttachment(attachment.localId)">
            <Icon name="x" size="xs" />
          </button>
        </figure>
      </div>

      <div class="composer-editor">
        <div class="attachment-control">
          <button
            type="button"
            class="attachment-button"
            title="添加附件"
            :disabled="!modelName || streaming || attachmentsUploading"
            :aria-expanded="attachmentMenuOpen"
            @click="toggleAttachmentMenu"
          >
            <Icon name="paperclip" size="sm" />
          </button>
          <Transition name="attachment-menu">
            <div v-if="attachmentMenuOpen" class="attachment-menu">
              <button type="button" @click="openFilePicker('file')">
                <Icon name="document" size="sm" />
                <span>选择文件</span>
              </button>
              <button type="button" @click="openFilePicker('folder')">
                <Icon name="folder" size="sm" />
                <span>选择文件夹</span>
              </button>
            </div>
          </Transition>
        </div>
        <textarea
          ref="textareaRef"
          v-model="value"
          rows="1"
          :placeholder="modelName ? `发送消息给 ${modelName}` : '请先添加可用模型'"
          :disabled="!modelName"
          aria-label="消息内容"
          @keydown="onKeydown"
          @paste="onPaste"
        />
      </div>

      <button
        v-if="streaming"
        type="button"
        class="composer-action composer-action--stop"
        title="停止生成"
        @click="emit('stop')"
      >
        <span />
      </button>
      <button
        v-else
        type="button"
        class="composer-action composer-action--send"
        title="发送"
        :disabled="!canSubmit"
        @click="submit"
      >
        <Icon name="arrowUp" size="sm" />
      </button>

      <input
        ref="fileInputRef"
        class="file-input"
        type="file"
        :accept="WORKBENCH_ATTACHMENT_ACCEPT"
        multiple
        @change="onFileChange"
      />
      <input
        ref="folderInputRef"
        class="file-input"
        type="file"
        multiple
        webkitdirectory
        directory
        @change="onFileChange"
      />
    </div>
    <span v-if="attachmentError" class="composer-error">{{ attachmentError }}</span>
  </footer>
</template>

<style scoped>
.composer-wrap {
  width: min(820px, calc(100% - 36px));
  margin: 0 auto;
  padding: 0 0 14px;
}

.composer {
  position: relative;
  display: grid;
  min-height: 58px;
  grid-template-columns: minmax(0, 1fr) 36px;
  align-items: end;
  gap: 8px;
  border: 1px solid rgba(23, 20, 17, 0.12);
  border-radius: 8px;
  padding: 7px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 8px 28px rgba(63, 53, 43, 0.07);
  transition: border-color 160ms ease, box-shadow 160ms ease, background 160ms ease;
}

.composer:focus-within,
.composer--dragging {
  border-color: #65a296;
  box-shadow: 0 0 0 3px rgba(101, 162, 150, 0.14), 0 8px 28px rgba(63, 53, 43, 0.07);
}

.composer--dragging { background: #edf7f4; }

.attachment-strip {
  display: flex;
  min-width: 0;
  grid-column: 1 / -1;
  gap: 8px;
  overflow-x: auto;
  padding: 2px 2px 5px;
}

.attachment-preview {
  position: relative;
  display: grid;
  width: 64px;
  height: 64px;
  flex: 0 0 64px;
  place-items: center;
  margin: 0;
  overflow: hidden;
  border: 1px solid rgba(23, 20, 17, 0.12);
  border-radius: 7px;
  background: #f3f0ea;
  color: #4f7d74;
}

.attachment-preview img { width: 100%; height: 100%; object-fit: cover; }

.attachment-preview--uploading::after {
  position: absolute;
  z-index: 1;
  inset: 0;
  background: rgba(15, 23, 42, 0.32);
  content: '';
}

.attachment-spinner {
  position: absolute;
  z-index: 2;
  top: 50%;
  left: 50%;
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.45);
  border-top-color: #fff;
  border-radius: 50%;
  animation: attachment-spin 700ms linear infinite;
  transform: translate(-50%, -50%);
}

@keyframes attachment-spin {
  to { transform: translate(-50%, -50%) rotate(360deg); }
}

.attachment-preview--file {
  grid-template-rows: 27px 11px 13px;
  padding: 4px 5px;
}

.attachment-preview--file strong {
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 8px;
}

.attachment-preview--file span {
  width: 100%;
  overflow: hidden;
  color: #746d64;
  font-size: 7px;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attachment-preview > button {
  position: absolute;
  top: 3px;
  right: 3px;
  display: inline-flex;
  width: 20px;
  height: 20px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.45);
  border-radius: 50%;
  background: rgba(23, 20, 17, 0.72);
  color: #fff;
}

.composer-editor {
  display: grid;
  min-width: 0;
  grid-template-columns: 30px minmax(0, 1fr);
  align-items: end;
}

.attachment-control { position: relative; }

.attachment-button {
  display: inline-flex;
  width: 30px;
  height: 34px;
  align-items: center;
  justify-content: center;
  margin-bottom: 1px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: #766e64;
}

.attachment-button:hover:not(:disabled) { background: rgba(72, 112, 104, 0.1); color: #356f64; }
.attachment-button:disabled { opacity: 0.35; }

.attachment-menu {
  position: absolute;
  bottom: 42px;
  left: 0;
  z-index: 12;
  width: 150px;
  overflow: hidden;
  border: 1px solid rgba(23, 20, 17, 0.11);
  border-radius: 7px;
  background: #fffdf9;
  box-shadow: 0 12px 30px rgba(52, 45, 38, 0.13);
}

.attachment-menu button {
  display: grid;
  width: 100%;
  min-height: 39px;
  grid-template-columns: 18px 1fr;
  align-items: center;
  gap: 8px;
  border: 0;
  padding: 0 11px;
  background: transparent;
  color: #49433d;
  font-size: 11px;
  text-align: left;
}

.attachment-menu button:hover { background: #edf4f1; color: #2f655b; }
.attachment-menu-enter-active,
.attachment-menu-leave-active { transition: opacity 150ms ease, transform 150ms ease; }
.attachment-menu-enter-from,
.attachment-menu-leave-to { opacity: 0; transform: translateY(4px); }

.composer textarea {
  width: 100%;
  min-height: 46px;
  max-height: 176px;
  resize: none;
  border: 0;
  outline: 0;
  padding: 12px 5px 8px;
  background: transparent;
  color: #231f1b;
  font-family: inherit;
  font-size: 13px;
  line-height: 1.55;
}

.composer textarea::placeholder { color: #9b9186; }

.composer-action {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 6px;
}

.composer-action--send { background: #171411; color: #fffdf8; }
.composer-action--send:disabled { background: #d8d2c9; color: #9d9489; cursor: not-allowed; }
.composer-action--stop { background: #b4534b; }
.composer-action--stop span { width: 10px; height: 10px; border-radius: 1px; background: #fff; }
.file-input { display: none; }

.composer-error {
  display: block;
  margin-top: 7px;
  color: #b4534b;
  font-size: 9px;
  text-align: center;
}

:global(.dark .composer) {
  border-color: rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.92);
  box-shadow: 0 12px 34px rgba(2, 6, 23, 0.24);
}

:global(.dark .composer--dragging) { background: rgba(20, 184, 166, 0.12); }
:global(.dark .composer textarea) { color: #f8fafc; }
:global(.dark .attachment-button) { color: #94a3b8; }
:global(.dark .attachment-preview) { border-color: rgba(148, 163, 184, 0.24); background: #111827; }
:global(.dark .attachment-menu) { border-color: rgba(148, 163, 184, 0.2); background: #111827; }
:global(.dark .attachment-menu button) { color: #cbd5e1; }
:global(.dark .attachment-menu button:hover) { background: rgba(20, 184, 166, 0.12); color: #99f6e4; }
:global(.dark .composer-action--send) { background: #14b8a6; color: #06211d; }
:global(.dark .composer-action--send:disabled) { background: #334155; color: #64748b; }

@media (prefers-reduced-motion: reduce) {
  .composer,
  .attachment-menu-enter-active,
  .attachment-menu-leave-active { transition: none; }
  .attachment-spinner { animation: none; }
}

@media (max-width: 640px) {
  .composer-wrap { width: calc(100% - 20px); padding-bottom: 10px; }
}
</style>
