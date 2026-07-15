<template>
  <AppLayout>
    <div class="redeem-page">
      <section class="redeem-hero">
        <div>
          <p class="console-kicker">REDEEM CENTER</p>
          <h1 class="redeem-title">{{ t('redeem.title') }}</h1>
          <p class="redeem-lead">
            输入兑换码，可为账户增加余额、并发额度或订阅权限，兑换结果和账户状态会自动更新。
          </p>
        </div>
        <aside class="redeem-note">
          <strong>快速兑换</strong>
          <span>兑换前可查看当前余额和并发额度，成功后可在最近活动中确认到账记录。</span>
        </aside>
      </section>

      <div class="redeem-layout">
        <section class="redeem-card redeem-main-card">
          <div class="section-heading">
            <p class="console-kicker">ENTER CODE</p>
            <h2>输入兑换码</h2>
            <p>兑换码可用于充值余额、增加并发额度或开通订阅，具体内容以兑换结果为准。</p>
          </div>

          <div class="balance-grid">
            <div class="balance-box">
              <span>{{ t('redeem.currentBalance') }}</span>
              <strong>USD {{ user?.balance?.toFixed(2) || '0.00' }}</strong>
            </div>
            <div class="balance-box">
              <span>{{ t('redeem.concurrency') }}</span>
              <strong>{{ user?.concurrency || 0 }}</strong>
              <small>{{ t('redeem.requests') }}</small>
            </div>
          </div>

          <form class="redeem-form" @submit.prevent="handleRedeem">
            <label class="code-label" for="code">{{ t('redeem.redeemCodeLabel') }}</label>
            <div class="code-input-wrap">
              <Icon name="gift" size="md" class="code-input-icon" />
              <input
                id="code"
                v-model="redeemCode"
                type="text"
                required
                :placeholder="t('redeem.redeemCodePlaceholder')"
                :disabled="submitting"
                class="code-input"
              />
            </div>
            <p class="input-hint">{{ t('redeem.redeemCodeHint') }}</p>

            <div class="submit-row">
              <button type="submit" :disabled="!redeemCode || submitting" class="redeem-button">
                <svg v-if="submitting" class="spin-icon" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <Icon v-else name="checkCircle" size="md" />
                <span>{{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}</span>
              </button>
              <p>兑换成功后，账户余额、并发额度或订阅状态将自动更新。</p>
            </div>
          </form>

          <transition name="fade">
            <div v-if="redeemResult" class="result-card success">
              <Icon name="checkCircle" size="lg" />
              <div>
                <h3>{{ t('redeem.redeemSuccess') }}</h3>
                <p>{{ redeemResult.message }}</p>
                <div class="result-lines">
                  <p v-if="redeemResult.type === 'balance'">
                    {{ t('redeem.added') }}: USD {{ redeemResult.value.toFixed(2) }}
                  </p>
                  <p v-else-if="redeemResult.type === 'concurrency'">
                    {{ t('redeem.added') }}: {{ redeemResult.value }} {{ t('redeem.concurrentRequests') }}
                  </p>
                  <p v-else-if="redeemResult.type === 'subscription'">
                    {{ t('redeem.subscriptionAssigned') }}
                    <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                    <span v-if="redeemResult.validity_days">
                      ({{ t('redeem.subscriptionDays', { days: redeemResult.validity_days }) }})
                    </span>
                  </p>
                  <p v-if="redeemResult.new_balance !== undefined">
                    {{ t('redeem.newBalance') }}: USD {{ redeemResult.new_balance.toFixed(2) }}
                  </p>
                  <p v-if="redeemResult.new_concurrency !== undefined">
                    {{ t('redeem.newConcurrency') }}: {{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}
                  </p>
                </div>
              </div>
            </div>
          </transition>

          <transition name="fade">
            <div v-if="errorMessage" class="result-card error">
              <Icon name="exclamationCircle" size="lg" />
              <div>
                <h3>{{ t('redeem.redeemFailed') }}</h3>
                <p>{{ errorMessage }}</p>
              </div>
            </div>
          </transition>
        </section>

        <aside class="redeem-side">
          <section class="redeem-card info-card">
            <div class="section-heading compact">
              <p class="console-kicker">RULES</p>
              <h2>{{ t('redeem.aboutCodes') }}</h2>
              <p>使用兑换码前，请确认兑换码的有效期与适用范围。</p>
            </div>
            <ul class="rule-list">
              <li>{{ t('redeem.codeRule1') }}</li>
              <li>{{ t('redeem.codeRule2') }}</li>
              <li>
                {{ t('redeem.codeRule3') }}
                <span v-if="contactInfo" class="contact-pill">{{ contactInfo }}</span>
              </li>
              <li>{{ t('redeem.codeRule4') }}</li>
            </ul>
          </section>

          <section class="redeem-card history-card">
            <div class="section-heading compact">
              <p class="console-kicker">RECENT</p>
              <h2>{{ t('redeem.recentActivity') }}</h2>
              <p>查看近期兑换结果以及余额、并发或订阅的到账记录。</p>
            </div>

            <div v-if="loadingHistory" class="history-loading">
              <svg class="spin-icon" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            </div>

            <div v-else-if="history.length > 0" class="history-list">
              <div v-for="item in history" :key="item.id" class="history-item">
                <div class="history-icon" :class="historyToneClass(item)">
                  <Icon :name="historyIconName(item)" size="md" />
                </div>
                <div class="history-content">
                  <p>{{ getHistoryItemTitle(item) }}</p>
                  <span>{{ formatDateTime(item.used_at) }}</span>
                  <small v-if="!isAdminAdjustment(item.type)">{{ item.code.slice(0, 8) }}...</small>
                  <small v-else>{{ t('redeem.adminAdjustment') }}</small>
                  <small v-if="item.notes" :title="item.notes">{{ item.notes }}</small>
                </div>
                <strong :class="historyValueClass(item)">{{ formatHistoryValue(item) }}</strong>
              </div>
            </div>

            <div v-else class="empty-history">
              <Icon name="clock" size="xl" />
              <p>{{ t('redeem.historyWillAppear') }}</p>
            </div>
          </section>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { redeemAPI, authAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import { PUBLIC_CONTACT_INFO } from '@/constants/site'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const user = computed(() => authStore.user)

const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
} | null>(null)
const errorMessage = ref('')

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const contactInfo = ref(PUBLIC_CONTACT_INFO)

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance' || type === 'admin_balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

const isAdminAdjustment = (type: string) => {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const historyIconName = (item: RedeemHistoryItem): 'dollar' | 'badge' | 'bolt' => {
  if (isBalanceType(item.type)) return 'dollar'
  if (isSubscriptionType(item.type)) return 'badge'
  return 'bolt'
}

const historyToneClass = (item: RedeemHistoryItem) => {
  if (isAdminAdjustment(item.type)) return 'admin'
  if (isBalanceType(item.type)) return 'balance'
  if (isSubscriptionType(item.type)) return 'subscription'
  return 'concurrency'
}

const historyValueClass = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) return item.value >= 0 ? 'value-green' : 'value-red'
  if (isSubscriptionType(item.type)) return 'value-violet'
  return item.value >= 0 ? 'value-blue' : 'value-red'
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    redeemResult.value = result

    // Refresh user data to get updated balance/concurrency
    await authStore.refreshUser()

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || t('redeem.failedToRedeem')

    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  fetchHistory()
  try {
    const settings = await authAPI.getPublicSettings()
    contactInfo.value = settings.contact_info || PUBLIC_CONTACT_INFO
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
.redeem-page {
  --panel: rgba(255, 255, 255, 0.82);
  --ink: #171513;
  --muted: #73695f;
  --line: rgba(23, 21, 19, 0.1);
  --line-strong: rgba(23, 21, 19, 0.16);
  --teal-soft: rgba(106, 147, 139, 0.14);
  --red-soft: rgba(187, 107, 105, 0.1);
  --green-soft: rgba(95, 142, 114, 0.11);
  max-width: 1560px;
  margin: 0 auto;
  color: var(--ink);
}
.redeem-hero { display: grid; grid-template-columns: minmax(0, 1fr) 330px; gap: 24px; align-items: end; margin-bottom: 18px; }
.console-kicker { margin: 0 0 8px; color: #8c7351; font-size: 12px; font-weight: 700; letter-spacing: 0.2em; }
.redeem-title { margin: 0; font-family: var(--console-font-sans); font-size: 30px; font-weight: 700; line-height: 1.1; }
.redeem-lead, .redeem-note span, .section-heading p, .submit-row p { color: var(--muted); font-size: 13px; line-height: 1.82; }
.redeem-lead { max-width: 840px; margin: 10px 0 0; }
.redeem-note, .redeem-card { border: 1px solid var(--line); background: var(--panel); box-shadow: 0 1px 0 rgba(44, 34, 24, 0.03); }
.redeem-note { border-radius: 18px; padding: 14px 16px; }
.redeem-note strong { display: block; margin-bottom: 8px; font-size: 13px; }
.redeem-layout { display: grid; grid-template-columns: 1.08fr 0.92fr; gap: 14px; }
.redeem-card { border-radius: 24px; padding: 18px; }
.redeem-side { display: grid; gap: 14px; align-content: start; }
.section-heading h2 { margin: 0; font-family: var(--console-font-sans); font-size: 24px; line-height: 1.2; }
.section-heading p { margin: 8px 0 0; }
.compact h2 { font-size: 22px; }
.balance-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; margin-top: 18px; }
.balance-box { min-height: 86px; border: 1px solid var(--line); border-radius: 18px; background: rgba(255, 255, 255, 0.62); padding: 14px 16px; }
.balance-box span, .balance-box small { display: block; color: var(--muted); font-size: 13px; }
.balance-box strong { display: inline-block; margin-top: 8px; font-family: var(--console-font-sans); font-size: 30px; line-height: 1; }
.redeem-form { margin-top: 18px; }
.code-label { display: block; margin-bottom: 8px; color: #50483f; font-size: 13px; font-weight: 700; }
.code-input-wrap { display: flex; align-items: center; gap: 10px; height: 58px; border: 1px solid var(--line-strong); border-radius: 18px; background: rgba(255, 255, 255, 0.84); padding: 0 16px; }
.code-input-icon { color: #8a7d70; flex: 0 0 auto; }
.code-input { min-width: 0; width: 100%; border: 0; outline: 0; background: transparent; color: var(--ink); font-size: 15px; }
.code-input::placeholder { color: #9d9183; }
.input-hint { margin: 10px 0 0; color: #8d8276; font-size: 13px; }
.submit-row { display: flex; align-items: center; gap: 12px; margin-top: 16px; }
.submit-row p { margin: 0; }
.redeem-button { display: inline-flex; min-width: 170px; height: 42px; align-items: center; justify-content: center; gap: 8px; border: 0; border-radius: 999px; background: var(--ink); color: #f7f3eb; font-size: 13px; font-weight: 800; }
.redeem-button:disabled { cursor: not-allowed; opacity: 0.45; }
.spin-icon { width: 20px; height: 20px; animation: spin 800ms linear infinite; }
.result-card { display: flex; align-items: flex-start; gap: 12px; margin-top: 16px; border: 1px solid var(--line); border-radius: 18px; padding: 14px 15px; font-size: 13px; line-height: 1.75; }
.result-card h3, .result-card p { margin: 0; }
.result-card h3 { font-size: 14px; }
.result-card.success { border-color: rgba(95, 142, 114, 0.2); background: var(--green-soft); color: #325241; }
.result-card.error { border-color: rgba(187, 107, 105, 0.22); background: var(--red-soft); color: #7c4746; }
.result-lines { margin-top: 8px; font-weight: 700; }
.rule-list { margin: 14px 0 0; padding-left: 18px; color: #5f584f; font-size: 13px; line-height: 1.9; }
.contact-pill { display: inline-flex; margin-left: 8px; border-radius: 999px; background: var(--teal-soft); padding: 3px 8px; color: #31584f; font-size: 13px; font-weight: 700; }
.history-card { min-height: 360px; }
.history-loading, .empty-history { display: grid; place-items: center; gap: 10px; min-height: 180px; color: var(--muted); text-align: center; }
.history-list { display: grid; gap: 10px; margin-top: 16px; }
.history-item { display: grid; grid-template-columns: 42px minmax(0, 1fr) auto; gap: 12px; align-items: center; border: 1px solid rgba(23, 21, 19, 0.06); border-radius: 16px; background: rgba(247, 243, 235, 0.88); padding: 12px; }
.history-icon { display: grid; place-items: center; width: 42px; height: 42px; border-radius: 14px; color: white; }
.history-icon.balance { background: linear-gradient(135deg, #6f9c80, #4c745c); }
.history-icon.concurrency { background: linear-gradient(135deg, #7ba7ca, #4f7191); }
.history-icon.subscription { background: linear-gradient(135deg, #9c86b6, #66557b); }
.history-icon.admin { background: linear-gradient(135deg, #c58b6d, #8a5f47); }
.history-content { min-width: 0; }
.history-content p { margin: 0; font-size: 13px; font-weight: 800; }
.history-content span, .history-content small { display: block; overflow: hidden; color: var(--muted); text-overflow: ellipsis; white-space: nowrap; font-size: 12px; line-height: 1.55; }
.history-item strong { text-align: right; font-size: 13px; line-height: 1.5; }
.value-green { color: #39654a; }
.value-blue { color: #456784; }
.value-violet { color: #5f4d76; }
.value-red { color: #7a4747; }
.fade-enter-active, .fade-leave-active { transition: all 0.22s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(-8px); }
@keyframes spin { to { transform: rotate(360deg); } }
:global(.dark) .redeem-page { --panel: rgba(28, 26, 24, 0.88); --ink: #f5efe6; --muted: #b7aa9c; --line: rgba(255, 255, 255, 0.1); --line-strong: rgba(255, 255, 255, 0.16); --teal-soft: rgba(106, 147, 139, 0.22); }
:global(.dark) .redeem-note, :global(.dark) .redeem-card { background: var(--panel); }
:global(.dark) .balance-box, :global(.dark) .code-input-wrap, :global(.dark) .history-item { background: rgba(255, 255, 255, 0.06); }
:global(.dark) .rule-list, :global(.dark) .code-label { color: var(--ink); }
@media (max-width: 1180px) { .redeem-hero, .redeem-layout { grid-template-columns: 1fr; } }
@media (max-width: 720px) { .balance-grid, .history-item { grid-template-columns: 1fr; } .submit-row { align-items: stretch; flex-direction: column; } .redeem-button { width: 100%; } }
</style>
