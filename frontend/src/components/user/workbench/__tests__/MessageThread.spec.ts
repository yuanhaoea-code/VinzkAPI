import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { WorkbenchMessage } from '@/api/workbench'
import MessageThread from '../MessageThread.vue'

const scrollTo = vi.fn()

function message(
  id: string,
  role: 'user' | 'assistant',
  content: string,
  status: WorkbenchMessage['status'] = 'completed',
  attachments: WorkbenchMessage['attachments'] = []
): WorkbenchMessage {
  return {
    id,
    conversation_id: 'conversation-a',
    sequence: role === 'user' ? 1 : 2,
    role,
    content,
    status,
    attachments,
    created_at: '2026-07-21T01:13:00Z',
    updated_at: '2026-07-21T01:13:00Z'
  }
}

describe('MessageThread', () => {
  beforeEach(() => {
    scrollTo.mockReset()
    Object.defineProperty(HTMLElement.prototype, 'scrollTo', {
      configurable: true,
      value: scrollTo
    })
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: vi.fn().mockResolvedValue(undefined) }
    })
  })

  it('opens a conversation at the bottom but stops following once the user scrolls up', async () => {
    const messages = [
      message('user-1', 'user', '第一轮问题'),
      message('assistant-1', 'assistant', '第一轮回答', 'in_progress')
    ]
    const wrapper = mount(MessageThread, {
      props: {
        messages,
        tasksByMessage: {},
        stageByMessage: {},
        conversationId: 'conversation-a',
        loading: false
      }
    })
    await flushPromises()
    expect(scrollTo).toHaveBeenCalled()

    const scroller = wrapper.get('.message-thread__scroller').element as HTMLElement
    Object.defineProperties(scroller, {
      scrollHeight: { configurable: true, value: 1200 },
      clientHeight: { configurable: true, value: 400 },
      scrollTop: { configurable: true, writable: true, value: 180 }
    })
    await wrapper.get('.message-thread__scroller').trigger('scroll')
    scrollTo.mockClear()

    await wrapper.setProps({
      messages: [messages[0], { ...messages[1], content: '第一轮回答，继续输出' }]
    })
    await flushPromises()

    expect(scrollTo).not.toHaveBeenCalled()
  })

  it('keeps the rail hidden for a short conversation', async () => {
    const wrapper = mount(MessageThread, {
      props: {
        conversationId: 'conversation-a',
        loading: false,
        tasksByMessage: {},
        stageByMessage: {},
        messages: [
          message('user-1', 'user', '继续完善工作台交互', 'completed', [
            {
              id: 'attachment-1',
              name: 'Icon.vue',
              mime_type: 'text/plain',
              size_bytes: 100,
              data_url: 'data:text/plain;base64,QQ=='
            },
            {
              id: 'attachment-2',
              name: 'useWorkbench.ts',
              mime_type: 'text/plain',
              size_bytes: 100,
              data_url: 'data:text/plain;base64,QQ=='
            },
            {
              id: 'attachment-3',
              name: 'MessageThread.vue',
              mime_type: 'text/plain',
              size_bytes: 100,
              data_url: 'data:text/plain;base64,QQ=='
            }
          ]),
          message('assistant-1', 'assistant', '正在调整真实消息刻度和悬浮预览。'),
          message('user-2', 'user', '第二轮问题'),
          message('assistant-2', 'assistant', '第二轮回答')
        ]
      }
    })
    await flushPromises()

    expect(wrapper.findAll('.conversation-rail__mark')).toHaveLength(0)
    expect(wrapper.find('.conversation-rail__ticks').exists()).toBe(false)
    const firstPreview = wrapper.findAll('.conversation-rail__preview')[0]
    if (firstPreview?.exists()) {
    expect(firstPreview.text()).toContain('继续完善工作台交互')
    expect(firstPreview.text()).toContain('正在调整真实消息刻度和悬浮预览。')
    expect(firstPreview.text()).toContain('Icon.vue')
    expect(firstPreview.text()).toContain('useWorkbench.ts')
    expect(firstPreview.text()).toContain('+1')
    }
    expect(wrapper.text()).not.toContain('09:13')
    expect(wrapper.find('button[title="复制消息"]').exists()).toBe(true)
  })

  it('shows one mark per user turn once the thread overflows', async () => {
    const messages = Array.from({ length: 4 }, (_, index) => [
      message(`user-${index + 1}`, 'user', `user turn ${index + 1}`),
      message(`assistant-${index + 1}`, 'assistant', `assistant turn ${index + 1}`)
    ]).flat()
    const wrapper = mount(MessageThread, {
      props: {
        conversationId: 'conversation-a',
        loading: false,
        tasksByMessage: {},
        stageByMessage: {},
        messages
      }
    })
    await flushPromises()

    const scroller = wrapper.get('.message-thread__scroller').element as HTMLElement
    Object.defineProperties(scroller, {
      scrollHeight: { configurable: true, value: 1400 },
      clientHeight: { configurable: true, value: 400 },
      scrollTop: { configurable: true, writable: true, value: 0 }
    })
    wrapper.findAll('[data-message-id]').forEach((element, index) => {
      Object.defineProperty(element.element, 'offsetTop', { configurable: true, value: index * 150 })
      Object.defineProperty(element.element, 'offsetHeight', { configurable: true, value: 100 })
    })
    await wrapper.get('.message-thread__scroller').trigger('scroll')
    await flushPromises()

    expect(wrapper.findAll('.conversation-rail__mark')).toHaveLength(4)
    expect(wrapper.findAll('.conversation-rail__mark')[0].attributes('style')).toContain('top:')
  })

  it('collapses generation details after the rendered answer completes', async () => {
    const assistant = {
      ...message('assistant-1', 'assistant', '回答完成'),
      reasoning_summary: '先分析问题再作答'
    }
    const wrapper = mount(MessageThread, {
      props: {
        conversationId: 'conversation-a',
        loading: false,
        messages: [message('user-1', 'user', '问题'), assistant],
        tasksByMessage: {
          'assistant-1': [{ id: 'answer', title: '回答已生成', status: 'completed' }]
        },
        stageByMessage: { 'assistant-1': 'completed' }
      }
    })
    await flushPromises()

    expect(wrapper.find('.generation-activity').exists()).toBe(false)
    expect(wrapper.find('.reasoning-panel').exists()).toBe(false)
    expect(wrapper.text()).toContain('回答完成')
  })

  it('shows a return-to-bottom button after the user scrolls away', async () => {
    const wrapper = mount(MessageThread, {
      props: {
        conversationId: 'conversation-a',
        loading: false,
        messages: [message('user-1', 'user', '问题'), message('assistant-1', 'assistant', '回答')],
        tasksByMessage: {},
        stageByMessage: {}
      }
    })
    await flushPromises()

    const scroller = wrapper.get('.message-thread__scroller').element as HTMLElement
    Object.defineProperties(scroller, {
      scrollHeight: { configurable: true, value: 1200 },
      clientHeight: { configurable: true, value: 400 },
      scrollTop: { configurable: true, writable: true, value: 180 }
    })
    await wrapper.get('.message-thread__scroller').trigger('scroll')

    const button = wrapper.get('button[aria-label="回到底部"]')
    await button.trigger('click')
    expect(scrollTo).toHaveBeenLastCalledWith({ top: 1200, behavior: 'smooth' })
  })
})
