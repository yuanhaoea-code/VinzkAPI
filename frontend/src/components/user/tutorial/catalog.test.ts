import { describe, expect, it } from 'vitest'
import {
  buildOpenAiBaseUrl,
  createTutorialCatalog,
  normalizeTutorialBaseUrl,
} from './catalog'

const context = {
  siteName: 'Vinzk AI',
  siteUrl: 'https://console.example.com',
  apiBaseUrl: 'https://api.example.com',
  openAiBaseUrl: 'https://api.example.com/v1',
  apiKeyPlaceholder: 'sk-your-api-key',
}

describe('tutorial catalog', () => {
  it('keeps article and navigation references valid', () => {
    const catalog = createTutorialCatalog(context)
    const slugs = catalog.articles.map((article) => article.slug)
    const knownSlugs = new Set(slugs)

    expect(knownSlugs.size).toBe(slugs.length)
    expect(catalog.groups.flatMap((group) => group.articleSlugs).every((slug) => knownSlugs.has(slug))).toBe(true)
    expect(catalog.articles.flatMap((article) => article.relatedSlugs).every((slug) => knownSlugs.has(slug))).toBe(true)
  })

  it('does not retain reference-site branding or support-group content', () => {
    const serialized = JSON.stringify(createTutorialCatalog(context))

    expect(serialized).not.toMatch(/geiliapi|链动小铺|售后群|QQ\s*群/i)
  })

  it('includes CCSwitch configuration and screenshot guidance', () => {
    const article = createTutorialCatalog(context).articles.find(({ slug }) => slug === 'ccswitch')
    const blocks = article?.sections.flatMap((section) => section.blocks) ?? []

    expect(article?.title).toContain('CCSwitch')
    expect(blocks.some((block) => block.type === 'image' && block.src.includes('ccswitch'))).toBe(true)
    expect(JSON.stringify(article)).toContain('/keys')
  })
})

describe('tutorial API URLs', () => {
  it('normalizes absolute, relative, empty, and trailing-slash base URLs', () => {
    expect(normalizeTutorialBaseUrl('https://api.example.com///', context.siteUrl)).toBe('https://api.example.com')
    expect(normalizeTutorialBaseUrl('/gateway/', context.siteUrl)).toBe('https://console.example.com/gateway')
    expect(normalizeTutorialBaseUrl('', `${context.siteUrl}/`)).toBe(context.siteUrl)
  })

  it('adds the OpenAI compatibility path exactly once', () => {
    expect(buildOpenAiBaseUrl('https://api.example.com')).toBe('https://api.example.com/v1')
    expect(buildOpenAiBaseUrl('https://api.example.com/v1')).toBe('https://api.example.com/v1')
  })
})
