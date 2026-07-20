<template>
  <AppLayout>
    <UserConsolePage
      :kicker="t('dashboard.console.kicker')"
      :title="t('dashboard.console.title')"
      compact
    >
      <template #headerActions>
        <button
          type="button"
          class="btn btn-secondary"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
          :disabled="loading || loadingCharts || loadingUsage"
          @click="refreshAll"
        >
          <Icon
            name="refresh"
            size="md"
            :class="loading || loadingCharts || loadingUsage ? 'animate-spin' : ''"
          />
        </button>
      </template>

      <div v-if="loading" class="dashboard-loading">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <UserDashboardStats
          :stats="stats"
          :balance="user?.balance || 0"
          :is-simple="authStore.isSimpleMode"
          :platform-quotas="platformQuotas"
        />
        <UserDashboardCharts
          v-model:startDate="startDate"
          v-model:endDate="endDate"
          v-model:granularity="granularity"
          :loading="loadingCharts"
          :trend="trendData"
          :models="modelStats"
          @dateRangeChange="loadCharts"
          @granularityChange="loadCharts"
          @refresh="refreshAll"
        />
        <div class="dashboard-lower-grid">
          <UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" />
          <UserDashboardQuickActions />
        </div>
      </template>
    </UserConsolePage>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserConsolePage from '@/components/user/console/UserConsolePage.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { ModelStat, PlatformQuotaItem, TrendDataPoint, UsageLog } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'

const { t } = useI18n()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)

const formatLocalDate = (date: Date) => date.toISOString().split('T')[0]
const startDate = ref(formatLocalDate(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatLocalDate(new Date()))
const granularity = ref('day')
const loadStats = async () => {
  loading.value = true
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  } finally {
    loading.value = false
  }
}

const loadCharts = async () => {
  loadingCharts.value = true
  try {
    const [trendResponse, modelResponse] = await Promise.all([
      usageAPI.getDashboardTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value as 'day' | 'hour',
      }),
      usageAPI.getDashboardModels({
        start_date: startDate.value,
        end_date: endDate.value,
      }),
    ])

    trendData.value = trendResponse.trend || []
    modelStats.value = modelResponse.models || []
  } catch (error) {
    console.error('Failed to load charts:', error)
  } finally {
    loadingCharts.value = false
  }
}

const loadRecent = async () => {
  loadingUsage.value = true
  try {
    const response = await usageAPI.getByDateRange(startDate.value, endDate.value)
    recentUsage.value = response.items.slice(0, 5)
  } catch (error) {
    console.error('Failed to load recent usage:', error)
  } finally {
    loadingUsage.value = false
  }
}

const loadPlatformQuotas = async () => {
  try {
    const data = await getMyPlatformQuotas()
    platformQuotas.value = data.platform_quotas ?? []
  } catch (error) {
    console.warn('Failed to load platform quotas:', error)
    platformQuotas.value = []
  }
}

const refreshAll = () => {
  void loadStats()
  void loadCharts()
  void loadRecent()
  void loadPlatformQuotas()
}

onMounted(() => {
  refreshAll()
})
</script>

<style scoped>
.dashboard-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 18rem;
}

.dashboard-lower-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 14px;
}

@media (min-width: 1180px) {
  .dashboard-lower-grid {
    grid-template-columns: minmax(0, 1.22fr) minmax(19rem, 0.78fr);
  }
}
</style>
