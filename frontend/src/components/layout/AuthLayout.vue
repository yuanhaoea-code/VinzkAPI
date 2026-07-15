<script setup lang="ts">
import { computed, nextTick, onMounted, shallowRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const { t } = useI18n()
const route = useRoute()

const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const currentYear = computed(() => new Date().getFullYear())
const introAnimationKey = shallowRef(0)

const serviceItems = computed(() => [
  {
    key: 'api',
    label: t('authLayout.service.api.label'),
    value: t('authLayout.service.api.value')
  },
  {
    key: 'ai',
    label: t('authLayout.service.ai.label'),
    value: t('authLayout.service.ai.value')
  },
  {
    key: 'compute',
    label: t('authLayout.service.compute.label'),
    value: t('authLayout.service.compute.value')
  }
])

const capabilityItems = computed(() => [
  {
    key: 'api',
    label: t('authLayout.capabilities.api')
  },
  {
    key: 'workspace',
    label: t('authLayout.capabilities.workspace')
  },
  {
    key: 'support',
    label: t('authLayout.capabilities.support')
  }
])

function replayIntroAnimation(): void {
  introAnimationKey.value += 1
}

onMounted(() => {
  appStore.fetchPublicSettings()
  nextTick(replayIntroAnimation)
})

watch(
  () => route.fullPath,
  async () => {
    await nextTick()
    replayIntroAnimation()
  }
)
</script>

<template>
  <main class="auth-shell">
    <header class="auth-topbar">
      <router-link to="/home" class="auth-brand" aria-label="Vinzk AI home">
        <span class="auth-brand__mark">
          <img v-if="settingsLoaded && siteLogo" :src="siteLogo" alt="" />
          <b v-else>{{ t('authLayout.brandMark') }}</b>
        </span>
        <strong>{{ t('authLayout.brand') }}</strong>
      </router-link>

      <div class="auth-topbar__actions">
        <LocaleSwitcher />
        <router-link to="/home" class="auth-home-link">
          {{ t('authLayout.backHome') }}
        </router-link>
      </div>
    </header>

    <div class="auth-grid">
      <aside class="auth-story" aria-label="Vinzk AI service introduction">
        <div :key="`diagram-${introAnimationKey}`" class="auth-story__diagram" aria-hidden="true">
          <span class="auth-story__track"></span>
          <span class="auth-story__beam"></span>
          <span
            v-for="item in serviceItems"
            :key="`node-${item.key}`"
            class="auth-story__service-node"
            :class="`auth-story__service-node--${item.key}`"
          >
            <i></i>
          </span>
        </div>

        <div :key="`tags-${introAnimationKey}`" class="auth-story__tags">
          <span
            v-for="item in serviceItems"
            :key="item.key"
            :class="`auth-story__service-tag--${item.key}`"
          >
            <b>{{ item.label }}</b>
            {{ item.value }}
          </span>
        </div>

        <p class="auth-story__kicker">{{ t('authLayout.kicker') }}</p>
        <h1>
          <span>{{ t('authLayout.titleOne') }}</span>
          <em>{{ t('authLayout.titleTwo') }}</em>
        </h1>
        <p class="auth-story__lead">{{ t('authLayout.lead') }}</p>

        <ul class="auth-story__capabilities">
          <li v-for="item in capabilityItems" :key="item.key">
            <span class="auth-story__capability-icon" :class="`auth-story__capability-icon--${item.key}`">
              <i></i>
            </span>
            {{ item.label }}
          </li>
        </ul>

        <div class="auth-story__support">
          <p>{{ t('authLayout.support.label') }}</p>
          <span>{{ t('authLayout.support.text') }}</span>
          <strong>客服微信：13387544600</strong>
        </div>
      </aside>

      <section class="auth-panel" aria-label="Authentication form">
        <div class="auth-card">
          <div class="auth-card__body">
            <slot />
          </div>

          <div class="auth-card__footer">
            <slot name="footer" />
          </div>
        </div>
      </section>
    </div>

    <footer class="auth-copyright">
      <span>@{{ currentYear }} www.Vinzk.cn</span>
    </footer>
  </main>
</template>

<style scoped>
.auth-shell {
  position: relative;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  width: 100%;
  height: 100vh;
  height: 100dvh;
  min-height: 0;
  overflow: hidden;
  padding: 22px clamp(20px, 4vw, 72px);
  background:
    linear-gradient(90deg, rgba(245, 243, 239, 0.96), rgba(251, 250, 247, 0.92)),
    #f7f5f1;
  color: #111;
}

:global(.dark) .auth-shell {
  background:
    linear-gradient(90deg, rgba(9, 10, 12, 0.98), rgba(16, 16, 18, 0.96)),
    #090a0c;
  color: #f8f4eb;
}

.auth-topbar {
  position: relative;
  z-index: 2;
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: min(1280px, 100%);
  margin: 0 auto;
}

.auth-brand,
.auth-topbar__actions,
.auth-home-link {
  display: inline-flex;
  align-items: center;
}

.auth-brand {
  gap: 12px;
  color: inherit;
  text-decoration: none;
}

.auth-brand__mark {
  display: inline-grid;
  overflow: hidden;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(17, 17, 17, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.64);
}

:global(.dark) .auth-brand__mark {
  border-color: rgba(255, 255, 255, 0.13);
  background: rgba(255, 255, 255, 0.06);
}

.auth-brand__mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.auth-brand__mark b {
  font-family: SimSun, 'Songti SC', serif;
  font-size: 15px;
  font-weight: 900;
}

.auth-brand strong {
  font-family: SimSun, 'Songti SC', serif;
  font-size: 18px;
  font-weight: 900;
  letter-spacing: 0;
}

.auth-topbar__actions {
  gap: 18px;
}

.auth-home-link {
  gap: 6px;
  color: #69645d;
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
  transition: color 160ms ease;
}

.auth-home-link::before {
  content: "<-";
  color: #8d877f;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.auth-home-link:hover {
  color: #111;
}

:global(.dark) .auth-home-link {
  color: #c9c0b4;
}

:global(.dark) .auth-home-link:hover {
  color: #fffaf0;
}

.auth-grid {
  flex: 1 1 auto;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(420px, 480px);
  gap: clamp(42px, 7vw, 104px);
  align-items: center;
  width: min(1210px, 100%);
  min-height: 0;
  margin: 0 auto;
  padding: clamp(18px, 3vh, 34px) 0 22px;
}

.auth-story {
  max-width: 690px;
  min-width: 0;
  padding-top: 0;
}

.auth-story__diagram {
  position: relative;
  height: 104px;
  margin-bottom: 18px;
  overflow: visible;
}

.auth-story__track,
.auth-story__beam {
  position: absolute;
  top: 51px;
  right: 0;
  left: 0;
  height: 3px;
  border-radius: 999px;
}

.auth-story__track {
  background: rgba(17, 17, 17, 0.12);
}

.auth-story__track::before,
.auth-story__track::after {
  content: "";
  position: absolute;
  top: -3px;
  width: 7px;
  height: 7px;
  border: 1px solid rgba(17, 17, 17, 0.28);
  border-radius: 50%;
  background: #f7f5f1;
}

.auth-story__track::before {
  left: 0;
}

.auth-story__track::after {
  right: 0;
}

.auth-story__beam {
  overflow: visible;
  background: linear-gradient(90deg, #111 0%, #111 76%, rgba(17, 17, 17, 0.78) 90%, rgba(17, 17, 17, 0) 100%);
  box-shadow: 0 0 18px rgba(17, 17, 17, 0.18);
  transform: scaleX(0);
  transform-origin: left center;
  animation: auth-line-enter 2.85s linear 0.18s both;
}

.auth-story__beam::after {
  content: "";
  position: absolute;
  top: 50%;
  right: -5px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #111;
  box-shadow:
    0 0 0 5px rgba(17, 17, 17, 0.08),
    0 0 22px rgba(17, 17, 17, 0.28);
  transform: translateY(-50%) scale(0);
  animation: auth-beam-head-enter 2.85s linear 0.18s both;
}

:global(.dark) .auth-story__track {
  background: rgba(255, 250, 240, 0.13);
}

:global(.dark) .auth-story__track::before,
:global(.dark) .auth-story__track::after {
  border-color: rgba(255, 255, 255, 0.26);
  background: #090a0c;
}

:global(.dark) .auth-story__beam {
  background: linear-gradient(90deg, #fffaf0 0%, #fffaf0 76%, rgba(255, 250, 240, 0.78) 90%, rgba(255, 250, 240, 0) 100%);
  box-shadow: 0 0 20px rgba(255, 250, 240, 0.18);
}

:global(.dark) .auth-story__beam::after {
  background: #fffaf0;
  box-shadow:
    0 0 0 5px rgba(255, 250, 240, 0.1),
    0 0 22px rgba(255, 250, 240, 0.3);
}

.auth-story__service-node {
  position: absolute;
  top: 52px;
  display: grid;
  place-items: center;
  width: 46px;
  height: 46px;
  border: 1px solid rgba(17, 17, 17, 0.24);
  border-radius: 50%;
  background: rgba(247, 245, 241, 0.96);
  box-shadow:
    0 18px 36px rgba(42, 37, 31, 0.12),
    inset 0 0 0 7px rgba(255, 255, 255, 0.54);
  opacity: 0;
  filter: blur(5px);
  transform: translate(-50%, -50%) scale(0.84);
  will-change: opacity, filter, transform;
}

.auth-story__service-node i,
.auth-story__service-node i::before,
.auth-story__service-node i::after {
  position: absolute;
  content: "";
  display: block;
}

.auth-story__service-node--api {
  left: 16%;
  animation: auth-node-enter 560ms ease-out 0.62s both;
}

.auth-story__service-node--api i {
  width: 26px;
  height: 14px;
  border-top: 1px solid #111;
  border-bottom: 1px solid #111;
}

.auth-story__service-node--api i::before {
  top: 6px;
  left: -5px;
  width: 36px;
  height: 1px;
  background: #111;
}

.auth-story__service-node--api i::after {
  top: 2px;
  left: 4px;
  width: 6px;
  height: 6px;
  border: 1px solid #111;
  border-radius: 50%;
  box-shadow: 12px 4px 0 -1px #f7f5f1, 12px 4px 0 0 #111;
}

.auth-story__service-node--ai {
  left: 50%;
  animation: auth-node-enter 560ms ease-out 1.58s both;
}

.auth-story__service-node--ai i {
  width: 25px;
  height: 25px;
}

.auth-story__service-node--ai i::before {
  top: 11px;
  left: 2px;
  width: 21px;
  height: 1px;
  background: #111;
  transform: rotate(45deg);
  box-shadow: 0 0 0 0 #111;
}

.auth-story__service-node--ai i::after {
  top: 11px;
  left: 2px;
  width: 21px;
  height: 1px;
  background: #111;
  transform: rotate(-45deg);
  box-shadow:
    0 -7px 0 -0.25px #111,
    0 7px 0 -0.25px #111;
}

.auth-story__service-node--compute {
  left: 84%;
  animation: auth-node-enter 560ms ease-out 2.52s both;
}

.auth-story__service-node--compute i {
  width: 24px;
  height: 22px;
  border: 1px solid #111;
  border-radius: 3px;
}

.auth-story__service-node--compute i::before,
.auth-story__service-node--compute i::after {
  left: 5px;
  width: 14px;
  height: 1px;
  background: #111;
}

.auth-story__service-node--compute i::before {
  top: 7px;
}

.auth-story__service-node--compute i::after {
  top: 14px;
  box-shadow:
    -5px -7px 0 -0.25px #111,
    5px -7px 0 -0.25px #111,
    -5px 7px 0 -0.25px #111,
    5px 7px 0 -0.25px #111;
}

:global(.dark) .auth-story__service-node {
  border-color: rgba(255, 255, 255, 0.22);
  background: rgba(15, 15, 17, 0.94);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.35);
}

:global(.dark) .auth-story__service-node--api i,
:global(.dark) .auth-story__service-node--ai i,
:global(.dark) .auth-story__service-node--compute i {
  border-color: #fffaf0;
}

:global(.dark) .auth-story__service-node--api i::before,
:global(.dark) .auth-story__service-node--ai i::after,
:global(.dark) .auth-story__service-node--compute i::before,
:global(.dark) .auth-story__service-node--compute i::after {
  background: #fffaf0;
}

:global(.dark) .auth-story__service-node--api i::before {
  box-shadow: none;
}

:global(.dark) .auth-story__service-node--api i::after {
  border-color: #fffaf0;
  box-shadow: 12px 4px 0 -1px #0f0f11, 12px 4px 0 0 #fffaf0;
}

:global(.dark) .auth-story__service-node--ai i::after,
:global(.dark) .auth-story__service-node--compute i::before,
:global(.dark) .auth-story__service-node--compute i::after {
  box-shadow: none;
}

:global(.dark) .auth-story__service-node--ai i::before {
  background: #fffaf0;
  border-color: transparent;
}

:global(.dark) .auth-story__service-node--compute i::after {
  box-shadow:
    -5px -7px 0 -0.25px #fffaf0,
    5px -7px 0 -0.25px #fffaf0,
    -5px 7px 0 -0.25px #fffaf0,
    5px 7px 0 -0.25px #fffaf0;
}

.auth-story__tags {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 28px;
  color: #777169;
  font-size: 12px;
}

.auth-story__tags span {
  display: grid;
  gap: 8px;
  opacity: 0;
  filter: blur(5px);
  transform: translateY(-6px);
  will-change: opacity, filter, transform;
}

.auth-story__service-tag--api {
  animation: auth-tag-enter 600ms ease-out 0.78s both;
}

.auth-story__service-tag--ai {
  animation: auth-tag-enter 600ms ease-out 1.74s both;
}

.auth-story__service-tag--compute {
  animation: auth-tag-enter 600ms ease-out 2.68s both;
}

.auth-story__tags b {
  color: #272420;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

:global(.dark) .auth-story__tags {
  color: #aaa197;
}

:global(.dark) .auth-story__tags b {
  color: #f4efe7;
}

.auth-story__kicker {
  margin: 0 0 18px;
  color: #6c665f;
  font-size: 13px;
  letter-spacing: 0.46em;
}

.auth-story h1 {
  margin: 0;
  padding-bottom: 22px;
  border-bottom: 1px solid rgba(17, 17, 17, 0.14);
  font-family: SimSun, 'Songti SC', serif;
  font-size: clamp(42px, 4.9vw, 62px);
  font-weight: 900;
  line-height: 1.1;
  letter-spacing: 0;
}

.auth-story h1 span,
.auth-story h1 em {
  display: block;
  font-style: normal;
}

.auth-story h1 em {
  color: #5a5650;
}

:global(.dark) .auth-story h1 {
  border-bottom-color: rgba(255, 255, 255, 0.14);
}

:global(.dark) .auth-story h1 em {
  color: #bcb4aa;
}

.auth-story__lead {
  max-width: 620px;
  margin: 20px 0 0;
  color: #56514b;
  font-size: 15px;
  line-height: 1.82;
}

:global(.dark) .auth-story__lead {
  color: #c8c0b7;
}

.auth-story__capabilities {
  display: grid;
  gap: 12px;
  margin: 24px 0 0;
  padding: 0;
  color: #25221f;
  font-size: 14px;
  list-style: none;
}

.auth-story__capabilities li {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auth-story__capability-icon {
  position: relative;
  display: inline-grid;
  flex: 0 0 auto;
  place-items: center;
  width: 20px;
  height: 20px;
  border: 0;
  border-radius: 0;
}

.auth-story__capability-icon i,
.auth-story__capability-icon i::before,
.auth-story__capability-icon i::after {
  position: absolute;
  content: "";
  display: block;
}

.auth-story__capability-icon--api i {
  width: 18px;
  height: 10px;
  border-top: 1px solid #25221f;
  border-bottom: 1px solid #25221f;
}

.auth-story__capability-icon--api i::before {
  top: 4px;
  left: -2px;
  width: 22px;
  height: 1px;
  background: #25221f;
}

.auth-story__capability-icon--api i::after {
  top: 1px;
  left: 3px;
  width: 4px;
  height: 4px;
  border: 1px solid #25221f;
  border-radius: 50%;
  box-shadow: 8px 4px 0 -1px #f7f5f1, 8px 4px 0 0 #25221f;
}

.auth-story__capability-icon--workspace i {
  width: 18px;
  height: 18px;
}

.auth-story__capability-icon--workspace i::before,
.auth-story__capability-icon--workspace i::after {
  top: 8px;
  left: 2px;
  width: 14px;
  height: 1px;
  background: #25221f;
}

.auth-story__capability-icon--workspace i::before {
  transform: rotate(45deg);
  box-shadow: 0 -5px 0 -0.25px #25221f, 0 5px 0 -0.25px #25221f;
}

.auth-story__capability-icon--workspace i::after {
  transform: rotate(-45deg);
}

.auth-story__capability-icon--support i {
  width: 18px;
  height: 14px;
  border: 1px solid #25221f;
  border-radius: 3px;
}

.auth-story__capability-icon--support i::before {
  top: 4px;
  left: 4px;
  width: 8px;
  height: 1px;
  background: #25221f;
  box-shadow: 0 4px 0 #25221f;
}

.auth-story__capability-icon--support i::after {
  right: 2px;
  bottom: -5px;
  width: 6px;
  height: 5px;
  border-right: 1px solid #25221f;
  border-bottom: 1px solid #25221f;
  transform: skewY(-28deg);
}

:global(.dark) .auth-story__capabilities {
  color: #f5efe7;
}

:global(.dark) .auth-story__capability-icon--api i,
:global(.dark) .auth-story__capability-icon--support i {
  border-color: #f5efe7;
}

:global(.dark) .auth-story__capability-icon--api i::before,
:global(.dark) .auth-story__capability-icon--workspace i::before,
:global(.dark) .auth-story__capability-icon--workspace i::after,
:global(.dark) .auth-story__capability-icon--support i::before {
  background: #f5efe7;
}

:global(.dark) .auth-story__capability-icon--api i::after {
  border-color: #f5efe7;
  box-shadow: 8px 4px 0 -1px #090a0c, 8px 4px 0 0 #f5efe7;
}

:global(.dark) .auth-story__capability-icon--workspace i::before {
  box-shadow: 0 -5px 0 -0.25px #f5efe7, 0 5px 0 -0.25px #f5efe7;
}

:global(.dark) .auth-story__capability-icon--support i::after {
  border-color: #f5efe7;
}

.auth-story__support {
  margin-top: 34px;
  padding-top: 18px;
  border-top: 1px solid rgba(17, 17, 17, 0.14);
}

:global(.dark) .auth-story__support {
  border-top-color: rgba(255, 255, 255, 0.14);
}

.auth-story__support p {
  margin: 0 0 10px;
  color: #6f6962;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.auth-story__support p::before {
  content: "";
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 8px;
  border-radius: 50%;
  background: #5e5a54;
}

.auth-story__support span {
  display: block;
  max-width: 420px;
  color: #4c4741;
  font-size: 14px;
  line-height: 1.62;
}

.auth-story__support strong {
  display: block;
  margin-top: 12px;
  color: #111;
  font-family: 'Times New Roman', SimSun, 'Songti SC', serif;
  font-size: 28px;
  font-weight: 900;
  letter-spacing: 0.02em;
}

:global(.dark) .auth-story__support p,
:global(.dark) .auth-story__support span {
  color: #bdb4aa;
}

:global(.dark) .auth-story__support strong {
  color: #fffaf0;
}

.auth-panel {
  display: flex;
  min-height: 0;
  justify-content: center;
  padding-top: 0;
}

.auth-card {
  width: min(100%, 480px);
  max-height: calc(100vh - 112px);
  max-height: calc(100dvh - 112px);
  overflow: auto;
  padding: clamp(28px, 4vh, 42px) 42px 30px;
  border: 1px solid rgba(17, 17, 17, 0.16);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 28px 80px rgba(42, 37, 31, 0.09);
  scrollbar-width: thin;
}

:global(.dark) .auth-card {
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(18, 18, 20, 0.86);
  box-shadow: 0 28px 80px rgba(0, 0, 0, 0.34);
}

.auth-card__body {
  min-height: 0;
}

.auth-card__footer {
  margin-top: 24px;
  text-align: center;
  font-size: 13px;
}

.auth-copyright {
  position: absolute;
  right: clamp(20px, 4vw, 72px);
  bottom: 10px;
  display: block;
  width: auto;
  max-width: 1280px;
  margin: 0;
  color: #9a948d;
  font-size: 10px;
  text-align: right;
  pointer-events: none;
}

:deep(.auth-card__body > .space-y-6 > .text-center) {
  margin-bottom: 24px;
}

:deep(.auth-card__body > .space-y-6 > .text-center h2) {
  color: #111;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 28px;
  font-weight: 900;
  line-height: 1.2;
}

:deep(.auth-card__body > .space-y-6 > .text-center p) {
  margin-top: 12px;
  color: #6d675f;
  font-size: 14px;
}

:global(.dark) :deep(.auth-card__body > .space-y-6 > .text-center h2) {
  color: #fffaf0;
}

:global(.dark) :deep(.auth-card__body > .space-y-6 > .text-center p) {
  color: #bdb4aa;
}

.auth-card :deep(.input-label) {
  margin-bottom: 9px;
  color: #161412;
  font-size: 13px;
  font-weight: 800;
}

:global(.dark) .auth-card :deep(.input-label) {
  color: #f2ece3;
}

.auth-card :deep(.input) {
  min-height: 46px;
  border-color: rgba(17, 17, 17, 0.18);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.7);
  color: #111;
  box-shadow: none;
}

.auth-card :deep(.input:focus) {
  border-color: rgba(17, 17, 17, 0.55);
  box-shadow: 0 0 0 3px rgba(17, 17, 17, 0.06);
}

:global(.dark) .auth-card :deep(.input) {
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.06);
  color: #fffaf0;
}

:global(.dark) .auth-card :deep(.input:focus) {
  border-color: rgba(255, 255, 255, 0.48);
  box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.08);
}

.auth-card :deep(.btn-primary) {
  min-height: 48px;
  border-radius: 999px;
  background: #85827c;
  color: #fff;
  box-shadow: 0 12px 26px rgba(79, 76, 70, 0.22);
}

.auth-card :deep(.btn-primary:hover) {
  background: #6f6b65;
  box-shadow: 0 14px 32px rgba(79, 76, 70, 0.28);
}

.auth-card :deep(.btn-secondary) {
  min-height: 46px;
  border-color: rgba(17, 17, 17, 0.16);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  color: #1d1b18;
  box-shadow: 0 6px 18px rgba(42, 37, 31, 0.06);
}

.auth-card :deep(.btn-secondary:hover) {
  border-color: rgba(17, 17, 17, 0.32);
  background: #fff;
}

:global(.dark) .auth-card :deep(.btn-secondary) {
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.06);
  color: #fffaf0;
}

