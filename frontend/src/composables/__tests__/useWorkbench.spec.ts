import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import type {
  WorkbenchConversation,
  WorkbenchKey,
  WorkbenchMessage,
  WorkbenchModel,
  WorkbenchStreamEvent,
  WorkbenchTurn
} from '@/api/workbench'

const workbenchAPIMock = vi.hoisted(() => ({
  getModels: vi.fn(),
  addModel: vi.fn(),
  addModels: vi.fn(),
  hideModel: vi.fn(),
  listConversations: vi.fn(),
  createConversation: vi.fn(),
  getConversation: vi.fn(),
  updateConversation: vi.fn(),
  deleteConversation: vi.fn(),
  createMessage: vi.fn(),
  cancelGeneration: vi.fn(),
  streamGeneration: vi.fn(),
}))

vi.mock('@/api/workbench', async () => {
  const actual = await vi.importActual<typeof import('@/api/workbench')>('@/api/workbench')
  return { ...actual, workbenchAPI: workbenchAPIMock }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() }),
}))

import { resetWorkbenchWorkspace, useWorkbench } from '@/composables/useWorkbench'

const availableModel: WorkbenchModel = {
  id: 'gpt-5.6-sol',
  display_name: 'GPT-5.6',
  provider: 'openai',
  provider_label: 'OpenAI',
  default_visible: true,
  added: true,
  available: true,
  binding_id: 'binding-a',
  sort_order: 10,
  reasoning_presets: ['fast', 'standard', 'deep'],
}

const alternateModel: WorkbenchModel = {
  ...availableModel,
  id: 'gpt-5.5',
  display_name: 'GPT-5.5',
  binding_id: 'binding-b',
  sort_order: 9,
}

function keyWithModels(models: WorkbenchModel[], defaultModelId = models[0]?.id ?? ''): WorkbenchKey {
  return {
    api_key_id: 42,
    key_name: 'OpenAI 主密钥',
    group_id: 1,
    group_name: 'GPT',
    platform: 'openai',
    provider_label: 'OpenAI',
    rate: 1,
    default_model_id: defaultModelId,
    default_model_name: models.find(model => model.id === defaultModelId)?.display_name ?? '',
    models
  }
}

function conversation(id: string, bindingId = 'binding-a'): WorkbenchConversation {
  return {
    id,
    model_binding_id: bindingId,
    title: id,
    reasoning_preset: 'standard',
    created_at: '2026-07-20T00:00:00Z',
    updated_at: '2026-07-20T00:00:00Z',
    messages: [],
  }
}

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolver, rejecter) => {
    resolve = resolver
    reject = rejecter
  })
  return { promise, resolve, reject }
}

function mountWorkbench() {
  let workbench!: ReturnType<typeof useWorkbench>
  const Harness = defineComponent({
    setup() {
      workbench = useWorkbench()
      return () => h('div')
    }
  })
  return { wrapper: mount(Harness), workbench: () => workbench }
}

function workbenchMessage(
  id: string,
  conversationId: string,
  role: WorkbenchMessage['role'],
  status: WorkbenchMessage['status']
): WorkbenchMessage {
  return {
    id,
    conversation_id: conversationId,
    sequence: role === 'user' ? 1 : 2,
    role,
    content: role === 'user' ? `request-${conversationId}` : '',
    status,
    created_at: '2026-07-20T00:00:00Z',
    updated_at: '2026-07-20T00:00:00Z',
  }
}

function turn(conversationId: string): WorkbenchTurn {
  return {
    conversation: conversation(conversationId),
    user_message: workbenchMessage(`user-${conversationId}`, conversationId, 'user', 'completed'),
    assistant_message: workbenchMessage(`assistant-${conversationId}`, conversationId, 'assistant', 'pending'),
  }
}

