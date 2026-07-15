<script setup lang="ts">
import { computed, onMounted, onUnmounted, shallowRef, useTemplateRef } from 'vue'
import { useI18n } from 'vue-i18n'
import coastalSunset from '@/assets/home/coastal-sunset.png'

const { t } = useI18n()

const clipItems = ['Prompt', 'Lighting', 'Photoreal'] as const
const outputSizes = ['1K', '2K', '4K'] as const
const sectionRef = useTemplateRef<HTMLElement>('section')
const demoKey = shallowRef(0)
const isDemoAnimated = shallowRef(false)
const renderElapsed = shallowRef(0)
let imageObserver: IntersectionObserver | undefined
let wasVisible = false
let renderFrame: number | undefined
let renderStartedAt = 0
const promptTypeMs = 1450
const imageLoadMs = 2050
const readyHoldMs = 1500
const renderCycleMs = promptTypeMs + imageLoadMs + readyHoldMs
const replayStartMs = promptTypeMs + imageLoadMs + readyHoldMs

const photoStyle: Record<string, string> = {
  '--coastal-photo': `url(${coastalSunset})`
}

function clampRatio(value: number) {
  return Math.min(1, Math.max(0, value))
}

const promptTypeProgress = computed(() => {
  if (renderElapsed.value >= replayStartMs) return 0
  return clampRatio(renderElapsed.value / promptTypeMs)
})
const imageLoadProgress = computed(() =>
  renderElapsed.value >= replayStartMs
    ? 0
    : clampRatio((renderElapsed.value - promptTypeMs) / imageLoadMs)
)

const renderProgress = computed(() => {
  const elapsed = renderElapsed.value
  if (elapsed < promptTypeMs) return 0
  if (elapsed < promptTypeMs + imageLoadMs) return Math.round(imageLoadProgress.value * 100)
  if (elapsed < replayStartMs) return 100
  return 0
})

const renderState = computed(() => {
  const elapsed = renderElapsed.value
  if (elapsed < promptTypeMs) return 'PROMPTING'
  if (elapsed < promptTypeMs + imageLoadMs) return 'GENERATING'
  if (elapsed < replayStartMs) return 'IMAGE READY'
  return 'PROMPTING'
})

const isPromptTyping = computed(() => renderElapsed.value < promptTypeMs || renderElapsed.value >= replayStartMs)
const isImageLoading = computed(() =>
  renderElapsed.value >= promptTypeMs && renderElapsed.value < promptTypeMs + imageLoadMs
)
const isImageReady = computed(() =>
  renderElapsed.value >= promptTypeMs + imageLoadMs &&
  renderElapsed.value < replayStartMs
)
const progressStyle = computed(() => ({
  transform: `scaleX(${renderProgress.value / 100})`
}))
const promptTextStyle = computed(() => ({
  width: `${Math.round(promptTypeProgress.value * 34 * 100) / 100}em`,
  clipPath: `inset(0 ${Math.round((1 - promptTypeProgress.value) * 100)}% 0 0)`
}))
const imageStyle = computed<Record<string, string>>(() => {
  const load = imageLoadProgress.value
  const isReady = renderElapsed.value >= promptTypeMs + imageLoadMs && renderElapsed.value < replayStartMs

  if (isReady) {
    return {
      ...photoStyle,
      opacity: '1',
      filter: 'none',
      transform: 'scale(1)'
    }
  }

  const opacity = renderElapsed.value < promptTypeMs ? 0 : 0.12 + load * 0.88
  const blur = Math.round((1 - load) * 22)
  const gray = Math.round((1 - load) * 78) / 100
  const saturation = Math.round((0.58 + load * 0.48) * 100) / 100
  const brightness = Math.round((0.78 + load * 0.22) * 100) / 100
  const scale = Math.round((1.055 - load * 0.055) * 1000) / 1000

  return {
    ...photoStyle,
    opacity: String(opacity),
    filter: `grayscale(${gray}) blur(${blur}px) saturate(${saturation}) brightness(${brightness})`,
    transform: `scale(${scale})`
  }
})

