<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <footer class="home-footer">
    <p>{{ t('home.landing.footer.tagline') }}</p>
    <small>{{ t('home.landing.footer.copyright') }}</small>
  </footer>
</template>

<style scoped>
.home-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 28px 0 34px;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

:global(.dark) .home-footer {
  border-top-color: rgba(255, 255, 255, 0.14);
}

.home-footer p {
  margin: 0;
  color: #8b857d;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
}

:global(.dark) .home-footer p {
  color: #9b9287;
}

.home-footer small {
  color: #9b958d;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 500;
}

@media (max-width: 760px) {
  .home-footer {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
    width: min(100% - 32px, 1180px);
    padding: 22px 0 30px;
  }
}
</style>
