<template>
  <UserConsolePanel
    :title="t('dashboard.quickActions')"
    :description="t('dashboard.console.quickActionsDescription')"
  >
    <div class="dashboard-actions">
      <button
        v-for="action in actions"
        :key="action.title"
        type="button"
        class="dashboard-actions__item"
        @click="router.push(action.to)"
      >
        <div class="dashboard-actions__icon" :class="`dashboard-actions__icon--${action.tone}`">
          <Icon :name="action.icon" size="md" />
        </div>
        <div class="dashboard-actions__copy">
          <p class="dashboard-actions__title">{{ action.title }}</p>
          <p class="dashboard-actions__description">{{ action.description }}</p>
        </div>
        <Icon name="arrowRight" size="sm" class="dashboard-actions__arrow" />
      </button>
    </div>
  </UserConsolePanel>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import UserConsolePanel from '@/components/user/console/UserConsolePanel.vue'

const router = useRouter()
const { t } = useI18n()

const actions = computed(() => [
  {
    title: t('dashboard.createApiKey'),
    description: t('dashboard.generateNewKey'),
    to: '/keys',
    tone: 'sky',
    icon: 'key' as const,
  },
  {
    title: t('dashboard.viewUsage'),
    description: t('dashboard.checkDetailedLogs'),
    to: '/usage',
    tone: 'emerald',
    icon: 'chart' as const,
  },
  {
    title: t('dashboard.redeemCode'),
    description: t('dashboard.addBalanceWithCode'),
    to: '/redeem',
    tone: 'amber',
    icon: 'gift' as const,
  },
  {
    title: t('monitor.title'),
    description: t('dashboard.console.monitorDescription'),
    to: '/monitor',
    tone: 'violet',
    icon: 'server' as const,
  },
])
</script>

<style scoped>
.dashboard-actions {
  display: flex;
  flex-direction: column;
  gap: 9px;
}

.dashboard-actions__item {
  display: flex;
  align-items: center;
  gap: 11px;
  width: 100%;
  padding: 11px 13px;
  border: 1px solid rgba(17, 24, 39, 0.08);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.56);
  text-align: left;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.dashboard-actions__item:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(36, 29, 22, 0.035);
}

.dashboard-actions__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 12px;
  flex-shrink: 0;
}

.dashboard-actions__icon--sky {
  color: #587d77;
  background: rgba(127, 159, 152, 0.12);
}

.dashboard-actions__icon--emerald {
  color: #6d7b8a;
  background: rgba(142, 170, 189, 0.12);
}

.dashboard-actions__icon--amber {
  color: #8f7035;
  background: rgba(210, 177, 111, 0.14);
}

.dashboard-actions__icon--violet {
  color: #80675a;
  background: rgba(179, 154, 137, 0.13);
}

.dashboard-actions__copy {
  flex: 1;
  min-width: 0;
}

.dashboard-actions__title,
.dashboard-actions__description {
  margin: 0;
}

.dashboard-actions__title {
  font-size: 0.86rem;
  font-weight: 600;
  color: #111827;
}

.dashboard-actions__description {
  margin-top: 3px;
  font-size: 0.75rem;
  line-height: 1.36;
  color: #6b7280;
}

.dashboard-actions__arrow {
  color: #6b7280;
  flex-shrink: 0;
}

:global(.dark) .dashboard-actions__item {
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.76), rgba(15, 23, 42, 0.92));
}

:global(.dark) .dashboard-actions__title {
  color: #f8fafc;
}

:global(.dark) .dashboard-actions__description,
:global(.dark) .dashboard-actions__arrow {
  color: #94a3b8;
}
</style>
