import { onScopeDispose, shallowRef, watch, type Ref } from 'vue'
import { getWorkbenchAttachmentContent, type WorkbenchAttachment } from '@/api/workbench'

export function useWorkbenchAttachmentContent(attachments: Ref<WorkbenchAttachment[]>) {
  const objectURLs = shallowRef<Record<string, string>>({})
  let loadVersion = 0

  watch(attachments, async current => {
    const version = ++loadVersion
    revokeAll()
    const next: Record<string, string> = {}
    await Promise.all(current.map(async attachment => {
      if (attachment.data_url || !attachment.mime_type.startsWith('image/')) return
      try {
        const blob = await getWorkbenchAttachmentContent(attachment.id)
        const url = URL.createObjectURL(blob)
        if (version !== loadVersion) {
          URL.revokeObjectURL(url)
          return
        }
        next[attachment.id] = url
      } catch {
        // Keep the attachment tile visible; download can be retried by the user.
      }
    }))
    if (version === loadVersion) objectURLs.value = next
  }, { immediate: true })

  function contentURL(attachment: WorkbenchAttachment): string {
    return attachment.data_url || objectURLs.value[attachment.id] || ''
  }

  async function download(attachment: WorkbenchAttachment): Promise<void> {
    if (attachment.data_url) {
      triggerDownload(attachment.data_url, attachment.name)
      return
    }
    const blob = await getWorkbenchAttachmentContent(attachment.id, true)
    const url = URL.createObjectURL(blob)
    try {
      triggerDownload(url, attachment.name)
    } finally {
      setTimeout(() => URL.revokeObjectURL(url), 0)
    }
  }

  function revokeAll(): void {
    for (const url of Object.values(objectURLs.value)) URL.revokeObjectURL(url)
    objectURLs.value = {}
  }

  onScopeDispose(() => {
    loadVersion += 1
    revokeAll()
  })

  return { contentURL, download }
}

function triggerDownload(url: string, name: string): void {
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = name
  anchor.rel = 'noopener noreferrer'
  anchor.click()
}
