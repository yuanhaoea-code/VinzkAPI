<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AnimatedTerminal from './AnimatedTerminal.vue'

interface TerminalLine {
  text: string
  tone?: 'normal' | 'muted' | 'ok' | 'warn'
}

const { t } = useI18n()

const steps = ['copy', 'paste', 'create'] as const
const terminalLines = computed<TerminalLine[]>(() => [
  { text: '$ vinzk connect --key sk-******' },
  { text: '✓ models loaded: GPT / Claude / Gemini', tone: 'ok' },
  { text: '✓ agent tools ready', tone: 'ok' },
  { text: '✓ image2 and seedance queued', tone: 'ok' },
  { text: `run: "${t('home.landing.onboarding.idea')}"`, tone: 'warn' }
])
</script>

<template>
  <section class="onboarding-section">
    <div class="onboarding-section__copy">
      <p class="onboarding-section__eyebrow">{{ t('home.landing.onboarding.eyebrow') }}</p>
      <h2>{{ t('home.landing.onboarding.title') }}</h2>
      <p>{{ t('home.landing.onboarding.lead') }}</p>

      <div class="onboarding-section__steps">
        <div v-for="(step, index) in steps" :key="step" class="onboarding-section__step">
          <span>{{ index + 1 }}</span>
          <div>
            <strong>{{ t(`home.landing.onboarding.steps.${step}.title`) }}</strong>
            <p>{{ t(`home.landing.onboarding.steps.${step}.body`) }}</p>
          </div>
        </div>
      </div>
    </div>

    <AnimatedTerminal
      class="onboarding-section__terminal"
      title="bash - 一念成事"
      :lines="terminalLines"
      :body-height="310"
      :cycle-pause-ms="2300"
    />
  </section>
</template>

<style scoped>
.onboarding-section {
  display: grid;
  grid-template-columns: minmax(0, 0.76fr) minmax(420px, 1.24fr);
  gap: 74px;
  align-items: center;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.onboarding-section__eyebrow {
  margin: 0 0 24px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.onboarding-section__copy h2 {
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 52px;
  font-weight: 900;
  line-height: 1.16;
}

.onboarding-section__copy > p {
  margin: 24px 0 0;
  color: #5d5850;
  font-size: 16px;
  line-height: 1.9;
}

.onboarding-section__steps {
  display: grid;
  gap: 0;
  margin-top: 48px;
}

.onboarding-section__step {
  display: grid;
  grid-template-columns: 38px 1fr;
  gap: 18px;
  min-height: 100px;
}

.onboarding-section__step span {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  color: #fff;
  border-radius: 50%;
  background: #111;
  font-weight: 800;
}

.onboarding-section__step:not(:last-child) span::after {
  content: "";
  width: 1px;
  height: 70px;
  margin-top: 8px;
  background: rgba(17, 17, 17, 0.16);
}

.onboarding-section__step strong {
  color: #111;
  font-size: 18px;
}

.onboarding-section__step p {
  margin: 10px 0 0;
  color: #716b62;
  font-size: 14px;
  line-height: 1.7;
}

.onboarding-section__terminal {
  min-height: 352px;
}

@media (max-width: 980px) {
  .onboarding-section {
    grid-template-columns: 1fr;
    width: min(100% - 32px, 1180px);
    padding: 82px 0;
  }

  .onboarding-section__copy h2 {
    font-size: 38px;
  }
}

@media (max-width: 560px) {
  .onboarding-section__terminal {
    min-height: 300px;
  }
}
</style>
