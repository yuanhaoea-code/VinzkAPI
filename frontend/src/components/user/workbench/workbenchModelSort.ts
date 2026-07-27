import type { WorkbenchModel } from '@/api/workbench'

const PROVIDER_ORDER: Record<string, number> = {
  openai: 0,
  anthropic: 1,
  gemini: 2,
  grok: 3
}

export function compareWorkbenchModels(a: WorkbenchModel, b: WorkbenchModel): number {
  const providerDifference = (PROVIDER_ORDER[a.provider] ?? 99) - (PROVIDER_ORDER[b.provider] ?? 99)
  if (providerDifference !== 0) return providerDifference

  const aVersion = modelVersion(a.id)
  const bVersion = modelVersion(b.id)
  if (aVersion && bVersion) {
    if (aVersion[0] !== bVersion[0]) return bVersion[0] - aVersion[0]
    if (aVersion[1] !== bVersion[1]) return bVersion[1] - aVersion[1]
    if (a.id.length !== b.id.length) return a.id.length - b.id.length
  } else if (aVersion) {
    return -1
  } else if (bVersion) {
    return 1
  }

  if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
  return a.display_name.localeCompare(b.display_name, 'zh-CN', { numeric: true })
}

function modelVersion(modelId: string): [number, number] | null {
  const numbers = modelId.match(/\d+/g)
  if (!numbers?.length) return null
  return [Number(numbers[0]) || 0, Number(numbers[1]) || 0]
}
