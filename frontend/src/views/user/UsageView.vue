<template>
  <AppLayout>
    <UserConsolePage
      :kicker="t('usage.console.kicker')"
      :title="t('usage.console.title')"
      :description="t('usage.console.description')"
      class="usage-console"
    >
      <template #heroAside>
        <div class="console-summary">
          <div>
            <div class="console-summary__label">SELECTED RANGE</div>
            <div class="console-summary__value">{{ usageStats?.total_requests?.toLocaleString() || '0' }}</div>
            <div class="console-summary__desc">
              {{ t('usage.console.summaryDescription') }}
            </div>
          </div>
          <div class="console-summary__micro">
            <div>
              <b>{{ formatCompactTokens(usageStats?.total_tokens || 0) }}</b>
              <span>{{ t('usage.tokens') }}</span>
            </div>
            <div>
              <b>${{ (usageStats?.total_actual_cost || 0).toFixed(2) }}</b>
              <span>{{ t('usage.console.actualCost') }}</span>
            </div>
          </div>
        </div>
      </template>

      <template #heroNotes>
        <div class="console-note">
          <strong class="console-note__title">{{ t('usage.console.noteBillingTitle') }}</strong>
          <span class="console-note__copy">{{ t('usage.console.noteBillingCopy') }}</span>
        </div>
        <div class="console-note">
          <strong class="console-note__title">{{ t('usage.console.noteExportTitle') }}</strong>
          <span class="console-note__copy">{{ t('usage.console.noteExportCopy') }}</span>
        </div>
        <div class="console-note">
          <strong class="console-note__title">{{ t('usage.console.noteErrorTitle') }}</strong>
          <span class="console-note__copy">{{ t('usage.console.noteErrorCopy') }}</span>
        </div>
      </template>

      <div class="usage-console__stats">
        <UserConsoleStatCard
          v-for="item in usageStatCards"
          :key="item.title"
          :title="item.title"
          :value="item.value"
          :hint="item.hint"
          :detail="item.detail"
          :meta="item.meta"
          :tone="item.tone"
        />
      </div>

      <UserConsolePanel
        :title="t('usage.console.chartsTitle')"
        :description="t('usage.console.chartsDescription')"
      >
        <template #headerActions>
          <div class="usage-console__chart-controls">
            <div class="usage-console__chart-control">
              <span class="usage-console__control-label">{{ t('admin.dashboard.timeRange') }}</span>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
            </div>
            <div class="usage-console__chart-control usage-console__chart-control--small">
              <span class="usage-console__control-label">{{ t('admin.dashboard.granularity') }}</span>
              <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
            </div>
          </div>
        </template>

        <div class="usage-console__chart-grid">
          <ModelDistributionChart
            v-model:metric="modelDistributionMetric"
            :model-stats="requestedModelStats"
            :loading="modelStatsLoading"
            :show-source-toggle="false"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :show-account-cost="false"
            :start-date="startDate"
            :end-date="endDate"
          />
          <GroupDistributionChart
            v-model:metric="groupDistributionMetric"
            :group-stats="groupStats"
            :loading="chartsLoading"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :show-account-cost="false"
            :start-date="startDate"
            :end-date="endDate"
          />
        </div>
        <div class="usage-console__chart-grid">
          <EndpointDistributionChart
            v-model:source="endpointDistributionSource"
            v-model:metric="endpointDistributionMetric"
            :endpoint-stats="inboundEndpointStats"
            :upstream-endpoint-stats="upstreamEndpointStats"
            :endpoint-path-stats="endpointPathStats"
            :loading="endpointStatsLoading"
            :show-source-toggle="false"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :title="t('usage.endpointDistribution')"
            :start-date="startDate"
            :end-date="endDate"
          />
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </UserConsolePanel>

      <UserConsolePanel
        :title="t('usage.console.filtersTitle')"
        :description="t('usage.console.filtersDescription')"
      >
        <div class="flex flex-wrap items-end justify-between gap-4">
          <div class="flex flex-1 flex-wrap items-end gap-4">
            <div class="w-full sm:w-auto sm:min-w-[220px]">
              <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
              <Select v-model="filters.api_key_id" :options="apiKeyOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[220px]">
              <label class="input-label">{{ t('usage.model') }}</label>
              <Select v-model="filters.model" :options="modelOptions" searchable @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.group') }}</label>
              <Select v-model="filters.group_id" :options="groupOptions" searchable @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[180px]">
              <label class="input-label">{{ t('usage.type') }}</label>
              <Select v-model="filters.request_type" :options="requestTypeOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.billingType') }}</label>
              <Select v-model="filters.billing_type" :options="billingTypeOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.billingMode') }}</label>
              <Select v-model="filters.billing_mode" :options="billingModeOptions" @change="applyFilters" />
            </div>
          </div>

          <div class="flex w-full flex-wrap items-center justify-end gap-3 sm:w-auto">
            <button type="button" @click="refreshData" :disabled="loading" class="btn btn-secondary">
              {{ t('common.refresh') }}
            </button>
            <button type="button" @click="resetFilters" class="btn btn-secondary">
              {{ t('common.reset') }}
            </button>
            <div class="relative" ref="columnDropdownRef">
              <button
                type="button"
                @click="showColumnDropdown = !showColumnDropdown"
                class="btn btn-secondary px-2 md:px-3"
                :title="t('admin.users.columnSettings')"
                :aria-label="t('admin.users.columnSettings')"
              >
                <Icon name="grid" size="sm" />
                <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
              </button>
              <div
                v-if="showColumnDropdown"
                class="absolute right-0 top-full z-50 mt-1 max-h-80 w-48 overflow-y-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
              >
                <button
                  v-for="col in toggleableColumns"
                  :key="col.key"
                  type="button"
                  @click="toggleColumn(col.key)"
                  class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
                >
                  <span>{{ col.label }}</span>
                  <Icon v-if="isColumnVisible(col.key)" name="check" size="sm" class="text-primary-500" />
                </button>
              </div>
            </div>
            <button type="button" @click="exportToCSV" :disabled="exporting" class="btn btn-primary">
              {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
            </button>
            </div>
          </div>
      </UserConsolePanel>

      <UserConsolePanel
        :title="activeTab === 'usage' ? t('usage.console.recordsTitle') : t('usage.tabs.errors')"
        :description="activeTab === 'usage' ? t('usage.console.recordsDescription') : t('usage.console.errorsDescription')"
      >
        <template #headerActions>
          <div v-if="errorViewEnabled" class="usage-console__tabs">
            <button class="usage-console__tab" :class="{ 'usage-console__tab--active': activeTab === 'usage' }" @click="activeTab = 'usage'">
              {{ t('usage.tabs.usage') }}
            </button>
            <button class="usage-console__tab" :class="{ 'usage-console__tab--active': activeTab === 'errors' }" @click="switchToErrors">
              {{ t('usage.tabs.errors') }}
            </button>
          </div>
        </template>

        <template v-if="activeTab === 'usage'">
          <UsageTable
            :data="usageLogs"
            :loading="loading"
            :columns="visibleColumns"
            :server-side-sort="true"
            :show-account-billing="false"
            :show-upstream-endpoint="false"
            default-sort-key="created_at"
            default-sort-order="desc"
            @sort="handleSort"
            @ipGeoBatchFailed="handleIpGeoBatchFailed"
          />

          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </template>

        <UserErrorRequestsTable
          v-else-if="errorViewEnabled"
          :rows="errorRows"
          :total="errorTotal"
          :loading="errorLoading"
          :page="errorPage"
          :page-size="errorPageSize"
          :api-keys="apiKeys"
          @filter="onErrorFilter"
          @update:page="onErrorPage"
          @update:pageSize="onErrorPageSize"
        />
      </UserConsolePanel>
    </UserConsolePage>
  </AppLayout>

</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import UserConsolePage from '@/components/user/console/UserConsolePage.vue'
import UserConsolePanel from '@/components/user/console/UserConsolePanel.vue'
import UserConsoleStatCard from '@/components/user/console/UserConsoleStatCard.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import GroupDistributionChart from '@/components/charts/GroupDistributionChart.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import Icon from '@/components/icons/Icon.vue'
import UserErrorRequestsTable from '@/components/user/UserErrorRequestsTable.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatReasoningEffort } from '@/utils/format'
import { BILLING_MODE_IMAGE, getBillingModeLabel } from '@/utils/billingMode'
import { resolveUsageRequestType, requestTypeToLegacyStream } from '@/utils/usageRequestType'
import type {
  ApiKey,
  EndpointStat,
  Group,
  GroupStat,
  ModelStat,
  TrendDataPoint,
  UsageLog,
  UsageQueryParams,
  UsageStatsResponse,
  UserErrorRequest,
} from '@/types'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()

type DistributionMetric = 'tokens' | 'actual_cost'
type EndpointSource = 'inbound' | 'upstream' | 'path'
type UsageStatTone = 'neutral' | 'sky' | 'emerald' | 'amber' | 'violet'

const usageStats = ref<UsageStatsResponse | null>(null)
const usageLogs = ref<UsageLog[]>([])
const trendData = ref<TrendDataPoint[]>([])
const requestedModelStats = ref<ModelStat[]>([])
const groupStats = ref<GroupStat[]>([])
const inboundEndpointStats = ref<EndpointStat[]>([])
const upstreamEndpointStats = ref<EndpointStat[]>([])
const endpointPathStats = ref<EndpointStat[]>([])

const loading = ref(false)
const chartsLoading = ref(false)
const modelStatsLoading = ref(false)
const endpointStatsLoading = ref(false)
const exporting = ref(false)
const errorRows = ref<UserErrorRequest[]>([])
const errorLoading = ref(false)
const errorPage = ref(1)
const errorPageSize = ref(20)
const errorTotal = ref(0)
const errorFilter = ref<{ model: string; category: string; api_key_id: number | null }>({
  model: '',
  category: '',
  api_key_id: null,
})

let abortController: AbortController | null = null
let chartReqSeq = 0
let statsReqSeq = 0
let modelStatsReqSeq = 0

const formatLocalDate = (date: Date): string =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

const getLast24HoursRangeDates = () => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return { start: formatLocalDate(start), end: formatLocalDate(end) }
}

