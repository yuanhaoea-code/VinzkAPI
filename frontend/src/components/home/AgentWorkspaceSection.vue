<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <section id="workspace" class="agent-workspace">
    <div class="agent-workspace__panel">
      <div class="agent-workspace__chat">
        <div class="agent-workspace__message agent-workspace__message--user">
          {{ t('home.landing.workspace.userMessage') }}
        </div>
        <div class="agent-workspace__message agent-workspace__message--ai">
          {{ t('home.landing.workspace.aiMessage') }}
        </div>
        <div class="agent-workspace__typing">
          <span></span>
          <span></span>
          <span></span>
        </div>
      </div>
      <div class="agent-workspace__files">
        <span>report.xlsx</span>
        <span>contract.pdf</span>
        <span>brief.docx</span>
      </div>
      <div class="agent-workspace__nodes" aria-hidden="true">
        <span>Hermes</span>
        <span>OpenClaw</span>
        <span>API</span>
      </div>
    </div>

    <div class="agent-workspace__copy">
      <p class="agent-workspace__eyebrow">{{ t('home.landing.workspace.eyebrow') }}</p>
      <h2>{{ t('home.landing.workspace.title') }}</h2>
      <p>{{ t('home.landing.workspace.lead') }}</p>
      <div class="agent-workspace__status">{{ t('home.landing.workspace.status') }}</div>
    </div>
  </section>
</template>

<style scoped>
.agent-workspace {
  display: grid;
  grid-template-columns: minmax(420px, 1.1fr) minmax(0, 0.9fr);
  gap: 78px;
  align-items: center;
  scroll-margin-top: 90px;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
  padding: 118px 0;
  border-top: 1px solid rgba(20, 20, 20, 0.14);
}

.agent-workspace__panel {
  display: grid;
  grid-template-columns: 1fr 150px;
  gap: 16px;
  min-height: 420px;
  padding: 18px;
  border: 1px solid rgba(17, 17, 17, 0.18);
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 24px 58px rgba(0, 0, 0, 0.08);
}

:global(.dark) .agent-workspace__panel {
  border-color: rgba(255, 255, 255, 0.16);
  background: #15130f;
}

.agent-workspace__chat {
  position: relative;
  display: grid;
  align-content: end;
  gap: 18px;
  padding: 24px;
  border-radius: 8px;
  background:
    linear-gradient(rgba(17, 17, 17, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(17, 17, 17, 0.04) 1px, transparent 1px),
    #faf9f6;
  background-size: 28px 28px;
}

:global(.dark) .agent-workspace__chat {
  background:
    linear-gradient(rgba(255, 255, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.06) 1px, transparent 1px),
    #080808;
}

.agent-workspace__message {
  max-width: 82%;
  padding: 14px 16px;
  border: 1px solid rgba(17, 17, 17, 0.12);
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.7;
}

.agent-workspace__message--user {
  justify-self: end;
  color: #fff;
  background: #111;
}

.agent-workspace__message--ai {
  justify-self: start;
  color: #27231e;
  background: #fff;
}

.agent-workspace__typing {
  display: inline-flex;
  gap: 6px;
  width: max-content;
  padding: 12px 14px;
  border: 1px solid rgba(17, 17, 17, 0.12);
  border-radius: 999px;
  background: #fff;
}

.agent-workspace__typing span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #111;
  animation: typing-dot 1.2s ease-in-out infinite;
}

.agent-workspace__typing span:nth-child(2) {
  animation-delay: 120ms;
}

.agent-workspace__typing span:nth-child(3) {
  animation-delay: 240ms;
}

.agent-workspace__files {
  display: grid;
  align-content: start;
  gap: 12px;
}

.agent-workspace__files span {
  padding: 14px 12px;
  border: 1px solid rgba(17, 17, 17, 0.14);
  border-radius: 8px;
  color: #4f4a43;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  background: #faf9f6;
}

.agent-workspace__nodes {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.agent-workspace__nodes span {
  padding: 8px 10px;
  border: 1px solid rgba(17, 17, 17, 0.14);
  border-radius: 999px;
  color: #111;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.agent-workspace__eyebrow {
  margin: 0 0 24px;
  color: #88827a;
  font-size: 13px;
  font-weight: 700;
}

.agent-workspace__copy h2 {
  margin: 0;
  color: #0a0a0a;
  font-family: SimSun, 'Songti SC', serif;
  font-size: 52px;
  font-weight: 900;
  line-height: 1.16;
}

:global(.dark) .agent-workspace__copy h2 {
  color: #fffaf0;
}

.agent-workspace__copy p {
  margin: 26px 0 0;
  color: #5d5850;
  font-size: 16px;
  line-height: 1.9;
}

:global(.dark) .agent-workspace__copy p {
  color: #c8c0b4;
}

.agent-workspace__status {
  display: inline-flex;
  margin-top: 28px;
  padding: 9px 12px;
  border: 1px solid rgba(17, 17, 17, 0.14);
  border-radius: 999px;
  color: #7d5a16;
  background: rgba(242, 195, 107, 0.18);
  font-size: 13px;
  font-weight: 800;
}

@keyframes typing-dot {
  50% {
    transform: translateY(-4px);
    opacity: 0.45;
  }
}

@media (max-width: 980px) {
  .agent-workspace {
    grid-template-columns: 1fr;
    width: min(100% - 32px, 1180px);
    padding: 82px 0;
  }

  .agent-workspace__copy {
    order: -1;
  }

  .agent-workspace__copy h2 {
    font-size: 38px;
  }
}

@media (max-width: 620px) {
  .agent-workspace__panel {
    grid-template-columns: 1fr;
    min-height: auto;
  }
}

@media (prefers-reduced-motion: reduce) {
  .agent-workspace__typing span {
    animation: none;
  }
}
</style>
