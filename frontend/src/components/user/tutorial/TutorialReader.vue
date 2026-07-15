<script setup lang="ts">
import { computed, ref, shallowRef, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import { PUBLIC_API_BASE_URL } from '@/constants/site'
import { useTutorialCatalog } from '@/composables/useTutorialCatalog'
import { useTutorialOutline } from '@/composables/useTutorialOutline'
import { saveLastTutorialSlug } from '@/utils/tutorialProgress'
import TutorialArticleContent from './TutorialArticleContent.vue'
import TutorialDirectory from './TutorialDirectory.vue'

const route = useRoute()
const { catalog } = useTutorialCatalog()

const articleRoot = ref<HTMLElement | null>(null)
const mobileDirectoryOpen = shallowRef(false)

const currentSlug = computed(() => String(route.params.slug ?? ''))
const currentArticle = computed(() =>
  catalog.value.articles.find((article) => article.slug === currentSlug.value)
)

const { activeSectionId, scrollToSection } = useTutorialOutline(currentArticle, articleRoot)

watch(currentArticle, (article) => {
  if (article) saveLastTutorialSlug(article.slug)
}, { immediate: true })
</script>

<template>
  <div class="tutorial-reader">
    <div class="tutorial-reader__toolbar">
      <div>
        <button
          type="button"
          class="tutorial-reader__directory-toggle"
          aria-label="打开文档目录"
          @click="mobileDirectoryOpen = true"
        >
          <Icon name="menu" size="sm" />
        </button>
        <RouterLink to="/tutorials" class="tutorial-reader__home">
          <Icon name="book" size="sm" />
          <span>教程文档</span>
        </RouterLink>
        <span v-if="currentArticle" class="tutorial-reader__current">/ {{ currentArticle.shortTitle }}</span>
      </div>
      <RouterLink to="/tutorials" class="tutorial-reader__all">
        全部教程
        <Icon name="arrowRight" size="xs" />
      </RouterLink>
    </div>

    <div class="tutorial-reader__body">
      <TutorialDirectory
        class="tutorial-reader__directory"
        :class="{ 'tutorial-reader__directory--open': mobileDirectoryOpen }"
        :active-section-id="activeSectionId"
        :articles="catalog.articles"
        :current-slug="currentSlug"
        :groups="catalog.groups"
        @close="mobileDirectoryOpen = false"
        @section-click="scrollToSection"
      />

      <main ref="articleRoot" class="tutorial-reader__content">
        <TutorialArticleContent
          v-if="currentArticle"
          :article="currentArticle"
          :articles="catalog.articles"
        />
        <div v-else class="tutorial-reader__not-found">
          <Icon name="book" size="xl" />
          <h1>没有找到这篇教程</h1>
          <p>教程地址可能已变更，请从教程首页重新选择。</p>
          <RouterLink to="/tutorials">返回教程首页</RouterLink>
        </div>
      </main>

      <aside v-if="currentArticle" class="tutorial-reader__outline">
        <strong>本页内容</strong>
        <nav>
          <button
            v-for="section in currentArticle.sections"
            :key="section.id"
            type="button"
            :class="{ 'tutorial-reader__outline--active': activeSectionId === section.id }"
            @click="scrollToSection(section.id)"
          >
            {{ section.title }}
          </button>
        </nav>
        <div class="tutorial-reader__outline-footer">
          <span>API 地址</span>
          <code>{{ PUBLIC_API_BASE_URL }}</code>
        </div>
      </aside>
    </div>

    <button
      v-if="mobileDirectoryOpen"
      type="button"
      class="tutorial-reader__overlay"
      aria-label="关闭文档目录"
      @click="mobileDirectoryOpen = false"
    ></button>
  </div>
</template>

<style scoped>
.tutorial-reader {
  min-width: 0;
  overflow: clip;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.66);
}

.tutorial-reader__toolbar {
  position: sticky;
  top: 56px;
  z-index: 12;
  display: flex;
  min-height: 50px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 0 14px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.09);
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(14px);
}

.tutorial-reader__toolbar > div {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 7px;
}

.tutorial-reader__home,
.tutorial-reader__all {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #2c2824;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
}

.tutorial-reader__all {
  min-height: 32px;
  padding: 0 11px;
  border: 1px solid rgba(23, 20, 17, 0.09);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.66);
  font-size: 11px;
  font-weight: 600;
}

.tutorial-reader__current {
  overflow: hidden;
  color: #7c7267;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-reader__directory-toggle {
  display: none;
  width: 32px;
  height: 32px;
  place-items: center;
  border: 1px solid rgba(23, 20, 17, 0.09);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.66);
}

.tutorial-reader__body {
  display: grid;
  min-height: calc(100vh - 139px);
  grid-template-columns: 246px minmax(0, 1fr) 220px;
  align-items: start;
}

.tutorial-reader__directory {
  position: sticky;
  top: 106px;
  height: calc(100vh - 123px);
}