describe('useWorkbench', () => {
  beforeEach(() => {
    resetWorkbenchWorkspace()
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useRealTimers()
  })

  afterEach(() => {
    resetWorkbenchWorkspace()
    vi.useRealTimers()
  })

  it('initializes only once when the workbench page remounts', async () => {
    workbenchAPIMock.getModels.mockResolvedValue({ selected: [availableModel], available: [] })
    workbenchAPIMock.listConversations.mockResolvedValue([conversation('first')])
    workbenchAPIMock.getConversation.mockResolvedValue(conversation('first'))

    const first = useWorkbench()
    await first.initialize()
    const second = useWorkbench()
    await second.initialize()

    expect(second).toBe(first)
    expect(workbenchAPIMock.getModels).toHaveBeenCalledTimes(1)
    expect(workbenchAPIMock.listConversations).toHaveBeenCalledTimes(1)
    expect(workbenchAPIMock.getConversation).toHaveBeenCalledTimes(1)
  })

  it('keeps streaming after the workbench page unmounts and updates the remounted page', async () => {
    vi.useFakeTimers()
    const stream = deferred<void>()
    let handleEvent!: (event: WorkbenchStreamEvent) => void
    let streamSignal!: AbortSignal
    const completed = conversation('first')
    const completedTurn = turn('first')
    completed.messages = [
      completedTurn.user_message,
      { ...completedTurn.assistant_message, content: 'answer after return', status: 'completed' }
    ]
    workbenchAPIMock.getConversation
      .mockResolvedValueOnce(conversation('first'))
      .mockResolvedValueOnce(completed)
    workbenchAPIMock.createMessage.mockResolvedValue(completedTurn)
    workbenchAPIMock.streamGeneration.mockImplementation((
      _messageId: string,
      onEvent: (event: WorkbenchStreamEvent) => void,
      signal: AbortSignal
    ) => {
      handleEvent = onEvent
      streamSignal = signal
      signal.addEventListener('abort', () => {
        stream.reject(new DOMException('Aborted', 'AbortError'))
      }, { once: true })
      return stream.promise
    })

    const firstMount = mountWorkbench()
    const first = firstMount.workbench()
    first.models.value = { selected: [availableModel], available: [] }
    first.selectedBindingId.value = 'binding-a'
    await first.openConversation('first')
    await first.sendMessage('run the task')

    firstMount.wrapper.unmount()
    const secondMount = mountWorkbench()
    const second = secondMount.workbench()

    expect(second).toBe(first)
    expect(streamSignal.aborted).toBe(false)
    expect(second.streamingConversationIds.value).toEqual(['first'])
    expect(second.canSend.value).toBe(false)
    expect(workbenchAPIMock.cancelGeneration).not.toHaveBeenCalled()

    await second.sendMessage('duplicate request')
    expect(workbenchAPIMock.createMessage).toHaveBeenCalledTimes(1)

    handleEvent({ name: 'workbench.thinking.started', data: {} })
    handleEvent({ name: 'workbench.answer.delta', data: { delta: 'answer after return' } })
    handleEvent({ name: 'workbench.generation.completed', data: {} })
    await vi.advanceTimersByTimeAsync(3000)
    stream.resolve()
    await flushPromises()

    expect(second.streamingConversationIds.value).toEqual([])
    expect(second.messages.value.at(-1)).toMatchObject({
      content: 'answer after return',
      status: 'completed'
    })
    expect(second.canSend.value).toBe(true)
    expect(workbenchAPIMock.cancelGeneration).not.toHaveBeenCalled()
  })

  it('ignores a slower conversation response after a newer selection wins', async () => {
    const first = deferred<WorkbenchConversation>()
    const second = deferred<WorkbenchConversation>()
    workbenchAPIMock.getConversation.mockImplementation((id: string) => (
      id === 'first' ? first.promise : second.promise
    ))
    const workbench = useWorkbench()
    workbench.models.value = { selected: [availableModel], available: [] }
    workbench.selectedBindingId.value = 'binding-a'

    const firstOpen = workbench.openConversation('first')
    const secondOpen = workbench.openConversation('second')
    second.resolve(conversation('second'))
    await secondOpen
    first.resolve(conversation('first'))
    await firstOpen

    expect(workbench.activeConversation.value?.id).toBe('second')
    expect(workbench.loadingConversation.value).toBe(false)
  })

  it('keeps an available selection when a saved conversation points to an unavailable binding', async () => {
    workbenchAPIMock.getConversation.mockResolvedValue(conversation('legacy', 'deleted-binding'))
    const workbench = useWorkbench()
    workbench.models.value = { selected: [availableModel], available: [] }
    workbench.selectedBindingId.value = 'binding-a'

    await workbench.openConversation('legacy')

    expect(workbench.activeConversation.value?.id).toBe('legacy')
    expect(workbench.selectedBindingId.value).toBe('binding-a')
    expect(workbench.canSend.value).toBe(true)
  })

  it('starts streaming immediately while presenting each initial stage in order', async () => {
    vi.useFakeTimers()
    let handleEvent: ((event: WorkbenchStreamEvent) => void) | undefined
    workbenchAPIMock.getConversation.mockResolvedValue(conversation('first'))
    workbenchAPIMock.createMessage.mockResolvedValue(turn('first'))
    workbenchAPIMock.streamGeneration.mockImplementation((
      _messageId: string,
      onEvent: (event: WorkbenchStreamEvent) => void
    ) => {
      handleEvent = onEvent
      return new Promise(() => {})
    })
    workbenchAPIMock.cancelGeneration.mockResolvedValue(undefined)
    const workbench = useWorkbench()
    workbench.models.value = { selected: [availableModel], available: [] }
    workbench.selectedBindingId.value = 'binding-a'

    await workbench.openConversation('first')
    await workbench.sendMessage('run the task')

    expect(workbenchAPIMock.streamGeneration).toHaveBeenCalledWith(
      'assistant-first',
      expect.any(Function),
      expect.any(AbortSignal)
    )
    expect(workbench.stageByMessage.value['assistant-first']).toBe('preparing')
    expect(workbench.tasksByMessage.value['assistant-first']).toEqual([
      { id: 'accepted', title: '正在接收请求', status: 'in_progress' }
    ])

    handleEvent?.({ name: 'workbench.thinking.started', data: {} })
    handleEvent?.({ name: 'workbench.answer.started', data: {} })

    await vi.advanceTimersByTimeAsync(999)
    expect(workbench.stageByMessage.value['assistant-first']).toBe('preparing')

    await vi.advanceTimersByTimeAsync(1)
    expect(workbench.stageByMessage.value['assistant-first']).toBe('reasoning')

    await vi.advanceTimersByTimeAsync(719)
    expect(workbench.stageByMessage.value['assistant-first']).toBe('reasoning')

    await vi.advanceTimersByTimeAsync(1)
    expect(workbench.stageByMessage.value['assistant-first']).toBe('answering')

    await vi.advanceTimersByTimeAsync(320)
    await workbench.stopGeneration('first')
  })

  it('allows switching the next-turn model while the active conversation is generating', async () => {
    vi.useFakeTimers()
    workbenchAPIMock.getConversation.mockResolvedValue(conversation('first'))
    workbenchAPIMock.createMessage.mockResolvedValue(turn('first'))
    workbenchAPIMock.streamGeneration.mockReturnValue(new Promise(() => {}))
    workbenchAPIMock.updateConversation.mockResolvedValue(conversation('first', 'binding-b'))
    workbenchAPIMock.cancelGeneration.mockResolvedValue(undefined)
    const workbench = useWorkbench()
    workbench.models.value = { selected: [availableModel, alternateModel], available: [] }
    workbench.selectedBindingId.value = 'binding-a'

    await workbench.openConversation('first')
    await workbench.sendMessage('run the task')
    expect(workbench.streamingConversationIds.value).toEqual(['first'])

    await workbench.selectModel('binding-b')

    expect(workbenchAPIMock.updateConversation).toHaveBeenCalledWith('first', {
      model_binding_id: 'binding-b'
    })
    expect(workbench.selectedBindingId.value).toBe('binding-b')
    await workbench.stopGeneration('first')
  })

  it('falls back to the same key default after removing the selected model', async () => {
    const defaultModel = {
      ...availableModel,
      id: 'gpt-5.4-mini',
      display_name: 'GPT-5.4 Mini',
      binding_id: 'binding-default',
      api_key_id: 42
    }
    const selectedModel = { ...alternateModel, api_key_id: 42 }
    const initialKey = keyWithModels([defaultModel, selectedModel], defaultModel.id)
    const refreshedKey = keyWithModels([defaultModel], defaultModel.id)
    workbenchAPIMock.hideModel.mockResolvedValue(undefined)
    workbenchAPIMock.getModels.mockResolvedValue({
      keys: [refreshedKey],
      selected: [defaultModel],
      available: []
    })
    workbenchAPIMock.updateConversation.mockResolvedValue(conversation('first', defaultModel.binding_id))

    const workbench = useWorkbench()
    workbench.models.value = {
      keys: [initialKey],
      selected: [defaultModel, selectedModel],
      available: []
    }
    workbench.selectedBindingId.value = selectedModel.binding_id!
    workbench.conversations.value = [conversation('first', selectedModel.binding_id)]
    await workbench.openConversation('first')

    await workbench.hideModel(selectedModel.binding_id!)

    expect(workbenchAPIMock.hideModel).toHaveBeenCalledWith(selectedModel.binding_id)
    expect(workbench.selectedBindingId.value).toBe(defaultModel.binding_id)
    expect(workbenchAPIMock.updateConversation).toHaveBeenCalledWith('first', {
      model_binding_id: defaultModel.binding_id
    })
  })

  it('keeps parallel streams isolated while navigating between conversations', async () => {
    vi.useFakeTimers()
    const handlers = new Map<string, (event: WorkbenchStreamEvent) => void>()
    workbenchAPIMock.getConversation.mockImplementation((id: string) => (
      Promise.resolve(conversation(id))
    ))
    workbenchAPIMock.createMessage.mockImplementation((id: string) => Promise.resolve(turn(id)))
    workbenchAPIMock.streamGeneration.mockImplementation((
      messageId: string,
      onEvent: (event: WorkbenchStreamEvent) => void
    ) => {
      handlers.set(messageId, onEvent)
      return new Promise(() => {})
    })
    workbenchAPIMock.cancelGeneration.mockResolvedValue(undefined)
    const workbench = useWorkbench()
    workbench.models.value = { selected: [availableModel], available: [] }
    workbench.selectedBindingId.value = 'binding-a'

    await workbench.openConversation('first')
    await workbench.sendMessage('first task')
    await workbench.openConversation('second')
    await workbench.sendMessage('second task')

    expect(workbench.activeConversation.value?.id).toBe('second')
    expect(new Set(workbench.streamingConversationIds.value)).toEqual(new Set(['first', 'second']))

    await vi.advanceTimersByTimeAsync(1000)
    handlers.get('assistant-first')?.({
      name: 'workbench.answer.delta',
      data: { delta: 'answer-first' }
    })
    handlers.get('assistant-second')?.({
      name: 'response.output_text.delta',
      data: { delta: 'answer-second' }
    })
    await vi.advanceTimersByTimeAsync(1200)

    await workbench.openConversation('first')
    expect(workbench.messages.value.at(-1)?.content).toBe('answer-first')
    await workbench.openConversation('second')
    expect(workbench.messages.value.at(-1)?.content).toBe('answer-second')

    await workbench.stopGeneration('first')
    expect(workbench.streamingConversationIds.value).toEqual(['second'])
    await workbench.stopGeneration('second')
  })
})
