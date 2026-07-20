import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import UserConsolePage from '../UserConsolePage.vue'

describe('UserConsolePage', () => {
  it('renders a compact heading with page actions and hides hero content', () => {
    const wrapper = mount(UserConsolePage, {
      props: {
        kicker: 'Workspace',
        title: 'Usage records',
        description: 'Legacy hero description',
        compact: true,
      },
      slots: {
        headerActions: '<button type="button">Export</button>',
        heroNotes: '<span>Legacy note</span>',
        heroAside: '<span>Legacy summary</span>',
        default: '<div>Page content</div>',
      },
    })

    expect(wrapper.text()).toContain('Workspace')
    expect(wrapper.text()).toContain('Usage records')
    expect(wrapper.text()).toContain('Export')
    expect(wrapper.text()).toContain('Page content')
    expect(wrapper.text()).not.toContain('Legacy hero description')
    expect(wrapper.text()).not.toContain('Legacy note')
    expect(wrapper.text()).not.toContain('Legacy summary')
    expect(wrapper.find('.user-console-page__head--compact').exists()).toBe(true)
  })

  it('keeps the original hero layout as the default variant', () => {
    const wrapper = mount(UserConsolePage, {
      props: {
        kicker: 'Administration',
        title: 'Redeem codes',
        description: 'Manage redeem codes.',
      },
      slots: {
        heroNotes: '<span>Usage note</span>',
        heroAside: '<span>Summary</span>',
      },
    })

    expect(wrapper.text()).toContain('Manage redeem codes.')
    expect(wrapper.text()).toContain('Usage note')
    expect(wrapper.text()).toContain('Summary')
    expect(wrapper.find('.user-console-page__head--compact').exists()).toBe(false)
    expect(wrapper.find('.user-console-page__hero').exists()).toBe(true)
  })
})
