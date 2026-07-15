<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { scrollToHomeSection } from './useHomeAnchorScroll'

const props = defineProps<{
  actionPath: string
  isAuthenticated: boolean
}>()

const { t } = useI18n()

const tools = ['Codex CLI', 'Claude Code', 'Cline', 'Gemini CLI', 'Cursor', 'Continue'] as const

function handleCreativeClick(event: MouseEvent) {
  event.preventDefault()
  event.stopPropagation()
  scrollToHomeSection('#creative')
}
</script>

<template>
  <section id="top" class="hero-section">
    <div class="hero-section__orb" aria-hidden="true">
      <span class="hero-section__sphere"></span>
      <span class="hero-section__ring hero-section__ring--one"></span>
      <span class="hero-section__ring hero-section__ring--two"></span>
      <span class="hero-section__ring hero-section__ring--three"></span>
      <span class="hero-section__orbit hero-section__orbit--one"></span>
      <span class="hero-section__orbit hero-section__orbit--two"></span>
    </div>

    <p class="hero-section__eyebrow">{{ t('home.landing.hero.eyebrow') }}</p>
    <h1 class="hero-section__title">
      <span>{{ t('home.landing.brand') }}</span>
      <strong>{{ t('home.landing.hero.slogan') }}</strong>
    </h1>
    <p class="hero-section__subtitle">{{ t('home.landing.hero.subtitle') }}</p>

    <div class="hero-section__actions">
      <router-link class="hero-section__primary" :to="props.actionPath">
        {{ props.isAuthenticated ? t('home.goToDashboard') : t('home.landing.hero.primaryCta') }}
        <Icon name="arrowRight" size="sm" />
      </router-link>
      <a class="hero-section__secondary" href="#creative" @click="handleCreativeClick">
        {{ t('home.landing.hero.secondaryCta') }}
      </a>
    </div>

    <div class="hero-section__tools" aria-label="Works on">
      <span>Works on</span>
      <b v-for="tool in tools" :key="tool">{{ tool }}</b>
    </div>
  </section>
</template>

<style scoped>
.hero-section {
  position: relative;
  display: grid;
  justify-items: center;
  width: min(1180px, calc(100% - 48px));
  min-height: 680px;
  margin: 0 auto;
  padding: 168px 0 76px;
  text-align: center;
}

.hero-section__orb {
  position: absolute;
  top: 132px;
  left: 50%;
  width: 430px;
  height: 430px;
  perspective: 900px;
  transform-style: preserve-3d;
  transform: translateX(-50%);
  opacity: 0.86;
  animation: hero-orb-float 7s ease-in-out infinite;
}

.hero-section__sphere,
.hero-section__ring,
.hero-section__orbit {
  position: absolute;
  inset: 0;
  border-radius: 50%;
}

.hero-section__sphere {
  overflow: hidden;
  border: 1px solid rgba(17, 17, 17, 0.08);
  background:
    radial-gradient(circle at 34% 28%, rgba(255, 255, 255, 0.7), transparent 34%),
    radial-gradient(circle at 50% 50%, rgba(17, 17, 17, 0.28) 1px, transparent 1.5px),
    radial-gradient(circle at 50% 50%, transparent 0 56%, rgba(17, 17, 17, 0.1) 57%, transparent 59%),
    radial-gradient(circle, rgba(17, 17, 17, 0.04), transparent 68%);
  background-size:
    100% 100%,
    15px 15px,
    100% 100%,
    100% 100%;
  box-shadow:
    inset -28px -24px 64px rgba(17, 17, 17, 0.08),
    inset 28px 22px 72px rgba(255, 255, 255, 0.68);
  mask-image: radial-gradient(circle, black 0 70%, transparent 78%);
  transform: rotateX(62deg) rotateZ(-12deg);
  animation:
    hero-sphere-turn 18s linear infinite,
    hero-sphere-breathe 4.8s ease-in-out infinite;
}

.hero-section__sphere::before,
.hero-section__sphere::after {
  content: "";
  position: absolute;
  inset: 6%;
  border-radius: 50%;
  pointer-events: none;
}

