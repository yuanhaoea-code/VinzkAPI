import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import adminComplianceAPI, { type AdminComplianceStatus } from '@/api/admin/compliance'
import { getLocale } from '@/i18n'
import { useAppStore } from '@/stores/app'
import { PUBLIC_SITE_NAME } from '@/constants/site'

const LEGACY_ZH_PHRASE = '我已阅读、理解并同意 Sub2API 部署与运营合规承诺'
const LEGACY_EN_PHRASE = 'I have read, understood, and agree to the Sub2API Deployment and Operation Compliance Commitment'
const INVALID_PHRASE_CODE = 'ADMIN_COMPLIANCE_INVALID_PHRASE'

function phraseSiteName(siteName?: string): string {
  const trimmed = siteName?.trim()
  return trimmed || PUBLIC_SITE_NAME
}

function buildZhPhrase(siteName?: string): string {
  return `我已阅读、理解并同意 ${phraseSiteName(siteName)} 部署与运营合规承诺`
}

function buildEnPhrase(siteName?: string): string {
  return `I have read, understood, and agree to the ${phraseSiteName(siteName)} Deployment and Operation Compliance Commitment`
}

function isInvalidPhraseError(error: unknown): boolean {
  const apiError = error as { code?: unknown; reason?: unknown; message?: unknown }
  return (
    apiError?.code === INVALID_PHRASE_CODE ||
    apiError?.reason === INVALID_PHRASE_CODE ||
    apiError?.message === 'confirmation phrase does not match'
  )
}

export const useAdminComplianceStore = defineStore('adminCompliance', () => {
  const appStore = useAppStore()
  const status = ref<AdminComplianceStatus | null>(null)
  const loading = ref(false)
  const submitting = ref(false)
  const initialized = ref(false)
  const forceVisible = ref(false)

  const required = computed(() => status.value?.required === true)
  const shouldShow = computed(() => required.value || forceVisible.value)
  const currentLocale = computed(() => getLocale())
  const fallbackZhPhrase = computed(() => buildZhPhrase(appStore.siteName))
  const fallbackEnPhrase = computed(() => buildEnPhrase(appStore.siteName))
  const expectedPhrase = computed(() => {
    if (currentLocale.value === 'zh') {
      return status.value?.ack_phrase_zh || fallbackZhPhrase.value
    }
    return status.value?.ack_phrase_en || fallbackEnPhrase.value
  })

  async function fetchStatus(): Promise<AdminComplianceStatus> {
    loading.value = true
    try {
      const nextStatus = await adminComplianceAPI.getStatus()
      status.value = nextStatus
      initialized.value = true
      forceVisible.value = nextStatus.required
      return nextStatus
    } finally {
      loading.value = false
    }
  }

  async function accept(phrase: string): Promise<AdminComplianceStatus> {
    submitting.value = true
    try {
      let nextStatus: AdminComplianceStatus
      try {
        nextStatus = await adminComplianceAPI.accept({
          phrase,
          language: currentLocale.value
        })
      } catch (error) {
        const legacyPhrase = currentLocale.value === 'zh' ? LEGACY_ZH_PHRASE : LEGACY_EN_PHRASE
        if (!isInvalidPhraseError(error) || phrase === legacyPhrase) {
          throw error
        }
        nextStatus = await adminComplianceAPI.accept({
          phrase: legacyPhrase,
          language: currentLocale.value
        })
      }
      status.value = nextStatus
      forceVisible.value = nextStatus.required
      return nextStatus
    } finally {
      submitting.value = false
    }
  }

  function requireAcknowledgement(partialStatus?: Partial<AdminComplianceStatus>): void {
    status.value = {
      required: true,
      version: partialStatus?.version || status.value?.version || 'v2026.06.10',
      document_path_zh: partialStatus?.document_path_zh || status.value?.document_path_zh || 'docs/legal/admin-compliance.zh.md',
      document_path_en: partialStatus?.document_path_en || status.value?.document_path_en || 'docs/legal/admin-compliance.en.md',
      document_url_zh: partialStatus?.document_url_zh || status.value?.document_url_zh || 'https://github.com/Wei-Shaw/sub2api/blob/main/docs/legal/admin-compliance.zh.md',
      document_url_en: partialStatus?.document_url_en || status.value?.document_url_en || 'https://github.com/Wei-Shaw/sub2api/blob/main/docs/legal/admin-compliance.en.md',
      ack_phrase_zh: partialStatus?.ack_phrase_zh || status.value?.ack_phrase_zh || fallbackZhPhrase.value,
      ack_phrase_en: partialStatus?.ack_phrase_en || status.value?.ack_phrase_en || fallbackEnPhrase.value,
      acknowledgement: status.value?.acknowledgement
    }
    initialized.value = true
    forceVisible.value = true
  }

  function reset(): void {
    status.value = null
    loading.value = false
    submitting.value = false
    initialized.value = false
    forceVisible.value = false
  }

  return {
    status,
    loading,
    submitting,
    initialized,
    required,
    shouldShow,
    expectedPhrase,
    fetchStatus,
    accept,
    requireAcknowledgement,
    reset
  }
})
