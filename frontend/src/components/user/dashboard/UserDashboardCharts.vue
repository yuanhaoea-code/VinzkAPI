<template>
  <UserConsolePanel
    :title="t('dashboard.console.chartTitle')"
    :description="t('dashboard.console.chartDescription')"
    compact
  >
    <template #headerMeta>
      <span class="dashboard-chart-meta">{{ t('dashboard.console.chartMeta') }}</span>
    </template>

    <div class="dashboard-chart-controls">
      <div class="dashboard-chart-controls__main">
        <div class="dashboard-chart-controls__range">
          <span class="dashboard-chart-controls__label">{{ t('dashboard.timeRange') }}</span>
          <DateRangePicker
            :start-date="startDate"
            :end-date="endDate"
            @update:startDate="$emit('update:startDate', $event)"
            @update:endDate="$emit('update:endDate', $event)"
            @change="$emit('dateRangeChange', $event)"
          />
        </div>
        <div class="dashboard-chart-controls__select">
          <span class="dashboard-chart-controls__label">{{ t('dashboard.granularity') }}</span>
          <Select
            :model-value="granularity"
            :options="granularityOptions"
            @update:model-value="$emit('update:granularity', $event)"
            @change="$emit('granularityChange')"
          />
        </div>
        <div class="dashboard-chart-controls__view">
          <span>{{ t('dashboard.console.chartViewLabel') }}</span>
          <strong>{{ t('dashboard.modelDistribution') }} / Token</strong>
        </div>
      </div>
      <button type="button" class="btn btn-secondary dashboard-chart-controls__refresh" :disabled="loading" @click="$emit('refresh')">
        {{ t('common.refresh') }}
      </button>
    </div>

    <div class="dashboard-chart-grid">
      <section class="dashboard-chart-box">
        <div class="dashboard-chart-box__head">
          <h3>{{ t('dashboard.modelDistribution') }}</h3>
          <span>Model Distribution</span>
        </div>
        <div class="dashboard-models">
          <div class="dashboard-models__chart">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="dashboard-models__empty">{{ t('dashboard.noDataAvailable') }}</div>
          </div>

          <div class="dashboard-models__table">
            <div v-if="loading" class="dashboard-models__loading">
              <LoadingSpinner size="md" />
            </div>
            <template v-else>
              <table v-if="models.length" class="dashboard-models__table-inner">
                <thead>
                  <tr>
                    <th>{{ t('dashboard.model') }}</th>
                    <th>{{ t('dashboard.requests') }}</th>
                    <th>{{ t('dashboard.tokens') }}</th>
                    <th>{{ t('dashboard.actual') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="model in models" :key="model.model">
                    <td :title="model.model">{{ model.model }}</td>
                    <td>{{ formatNumber(model.requests) }}</td>
                    <td>{{ formatTokens(model.total_tokens) }}</td>
                    <td>${{ formatCost(model.actual_cost) }}</td>
                  </tr>
                </tbody>
              </table>
              <div v-else class="dashboard-models__empty">{{ t('dashboard.noDataAvailable') }}</div>
            </template>
          </div>
        </div>
      </section>

      <section class="dashboard-chart-box dashboard-chart-trend">
        <TokenUsageTrend :trend-data="trend" :loading="loading" />
      </section>
    </div>
  </UserConsolePanel>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import UserConsolePanel from '@/components/user/console/UserConsolePanel.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { ModelStat, TrendDataPoint } from '@/types'
import {
  formatCostFixed as formatCost,
  formatNumberLocaleString as formatNumber,
  formatTokensK as formatTokens,
} from '@/utils/format'
import {
  ArcElement,
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  loading: boolean
  startDate: string
  endDate: string
  granularity: string
  trend: TrendDataPoint[]
  models: ModelStat[]
}>()

defineEmits([
  'update:startDate',
  'update:endDate',
  'update:granularity',
  'dateRangeChange',
  'granularityChange',
  'refresh',
])

const { t } = useI18n()

const granularityOptions = computed<SelectOption[]>(() => [
  { value: 'day', label: t('dashboard.day') },
  { value: 'hour', label: t('dashboard.hour') },
])

const chartColors = ['#7f9f98', '#d4b165', '#8eaaae', '#b99a89', '#7ea27e', '#ba948e', '#a7b0a3', '#c8c2b6']

const modelData = computed(() => {
  if (!props.models?.length) return null
  return {
    labels: props.models.map((model) => model.model),
    datasets: [
      {
        data: props.models.map((model) => model.total_tokens),
        backgroundColor: chartColors,
        borderWidth: 0,
      },
    ],
  }
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: { label: string; parsed: number }) =>
          `${context.label}: ${formatTokens(context.parsed)} tokens`,
      },
    },
  },
}
</script>

