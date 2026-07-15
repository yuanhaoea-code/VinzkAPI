<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { scrollToHomeSection } from './useHomeAnchorScroll'

const props = defineProps<{
  siteLogo: string
  isDark: boolean
  isAuthenticated: boolean
  dashboardPath: string
  userInitial: string
}>()

const emit = defineEmits<{
  toggleTheme: []
}>()

const { t } = useI18n()

const navItems = [
  { href: '#gateway', labelKey: 'home.landing.nav.gateway' },
  { href: '#creative', labelKey: 'home.landing.nav.creative' },
  { href: '#workspace', labelKey: 'home.landing.nav.workspace' },
  { href: '#pricing', labelKey: 'home.landing.nav.pricing' }
] as const

function handleAnchorClick(hash: string) {
  scrollToHomeSection(hash)
}
</script>

<template>
  <header class="home-nav">
    <nav class="home-nav__inner" :aria-label="t('home.landing.nav.label')">
      <a class="home-nav__brand" href="#top" @click.prevent="handleAnchorClick('#top')">
        <span class="home-nav__mark">
          <img v-if="props.siteLogo" :src="props.siteLogo" alt="" class="home-nav__logo" />
          <span v-else class="home-nav__glyph">维</span>
        </span>
        <span class="home-nav__name">{{ t('home.landing.brand') }}</span>
      </a>

      <div class="home-nav__links">
        <a
          v-for="item in navItems"
          :key="item.href"
          :href="item.href"
          class="home-nav__link"
          @click.prevent="handleAnchorClick(item.href)"
        >
          {{ t(item.labelKey) }}
        </a>
      </div>

      <div class="home-nav__actions">
        <LocaleSwitcher />
        <button
          class="home-nav__icon-button"
          type="button"
          :aria-label="props.isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="emit('toggleTheme')"
        >
          <Icon v-if="props.isDark" name="sun" size="sm" />
          <Icon v-else name="moon" size="sm" />
        </button>
        <router-link class="home-nav__console" :to="props.isAuthenticated ? props.dashboardPath : '/login'">
          <span v-if="props.isAuthenticated" class="home-nav__avatar">{{ props.userInitial }}</span>
          <span>{{ props.isAuthenticated ? t('home.dashboard') : t('home.login') }}</span>
          <Icon name="externalLink" size="xs" />
        </router-link>
      </div>
    </nav>
  </header>
</template>

<style scoped>
.home-nav {
  position: sticky;
  top: 0;
  z-index: 30;
  border-bottom: 1px solid rgba(20, 20, 20, 0.08);
  background: rgba(250, 249, 246, 0.88);
  backdrop-filter: blur(18px);
}

:global(.dark) .home-nav {
  border-bottom-color: rgba(255, 255, 255, 0.12);
  background: rgba(10, 10, 10, 0.84);
}

.home-nav__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  width: min(1180px, calc(100% - 48px));
  min-height: 72px;
  margin: 0 auto;
}

.home-nav__brand,
.home-nav__actions,
.home-nav__links {
  display: flex;
  align-items: center;
}

.home-nav__brand {
  flex: 0 0 auto;
  gap: 12px;
  color: #111;
  text-decoration: none;
}

:global(.dark) .home-nav__brand {
  color: #f6f2e9;
}

.home-nav__mark {
  display: inline-grid;
  overflow: hidden;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(17, 17, 17, 0.1);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.64);
}

:global(.dark) .home-nav__mark {
  border-color: rgba(255, 255, 255, 0.13);
  background: rgba(255, 255, 255, 0.06);
}

.home-nav__logo {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.home-nav__glyph {
  font-family: SimSun, 'Songti SC', serif;
  font-size: 18px;
  font-weight: 700;
}

.home-nav__name {
  font-size: 16px;
  font-weight: 700;
}

.home-nav__links {
  gap: 28px;
}

.home-nav__link {
  color: #696761;
  font-size: 14px;
  text-decoration: none;
  transition: color 180ms ease;
}

.home-nav__link:hover {
  color: #111;
}

:global(.dark) .home-nav__link {
  color: #b8b3a8;
}

:global(.dark) .home-nav__link:hover {
  color: #fffaf0;
}

.home-nav__actions {
  flex: 0 0 auto;
  gap: 10px;
}

.home-nav__icon-button {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  color: #111;
  border: 1px solid rgba(20, 20, 20, 0.12);
  border-radius: 999px;
  background: transparent;
  transition:
    border-color 180ms ease,
    transform 180ms ease;
}

.home-nav__icon-button:hover {
  border-color: rgba(20, 20, 20, 0.42);
  transform: translateY(-1px);
}

:global(.dark) .home-nav__icon-button {
  color: #fffaf0;
  border-color: rgba(255, 255, 255, 0.18);
}

.home-nav__console {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 14px;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
  border-radius: 999px;
  background: #090909;
  transition:
    transform 180ms ease,
    background 180ms ease;
}

.home-nav__console:hover {
  background: #26231e;
  transform: translateY(-1px);
}

.home-nav__avatar {
  display: grid;
  place-items: center;
  width: 22px;
  height: 22px;
  color: #111;
  border-radius: 999px;
  background: #f2c36b;
  font-size: 12px;
}

@media (max-width: 860px) {
  .home-nav__inner {
    width: min(100% - 32px, 1180px);
  }

  .home-nav__links {
    display: none;
  }
}

@media (max-width: 520px) {
  .home-nav__inner {
    gap: 12px;
    min-height: 64px;
  }

  .home-nav__name {
    display: none;
  }

  .home-nav__console {
    min-height: 34px;
    padding: 0 10px;
    font-size: 12px;
  }
}
</style>
