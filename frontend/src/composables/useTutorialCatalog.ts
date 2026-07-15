import { computed } from 'vue'
import { useAppStore } from '@/stores/app'
import {
  PUBLIC_API_BASE_URL,
  PUBLIC_SITE_NAME,
  PUBLIC_SITE_URL,
  resolvePublicUrl,
} from '@/constants/site'
import {
  buildOpenAiBaseUrl,
  createTutorialCatalog,
  normalizeTutorialBaseUrl,
} from '@/components/user/tutorial/catalog'

export function useTutorialCatalog() {
  const appStore = useAppStore()

  const siteUrl = computed(() => PUBLIC_SITE_URL)

  const apiBaseUrl = computed(() =>
    normalizeTutorialBaseUrl(
      resolvePublicUrl(appStore.apiBaseUrl, PUBLIC_API_BASE_URL),
      siteUrl.value,
    )
  )

  const catalog = computed(() => createTutorialCatalog({
    siteName: appStore.siteName || PUBLIC_SITE_NAME,
    siteUrl: siteUrl.value,
    apiBaseUrl: apiBaseUrl.value,
    openAiBaseUrl: buildOpenAiBaseUrl(apiBaseUrl.value),
    apiKeyPlaceholder: 'sk-your-api-key',
  }))

  return {
    apiBaseUrl,
    catalog,
    siteUrl,
  }
}
