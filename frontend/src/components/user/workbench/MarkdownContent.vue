<script setup lang="ts">
import { computed, onBeforeUnmount } from 'vue'
import DOMPurify from 'dompurify'
import { marked, Renderer } from 'marked'

const props = defineProps<{ content: string }>()

const renderer = new Renderer()
renderer.code = ({ text, lang }) => {
  const language = (lang || '').trim().split(/\s+/)[0]
  const languageClass = language ? ` class="language-${escapeHTML(language)}"` : ''
  return [
    '<div class="code-block">',
    '<div class="code-block__toolbar">',
    `<span>${escapeHTML(language || '代码')}</span>`,
    '<button type="button" data-code-copy aria-label="复制代码">复制</button>',
    '</div>',
    `<pre><code${languageClass}>${escapeHTML(text)}</code></pre>`,
    '</div>'
  ].join('')
}

let feedbackTimer: number | undefined
let feedbackButton: HTMLButtonElement | null = null

const html = computed(() => {
  const rendered = marked.parse(props.content || '', {
    async: false,
    breaks: true,
    gfm: true,
    renderer
  }) as string
  return DOMPurify.sanitize(rendered, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'iframe', 'form'],
    FORBID_ATTR: ['style']
  })
})

function escapeHTML(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

async function onContentClick(event: MouseEvent): Promise<void> {
  const target = event.target
  if (!(target instanceof Element)) return
  const button = target.closest<HTMLButtonElement>('[data-code-copy]')
  if (!button) return
  const code = button.closest('.code-block')?.querySelector('code')?.textContent
  if (!code) return
  await navigator.clipboard.writeText(code)
  resetFeedback()
  feedbackButton = button
  button.textContent = '已复制'
  feedbackTimer = window.setTimeout(resetFeedback, 1400)
}

function resetFeedback(): void {
  window.clearTimeout(feedbackTimer)
  if (feedbackButton?.isConnected) feedbackButton.textContent = '复制'
  feedbackButton = null
  feedbackTimer = undefined
}

onBeforeUnmount(resetFeedback)
</script>

<template>
  <div class="markdown-content" @click="onContentClick" v-html="html" />
</template>

<style scoped>
.markdown-content {
  min-width: 0;
  color: inherit;
  font-size: 14px;
  line-height: 1.72;
  overflow-wrap: anywhere;
}

.markdown-content :deep(p) {
  margin: 0 0 10px;
}

.markdown-content :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3) {
  margin: 18px 0 8px;
  color: inherit;
  font-weight: 700;
  letter-spacing: 0;
}

.markdown-content :deep(h1) { font-size: 19px; }
.markdown-content :deep(h2) { font-size: 17px; }
.markdown-content :deep(h3) { font-size: 15px; }

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 8px 0 12px;
  padding-left: 21px;
}

.markdown-content :deep(li) {
  margin: 4px 0;
}

.markdown-content :deep(.code-block) {
  max-width: 100%;
  margin: 12px 0;
  overflow-x: auto;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 6px;
  background: #1c1b19;
  color: #f7f3eb;
}

.markdown-content :deep(.code-block__toolbar) {
  display: flex;
  min-height: 32px;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0 8px 0 13px;
  color: #a9a49b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 10px;
}

.markdown-content :deep(.code-block__toolbar button) {
  min-width: 42px;
  min-height: 24px;
  border: 0;
  border-radius: 4px;
  padding: 0 7px;
  background: transparent;
  color: #d8d3ca;
  font-family: inherit;
  font-size: 10px;
}

.markdown-content :deep(.code-block__toolbar button:hover),
.markdown-content :deep(.code-block__toolbar button:focus-visible) {
  outline: 0;
  background: rgba(255, 255, 255, 0.09);
  color: #fff;
}

.markdown-content :deep(pre) {
  max-width: 100%;
  margin: 0;
  overflow-x: auto;
  padding: 12px 13px;
  font-size: 12px;
  line-height: 1.6;
}

.markdown-content :deep(code) {
  border-radius: 3px;
  padding: 1px 4px;
  background: rgba(23, 20, 17, 0.07);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.9em;
}

.markdown-content :deep(pre code) {
  padding: 0;
  background: transparent;
}

.markdown-content :deep(blockquote) {
  margin: 12px 0;
  border-left: 2px solid #7fa69e;
  padding-left: 12px;
  color: #6d655b;
}

.markdown-content :deep(table) {
  width: 100%;
  margin: 12px 0;
  border-collapse: collapse;
  font-size: 12px;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  border: 1px solid rgba(23, 20, 17, 0.1);
  padding: 7px 8px;
  text-align: left;
}

.markdown-content :deep(a) {
  color: #356f64;
  text-decoration: underline;
  text-underline-offset: 2px;
}

:global(.dark .markdown-content) :deep(code) {
  background: rgba(148, 163, 184, 0.12);
}

:global(.dark .markdown-content) :deep(blockquote) {
  color: #94a3b8;
}

:global(.dark .markdown-content) :deep(th),
:global(.dark .markdown-content) :deep(td) {
  border-color: rgba(148, 163, 184, 0.16);
}
</style>
