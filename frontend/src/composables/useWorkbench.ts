import { computed, effectScope, shallowRef, type EffectScope } from 'vue'
import { workbenchAPI } from '@/api/workbench'
import type {
  WorkbenchAttachment,
  WorkbenchConversation,
  WorkbenchKey,
  WorkbenchMessage,
  WorkbenchModel,
  WorkbenchModels,
  WorkbenchReasoningPreset,
  WorkbenchStreamEvent
} from '@/api/workbench'
import type {
  WorkbenchGenerationStage,
  WorkbenchTask
} from '@/components/user/workbench/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const EMPTY_MODELS: WorkbenchModels = { selected: [], available: [] }
const ACCEPTED_DWELL_MS = 1000
const REASONING_PLACEHOLDER_MS = 420
const REASONING_DWELL_MS = 720
const ANSWER_PLACEHOLDER_MS = 320

function findWorkbenchModelByBinding(models: WorkbenchModels, bindingID: string): WorkbenchModel | null {
  if (!bindingID) return null
  for (const key of models.keys ?? []) {
    const model = key.models.find(item => item.binding_id === bindingID)
    if (model) return model
  }
  return models.selected.find(item => item.binding_id === bindingID) ?? null
}

function firstWorkbenchBindingID(models: WorkbenchModels): string {
  for (const key of models.keys ?? []) {
    const model = key.models.find(item => item.id === key.default_model_id && item.available)
      ?? key.models.find(item => item.available)
    if (model?.binding_id) return model.binding_id
  }
  return models.selected.find(model => model.available)?.binding_id ?? ''
}

function defaultWorkbenchBindingIDForKey(models: WorkbenchModels, apiKeyId: number): string {
  const key = (models.keys ?? []).find(item => item.api_key_id === apiKeyId)
  if (!key) return ''
  const model = key.models.find(item => item.id === key.default_model_id && item.available)
    ?? key.models.find(item => item.available)
  return model?.binding_id ?? ''
}

interface GenerationRun {
  conversationId: string
  messageId: string
  runId: number
  controller: AbortController
}