.hero-section__sphere::before {
  background:
    radial-gradient(circle, rgba(17, 17, 17, 0.42) 0 1.1px, transparent 1.4px),
    radial-gradient(circle at 50% 50%, transparent 0 56%, rgba(17, 17, 17, 0.08) 57%, transparent 58%);
  background-size: 18px 18px, 100% 100%;
  mask-image: radial-gradient(circle, black 0 68%, transparent 73%);
  animation: hero-dot-current 9s linear infinite;
}

.hero-section__sphere::after {
  border: 1px solid rgba(17, 17, 17, 0.1);
  background:
    repeating-radial-gradient(ellipse at center, transparent 0 30px, rgba(17, 17, 17, 0.1) 31px 32px, transparent 33px 54px),
    repeating-linear-gradient(90deg, transparent 0 42px, rgba(17, 17, 17, 0.08) 43px 44px, transparent 45px 84px);
  opacity: 0.62;
  transform: rotateZ(16deg);
  animation: hero-grid-roll 13s linear infinite;
}

.hero-section__ring {
  border: 1px solid rgba(17, 17, 17, 0.1);
  transform-style: preserve-3d;
  animation: hero-ring-turn 16s linear infinite;
}

.hero-section__ring--one {
  transform: rotateX(68deg);
}

.hero-section__ring--two {
  transform: rotateY(68deg);
  animation-duration: 21s;
}

.hero-section__ring--three {
  transform: rotateX(28deg) rotateY(58deg);
  animation-duration: 26s;
  opacity: 0.7;
}

.hero-section__orbit {
  inset: 10%;
  border: 1px dashed rgba(17, 17, 17, 0.16);
  transform-style: preserve-3d;
  animation: hero-orbit-spin 11s linear infinite;
}

.hero-section__orbit--one {
  transform: rotateX(76deg) rotateZ(12deg);
}

.hero-section__orbit--two {
  inset: 18%;
  border-style: solid;
  opacity: 0.5;
  transform: rotateY(72deg) rotateZ(-18deg);
  animation-duration: 15s;
  animation-direction: reverse;
}

:global(.dark) .hero-section__sphere {
  border-color: rgba(246, 242, 233, 0.08);
  background:
    radial-gradient(circle at 34% 28%, rgba(255, 250, 240, 0.18), transparent 34%),
    radial-gradient(circle at 50% 50%, rgba(246, 242, 233, 0.3) 1px, transparent 1.5px),
    radial-gradient(circle at 50% 50%, transparent 0 56%, rgba(246, 242, 233, 0.13) 57%, transparent 59%),
    radial-gradient(circle, rgba(246, 242, 233, 0.04), transparent 68%);
  background-size:
    100% 100%,
    15px 15px,
    100% 100%,
    100% 100%;
}

:global(.dark) .hero-section__ring {
  border-color: rgba(246, 242, 233, 0.12);
}

:global(.dark) .hero-section__sphere::before {
  background:
    radial-gradient(circle, rgba(246, 242, 233, 0.48) 0 1.1px, transparent 1.4px),
    radial-gradient(circle at 50% 50%, transparent 0 56%, rgba(246, 242, 233, 0.08) 57%, transparent 58%);
  background-size: 18px 18px, 100% 100%;
}

:global(.dark) .hero-section__sphere::after {
  border-color: rgba(246, 242, 233, 0.11);
  background:
    repeating-radial-gradient(ellipse at center, transparent 0 30px, rgba(246, 242, 233, 0.11) 31px 32px, transparent 33px 54px),
    repeating-linear-gradient(90deg, transparent 0 42px, rgba(246, 242, 233, 0.08) 43px 44px, transparent 45px 84px);
}

:global(.dark) .hero-section__orbit {
  border-color: rgba(246, 242, 233, 0.18);
}

.hero-section__eyebrow,
.hero-section__title,
.hero-section__subtitle,
.hero-section__actions,
.hero-section__tools {
  position: relative;
  z-index: 1;
}

