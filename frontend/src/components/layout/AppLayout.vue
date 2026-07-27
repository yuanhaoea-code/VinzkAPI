<template>
  <div
    class="app-layout min-h-screen bg-gray-50 dark:bg-dark-950"
    :class="{ 'app-layout--paper-console': isPaperConsoleRoute }"
  >
    <!-- Background Decoration -->
    <div v-if="!isPaperConsoleRoute" class="pointer-events-none fixed inset-0 bg-mesh-gradient"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="app-layout__content relative min-h-screen transition-all duration-300"
      :class="[
        sidebarCollapsed
          ? 'lg:ml-[72px]'
          : isPaperConsoleRoute
            ? 'lg:ml-[224px]'
            : 'lg:ml-64'
      ]"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main
        class="app-layout__main"
        :class="isPaperConsoleRoute ? 'app-layout__main--paper px-4 py-3 md:px-5 md:py-4 lg:px-6 lg:py-4' : 'p-4 md:p-6 lg:p-8'"
      >
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')
const userConsoleRouteNames = new Set([
  'Dashboard',
  'Keys',
  'Usage',
  'AvailableChannels',
  'ChannelStatus',
  'Subscriptions',
  'Redeem',
  'Recharge',
  'ModelMarket',
  'Workbench',
  'ImageGeneration',
  'Affiliate',
  'Profile',
  'TutorialHome',
  'TutorialArticle',
])
const isPaperConsoleRoute = computed(() => {
  const path = route.path
  const routeName = typeof route.name === 'string' ? route.name : ''
  return path.startsWith('/admin') || userConsoleRouteNames.has(routeName) || (path !== '/home' && !path.startsWith('/custom/'))
})

watch(
  isPaperConsoleRoute,
  (enabled) => {
    if (typeof document !== 'undefined') {
      document.body.classList.toggle('paper-console-active', enabled)
    }
  },
  { immediate: true }
)

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

onUnmounted(() => {
  if (typeof document !== 'undefined') {
    document.body.classList.remove('paper-console-active')
  }
})

defineExpose({ replayTour })
</script>

<style scoped>
.app-layout__content,
.app-layout__main {
  min-width: 0;
}

.app-layout--paper-console {
  --console-font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC",
    "Microsoft YaHei", "Noto Sans CJK SC", Arial, sans-serif;
  font-family: var(--console-font-sans);
  color: #171411;
  background:
    radial-gradient(circle at 12% 0, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0) 28%),
    linear-gradient(180deg, #fbf8f1 0%, #f3ede2 100%);
}

.app-layout__main--paper {
  min-height: calc(100vh - 56px);
  background: transparent;
}

:global(.dark) .app-layout--paper-console {
  color: #e5e7eb;
  background:
    radial-gradient(circle at 14% 0, rgba(20, 184, 166, 0.1), rgba(20, 184, 166, 0) 30%),
    radial-gradient(circle at 86% 8%, rgba(56, 189, 248, 0.08), rgba(56, 189, 248, 0) 32%),
    linear-gradient(180deg, #0b1220 0%, #08111f 100%);
}

:global(.app-layout--paper-console .btn) {
  min-height: 34px;
  border-radius: 999px;
  padding: 0 13px;
  font-size: 13px;
  font-weight: 600;
  box-shadow: none;
}

:global(.app-layout--paper-console .btn-primary) {
  border: 1px solid #171411;
  background: #171411;
  color: #fbf8f3;
}

:global(.app-layout--paper-console .btn-primary:hover) {
  background: #2b2721;
}

:global(.app-layout--paper-console .btn-secondary),
:global(.app-layout--paper-console .btn-ghost) {
  border: 1px solid rgba(23, 20, 17, 0.09);
  background: rgba(255, 255, 255, 0.72);
  color: #171411;
}

:global(.app-layout--paper-console .btn-secondary:hover),
:global(.app-layout--paper-console .btn-ghost:hover) {
  border-color: rgba(23, 20, 17, 0.14);
  background: rgba(255, 255, 255, 0.9);
}

:global(.app-layout--paper-console .input),
:global(.app-layout--paper-console .select-trigger) {
  min-height: 34px;
  border-color: rgba(23, 20, 17, 0.09);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  color: #171411;
  font-size: 13px;
  box-shadow: none;
}

:global(.app-layout--paper-console .input:focus),
:global(.app-layout--paper-console .select-trigger-open),
:global(.app-layout--paper-console .select-trigger:focus) {
  border-color: rgba(127, 159, 152, 0.46);
  box-shadow: 0 0 0 3px rgba(127, 159, 152, 0.13);
}

:global(.app-layout--paper-console .input-label) {
  margin-bottom: 6px;
  color: #7c7267;
  font-size: 13px;
  font-weight: 500;
}

:global(.app-layout--paper-console .card) {
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.66);
  box-shadow: none;
}

