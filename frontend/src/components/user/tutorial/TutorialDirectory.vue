<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import type { TutorialArticle, TutorialGroup } from './types'

const props = defineProps<{
  activeSectionId: string
  articles: TutorialArticle[]
  currentSlug: string
  groups: TutorialGroup[]
}>()

const emit = defineEmits<{
  close: []
  sectionClick: [id: string]
}>()

const query = shallowRef('')

const filteredGroups = computed(() => {
  const normalizedQuery = query.value.trim().toLowerCase()
  return props.groups
    .map((group) => ({
      ...group,
      articles: group.articleSlugs
        .map((slug) => props.articles.find((article) => article.slug === slug))
        .filter((article): article is TutorialArticle => Boolean(article))
        .filter((article) => {
          if (!normalizedQuery) return true
          const searchText = [
            article.title,
            article.description,
            ...article.sections.map((section) => section.title),
          ].join(' ').toLowerCase()
          return searchText.includes(normalizedQuery)
        }),
    }))
    .filter((group) => group.articles.length > 0)
})

function handleArticleClick() {
  emit('close')
}

function handleSectionClick(id: string) {
  emit('sectionClick', id)
  emit('close')
}
</script>

<template>
  <aside class="tutorial-directory">
    <div class="tutorial-directory__header">
      <div>
        <span>文档目录</span>
        <strong>教程文档</strong>
      </div>
      <button type="button" aria-label="关闭目录" @click="emit('close')">
        <Icon name="x" size="sm" />
      </button>
    </div>

    <label class="tutorial-directory__search">
      <Icon name="search" size="sm" />
      <input v-model="query" type="search" placeholder="搜索教程或配置" />
    </label>

    <nav class="tutorial-directory__nav" aria-label="教程文档目录">
      <div v-for="group in filteredGroups" :key="group.id" class="tutorial-directory__group">
        <div class="tutorial-directory__group-label">{{ group.label }}</div>
        <div v-for="article in group.articles" :key="article.slug">
          <RouterLink
            :to="`/tutorials/${article.slug}`"
            class="tutorial-directory__article"
            :class="{ 'tutorial-directory__article--active': currentSlug === article.slug }"
            @click="handleArticleClick"
          >
            <Icon :name="article.icon" size="sm" />
            <span>{{ article.shortTitle }}</span>
          </RouterLink>

          <div v-if="currentSlug === article.slug" class="tutorial-directory__sections">
            <button
              v-for="section in article.sections"
              :key="section.id"
              type="button"
              :class="{ 'tutorial-directory__section--active': activeSectionId === section.id }"
              @click="handleSectionClick(section.id)"
            >
              {{ section.title }}
            </button>
          </div>
        </div>
      </div>
    </nav>

    <div v-if="filteredGroups.length === 0" class="tutorial-directory__empty">
      没有找到相关教程
    </div>
  </aside>
</template>

<style scoped>
.tutorial-directory {
  display: flex;
  min-width: 0;
  height: 100%;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid rgba(23, 20, 17, 0.1);
  background: rgba(250, 248, 244, 0.72);
}

.tutorial-directory__header {
  display: flex;
  min-height: 58px;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px 9px 16px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.08);
}

.tutorial-directory__header span,
.tutorial-directory__header strong {
  display: block;
}

.tutorial-directory__header span {
  margin-bottom: 3px;
  color: #7c7267;
  font-size: 12px;
  letter-spacing: 2px;
}

.tutorial-directory__header strong {
  font-size: 15px;
}

.tutorial-directory__header button {
  display: none;
  width: 30px;
  height: 30px;
  place-items: center;
  border: 1px solid rgba(23, 20, 17, 0.09);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.72);
}

.tutorial-directory__search {
  display: flex;
  height: 34px;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
  margin: 13px 13px 8px;
  padding: 0 10px;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 999px;
  color: #9b938b;
  background: rgba(255, 255, 255, 0.72);
}

.tutorial-directory__search input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  color: #2b2723;
  background: transparent;
  font-size: 13px;
}

.tutorial-directory__search input::placeholder {
  color: #9b938b;
}

.tutorial-directory__nav {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding: 4px 11px 24px;
  scrollbar-width: thin;
  scrollbar-color: rgba(124, 114, 103, 0.3) transparent;
}

.tutorial-directory__group {
  padding-top: 13px;
}

.tutorial-directory__group-label {
  margin: 0 9px 7px;
  color: #7d4b1d;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 2px;
}

.tutorial-directory__article {
  display: grid;
  min-height: 42px;
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  margin: 2px 0;
  padding: 7px 9px;
  border: 1px solid transparent;
  border-radius: 9px;
  color: #5f5851;
  font-size: 14px;
  text-decoration: none;
}

.tutorial-directory__article span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-directory__article:hover {
  color: #2b2723;
  background: rgba(255, 255, 255, 0.62);
}

.tutorial-directory__article--active {
  border-color: rgba(15, 159, 144, 0.12);
  color: #087b70;
  background: rgba(15, 159, 144, 0.1);
  font-weight: 600;
}

.tutorial-directory__sections {
  margin: 3px 0 8px 17px;
  border-left: 1px solid #dcd3c9;
}

.tutorial-directory__sections button {
  display: block;
  width: 100%;
  padding: 8px 8px 8px 13px;
  border: 0;
  border-left: 2px solid transparent;
  color: #857c74;
  background: transparent;
  font-size: 12px;
  line-height: 1.5;
  text-align: left;
}

.tutorial-directory__sections button:hover,
.tutorial-directory__section--active {
  border-left-color: #0f9f90 !important;
  color: #087b70 !important;
}

.tutorial-directory__empty {
  padding: 24px 15px;
  color: #8c837b;
  font-size: 13px;
  text-align: center;
}

:global(.dark .tutorial-directory) {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.72);
}

:global(.dark .tutorial-directory__header),
:global(.dark .tutorial-directory__sections) {
  border-color: rgba(148, 163, 184, 0.14);
}

:global(.dark .tutorial-directory__header span),
:global(.dark .tutorial-directory__article),
:global(.dark .tutorial-directory__sections button),
:global(.dark .tutorial-directory__empty) {
  color: #94a3b8;
}

:global(.dark .tutorial-directory__header strong) {
  color: #f8fafc;
}

:global(.dark .tutorial-directory__search),
:global(.dark .tutorial-directory__header button) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(30, 41, 59, 0.72);
}

:global(.dark .tutorial-directory__search input) {
  color: #f8fafc;
}

:global(.dark .tutorial-directory__article--active) {
  border-color: rgba(45, 212, 191, 0.2);
  color: #5eead4;
  background: rgba(20, 184, 166, 0.12);
}

@media (max-width: 900px) {
  .tutorial-directory__header button {
    display: grid;
  }
}
</style>