.hero-section__eyebrow {
  margin: 0 0 26px;
  color: #6c6962;
  font-size: 14px;
  font-weight: 700;
}

.hero-section__title {
  display: grid;
  gap: 22px;
  margin: 0;
  color: #090909;
  font-family: SimSun, 'Songti SC', 'Noto Serif CJK SC', serif;
  font-weight: 900;
  line-height: 1.02;
}

.hero-section__title span {
  font-size: 38px;
}

.hero-section__title strong {
  font-size: 84px;
  font-weight: inherit;
}

:global(.dark) .hero-section__title {
  color: #fffaf0;
}

.hero-section__subtitle {
  width: min(760px, 100%);
  margin: 28px 0 0;
  color: #55514a;
  font-size: 20px;
  line-height: 1.9;
}

:global(.dark) .hero-section__subtitle,
:global(.dark) .hero-section__eyebrow {
  color: #c8c0b4;
}

.hero-section__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 14px;
  margin-top: 34px;
}

.hero-section__primary,
.hero-section__secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 50px;
  padding: 0 24px;
  font-size: 15px;
  font-weight: 800;
  text-decoration: none;
  border-radius: 999px;
  transition:
    transform 180ms ease,
    background 180ms ease,
    border-color 180ms ease;
}

.hero-section__primary {
  color: #fff;
  background: #090909;
}

.hero-section__primary:hover,
.hero-section__secondary:hover {
  transform: translateY(-2px);
}

.hero-section__primary:hover {
  background: #2a261f;
}

.hero-section__secondary {
  color: #111;
  border: 1px solid rgba(17, 17, 17, 0.18);
  background: rgba(250, 249, 246, 0.72);
}

:global(.dark) .hero-section__secondary {
  color: #fffaf0;
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.04);
}

.hero-section__tools {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 13px;
  margin-top: 48px;
  color: #6c6962;
  font-size: 13px;
}

.hero-section__tools span {
  color: #9a968d;
  text-transform: uppercase;
}

.hero-section__tools b {
  font-weight: 600;
}

.hero-section__tools b::after {
  content: "·";
  margin-left: 13px;
  color: #bbb5a8;
}

.hero-section__tools b:last-child::after {
  content: "";
  margin: 0;
}

@keyframes hero-sphere-turn {
  to {
    background-position:
      0 0,
      90px 0,
      0 0,
      0 0;
    transform: rotateX(62deg) rotateZ(348deg);
  }
}

@keyframes hero-orb-float {
  50% {
    transform: translateX(-50%) translateY(-12px);
  }
}

@keyframes hero-dot-current {
  to {
    background-position:
      72px 18px,
      0 0;
  }
}

@keyframes hero-grid-roll {
  to {
    transform: rotateZ(376deg);
  }
}

@keyframes hero-sphere-breathe {
  50% {
    filter: drop-shadow(0 20px 36px rgba(17, 17, 17, 0.08));
  }
}

@keyframes hero-orbit-spin {
  to {
    rotate: -360deg;
  }
}

@keyframes hero-ring-turn {
  to {
    rotate: 360deg;
  }
}

@media (max-width: 860px) {
  .hero-section {
    width: min(100% - 32px, 1180px);
    min-height: 620px;
    padding-top: 118px;
  }

  .hero-section__orb {
    top: 100px;
    width: 340px;
    height: 340px;
  }

  .hero-section__title strong {
    font-size: 56px;
  }

  .hero-section__title span {
    font-size: 30px;
  }

  .hero-section__subtitle {
    font-size: 17px;
  }
}

@media (max-width: 520px) {
  .hero-section {
    min-height: 560px;
    padding-top: 92px;
  }

  .hero-section__orb {
    width: 280px;
    height: 280px;
  }

  .hero-section__title strong {
    font-size: 36px;
  }

  .hero-section__subtitle {
    font-size: 15px;
    line-height: 1.75;
  }

  .hero-section__tools {
    margin-top: 32px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .hero-section__sphere,
  .hero-section__ring,
  .hero-section__orbit {
    animation: none;
  }
}
</style>