.auth-card :deep(a) {
  color: #2f2c28;
}

.auth-card :deep(a:hover) {
  color: #111;
}

:global(.dark) .auth-card :deep(a) {
  color: #eee6db;
}

.auth-card :deep(.input-hint) {
  color: #777169;
}

.auth-card :deep(.space-y-5 > :not([hidden]) ~ :not([hidden])) {
  margin-top: 18px;
}

.auth-card :deep(.space-y-6 > :not([hidden]) ~ :not([hidden])) {
  margin-top: 24px;
}

.auth-card :deep(.auth-user-type-select .select-trigger) {
  position: relative;
  z-index: 1;
  min-height: 46px;
  padding: 0 14px 0 44px;
  border-color: rgba(17, 17, 17, 0.18);
  border-radius: 10px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(250, 249, 246, 0.72)),
    rgba(255, 255, 255, 0.7);
  color: #111;
  box-shadow: none;
}

.auth-card :deep(.auth-user-type-select .select-trigger:hover) {
  border-color: rgba(17, 17, 17, 0.32);
  background: rgba(255, 255, 255, 0.9);
}

.auth-card :deep(.auth-user-type-select .select-trigger-open),
.auth-card :deep(.auth-user-type-select .select-trigger:focus) {
  border-color: rgba(17, 17, 17, 0.55);
  box-shadow: 0 0 0 3px rgba(17, 17, 17, 0.06);
}

