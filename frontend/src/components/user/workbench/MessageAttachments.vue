<script setup lang="ts">
import { toRef } from 'vue'
import type { WorkbenchAttachment } from '@/api/workbench'
import Icon from '@/components/icons/Icon.vue'
import { useWorkbenchAttachmentContent } from '@/composables/useWorkbenchAttachmentContent'

const props = defineProps<{ attachments: WorkbenchAttachment[] }>()
const { contentURL, download } = useWorkbenchAttachmentContent(toRef(props, 'attachments'))

function isImage(attachment: WorkbenchAttachment): boolean {
  return attachment.mime_type.startsWith('image/')
}

function extension(name: string): string {
  const value = name.split('.').pop()?.trim().toUpperCase()
  return value && value.length <= 6 ? value : 'FILE'
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${Math.ceil(bytes / 1024)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function onAttachmentClick(event: MouseEvent, attachment: WorkbenchAttachment): void {
  if (isImage(attachment)) return
  event.preventDefault()
  void download(attachment)
}
</script>

<template>
  <div v-if="attachments.length" class="message-attachments">
    <a
      v-for="attachment in attachments"
      :key="attachment.id"
      class="message-attachment"
      :class="isImage(attachment) ? 'message-attachment--image' : 'message-attachment--file'"
      :href="isImage(attachment) ? contentURL(attachment) || undefined : undefined"
      :target="isImage(attachment) ? '_blank' : undefined"
      rel="noopener noreferrer"
      :title="attachment.name"
      @click="onAttachmentClick($event, attachment)"
    >
      <img v-if="isImage(attachment) && contentURL(attachment)" :src="contentURL(attachment)" :alt="attachment.name" />
      <span v-else-if="isImage(attachment)" class="image-loading" aria-label="正在加载附件">
        <Icon name="document" size="md" />
      </span>
      <template v-else>
        <span class="file-mark">
          <Icon name="document" size="md" />
          <small>{{ extension(attachment.name) }}</small>
        </span>
        <span class="file-copy">
          <strong>{{ attachment.name }}</strong>
          <small>{{ formatSize(attachment.size_bytes) }}</small>
        </span>
        <Icon name="download" size="sm" class="file-download" />
      </template>
    </a>
  </div>
</template>

<style scoped>
.message-attachments {
  display: flex;
  max-width: min(560px, 78vw);
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  margin-bottom: 7px;
}

.message-attachment--image {
  display: block;
  width: 116px;
  height: 92px;
  overflow: hidden;
  border: 1px solid rgba(91, 108, 120, 0.15);
  border-radius: 7px;
  background: #eef2f3;
  box-shadow: 0 3px 10px rgba(42, 51, 57, 0.07);
}

.message-attachment--image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 180ms ease, filter 180ms ease;
}

.image-loading {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  color: #789087;
}

.message-attachment--image:hover img {
  filter: saturate(1.04);
  transform: scale(1.025);
}

.message-attachment--file {
  display: grid;
  width: min(300px, 72vw);
  min-height: 58px;
  grid-template-columns: 38px minmax(0, 1fr) 18px;
  align-items: center;
  gap: 9px;
  border: 1px solid rgba(91, 108, 120, 0.13);
  border-radius: 7px;
  padding: 7px 10px 7px 8px;
  background: #edf4fa;
  color: #2e3c45;
  text-decoration: none;
}

.file-mark {
  display: grid;
  width: 38px;
  height: 42px;
  place-items: center;
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.78);
  color: #4f7b75;
}

.file-mark small {
  margin-top: -8px;
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 7px;
  font-weight: 700;
}

.file-copy {
  display: grid;
  min-width: 0;
  gap: 3px;
}

.file-copy strong {
  overflow: hidden;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-copy small { color: #7e8d96; font-size: 9px; }
.file-download { color: #789087; }

:global(.dark .message-attachment--image) {
  border-color: rgba(148, 163, 184, 0.2);
  background: #182231;
}

:global(.dark .message-attachment--file) {
  border-color: rgba(125, 211, 252, 0.12);
  background: #263746;
  color: #edf7ff;
}

:global(.dark .file-mark) { background: rgba(15, 23, 42, 0.58); color: #5eead4; }

@media (prefers-reduced-motion: reduce) {
  .message-attachment--image img { transition: none; }
}
</style>
