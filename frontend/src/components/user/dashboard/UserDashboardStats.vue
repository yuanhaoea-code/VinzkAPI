<template>
  <div class="dashboard-stats">
    <div class="dashboard-stats__grid">
      <UserConsoleStatCard
        v-for="item in statCards"
        :key="item.title"
        :title="item.title"
        :value="item.value"
        :hint="item.hint"
        :detail="item.detail"
        :tone="item.tone"
      >
        <template #icon>
          <Icon :name="item.icon" size="md" />
        </template>
      </UserConsoleStatCard>
    </div>

    <UserConsolePanel
      v-if="!isSimple && platformCards.length > 0"
      :title="t('dashboard.platformBreakdown')"
      :description="t('dashboard.console.platformDescription')"
    >
      <template #headerMeta>
        <span class="dashboard-platforms__meta">
          {{ t('dashboard.platformCount', { count: sortedPlatforms.length }) }}
        </span>
      </template>

      <div class="dashboard-platforms">
        <article
          v-for="item in platformCards"
          :key="item.platform"
          class="dashboard-platforms__card"
          :class="{ 'dashboard-platforms__card--other': item.isOther }"
        >
          <div class="dashboard-platforms__top">
            <div>
              <p class="dashboard-platforms__name">
                {{ item.isOther ? t('dashboard.platformOther') : platformLabel(item.platform) }}
              </p>
              <p class="dashboard-platforms__subtle">
                {{ t('dashboard.todayCost') }} ${{ formatCost(item.today_actual_cost) }}
              </p>
            </div>
            <p class="dashboard-platforms__cost">${{ formatCost(item.total_actual_cost) }}</p>
          </div>

          <div class="dashboard-platforms__rows">
            <div class="dashboard-platforms__row">
              <span>{{ t('dashboard.requests') }}</span>
              <strong>{{ item.total_requests > 0 ? formatNumber(item.total_requests) : '-' }}</strong>
            </div>
            <div class="dashboard-platforms__row">
              <span>{{ t('dashboard.tokens') }}</span>
              <strong>{{ item.total_tokens > 0 ? formatTokens(item.total_tokens) : '-' }}</strong>
            </div>
          </div>

          <div
            v-if="hasAnyLimit(item.quota) && !item.isOther"
            class="dashboard-platforms__quota"
          >
            <p class="dashboard-platforms__quota-title">
              {{ t('dashboard.platformQuota.title') }}
            </p>
            <template v-for="windowName in quotaWindows" :key="windowName">
              <div
                v-if="quotaVal(item.quota, `${windowName}_limit_usd`) != null"
                class="dashboard-platforms__quota-item"
              >
                <template v-if="(quotaVal(item.quota, `${windowName}_limit_usd`) as number) === 0">
                  <div class="dashboard-platforms__row">
                    <span>{{ t(`dashboard.platformQuota.${windowName}`) }}</span>
                    <strong class="dashboard-platforms__danger">
                      {{ t('dashboard.platformQuota.disabled') }}
                    </strong>
                  </div>
                  <div class="dashboard-platforms__bar">
                    <div class="dashboard-platforms__bar-fill dashboard-platforms__bar-fill--danger" />
                  </div>
                </template>
                <template v-else>
                  <div class="dashboard-platforms__row">
                    <span>{{ t(`dashboard.platformQuota.${windowName}`) }}</span>
                    <strong>
                      ${{ formatUsd((quotaVal(item.quota, `${windowName}_usage_usd`) as number) ?? 0) }}
                      /
                      ${{ formatUsd(quotaVal(item.quota, `${windowName}_limit_usd`) as number) }}
                    </strong>
                  </div>
                  <div class="dashboard-platforms__bar">
                    <div
                      class="dashboard-platforms__bar-fill"
                      :class="quotaBarClass(calcPercent((quotaVal(item.quota, `${windowName}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${windowName}_limit_usd`) as number))"
                      :style="{ width: `${calcPercent((quotaVal(item.quota, `${windowName}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${windowName}_limit_usd`) as number)}%` }"
                    />
                  </div>
                  <p
                    v-if="quotaVal(item.quota, `${windowName}_window_resets_at`)"
                    class="dashboard-platforms__quota-reset"
                  >
                    {{
                      t('dashboard.platformQuota.resetsAt', {
                        time: formatResetTime(quotaVal(item.quota, `${windowName}_window_resets_at`) as string),
                      })
                    }}
                  </p>
                </template>
              </div>
            </template>
          </div>
        </article>
      </div>
    </UserConsolePanel>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import UserConsolePanel from '@/components/user/console/UserConsolePanel.vue'
