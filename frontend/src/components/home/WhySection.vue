<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'

type ReasonKey = 'multi' | 'reliable' | 'billing' | 'enterprise' | 'market' | 'recharge'

const { t } = useI18n()
const activeReason = shallowRef<ReasonKey | null>(null)

const reasons = computed(() => [
  { key: 'multi' as const, label: '01', title: t('home.landing.why.items.multi.title'), tag: 'Multi-model' },
  { key: 'reliable' as const, label: '02', title: t('home.landing.why.items.reliable.title'), tag: 'Reliability' },
  { key: 'billing' as const, label: '03', title: t('home.landing.why.items.billing.title'), tag: 'Fair billing' },
  { key: 'enterprise' as const, label: '04', title: t('home.landing.why.items.enterprise.title'), tag: 'Service' },
  { key: 'market' as const, label: '05', title: t('home.landing.why.items.market.title'), tag: 'Model market' },
  { key: 'recharge' as const, label: '06', title: t('home.landing.why.items.recharge.title'), tag: 'Self recharge' }
])

const activeIndex = computed(() => {
  if (!activeReason.value) return -1
  return reasons.value.findIndex((reason) => reason.key === activeReason.value)
})

const preview = computed(() => {
  if (!activeReason.value) return null

  return {
    key: activeReason.value,
    title: t(`home.landing.why.items.${activeReason.value}.previewTitle`),
    body: t(`home.landing.why.items.${activeReason.value}.body`)
  }
})

const previewStyle = computed(() => ({
  '--preview-row': `${Math.min(Math.max(activeIndex.value, 0), 3)}`
}))

function setReason(reason: ReasonKey) {
  activeReason.value = reason
}

function clearReason() {
  activeReason.value = null
}
</script>

<template>
  <section class="why-section">
    <p class="why-section__eyebrow">{{ t('home.landing.why.eyebrow') }}</p>
    <h2>{{ t('home.landing.why.title') }}</h2>

    <div class="why-section__rows" @mouseleave="clearReason">
      <button
        v-for="reason in reasons"
        :key="reason.key"
        type="button"
        class="why-section__row"
        :class="{ 'is-active': activeReason === reason.key }"
        @mouseenter="setReason(reason.key)"
        @focus="setReason(reason.key)"
        @click="setReason(reason.key)"
      >
        <span class="why-section__index">{{ reason.label }}</span>
        <strong>{{ reason.title }}</strong>
        <small>{{ reason.tag }}</small>
        <p>{{ t(`home.landing.why.items.${reason.key}.body`) }}</p>
      </button>
    </div>

    <aside
      v-if="preview"
      class="why-section__preview"
      :class="`why-section__preview--${preview.key}`"
      :style="previewStyle"
      aria-live="polite"
    >
      <span>{{ preview.title }}</span>
      <p>{{ preview.body }}</p>
      <div class="why-section__map" :class="`why-section__map--${preview.key}`">
        <template v-if="preview.key === 'multi'">
          <i class="why-section__model-chip why-section__model-chip--one">GPT</i>
          <i class="why-section__model-chip why-section__model-chip--two">Claude</i>
          <i class="why-section__model-chip why-section__model-chip--three">Gemini</i>
          <i class="why-section__model-chip why-section__model-chip--four">image2</i>
          <em class="why-section__orbit"></em>
        </template>

        <template v-else-if="preview.key === 'reliable'">
          <i class="why-section__route why-section__route--one"></i>
          <i class="why-section__route why-section__route--two"></i>
          <i class="why-section__route why-section__route--three"></i>
          <span class="why-section__status-dot why-section__status-dot--one"></span>
          <span class="why-section__status-dot why-section__status-dot--two"></span>
          <span class="why-section__status-dot why-section__status-dot--three"></span>
        </template>

        <template v-else-if="preview.key === 'billing'">
          <i class="why-section__ledger why-section__ledger--one"></i>
          <i class="why-section__ledger why-section__ledger--two"></i>
          <i class="why-section__ledger why-section__ledger--three"></i>
          <span class="why-section__coin why-section__coin--one">T</span>
          <span class="why-section__coin why-section__coin--two">¥</span>
        </template>

        <template v-else-if="preview.key === 'enterprise'">
          <i class="why-section__invoice"></i>
          <span class="why-section__stamp">可开票</span>
          <span class="why-section__service-line why-section__service-line--one"></span>
          <span class="why-section__service-line why-section__service-line--two"></span>
        </template>

        <template v-else-if="preview.key === 'market'">
          <i class="why-section__market-card why-section__market-card--one">GPT-5.5</i>
          <i class="why-section__market-card why-section__market-card--two">Claude</i>
          <i class="why-section__market-card why-section__market-card--three">Gemini</i>
          <i class="why-section__market-card why-section__market-card--four">Seedance</i>
        </template>

        <template v-else>
          <span class="why-section__recharge-panel">
            <small>充值金额</small>
            <strong>¥100</strong>
          </span>
          <i class="why-section__recharge-flow why-section__recharge-flow--one"></i>
          <i class="why-section__recharge-flow why-section__recharge-flow--two"></i>
          <span class="why-section__recharge-button">自助支付</span>
          <span class="why-section__recharge-balance">余额 +100</span>
          <span class="why-section__recharge-done">到账</span>
        </template>

      </div>
    </aside>
  </section>
