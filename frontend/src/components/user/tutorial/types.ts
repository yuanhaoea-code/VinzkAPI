export type TutorialTone = 'info' | 'success' | 'warning'

export type TutorialTag = 'start' | 'codex' | 'claude' | 'gemini' | 'ccswitch' | 'openclaw' | 'ops'

export type TutorialIconName =
  | 'book'
  | 'terminal'
  | 'key'
  | 'swap'
  | 'gift'
  | 'server'
  | 'exclamationTriangle'
  | 'upload'
  | 'sparkles'
  | 'chart'

export interface TutorialLink {
  label: string
  to: string
  external?: boolean
}

export interface TutorialStep {
  title: string
  description: string
}

export type TutorialBlock =
  | { type: 'paragraph'; text: string }
  | { type: 'list'; items: string[]; ordered?: boolean }
  | { type: 'steps'; items: TutorialStep[] }
  | { type: 'code'; label: string; language: string; code: string }
  | { type: 'callout'; tone: TutorialTone; title: string; text: string }
  | { type: 'image'; src: string; alt: string; caption: string }
  | { type: 'links'; links: TutorialLink[] }

export interface TutorialSection {
  id: string
  title: string
  description?: string
  blocks: TutorialBlock[]
}

export interface TutorialArticle {
  slug: string
  group: string
  tag: TutorialTag
  icon: TutorialIconName
  title: string
  shortTitle: string
  eyebrow: string
  description: string
  duration: string
  updatedAt: string
  sections: TutorialSection[]
  relatedSlugs: string[]
}

export interface TutorialGroup {
  id: string
  label: string
  articleSlugs: string[]
}

export interface TutorialCatalog {
  articles: TutorialArticle[]
  groups: TutorialGroup[]
}

export interface TutorialContext {
  siteName: string
  siteUrl: string
  apiBaseUrl: string
  openAiBaseUrl: string
  apiKeyPlaceholder: string
}
