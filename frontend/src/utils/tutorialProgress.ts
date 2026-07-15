const LAST_TUTORIAL_STORAGE_KEY = 'user_tutorial_last_read'

export function getLastTutorialSlug(): string {
  if (typeof window === 'undefined') return ''
  return window.localStorage.getItem(LAST_TUTORIAL_STORAGE_KEY) ?? ''
}

export function saveLastTutorialSlug(slug: string): void {
  if (typeof window === 'undefined' || !slug) return
  window.localStorage.setItem(LAST_TUTORIAL_STORAGE_KEY, slug)
}
