import { describe, expect, it } from 'vitest'
import { PUBLIC_API_BASE_URL, resolvePublicUrl } from './site'

describe('resolvePublicUrl', () => {
  it.each([
    '',
    'http://127.0.0.1:3001',
    'https://localhost:3001',
    'https://192.168.1.8',
    'https://api.example.com',
    'https://your-domain.com',
    'not-a-url',
  ])('uses the production fallback for non-production value %s', (value) => {
    expect(resolvePublicUrl(value, PUBLIC_API_BASE_URL)).toBe(PUBLIC_API_BASE_URL)
  })

  it('keeps a valid HTTPS custom domain and removes trailing slashes', () => {
    expect(resolvePublicUrl(' https://gateway.vinzk.cn/// ', PUBLIC_API_BASE_URL)).toBe(
      'https://gateway.vinzk.cn',
    )
  })
})
