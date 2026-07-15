<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    title: string
    value: string | number
    hint?: string
    detail?: string
    meta?: string
    tone?: 'neutral' | 'sky' | 'emerald' | 'amber' | 'violet'
  }>(),
  {
    hint: '',
    detail: '',
    meta: '',
    tone: 'neutral',
  }
)

const toneClassMap = {
  neutral: 'user-console-stat-card--neutral',
  sky: 'user-console-stat-card--sky',
  emerald: 'user-console-stat-card--emerald',
  amber: 'user-console-stat-card--amber',
  violet: 'user-console-stat-card--violet',
} as const

const toneClass = toneClassMap[props.tone]
</script>

<template>
  <article class="user-console-stat-card" :class="toneClass">
    <div class="user-console-stat-card__top">
      <div class="user-console-stat-card__label-wrap">
        <div v-if="$slots.icon" class="user-console-stat-card__icon" aria-hidden="true">
          <slot name="icon" />
        </div>
        <p class="user-console-stat-card__title">{{ title }}</p>
      </div>
      <p v-if="meta" class="user-console-stat-card__meta">{{ meta }}</p>
    </div>
    <p class="user-console-stat-card__value">{{ value }}</p>
    <p v-if="hint" class="user-console-stat-card__hint">{{ hint }}</p>
    <p v-if="detail" class="user-console-stat-card__detail">{{ detail }}</p>
  </article>
</template>

<style scoped>
.user-console-stat-card {
  min-height: 82px;
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 16px;
  padding: 11px 14px 10px;
  background: rgba(255, 255, 255, 0.68);
  box-shadow: 0 1px 0 rgba(36, 29, 22, 0.02);
}

.user-console-stat-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.user-console-stat-card__label-wrap {
  display: flex;
  min-width: 0;
  align-items: center;
}

.user-console-stat-card__icon {
  display: none;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 10px;
  background: rgba(127, 159, 152, 0.12);
  color: #6e958d;
  flex: 0 0 auto;
}

.user-console-stat-card__title {
  margin: 0;
  overflow: hidden;
  color: #6f6559;
  font-size: 13px;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-console-stat-card__meta {
  margin: 0;
  color: #7c5f43;
  font-size: 13px;
  line-height: 1.3;
  white-space: nowrap;
}

.user-console-stat-card__value {
  margin: 8px 0 0;
  color: #050505;
  font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
  font-size: 23px;
  font-weight: 700;
  line-height: 1.05;
}

.user-console-stat-card__hint,
.user-console-stat-card__detail {
  margin: 4px 0 0;
  font-size: 13px;
  line-height: 1.32;
}

.user-console-stat-card__hint {
  color: #5f584f;
}

.user-console-stat-card__detail {
  color: #7c7267;
}

.user-console-stat-card--neutral {
  border-color: rgba(17, 24, 39, 0.08);
}

.user-console-stat-card--sky .user-console-stat-card__icon {
  background: rgba(142, 170, 189, 0.14);
  color: #59788e;
}

.user-console-stat-card--emerald .user-console-stat-card__icon {
  background: rgba(127, 159, 152, 0.14);
  color: #5f847d;
}

.user-console-stat-card--amber .user-console-stat-card__icon {
  background: rgba(210, 177, 111, 0.16);
  color: #8f7035;
}

.user-console-stat-card--violet .user-console-stat-card__icon {
  background: rgba(179, 154, 137, 0.16);
  color: #80675a;
}

:global(.dark) .user-console-stat-card {
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.76), rgba(15, 23, 42, 0.92));
}

:global(.dark) .user-console-stat-card__icon {
  background: rgba(255, 255, 255, 0.06);
}

:global(.dark) .user-console-stat-card__title,
:global(.dark) .user-console-stat-card__meta {
  color: #94a3b8;
}

:global(.dark) .user-console-stat-card__value {
  color: #f8fafc;
}

:global(.dark) .user-console-stat-card__hint {
  color: #e2e8f0;
}

:global(.dark) .user-console-stat-card__detail {
  color: #94a3b8;
}
</style>