function updateRenderFrame() {
  renderElapsed.value = (performance.now() - renderStartedAt) % renderCycleMs
  renderFrame = window.requestAnimationFrame(updateRenderFrame)
}

function startRenderLoop() {
  stopRenderLoop()
  renderStartedAt = performance.now()
  renderElapsed.value = 0
  renderFrame = window.requestAnimationFrame(updateRenderFrame)
}

function stopRenderLoop() {
  if (renderFrame !== undefined) {
    window.cancelAnimationFrame(renderFrame)
    renderFrame = undefined
  }
}

onMounted(() => {
  if (!('IntersectionObserver' in window) || !sectionRef.value) {
    isDemoAnimated.value = true
    startRenderLoop()
    return
  }

  imageObserver = new IntersectionObserver(
    ([entry]) => {
      if (!entry) return

      if (entry.isIntersecting) {
        if (!wasVisible) {
          demoKey.value += 1
          startRenderLoop()
        }
        wasVisible = true
        isDemoAnimated.value = true
        return
      }

      wasVisible = false
      isDemoAnimated.value = false
      stopRenderLoop()
      renderElapsed.value = 0
    },
    {
      threshold: 0.32,
      rootMargin: '0px 0px -12% 0px'
    }
  )

  imageObserver.observe(sectionRef.value)
})

onUnmounted(() => {
  stopRenderLoop()
  imageObserver?.disconnect()
})
</script>

