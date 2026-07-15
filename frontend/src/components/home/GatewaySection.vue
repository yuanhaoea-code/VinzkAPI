<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AnimatedTerminal from './AnimatedTerminal.vue'

interface TerminalLine {
  text: string
  tone?: 'normal' | 'muted' | 'ok' | 'warn'
}

const { t } = useI18n()

const endpoints = ['/v1/chat/completions', '/v1/messages', '/v1/images/generations'] as const
const tools = ['Codex CLI', 'Claude Code', 'Cursor', 'Cline', 'Gemini CLI'] as const
const terminalLines = computed<TerminalLine[]>(() => [
  { text: '$ vinzk route --watch --model gpt-5.5', tone: 'warn' },
  { text: '$ curl -X POST https://api.vinzk.cn/v1/chat/completions \\' },
  { text: '  -H "Authorization: Bearer sk-******" \\' },
  { text: '  -d \'{"model":"gpt-5.5","messages":[...]}\' ' },
  { text: '# routing: Claude / GPT / Gemini / image2 / seedance', tone: 'muted' },
  { text: '200 OK  { "content": "一念既起，万象可成" }', tone: 'ok' }
])
</script>

<template>
  <section id="gateway" class="gateway-section">
    <div class="gateway-section__copy">
      <p class="gateway-section__eyebrow">{{ t('home.landing.gateway.eyebrow') }}</p>
      <h2>{{ t('home.landing.gateway.title') }}</h2>
      <p class="gateway-section__lead">{{ t('home.landing.gateway.lead') }}</p>
      <div class="gateway-section__endpoints">
        <div v-for="endpoint in endpoints" :key="endpoint" class="gateway-section__endpoint">
          <span>POST</span>
          <code>{{ endpoint }}</code>
        </div>
      </div>
      <div class="gateway-section__tools">
        <span v-for="tool in tools" :key="tool">{{ tool }}</span>
      </div>
    </div>

    <AnimatedTerminal
      class="gateway-section__terminal"
      title="bash - vinzk ai gateway"
      :lines="terminalLines"
      :body-height="318"
    />
  </section>
</template>

<style scoped>
.gateway-section {
  display: grid;
  grid-template-columns: minmax(0, 0.82fr) minmax(420px, 1.18fr);
  gap: 74px;
  align-items: center;
  scroll-margin-top: 90px;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0;
}

.gateway-section__eyebrow {
  margin: 0 0 24px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.gateway-section__copy h2 {
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 52px;
  font-weight: 900;
  line-height: 1.16;
}

:global(.dark) .gateway-section__copy h2 {
  color: #fffaf0;
}

.gateway-section__lead {
  margin: 26px 0 0;
  color: #5d5850;
  font-size: 16px;
  line-height: 1.9;
}

:global(.dark) .gateway-section__lead {
  color: #c8c0b4;
}

.gateway-section__endpoints {
  display: grid;
  gap: 0;
  margin-top: 42px;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.gateway-section__endpoint {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px solid rgba(20, 20, 20, 0.14);
}

.gateway-section__endpoint span {
  padding: 3px 8px;
  color: #fff;
  border-radius: 5px;
  background: #0b0b0b;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 11px;
  font-weight: 800;
}

.gateway-section__endpoint code {
  color: #4d4942;
  font-size: 13px;
}

.gateway-section__tools {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 24px;
}

.gateway-section__tools span {
  padding: 8px 10px;
  color: #4f4a43;
  border: 1px solid rgba(20, 20, 20, 0.12);
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

@media (max-width: 980px) {
  .gateway-section {
    grid-template-columns: 1fr;
    width: min(100% - 32px, 1180px);
    padding: 82px 0;
  }

  .gateway-section__copy h2 {
    font-size: 38px;
  }
}

</style>