.auth-card :deep(.auth-user-type-select .select-trigger-error) {
  border-color: rgba(180, 38, 38, 0.62);
  box-shadow: 0 0 0 3px rgba(180, 38, 38, 0.08);
}

.auth-card :deep(.auth-user-type-select .select-value) {
  color: #111;
  font-size: 14px;
}

.auth-card :deep(.auth-user-type-select .select-icon) {
  color: #6f6962;
}

.auth-card :deep(.auth-select-leading-icon) {
  z-index: 2;
}

.auth-card :deep(.auth-select-leading-icon svg) {
  opacity: 1;
}

:global(.dark) .auth-card :deep(.auth-user-type-select .select-trigger) {
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.06);
  color: #fffaf0;
}

:global(.dark) .auth-card :deep(.auth-user-type-select .select-trigger:hover),
:global(.dark) .auth-card :deep(.auth-user-type-select .select-trigger-open),
:global(.dark) .auth-card :deep(.auth-user-type-select .select-trigger:focus) {
  border-color: rgba(255, 255, 255, 0.48);
  box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.08);
}

:global(.dark) .auth-card :deep(.auth-user-type-select .select-value) {
  color: #fffaf0;
}

:global(.dark) .auth-card :deep(.auth-select-leading-icon) {
  z-index: 2;
}

