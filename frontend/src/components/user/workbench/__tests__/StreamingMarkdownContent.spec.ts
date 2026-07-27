import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import StreamingMarkdownContent from '../StreamingMarkdownContent.vue'

describe('StreamingMarkdownContent', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('reveals streamed content one character at a time', async () => {
    const wrapper = mount(StreamingMarkdownContent, {
      props: { content: '', streaming: true }
    })

    await wrapper.setProps({ content: '你好呀' })
    expect(wrapper.text()).toBe('')

    await vi.advanceTimersByTimeAsync(16)
    expect(wrapper.text()).toBe('你')
    await vi.advanceTimersByTimeAsync(16)
    expect(wrapper.text()).toBe('你好')
    await vi.advanceTimersByTimeAsync(16)
    expect(wrapper.text()).toBe('你好呀')
  })

  it('renders historical completed messages immediately', () => {
    const wrapper = mount(StreamingMarkdownContent, {
      props: { content: '已经完成的回答', streaming: false }
    })

    expect(wrapper.text()).toBe('已经完成的回答')
  })

  it('finishes its existing queue smoothly after the stream closes', async () => {
    const wrapper = mount(StreamingMarkdownContent, {
      props: { content: '平滑结束', streaming: true }
    })

    await vi.advanceTimersByTimeAsync(16)
    expect(wrapper.text()).toBe('平')
    await wrapper.setProps({ streaming: false })
    expect(wrapper.text()).toBe('平')

    await vi.advanceTimersByTimeAsync(48)
    expect(wrapper.text()).toBe('平滑结束')
    expect(wrapper.emitted('typingChange')?.at(-1)).toEqual([false])
  })
})
