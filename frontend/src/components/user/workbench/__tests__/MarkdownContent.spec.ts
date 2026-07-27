import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MarkdownContent from '../MarkdownContent.vue'

describe('MarkdownContent', () => {
  const writeText = vi.fn()

  beforeEach(() => {
    writeText.mockReset()
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText }
    })
  })

  it('renders a copy control for fenced code and copies only the code text', async () => {
    const wrapper = mount(MarkdownContent, {
      props: { content: '```ts\nconst total = 42\n```' }
    })

    expect(wrapper.get('.code-block__toolbar').text()).toContain('ts')
    await wrapper.get('[data-code-copy]').trigger('click')
    await flushPromises()

    expect(writeText).toHaveBeenCalledWith('const total = 42')
    expect(wrapper.get('[data-code-copy]').text()).toBe('已复制')
    wrapper.unmount()
  })
})
