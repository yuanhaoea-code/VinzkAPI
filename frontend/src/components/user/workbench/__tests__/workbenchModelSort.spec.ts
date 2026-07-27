import { describe, expect, it } from 'vitest'
import type { WorkbenchModel } from '@/api/workbench'
import { compareWorkbenchModels } from '../workbenchModelSort'

function model(id: string, provider = 'openai'): WorkbenchModel {
  return {
    id,
    display_name: id,
    provider,
    provider_label: provider,
    default_visible: false,
    added: false,
    available: true,
    sort_order: 1000,
    reasoning_presets: ['fast', 'standard', 'deep']
  }
}

describe('compareWorkbenchModels', () => {
  it('sorts mainstream model versions from newest to oldest', () => {
    const ids = ['gpt-5.2', 'gpt-5.4-mini', 'gpt-5.6-sol', 'gpt-5.5', 'gpt-5.4']
    expect(ids.map(id => model(id)).sort(compareWorkbenchModels).map(item => item.id)).toEqual([
      'gpt-5.6-sol',
      'gpt-5.5',
      'gpt-5.4',
      'gpt-5.4-mini',
      'gpt-5.2'
    ])
  })

  it('keeps versioned models ahead of utility aliases', () => {
    const items = [model('codex-auto-review'), model('gpt-5.3-codex-spark')]
    expect(items.sort(compareWorkbenchModels).map(item => item.id)).toEqual([
      'gpt-5.3-codex-spark',
      'codex-auto-review'
    ])
  })
})