</template>

<style scoped>
.why-section {
  position: relative;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.why-section__eyebrow {
  margin: 0 0 24px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.why-section h2 {
  max-width: 560px;
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 52px;
  font-weight: 900;
  line-height: 1.16;
}

:global(.dark) .why-section h2 {
  color: #fffaf0;
}

.why-section__rows {
  margin-top: 58px;
  border-top: 2px solid #111;
}

.why-section__row {
  position: relative;
  overflow: hidden;
  display: grid;
  grid-template-columns: 72px 210px 130px minmax(0, 1fr);
  gap: 22px;
  align-items: center;
  width: 100%;
  min-height: 110px;
  padding: 22px 0;
  color: #111;
  border: 0;
  border-bottom: 1px solid rgba(20, 20, 20, 0.14);
  background: transparent;
  text-align: left;
  transition:
    background 180ms ease,
    transform 180ms ease;
}

.why-section__row::before {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(17, 17, 17, 0.055), transparent);
  opacity: 0;
  transform: translateX(-100%);
  pointer-events: none;
}

.why-section__row strong {
  font-family: SimSun, 'Songti SC', serif;
  font-size: 28px;
  font-weight: 900;
}

.why-section__row small,
.why-section__index {
  color: #8c867d;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.why-section__row p {
  margin: 0;
  color: #5d5850;
  font-size: 15px;
  line-height: 1.75;
}

.why-section__row:hover strong,
.why-section__row.is-active strong {
  text-decoration: underline;
  text-decoration-thickness: 2px;
  text-underline-offset: 8px;
}

.why-section__row:hover,
.why-section__row.is-active {
  background: rgba(255, 255, 255, 0.52);
  transform: translateX(8px);
}

.why-section__row:hover::before,
.why-section__row.is-active::before {
  opacity: 1;
  animation: why-row-sweep 1.4s ease-in-out infinite;
}

.why-section__preview {
  position: absolute;
  top: calc(318px + var(--preview-row) * 110px);
  right: 5%;
  width: 372px;
  padding: 18px;
  border: 1px solid rgba(17, 17, 17, 0.18);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.84);
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.12);
  transform: translateY(-42px);
  pointer-events: none;
  animation: why-preview-in 180ms ease-out both;
}

.why-section__preview span {
  color: #8a8174;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 800;
}

.why-section__preview p {
  margin: 14px 0 0;
  color: #4f4a43;
  font-size: 13px;
  line-height: 1.7;
}

