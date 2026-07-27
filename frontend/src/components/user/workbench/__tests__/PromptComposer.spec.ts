import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import PromptComposer from '../PromptComposer.vue'

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
    await wrapper.get('button[title="发送"]').trigger('click')

    const sent = wrapper.emitted('send')?.[0]
    expect(sent?.[0]).toBe('')
    expect(sent?.[1]).toEqual([
      expect.objectContaining({ name: 'clipboard.png', mime_type: 'image/png' })
    ])
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
    await wrapper.get('.composer-action--send').trigger('click')

    expect(wrapper.emitted('send')?.[0]?.[1]).toEqual([
      expect.objectContaining({ name: 'project/docs/README.md', mime_type: 'text/markdown' })
    ])
  })
})
