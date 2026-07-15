<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  actionPath: string
  isAuthenticated: boolean
}>()

const { t } = useI18n()

const services = ['api', 'workspace', 'invoice'] as const
</script>

<template>
  <section id="pricing" class="pricing-cta">
    <p class="pricing-cta__eyebrow">{{ t('home.landing.pricing.eyebrow') }}</p>
    <h2>{{ t('home.landing.pricing.title') }}</h2>
    <p class="pricing-cta__lead">{{ t('home.landing.pricing.lead') }}</p>

    <div class="pricing-cta__services">
      <div v-for="service in services" :key="service" class="pricing-cta__service">
        <span>{{ t(`home.landing.pricing.services.${service}.tag`) }}</span>
        <strong>{{ t(`home.landing.pricing.services.${service}.title`) }}</strong>
        <p>{{ t(`home.landing.pricing.services.${service}.body`) }}</p>
      </div>
    </div>

    <router-link class="pricing-cta__button" :to="props.actionPath">
      {{ props.isAuthenticated ? t('home.goToDashboard') : t('home.landing.pricing.cta') }}
      <Icon name="arrowRight" size="sm" />
    </router-link>
  </section>
</template>

<style scoped>
.pricing-cta {
  position: relative;
  overflow: hidden;
  scroll-margin-top: 90px;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0 132px;
  text-align: center;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.pricing-cta::before {
  content: "";
  position: absolute;
  left: 50%;
  top: 52%;
  width: 420px;
  height: 420px;
  transform: translate(-50%, -50%);
  border-radius: 50%;
  background-image: radial-gradient(circle, rgba(17, 17, 17, 0.18) 1px, transparent 1.4px);
  background-size: 14px 14px;
  mask-image: radial-gradient(circle, black 0 50%, transparent 70%);
  opacity: 0.28;
}

.pricing-cta__eyebrow,
.pricing-cta h2,
.pricing-cta__lead,
.pricing-cta__services,
.pricing-cta__button {
  position: relative;
  z-index: 1;
}

.pricing-cta__eyebrow {
  margin: 0 0 24px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.pricing-cta h2 {
  width: min(840px, 100%);
  margin: 0 auto;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 60px;
  font-weight: 900;
  line-height: 1.12;
}

:global(.dark) .pricing-cta h2 {
  color: #fffaf0;
}

.pricing-cta__lead {
  width: min(690px, 100%);
  margin: 28px auto 0;
  color: #5d5850;
  font-size: 17px;
  line-height: 1.9;
}

.pricing-cta__services {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px;
  margin-top: 56px;
}

.pricing-cta__service {
  min-height: 180px;
  padding: 24px;
  border: 1px solid rgba(17, 17, 17, 0.14);
  border-radius: 8px;
  background: rgba(250, 249, 246, 0.82);
  text-align: left;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease;
}

.pricing-cta__service:hover {
  transform: translateY(-6px);
  box-shadow: 0 24px 44px rgba(0, 0, 0, 0.1);
}

.pricing-cta__service span {
  color: #9a7a35;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 900;
}

.pricing-cta__service strong {
  display: block;
  margin-top: 18px;
  color: #111;
  font-size: 20px;
}

.pricing-cta__service p {
  margin: 14px 0 0;
  color: #5d5850;
  font-size: 14px;
  line-height: 1.75;
}

.pricing-cta__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 52px;
  margin-top: 46px;
  padding: 0 26px;
  color: #fff;
  border-radius: 999px;
  background: #090909;
  font-size: 15px;
  font-weight: 900;
  text-decoration: none;
  transition:
    transform 180ms ease,
    background 180ms ease;
}

.pricing-cta__button:hover {
  background: #2a261f;
  transform: translateY(-2px);
}

@media (max-width: 860px) {
  .pricing-cta {
    width: min(100% - 32px, 1180px);
    padding: 82px 0 92px;
  }

  .pricing-cta h2 {
    font-size: 40px;
  }

  .pricing-cta__services {
    grid-template-columns: 1fr;
  }
}
</style>
