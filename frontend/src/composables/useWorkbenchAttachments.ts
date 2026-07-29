import { computed, onScopeDispose, shallowRef } from 'vue'
import {
  completeWorkbenchAttachmentUpload,
  createWorkbenchAttachmentUpload,
  deleteWorkbenchAttachment,
  uploadWorkbenchAttachment,
  type WorkbenchAttachment
} from '@/api/workbench'
import { extractApiErrorMessage } from '@/utils/apiError'

const MAX_ATTACHMENTS = 8
const MAX_ATTACHMENT_BYTES = 10 * 1024 * 1024
const MAX_TOTAL_BYTES = 24 * 1024 * 1024

const MIME_BY_EXTENSION: Record<string, string> = {
  png: 'image/png', jpg: 'image/jpeg', jpeg: 'image/jpeg', webp: 'image/webp', gif: 'image/gif',
  pdf: 'application/pdf', doc: 'application/msword',
  docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  xls: 'application/vnd.ms-excel',
  xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ppt: 'application/vnd.ms-powerpoint',
  pptx: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  json: 'application/json', xml: 'application/xml', rtf: 'application/rtf',
  txt: 'text/plain', log: 'text/plain', md: 'text/markdown', csv: 'text/csv', html: 'text/html',
  yaml: 'text/yaml', yml: 'text/yaml', js: 'text/plain', ts: 'text/plain', jsx: 'text/plain',
  tsx: 'text/plain', py: 'text/plain', go: 'text/plain', java: 'text/plain', sql: 'text/plain',
  c: 'text/plain', cpp: 'text/plain', h: 'text/plain', css: 'text/plain', sh: 'text/plain'
}

const ACCEPTED_MIME_TYPES = new Set(Object.values(MIME_BY_EXTENSION))

export const WORKBENCH_ATTACHMENT_ACCEPT = Object.keys(MIME_BY_EXTENSION)
  .map(extension => `.${extension}`)
  .join(',')

export interface ComposerAttachment {
  localId: string
  name: string
  mimeType: string
  sizeBytes: number
  previewURL: string
  status: 'uploading' | 'ready'
  attachment?: WorkbenchAttachment
}

export function useWorkbenchAttachments() {
  const items = shallowRef<ComposerAttachment[]>([])
  const errorMessage = shallowRef('')
  const canceled = new Set<string>()
  let queue: Promise<void> = Promise.resolve()
  let disposed = false

  const uploading = computed(() => items.value.some(item => item.status === 'uploading'))

  function addFiles(files: File[]): Promise<void> {
    const task = queue.then(() => processFiles(files))
    queue = task.catch(() => undefined)
    return task
  }

  async function processFiles(files: File[]): Promise<void> {
    errorMessage.value = ''
    let skipped = 0
    for (const file of files) {
      if (disposed) break
      if (items.value.length >= MAX_ATTACHMENTS) {
        errorMessage.value = `每条消息最多添加 ${MAX_ATTACHMENTS} 个附件`
        break
      }
      const mimeType = normalizedMimeType(file)
      if (!mimeType) {
        skipped += 1
        continue
      }
      if (file.size > MAX_ATTACHMENT_BYTES) {
        errorMessage.value = `${file.name} 超过 10 MB`
        continue
      }
      const totalBytes = items.value.reduce((sum, item) => sum + item.sizeBytes, 0) + file.size
      if (totalBytes > MAX_TOTAL_BYTES) {
        errorMessage.value = '本条消息的附件总大小不能超过 24 MB'
        break
      }

      const name = file.webkitRelativePath || file.name || 'attachment'
      const localId = createLocalId()
      let previewURL = ''
      try {
        previewURL = mimeType.startsWith('image/') ? await createPreviewURL(file) : ''
      } catch {
        errorMessage.value = `${name} 读取失败`
        continue
      }
      const item: ComposerAttachment = {
        localId, name, mimeType, sizeBytes: file.size, previewURL, status: 'uploading'
      }
      items.value = [...items.value, item]

      let remoteId = ''
      try {
        const ticket = await createWorkbenchAttachmentUpload({
          name,
          mime_type: mimeType,
          size_bytes: file.size
        })
        remoteId = ticket.attachment.id
        if (disposed || canceled.has(localId)) {
          await deleteWorkbenchAttachment(remoteId)
          continue
        }
        await uploadWorkbenchAttachment(ticket, file)
        const attachment = await completeWorkbenchAttachmentUpload(remoteId)
        if (disposed || canceled.has(localId)) {
          await deleteWorkbenchAttachment(remoteId)
          continue
        }
        replaceItem(localId, { ...item, status: 'ready', attachment })
      } catch (error) {
        if (remoteId) void deleteWorkbenchAttachment(remoteId).catch(() => undefined)
        removeLocalItem(localId)
        errorMessage.value = extractApiErrorMessage(error, `${name} 上传失败`)
      }
    }
    if (!errorMessage.value && skipped > 0) {
      errorMessage.value = `已跳过 ${skipped} 个不支持的文件`
    }
  }

  function removeAttachment(localId: string): void {
    const item = items.value.find(candidate => candidate.localId === localId)
    if (!item) return
    canceled.add(localId)
    removeLocalItem(localId)
    errorMessage.value = ''
    if (item.attachment?.id) {
      void deleteWorkbenchAttachment(item.attachment.id).catch(() => undefined)
    }
  }

  function attachmentsForSend(): WorkbenchAttachment[] {
    return items.value.flatMap(item => item.status === 'ready' && item.attachment ? [item.attachment] : [])
  }

  function clearAfterSend(): void {
    for (const item of items.value) revokePreviewURL(item.previewURL)
    items.value = []
    errorMessage.value = ''
  }

  function replaceItem(localId: string, replacement: ComposerAttachment): void {
    items.value = items.value.map(item => item.localId === localId ? replacement : item)
  }

  function removeLocalItem(localId: string): void {
    const item = items.value.find(candidate => candidate.localId === localId)
    if (item) revokePreviewURL(item.previewURL)
    items.value = items.value.filter(candidate => candidate.localId !== localId)
  }

  onScopeDispose(() => {
    disposed = true
    for (const item of items.value) {
      canceled.add(item.localId)
      revokePreviewURL(item.previewURL)
      if (item.attachment?.id) void deleteWorkbenchAttachment(item.attachment.id).catch(() => undefined)
    }
  })

  return {
    items,
    errorMessage,
    uploading,
    addFiles,
    removeAttachment,
    attachmentsForSend,
    clearAfterSend
  }
}

function normalizedMimeType(file: File): string {
  const declared = file.type.toLowerCase().trim()
  if (ACCEPTED_MIME_TYPES.has(declared)) return declared
  const extension = file.name.split('.').pop()?.toLowerCase() ?? ''
  return MIME_BY_EXTENSION[extension] ?? ''
}

function createLocalId(): string {
  return typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : `attachment-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

async function createPreviewURL(file: File): Promise<string> {
  if (typeof URL.createObjectURL === 'function') return URL.createObjectURL(file)
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error ?? new Error('读取附件失败'))
    reader.readAsDataURL(file)
  })
}

function revokePreviewURL(url: string): void {
  if (url.startsWith('blob:') && typeof URL.revokeObjectURL === 'function') URL.revokeObjectURL(url)
}
