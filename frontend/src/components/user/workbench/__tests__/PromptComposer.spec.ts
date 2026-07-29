import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PromptComposer from '../PromptComposer.vue'

const attachmentAPI = vi.hoisted(() => ({
  create: vi.fn(),
  upload: vi.fn(),
  complete: vi.fn(),
  remove: vi.fn()
}))

vi.mock('@/api/workbench', async importOriginal => {
  const actual = await importOriginal<typeof import('@/api/workbench')>()
  return {
    ...actual,
    createWorkbenchAttachmentUpload: attachmentAPI.create,
    uploadWorkbenchAttachment: attachmentAPI.upload,
    completeWorkbenchAttachmentUpload: attachmentAPI.complete,
    deleteWorkbenchAttachment: attachmentAPI.remove
  }
})

const uploadedAttachments = new Map<string, { id: string; name: string; mime_type: string; size_bytes: number }>()

beforeEach(() => {
  vi.clearAllMocks()
  uploadedAttachments.clear()
  let sequence = 0
  attachmentAPI.create.mockImplementation(async input => {
    sequence += 1
    const attachment = {
      id: `00000000-0000-4000-8000-${String(sequence).padStart(12, '0')}`,
      name: input.name,
      mime_type: input.mime_type,
      size_bytes: input.size_bytes
    }
    uploadedAttachments.set(attachment.id, attachment)
    return {
      attachment,
      upload: { url: `/upload/${attachment.id}`, method: 'PUT', requires_auth: true },
      expires_at: '2026-07-28T12:00:00Z'
    }
  })
  attachmentAPI.upload.mockResolvedValue(undefined)
  attachmentAPI.complete.mockImplementation(async id => uploadedAttachments.get(id))
  attachmentAPI.remove.mockResolvedValue(undefined)
})

describe('PromptComposer', () => {
  it('accepts a pasted image and sends it with an empty text prompt', async () => {
    const wrapper = mount(PromptComposer, {
      props: { canSend: true, streaming: false, modelName: 'GPT-5.5' }
    })
    const file = new File(['image'], 'clipboard.png', { type: 'image/png' })

    await wrapper.get('textarea').trigger('paste', {
      clipboardData: {
        items: [{ kind: 'file', type: 'image/png', getAsFile: () => file }]
      }
    })
    await vi.waitFor(() => {
      expect(wrapper.find('.attachment-preview img').exists()).toBe(true)
    })
    await vi.waitFor(() => expect(wrapper.get('button[title="发送"]').attributes('disabled')).toBeUndefined())
    await wrapper.get('button[title="发送"]').trigger('click')

    const sent = wrapper.emitted('send')?.[0]
    expect(sent?.[0]).toBe('')
    expect(sent?.[1]).toEqual([
      expect.objectContaining({ name: 'clipboard.png', mime_type: 'image/png' })
    ])
    expect(sent?.[1]?.[0]).not.toHaveProperty('data_url')
    expect(attachmentAPI.create).toHaveBeenCalledOnce()
    expect(attachmentAPI.upload).toHaveBeenCalledOnce()
    expect(attachmentAPI.complete).toHaveBeenCalledOnce()
  })

  it('accepts PDF, Word, Excel and text files', async () => {
    const wrapper = mount(PromptComposer, {
      props: { canSend: true, streaming: false, modelName: 'GPT-5.5' }
    })
    const input = wrapper.findAll('input[type="file"]')[0]
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [
        new File(['pdf'], 'brief.pdf', { type: 'application/pdf' }),
        new File(['word'], 'brief.docx', {
          type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
        }),
        new File(['sheet'], 'metrics.xlsx', {
          type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        }),
        new File(['notes'], 'notes.txt', { type: 'text/plain' })
      ]
    })

    await input.trigger('change')
    await vi.waitFor(() => expect(wrapper.findAll('.attachment-preview')).toHaveLength(4))

    expect(wrapper.text()).toContain('PDF')
    expect(wrapper.text()).toContain('DOCX')
    expect(wrapper.text()).toContain('XLSX')
    expect(wrapper.text()).toContain('TXT')
  })

  it('rejects unsupported executable files', async () => {
    const wrapper = mount(PromptComposer, {
      props: { canSend: true, streaming: false, modelName: 'GPT-5.5' }
    })
    const input = wrapper.findAll('input[type="file"]')[0]
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [new File(['binary'], 'installer.exe', { type: 'application/octet-stream' })]
    })

    await input.trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('已跳过 1 个不支持的文件')
    expect(wrapper.find('.attachment-preview').exists()).toBe(false)
  })

  it('shows a clear message when a file exceeds 10 MB', async () => {
    const wrapper = mount(PromptComposer, {
      props: { canSend: true, streaming: false, modelName: 'GPT-5.5' }
    })
    const file = new File(['x'], 'large.pdf', { type: 'application/pdf' })
    Object.defineProperty(file, 'size', { configurable: true, value: 10 * 1024 * 1024 + 1 })
    const input = wrapper.findAll('input[type="file"]')[0]
    Object.defineProperty(input.element, 'files', { configurable: true, value: [file] })

    await input.trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('large.pdf 超过 10 MB')
    expect(attachmentAPI.create).not.toHaveBeenCalled()
  })

  it('keeps sending disabled until the object upload is complete', async () => {
    let finishUpload: (() => void) | undefined
    attachmentAPI.upload.mockImplementation(() => new Promise<void>(resolve => {
      finishUpload = resolve
    }))
    const wrapper = mount(PromptComposer, {
      props: { canSend: true, streaming: false, modelName: 'GPT-5.5' }
    })
    const input = wrapper.findAll('input[type="file"]')[0]
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [new File(['notes'], 'notes.txt', { type: 'text/plain' })]
    })

    await input.trigger('change')
    await vi.waitFor(() => expect(attachmentAPI.upload).toHaveBeenCalledOnce())

    expect(wrapper.get('.composer-action--send').attributes('disabled')).toBeDefined()
    expect(wrapper.get('.attachment-spinner').attributes('aria-label')).toBe('正在上传')

    finishUpload?.()
    await vi.waitFor(() => expect(wrapper.get('.composer-action--send').attributes('disabled')).toBeUndefined())
  })

  it('keeps folder-relative paths when sending selected files', async () => {
    const wrapper = mount(PromptComposer, {
      props: { canSend: true, streaming: false, modelName: 'GPT-5.5' }
    })
    const file = new File(['readme'], 'README.md', { type: 'text/markdown' })
    Object.defineProperty(file, 'webkitRelativePath', {
      configurable: true,
      value: 'project/docs/README.md'
    })
    const folderInput = wrapper.findAll('input[type="file"]')[1]
    Object.defineProperty(folderInput.element, 'files', {
      configurable: true,
      value: [file]
    })

    await folderInput.trigger('change')
    await vi.waitFor(() => expect(wrapper.find('.attachment-preview').exists()).toBe(true))
    await vi.waitFor(() => expect(wrapper.get('.composer-action--send').attributes('disabled')).toBeUndefined())
    await wrapper.get('.composer-action--send').trigger('click')

    expect(wrapper.emitted('send')?.[0]?.[1]).toEqual([
      expect.objectContaining({ name: 'project/docs/README.md', mime_type: 'text/markdown' })
    ])
  })
})