<template>
  <section id="creative" ref="section" class="image-service">
    <div class="image-service__copy">
      <p class="image-service__eyebrow">{{ t('home.landing.creative.eyebrow') }}</p>
      <h2>{{ t('home.landing.creative.title') }}</h2>
      <p>{{ t('home.landing.creative.lead') }}</p>

      <div class="image-service__model">
        <span class="image-service__dash"></span>
        <strong>image2</strong>
        <b>IMAGE</b>
        <small>OpenAI-compatible</small>
      </div>

      <div class="image-service__sizes" aria-label="Supported image sizes">
        <span>{{ t('home.landing.creative.sizeLabel') }}</span>
        <b v-for="size in outputSizes" :key="size">{{ size }}</b>
      </div>

      <div class="image-service__routes">
        <span>POST <code>/v1/images/generations</code></span>
        <span>POST <code>/v1/videos/generations</code></span>
      </div>
    </div>

    <div
      :key="demoKey"
      class="image-service__demo"
      :class="{ 'is-animated': isDemoAnimated }"
      aria-label="image generation animation"
    >
      <div class="image-service__topbar" aria-hidden="true">
        <span></span>
        <span></span>
        <span></span>
        <b>image2 render lab</b>
      </div>

      <div class="image-service__prompt">
        <span>PROMPT</span>
        <p class="image-service__prompt-text" :style="promptTextStyle">
          {{ t('home.landing.creative.prompt') }}
        </p>
        <i class="image-service__caret"></i>
      </div>

      <div class="image-service__queue" aria-hidden="true">
        <span v-for="(clip, index) in clipItems" :key="clip">
          <b></b>CLIP {{ String(index + 1).padStart(2, '0') }} / {{ clip }}
        </span>
      </div>

      <div
        class="image-service__canvas"
        :class="{
          'is-typing': isPromptTyping,
          'is-loading': isImageLoading,
          'is-ready': isImageReady
        }"
      >
        <div class="image-service__stage">
          <span>{{ t('home.landing.creative.generating') }}</span>
          <strong>SKYLINE PIPELINE</strong>
        </div>

        <div class="image-service__latent" aria-hidden="true"></div>
        <div class="image-service__tiles" aria-hidden="true">
          <span v-for="index in 64" :key="index"></span>
        </div>
        <div class="image-service__packets" aria-hidden="true">
          <i v-for="index in 10" :key="index"></i>
        </div>
        <div
          class="image-service__photo"
          :class="{ 'is-ready': isImageReady }"
          :style="imageStyle"
          aria-hidden="true"
        ></div>
        <div class="image-service__fog" aria-hidden="true"></div>
        <div class="image-service__scan" aria-hidden="true"></div>
        <div class="image-service__ready" :class="{ 'is-visible': isImageReady }" aria-hidden="true">
          {{ renderState }}
        </div>
      </div>

      <div class="image-service__progress" aria-hidden="true">
        <span :style="progressStyle"></span>
      </div>

      <div class="image-service__meta">
        <span>{{ renderState }}</span>
        <b>1024×1024 · ~5s</b>
        <strong :class="{ 'is-visible': isImageReady }">{{ renderProgress }}%</strong>
      </div>

      <div class="image-service__video">
        <span>seedance</span>
        <i></i>
        <i></i>
        <i></i>
        <p>{{ t('home.landing.creative.video') }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.image-service {
  display: grid;
  grid-template-columns: minmax(0, 0.82fr) minmax(420px, 1.18fr);
  gap: 76px;
  align-items: center;
  scroll-margin-top: 90px;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

:global(.dark) .image-service {
  border-top-color: rgba(255, 255, 255, 0.14);
}

.image-service__eyebrow {
  margin: 0 0 22px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.image-service__copy h2 {
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 52px;
  font-weight: 900;
  line-height: 1.16;
}

:global(.dark) .image-service__copy h2 {
  color: #fffaf0;
}

.image-service__copy p {
  margin: 26px 0 0;
  color: #5d5850;
  font-size: 16px;
  line-height: 1.9;
}

:global(.dark) .image-service__copy p {
  color: #c8c0b4;
}

.image-service__model {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 74px;
  color: #111;
}

.image-service__dash {
  width: 36px;
  height: 2px;
  background: #111;
}

.image-service__model strong {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 24px;
}

.image-service__model b {
  padding: 4px 8px;
  color: #fff;
  border-radius: 5px;
  background: #38bdf8;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
}

.image-service__model small {
  color: #777169;
  font-size: 13px;
}

.image-service__sizes {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-top: 22px;
}

.image-service__sizes span {
  color: #5d5850;
  font-size: 13px;
  font-weight: 800;
}

.image-service__sizes b {
  display: inline-grid;
  place-items: center;
  min-width: 42px;
  height: 30px;
  color: #0b1822;
  border: 1px solid rgba(56, 189, 248, 0.38);
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(56, 189, 248, 0.22), rgba(186, 230, 253, 0.62));
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 900;
}

.image-service__routes {
  display: grid;
  gap: 0;
  margin-top: 28px;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.image-service__routes span {
  padding: 13px 0;
  border-bottom: 1px solid rgba(20, 20, 20, 0.14);
  color: #111;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
}

.image-service__routes code {
  margin-left: 12px;
  color: #55514a;
}

.image-service__demo {
  --creative-cycle: 5s;
  position: relative;
  overflow: hidden;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.11);
  border-radius: 8px;
  background:
    radial-gradient(circle at 18% 0%, rgba(56, 189, 248, 0.16), transparent 32%),
    radial-gradient(circle at 90% 100%, rgba(192, 132, 252, 0.14), transparent 34%),
    #07080d;
  box-shadow: 0 28px 68px rgba(0, 0, 0, 0.22);
}

.image-service__demo:not(.is-animated)::before,
.image-service__demo:not(.is-animated) *,
.image-service__demo:not(.is-animated) *::before,
.image-service__demo:not(.is-animated) *::after {
  animation-play-state: paused !important;
}

.image-service__demo::before {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(115deg, transparent 0 36%, rgba(255, 255, 255, 0.07) 46%, transparent 58%);
  transform: translateX(-120%);
  animation: creative-panel-sweep var(--creative-cycle) ease-in-out infinite;
  pointer-events: none;
}

.image-service__topbar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 28px;
  color: #747989;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
}

.image-service__topbar span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #3b3f4b;
}