:global(.auth-select-dropdown) {
  border-color: rgba(17, 17, 17, 0.16) !important;
  border-radius: 12px !important;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 247, 244, 0.96)) !important;
  box-shadow: 0 18px 44px rgba(42, 37, 31, 0.13) !important;
}

:global(.auth-select-dropdown .select-options) {
  padding: 6px !important;
}

:global(.auth-select-dropdown .select-option) {
  min-height: 38px !important;
  border-radius: 8px !important;
  padding: 9px 12px !important;
  color: #2b2722 !important;
  font-size: 13px !important;
}

:global(.auth-select-dropdown .select-option:hover),
:global(.auth-select-dropdown .select-option-focused) {
  background: rgba(17, 17, 17, 0.055) !important;
}

:global(.auth-select-dropdown .select-option-selected) {
  background: #111 !important;
  color: #fffaf0 !important;
}

:global(.auth-select-dropdown .select-option-selected:hover),
:global(.auth-select-dropdown .select-option-selected.select-option-focused) {
  background: #111 !important;
  color: #fffaf0 !important;
}

:global(.auth-select-dropdown .select-option-selected svg) {
  color: #fffaf0 !important;
}

:global(.dark .auth-select-dropdown) {
  border-color: rgba(255, 255, 255, 0.14) !important;
  background: rgba(20, 20, 22, 0.98) !important;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.38) !important;
}

