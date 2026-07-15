<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const stats = [
  { valueKey: 'requestsValue', unitKey: 'requestsUnit', labelKey: 'requestsLabel' },
  { valueKey: 'availabilityValue', unitKey: 'availabilityUnit', labelKey: 'availabilityLabel' },
  { valueKey: 'latencyValue', unitKey: 'latencyUnit', labelKey: 'latencyLabel' }
] as const
</script>

<template>
  <section class="stats-strip" aria-label="VinzkAI stats">
    <div v-for="stat in stats" :key="stat.labelKey" class="stats-strip__item">
      <strong>
        <span>{{ t(`home.landing.stats.${stat.valueKey}`) }}</span>
        <small>{{ t(`home.landing.stats.${stat.unitKey}`) }}</small>
      </strong>
      <p>{{ t(`home.landing.stats.${stat.labelKey}`) }}</p>
    </div>
  </section>
</template>

<style scoped>
.stats-strip {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
  border-bottom: 1px solid rgba(20, 20, 20, 0.14);
}

:global(.dark) .stats-strip {
  border-color: rgba(255, 255, 255, 0.14);
}

.stats-strip__item {
  min-height: 178px;
  padding: 42px 34px 34px;
  text-align: center;
}

.stats-strip__item + .stats-strip__item {
  border-left: 1px solid rgba(20, 20, 20, 0.12);
}

:global(.dark) .stats-strip__item + .stats-strip__item {
  border-left-color: rgba(255, 255, 255, 0.12);
}

.stats-strip__item strong {
  display: inline-flex;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
  color: #090909;
  font-family: 'Times New Roman', SimSun, 'Songti SC', serif;
  font-size: clamp(38px, 5vw, 68px);
  font-weight: 900;
  line-height: 0.96;
  letter-spacing: 0;
  white-space: nowrap;
}

:global(.dark) .stats-strip__item strong {
  color: #fffaf0;
}

.stats-strip__item small {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: clamp(18px, 2vw, 28px);
  font-weight: 900;
}

.stats-strip__item p {
  margin: 18px 0 0;
  color: #777169;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

@media (max-width: 760px) {
  .stats-strip {
    grid-template-columns: 1fr;
    width: min(100% - 32px, 1180px);
  }

  .stats-strip__item {
    min-height: auto;
    padding: 30px 20px;
  }

  .stats-strip__item + .stats-strip__item {
    border-top: 1px solid rgba(20, 20, 20, 0.12);
    border-left: 0;
  }

  .stats-strip__item strong {
    font-size: 38px;
  }

  .stats-strip__item small {
    font-size: 18px;
  }
}
</style>
