<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const promises = [
  { icon: 'key', key: 'oneKey' },
  { icon: 'sync', key: 'dispatch' },
  { icon: 'chart', key: 'billing' },
  { icon: 'shield', key: 'service' }
] as const
</script>

<template>
  <section class="promise-section" aria-labelledby="promise-title">
    <p class="promise-section__eyebrow">{{ t('home.landing.promise.eyebrow') }}</p>
    <h2 id="promise-title" class="promise-section__title">
      {{ t('home.landing.promise.title') }}
    </h2>
    <div class="promise-section__copy">
      <p>{{ t('home.landing.promise.copyOne') }}</p>
      <p>{{ t('home.landing.promise.copyTwo') }}</p>
    </div>
    <div class="promise-section__items">
      <div v-for="item in promises" :key="item.key" class="promise-section__item">
        <Icon :name="item.icon" size="sm" />
        <span>{{ t(`home.landing.promise.items.${item.key}`) }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.promise-section {
  width: min(980px, calc(100% - 48px));
  margin: 0 auto;
  padding: 108px 0 82px;
  text-align: center;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

:global(.dark) .promise-section {
  border-top-color: rgba(255, 255, 255, 0.14);
}

.promise-section__eyebrow {
  margin: 0 0 28px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.promise-section__title {
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 58px;
  font-weight: 900;
  line-height: 1.16;
}

:global(.dark) .promise-section__title {
  color: #fffaf0;
}

.promise-section__copy {
  display: grid;
  gap: 18px;
  width: min(680px, 100%);
  margin: 54px auto 0;
  color: #5d5850;
  font-size: 17px;
  line-height: 1.95;
}

:global(.dark) .promise-section__copy {
  color: #c8c0b4;
}

.promise-section__items {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 34px;
  margin-top: 54px;
  padding-top: 30px;
  border-top: 1px solid rgba(20, 20, 20, 0.12);
}

:global(.dark) .promise-section__items {
  border-top-color: rgba(255, 255, 255, 0.12);
}

.promise-section__item {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #151515;
  font-size: 14px;
  font-weight: 800;
}

:global(.dark) .promise-section__item {
  color: #fffaf0;
}

@media (max-width: 760px) {
  .promise-section {
    width: min(100% - 32px, 980px);
    padding: 76px 0 58px;
  }

  .promise-section__title {
    font-size: 38px;
  }

  .promise-section__copy {
    margin-top: 32px;
    font-size: 15px;
  }
}
</style>
