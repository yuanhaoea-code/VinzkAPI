<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  src: string
  alt: string
  caption: string
}>()

const previewOpen = shallowRef(false)
const previewTriggerRef = ref<HTMLButtonElement | null>(null)
const closeButtonRef = ref<HTMLButtonElement | null>(null)
let previousActiveElement: HTMLElement | null = null
let addedScrollLock = false

function openPreview() {
  previewOpen.value = true
}

function closePreview() {
  previewOpen.value = false
}

function handleEscape(event: KeyboardEvent) {
  if (previewOpen.value && event.key === 'Escape') {
    closePreview()
  }
}

watch(previewOpen, async (isOpen) => {
  if (isOpen) {
    previousActiveElement = previewTriggerRef.value
    if (!document.body.classList.contains('modal-open')) {
      document.body.classList.add('modal-open')
      addedScrollLock = true
    }
    await nextTick()
    closeButtonRef.value?.focus()
    return
  }

  if (addedScrollLock) {
    document.body.classList.remove('modal-open')
    addedScrollLock = false
  }
  previousActiveElement?.focus()
  previousActiveElement = null
})

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  if (addedScrollLock) document.body.classList.remove('modal-open')
})
</script>

<template>
  <figure class="tutorial-shot">
    <button
      ref="previewTriggerRef"
      type="button"
      class="tutorial-shot__trigger"
      :aria-label="`查看大图：${alt}`"
      @click="openPreview"
    >
      <img :src="src" :alt="alt" loading="lazy" decoding="async" />
    </button>
    <figcaption>{{ caption }}</figcaption>
  </figure>

  <Teleport to="body">
    <div
      v-if="previewOpen"
      class="tutorial-lightbox"
      role="dialog"
      aria-modal="true"
      :aria-label="`图片预览：${alt}`"
      @click.self="closePreview"
    >
      <button
        ref="closeButtonRef"
        type="button"
        class="tutorial-lightbox__close"
        aria-label="关闭图片预览"
        title="关闭图片预览"
        @click="closePreview"
      >
        <Icon name="x" size="md" :stroke-width="2" />
      </button>

      <div class="tutorial-lightbox__stage" @click.stop>
        <img :src="src" :alt="alt" decoding="async" />
        <p v-if="caption">{{ caption }}</p>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.tutorial-shot {
  margin: 0;
}

.tutorial-shot__trigger {
  display: block;
  width: 100%;
  overflow: hidden;
  padding: 0;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.76);
  cursor: zoom-in;
  text-align: initial;
}

.tutorial-shot__trigger:focus-visible {
  outline: 2px solid rgba(15, 159, 144, 0.72);
  outline-offset: 3px;
}

.tutorial-shot__trigger img {
  display: block;
  width: 100%;
  max-height: 560px;
  object-fit: contain;
  background: #f8fafc;
}

.tutorial-shot figcaption {
  margin-top: 8px;
  color: #7c7267;
  font-size: 11px;
  line-height: 1.55;
}

.tutorial-lightbox {
  position: fixed;
  inset: 0;
  z-index: 140;
  display: grid;
  overflow: auto;
  place-items: center;
  padding: 72px 24px 28px;
  background: rgba(10, 12, 14, 0.9);
  backdrop-filter: blur(10px);
}

.tutorial-lightbox__close {
  position: fixed;
  top: max(18px, env(safe-area-inset-top));
  right: max(18px, env(safe-area-inset-right));
  z-index: 2;
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 50%;
  color: #ffffff;
  background: rgba(26, 29, 32, 0.82);
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.28);
  transition: border-color 0.16s ease, background 0.16s ease;
}

.tutorial-lightbox__close:hover {
  border-color: rgba(255, 255, 255, 0.58);
  background: rgba(55, 60, 64, 0.92);
}

.tutorial-lightbox__close:focus-visible {
  outline: 2px solid #5eead4;
  outline-offset: 3px;
}

.tutorial-lightbox__stage {
  display: grid;
  width: min(94vw, 1600px);
  justify-items: center;
  gap: 12px;
}

.tutorial-lightbox__stage img {
  display: block;
  max-width: 100%;
  max-height: calc(100dvh - 126px);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 8px;
  object-fit: contain;
  background: #f8fafc;
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.36);
}

.tutorial-lightbox__stage p {
  margin: 0;
  color: rgba(255, 255, 255, 0.76);
  font-size: 13px;
  line-height: 1.6;
  text-align: center;
}

:global(.dark .tutorial-shot__trigger) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.72);
}

:global(.dark .tutorial-shot__trigger img) {
  background: #0f172a;
}

:global(.dark .tutorial-shot figcaption) {
  color: #94a3b8;
}

@media (max-width: 640px) {
  .tutorial-lightbox {
    padding: 64px 12px 20px;
  }

  .tutorial-lightbox__close {
    top: max(12px, env(safe-area-inset-top));
    right: max(12px, env(safe-area-inset-right));
  }

  .tutorial-lightbox__stage {
    width: 100%;
  }

  .tutorial-lightbox__stage img {
    max-height: calc(100dvh - 104px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .tutorial-lightbox__close {
    transition: none;
  }
}
</style>