.image-service__topbar b {
  margin-left: 6px;
  font-weight: 800;
}

.image-service__prompt {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 16px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  background: #11131b;
}

.image-service__prompt span {
  flex: 0 0 auto;
  padding: 4px 8px;
  color: #bae6fd;
  border-radius: 5px;
  background: rgba(56, 189, 248, 0.16);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 900;
}

.image-service__prompt-text {
  overflow: hidden;
  width: 0;
  max-width: calc(100% - 96px);
  margin: 0;
  color: #f4ecdf;
  font-size: 13px;
  white-space: nowrap;
  transition:
    width 80ms linear,
    clip-path 80ms linear;
}

.image-service__caret {
  flex: 0 0 auto;
  width: 7px;
  height: 1.2em;
  background: #bae6fd;
  animation: creative-caret-blink 720ms steps(1, end) infinite;
}

.image-service__queue {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  color: #8d8794;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 800;
}

.image-service__queue span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 5px 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.04);
  opacity: 0.46;
  animation: creative-clip-state var(--creative-cycle) ease-in-out infinite;
}

.image-service__queue span:nth-child(2) {
  animation-delay: 220ms;
}

.image-service__queue span:nth-child(3) {
  animation-delay: 440ms;
}

.image-service__queue b {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #38bdf8;
  box-shadow: 0 0 12px rgba(56, 189, 248, 0.78);
}

.image-service__canvas {
  position: relative;
  overflow: hidden;
  height: 350px;
  margin-top: 18px;
  border-radius: 8px;
  background:
    radial-gradient(circle at 22% 18%, rgba(56, 189, 248, 0.12), transparent 28%),
    radial-gradient(circle at 76% 28%, rgba(192, 132, 252, 0.13), transparent 30%),
    linear-gradient(135deg, #050812, #030407 56%, #0b0f17);
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.04),
    inset 0 -80px 120px rgba(0, 0, 0, 0.36);
}

.image-service__canvas::before {
  content: "";
  position: absolute;
  z-index: 8;
  inset: 0;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: inherit;
  pointer-events: none;
}

.image-service__canvas::after {
  content: "";
  position: absolute;
  z-index: 9;
  inset: 0;
  background:
    linear-gradient(90deg, transparent 0 42%, rgba(125, 211, 252, 0.22) 48%, transparent 56%),
    repeating-linear-gradient(90deg, transparent 0 62px, rgba(255, 255, 255, 0.07) 63px 64px),
    repeating-linear-gradient(0deg, transparent 0 42px, rgba(255, 255, 255, 0.05) 43px 44px);
  opacity: 0.78;
  mix-blend-mode: screen;
  transform: translateX(-72%);
  animation: creative-render-curtain 2.2s cubic-bezier(0.19, 1, 0.22, 1) infinite;
  pointer-events: none;
}

.image-service__canvas.is-typing::after {
  opacity: 0.16;
  transform: translateX(0);
  animation: none;
}

.image-service__canvas.is-ready::after {
  opacity: 0;
  animation: none;
}

.image-service__stage {
  position: absolute;
  z-index: 10;
  top: 16px;
  left: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 7px;
  background: rgba(7, 8, 13, 0.74);
  backdrop-filter: blur(10px);
  transition:
    opacity 220ms ease,
    transform 220ms ease;
}

.image-service__canvas.is-typing .image-service__stage {
  opacity: 0;
  transform: translateY(-6px);
}

.image-service__stage span,
.image-service__stage strong {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 900;
}

.image-service__stage span {
  color: #d8d1c4;
}

.image-service__stage strong {
  color: #7dd3fc;
}

.image-service__latent,
.image-service__tiles,
.image-service__packets,
.image-service__photo,
.image-service__fog,
.image-service__scan {
  position: absolute;
  inset: 0;
}