.why-section__map {
  position: relative;
  overflow: hidden;
  height: 120px;
  margin-top: 18px;
  border: 1px solid rgba(17, 17, 17, 0.1);
  border-radius: 8px;
  background:
    linear-gradient(rgba(17, 17, 17, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(17, 17, 17, 0.04) 1px, transparent 1px);
  background-size: 28px 28px;
  animation: why-map-pan 8s linear infinite;
}

.why-section__map::before,
.why-section__map::after,
.why-section__map i,
.why-section__map em,
.why-section__map span {
  position: absolute;
}

.why-section__map::before,
.why-section__map::after {
  content: "";
  left: 22px;
  right: 22px;
  top: 50%;
  height: 1px;
  background: rgba(17, 17, 17, 0.18);
  transform-origin: left center;
  animation: why-line-sweep 2.4s ease-in-out infinite;
}

.why-section__map::after {
  transform: rotate(18deg);
  animation-delay: 260ms;
}

.why-section__model-chip,
.why-section__market-card,
.why-section__coin,
.why-section__stamp,
.why-section__recharge-panel,
.why-section__recharge-button,
.why-section__recharge-balance,
.why-section__recharge-done {
  z-index: 3;
  display: grid;
  place-items: center;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-style: normal;
  font-weight: 900;
}

.why-section__model-chip {
  min-width: 58px;
  height: 24px;
  padding: 0 8px;
  color: #111;
  border: 1px solid #111;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.86);
  font-size: 11px;
  animation: why-chip-orbit 3.2s ease-in-out infinite;
}

.why-section__model-chip--one {
  left: 28px;
  top: 22px;
}

.why-section__model-chip--two {
  right: 26px;
  top: 18px;
  animation-delay: 180ms;
}

.why-section__model-chip--three {
  left: 38px;
  bottom: 20px;
  animation-delay: 360ms;
}

.why-section__model-chip--four {
  right: 34px;
  bottom: 22px;
  animation-delay: 540ms;
}

.why-section__orbit {
  z-index: 1;
  inset: 16px 76px;
  border: 1px dashed rgba(17, 17, 17, 0.22);
  border-radius: 50%;
  animation: why-orbit-spin 7s linear infinite;
}

.why-section__preview--reliable .why-section__map::before,
.why-section__preview--reliable .why-section__map::after {
  animation-name: why-pulse-line;
}

.why-section__route {
  left: 36px;
  right: 36px;
  height: 3px;
  border-radius: 999px;
  background:
    linear-gradient(90deg, #111 0 34%, transparent 34% 44%, #111 44% 100%);
  background-size: 180% 100%;
  animation: why-route-flow 1.4s linear infinite;
}

.why-section__route--one {
  top: 30px;
  transform: rotate(-13deg);
}

.why-section__route--two {
  top: 58px;
  animation-delay: 160ms;
}

.why-section__route--three {
  top: 86px;
  transform: rotate(13deg);
  animation-delay: 320ms;
}

.why-section__status-dot {
  z-index: 3;
  width: 18px;
  height: 18px;
  border: 1px solid #111;
  border-radius: 50%;
  background: #fff;
  animation: why-dot-pulse 1.8s ease-in-out infinite;
}

.why-section__status-dot--one {
  left: 34px;
  top: 50px;
}

.why-section__status-dot--two {
  right: 42px;
  top: 24px;
  animation-delay: 220ms;
}

.why-section__status-dot--three {
  right: 48px;
  bottom: 22px;
  animation-delay: 440ms;
}

.why-section__ledger {
  left: 26px;
  right: 26px;
  height: 16px;
  border-bottom: 1px solid rgba(17, 17, 17, 0.22);
  background:
    linear-gradient(90deg, #111 0 28%, transparent 28% 36%, rgba(17, 17, 17, 0.5) 36% 52%, transparent 52% 66%, rgba(17, 17, 17, 0.28) 66% 100%);
  animation: why-ledger-slide 2.4s linear infinite;
}

.why-section__ledger--one {
  top: 24px;
}

.why-section__ledger--two {
  top: 52px;
  animation-delay: 160ms;
}

.why-section__ledger--three {
  top: 80px;
  animation-delay: 320ms;
}

.why-section__coin {
  width: 28px;
  height: 28px;
  color: #111;
  border: 1px solid #111;
  border-radius: 50%;
  background: #fff;
  font-size: 13px;
  animation: why-coin-drop 2.4s ease-in-out infinite;
}

.why-section__coin--one {
  right: 44px;
  top: 24px;
}

.why-section__coin--two {
  right: 78px;
  bottom: 22px;
  animation-delay: 360ms;
}

.why-section__invoice {
  z-index: 2;
  right: 34px;
  top: 18px;
  width: 92px;
  height: 74px;
  border: 1px solid #111;
  border-radius: 6px;
  background:
    linear-gradient(#111 0 0) 16px 18px / 48px 2px no-repeat,
    linear-gradient(#111 0 0) 16px 34px / 62px 2px no-repeat,
    linear-gradient(#111 0 0) 16px 50px / 38px 2px no-repeat,
    rgba(255, 255, 255, 0.78);
  animation: why-invoice-rise 2.8s ease-in-out infinite;
}

.why-section__stamp {
  right: 48px;
  bottom: 20px;
  padding: 5px 9px;
  color: #111;
  border: 1px solid #111;
  border-radius: 999px;
  background: #fff;
  font-size: 12px;
  animation: why-invoice-stamp 2.6s ease-in-out infinite;
}

.why-section__service-line {
  left: 28px;
  width: 118px;
  height: 2px;
  background: #111;
  transform-origin: left center;
  animation: why-service-line 2s ease-in-out infinite;
}

.why-section__service-line--one {
  top: 40px;
  transform: rotate(10deg);
}

.why-section__service-line--two {
  bottom: 36px;
  transform: rotate(-12deg);
  animation-delay: 240ms;
}

.why-section__market-card {
  min-width: 74px;
  height: 28px;
  padding: 0 8px;
  color: #111;
  border: 1px solid #111;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.86);
  font-size: 11px;
  animation: why-market-float 3s ease-in-out infinite;
}

.why-section__market-card--one {
  left: 24px;
  top: 22px;
}

.why-section__market-card--two {
  right: 28px;
  top: 28px;
  animation-delay: 160ms;
}

.why-section__market-card--three {
  left: 38px;
  bottom: 22px;
  animation-delay: 320ms;
}

.why-section__market-card--four {
  right: 42px;
  bottom: 18px;
  animation-delay: 480ms;
}

.why-section__recharge-panel {
  left: 24px;
  top: 20px;
  width: 114px;
  height: 72px;
  align-content: center;
  gap: 6px;
  color: #111;
  border: 1px solid #111;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.86);
  animation: why-recharge-card 2.8s ease-in-out infinite;
}

.why-section__recharge-panel small {
  color: #777169;
  font-size: 10px;
  font-weight: 900;
}

.why-section__recharge-panel strong {
  font-size: 24px;
  line-height: 1;
}

.why-section__recharge-flow {
  z-index: 2;
  left: 136px;
  width: 112px;
  height: 2px;
  border-radius: 999px;
  background:
    linear-gradient(90deg, #111 0 34%, transparent 34% 45%, #111 45% 100%);
  background-size: 180% 100%;
  animation: why-recharge-flow 1.5s linear infinite;
}

.why-section__recharge-flow--one {
  top: 42px;
}

.why-section__recharge-flow--two {
  top: 76px;
  animation-delay: 180ms;
}

.why-section__recharge-button {
  right: 26px;
  top: 24px;
  min-width: 84px;
  height: 30px;
  color: #fff;
  border-radius: 999px;
  background: #111;
  font-size: 11px;
  animation: why-recharge-pay 2.8s ease-in-out infinite;
}

.why-section__recharge-balance {
  right: 28px;
  bottom: 24px;
  min-width: 94px;
  height: 30px;
  color: #111;
  border: 1px solid #111;
  border-radius: 999px;
  background: #fff;
  font-size: 11px;
  animation: why-recharge-balance 2.8s ease-in-out infinite;
}

.why-section__recharge-done {
  left: 50%;
  bottom: 18px;
  min-width: 46px;
  height: 26px;
  color: #111;
  border: 1px solid #111;
  border-radius: 999px;
  background: #fff;
  font-size: 11px;
  transform: translateX(-50%) rotate(-7deg);
  animation: why-recharge-done 2.8s ease-in-out infinite;
}

@keyframes why-preview-in {
  from {
    opacity: 0;
    transform: translateY(-28px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(-42px) scale(1);
  }
}

@keyframes why-row-sweep {
  0% {
    transform: translateX(-100%);
  }

  100% {
    transform: translateX(100%);
  }
}

@keyframes why-map-pan {
  to {
    background-position:
      28px 28px,
      28px 28px;
  }
}

@keyframes why-line-sweep {
  0%,
  100% {
    scale: 0.35 1;
    opacity: 0.25;
  }

  50% {
    scale: 1 1;
    opacity: 0.75;
  }
}

@keyframes why-pulse-line {
  0%,
  100% {
    opacity: 0.18;
    transform: scaleX(0.42);
  }

  50% {
    opacity: 0.9;
    transform: scaleX(1);
  }
}

@keyframes why-dot-pulse {
  50% {
    transform: scale(1.16);
    box-shadow: 0 0 0 8px rgba(17, 17, 17, 0.05);
  }
}

@keyframes why-chip-orbit {
  50% {
    transform: translateY(-8px);
    box-shadow: 0 10px 22px rgba(17, 17, 17, 0.08);
  }
}

@keyframes why-orbit-spin {
  to {
    rotate: 360deg;
  }
}

@keyframes why-route-flow {
  to {
    background-position: -180% 0;
  }
}

@keyframes why-ledger-slide {
  50% {
    transform: translateX(10px);
    opacity: 0.58;
  }
}

@keyframes why-coin-drop {
  0%,
  30% {
    opacity: 0;
    transform: translateY(-14px) scale(0.82);
  }

  54%,
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes why-invoice-rise {
  50% {
    transform: translateY(-6px);
    box-shadow: 0 12px 24px rgba(17, 17, 17, 0.08);
  }
}

@keyframes why-invoice-stamp {
  0%,
  42% {
    transform: rotate(-8deg) scale(0.86);
  }

  58%,
  100% {
    transform: rotate(-8deg) scale(1);
  }
}

@keyframes why-service-line {
  0%,
  100% {
    scale: 0.72 1;
    opacity: 0.42;
  }

  50% {
    scale: 1 1;
    opacity: 1;
  }
}

@keyframes why-market-float {
  50% {
    transform: translateY(-7px);
  }
}

@keyframes why-recharge-card {
  50% {
    transform: translateY(-5px);
    box-shadow: 0 12px 24px rgba(17, 17, 17, 0.08);
  }
}

@keyframes why-recharge-flow {
  to {
    background-position: -180% 0;
  }
}

@keyframes why-recharge-pay {
  0%,
  32% {
    transform: translateY(0);
  }

  50%,
  74% {
    transform: translateY(-4px);
    box-shadow: 0 10px 22px rgba(17, 17, 17, 0.12);
  }
}

@keyframes why-recharge-balance {
  0%,
  48% {
    opacity: 0.28;
    transform: translateY(8px) scale(0.94);
  }

  66%,
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes why-recharge-done {
  0%,
  54% {
    opacity: 0;
    transform: translateX(-50%) rotate(-7deg) scale(0.82);
  }

  70%,
  100% {
    opacity: 1;
    transform: translateX(-50%) rotate(-7deg) scale(1);
  }
}

@media (max-width: 1080px) {
  .why-section__preview {
    position: static;
    width: 100%;
    margin-top: 28px;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .why-section__preview,
  .why-section__row::before,
  .why-section__map::before,
  .why-section__map::after,
  .why-section__map i,
  .why-section__map em,
  .why-section__map span {
    animation: none;
  }
}

@media (max-width: 860px) {
  .why-section {
    width: min(100% - 32px, 1180px);
    padding: 82px 0;
  }

  .why-section h2 {
    font-size: 38px;
  }

  .why-section__row {
    grid-template-columns: 44px 1fr;
    gap: 12px;
  }

  .why-section__row small,
  .why-section__row p {
    grid-column: 2;
  }
}
</style>