const getGranularityForRange = (start: string, end: string): 'day' | 'hour' => {
  const startTime = new Date(`${start}T00:00:00`).getTime()
  const endTime = new Date(`${end}T00:00:00`).getTime()
  return Math.ceil((endTime - startTime) / (1000 * 60 * 60 * 24)) <= 1 ? 'hour' : 'day'
}

const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)
const granularity = ref<'day' | 'hour'>(getGranularityForRange(startDate.value, endDate.value))
const formatCompactTokens = (value: number) => {
  if (value >= 1e9) return `${(value / 1e9).toFixed(1)}B`
  if (value >= 1e6) return `${Math.round(value / 1e6)}M`
  if (value >= 1e3) return `${Math.round(value / 1e3)}K`
  return Number(value || 0).toLocaleString()
}
const formatUsageDuration = (ms: number) => (ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`)

const usageStatCards = computed<Array<{
  title: string
  value: string
  hint: string
  detail: string
  meta: string
  tone: UsageStatTone
}>>(() => [
  {
    title: t('usage.totalRequests'),
    value: (usageStats.value?.total_requests || 0).toLocaleString(),
    hint: t('usage.inSelectedRange'),
    detail: t('usage.console.recordsDescription'),
    meta: 'Range',
    tone: 'sky',
  },
  {
    title: t('usage.totalTokens'),
    value: formatCompactTokens(usageStats.value?.total_tokens || 0),
    hint: `${t('usage.in')}: ${formatCompactTokens(usageStats.value?.total_input_tokens || 0)}`,
    detail: `${t('usage.out')}: ${formatCompactTokens(usageStats.value?.total_output_tokens || 0)}`,
    meta: 'Token',
    tone: 'amber',
  },
  {
    title: t('usage.totalCost'),
    value: `$${(usageStats.value?.total_actual_cost || 0).toFixed(4)}`,
    hint: `${t('usage.standardCost')} $${(usageStats.value?.total_cost || 0).toFixed(4)}`,
    detail: t('usage.console.actualCost'),
    meta: 'Actual',
    tone: 'emerald',
  },
  {
    title: t('usage.avgDuration'),
    value: formatUsageDuration(usageStats.value?.average_duration_ms || 0),
    hint: t('dashboard.averageTime'),
    detail: t('usage.console.filtersDescription'),
    meta: 'Avg',
    tone: 'violet',
  },
])

const modelDistributionMetric = ref<DistributionMetric>('tokens')
const groupDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionSource = ref<EndpointSource>('inbound')
const activeTab = ref<'usage' | 'errors'>('usage')
const errorViewEnabled = computed(() => appStore.cachedPublicSettings?.allow_user_view_error_requests ?? false)

const filters = ref<UsageQueryParams>({
  start_date: startDate.value,
  end_date: endDate.value,
  request_type: undefined,
  billing_type: null,
  billing_mode: null,
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc',
})

const granularityOptions = computed<SelectOption[]>(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') },
])
const requestTypeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allTypes') },
  { value: 'ws_v2', label: t('usage.ws') },
  { value: 'stream', label: t('usage.stream') },
  { value: 'sync', label: t('usage.sync') },
])
const billingTypeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allBillingTypes') },
  { value: 0, label: t('admin.usage.billingTypeBalance') },
  { value: 1, label: t('admin.usage.billingTypeSubscription') },
])
const billingModeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allBillingModes') },
  { value: 'token', label: t('admin.usage.billingModeToken') },
  { value: 'per_request', label: t('admin.usage.billingModePerRequest') },
  { value: 'image', label: t('admin.usage.billingModeImage') },
])

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const modelOptionValues = ref<string[]>([])

const apiKeyOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.allApiKeys') },
  ...apiKeys.value.map((key) => ({ value: key.id, label: key.name })),
])
const groupOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allGroups') },
  ...groups.value.map((group) => ({ value: group.id, label: group.name })),
])
const modelOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allModels') },
  ...modelOptionValues.value.map((model) => ({ value: model, label: model })),
])

const normalizedFilters = computed<UsageQueryParams>(() => {
  const requestType = filters.value.request_type
  const legacyStream = requestType ? requestTypeToLegacyStream(requestType) : filters.value.stream
  return {
    ...filters.value,
    start_date: startDate.value,
    end_date: endDate.value,
    stream: legacyStream === null ? undefined : legacyStream,
  }
})

const buildUsageListParams = (page: number, pageSize: number): UsageQueryParams => ({
  page,
  page_size: pageSize,
  ...normalizedFilters.value,
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order,
})

const loadLogs = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  loading.value = true
  try {
    const res = await usageAPI.query(buildUsageListParams(pagination.page, pagination.page_size), {
      signal: controller.signal,
    })
    if (!controller.signal.aborted) {
      usageLogs.value = res.items
      pagination.total = res.total
    }
  } catch (error: any) {
    if (error?.name !== 'AbortError' && error?.code !== 'ERR_CANCELED') {
      appStore.showError(t('usage.failedToLoad'))
    }
  } finally {
    if (abortController === controller) loading.value = false
  }
}

const loadStats = async () => {
  const seq = ++statsReqSeq
  endpointStatsLoading.value = true
  try {
    const stats = await usageAPI.getStats(normalizedFilters.value)
    if (seq !== statsReqSeq) return
    usageStats.value = stats
    inboundEndpointStats.value = stats.endpoints || []
    upstreamEndpointStats.value = []
    endpointPathStats.value = []
  } catch (error) {
    if (seq !== statsReqSeq) return
    console.error('Failed to load usage stats:', error)
    inboundEndpointStats.value = []
    upstreamEndpointStats.value = []
    endpointPathStats.value = []
  } finally {
    if (seq === statsReqSeq) endpointStatsLoading.value = false
  }
}

const loadModelStats = async () => {
  const seq = ++modelStatsReqSeq
  modelStatsLoading.value = true
  try {
    const response = await usageAPI.getDashboardModels({
      ...normalizedFilters.value,
      model_source: 'requested',
    })
    if (seq !== modelStatsReqSeq) return
    requestedModelStats.value = response.models || []
    refreshModelOptions(response.models || [])
  } catch (error) {
    if (seq !== modelStatsReqSeq) return
    console.error('Failed to load model stats:', error)
    requestedModelStats.value = []
  } finally {
    if (seq === modelStatsReqSeq) modelStatsLoading.value = false
  }
}

const loadChartData = async () => {
  const seq = ++chartReqSeq
  chartsLoading.value = true
  try {
    const snapshot = await usageAPI.getDashboardSnapshotV2({
      ...normalizedFilters.value,
      granularity: granularity.value,
      include_trend: true,
      include_model_stats: false,
      include_group_stats: true,
    })
    if (seq !== chartReqSeq) return
    trendData.value = snapshot.trend || []
    groupStats.value = snapshot.groups || []
  } catch (error) {
    if (seq !== chartReqSeq) return
    console.error('Failed to load chart data:', error)
    trendData.value = []
    groupStats.value = []
  } finally {
    if (seq === chartReqSeq) chartsLoading.value = false
  }
}

const refreshModelOptions = (models: ModelStat[]) => {
  const current = filters.value.model
  const set = new Set(modelOptionValues.value)
  models.forEach((item) => {
    if (item.model) set.add(item.model)
  })
  if (current) set.add(current)
  modelOptionValues.value = Array.from(set).sort()
}

const applyFilters = () => {
  pagination.page = 1
  void loadLogs()
  void loadStats()
  void loadModelStats()
  void loadChartData()
  resetErrorRows()
}

const refreshData = () => {
  void loadLogs()
  void loadStats()
  void loadModelStats()
  void loadChartData()
  if (activeTab.value === 'errors') void loadErrors()
}

const resetFilters = () => {
  const range = getLast24HoursRangeDates()
  startDate.value = range.start
  endDate.value = range.end
  filters.value = {
    start_date: range.start,
    end_date: range.end,
    request_type: undefined,
    billing_type: null,
    billing_mode: null,
  }
  granularity.value = getGranularityForRange(range.start, range.end)
  applyFilters()
}

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  startDate.value = range.startDate
  endDate.value = range.endDate
  filters.value.start_date = range.startDate
  filters.value.end_date = range.endDate
  granularity.value = getGranularityForRange(range.startDate, range.endDate)
  applyFilters()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  void loadLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadLogs()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  void loadLogs()
}

const handleIpGeoBatchFailed = () => {
  appStore.showError(t('usage.ipGeo.batchFailed'))
}

const getRequestTypeExportText = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'Cyber'
  if (requestType === 'ws_v2') return 'WS'
  if (requestType === 'stream') return 'Stream'
  if (requestType === 'sync') return 'Sync'
  return 'Unknown'
}

const getDisplayBillingMode = (
  row: Pick<UsageLog, 'billing_mode' | 'image_count'> | null | undefined
): string | null | undefined => {
  if ((row?.image_count ?? 0) > 0) return BILLING_MODE_IMAGE
  return row?.billing_mode
}

const escapeCSVValue = (value: unknown): string => {
  if (value == null) return ''
  const str = String(value)
  const escaped = str.replace(/"/g, '""')
  if (/^[=+\-@\t\r]/.test(str)) return `"\'${escaped}"`
  if (/[,"\n\r]/.test(str)) return `"${escaped}"`
  return str
}

const exportToCSV = async () => {
  if (pagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }
  exporting.value = true
  appStore.showInfo(t('usage.preparingExport'))
  try {
    const allLogs: UsageLog[] = []
    const pageSize = 100
    const totalPages = Math.ceil(pagination.total / pageSize)
    for (let page = 1; page <= totalPages; page++) {
      const response = await usageAPI.query(buildUsageListParams(page, pageSize))
      allLogs.push(...response.items)
    }
    if (allLogs.length === 0) {
      appStore.showWarning(t('usage.noDataToExport'))
      return
    }
    const headers = [
      'Time',
      'API Key Name',
      'Model',
      'Reasoning Effort',
      'Inbound Endpoint',
      'IP Address',
      'Type',
      'Billing Mode',
      'Input Tokens',
      'Output Tokens',
      'Cache Read Tokens',
      'Cache Creation Tokens',
      'Rate Multiplier',
      'Billed Cost',
      'Original Cost',
      'First Token (ms)',
      'Duration (ms)',
    ]
    const rows = allLogs.map((log) => [
      log.created_at,
      log.api_key?.name || '',
      log.model,
      formatReasoningEffort(log.reasoning_effort),
      log.inbound_endpoint || '',
      log.ip_address || '',
      getRequestTypeExportText(log),
      getBillingModeLabel(getDisplayBillingMode(log), t),
      log.input_tokens,
      log.output_tokens,
      log.cache_read_tokens,
      log.cache_creation_tokens,
      log.rate_multiplier,
      log.actual_cost.toFixed(8),
      log.total_cost.toFixed(8),
      log.first_token_ms ?? '',
      log.duration_ms ?? '',
    ].map(escapeCSVValue))
    const csvContent = [
      headers.map(escapeCSVValue).join(','),
      ...rows.map((row) => row.join(',')),
    ].join('\n')
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `usage_${startDate.value}_to_${endDate.value}.csv`
    link.click()
    window.URL.revokeObjectURL(url)
    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    console.error('CSV Export failed:', error)
    appStore.showError(t('usage.exportFailed'))
  } finally {
    exporting.value = false
  }
}

const ALWAYS_VISIBLE = ['created_at']
const DEFAULT_HIDDEN_COLUMNS = ['user_agent']
const HIDDEN_COLUMNS_KEY = 'user-usage-hidden-columns'

const allColumns = computed<Column[]>(() => [
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'ip_address', label: 'IP', sortable: false },
  { key: 'group', label: t('admin.usage.group'), sortable: false },
  { key: 'stream', label: t('usage.type'), sortable: false },
  { key: 'billing_mode', label: t('admin.usage.billingMode'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'first_token', label: t('usage.firstToken'), sortable: false },
  { key: 'duration', label: t('usage.duration'), sortable: false },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
])

const hiddenColumns = reactive<Set<string>>(new Set())
const toggleableColumns = computed(() => allColumns.value.filter((col) => !ALWAYS_VISIBLE.includes(col.key)))
const visibleColumns = computed(() =>
  allColumns.value.filter((col) => ALWAYS_VISIBLE.includes(col.key) || !hiddenColumns.has(col.key))
)
const isColumnVisible = (key: string) => !hiddenColumns.has(key)
const toggleColumn = (key: string) => {
  if (hiddenColumns.has(key)) hiddenColumns.delete(key)
  else hiddenColumns.add(key)
  localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
}
const loadSavedColumns = () => {
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    const values = saved ? JSON.parse(saved) as string[] : DEFAULT_HIDDEN_COLUMNS
    values.forEach((key) => hiddenColumns.add(key))
  } catch {
    DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
  }
}

const showColumnDropdown = ref(false)
const columnDropdownRef = ref<HTMLElement | null>(null)
const handleColumnClickOutside = (event: MouseEvent) => {
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(event.target as HTMLElement)) {
    showColumnDropdown.value = false
  }
}

const loadFilterOptions = async () => {
  try {
    const [keys, availableGroups] = await Promise.all([
      keysAPI.list(1, 100),
      userGroupsAPI.getAvailable(),
    ])
    apiKeys.value = keys.items
    groups.value = availableGroups
  } catch (error) {
    console.error('Failed to load usage filter options:', error)
  }
}

const resetErrorRows = () => {
  errorPage.value = 1
  if (activeTab.value === 'errors') {
    void loadErrors()
  } else {
    errorRows.value = []
    errorTotal.value = 0
  }
}

const loadErrors = async () => {
  errorLoading.value = true
  try {
    const resp = await usageAPI.listMyErrorRequests({
      page: errorPage.value,
      page_size: errorPageSize.value,
      start_date: startDate.value,
      end_date: endDate.value,
      model: errorFilter.value.model || undefined,
      category: errorFilter.value.category || undefined,
      api_key_id: errorFilter.value.api_key_id ?? undefined,
    })
    errorRows.value = resp.items
    errorTotal.value = resp.total
  } catch (error) {
    console.error('[UsageView] loadErrors failed:', error)
    appStore.showError(t('usage.errors.failedToLoad'))
  } finally {
    errorLoading.value = false
  }
}

const onErrorFilter = (filter: { model: string; category: string; api_key_id: number | null }) => {
  errorFilter.value = filter
  errorPage.value = 1
  void loadErrors()
}

const onErrorPage = (page: number) => {
  errorPage.value = page
  void loadErrors()
}

const onErrorPageSize = (pageSize: number) => {
  errorPageSize.value = pageSize
  errorPage.value = 1
  void loadErrors()
}

const switchToErrors = () => {
  activeTab.value = 'errors'
  if (errorRows.value.length === 0) void loadErrors()
}

onMounted(() => {
  loadSavedColumns()
  document.addEventListener('click', handleColumnClickOutside)
  void loadFilterOptions()
  refreshData()
})

onUnmounted(() => {
  abortController?.abort()
  document.removeEventListener('click', handleColumnClickOutside)
})

watch(endpointDistributionSource, () => {
  // Endpoint source switching is handled by the chart component using already loaded stats.
})
</script>

<style scoped>
.usage-console__stats {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 12px;
}

.usage-console :deep(.card) {
  border-radius: 16px;
  border-color: rgba(23, 20, 17, 0.08);
  background: rgba(255, 255, 255, 0.68);
  box-shadow: 0 1px 0 rgba(36, 29, 22, 0.02);
}

.usage-console__chart-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 12px;
}

.usage-console__chart-grid + .usage-console__chart-grid {
  margin-top: 12px;
}

.usage-console__chart-grid :deep(.card) {
  min-height: 252px;
  padding: 12px 13px;
}

.usage-console__chart-grid :deep(canvas) {
  filter: saturate(0.72) sepia(0.08) hue-rotate(8deg);
}

.usage-console__chart-grid :deep(h3) {
  font-family: var(--console-font-sans);
  font-size: 17px;
  line-height: 1.2;
}

.usage-console__chart-grid :deep(.h-48) {
  height: 164px;
}

.usage-console__chart-grid :deep(.w-48) {
  width: 164px;
}

.usage-console__chart-grid :deep(.gap-6) {
  gap: 14px;
}

.usage-console__chart-controls {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.usage-console__chart-control {
  min-width: 0;
}

.usage-console__chart-control--small {
  width: 8rem;
}

.usage-console__control-label {
  display: block;
  margin-bottom: 0.35rem;
  color: #7c7267;
  font-size: 13px;
  letter-spacing: 0;
  text-transform: none;
}

.usage-console__tabs {
  display: inline-flex;
  gap: 0.35rem;
  padding: 0.2rem;
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 999px;
  background: rgba(248, 247, 244, 0.82);
}

.usage-console__tab {
  border: 0;
  border-radius: 999px;
  padding: 0.42rem 0.82rem;
  font-size: 0.84rem;
  color: #6b7280;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.usage-console__tab--active {
  background: #171411;
  color: #fbf8f3;
}

.usage-console :deep(.table-wrapper thead),
.usage-console :deep(.table-header) {
  background: rgba(248, 247, 244, 0.92);
}

.usage-console :deep(.table-wrapper) {
  border: 1px solid rgba(23, 20, 17, 0.075);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.58);
  min-height: 328px;
}

.usage-console :deep(th) {
  color: #7c7267;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0;
  text-transform: none;
  padding-top: 10px !important;
  padding-bottom: 10px !important;
}

.usage-console :deep(td) {
  color: #171411;
  font-size: 14px;
  padding-top: 10px !important;
  padding-bottom: 10px !important;
}

.usage-console :deep(.pagination),
.usage-console :deep(nav[aria-label='Pagination']) {
  margin-top: 12px;
}

.usage-console :deep(.input-label) {
  margin-bottom: 6px;
}

:global(.dark) .usage-console__tabs,
:global(.dark) .usage-console :deep(.card) {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.78);
}

:global(.dark) .usage-console__control-label,
:global(.dark) .usage-console__tab {
  color: #94a3b8;
}

:global(.dark) .usage-console__tab--active {
  background: #f8fafc;
  color: #0f172a;
}

:global(.dark) .usage-console :deep(.table-wrapper thead),
:global(.dark) .usage-console :deep(.table-header) {
  background: rgba(15, 23, 42, 0.96);
}

:global(.dark) .usage-console :deep(.table-wrapper) {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.78);
}

@media (min-width: 1180px) {
  .usage-console__stats {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .usage-console__chart-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 700px) and (max-width: 1179px) {
  .usage-console__stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .usage-console__chart-controls {
    justify-content: stretch;
  }

  .usage-console__chart-control,
  .usage-console__chart-control--small {
    width: 100%;
  }

  .usage-console__tabs {
    width: 100%;
    justify-content: stretch;
  }

  .usage-console__tab {
    flex: 1;
  }
}
</style>
