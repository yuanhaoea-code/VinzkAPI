<template>
  <UserConsolePanel
    :title="t('dashboard.recentUsage')"
    :description="t('dashboard.console.recentDescription')"
  >
    <template #headerMeta>
      <span class="dashboard-recent__meta">{{ t('dashboard.last7Days') }}</span>
    </template>

    <div v-if="loading" class="dashboard-recent__loading">
      <LoadingSpinner size="lg" />
    </div>

    <div v-else-if="data.length === 0" class="dashboard-recent__empty">
      <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
    </div>

    <div v-else class="dashboard-recent__list">
      <article v-for="log in data" :key="log.id" class="dashboard-recent__item">
        <div class="dashboard-recent__item-main">
          <div class="dashboard-recent__icon">
            <Icon name="beaker" size="md" />
          </div>
          <div class="dashboard-recent__copy">
            <p class="dashboard-recent__model">{{ log.model }}</p>
            <p class="dashboard-recent__time">{{ formatDateTime(log.created_at) }}</p>
          </div>
        </div>

        <div class="dashboard-recent__item-side">
          <p class="dashboard-recent__cost">
            ${{ formatCost(log.actual_cost) }}
            <span>/ ${{ formatCost(log.total_cost) }}</span>
          </p>
          <p class="dashboard-recent__tokens">
            {{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens
          </p>
        </div>
      </article>

      <router-link to="/usage" class="dashboard-recent__link">
        <span>{{ t('dashboard.viewAllUsage') }}</span>
        <Icon name="arrowRight" size="sm" />
      </router-link>
    </div>
  </UserConsolePanel>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import UserConsolePanel from '@/components/user/console/UserConsolePanel.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()

const { t } = useI18n()
const formatCost = (cost: number) => cost.toFixed(4)
</script>

<style scoped>
.dashboard-recent__meta {
  font-size: 0.8rem;
  color: #6b7280;
}

.dashboard-recent__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 10rem;
}

.dashboard-recent__empty {
  padding: 1rem 0;
}

.dashboard-recent__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.dashboard-recent__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 11px 13px;
  border: 1px solid rgba(17, 24, 39, 0.08);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.58);
}

.dashboard-recent__item-main {
  display: flex;
  align-items: center;
  gap: 11px;
  min-width: 0;
}

.dashboard-recent__icon {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background: rgba(127, 159, 152, 0.1);
  color: #587d77;
  flex-shrink: 0;
}

.dashboard-recent__copy {
  min-width: 0;
}

.dashboard-recent__model,
.dashboard-recent__cost,
.dashboard-recent__tokens,
.dashboard-recent__time {
  margin: 0;
}

.dashboard-recent__model {
  font-size: 0.88rem;
  font-weight: 600;
  color: #111827;
}

.dashboard-recent__time,
.dashboard-recent__tokens {
  margin-top: 3px;
  font-size: 0.76rem;
  color: #6b7280;
}

.dashboard-recent__item-side {
  text-align: right;
}

.dashboard-recent__cost {
  font-size: 0.88rem;
  font-weight: 600;
  color: #587d77;
}

.dashboard-recent__cost span {
  color: #9ca3af;
  font-weight: 400;
}

.dashboard-recent__link {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  margin-top: 0.25rem;
  align-self: flex-start;
  color: #111827;
  font-size: 0.88rem;
  text-decoration: none;
}

:global(.dark) .dashboard-recent__meta,
:global(.dark) .dashboard-recent__time,
:global(.dark) .dashboard-recent__tokens {
  color: #94a3b8;
}

:global(.dark) .dashboard-recent__item {
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.76), rgba(15, 23, 42, 0.92));
}

:global(.dark) .dashboard-recent__model,
:global(.dark) .dashboard-recent__link {
  color: #f8fafc;
}

:global(.dark) .dashboard-recent__cost {
  color: #5eead4;
}

:global(.dark) .dashboard-recent__cost span {
  color: #94a3b8;
}

@media (max-width: 640px) {
  .dashboard-recent__item {
    flex-direction: column;
    align-items: stretch;
  }

  .dashboard-recent__item-side {
    text-align: left;
  }
}
</style>