import UserConsoleStatCard from '@/components/user/console/UserConsoleStatCard.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { PlatformQuotaItem } from '@/types'

interface FusedPlatformCard {
  platform: string
  total_actual_cost: number
  today_actual_cost: number
  total_requests: number
  total_tokens: number
  isOther?: boolean
  quota?: PlatformQuotaItem
}

type StatTone = 'neutral' | 'sky' | 'emerald' | 'amber' | 'violet'

interface DashboardStatCard {
  title: string
  value: string | number
  hint: string
  detail: string
  tone: StatTone
  icon:
    | 'creditCard'
    | 'key'
    | 'chart'
    | 'dollar'
    | 'cube'
    | 'database'
    | 'bolt'
    | 'clock'
}

type QuotaWindow = 'daily' | 'weekly' | 'monthly'
type QuotaField =
  | `${QuotaWindow}_limit_usd`
  | `${QuotaWindow}_usage_usd`
  | `${QuotaWindow}_window_resets_at`

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
  platformQuotas?: PlatformQuotaItem[] | null
}>()

const { t } = useI18n()
const quotaWindows: QuotaWindow[] = ['daily', 'weekly', 'monthly']

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity',
}

const formatBalance = (value: number) =>
  Number(value || 0).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })

const formatCost = (value: number) => Number(value || 0).toFixed(4)

const formatNumber = (value: number) => Number(value || 0).toLocaleString()

