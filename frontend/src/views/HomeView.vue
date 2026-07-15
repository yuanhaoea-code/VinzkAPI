<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import AgentWorkspaceSection from '@/components/home/AgentWorkspaceSection.vue'
import GatewaySection from '@/components/home/GatewaySection.vue'
import HeroSection from '@/components/home/HeroSection.vue'
import HomeFooter from '@/components/home/HomeFooter.vue'
import HomeNav from '@/components/home/HomeNav.vue'
import ImageServiceSection from '@/components/home/ImageServiceSection.vue'
import OnboardingSection from '@/components/home/OnboardingSection.vue'
import PlaybookSection from '@/components/home/PlaybookSection.vue'
import PricingCTASection from '@/components/home/PricingCTASection.vue'
import PromiseSection from '@/components/home/PromiseSection.vue'
import StatsStrip from '@/components/home/StatsStrip.vue'
import WhySection from '@/components/home/WhySection.vue'

const authStore = useAuthStore()
const appStore = useAppStore()

const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = shallowRef(document.documentElement.classList.contains('dark'))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const actionPath = computed(() => (isAuthenticated.value ? dashboardPath.value : '/login'))
const userInitial = computed(() => {
  const email = authStore.user?.email
  return email ? email.charAt(0).toUpperCase() : ''
})

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- homeContent is an admin-only setting and intentionally supports custom HTML. -->
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="home-page">
    <HomeNav
      :site-logo="siteLogo"
      :is-dark="isDark"
      :is-authenticated="isAuthenticated"
      :dashboard-path="dashboardPath"
      :user-initial="userInitial"
      @toggle-theme="toggleTheme"
    />

    <main>
      <HeroSection :action-path="actionPath" :is-authenticated="isAuthenticated" />
      <PromiseSection />
      <StatsStrip />
      <GatewaySection />
      <ImageServiceSection />
      <PlaybookSection />
      <AgentWorkspaceSection />
      <WhySection />
      <OnboardingSection />
      <PricingCTASection :action-path="actionPath" :is-authenticated="isAuthenticated" />
      <HomeFooter />
    </main>
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  color: #111;
  background:
    linear-gradient(rgba(17, 17, 17, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(17, 17, 17, 0.025) 1px, transparent 1px),
    #faf9f6;
  background-size: 64px 64px;
}

:global(.dark) .home-page {
  color: #fffaf0;
  background:
    linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.04) 1px, transparent 1px),
    #090909;
}
</style>
