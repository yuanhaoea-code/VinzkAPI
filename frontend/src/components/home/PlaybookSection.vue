<script setup lang="ts">
import { computed, onMounted, onUnmounted, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'

type ModeKey = 'signin' | 'agent' | 'token' | 'code'

const { t } = useI18n()

const activeMode = shallowRef<ModeKey>('signin')
const modeOrder: ModeKey[] = ['signin', 'agent', 'token', 'code']
const userHoldUntil = shallowRef(0)
let rotateTimer: ReturnType<typeof setInterval> | undefined

const modes = computed(() => [
  { key: 'signin' as const, channel: 'CH.01', title: t('home.landing.playbook.modes.signin.title') },
  { key: 'agent' as const, channel: 'CH.02', title: t('home.landing.playbook.modes.agent.title') },
  { key: 'token' as const, channel: 'CH.03', title: t('home.landing.playbook.modes.token.title') },
  { key: 'code' as const, channel: 'CH.04', title: t('home.landing.playbook.modes.code.title') }
])

const selectedMode = computed(() => modes.value.find((mode) => mode.key === activeMode.value) ?? modes.value[0])

function setMode(mode: ModeKey, hold = false) {
  activeMode.value = mode
  if (hold) {
    userHoldUntil.value = Date.now() + 9000
  }
}

function rotateMode() {
  if (Date.now() < userHoldUntil.value) return

  const currentIndex = modeOrder.indexOf(activeMode.value)
  const nextIndex = (currentIndex + 1) % modeOrder.length
  activeMode.value = modeOrder[nextIndex]
}

onMounted(() => {
  rotateTimer = setInterval(rotateMode, 3600)
})

onUnmounted(() => {
  if (rotateTimer) {
    clearInterval(rotateTimer)
  }
})
</script>

<template>
  <section class="playbook-section">
    <div class="playbook-section__copy">
      <p class="playbook-section__eyebrow">{{ t('home.landing.playbook.eyebrow') }}</p>
      <h2>{{ t('home.landing.playbook.title') }}</h2>
      <p>{{ t('home.landing.playbook.lead') }}</p>
    </div>

    <div class="playbook-section__machine">
      <div class="playbook-section__antenna" aria-hidden="true"></div>
      <div class="playbook-section__screen">
        <div class="playbook-section__screen-title">
          {{ selectedMode.channel }} - {{ selectedMode.title }}
        </div>

        <div
          :key="activeMode"
          class="playbook-section__scene"
          :class="`playbook-section__scene--${activeMode}`"
        >
          <template v-if="activeMode === 'signin'">
            <div class="playbook-section__signin-card">
              <span>DAILY CHECK-IN</span>
              <strong>100万 Token</strong>
              <b>已签到</b>
            </div>
            <div class="playbook-section__signin-reward">
              <span>体验额度</span>
              <strong>+ ¥0.25</strong>
            </div>
            <div class="playbook-section__tap-ring"></div>
            <div class="playbook-section__cursor"></div>
            <i class="playbook-section__coin playbook-section__coin--one">T</i>
            <i class="playbook-section__coin playbook-section__coin--two">T</i>
            <i class="playbook-section__coin playbook-section__coin--three">T</i>
            <p>{{ t('home.landing.playbook.modes.signin.caption') }}</p>
          </template>

          <template v-else-if="activeMode === 'agent'">
            <div class="playbook-section__agent-flow">
              <div class="playbook-section__agent-wire">
                <i></i>
                <i></i>
                <i></i>
              </div>
              <div class="playbook-section__agent-card playbook-section__agent-card--owner">
                <span>你的邀请链接</span>
                <strong>vinzk.cn/i/8K</strong>
                <b>分享给好友</b>
              </div>
              <div class="playbook-section__agent-card playbook-section__agent-card--order">
                <span>好友消费</span>
                <strong>¥100</strong>
                <b>API 调用 / 网页对话</b>
              </div>
              <div class="playbook-section__agent-card playbook-section__agent-card--wallet">
                <span>返利自动入账</span>
                <strong>5%</strong>
                <b>+ ¥5.00 余额</b>
              </div>
              <div class="playbook-section__agent-ledger">
                <span><i></i>好友付款成功</span>
                <span><i></i>佣金记录可查</span>
                <span><i></i>余额实时到账</span>
              </div>
              <em class="playbook-section__agent-cash playbook-section__agent-cash--one">+¥5</em>
              <em class="playbook-section__agent-cash playbook-section__agent-cash--two">5%</em>
              <div class="playbook-section__agent-team">
                <span>好友 A</span>
                <span>好友 B</span>
                <span>好友 C</span>
              </div>
            </div>
            <p>{{ t('home.landing.playbook.modes.agent.caption') }}</p>
          </template>

          <template v-else-if="activeMode === 'token'">
            <div class="playbook-section__rank">
              <div class="playbook-section__rank-row playbook-section__rank-row--one">
                <span>01</span>
                <strong>1.28亿 Token</strong>
                <em>+10元额度</em>
                <i></i>
              </div>
              <div class="playbook-section__rank-row playbook-section__rank-row--two">
                <span>02</span>
                <strong>9,600万 Token</strong>
                <em>+5元额度</em>
                <i></i>
              </div>
              <div class="playbook-section__rank-row playbook-section__rank-row--three">
                <span>03</span>
                <strong>7,300万 Token</strong>
                <em>+3元额度</em>
                <i></i>
              </div>
              <b class="playbook-section__crown">月度 Token 消耗榜奖励</b>
            </div>
            <p>{{ t('home.landing.playbook.modes.token.caption') }}</p>
          </template>

          <template v-else>
            <div class="playbook-section__codebox">
              <span style="--line-width: 38ch; --line-steps: 38; --line-delay: 0ms">$ build app --idea "{{ t('home.landing.playbook.modes.code.prompt') }}"</span>
              <span style="--line-width: 22ch; --line-steps: 22; --line-delay: 720ms">[ok] scaffold project</span>
              <span style="--line-width: 20ch; --line-steps: 20; --line-delay: 1440ms">[ok] write api client</span>
              <span style="--line-width: 17ch; --line-steps: 17; --line-delay: 2160ms">[ok] generate UI</span>
              <span style="--line-width: 14ch; --line-steps: 14; --line-delay: 2880ms">[ok] run tests</span>
            </div>
            <p>{{ t('home.landing.playbook.modes.code.caption') }}</p>
          </template>
        </div>
      </div>

      <aside class="playbook-section__channels" aria-label="Playbook channels">
        <span class="playbook-section__channels-title">Channels</span>
        <button
          v-for="mode in modes"
          :key="mode.key"
          type="button"
          :class="{ 'is-active': activeMode === mode.key }"
          @mouseenter="setMode(mode.key, true)"
          @focus="setMode(mode.key, true)"
          @click="setMode(mode.key, true)"
        >
          <small>{{ mode.channel.replace('CH.', '') }}</small>
          {{ mode.title }}
        </button>
        <div class="playbook-section__dial" aria-hidden="true"></div>
        <span class="playbook-section__dial-label">{{ t('home.landing.playbook.switch') }}</span>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.playbook-section {
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.playbook-section__copy {
  max-width: 620px;
}

.playbook-section__eyebrow {
  margin: 0 0 24px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.playbook-section__copy h2 {
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 52px;
  font-weight: 900;
  line-height: 1.16;
}

:global(.dark) .playbook-section__copy h2 {
  color: #fffaf0;
}

.playbook-section__copy p {
  margin: 24px 0 0;
  color: #5d5850;
  font-size: 16px;
  line-height: 1.9;
}

.playbook-section__machine {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 148px;
  gap: 22px;
  margin-top: 70px;
  padding: 22px;
  border: 1px solid #111;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.42);
}

:global(.dark) .playbook-section__machine {
  border-color: rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.04);
}

.playbook-section__antenna {
  position: absolute;
  top: -60px;
  left: 50%;
  width: 120px;
  height: 60px;
  transform: translateX(-50%);
}

.playbook-section__antenna::before,
.playbook-section__antenna::after {
  content: "";
  position: absolute;
  bottom: 0;
  width: 2px;
  height: 70px;
  background: #111;
  transform-origin: bottom;
}

.playbook-section__antenna::before {
  left: 50%;
  transform: rotate(-45deg);
}

.playbook-section__antenna::after {
  right: 50%;
  transform: rotate(45deg);
}

.playbook-section__screen {
  position: relative;
  overflow: hidden;
  min-height: 360px;
  border-radius: 8px;
  background:
    repeating-linear-gradient(0deg, rgba(255, 255, 255, 0.025) 0 1px, transparent 1px 4px),
    #060606;
}

.playbook-section__screen::after {
  content: "";
  position: absolute;
  inset: 0;
  z-index: 1;
  background: linear-gradient(90deg, transparent, rgba(242, 195, 107, 0.12), transparent);
  opacity: 0.36;
  transform: translateX(-120%);
  animation: playbook-monitor-scan 5.2s ease-in-out infinite;
  pointer-events: none;
}

.playbook-section__screen-title {
  position: absolute;
  z-index: 3;
  top: 16px;
  left: 18px;
  color: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 800;
}

.playbook-section__scene {
  position: absolute;
  z-index: 2;
  inset: 0;
  display: grid;
  place-items: center;
  color: #fff;
}

.playbook-section__scene::before {
  content: "";
  position: absolute;
  inset: 0;
  opacity: 0.42;
  background:
    radial-gradient(circle at 50% 50%, rgba(255, 255, 255, 0.14), transparent 34%),
    linear-gradient(110deg, transparent 0 42%, rgba(255, 255, 255, 0.12) 48%, transparent 55%);
  transform: translateX(-80%);
  animation: playbook-scene-sweep 3.8s ease-in-out infinite;
}

.playbook-section__scene--agent::before {
  background:
    radial-gradient(circle at 20% 52%, rgba(255, 255, 255, 0.12), transparent 18%),
    radial-gradient(circle at 78% 28%, rgba(255, 255, 255, 0.12), transparent 18%),
    linear-gradient(110deg, transparent 0 42%, rgba(255, 255, 255, 0.1) 48%, transparent 55%);
}

.playbook-section__scene--token::before {
  background:
    repeating-linear-gradient(90deg, transparent 0 52px, rgba(255, 255, 255, 0.06) 53px 54px),
    linear-gradient(110deg, transparent 0 42%, rgba(242, 195, 107, 0.16) 48%, transparent 55%);
}

.playbook-section__scene--code::before {
  background:
    linear-gradient(rgba(255, 255, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.045) 1px, transparent 1px);
  background-size: 28px 28px;
  transform: none;
  animation: code-grid-pan 8s linear infinite;
}

.playbook-section__scene p {
  position: absolute;
  left: 18px;
  bottom: 14px;
  margin: 0;
  color: #a7a39b;
  font-size: 12px;
}

.playbook-section__signin-card {
  position: relative;
  display: grid;
  grid-template-rows: auto auto;
  gap: 14px;
  align-content: center;
  justify-items: center;
  box-sizing: border-box;
  width: 264px;
  height: 150px;
  padding: 22px 18px 20px;
  border: 4px solid #fff;
  border-top-width: 18px;
  background: rgba(255, 255, 255, 0.03);
  animation: signin-card-pop 2.8s ease-in-out infinite;
}

.playbook-section__signin-card span,
.playbook-section__signin-card strong {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.playbook-section__signin-card span {
  color: #a7a39b;
  font-size: 11px;
  letter-spacing: 0.08em;
}

.playbook-section__signin-card strong {
  color: #fff;
  font-size: 20px;
  letter-spacing: 0;
  white-space: nowrap;
}

.playbook-section__signin-card b {
  position: absolute;
  top: 12px;
  right: 12px;
  padding: 6px 10px;
  color: #111;
  border-radius: 999px;
  background: #fff;
  font-size: 13px;
  transform: rotate(-8deg) scale(0);
  animation: signin-stamp 2.8s ease-in-out infinite;
}

.playbook-section__signin-reward {
  position: absolute;
  left: 50%;
  top: 67%;
  z-index: 4;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  color: #111;
  border-radius: 999px;
  background: #fff;
  box-shadow: 0 12px 28px rgba(255, 255, 255, 0.16);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  opacity: 1;
  transform: translate(-50%, 0) scale(1);
  animation: signin-reward-float 2.8s ease-in-out infinite;
}

.playbook-section__signin-reward span {
  color: #5f5a52;
  font-size: 11px;
  font-weight: 800;
}

.playbook-section__signin-reward strong {
  font-size: 16px;
}

.playbook-section__tap-ring {
  position: absolute;
  width: 64px;
  height: 64px;
  border: 2px solid #fff;
  border-radius: 50%;
  transform: translate(92px, 54px) scale(0.4);
  animation: signin-tap-ring 2.8s ease-in-out infinite;
}

.playbook-section__coin {
  position: absolute;
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  color: #111;
  border-radius: 50%;
  background: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 16px;
  font-style: normal;
  font-weight: 900;
  opacity: 0;
  animation: signin-coin 2.8s ease-in-out infinite;
}

.playbook-section__coin--one {
  transform: translate(-96px, 22px);
}

.playbook-section__coin--two {
  animation-delay: 140ms;
  transform: translate(-62px, -58px);
}

.playbook-section__coin--three {
  animation-delay: 280ms;
  transform: translate(108px, -26px);
}

.playbook-section__agent-flow {
  position: absolute;
  z-index: 2;
  inset: 62px 54px 54px;
  display: grid;
  grid-template-columns: minmax(150px, 0.98fr) minmax(122px, 0.72fr) minmax(160px, 0.98fr);
  gap: 28px;
  align-items: center;
  isolation: isolate;
}

.playbook-section__agent-flow::before {
  content: "";
  position: absolute;
  inset: 22% -2% 18%;
  z-index: -1;
  border-radius: 999px;
  background:
    radial-gradient(circle at 16% 50%, rgba(255, 255, 255, 0.16), transparent 18%),
    radial-gradient(circle at 50% 50%, rgba(242, 195, 107, 0.16), transparent 18%),
    radial-gradient(circle at 84% 50%, rgba(142, 240, 165, 0.16), transparent 18%);
  filter: blur(16px);
  opacity: 0.85;
}

.playbook-section__agent-card {
  position: relative;
  z-index: 3;
  display: grid;
  gap: 8px;
  min-height: 132px;
  padding: 18px 18px 16px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.055);
  box-shadow:
    inset 0 1px rgba(255, 255, 255, 0.08),
    0 18px 42px rgba(0, 0, 0, 0.24);
  backdrop-filter: blur(10px);
  animation: agent-card-rise 4.8s cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.playbook-section__agent-card--order {
  min-height: 118px;
  text-align: center;
  animation-delay: 260ms;
}

.playbook-section__agent-card--wallet {
  color: #111;
  border-color: rgba(255, 255, 255, 0.72);
  background:
    linear-gradient(135deg, #fff, #ebe6dc),
    #fff;
  animation-delay: 520ms;
}

.playbook-section__agent-card span,
.playbook-section__agent-card b,
.playbook-section__agent-ledger,
.playbook-section__agent-team,
.playbook-section__agent-cash {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.playbook-section__agent-card span {
  color: #a9a49b;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.playbook-section__agent-card strong {
  color: #fff;
  font-size: 23px;
  line-height: 1.05;
}

.playbook-section__agent-card b {
  color: #d6d0c5;
  font-size: 12px;
}

.playbook-section__agent-card--wallet span,
.playbook-section__agent-card--wallet b {
  color: #6b665f;
}

.playbook-section__agent-card--wallet strong {
  color: #111;
  font-size: 48px;
  line-height: 0.92;
  letter-spacing: -0.02em;
}

.playbook-section__agent-wire {
  position: absolute;
  z-index: 1;
  left: 13%;
  right: 13%;
  top: 50%;
  height: 2px;
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.18), #fff 48%, rgba(255, 255, 255, 0.18));
  transform: translateY(-50%);
}

.playbook-section__agent-wire::before,
.playbook-section__agent-wire::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 14px;
  height: 14px;
  border: 2px solid #fff;
  border-radius: 50%;
  background: #060606;
  transform: translateY(-50%);
}

.playbook-section__agent-wire::before {
  left: 31%;
}

.playbook-section__agent-wire::after {
  right: 31%;
}

.playbook-section__agent-wire i {
  position: absolute;
  top: 50%;
  width: 58px;
  height: 10px;
  border-radius: 999px;
  background: linear-gradient(90deg, transparent, #fff, transparent);
  opacity: 0;
  transform: translate3d(-42px, -50%, 0);
  animation: agent-flow-packet 2.4s cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.playbook-section__agent-wire i:nth-child(2) {
  animation-delay: 480ms;
}

.playbook-section__agent-wire i:nth-child(3) {
  animation-delay: 960ms;
}

.playbook-section__agent-ledger {
  position: absolute;
  z-index: 4;
  right: 7%;
  bottom: 2%;
  display: grid;
  gap: 6px;
  min-width: 178px;
  padding: 12px;
  color: #191919;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 18px 32px rgba(0, 0, 0, 0.24);
  font-size: 11px;
  font-weight: 900;
  animation: agent-ledger-in 4.8s cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.playbook-section__agent-ledger span {
  display: flex;
  align-items: center;
  gap: 7px;
}

.playbook-section__agent-ledger i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #111;
}

.playbook-section__agent-cash {
  position: absolute;
  z-index: 5;
  display: grid;
  place-items: center;
  min-width: 54px;
  height: 32px;
  padding: 0 10px;
  color: #111;
  border-radius: 999px;
  background: #fff;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.5),
    0 14px 32px rgba(255, 255, 255, 0.12);
  font-size: 13px;
  font-style: normal;
  font-weight: 900;
  animation: agent-cash-travel 4.8s cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.playbook-section__agent-cash--one {
  left: 34%;
  top: 31%;
}

.playbook-section__agent-cash--two {
  right: 23%;
  top: 27%;
  animation-delay: 520ms;
}

.playbook-section__agent-team {
  position: absolute;
  z-index: 2;
  left: 12%;
  bottom: 6%;
  display: flex;
  gap: 8px;
}

.playbook-section__agent-team span {
  display: grid;
  place-items: center;
  width: 58px;
  height: 28px;
  color: #d6d0c5;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.05);
  font-size: 11px;
  font-weight: 900;
  opacity: 0;
  animation: agent-team-pop 4.8s cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.playbook-section__agent-team span:nth-child(2) {
  animation-delay: 180ms;
}

.playbook-section__agent-team span:nth-child(3) {
  animation-delay: 360ms;
}

.playbook-section__rank {
  width: min(720px, 82%);
}

.playbook-section__rank-row {
  position: relative;
  display: grid;
  grid-template-columns: 48px 150px 96px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  margin: 20px 0;
  color: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.playbook-section__rank-row span {
  color: #a7a39b;
}

.playbook-section__rank-row strong {
  font-size: 17px;
}

.playbook-section__rank-row em {
  display: inline-flex;
  justify-content: center;
  padding: 6px 9px;
  color: #111;
  border-radius: 999px;
  background: #fff;
  font-size: 12px;
  font-style: normal;
  font-weight: 900;
  animation: rank-cash-pop 3.2s ease-in-out infinite;
}

.playbook-section__rank-row--two em {
  animation-delay: 180ms;
}

.playbook-section__rank-row--three em {
  animation-delay: 360ms;
}

.playbook-section__rank-row i {
  display: block;
  height: 18px;
  border: 2px solid #fff;
  border-radius: 999px;
  transform-origin: left center;
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.9), transparent);
  scale: 0.42 1;
  animation: rank-bar-grow 3.2s ease-in-out infinite;
}

.playbook-section__rank-row--one i {
  animation-delay: 0ms;
}

.playbook-section__rank-row--two i {
  max-width: 74%;
  animation-delay: 180ms;
}

.playbook-section__rank-row--three i {
  max-width: 58%;
  animation-delay: 360ms;
}

.playbook-section__crown {
  display: inline-flex;
  margin: 20px 0 0 138px;
  padding: 8px 12px;
  color: #111;
  border-radius: 999px;
  background: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  animation: rank-reward 3.2s ease-in-out infinite;
}

.playbook-section__codebox {
  display: grid;
  gap: 10px;
  width: min(640px, 82%);
  color: #e8e1d4;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 14px;
}

.playbook-section__codebox span {
  position: relative;
  overflow: hidden;
  width: 0;
  max-width: min(var(--line-width), 100%);
  white-space: nowrap;
  animation: playbook-code-type 680ms steps(var(--line-steps), end) forwards;
  animation-delay: var(--line-delay);
}

.playbook-section__codebox span::after {
  content: "";
  display: inline-block;
  width: 7px;
  height: 1em;
  margin-left: 3px;
  vertical-align: -0.12em;
  background: #f2c36b;
  animation: code-caret 0.72s steps(1, end) infinite;
}

.playbook-section__cursor {
  position: absolute;
  width: 0;
  height: 0;
  border-top: 18px solid transparent;
  border-bottom: 18px solid transparent;
  border-left: 28px solid #fff;
  transform: translate(84px, 50px) rotate(-24deg);
  filter: drop-shadow(0 0 10px rgba(255, 255, 255, 0.6));
  animation: signin-cursor-click 2.8s ease-in-out infinite;
}

.playbook-section__channels {
  display: grid;
  align-content: start;
  gap: 10px;
  padding: 20px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.7);
}

:global(.dark) .playbook-section__channels {
  background: rgba(255, 255, 255, 0.06);
}

.playbook-section__channels-title {
  margin-bottom: 8px;
  color: #9a958c;
  text-align: center;
  text-transform: uppercase;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
}

.playbook-section__channels button {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  min-height: 42px;
  padding: 0 12px;
  color: #6b665f;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  font-size: 14px;
  font-weight: 800;
  text-align: left;
  transition:
    color 160ms ease,
    border-color 160ms ease,
    background 160ms ease,
    transform 160ms ease;
}

.playbook-section__channels button small {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
}

.playbook-section__channels button.is-active,
.playbook-section__channels button:hover {
  color: #111;
  border-color: #111;
  background: #fff;
  transform: translateX(-3px);
}

.playbook-section__dial {
  width: 58px;
  height: 58px;
  margin: 22px auto 0;
  border: 1px solid #111;
  border-radius: 50%;
}

.playbook-section__dial::after {
  content: "";
  display: block;
  width: 2px;
  height: 18px;
  margin: 12px auto;
  background: #111;
  animation: dial-turn 2.2s ease-in-out infinite;
  transform-origin: bottom;
}

.playbook-section__dial-label {
  color: #6c675f;
  text-align: center;
  font-size: 12px;
}

@keyframes dial-turn {
  50% {
    transform: rotate(90deg);
  }
}

@keyframes playbook-monitor-scan {
  0%,
  22% {
    transform: translateX(-120%);
  }

  58% {
    transform: translateX(120%);
  }

  100% {
    transform: translateX(120%);
  }
}

@keyframes playbook-scene-sweep {
  0%,
  28% {
    transform: translateX(-86%);
  }

  58% {
    transform: translateX(86%);
  }

  100% {
    transform: translateX(86%);
  }
}

@keyframes code-grid-pan {
  to {
    background-position:
      0 28px,
      28px 0;
  }
}

@keyframes signin-card-pop {
  0%,
  100% {
    transform: translateY(0);
  }

  42% {
    transform: translateY(-6px);
  }
}

@keyframes signin-cursor-click {
  0%,
  26% {
    transform: translate(118px, 78px) rotate(-24deg);
  }

  42%,
  58% {
    transform: translate(90px, 54px) rotate(-24deg) scale(0.88);
  }

  78%,
  100% {
    transform: translate(118px, 78px) rotate(-24deg);
  }
}

@keyframes signin-stamp {
  0%,
  38% {
    transform: rotate(-8deg) scale(0);
  }

  48%,
  82% {
    transform: rotate(-8deg) scale(1);
  }

  100% {
    transform: rotate(-8deg) scale(0.86);
  }
}

@keyframes signin-reward-float {
  50% {
    transform: translate(-50%, -8px) scale(1.03);
    box-shadow: 0 16px 34px rgba(255, 255, 255, 0.2);
  }
}

@keyframes signin-tap-ring {
  0%,
  30% {
    opacity: 0;
    transform: translate(92px, 54px) scale(0.4);
  }

  52% {
    opacity: 1;
    transform: translate(92px, 54px) scale(1);
  }

  82%,
  100% {
    opacity: 0;
    transform: translate(92px, 54px) scale(1.45);
  }
}

@keyframes signin-coin {
  0% {
    opacity: 0.28;
  }

  42%,
  72% {
    opacity: 1;
  }

  100% {
    opacity: 0.18;
    translate: 0 -84px;
  }
}

@keyframes agent-card-rise {
  0%,
  100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(-6px);
  }
}

@keyframes agent-flow-packet {
  0% {
    opacity: 0;
    transform: translate3d(-42px, -50%, 0) scaleX(0.45);
  }

  18%,
  70% {
    opacity: 0.78;
  }

  100% {
    opacity: 0;
    transform: translate3d(calc(100vw * 0.42), -50%, 0) scaleX(1);
  }
}

@keyframes agent-ledger-in {
  0%,
  46% {
    opacity: 0;
    transform: translateY(12px) scale(0.96);
  }

  62%,
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes agent-cash-travel {
  0%,
  32% {
    opacity: 0;
    transform: translate3d(-28px, 10px, 0) scale(0.86);
  }

  52%,
  78% {
    opacity: 1;
    transform: translate3d(0, 0, 0) scale(1);
  }

  100% {
    opacity: 0.36;
    transform: translate3d(30px, -14px, 0) scale(0.92);
  }
}

@keyframes agent-team-pop {
  0%,
  24% {
    opacity: 0;
    transform: translateY(10px);
  }

  42%,
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes agent-node-pulse {
  50% {
    box-shadow: 0 0 0 10px rgba(255, 255, 255, 0.08);
  }
}

@keyframes agent-flow {
  to {
    background-position: -180% 0;
  }
}

@keyframes rebate-float {
  50% {
    opacity: 0.72;
    transform: translateY(-10px);
  }
}

@keyframes agent-action-pulse {
  50% {
    border-color: rgba(255, 255, 255, 0.9);
    background: rgba(255, 255, 255, 0.16);
    transform: translateY(-4px);
  }
}

@keyframes agent-formula-pulse {
  50% {
    transform: translateY(-5px);
    box-shadow: 0 20px 34px rgba(255, 255, 255, 0.16);
  }
}

@keyframes rank-bar-grow {
  0%,
  18% {
    scale: 0.42 1;
  }

  58%,
  100% {
    scale: 1 1;
  }
}

@keyframes rank-reward {
  0%,
  54% {
    opacity: 0;
    transform: translateY(8px);
  }

  72%,
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes rank-cash-pop {
  0%,
  44% {
    transform: translateY(4px) scale(0.92);
  }

  62%,
  100% {
    transform: translateY(0) scale(1);
  }
}

@keyframes playbook-code-type {
  from {
    width: 0;
  }

  to {
    width: min(var(--line-width), 100%);
  }
}

@keyframes code-caret {
  50% {
    opacity: 0;
  }
}

@media (max-width: 980px) {
  .playbook-section {
    width: min(100% - 32px, 1180px);
    padding: 82px 0;
  }

  .playbook-section__copy h2 {
    font-size: 38px;
  }

  .playbook-section__machine {
    grid-template-columns: 1fr;
  }

  .playbook-section__channels {
    grid-template-columns: repeat(2, 1fr);
  }

  .playbook-section__channels-title,
  .playbook-section__dial,
  .playbook-section__dial-label {
    display: none;
  }
}

@media (max-width: 560px) {
  .playbook-section__screen {
    min-height: 460px;
  }

  .playbook-section__channels {
    grid-template-columns: 1fr;
  }

  .playbook-section__agent-flow {
    inset: 78px 18px 42px;
    grid-template-columns: 1fr;
    gap: 10px;
    align-content: center;
  }

  .playbook-section__agent-card {
    min-height: auto;
    padding: 11px 12px;
  }

  .playbook-section__agent-card--wallet strong {
    font-size: 32px;
  }

  .playbook-section__agent-wire,
  .playbook-section__agent-team,
  .playbook-section__agent-ledger,
  .playbook-section__agent-cash {
    display: none;
  }

  .playbook-section__rank {
    width: min(100% - 36px, 340px);
  }

  .playbook-section__rank-row {
    grid-template-columns: 34px minmax(0, 1fr);
    gap: 8px 10px;
    margin: 16px 0;
  }

  .playbook-section__rank-row em,
  .playbook-section__rank-row i {
    grid-column: 2;
  }

  .playbook-section__rank-row strong {
    font-size: 14px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .playbook-section__scene::before,
  .playbook-section__screen::after,
  .playbook-section__dial::after,
  .playbook-section__signin-card,
  .playbook-section__signin-card b,
  .playbook-section__signin-reward,
  .playbook-section__tap-ring,
  .playbook-section__cursor,
  .playbook-section__coin,
  .playbook-section__agent-card,
  .playbook-section__agent-wire i,
  .playbook-section__agent-ledger,
  .playbook-section__agent-cash,
  .playbook-section__agent-team span,
  .playbook-section__rank-row i,
  .playbook-section__rank-row em,
  .playbook-section__crown,
  .playbook-section__codebox span {
    animation: none;
  }

  .playbook-section__codebox span {
    width: auto;
    max-width: 100%;
  }
}
</style>