:global(.dark .auth-select-dropdown .select-option) {
  color: #f5efe7 !important;
}

:global(.dark .auth-select-dropdown .select-option:hover),
:global(.dark .auth-select-dropdown .select-option-focused) {
  background: rgba(255, 255, 255, 0.08) !important;
}

:global(.dark .auth-select-dropdown .select-option-selected) {
  background: #fffaf0 !important;
  color: #111 !important;
}

:global(.dark .auth-select-dropdown .select-option-selected:hover),
:global(.dark .auth-select-dropdown .select-option-selected.select-option-focused) {
  background: #fffaf0 !important;
  color: #111 !important;
}

:global(.dark .auth-select-dropdown .select-option-selected svg) {
  color: #111 !important;
}

@keyframes auth-line-enter {
  0% {
    opacity: 0;
    transform: scaleX(0);
  }

  4% {
    opacity: 1;
  }

  46% {
    opacity: 1;
    transform: scaleX(0.46);
  }

  76% {
    opacity: 1;
    transform: scaleX(0.76);
  }

  100% {
    opacity: 1;
    transform: scaleX(1);
  }
}

@keyframes auth-beam-head-enter {
  0% {
    opacity: 0;
    transform: translateY(-50%) scale(0.5);
  }

  6% {
    opacity: 1;
    transform: translateY(-50%) scale(0.9);
  }

  92% {
    opacity: 1;
    transform: translateY(-50%) scale(1);
  }

  100% {
    opacity: 0;
    transform: translateY(-50%) scale(0.82);
  }
}

