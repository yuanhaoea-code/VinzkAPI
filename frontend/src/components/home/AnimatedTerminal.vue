<script setup lang="ts">
import { computed, onMounted, onUnmounted, shallowRef, watch } from 'vue'

interface TerminalLine {
  text: string
  tone?: 'normal' | 'muted' | 'ok' | 'warn'
}

const props = withDefaults(
  defineProps<{
    title: string
    lines: TerminalLine[]
    bodyHeight?: number
    typeMs?: number
    linePauseMs?: number
    cyclePauseMs?: number
  }>(),
  {
    bodyHeight: 318,
    typeMs: 22,
    linePauseMs: 160,
    cyclePauseMs: 1900
  }
)

const renderedLines = shallowRef<string[]>(props.lines.map((line, index) => (index === 0 ? line.text.slice(0, 1) : '')))
const activeLine = shallowRef(0)
const isMounted = shallowRef(false)

let timer: ReturnType<typeof setTimeout> | undefined

const terminalStyle = computed(() => ({
  '--terminal-body-height': `${props.bodyHeight}px`
}))

function clearTimer() {
  if (!timer) return
  clearTimeout(timer)
  timer = undefined
}

function schedule(callback: () => void, delay: number) {
  clearTimer()
  timer = setTimeout(callback, delay)
}

function updateLine(lineIndex: number, text: string) {
  const nextLines = [...renderedLines.value]
  nextLines[lineIndex] = text
  renderedLines.value = nextLines
}

function typeLine(lineIndex: number, charIndex: number) {
  const line = props.lines[lineIndex]

  if (!line || !isMounted.value) {
    return
  }

  activeLine.value = lineIndex
  updateLine(lineIndex, line.text.slice(0, Math.max(1, charIndex)))

  if (charIndex < line.text.length) {
    schedule(() => typeLine(lineIndex, charIndex + 1), props.typeMs)
    return
  }

  if (lineIndex < props.lines.length - 1) {
    schedule(() => typeLine(lineIndex + 1, 1), props.linePauseMs)
    return
  }

  schedule(startPlayback, props.cyclePauseMs)
}

function startPlayback() {
  clearTimer()

  if (!props.lines.length || !isMounted.value) {
    renderedLines.value = []
    return
  }

  renderedLines.value = props.lines.map((line, index) => (index === 0 ? line.text.slice(0, 1) : ''))
  activeLine.value = 0
  schedule(() => typeLine(0, 2), props.typeMs)
}

watch(
  () => props.lines.map((line) => line.text).join('\n'),
  () => {
    if (isMounted.value) {
      startPlayback()
    }
  }
)

onMounted(() => {
  isMounted.value = true
  startPlayback()
})

onUnmounted(() => {
  isMounted.value = false
  clearTimer()
})
</script>

<template>
  <div class="animated-terminal" :style="terminalStyle" aria-label="terminal animation">
    <div class="animated-terminal__bar">
      <span></span>
      <span></span>
      <span></span>
      <b>{{ props.title }}</b>
    </div>

    <pre class="animated-terminal__body"><code><span
      v-for="(line, index) in props.lines"
      :key="`${index}-${line.text}`"
      class="animated-terminal__line"
      :class="[
        `animated-terminal__line--${line.tone ?? 'normal'}`,
        { 'is-active': activeLine === index }
      ]"
    >{{ renderedLines[index] || '\u00a0' }}</span></code></pre>
  </div>
</template>

<style scoped>
.animated-terminal {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: #080808;
  box-shadow: 0 28px 64px rgba(0, 0, 0, 0.18);
}

.animated-terminal::before {
  content: "";
  position: absolute;
  inset: 42px 0 0;
  z-index: 1;
  background:
    linear-gradient(90deg, transparent, rgba(242, 195, 107, 0.12), transparent),
    repeating-linear-gradient(0deg, transparent 0 30px, rgba(255, 255, 255, 0.03) 31px 32px);
  opacity: 0.72;
  transform: translateX(-100%);
  animation: terminal-scan 4.8s ease-in-out infinite;
  pointer-events: none;
}

.animated-terminal__bar {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: #171717;
}

.animated-terminal__bar span {
  width: 11px;
  height: 11px;
  border-radius: 999px;
}

.animated-terminal__bar span:nth-child(1) {
  background: #ff5f57;
}

.animated-terminal__bar span:nth-child(2) {
  background: #ffbd2e;
}

.animated-terminal__bar span:nth-child(3) {
  background: #28c840;
}

.animated-terminal__bar b {
  margin-left: 8px;
  color: #777;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 700;
}

.animated-terminal__body {
  position: relative;
  z-index: 2;
  overflow: hidden;
  min-height: var(--terminal-body-height);
  margin: 0;
  padding: 28px;
  color: #e8e1d4;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 14px;
  line-height: 2;
  white-space: pre;
}

.animated-terminal__body::before {
  content: "VINZK_AI_STREAM";
  position: absolute;
  right: 28px;
  bottom: 20px;
  color: rgba(242, 195, 107, 0.14);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  letter-spacing: 0.16em;
  animation: terminal-watermark 2.6s ease-in-out infinite;
  pointer-events: none;
}

.animated-terminal__body::after {
  content: "";
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.2) 1px, transparent 1.5px);
  background-size: 9px 9px;
  opacity: 0.08;
  animation: terminal-noise 1.4s steps(2, end) infinite;
  pointer-events: none;
}

.animated-terminal__line {
  position: relative;
  display: block;
  min-height: 2em;
}

.animated-terminal__line.is-active::after {
  content: "";
  display: inline-block;
  width: 7px;
  height: 1.1em;
  margin-left: 3px;
  vertical-align: -0.16em;
  background: currentColor;
  animation: terminal-caret 720ms steps(1, end) infinite;
}

.animated-terminal__line--muted {
  color: #8a8174;
}

.animated-terminal__line--ok {
  color: #8ef0a5;
}

.animated-terminal__line--warn {
  color: #f2c36b;
}

@keyframes terminal-scan {
  0%,
  18% {
    transform: translateX(-110%);
  }

  52% {
    transform: translateX(110%);
  }

  100% {
    transform: translateX(110%);
  }
}

@keyframes terminal-noise {
  50% {
    opacity: 0.15;
    transform: translateY(1px);
  }
}

@keyframes terminal-watermark {
  50% {
    opacity: 0.36;
  }
}

@keyframes terminal-caret {
  50% {
    opacity: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .animated-terminal::before,
  .animated-terminal__body::before,
  .animated-terminal__body::after,
  .animated-terminal__line.is-active::after {
    animation: none;
  }
}
</style>