const formatDuration = (ms: number) => (ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`)

const formatTokens = (value: number) => {
  if (value >= 1e9) return `${(value / 1e9).toFixed(2)}B`
  if (value >= 1e6) return `${(value / 1e6).toFixed(2)}M`
  if (value >= 1e3) return `${(value / 1e3).toFixed(2)}K`
  return Number(value || 0).toLocaleString()
}

const platformLabel = (platform: string) => PLATFORM_LABELS[platform] ?? platform

const statCards = computed(() => {
  const cards: Array<DashboardStatCard | null> = [
    !props.isSimple
      ? {
          title: t('dashboard.balance'),
          value: `$${formatBalance(props.balance)}`,
          hint: t('dashboard.console.balanceHint'),
          detail: t('common.available'),
          tone: 'emerald',
          icon: 'creditCard',
        }
      : null,
    {
      title: t('dashboard.apiKeys'),
      value: props.stats?.total_api_keys || 0,
      hint: `${props.stats?.active_api_keys || 0} ${t('common.active')}`,
      detail: t('dashboard.console.keysDetail'),
      tone: 'sky',
      icon: 'key',
    },
    {
      title: t('dashboard.todayRequests'),
      value: formatNumber(props.stats?.today_requests || 0),
      hint: `${t('common.total')}: ${formatNumber(props.stats?.total_requests || 0)}`,
      detail: t('dashboard.console.requestsDetail'),
      tone: 'emerald',
      icon: 'chart',
    },
    {
      title: t('dashboard.todayCost'),
      value: `$${formatCost(props.stats?.today_actual_cost || 0)}`,
      hint: `${t('dashboard.actual')} / ${t('dashboard.standard')}`,
      detail: `$${formatCost(props.stats?.total_actual_cost || 0)} / $${formatCost(props.stats?.total_cost || 0)}`,
      tone: 'violet',
      icon: 'dollar',
    },
    {
      title: t('dashboard.todayTokens'),
      value: formatTokens(props.stats?.today_tokens || 0),
      hint: `${t('dashboard.input')} ${formatTokens(props.stats?.today_input_tokens || 0)}`,
      detail: `${t('dashboard.output')} ${formatTokens(props.stats?.today_output_tokens || 0)}`,
      tone: 'amber',
      icon: 'cube',
    },
    {
      title: t('dashboard.totalTokens'),
      value: formatTokens(props.stats?.total_tokens || 0),
      hint: `${t('dashboard.input')} ${formatTokens(props.stats?.total_input_tokens || 0)}`,
      detail: `${t('dashboard.output')} ${formatTokens(props.stats?.total_output_tokens || 0)}`,
      tone: 'sky',
      icon: 'database',
    },
    {
      title: t('dashboard.performance'),
      value: `${formatTokens(props.stats?.rpm || 0)} RPM`,
      hint: `${formatTokens(props.stats?.tpm || 0)} TPM`,
      detail: t('dashboard.console.performanceDetail'),
      tone: 'violet',
      icon: 'bolt',
    },
    {
      title: t('dashboard.avgResponse'),
      value: formatDuration(props.stats?.average_duration_ms || 0),
      hint: t('dashboard.averageTime'),
      detail: t('dashboard.console.responseDetail'),
      tone: 'amber',
      icon: 'clock',
    },
  ]

  return cards.filter((item): item is DashboardStatCard => item !== null)
})

const sortedPlatforms = computed(() => {
  const list = props.stats?.by_platform ?? []
  return [...list].sort((a, b) => b.total_actual_cost - a.total_actual_cost)
})

const platformCards = computed<FusedPlatformCard[]>(() => {
  const byPlatform = new Map<string, (typeof sortedPlatforms.value)[number]>()
  for (const item of props.stats?.by_platform ?? []) byPlatform.set(item.platform, item)

  const byQuota = new Map<string, PlatformQuotaItem>()
  for (const item of props.platformQuotas ?? []) byQuota.set(item.platform, item)

  const platforms = new Set<string>([...byPlatform.keys(), ...byQuota.keys()])
  const platformOrder = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
  const cards: FusedPlatformCard[] = []

  for (const platform of platforms) {
    const stat = byPlatform.get(platform)
    cards.push({
      platform,
      total_actual_cost: stat?.total_actual_cost ?? 0,
      today_actual_cost: stat?.today_actual_cost ?? 0,
      total_requests: stat?.total_requests ?? 0,
      total_tokens: stat?.total_tokens ?? 0,
      quota: byQuota.get(platform),
    })
  }

  cards.sort((left, right) => {
    const leftIndex = platformOrder.indexOf(left.platform)
    const rightIndex = platformOrder.indexOf(right.platform)
    if (leftIndex === -1 && rightIndex === -1) return left.platform.localeCompare(right.platform)
    if (leftIndex === -1) return 1
    if (rightIndex === -1) return -1
    return leftIndex - rightIndex
  })

  const otherThreshold = 0.0001
  const totalCost = props.stats?.total_actual_cost ?? 0
  const todayCost = props.stats?.today_actual_cost ?? 0
  const sumTotal = cards.reduce((sum, item) => sum + item.total_actual_cost, 0)
  const sumToday = cards.reduce((sum, item) => sum + item.today_actual_cost, 0)

  const diffTotal = Math.max(0, totalCost - sumTotal)
  const diffToday = Math.max(0, todayCost - sumToday)

  if (diffTotal > otherThreshold || diffToday > otherThreshold) {
    cards.push({
      platform: '__other__',
      total_actual_cost: diffTotal,
      today_actual_cost: diffToday,
      total_requests: 0,
      total_tokens: 0,
      isOther: true,
    })
  }

  return cards
})

function quotaVal(quota: PlatformQuotaItem | undefined, key: QuotaField): PlatformQuotaItem[QuotaField] {
  return quota?.[key]
}

function hasAnyLimit(quota: PlatformQuotaItem | undefined): boolean {
  if (!quota) return false
  return quota.daily_limit_usd != null || quota.weekly_limit_usd != null || quota.monthly_limit_usd != null
}

function calcPercent(usage: number, limit: number): number {
  if (!limit || limit <= 0) return 0
  return Math.min(100, Math.max(0, Math.round((usage / limit) * 100)))
}

function quotaBarClass(percent: number): string {
  if (percent >= 95) return 'dashboard-platforms__bar-fill--danger'
  if (percent >= 75) return 'dashboard-platforms__bar-fill--warn'
  return 'dashboard-platforms__bar-fill--good'
}

const usdFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

function formatUsd(value: number): string {
  if (!Number.isFinite(value)) return '0.00'
  return usdFormatter.format(value)
}

function formatResetTime(iso: string | null | undefined): string {
  if (!iso) return ''
  const date = new Date(iso)
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
</script>

<style scoped>
.dashboard-stats {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dashboard-stats__grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 12px;
}

.dashboard-platforms__meta {
  font-size: 0.8rem;
  color: #6b7280;
}

.dashboard-platforms {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  gap: 12px;
}

.dashboard-platforms__card {
  border: 1px solid rgba(17, 24, 39, 0.08);
  border-radius: 16px;
  padding: 12px 13px;
  background: rgba(255, 255, 255, 0.58);
}

.dashboard-platforms__card--other {
  border-style: dashed;
}

.dashboard-platforms__top {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
}

.dashboard-platforms__name,
.dashboard-platforms__cost {
  margin: 0;
}

.dashboard-platforms__name {
  font-size: 0.92rem;
  font-weight: 600;
  color: #111827;
}

.dashboard-platforms__cost {
  font-size: 0.95rem;
  font-weight: 600;
  color: #6f5f43;
}

.dashboard-platforms__subtle,
.dashboard-platforms__quota-reset {
  margin: 0.25rem 0 0;
  font-size: 0.77rem;
  color: #6b7280;
}

.dashboard-platforms__rows,
.dashboard-platforms__quota {
  margin-top: 0.75rem;
}

.dashboard-platforms__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  font-size: 0.82rem;
  color: #4b5563;
}

.dashboard-platforms__row + .dashboard-platforms__row {
  margin-top: 0.45rem;
}

.dashboard-platforms__row strong {
  color: #111827;
  font-weight: 600;
}

.dashboard-platforms__danger {
  color: #dc2626 !important;
}

.dashboard-platforms__quota {
  border-top: 1px solid rgba(17, 24, 39, 0.08);
  padding-top: 0.75rem;
}

.dashboard-platforms__quota-title {
  margin: 0 0 0.6rem;
  font-size: 0.72rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #9ca3af;
}

.dashboard-platforms__quota-item + .dashboard-platforms__quota-item {
  margin-top: 0.65rem;
}

.dashboard-platforms__bar {
  margin-top: 0.4rem;
  height: 0.38rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.22);
}

.dashboard-platforms__bar-fill {
  height: 100%;
  border-radius: inherit;
}

.dashboard-platforms__bar-fill--good {
  background: #7f9f98;
}

.dashboard-platforms__bar-fill--warn {
  background: #d4b165;
}

.dashboard-platforms__bar-fill--danger {
  background: #ba948e;
}

:global(.dark) .dashboard-platforms__meta,
:global(.dark) .dashboard-platforms__subtle,
:global(.dark) .dashboard-platforms__quota-reset {
  color: #94a3b8;
}

:global(.dark) .dashboard-platforms__card {
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.76), rgba(15, 23, 42, 0.92));
}

:global(.dark) .dashboard-platforms__name,
:global(.dark) .dashboard-platforms__row strong {
  color: #f8fafc;
}

:global(.dark) .dashboard-platforms__row {
  color: #cbd5e1;
}

:global(.dark) .dashboard-platforms__quota {
  border-top-color: rgba(255, 255, 255, 0.08);
}

@media (min-width: 640px) {
  .dashboard-stats__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1180px) {
  .dashboard-stats__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