:global(.app-layout--paper-console .bg-primary-50),
:global(.app-layout--paper-console .bg-blue-50),
:global(.app-layout--paper-console .bg-green-50),
:global(.app-layout--paper-console .bg-purple-50),
:global(.app-layout--paper-console .bg-amber-50) {
  background-color: rgba(255, 255, 255, 0.68) !important;
}

:global(.app-layout--paper-console .table-wrapper th) {
  color: #7c7267;
  font-size: 13px;
  font-weight: 500;
}

:global(.app-layout--paper-console .table-wrapper td) {
  color: #171411;
  font-size: 14px;
}

:global(.app-layout--paper-console .table-wrapper),
:global(.app-layout--paper-console .table-scroll-container) {
  scrollbar-color: rgba(124, 114, 103, 0.34) rgba(23, 20, 17, 0.04) !important;
}

:global(.dark .app-layout--paper-console .btn-secondary),
:global(.dark .app-layout--paper-console .btn-ghost),
:global(.dark .app-layout--paper-console .input),
:global(.dark .app-layout--paper-console .select-trigger),
:global(.dark .app-layout--paper-console .card) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(17, 24, 39, 0.82);
  color: #e5e7eb;
}

:global(.dark .app-layout--paper-console .btn-primary) {
  border-color: rgba(45, 212, 191, 0.36);
  background: #14b8a6;
  color: #06211d;
}

:global(.dark .app-layout--paper-console .btn-primary:hover) {
  background: #2dd4bf;
}

:global(.dark .app-layout--paper-console .input::placeholder) {
  color: #64748b;
}

:global(.dark .app-layout--paper-console .input:focus),
:global(.dark .app-layout--paper-console .select-trigger-open),
:global(.dark .app-layout--paper-console .select-trigger:focus) {
  border-color: rgba(45, 212, 191, 0.44);
  box-shadow: 0 0 0 3px rgba(20, 184, 166, 0.14);
}

:global(.dark .app-layout--paper-console .input-label),
:global(.dark .app-layout--paper-console .page-description),
:global(.dark .app-layout--paper-console .text-gray-500),
:global(.dark .app-layout--paper-console .text-dark-400) {
  color: #94a3b8 !important;
}

:global(.dark .app-layout--paper-console .page-title),
:global(.dark .app-layout--paper-console .text-gray-900),
:global(.dark .app-layout--paper-console .dark\:text-white) {
  color: #f8fafc !important;
}

:global(.dark .app-layout--paper-console .bg-white),
:global(.dark .app-layout--paper-console .dark\:bg-dark-900),
:global(.dark .app-layout--paper-console .dark\:bg-dark-800),
:global(.dark .app-layout--paper-console .table-container),
:global(.dark .app-layout--paper-console .table-wrapper),
:global(.dark .app-layout--paper-console .table-scroll-container) {
  border-color: rgba(148, 163, 184, 0.16) !important;
  background-color: rgba(17, 24, 39, 0.82) !important;
}

:global(.dark .app-layout--paper-console .table th),
:global(.dark .app-layout--paper-console .table-wrapper th),
:global(.dark .app-layout--paper-console thead) {
  border-color: rgba(148, 163, 184, 0.14) !important;
  background-color: rgba(15, 23, 42, 0.88) !important;
  color: #94a3b8 !important;
}

:global(.dark .app-layout--paper-console .table-wrapper td) {
  color: #f8fafc;
}

:global(.dark .app-layout--paper-console .table td),
:global(.dark .app-layout--paper-console tbody tr),
:global(.dark .app-layout--paper-console .divide-y > :not([hidden]) ~ :not([hidden])) {
  border-color: rgba(148, 163, 184, 0.12) !important;
}

:global(.dark .app-layout--paper-console tbody tr:hover),
:global(.dark .app-layout--paper-console .hover\:bg-gray-50:hover),
:global(.dark .app-layout--paper-console .dark\:hover\:bg-dark-800:hover) {
  background-color: rgba(20, 184, 166, 0.08) !important;
}

:global(.dark .app-layout--paper-console .badge-gray),
:global(.dark .app-layout--paper-console .bg-gray-100),
:global(.dark .app-layout--paper-console .bg-gray-50) {
  background-color: rgba(30, 41, 59, 0.72) !important;
  color: #cbd5e1 !important;
}
</style>