@keyframes auth-node-enter {
  0% {
    opacity: 0;
    filter: blur(5px);
    transform: translate(-50%, -50%) scale(0.84);
  }

  100% {
    opacity: 1;
    filter: blur(0);
    transform: translate(-50%, -50%) scale(1);
  }
}

@keyframes auth-tag-enter {
  0% {
    opacity: 0;
    filter: blur(5px);
    transform: translateY(-6px);
  }

  100% {
    opacity: 1;
    filter: blur(0);
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-story__beam,
  .auth-story__beam::after,
  .auth-story__service-node,
  .auth-story__tags span {
    animation: none;
    opacity: 1;
  }

  .auth-story__beam::after {
    opacity: 0;
  }

  .auth-story__beam {
    transform: scaleX(1);
  }

  .auth-story__service-node {
    filter: none;
    transform: translate(-50%, -50%) scale(1);
  }

  .auth-story__tags span {
    filter: none;
    transform: translateY(0);
  }
}

@media (max-width: 1080px) {
  .auth-grid {
    grid-template-columns: minmax(0, 520px);
    justify-content: center;
    gap: 0;
  }

  .auth-story {
    display: none;
  }
}

@media (max-width: 720px) {
  .auth-shell {
    padding: 18px 16px 22px;
  }

  .auth-topbar {
    gap: 16px;
  }

  .auth-brand strong {
    font-size: 16px;
  }

  .auth-topbar__actions {
    gap: 10px;
  }

  .auth-grid {
    padding-top: 16px;
  }

  .auth-panel {
    padding-top: 0;
  }

  .auth-home-link {
    font-size: 12px;
  }

  .auth-card {
    max-height: calc(100vh - 92px);
    max-height: calc(100dvh - 92px);
    padding: 30px 22px 24px;
    border-radius: 16px;
  }

  :deep(.auth-card__body > .space-y-6 > .text-center h2) {
    font-size: 24px;
  }

  .auth-copyright {
    flex-direction: column;
    gap: 6px;
  }
}
</style>
