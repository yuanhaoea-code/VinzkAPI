import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ReasoningPanel from '../ReasoningPanel.vue'

describe('ReasoningPanel', () => {
  it('opens while the public reasoning summary streams and closes when the answer begins', async () => {
    const wrapper = mount(ReasoningPanel, {
      props: { streaming: true, answerStarted: false, summary: '', visible: true }
    })

    expect(wrapper.get('.reasoning-trigger').attributes('aria-expanded')).toBe('true')
    expect(wrapper.get('.reasoning-content').isVisible()).toBe(true)
    expect(wrapper.find('.reasoning-inline-dots').exists()).toBe(true)

    await wrapper.setProps({ summary: '先分析用户目标' })
    expect(wrapper.text()).toContain('先分析用户目标')
    expect(wrapper.find('.reasoning-cursor').exists()).toBe(true)

    await wrapper.setProps({ answerStarted: true })
    expect(wrapper.get('.reasoning-trigger').attributes('aria-expanded')).toBe('false')
    await vi.waitFor(() => expect(wrapper.find('.reasoning-content').exists()).toBe(false))
    expect(wrapper.find('.reasoning-cursor').exists()).toBe(false)
  })
})
