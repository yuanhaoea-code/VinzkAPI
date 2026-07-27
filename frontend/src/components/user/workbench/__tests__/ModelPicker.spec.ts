import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { WorkbenchKey, WorkbenchModel } from '@/api/workbench'
import ModelPicker from '../ModelPicker.vue'

const IconStub = defineComponent({
  name: 'Icon',
  props: { name: String },
  setup: () => () => h('span')
})

function model(overrides: Partial<WorkbenchModel> = {}): WorkbenchModel {
  return {
    id: 'gpt-5.6',
    display_name: 'GPT-5.6',
    provider: 'openai',
    provider_label: 'OpenAI',
    default_visible: true,
    added: true,
    available: true,
    binding_id: 'binding-gpt-56',
    api_key_id: 42,
    key_name: 'gateway-key',
    group_name: 'main',
    sort_order: 1,
    reasoning_presets: ['fast', 'standard', 'deep'],
    ...overrides
  }
}

describe('ModelPicker', () => {
  it('selects a key independently and exposes the add-model entry for that key', async () => {
    const firstModel = model({ id: 'gpt-5.4-mini', display_name: 'GPT-5.4 Mini' })
    const secondModel = model({
      id: 'claude-3-5-haiku-20241022',
      display_name: 'Claude 3.5 Haiku',
      provider: 'anthropic',
      provider_label: 'Anthropic',
      binding_id: 'binding-claude',
      api_key_id: 77
    })
    const keys: WorkbenchKey[] = [
      {
        api_key_id: 42,
        key_name: 'OpenAI 主密钥',
        group_id: 1,
        group_name: 'GPT',
        platform: 'openai',
        provider_label: 'OpenAI',
        rate: 1,
        default_model_id: firstModel.id,
        default_model_name: firstModel.display_name,
        models: [firstModel]
      },
      {
        api_key_id: 77,
        key_name: 'Claude 主密钥',
        group_id: 2,
        group_name: 'Claude',
        platform: 'anthropic',
        provider_label: 'Anthropic',
        rate: 1,
        default_model_id: secondModel.id,
        default_model_name: secondModel.display_name,
        models: [secondModel]
      }
    ]
    const wrapper = mount(ModelPicker, {
      props: { keys, modelValue: firstModel.binding_id! },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('.key-trigger').trigger('click')
    await wrapper.findAll('.picker-option--key')[1].trigger('click')

    expect(wrapper.emitted('selectKey')).toEqual([[77]])
    await wrapper.get('.model-trigger').trigger('click')
    expect(wrapper.find('.model-menu__add').exists()).toBe(true)
  })

  it('emits a model change when an available model is selected', async () => {
    const selectedModel = model()
    const key: WorkbenchKey = {
      api_key_id: 42,
      key_name: 'OpenAI 主密钥',
      group_id: 1,
      group_name: 'GPT',
      platform: 'openai',
      provider_label: 'OpenAI',
      rate: 1,
      default_model_id: selectedModel.id,
      default_model_name: selectedModel.display_name,
      models: [selectedModel]
    }
    const wrapper = mount(ModelPicker, {
      props: {
        keys: [key],
        modelValue: 'binding-gpt-56'
      },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('.model-trigger').trigger('click')
    await wrapper.get('.model-option__select').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([['binding-gpt-56']])
    expect(wrapper.find('.model-menu').exists()).toBe(false)
  })

  it('opens the add-model dialog from the bottom of the model menu', async () => {
    const selectedModel = model()
    const key: WorkbenchKey = {
      api_key_id: 42,
      key_name: 'OpenAI 主密钥',
      group_id: 1,
      group_name: 'GPT',
      platform: 'openai',
      provider_label: 'OpenAI',
      rate: 1,
      default_model_id: selectedModel.id,
      default_model_name: selectedModel.display_name,
      models: [selectedModel],
      available_models: [selectedModel, model({ id: 'gpt-5.2', binding_id: undefined, added: false })]
    }
    const wrapper = mount(ModelPicker, {
      attachTo: document.body,
      props: { keys: [key], modelValue: selectedModel.binding_id! },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('.model-trigger').trigger('click')
    await wrapper.get('.model-menu__add').trigger('click')

    expect(document.body.textContent).toContain('添加模型')
    expect(document.body.textContent).toContain('OpenAI 主密钥')
    wrapper.unmount()
  })

  it('sorts default and manually added models from newest to oldest', async () => {
    const defaultModel = model({
      id: 'gpt-5.4-mini',
      display_name: 'GPT-5.4 Mini',
      binding_id: 'binding-default'
    })
    const key: WorkbenchKey = {
      api_key_id: 42,
      key_name: 'OpenAI 主密钥',
      group_id: 1,
      group_name: 'GPT',
      platform: 'openai',
      provider_label: 'OpenAI',
      rate: 1,
      default_model_id: defaultModel.id,
      default_model_name: defaultModel.display_name,
      models: [
        model({ id: 'gpt-5.2', display_name: 'GPT-5.2', binding_id: 'binding-52' }),
        model({ id: 'gpt-5.3-codex-spark', display_name: 'GPT-5.3 Codex Spark', binding_id: 'binding-53' }),
        defaultModel,
        model({ id: 'gpt-5.6-sol', display_name: 'GPT-5.6 Sol', binding_id: 'binding-56' })
      ]
    }
    const wrapper = mount(ModelPicker, {
      props: { keys: [key], modelValue: defaultModel.binding_id! },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('.model-trigger').trigger('click')
    const ids = wrapper.findAll('.model-option__select .picker-option__body small').map(item => item.text())

    expect(ids).toEqual(['GPT-5.4 Mini', 'gpt-5.6-sol', 'gpt-5.3-codex-spark', 'gpt-5.2'])
  })

  it('removes an explicit model without selecting it', async () => {
    const defaultModel = model({
      id: 'gpt-5.4-mini',
      display_name: 'GPT-5.4 Mini',
      binding_id: 'binding-default'
    })
    const removableModel = model({ id: 'gpt-5.6-sol', binding_id: 'binding-removable' })
    const key: WorkbenchKey = {
      api_key_id: 42,
      key_name: 'OpenAI 主密钥',
      group_id: 1,
      group_name: 'GPT',
      platform: 'openai',
      provider_label: 'OpenAI',
      rate: 1,
      default_model_id: defaultModel.id,
      default_model_name: defaultModel.display_name,
      models: [defaultModel, removableModel]
    }
    const wrapper = mount(ModelPicker, {
      props: { keys: [key], modelValue: defaultModel.binding_id! },
      global: { stubs: { Icon: IconStub } }
    })

    await wrapper.get('.model-trigger').trigger('click')
    expect(wrapper.findAll('.model-option__remove')).toHaveLength(1)
    await wrapper.get('.model-option__remove').trigger('click')

    expect(wrapper.emitted('removeModel')).toEqual([['binding-removable']])
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
