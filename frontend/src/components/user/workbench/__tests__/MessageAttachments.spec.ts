import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MessageAttachments from '../MessageAttachments.vue'

const attachmentAPI = vi.hoisted(() => ({
  getContent: vi.fn()
}))

vi.mock('@/api/workbench', async importOriginal => {
  const actual = await importOriginal<typeof import('@/api/workbench')>()
  return { ...actual, getWorkbenchAttachmentContent: attachmentAPI.getContent }
})

beforeEach(() => {
  vi.clearAllMocks()
  attachmentAPI.getContent.mockResolvedValue(new Blob(['content'], { type: 'image/png' }))
  Object.defineProperty(URL, 'createObjectURL', {
    configurable: true,
    value: vi.fn(() => 'blob:workbench-attachment')
  })
  Object.defineProperty(URL, 'revokeObjectURL', {
    configurable: true,
    value: vi.fn()
  })
})

describe('MessageAttachments', () => {
  it('loads an authenticated image blob when the message contains stored metadata', async () => {
    const wrapper = mount(MessageAttachments, {
      props: {
        attachments: [{
          id: 'attachment-1',
          name: 'diagram.png',
          mime_type: 'image/png',
          size_bytes: 7
        }]
      }
    })

    await flushPromises()

    expect(attachmentAPI.getContent).toHaveBeenCalledWith('attachment-1')
    expect(wrapper.get('img').attributes('src')).toBe('blob:workbench-attachment')
  })

  it('downloads a stored document through the authenticated content endpoint', async () => {
    attachmentAPI.getContent.mockResolvedValue(new Blob(['document'], { type: 'application/pdf' }))
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined)
    const wrapper = mount(MessageAttachments, {
      props: {
        attachments: [{
          id: 'attachment-2',
          name: 'brief.pdf',
          mime_type: 'application/pdf',
          size_bytes: 8
        }]
      }
    })

    await wrapper.get('.message-attachment--file').trigger('click')
    await flushPromises()

    expect(attachmentAPI.getContent).toHaveBeenCalledWith('attachment-2', true)
    expect(click).toHaveBeenCalledOnce()
    click.mockRestore()
  })
})