.tutorial-reader__content {
  min-width: 0;
}

.tutorial-reader__outline {
  position: sticky;
  top: 106px;
  display: flex;
  height: calc(100vh - 123px);
  flex-direction: column;
  overflow-y: auto;
  border-left: 1px solid rgba(23, 20, 17, 0.09);
  padding: 23px 15px 18px;
  background: rgba(250, 248, 244, 0.48);
}

.tutorial-reader__outline > strong {
  margin-bottom: 14px;
  color: #6f665d;
  font-size: 13px;
  letter-spacing: 0;
}

.tutorial-reader__outline nav {
  border-left: 1px solid #ddd5cc;
}

.tutorial-reader__outline button {
  display: block;
  width: 100%;
  padding: 9px 0 9px 12px;
  border: 0;
  border-left: 2px solid transparent;
  color: #817870;
  background: transparent;
  font-size: 12px;
  line-height: 1.55;
  text-align: left;
}

.tutorial-reader__outline button:hover,
.tutorial-reader__outline--active {
  border-left-color: #0f9f90 !important;
  color: #087b70 !important;
  font-weight: 600;
}

.tutorial-reader__outline-footer {
  margin-top: auto;
  padding-top: 15px;
  border-top: 1px solid rgba(23, 20, 17, 0.09);
}

.tutorial-reader__outline-footer span,
.tutorial-reader__outline-footer code {
  display: block;
}

.tutorial-reader__outline-footer span {
  margin-bottom: 6px;
  color: #8d837a;
  font-size: 11px;
}

.tutorial-reader__outline-footer code {
  overflow: hidden;
  color: #35675f;
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
}

.tutorial-reader__not-found {
  display: grid;
  min-height: 520px;
  place-items: center;
  align-content: center;
  gap: 10px;
  padding: 30px;
  color: #7c7267;
  text-align: center;
}

.tutorial-reader__not-found h1,
.tutorial-reader__not-found p {
  margin: 0;
}

.tutorial-reader__not-found h1 {
  color: #231f1b;
  font-family: "Songti SC", "STSong", Georgia, serif;
  font-size: 26px;
}

.tutorial-reader__not-found a {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  margin-top: 5px;
  padding: 0 13px;
  border-radius: 999px;
  color: #fff;
  background: #171411;
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
}

.tutorial-reader__overlay {
  position: fixed;
  inset: 0;
  z-index: 39;
  border: 0;
  background: rgba(15, 23, 42, 0.42);
}

:global(.dark .tutorial-reader) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(11, 18, 32, 0.7);
}

:global(.dark .tutorial-reader__toolbar) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.88);
}

:global(.dark .tutorial-reader__home),
:global(.dark .tutorial-reader__all) {
  color: #f8fafc;
}

:global(.dark .tutorial-reader__current),
:global(.dark .tutorial-reader__outline > strong),
:global(.dark .tutorial-reader__outline button),
:global(.dark .tutorial-reader__outline-footer span) {
  color: #94a3b8;
}

:global(.dark .tutorial-reader__all),
:global(.dark .tutorial-reader__directory-toggle) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(30, 41, 59, 0.72);
}

:global(.dark .tutorial-reader__outline) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.54);
}

:global(.dark .tutorial-reader__outline nav),
:global(.dark .tutorial-reader__outline-footer) {
  border-color: rgba(148, 163, 184, 0.16);
}

:global(.dark .tutorial-reader__outline-footer code) {
  color: #5eead4;
}

@media (max-width: 1180px) {
  .tutorial-reader__body {
    grid-template-columns: 238px minmax(0, 1fr);
  }

  .tutorial-reader__outline {
    display: none;
  }
}

@media (max-width: 900px) {
  .tutorial-reader__body {
    grid-template-columns: minmax(0, 1fr);
  }

  .tutorial-reader__directory-toggle {
    display: grid;
  }

  .tutorial-reader__directory {
    position: fixed;
    top: 56px;
    bottom: 0;
    left: 0;
    z-index: 40;
    width: min(84vw, 300px);
    height: auto;
    background: #faf8f4;
    box-shadow: 18px 0 36px rgba(23, 20, 17, 0.16);
    transform: translateX(-105%);
    transition: transform 0.2s ease;
  }

  :global(.dark .tutorial-reader__directory) {
    background: #0f172a;
    box-shadow: 18px 0 36px rgba(0, 0, 0, 0.34);
  }

  .tutorial-reader__directory--open {
    transform: translateX(0);
  }
}

@media (max-width: 640px) {
  .tutorial-reader {
    border-radius: 12px;
  }

  .tutorial-reader__toolbar {
    top: 56px;
    min-height: 46px;
    padding: 0 10px;
  }

  .tutorial-reader__current,
  .tutorial-reader__all {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .tutorial-reader__directory {
    transition: none;
  }
}
</style>