.image-service__latent {
  z-index: 1;
  background:
    radial-gradient(circle at 18% 26%, rgba(125, 211, 252, 0.72), transparent 20%),
    radial-gradient(circle at 70% 28%, rgba(147, 197, 253, 0.44), transparent 26%),
    radial-gradient(circle at 58% 78%, rgba(244, 114, 182, 0.3), transparent 28%),
    linear-gradient(125deg, rgba(3, 7, 18, 0), rgba(56, 189, 248, 0.26), rgba(192, 132, 252, 0.24), rgba(3, 7, 18, 0));
  background-size: 100% 100%, 100% 100%, 100% 100%, 220% 220%;
  opacity: 0;
  filter: blur(14px) saturate(1.2);
  animation:
    creative-latent-reveal var(--creative-cycle) cubic-bezier(0.19, 1, 0.22, 1) infinite,
    creative-holo 6s linear infinite;
}

.image-service__canvas.is-typing .image-service__latent {
  background:
    radial-gradient(circle at 12% 16%, rgba(255, 255, 255, 0.34), transparent 24%),
    linear-gradient(135deg, #d8a4ff 0%, #8bb8ff 46%, #35c8ff 100%);
  background-size: 100% 100%, 160% 160%;
  opacity: 0.96;
  filter: saturate(1.12);
  transform: none;
  animation: creative-holo 3.8s linear infinite;
}

.image-service__canvas.is-typing .image-service__tiles,
.image-service__canvas.is-typing .image-service__packets,
.image-service__canvas.is-typing .image-service__fog,
.image-service__canvas.is-typing .image-service__scan {
  opacity: 0;
}

.image-service__canvas.is-ready .image-service__latent,
.image-service__canvas.is-ready .image-service__tiles,
.image-service__canvas.is-ready .image-service__packets,
.image-service__canvas.is-ready .image-service__fog,
.image-service__canvas.is-ready .image-service__scan {
  opacity: 0;
  animation: none;
}

.image-service__tiles {
  z-index: 2;
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  grid-template-rows: repeat(8, 1fr);
  gap: 2px;
  padding: 10px;
  opacity: 0;
  pointer-events: none;
  animation: creative-tiles-wrap var(--creative-cycle) ease-in-out infinite;
}

.image-service__tiles span {
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 2px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.18), transparent 62%),
    rgba(125, 211, 252, 0.12);
  transform: scale(0.72) translateY(8px);
  opacity: 0;
  animation: creative-tile-build var(--creative-cycle) cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.image-service__tiles span:nth-child(3n + 1) {
  animation-delay: 90ms;
}

.image-service__tiles span:nth-child(3n + 2) {
  animation-delay: 180ms;
}

.image-service__tiles span:nth-child(4n) {
  animation-delay: 270ms;
}

.image-service__packets {
  z-index: 4;
  pointer-events: none;
}

.image-service__packets i {
  position: absolute;
  width: 52px;
  height: 9px;
  border-radius: 999px;
  background: linear-gradient(90deg, transparent, rgba(186, 230, 253, 0.92), transparent);
  opacity: 0;
  transform: translate3d(-70px, 0, 0);
  animation: creative-packet-run 2.1s ease-in-out infinite;
}

.image-service__packets i:nth-child(1) {
  top: 18%;
  left: 9%;
}

.image-service__packets i:nth-child(2) {
  top: 32%;
  left: 4%;
  animation-delay: 120ms;
}

.image-service__packets i:nth-child(3) {
  top: 48%;
  left: 14%;
  animation-delay: 240ms;
}

.image-service__packets i:nth-child(4) {
  top: 65%;
  left: 7%;
  animation-delay: 360ms;
}

.image-service__packets i:nth-child(5) {
  top: 78%;
  left: 20%;
  animation-delay: 480ms;
}

.image-service__packets i:nth-child(6) {
  top: 24%;
  right: 8%;
  animation-delay: 600ms;
}

.image-service__packets i:nth-child(7) {
  top: 40%;
  right: 14%;
  animation-delay: 720ms;
}

.image-service__packets i:nth-child(8) {
  top: 57%;
  right: 5%;
  animation-delay: 840ms;
}

