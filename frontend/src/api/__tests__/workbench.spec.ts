import { beforeEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from '@/api/client'

const { refreshTokenMock } = vi.hoisted(() => ({ refreshTokenMock: vi.fn() }))

vi.mock('@/api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    request: vi.fn(),
  },
  buildApiUrl: (path: string) => `/api/v1${path}`,
}))

vi.mock('@/api/auth', () => ({
  refreshToken: refreshTokenMock,
}))

import {
  addWorkbenchModels,
  createWorkbenchAttachmentUpload,
  streamWorkbenchGeneration,
  uploadWorkbenchAttachment,
  type WorkbenchAttachmentUploadTicket
} from '@/api/workbench'

function streamResponse(chunks: string[]): Response {
  const encoder = new TextEncoder()
  return new Response(new ReadableStream({
    start(controller) {
      for (const chunk of chunks) controller.enqueue(encoder.encode(chunk))
      controller.close()
    },
  }), {
    status: 200,
    headers: { 'Content-Type': 'text/event-stream' },
  })
}

describe('workbench stream API', () => {
  beforeEach(() => {
    localStorage.clear()
    refreshTokenMock.mockReset()
    vi.unstubAllGlobals()
  })

  it('parses SSE events split across network chunks', async () => {
    localStorage.setItem('auth_token', 'access-token')
    const fetchMock = vi.fn().mockResolvedValue(streamResponse([
      'event: response.output_text.delta\ndata: {"type":"response.output_',
      'text.delta","delta":"完成"}\n\nevent: workbench.generation.completed\n',
      'data: {"message_id":"message-1"}\n\n',
    ]))
    vi.stubGlobal('fetch', fetchMock)
    const events: Array<{ name: string; data: Record<string, unknown> }> = []

    await streamWorkbenchGeneration('message-1', event => events.push(event))

    expect(events).toEqual([
      { name: 'response.output_text.delta', data: { type: 'response.output_text.delta', delta: '完成' } },
      { name: 'workbench.generation.completed', data: { message_id: 'message-1' } },
    ])
    expect(fetchMock.mock.calls[0][1]?.headers).toMatchObject({
      Authorization: 'Bearer access-token',
    })
  })

  it('refreshes an expired access token once before retrying the stream', async () => {
    localStorage.setItem('auth_token', 'expired-token')
    localStorage.setItem('refresh_token', 'refresh-token')
    refreshTokenMock.mockResolvedValue({
      access_token: 'fresh-token',
      refresh_token: 'next-refresh-token',
      expires_in: 3600,
    })
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response('{"message":"expired"}', { status: 401 }))
      .mockResolvedValueOnce(streamResponse([
        'event: workbench.generation.completed\ndata: {"message_id":"message-1"}\n\n',
      ]))
    vi.stubGlobal('fetch', fetchMock)

    await streamWorkbenchGeneration('message-1', () => undefined)

    expect(refreshTokenMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(fetchMock.mock.calls[0][1]?.headers).toMatchObject({ Authorization: 'Bearer expired-token' })
    expect(fetchMock.mock.calls[1][1]?.headers).toMatchObject({ Authorization: 'Bearer fresh-token' })
  })
})

describe('workbench model API', () => {
  beforeEach(() => vi.clearAllMocks())

  it('adds multiple models for one key in a single request', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({ data: { items: [] } })

    await addWorkbenchModels(['gpt-5.3-codex-spark', 'gpt-5.2'], 42)

    expect(apiClient.post).toHaveBeenCalledWith('/workbench/models', {
      model_ids: ['gpt-5.3-codex-spark', 'gpt-5.2'],
      api_key_id: 42
    })
  })
})

describe('workbench attachment API', () => {
  beforeEach(() => vi.clearAllMocks())

  it('uploads local-development attachments through the authenticated client', async () => {
    const ticket: WorkbenchAttachmentUploadTicket = {
      attachment: { id: 'attachment-1', name: 'notes.txt', mime_type: 'text/plain', size_bytes: 5 },
      upload: { url: '/workbench/attachments/attachment-1/content', method: 'PUT', requires_auth: true },
      expires_at: '2026-07-28T12:00:00Z'
    }
    const file = new File(['hello'], 'notes.txt', { type: 'text/plain' })

    await uploadWorkbenchAttachment(ticket, file)

    expect(apiClient.request).toHaveBeenCalledWith(expect.objectContaining({
      url: ticket.upload.url,
      method: 'PUT',
      data: file,
      headers: { 'Content-Type': 'text/plain' }
    }))
  })

  it('requests an upload ticket without sending file content to the API', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({ data: { attachment: { id: 'attachment-1' } } })

    await createWorkbenchAttachmentUpload({ name: 'brief.pdf', mime_type: 'application/pdf', size_bytes: 1024 })

    expect(apiClient.post).toHaveBeenCalledWith('/workbench/attachments', {
      name: 'brief.pdf',
      mime_type: 'application/pdf',
      size_bytes: 1024
    })
  })
})
