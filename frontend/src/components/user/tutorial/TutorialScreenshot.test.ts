import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import TutorialScreenshot from './TutorialScreenshot.vue'

const mountedWrappers: Array<ReturnType<typeof mount>> = []

function renderScreenshot() {
  const wrapper = mount(TutorialScreenshot, {
    attachTo: document.body,
    props: {
      src: '/tutorials/keys-page.png',
      alt: 'API 密钥页面',
      caption: 'API 密钥操作示例',
    },
  })
  mountedWrappers.push(wrapper)
  return wrapper
}

afterEach(() => {
  mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  document.body.classList.remove('modal-open')
  document.body.innerHTML = ''
})

describe('TutorialScreenshot', () => {
  it('opens inside the tutorial and closes from the top-right button', async () => {
    const wrapper = renderScreenshot()
    const trigger = wrapper.get('button[aria-label="查看大图：API 密钥页面"]')

    expect(wrapper.find('a').exists()).toBe(false)
    await trigger.trigger('click')
    await nextTick()

    const dialog = document.querySelector<HTMLElement>('[role="dialog"]')
    const closeButton = document.querySelector<HTMLButtonElement>('button[aria-label="关闭图片预览"]')
    expect(dialog?.getAttribute('aria-label')).toBe('图片预览：API 密钥页面')
    expect(closeButton).toBe(document.activeElement)
    expect(document.body.classList.contains('modal-open')).toBe(true)

    closeButton?.click()
    await nextTick()

    expect(document.querySelector('[role="dialog"]')).toBeNull()
    expect(document.activeElement).toBe(trigger.element)
    expect(document.body.classList.contains('modal-open')).toBe(false)
  })

  it('closes when Escape is pressed', async () => {
    const wrapper = renderScreenshot()
    await wrapper.get('button[aria-label="查看大图：API 密钥页面"]').trigger('click')

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()

    expect(document.querySelector('[role="dialog"]')).toBeNull()
  })

  it('closes when the dark backdrop is clicked', async () => {
    const wrapper = renderScreenshot()
    await wrapper.get('button[aria-label="查看大图：API 密钥页面"]').trigger('click')

    const dialog = document.querySelector<HTMLElement>('[role="dialog"]')
    dialog?.click()
    await nextTick()

    expect(document.querySelector('[role="dialog"]')).toBeNull()
  })
})
