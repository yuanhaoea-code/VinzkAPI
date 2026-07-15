import { nextTick, onUnmounted, shallowRef, watch, type ComputedRef, type Ref } from 'vue'
import type { TutorialArticle } from '@/components/user/tutorial/types'

export function useTutorialOutline(
  article: ComputedRef<TutorialArticle | undefined>,
  articleRoot: Ref<HTMLElement | null>,
) {
  const activeSectionId = shallowRef('')
  let observer: IntersectionObserver | null = null

  function disconnectObserver() {
    observer?.disconnect()
    observer = null
  }

  async function observeSections() {
    disconnectObserver()
    activeSectionId.value = article.value?.sections[0]?.id ?? ''
    await nextTick()

    if (!articleRoot.value || typeof IntersectionObserver === 'undefined') return

    observer = new IntersectionObserver((entries) => {
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
      const id = visible[0]?.target.id
      if (id) activeSectionId.value = id
    }, {
      rootMargin: '-90px 0px -68% 0px',
      threshold: [0, 0.1, 0.5],
    })

    articleRoot.value
      .querySelectorAll<HTMLElement>('[data-tutorial-section]')
      .forEach((section) => observer?.observe(section))
  }

  function scrollToSection(id: string) {
    const section = articleRoot.value?.querySelector<HTMLElement>(`#${CSS.escape(id)}`)
    section?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeSectionId.value = id
  }

  watch(() => article.value?.slug, observeSections, { immediate: true })
  onUnmounted(disconnectObserver)

  return {
    activeSectionId,
    scrollToSection,
  }
}