function createWorkbench() {
  const appStore = useAppStore()
  const models = shallowRef<WorkbenchModels>(EMPTY_MODELS)
  const conversations = shallowRef<WorkbenchConversation[]>([])
  const conversationCache = shallowRef<Record<string, WorkbenchConversation>>({})
  const activeConversationId = shallowRef('')
  const tasksByMessage = shallowRef<Record<string, WorkbenchTask[]>>({})
  const stageByMessage = shallowRef<Record<string, WorkbenchGenerationStage>>({})
  const selectedBindingId = shallowRef('')
  const reasoningPreset = shallowRef<WorkbenchReasoningPreset>('standard')
  const loading = shallowRef(true)
  const loadingConversation = shallowRef(false)
  const managingModels = shallowRef(false)
  const creatingConversation = shallowRef(false)
  const submittingByConversation = shallowRef<Record<string, boolean>>({})
  const updatingByConversation = shallowRef<Record<string, boolean>>({})
  const streamingByConversation = shallowRef<Record<string, string>>({})
  const controllers = new Map<string, AbortController>()
  const runIds = new Map<string, number>()
  const stageStartedAt = new Map<string, number>()
  let conversationLoadRunId = 0
  let initialized = false
  let initializePromise: Promise<void> | null = null

  const activeConversation = computed<WorkbenchConversation | null>(() => {
    const id = activeConversationId.value
    if (!id) return null
    return conversationCache.value[id] ?? conversations.value.find(item => item.id === id) ?? null
  })
  const messages = computed<WorkbenchMessage[]>(() => activeConversation.value?.messages ?? [])
  const selectedKey = computed<WorkbenchKey | null>(() => (
    (models.value.keys ?? []).find(key => key.models.some(model => model.binding_id === selectedBindingId.value))
    ?? null
  ))
  const selectedModel = computed<WorkbenchModel | null>(() =>
    selectedKey.value?.models.find(model => model.binding_id === selectedBindingId.value)
    ?? models.value.selected.find(model => model.binding_id === selectedBindingId.value)
    ?? null
  )
  const streamingMessageId = computed(() => (
    activeConversationId.value ? streamingByConversation.value[activeConversationId.value] ?? '' : ''
  ))
  const streamingConversationIds = computed(() => Object.keys(streamingByConversation.value))
  const activeConversationUpdating = computed(() => {
    const id = activeConversationId.value
    return Boolean(id && updatingByConversation.value[id])
  })
  const activeConversationBusy = computed(() => {
    const id = activeConversationId.value
    return Boolean(
      id && (
        submittingByConversation.value[id] ||
        updatingByConversation.value[id] ||
        streamingByConversation.value[id]
      )
    )
  })
  const interactionBusy = computed(() => Boolean(
    loadingConversation.value || managingModels.value || creatingConversation.value || activeConversationBusy.value
  ))
  const canSend = computed(() => Boolean(
    selectedModel.value?.available && activeConversation.value && !activeConversationBusy.value
  ))

  function initialize(): Promise<void> {
    if (initialized) return Promise.resolve()
    if (initializePromise) return initializePromise

    loading.value = true
    initializePromise = (async () => {
      try {
        const [modelData, conversationData] = await Promise.all([
          workbenchAPI.getModels(),
          workbenchAPI.listConversations()
        ])
        models.value = modelData
        conversations.value = conversationData
        selectedBindingId.value = firstWorkbenchBindingID(modelData)
        if (conversationData.length > 0) {
          await openConversation(conversationData[0].id)
        } else if (selectedBindingId.value) {
          await newConversation()
        }
        initialized = true
      } catch (error) {
        appStore.showError(extractApiErrorMessage(error, '工作台加载失败'))
      } finally {
        loading.value = false
        initializePromise = null
      }
    })()
    return initializePromise
  }

  async function reloadModels(): Promise<void> {
    const data = await workbenchAPI.getModels()
    models.value = data
    const selectedStillExists = Boolean(findWorkbenchModelByBinding(data, selectedBindingId.value)?.available)
    if (!selectedStillExists) {
      selectedBindingId.value = firstWorkbenchBindingID(data)
    }
  }

  async function newConversation(): Promise<void> {
    if (!selectedBindingId.value || creatingConversation.value || managingModels.value) return
    conversationLoadRunId += 1
    creatingConversation.value = true
    try {
      const conversation = await workbenchAPI.createConversation({
        model_binding_id: selectedBindingId.value,
        reasoning_preset: reasoningPreset.value
      })
      const hydrated = { ...conversation, messages: conversation.messages ?? [] }
      cacheConversation(hydrated)
      updateConversationList(hydrated)
      activeConversationId.value = hydrated.id
    } catch (error) {
      appStore.showError(extractApiErrorMessage(error, '新建对话失败'))
    } finally {
      creatingConversation.value = false
    }
  }

  async function openConversation(id: string): Promise<void> {
    if (!id) return
    const cached = conversationCache.value[id]
    if (cached?.messages) {
      activeConversationId.value = id
      syncToolbar(cached)
      loadingConversation.value = false
      return
    }

    const loadRunId = ++conversationLoadRunId
    activeConversationId.value = id
    loadingConversation.value = true
    try {
      const conversation = await workbenchAPI.getConversation(id)
      if (loadRunId !== conversationLoadRunId) return
      cacheConversation(conversation)
      updateConversationList(conversation, false)
      syncToolbar(conversation)
    } catch (error) {
      if (loadRunId === conversationLoadRunId) {
        appStore.showError(extractApiErrorMessage(error, '无法打开对话'))
      }
    } finally {
      if (loadRunId === conversationLoadRunId) loadingConversation.value = false
    }
  }

  async function deleteConversation(id: string): Promise<void> {
    if (streamingByConversation.value[id] || submittingByConversation.value[id]) {
      appStore.showError('请先停止该对话的生成任务')
      return
    }
    try {
      await workbenchAPI.deleteConversation(id)
      const remaining = conversations.value.filter(item => item.id !== id)
      const nextCache = { ...conversationCache.value }
      delete nextCache[id]
      conversationCache.value = nextCache
      conversations.value = remaining
      if (activeConversationId.value !== id) return
      activeConversationId.value = ''
      if (remaining.length > 0) await openConversation(remaining[0].id)
      else if (selectedBindingId.value) await newConversation()
    } catch (error) {
      appStore.showError(extractApiErrorMessage(error, '删除对话失败'))
    }
  }

  async function selectModel(bindingId: string): Promise<void> {
    const model = findWorkbenchModelByBinding(models.value, bindingId)
    if (!model?.available) return
    const conversationId = activeConversationId.value
    if (!conversationId) {
      selectedBindingId.value = bindingId
      return
    }
    if (updatingByConversation.value[conversationId]) return

    const previous = selectedBindingId.value
    selectedBindingId.value = bindingId
    setConversationFlag(updatingByConversation, conversationId, true)
    try {
      const updated = await workbenchAPI.updateConversation(conversationId, {
        model_binding_id: bindingId
      })
      mergeConversation(updated)
      updateConversationList(updated)
    } catch (error) {
      if (activeConversationId.value === conversationId) selectedBindingId.value = previous
      appStore.showError(extractApiErrorMessage(error, '切换模型失败'))
    } finally {
      setConversationFlag(updatingByConversation, conversationId, false)
    }
  }

  async function selectKey(apiKeyId: number): Promise<void> {
    const key = (models.value.keys ?? []).find(item => item.api_key_id === apiKeyId)
    if (!key) return
    const model = key.models.find(item => item.id === key.default_model_id && item.available)
      ?? key.models.find(item => item.available)
    if (model?.binding_id) await selectModel(model.binding_id)
  }

  async function selectReasoningPreset(preset: WorkbenchReasoningPreset): Promise<void> {
    const conversationId = activeConversationId.value
    if (!conversationId) {
      reasoningPreset.value = preset
      return
    }
    if (updatingByConversation.value[conversationId]) return

    const previous = reasoningPreset.value
    reasoningPreset.value = preset
    setConversationFlag(updatingByConversation, conversationId, true)
    try {
      const updated = await workbenchAPI.updateConversation(conversationId, {
        reasoning_preset: preset
      })
      mergeConversation(updated)
      updateConversationList(updated)
    } catch (error) {
      if (activeConversationId.value === conversationId) reasoningPreset.value = previous
      appStore.showError(extractApiErrorMessage(error, '切换思考程度失败'))
    } finally {
      setConversationFlag(updatingByConversation, conversationId, false)
    }
  }

  async function addModel(modelId: string, apiKeyId: number): Promise<void> {
    await addModels([modelId], apiKeyId)
  }

  async function addModels(modelIds: string[], apiKeyId: number): Promise<void> {
    const normalized = [...new Set(modelIds.map(modelId => modelId.trim()).filter(Boolean))]
    if (normalized.length === 0 || !apiKeyId) return
    managingModels.value = true
    try {
      await workbenchAPI.addModels(normalized, apiKeyId)
      await reloadModels()
      appStore.showSuccess(normalized.length === 1 ? '模型已添加到工作台' : `已添加 ${normalized.length} 个模型`)
    } catch (error) {
      appStore.showError(extractApiErrorMessage(error, '添加模型失败'))
    } finally {
      managingModels.value = false
    }
  }

  async function hideModel(bindingId: string): Promise<void> {
    const key = (models.value.keys ?? []).find(item => (
      item.models.some(model => model.binding_id === bindingId)
    ))
    const removingSelectedModel = selectedBindingId.value === bindingId
    managingModels.value = true
    try {
      await workbenchAPI.hideModel(bindingId)
      const data = await workbenchAPI.getModels()
      models.value = data
      if (removingSelectedModel) {
        const fallbackBindingId = key
          ? defaultWorkbenchBindingIDForKey(data, key.api_key_id)
          : firstWorkbenchBindingID(data)
        selectedBindingId.value = fallbackBindingId
        if (fallbackBindingId && activeConversationId.value) {
          await selectModel(fallbackBindingId)
        }
      } else if (!findWorkbenchModelByBinding(data, selectedBindingId.value)?.available) {
        selectedBindingId.value = firstWorkbenchBindingID(data)
      }
      appStore.showSuccess('模型已从工作台移除')
    } catch (error) {
      appStore.showError(extractApiErrorMessage(error, '移除模型失败'))
    } finally {
      managingModels.value = false
    }
  }

  async function sendMessage(content: string, attachments: WorkbenchAttachment[] = []): Promise<void> {
    const text = content.trim()
    const conversation = activeConversation.value
    const bindingId = selectedBindingId.value
    const preset = reasoningPreset.value
    if ((!text && attachments.length === 0) || !conversation || !bindingId) return
    if (streamingByConversation.value[conversation.id] || submittingByConversation.value[conversation.id]) return

    const acceptedStartedAt = Date.now()
    setConversationFlag(submittingByConversation, conversation.id, true)
    try {
      const turn = await workbenchAPI.createMessage(conversation.id, {
        content: text,
        model_binding_id: bindingId,
        reasoning_preset: preset,
        attachments
      })
      const current = conversationCache.value[conversation.id] ?? conversation
      const nextConversation: WorkbenchConversation = {
        ...current,
        ...turn.conversation,
        messages: [...(current.messages ?? []), turn.user_message, turn.assistant_message]
      }
      cacheConversation(nextConversation)
      updateConversationList(turn.conversation)
      setTasks(turn.assistant_message.id, [
        { id: 'accepted', title: '正在接收请求', status: 'in_progress' }
      ])
      stageStartedAt.set(turn.assistant_message.id, acceptedStartedAt)
      setStage(turn.assistant_message.id, 'preparing')
      setStreamingMessage(conversation.id, turn.assistant_message.id)
      void runGeneration(conversation.id, turn.assistant_message.id)
    } catch (error) {
      appStore.showError(extractApiErrorMessage(error, '发送失败'))
    } finally {
      setConversationFlag(submittingByConversation, conversation.id, false)
    }
  }

  async function runGeneration(conversationId: string, messageId: string): Promise<void> {
    const runId = (runIds.get(conversationId) ?? 0) + 1
    runIds.set(conversationId, runId)
    const controller = new AbortController()
    controllers.set(conversationId, controller)
    const run: GenerationRun = { conversationId, messageId, runId, controller }

    try {
      let presentation = ensureDispatchStage(run)
      await workbenchAPI.streamGeneration(
        messageId,
        event => {
          presentation = presentation.then(() => handleStreamEvent(run, event))
        },
        controller.signal
      )
      await presentation
      if (!isCurrentRun(run)) return
      await refreshConversation(conversationId)
    } catch (error) {
      if (!isCurrentRun(run)) return
      if (!isAbortError(error)) {
        failGeneration(conversationId, messageId, extractApiErrorMessage(error, '生成失败'))
        appStore.showError(extractApiErrorMessage(error, '生成失败'))
      }
    } finally {
      if (isCurrentRun(run)) {
        clearStreamingMessage(conversationId, messageId)
        controllers.delete(conversationId)
      }
    }
  }

  async function stopGeneration(conversationId = activeConversationId.value): Promise<void> {
    const messageId = streamingByConversation.value[conversationId]
    if (!conversationId || !messageId) return
    runIds.set(conversationId, (runIds.get(conversationId) ?? 0) + 1)
    controllers.get(conversationId)?.abort()
    controllers.delete(conversationId)
    setMessageState(conversationId, messageId, { status: 'canceled', error_message: '生成已停止' })
    updateTask(messageId, { id: 'answer', title: '生成已停止', status: 'failed' })
    setStage(messageId, 'canceled')
    clearStreamingMessage(conversationId, messageId)
    try {
      await workbenchAPI.cancelGeneration(messageId)
    } catch {
      // Aborting the stream already cancels the local gateway request.
    }
  }

  async function stopAllGenerations(): Promise<void> {
    await Promise.all(Object.keys(streamingByConversation.value).map(id => stopGeneration(id)))
  }

  async function handleStreamEvent(run: GenerationRun, event: WorkbenchStreamEvent): Promise<void> {
    if (!isCurrentRun(run)) return
    const { conversationId, messageId } = run
    switch (event.name) {
      case 'workbench.request.accepted':
      case 'workbench.generation.started':
        await ensureDispatchStage(run)
        break
      case 'workbench.thinking.started':
        await ensureReasoningStage(run)
        break
      case 'workbench.thinking.delta':
      case 'response.reasoning_summary_text.delta':
        await ensureReasoningStage(run)
        appendReasoningSummary(conversationId, messageId, String(event.data.delta ?? ''))
        break
      case 'workbench.answer.started':
        await ensureAnswerStage(run)
        break
      case 'workbench.answer.delta':
      case 'response.output_text.delta':
        await ensureAnswerStage(run)
        appendMessageText(conversationId, messageId, String(event.data.delta ?? ''))
        break
      case 'response.completed':
      case 'response.done':
      case 'workbench.generation.completed':
        await ensureAnswerStage(run)
        completeGeneration(conversationId, messageId)
        break
      case 'workbench.generation.failed': {
        const canceled = event.data.status === 'canceled'
        if (canceled) {
          setMessageState(conversationId, messageId, { status: 'canceled', error_message: '生成已停止' })
          setStage(messageId, 'canceled')
        } else {
          failGeneration(conversationId, messageId, String(event.data.message ?? '生成失败'))
        }
        break
      }
      case 'workbench.task':
        if (String(event.data.id ?? '') !== 'dispatch') {
          updateTask(messageId, {
            id: String(event.data.id ?? 'task'),
            title: String(event.data.title ?? '处理中'),
            status: normalizeTaskStatus(event.data.status)
          })
        }
        break
    }
  }

  async function ensureDispatchStage(run: GenerationRun): Promise<void> {
    if (!isCurrentRun(run)) return
    const currentStage = stageByMessage.value[run.messageId]
    if (
      currentStage === 'dispatching' ||
      currentStage === 'reasoning' ||
      currentStage === 'answering' ||
      currentStage === 'completed' ||
      currentStage === 'failed' ||
      currentStage === 'canceled'
    ) return

    const acceptedStartedAt = stageStartedAt.get(run.messageId) ?? Date.now()
    await delay(Math.max(0, ACCEPTED_DWELL_MS - (Date.now() - acceptedStartedAt)))
    if (!isCurrentRun(run)) return
    setTasks(run.messageId, [
      { id: 'accepted', title: '已接收请求', status: 'completed' },
      { id: 'dispatch', title: '正在调用模型', status: 'in_progress' }
    ])
    setStage(run.messageId, 'dispatching')
  }

  async function ensureReasoningStage(run: GenerationRun): Promise<void> {
    if (!isCurrentRun(run)) return
    let currentStage = stageByMessage.value[run.messageId]
    if (currentStage === 'reasoning' || currentStage === 'answering' || currentStage === 'completed') return
    if (currentStage !== 'dispatching') {
      await ensureDispatchStage(run)
      currentStage = stageByMessage.value[run.messageId]
    }
    if (!isCurrentRun(run) || currentStage !== 'dispatching') return
    setTasks(run.messageId, [
      { id: 'accepted', title: '已接收请求', status: 'completed' },
      { id: 'dispatch', title: '模型已开始处理', status: 'completed' },
      { id: 'reasoning', title: '正在分析上下文与组织思路', status: 'in_progress' }
    ])
    setStage(run.messageId, 'reasoning')
    stageStartedAt.set(run.messageId, Date.now())
    await delay(REASONING_PLACEHOLDER_MS)
  }

  async function ensureAnswerStage(run: GenerationRun): Promise<void> {
    if (!isCurrentRun(run)) return
    let currentStage = stageByMessage.value[run.messageId]
    if (currentStage === 'answering' || currentStage === 'completed') return
    if (currentStage !== 'reasoning') {
      await ensureReasoningStage(run)
      currentStage = stageByMessage.value[run.messageId]
    }
    if (!isCurrentRun(run) || currentStage !== 'reasoning') return
    const reasoningStartedAt = stageStartedAt.get(run.messageId) ?? Date.now()
    await delay(Math.max(0, REASONING_DWELL_MS - (Date.now() - reasoningStartedAt)))
    if (!isCurrentRun(run)) return
    setTasks(run.messageId, [
      { id: 'accepted', title: '已接收请求', status: 'completed' },
      { id: 'dispatch', title: '模型已开始处理', status: 'completed' },
      { id: 'reasoning', title: '思考过程已生成', status: 'completed' },
      { id: 'answer', title: '正在生成回答', status: 'in_progress' }
    ])
    setStage(run.messageId, 'answering')
    await delay(ANSWER_PLACEHOLDER_MS)
  }

  function completeGeneration(conversationId: string, messageId: string): void {
    setMessageState(conversationId, messageId, { status: 'completed' })
    updateTask(messageId, { id: 'answer', title: '回答已生成', status: 'completed' })
    setStage(messageId, 'completed')
  }

  function failGeneration(conversationId: string, messageId: string, message: string): void {
    setMessageState(conversationId, messageId, { status: 'failed', error_message: message })
    updateTask(messageId, { id: 'answer', title: '生成失败', status: 'failed' })
    setStage(messageId, 'failed')
  }

  function appendMessageText(conversationId: string, messageId: string, delta: string): void {
    if (!delta) return
    updateMessages(conversationId, message =>
      message.id === messageId
        ? { ...message, content: message.content + delta, status: 'in_progress' }
        : message
    )
  }

  function appendReasoningSummary(conversationId: string, messageId: string, delta: string): void {
    if (!delta) return
    updateMessages(conversationId, message =>
      message.id === messageId
        ? { ...message, reasoning_summary: (message.reasoning_summary ?? '') + delta }
        : message
    )
  }

  function setMessageState(conversationId: string, messageId: string, patch: Partial<WorkbenchMessage>): void {
    if (!messageId) return
    updateMessages(conversationId, message => message.id === messageId ? { ...message, ...patch } : message)
  }

  function updateMessages(conversationId: string, update: (message: WorkbenchMessage) => WorkbenchMessage): void {
    const conversation = conversationCache.value[conversationId]
    if (!conversation) return
    cacheConversation({
      ...conversation,
      messages: (conversation.messages ?? []).map(update)
    })
  }

  function setTasks(messageId: string, tasks: WorkbenchTask[]): void {
    tasksByMessage.value = { ...tasksByMessage.value, [messageId]: tasks }
  }

  function updateTask(messageId: string, task: WorkbenchTask): void {
    const current = tasksByMessage.value[messageId] ?? []
    const exists = current.some(item => item.id === task.id)
    setTasks(messageId, exists
      ? current.map(item => item.id === task.id ? task : item)
      : [...current, task]
    )
  }

  function setStage(messageId: string, stage: WorkbenchGenerationStage): void {
    stageByMessage.value = { ...stageByMessage.value, [messageId]: stage }
  }

  function setStreamingMessage(conversationId: string, messageId: string): void {
    streamingByConversation.value = { ...streamingByConversation.value, [conversationId]: messageId }
  }

  function clearStreamingMessage(conversationId: string, messageId: string): void {
    if (streamingByConversation.value[conversationId] !== messageId) return
    const next = { ...streamingByConversation.value }
    delete next[conversationId]
    streamingByConversation.value = next
    stageStartedAt.delete(messageId)
  }

  function cacheConversation(conversation: WorkbenchConversation): void {
    conversationCache.value = { ...conversationCache.value, [conversation.id]: conversation }
  }

  function mergeConversation(conversation: WorkbenchConversation): void {
    const current = conversationCache.value[conversation.id]
    cacheConversation({
      ...current,
      ...conversation,
      messages: conversation.messages ?? current?.messages
    })
  }

  function updateConversationList(conversation: WorkbenchConversation, promote = true): void {
    const existingIndex = conversations.value.findIndex(item => item.id === conversation.id)
    const next = conversations.value.filter(item => item.id !== conversation.id)
    if (promote || existingIndex < 0) {
      conversations.value = [conversation, ...next]
      return
    }
    next.splice(existingIndex, 0, conversation)
    conversations.value = next
  }

  async function refreshConversation(id: string): Promise<void> {
    const conversation = await workbenchAPI.getConversation(id)
    cacheConversation(conversation)
    updateConversationList(conversation)
  }

  function syncToolbar(conversation: WorkbenchConversation): void {
    reasoningPreset.value = conversation.reasoning_preset
    const conversationModelAvailable = models.value.selected.some(
      model => model.binding_id === conversation.model_binding_id && model.available
    )
    const conversationModel = findWorkbenchModelByBinding(models.value, conversation.model_binding_id ?? '')
    if ((conversationModelAvailable || conversationModel?.available) && conversation.model_binding_id) {
      selectedBindingId.value = conversation.model_binding_id
      return
    }
    const currentModelAvailable = Boolean(findWorkbenchModelByBinding(models.value, selectedBindingId.value)?.available)
    if (!currentModelAvailable) {
      selectedBindingId.value = firstWorkbenchBindingID(models.value)
    }
  }

  function setConversationFlag(
    target: typeof submittingByConversation,
    conversationId: string,
    value: boolean
  ): void {
    const next = { ...target.value }
    if (value) next[conversationId] = true
    else delete next[conversationId]
    target.value = next
  }

  function isCurrentRun(run: GenerationRun): boolean {
    return runIds.get(run.conversationId) === run.runId &&
      streamingByConversation.value[run.conversationId] === run.messageId
  }

  function dispose(): void {
    for (const [conversationId, controller] of controllers) {
      runIds.set(conversationId, (runIds.get(conversationId) ?? 0) + 1)
      controller.abort()
    }
    controllers.clear()
    streamingByConversation.value = {}
    stageStartedAt.clear()
  }

  return {
    models,
    conversations,
    activeConversation,
    messages,
    tasksByMessage,
    stageByMessage,
    selectedBindingId,
    selectedKey,
    selectedModel,
    reasoningPreset,
    loading,
    loadingConversation,
    managingModels,
    creatingConversation,
    streamingMessageId,
    streamingConversationIds,
    activeConversationUpdating,
    activeConversationBusy,
    interactionBusy,
    canSend,
    initialize,
    reloadModels,
    newConversation,
    openConversation,
    deleteConversation,
    selectKey,
    selectModel,
    selectReasoningPreset,
    addModel,
    addModels,
    hideModel,
    sendMessage,
    stopGeneration,
    stopAllGenerations,
    dispose
  }
}

export type WorkbenchWorkspaceState = ReturnType<typeof createWorkbench>

let persistentWorkbench: WorkbenchWorkspaceState | null = null
let persistentWorkbenchScope: EffectScope | null = null

export function useWorkbench(): WorkbenchWorkspaceState {
  if (persistentWorkbench) return persistentWorkbench

  persistentWorkbenchScope = effectScope(true)
  const workbench = persistentWorkbenchScope.run(createWorkbench)
  if (!workbench) {
    persistentWorkbenchScope.stop()
    persistentWorkbenchScope = null
    throw new Error('Failed to initialize AI workbench')
  }

  persistentWorkbench = workbench
  return workbench
}

export function resetWorkbenchWorkspace(): void {
  persistentWorkbench?.dispose()
  persistentWorkbenchScope?.stop()
  persistentWorkbench = null
  persistentWorkbenchScope = null
}

function normalizeTaskStatus(value: unknown): WorkbenchTask['status'] {
  if (value === 'pending' || value === 'completed' || value === 'failed') return value
  return 'in_progress'
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

function delay(milliseconds: number): Promise<void> {
  if (milliseconds <= 0) return Promise.resolve()
  return new Promise(resolve => window.setTimeout(resolve, milliseconds))
}