.image-service__packets i:nth-child(9) {
  top: 72%;
  right: 12%;
  animation-delay: 960ms;
}

.image-service__packets i:nth-child(10) {
  top: 84%;
  right: 25%;
  animation-delay: 1080ms;
}

.image-service__photo {
  z-index: 5;
  opacity: 0;
  transform: scale(1.055);
  filter: grayscale(0.78) blur(20px) saturate(0.58) brightness(0.78);
  background: var(--coastal-photo);
  background-position: center;
  background-size: cover;
  transition:
    opacity 80ms linear,
    filter 80ms linear,
    transform 80ms linear;
}

.image-service__canvas.is-typing .image-service__photo {
  opacity: 0 !important;
  transition: none;
}

.image-service__photo.is-revealing {
  opacity: 0.36;
  transform: scale(1.032);
  filter: grayscale(0.68) blur(15px) saturate(0.7) brightness(0.84);
}

.image-service__photo.is-ready {
  opacity: 1;
  transform: scale(1);
  filter: grayscale(0) blur(0) saturate(1.06) brightness(1);
}

.image-service__photo::before,
.image-service__photo::after {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.image-service__photo::before {
  background:
    linear-gradient(180deg, rgba(229, 231, 235, 0.58), transparent 46%, rgba(229, 231, 235, 0.34)),
    radial-gradient(circle at 50% 42%, rgba(255, 255, 255, 0.54), transparent 34%);
  opacity: 0.9;
  transition: opacity 120ms ease;
}

.image-service__photo.is-ready::before {
  opacity: 0;
}

.image-service__photo::after {
  background: linear-gradient(100deg, transparent 0 34%, rgba(255, 255, 255, 0.58) 47%, transparent 60%);
  opacity: 0;
  transform: translate3d(-120%, 0, 0) skewX(-8deg);
}

.image-service__photo.is-ready::after {
  opacity: 0;
  animation: none;
}

.image-service__fog {
  z-index: 6;
  background:
    radial-gradient(circle at 32% 46%, rgba(226, 232, 240, 0.72), transparent 26%),
    radial-gradient(circle at 64% 52%, rgba(226, 232, 240, 0.5), transparent 32%),
    linear-gradient(110deg, rgba(226, 232, 240, 0), rgba(226, 232, 240, 0.32), rgba(226, 232, 240, 0));
  opacity: 0;
  filter: blur(22px);
  animation: creative-fog-clear var(--creative-cycle) ease-in-out infinite;
}

.image-service__scan {
  z-index: 7;
  background:
    linear-gradient(100deg, transparent 0 36%, rgba(186, 230, 253, 0.58) 47%, transparent 58%),
    repeating-linear-gradient(0deg, transparent 0 13px, rgba(255, 255, 255, 0.08) 14px 15px);
  opacity: 0;
  transform: translateX(-130%) skewX(-8deg);
  animation: creative-scan var(--creative-cycle) cubic-bezier(0.19, 1, 0.22, 1) infinite;
}

.image-service__ready {
  position: absolute;
  z-index: 10;
  right: 16px;
  bottom: 16px;
  padding: 8px 10px;
  color: #06131c;
  border-radius: 999px;
  background: linear-gradient(135deg, #e0f2fe, #7dd3fc);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 900;
  opacity: 0;
  transform: translateY(8px);
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.image-service__ready.is-visible {
  opacity: 1;
  transform: translateY(0);
}

.image-service__progress {
  height: 7px;
  margin-top: 18px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
}

.image-service__progress span {
  display: block;
  width: 100%;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #0ea5e9 0%, #38bdf8 34%, #7dd3fc 68%, #d8b4fe 100%);
  box-shadow: 0 0 24px rgba(56, 189, 248, 0.46);
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 90ms linear;
}

.image-service__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 13px;
  color: #9f9a91;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.image-service__meta span::before {
  content: "";
  display: inline-block;
  width: 7px;
  height: 7px;
  margin-right: 8px;
  border-radius: 50%;
  background: #38bdf8;
  box-shadow: 0 0 12px rgba(56, 189, 248, 0.9);
  animation: creative-dot 1.4s ease-in-out infinite;
}

.image-service__meta strong {
  color: #7dd3fc;
  min-width: 36px;
  opacity: 0.76;
  transition: opacity 180ms ease;
}

.image-service__meta strong.is-visible {
  opacity: 1;
  text-shadow: 0 0 12px rgba(125, 211, 252, 0.72);
}

.image-service__video {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
}

.image-service__video span {
  color: #f2c36b;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 900;
}

.image-service__video i {
  width: 34px;
  height: 22px;
  border-radius: 5px;
  background: linear-gradient(135deg, #0ea5e9, #d8b4fe);
  animation: video-frame 1.5s ease-in-out infinite;
}

.image-service__video i:nth-child(3) {
  animation-delay: 120ms;
}

.image-service__video i:nth-child(4) {
  animation-delay: 240ms;
}

.image-service__video p {
  margin: 0;
  color: #d8d1c4;
  font-size: 12px;
}

@keyframes creative-prompt-type {
  0%,
  7% {
    width: 0;
  }

  28%,
  100% {
    width: min(34em, 100%);
  }
}

@keyframes creative-prompt-clip {
  0%,
  7% {
    clip-path: inset(0 100% 0 0);
  }

  28%,
  100% {
    clip-path: inset(0 0 0 0);
  }
}

@keyframes creative-caret-blink {
  50% {
    opacity: 0;
  }
}

@keyframes creative-clip-state {
  0%,
  26% {
    opacity: 0.38;
    transform: translateY(0);
  }

  36%,
  76% {
    opacity: 1;
    border-color: rgba(56, 189, 248, 0.45);
    background: rgba(56, 189, 248, 0.12);
    transform: translateY(-2px);
  }

  94%,
  100% {
    opacity: 0.46;
    transform: translateY(0);
  }
}

@keyframes creative-panel-sweep {
  0%,
  24% {
    transform: translateX(-120%);
  }

  52% {
    transform: translateX(120%);
  }

  100% {
    transform: translateX(120%);
  }
}

@keyframes creative-render-curtain {
  0% {
    opacity: 0.18;
    transform: translateX(-72%);
  }

  42% {
    opacity: 0.82;
  }

  100% {
    opacity: 0.18;
    transform: translateX(72%);
  }
}

@keyframes creative-latent-reveal {
  0%,
  24% {
    opacity: 0;
    transform: scale(1.08);
  }

  34%,
  66% {
    opacity: 0.88;
    transform: scale(1);
  }

  70%,
  100% {
    opacity: 0;
    transform: scale(0.98);
  }
}

@keyframes creative-holo {
  to {
    background-position: 0 0, 0 0, 0 0, 180% 0;
  }
}

@keyframes creative-tiles-wrap {
  0%,
  24% {
    opacity: 0;
  }

  35%,
  64% {
    opacity: 0.86;
  }

  72%,
  100% {
    opacity: 0;
  }
}

@keyframes creative-tile-build {
  0%,
  24% {
    opacity: 0;
    transform: scale(0.72) translateY(8px);
  }

  38%,
  62% {
    opacity: 0.82;
    transform: scale(1);
  }

  74%,
  100% {
    opacity: 0;
    transform: scale(1.02) translateY(-3px);
  }
}

@keyframes creative-packet-run {
  0% {
    opacity: 0;
    transform: translateX(-70px) scaleX(0.42);
  }

  24%,
  62% {
    opacity: 0.82;
  }

  100% {
    opacity: 0;
    transform: translateX(112px) scaleX(1.18);
  }
}

@keyframes creative-photo-reveal {
  0%,
  68% {
    opacity: 0;
    transform: scale(1.055);
    filter: grayscale(0.78) blur(20px) saturate(0.58) brightness(0.78);
  }

  74% {
    opacity: 0.32;
    transform: scale(1.036);
    filter: grayscale(0.7) blur(16px) saturate(0.68) brightness(0.82);
  }

  90%,
  96% {
    opacity: 1;
    transform: scale(1);
    filter: grayscale(0) blur(0) saturate(1.06) brightness(1);
  }

  100% {
    opacity: 0;
    transform: scale(1);
    filter: grayscale(0) blur(0) saturate(1.06) brightness(1);
  }
}

@keyframes creative-photo-mist {
  0%,
  72% {
    opacity: 0.88;
  }

  90%,
  100% {
    opacity: 0.08;
  }
}

@keyframes creative-photo-glaze {
  0%,
  72% {
    opacity: 0;
    transform: translate3d(-120%, 0, 0) skewX(-8deg);
  }

  82% {
    opacity: 0.68;
  }

  96%,
  100% {
    opacity: 0;
    transform: translate3d(120%, 0, 0) skewX(-8deg);
  }
}

@keyframes creative-fog-clear {
  0%,
  68% {
    opacity: 0;
    transform: scale(1.08);
  }

  74% {
    opacity: 0.84;
  }

  92%,
  100% {
    opacity: 0;
    transform: scale(1);
  }
}

@keyframes creative-scan {
  24% {
    opacity: 0;
    transform: translate3d(-130%, 0, 0) skewX(-8deg);
  }

  44% {
    opacity: 0.52;
  }

  70%,
  100% {
    opacity: 0;
    transform: translate3d(130%, 0, 0) skewX(-8deg);
  }
}

@keyframes creative-ready {
  0%,
  74% {
    opacity: 0;
    transform: translateY(8px);
  }

  86%,
  96% {
    opacity: 1;
    transform: translateY(0);
  }

  100% {
    opacity: 0;
    transform: translateY(8px);
  }
}

@keyframes creative-progress {
  0%,
  26% {
    transform: scaleX(0);
  }

  68%,
  100% {
    transform: scaleX(1);
  }
}

@keyframes creative-ready-text {
  0%,
  74% {
    opacity: 0;
  }

  84%,
  96% {
    opacity: 1;
  }

  100% {
    opacity: 0;
  }
}

@keyframes creative-dot {
  50% {
    opacity: 0.3;
    transform: scale(0.7);
  }
}

@keyframes video-frame {
  50% {
    transform: translateY(-2px);
    filter: brightness(1.18);
  }
}

@media (max-width: 980px) {
  .image-service {
    grid-template-columns: 1fr;
    width: min(100% - 32px, 1180px);
    padding: 82px 0;
  }

  .image-service__copy h2 {
    font-size: 38px;
  }
}

@media (max-width: 560px) {
  .image-service__canvas {
    height: 260px;
  }

  .image-service__prompt {
    align-items: flex-start;
    flex-direction: column;
  }

  .image-service__prompt-text {
    max-width: 100%;
    line-height: 1.7;
    white-space: normal;
  }

  .image-service__caret {
    display: none;
  }

  .image-service__stage {
    right: 12px;
    left: 12px;
    justify-content: space-between;
  }

  .image-service__video {
    align-items: flex-start;
    flex-wrap: wrap;
  }
}

@media (prefers-reduced-motion: reduce) {
  .image-service__demo::before,
  .image-service__prompt-text,
  .image-service__caret,
  .image-service__queue span,
  .image-service__latent,
  .image-service__tiles,
  .image-service__tiles span,
  .image-service__packets i,
  .image-service__photo,
  .image-service__photo::before,
  .image-service__photo::after,
  .image-service__fog,
  .image-service__scan,
  .image-service__ready,
  .image-service__progress span,
  .image-service__meta span::before,
  .image-service__meta strong,
  .image-service__video i {
    animation: none;
  }

  .image-service__prompt-text {
    width: auto;
  }

  .image-service__progress span {
    transform: scaleX(1);
  }

  .image-service__photo {
    opacity: 1;
    transform: none;
    filter: none;
  }

  .image-service__ready,
  .image-service__meta strong {
    opacity: 1;
  }
}
</style>
