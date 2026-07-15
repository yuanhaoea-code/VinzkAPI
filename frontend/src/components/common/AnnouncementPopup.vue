<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="announcementStore.currentPopup"
        class="announcement-popup-overlay fixed inset-0 z-[120] flex items-start justify-center overflow-y-auto bg-gradient-to-br from-black/70 via-black/60 to-black/70 p-4 pt-[8vh] backdrop-blur-md"
      >
        <div
          class="announcement-popup w-full max-w-[680px] overflow-hidden rounded-3xl bg-white shadow-2xl ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
          @click.stop
        >
          <!-- Header with warm gradient -->
          <div class="announcement-popup__header relative overflow-hidden border-b border-amber-100/80 bg-gradient-to-br from-amber-50/80 via-orange-50/50 to-yellow-50/30 px-8 py-6 dark:border-dark-700/50 dark:from-amber-900/20 dark:via-orange-900/10 dark:to-yellow-900/5">
            <!-- Decorative background -->
            <div class="absolute right-0 top-0 h-full w-64 bg-gradient-to-l from-orange-100/30 to-transparent dark:from-orange-900/20"></div>
            <div class="absolute -right-8 -top-8 h-32 w-32 rounded-full bg-gradient-to-br from-amber-400/20 to-orange-500/20 blur-3xl"></div>
            <div class="absolute -left-4 -bottom-4 h-24 w-24 rounded-full bg-gradient-to-tr from-yellow-400/20 to-amber-500/20 blur-2xl"></div>

            <div class="relative z-10">
              <!-- Icon and badge -->
              <div class="mb-3 flex items-center gap-2">
                <div class="announcement-popup__icon flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-amber-500 to-orange-600 text-white shadow-lg shadow-amber-500/30">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                  </svg>
                </div>
                <span class="announcement-popup__badge inline-flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-amber-500 to-orange-600 px-2.5 py-1 text-xs font-medium text-white shadow-lg shadow-amber-500/30">
                  <span class="relative flex h-2 w-2">
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-white opacity-75"></span>
                    <span class="relative inline-flex h-2 w-2 rounded-full bg-white"></span>
                  </span>
                  {{ t('announcements.unread') }}
                </span>
              </div>

              <!-- Title -->
              <h2 class="mb-2 text-2xl font-bold leading-tight text-gray-900 dark:text-white">
                {{ announcementStore.currentPopup.title }}
              </h2>

              <!-- Time -->
              <div class="flex items-center gap-1.5 text-sm text-gray-600 dark:text-gray-400">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <time>{{ formatRelativeWithDateTime(announcementStore.currentPopup.created_at) }}</time>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div class="announcement-popup__body max-h-[50vh] overflow-y-auto bg-white px-8 py-8 dark:bg-dark-800">
            <div class="relative">
              <div class="announcement-popup__rail absolute left-0 top-0 bottom-0 w-1 rounded-full bg-gradient-to-b from-amber-500 via-orange-500 to-yellow-500"></div>
              <div class="pl-6">
                <div
                  class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                  v-html="renderedContent"
                ></div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="announcement-popup__footer border-t border-gray-100 bg-gray-50/50 px-8 py-5 dark:border-dark-700 dark:bg-dark-900/30">
            <div class="flex items-center justify-end">
              <button
                @click="handleDismiss"
                class="announcement-popup__primary rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 px-6 py-2.5 text-sm font-medium text-white shadow-lg shadow-amber-500/30 transition-all hover:shadow-xl hover:scale-105"
              >
                <span class="flex items-center gap-2">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  {{ t('announcements.markRead') }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = announcementStore.currentPopup?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  announcementStore.dismissPopup()
}

// Manage body overflow — only set, never unset (bell component handles restore)
watch(
  () => announcementStore.currentPopup,
  (popup) => {
    if (popup) {
      document.body.style.overflow = 'hidden'
    }
  }
)
</script>

<style scoped>
.announcement-popup-overlay {
  background: rgba(11, 18, 32, 0.62);
}

.announcement-popup {
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 22px;
  background: #fbf8f1;
  box-shadow: 0 28px 90px rgba(23, 20, 17, 0.2);
}

.announcement-popup__header {
  border-bottom-color: rgba(23, 20, 17, 0.08);
  background:
    radial-gradient(circle at 92% 0, rgba(127, 159, 152, 0.18), transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.76), rgba(248, 243, 233, 0.96));
}

.announcement-popup__header > .absolute {
  display: none;
}

.announcement-popup__icon,
.announcement-popup__badge,
.announcement-popup__primary {
  background: #171411;
  color: #fbf8f1;
  box-shadow: none;
}

.announcement-popup__badge,
.announcement-popup__primary {
  border-radius: 999px;
}

.announcement-popup__primary {
  border: 1px solid rgba(23, 20, 17, 0.14);
}

.announcement-popup__primary:hover {
  background: #2b2721;
  transform: none;
}

.announcement-popup__body {
  background: rgba(255, 255, 255, 0.62);
}

.announcement-popup__rail {
  background: #7f9f98;
}

.announcement-popup__footer {
  border-top-color: rgba(23, 20, 17, 0.08);
  background: rgba(255, 255, 255, 0.42);
}

:global(.dark) .announcement-popup {
  border-color: rgba(148, 163, 184, 0.16);
  background: #0f172a;
  box-shadow: 0 28px 90px rgba(0, 0, 0, 0.42);
}

:global(.dark) .announcement-popup__header {
  border-bottom-color: rgba(148, 163, 184, 0.14);
  background:
    radial-gradient(circle at 92% 0, rgba(45, 212, 191, 0.12), transparent 34%),
    linear-gradient(180deg, rgba(17, 24, 39, 0.96), rgba(15, 23, 42, 0.98));
}

:global(.dark) .announcement-popup__icon,
:global(.dark) .announcement-popup__badge,
:global(.dark) .announcement-popup__primary {
  border-color: rgba(45, 212, 191, 0.28);
  background: #14b8a6;
  color: #052e2b;
}

:global(.dark) .announcement-popup__primary:hover {
  background: #2dd4bf;
}

:global(.dark) .announcement-popup__body {
  background: rgba(15, 23, 42, 0.92);
}

:global(.dark) .announcement-popup__rail {
  background: #14b8a6;
}

:global(.dark) .announcement-popup__footer {
  border-top-color: rgba(148, 163, 184, 0.14);
  background: rgba(11, 18, 32, 0.74);
}

.popup-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.popup-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
}

.popup-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.popup-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #374151);
}
</style>
