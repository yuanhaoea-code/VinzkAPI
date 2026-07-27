import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { WorkbenchKey, WorkbenchModel } from '@/api/workbench'
import AddWorkbenchModelsDialog from '../AddWorkbenchModelsDialog.vue'

const IconStub = defineComponent({
  name: 'Icon',
  props: { name: String },
  setup: () => () => h('span')
})

function model(id: string, added = false): WorkbenchModel {
  return {
    id,
    display_name: id.toUpperCase(),
    provider: 'openai',
    provider_label: 'OpenAI',
    default_visible: false,
    added,
    available: true,
    binding_id: added ? `binding-${id}` : undefined,
    api_key_id: 42,
    sort_order: 1000,
    reasoning_presets: ['fast', 'standard', 'deep']
  }
}

const apiKey: WorkbenchKey = {
  api_key_id: 42,
  key_name: '主密钥',
  group_id: 1,
  group_name: 'GPT',
  platform: 'openai',
  provider_label: 'OpenAI',
  rate: 1,
  default_model_id: 'gpt-5.4-mini',
  default_model_name: 'GPT-5.4 Mini',
  models: [model('gpt-5.4-mini', true)],
  available_models: [model('gpt-5.4-mini', true), model('gpt-5.3-codex-spark'), model('gpt-5.2')]
}

describe('AddWorkbenchModelsDialog', () => {
  it('keeps added models checked and emits all newly selected models together', async () => {
    const wrapper = mount(AddWorkbenchModelsDialog, {
      attachTo: document.body,
      props: { show: true, apiKey },
      global: { stubs: { Icon: IconStub } }
    })
    const checkboxes = document.body.querySelectorAll<HTMLInputElement>('.catalog-row input')

    expect(checkboxes[0].disabled).toBe(true)
    await checkboxes[1].click()
    await checkboxes[2].click()
    await document.body.querySelector<HTMLButtonElement>('.catalog-button--primary')?.click()

    expect(wrapper.emitted('confirm')).toEqual([[['gpt-5.3-codex-spark', 'gpt-5.2']]])
    wrapper.unmount()
  })

  it('filters the current key catalog by model id', async () => {
    const wrapper = mount(AddWorkbenchModelsDialog, {
      attachTo: document.body,
      props: { show: true, apiKey },
      global: { stubs: { Icon: IconStub } }
    })
    const input = document.body.querySelector<HTMLInputElement>('.catalog-search input')!
    await input.focus()
    await input.setRangeText('spark')
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await wrapper.vm.$nextTick()

    expect(document.body.querySelectorAll('.catalog-row')).toHaveLength(1)
    expect(document.body.textContent).toContain('gpt-5.3-codex-spark')
    wrapper.unmount()
  })
})