<style scoped>
.dashboard-chart-meta {
  color: #7c7267;
  font-size: 13px;
}

.dashboard-chart-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 15px;
  background: rgba(246, 244, 239, 0.76);
  box-shadow: inset 0 0 0 1px rgba(23, 20, 17, 0.04);
}

.dashboard-chart-controls__main {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 9px;
}

.dashboard-chart-controls__range,
.dashboard-chart-controls__select {
  min-width: 0;
}

.dashboard-chart-controls__select {
  width: 8rem;
}

.dashboard-chart-controls__label {
  display: block;
  margin-bottom: 0.35rem;
  font-size: 13px;
  letter-spacing: 0;
  text-transform: none;
  color: #7c7267;
}

.dashboard-chart-controls__view {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  color: #7c7267;
  font-size: 13px;
}

.dashboard-chart-controls__view strong {
  color: #171411;
  font-weight: 700;
}

.dashboard-chart-controls__refresh {
  min-height: 34px;
  border-radius: 999px;
}

.dashboard-chart-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 14px;
}

.dashboard-chart-box {
  min-width: 0;
  min-height: 260px;
  border: 1px solid rgba(23, 20, 17, 0.06);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.5);
  box-shadow: inset 0 0 0 1px rgba(23, 20, 17, 0.02);
  padding: 13px;
}

.dashboard-chart-box :deep(canvas) {
  filter: saturate(0.72) sepia(0.08) hue-rotate(8deg);
}

.dashboard-chart-box__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.dashboard-chart-box__head h3 {
  margin: 0;
  font-family: var(--console-font-sans);
  font-size: 17px;
  line-height: 1.2;
  color: #171411;
}

.dashboard-chart-box__head span {
  color: #7c5f43;
  font-size: 13px;
}

.dashboard-models {
  display: grid;
  gap: 14px;
}

.dashboard-models__chart {
  position: relative;
  min-height: 152px;
}

.dashboard-models__table {
  position: relative;
  min-height: 112px;
}

.dashboard-models__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 10rem;
}

.dashboard-models__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 10rem;
  color: #6b7280;
  font-size: 0.9rem;
}

.dashboard-models__table-inner {
  width: 100%;
  border-collapse: collapse;
}

.dashboard-models__table-inner th,
.dashboard-models__table-inner td {
  padding: 0.52rem 0;
  border-bottom: 1px solid rgba(17, 24, 39, 0.08);
  font-size: 0.84rem;
  text-align: left;
}

.dashboard-models__table-inner th {
  color: #6b7280;
  font-weight: 500;
}

.dashboard-models__table-inner td {
  color: #111827;
}

.dashboard-models__table-inner td:first-child {
  max-width: 11rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dashboard-models__table-inner td:last-child {
  color: #0f766e;
}

.dashboard-chart-trend :deep(.card) {
  height: 100%;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}

.dashboard-chart-trend :deep(h3) {
  font-family: var(--console-font-sans);
  font-size: 17px;
  line-height: 1.2;
  color: #171411;
}

.dashboard-chart-trend :deep(.h-48) {
  height: 218px;
}

:global(.dark) .dashboard-chart-controls__label,
:global(.dark) .dashboard-chart-meta,
:global(.dark) .dashboard-models__empty,
:global(.dark) .dashboard-models__table-inner th {
  color: #94a3b8;
}

:global(.dark) .dashboard-chart-controls,
:global(.dark) .dashboard-chart-box {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
}

:global(.dark) .dashboard-chart-controls__view {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.7);
}

:global(.dark) .dashboard-chart-controls__view strong,
:global(.dark) .dashboard-chart-box__head h3,
:global(.dark) .dashboard-chart-trend :deep(h3) {
  color: #f8fafc;
}

:global(.dark) .dashboard-models__table-inner th,
:global(.dark) .dashboard-models__table-inner td {
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

:global(.dark) .dashboard-models__table-inner td {
  color: #f8fafc;
}

:global(.dark) .dashboard-models__table-inner td:last-child {
  color: #5eead4;
}

:global(.dark) .dashboard-chart-trend :deep(.card) {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.78);
  box-shadow: 0 16px 36px rgba(2, 6, 23, 0.18);
}

@media (min-width: 900px) {
  .dashboard-models {
    grid-template-columns: minmax(11rem, 14rem) minmax(0, 1fr);
    align-items: start;
  }
}

@media (min-width: 1180px) {
  .dashboard-chart-grid {
    grid-template-columns: minmax(0, 1fr) minmax(0, 0.98fr);
  }
}

@media (max-width: 640px) {
  .dashboard-chart-controls {
    justify-content: stretch;
  }

  .dashboard-chart-controls__range,
  .dashboard-chart-controls__select,
  .dashboard-chart-controls__view,
  .dashboard-chart-controls .btn {
    width: 100%;
  }
}
</style>
