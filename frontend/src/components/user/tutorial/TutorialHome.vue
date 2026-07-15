<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import { useTutorialCatalog } from '@/composables/useTutorialCatalog'
import { getLastTutorialSlug } from '@/utils/tutorialProgress'
import type { TutorialTag } from './types'

const { catalog } = useTutorialCatalog()

const selectedTag = shallowRef<TutorialTag | 'all'>('all')
const searchQuery = shallowRef('')
const lastReadSlug = shallowRef(getLastTutorialSlug())

const tabs: Array<{ id: TutorialTag | 'all'; label: string; icon: 'book' | 'terminal' | 'sparkles' | 'swap' | 'server' }> = [
  { id: 'all', label: '全部', icon: 'book' },
  { id: 'codex', label: 'Codex', icon: 'terminal' },
  { id: 'claude', label: 'Claude Code', icon: 'terminal' },
  { id: 'gemini', label: 'Gemini CLI', icon: 'sparkles' },
  { id: 'ccswitch', label: 'CCSwitch', icon: 'swap' },
  { id: 'openclaw', label: 'OpenClaw', icon: 'server' },
]

const continueArticle = computed(() =>
  catalog.value.articles.find((article) => article.slug === lastReadSlug.value)
    ?? catalog.value.articles[0]
)

const visibleArticles = computed(() => {
  const normalizedQuery = searchQuery.value.trim().toLowerCase()
  return catalog.value.articles.filter((article) => {
    if (selectedTag.value !== 'all' && article.tag !== selectedTag.value) return false
    if (!normalizedQuery) return true
    return [
      article.title,
      article.description,
      article.eyebrow,
      ...article.sections.map((section) => section.title),
    ].join(' ').toLowerCase().includes(normalizedQuery)
  })
})

const journey = computed(() => [
  {
    number: '1',
    title: '准备账户',
    description: '充值或兑换额度',
    to: '/tutorials/recharge',
  },
  {
    number: '2',
    title: '创建密钥',
    description: '独立命名并选择分组',
    to: '/tutorials/api-keys',
  },
  {
    number: '3',
    title: '配置客户端',
    description: '一键导入或手动配置',
    to: '/tutorials/ccswitch',
  },
  {
    number: '4',
    title: '验证调用',
    description: '检查模型、Token 与费用',
    to: '/tutorials/troubleshooting',
  },
])
</script>

<template>
  <div class="tutorial-home">
    <header class="tutorial-home__hero">
      <div>
        <span class="tutorial-home__eyebrow">DOCS / LEARNING PATH</span>
        <h1>教程文档</h1>
        <p>从账户准备到客户端接入，按真实操作顺序完成配置与排查。</p>
      </div>
      <RouterLink
        v-if="continueArticle"
        :to="`/tutorials/${continueArticle.slug}`"
        class="tutorial-home__continue"
      >
        <span>
          <small>继续上次阅读</small>
          {{ continueArticle.shortTitle }}
        </span>
        <Icon name="arrowRight" size="sm" />
      </RouterLink>
    </header>

    <div class="tutorial-home__filters">
      <label class="tutorial-home__search">
        <Icon name="search" size="sm" />
        <input v-model="searchQuery" type="search" placeholder="搜索客户端、配置或错误码" />
      </label>
      <div class="tutorial-home__tabs" aria-label="按客户端筛选">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          :class="{ 'tutorial-home__tab--active': selectedTag === tab.id }"
          @click="selectedTag = tab.id"
        >
          <Icon :name="tab.icon" size="sm" />
          {{ tab.label }}
        </button>
      </div>
    </div>

    <nav class="tutorial-journey" aria-label="首次接入流程">
      <RouterLink v-for="step in journey" :key="step.number" :to="step.to">
        <span class="tutorial-journey__number">{{ step.number }}</span>
        <strong>{{ step.title }}</strong>
        <small>{{ step.description }}</small>
      </RouterLink>
    </nav>

    <div class="tutorial-home__grid">
      <section class="tutorial-home__catalog">
        <div class="tutorial-home__section-heading">
          <div>
            <span>GUIDES</span>
            <h2>{{ selectedTag === 'all' ? '全部教程' : tabs.find((tab) => tab.id === selectedTag)?.label }}</h2>
          </div>
          <span>{{ visibleArticles.length }} 篇</span>
        </div>

        <div v-if="visibleArticles.length" class="tutorial-card-list">
          <RouterLink
            v-for="(article, index) in visibleArticles"
            :key="article.slug"
            :to="`/tutorials/${article.slug}`"
            class="tutorial-card"
          >
            <span class="tutorial-card__index">{{ String(index + 1).padStart(2, '0') }}</span>
            <span class="tutorial-card__icon"><Icon :name="article.icon" size="md" /></span>
            <span class="tutorial-card__content">
              <strong>{{ article.title }}</strong>
              <small>{{ article.description }}</small>
            </span>
            <span class="tutorial-card__meta">
              {{ article.duration }}
              <Icon name="arrowRight" size="xs" />
            </span>
          </RouterLink>
        </div>

        <div v-else class="tutorial-home__empty">
          <Icon name="search" size="lg" />
          <strong>没有找到相关教程</strong>
          <span>换一个关键词或客户端分类试试。</span>
        </div>
      </section>

      <aside class="tutorial-prep">
        <span class="tutorial-prep__eyebrow">BEFORE YOU START</span>
        <h2>开始前准备</h2>
        <p>先完成这三项，客户端接入会顺畅很多。</p>

        <div class="tutorial-prep__list">
          <RouterLink to="/redeem">
            <span>01</span>
            <div><strong>账户有可用额度</strong><small>自助充值或输入兑换码</small></div>
            <Icon name="chevronRight" size="xs" />
          </RouterLink>
          <RouterLink to="/keys">
            <span>02</span>
            <div><strong>为客户端创建独立密钥</strong><small>便于限额、停用与排查</small></div>
            <Icon name="chevronRight" size="xs" />
          </RouterLink>
          <RouterLink to="/model-market">
            <span>03</span>
            <div><strong>确认可用模型</strong><small>模型名称以模型广场为准</small></div>
            <Icon name="chevronRight" size="xs" />
          </RouterLink>
        </div>

        <div class="tutorial-prep__tip">
          <Icon name="lightbulb" size="sm" />
          <p>首次接入优先使用短请求验证，确认成功后再开启长上下文、工具调用或高推理强度。</p>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.tutorial-home {
  width: min(100%, 1180px);
  margin: 0 auto;
  padding: 4px 0 30px;
}

.tutorial-home__hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 30px;
  padding: 8px 0 24px;
}

.tutorial-home__eyebrow,
.tutorial-prep__eyebrow,
.tutorial-home__section-heading span:first-child {
  color: #7d4b1d;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 3px;
}

.tutorial-home__hero h1 {
  margin: 7px 0 7px;
  color: #171411;
  font-family: "Songti SC", "STSong", "Noto Serif CJK SC", Georgia, serif;
  font-size: 38px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.1;
}

.tutorial-home__hero p {
  margin: 0;
  color: #6d645c;
  line-height: 1.65;
}

.tutorial-home__continue {
  display: grid;
  min-width: 190px;
  min-height: 52px;
  grid-template-columns: minmax(0, 1fr) 18px;
  align-items: center;
  gap: 12px;
  padding: 8px 14px 8px 17px;
  border: 1px solid #171411;
  border-radius: 999px;
  color: #fff;
  background: #171411;
  text-decoration: none;
}

.tutorial-home__continue small,
.tutorial-home__continue span {
  display: block;
}

.tutorial-home__continue small {
  margin-bottom: 2px;
  color: #bdb8b2;
  font-size: 9px;
  font-weight: 500;
}

.tutorial-home__continue span {
  font-size: 12px;
  font-weight: 700;
}

.tutorial-home__filters {
  display: grid;
  grid-template-columns: minmax(230px, 0.75fr) minmax(0, 2fr);
  gap: 13px;
  padding: 14px 0;
  border-block: 1px solid rgba(23, 20, 17, 0.1);
}

.tutorial-home__search {
  display: flex;
  height: 42px;
  align-items: center;
  gap: 9px;
  padding: 0 13px;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 10px;
  color: #958b83;
  background: rgba(255, 255, 255, 0.66);
}

.tutorial-home__search input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  color: #2c2824;
  background: transparent;
  font-size: 13px;
}

.tutorial-home__tabs {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 7px;
}

.tutorial-home__tabs button {
  display: flex;
  min-width: 0;
  height: 42px;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 0 8px;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.58);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.tutorial-home__tabs button:hover,
.tutorial-home__tab--active {
  border-color: rgba(15, 159, 144, 0.3) !important;
  color: #087b70;
  background: rgba(220, 236, 231, 0.66) !important;
}

.tutorial-journey {
  position: relative;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
  margin: 26px 0 31px;
}

.tutorial-journey::before {
  position: absolute;
  top: 21px;
  right: 10%;
  left: 10%;
  height: 1px;
  background: #d8d0c6;
  content: '';
}

.tutorial-journey a {
  position: relative;
  display: grid;
  justify-items: center;
  color: #2b2723;
  text-align: center;
  text-decoration: none;
}

.tutorial-journey__number {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  margin-bottom: 10px;
  border: 1px solid rgba(23, 20, 17, 0.16);
  border-radius: 50%;
  color: #087b70;
  background: #fbf8f1;
  font-family: Georgia, serif;
  font-size: 16px;
  font-weight: 700;
  transition: border-color 0.16s ease, background 0.16s ease, color 0.16s ease;
}

.tutorial-journey a:hover .tutorial-journey__number {
  border-color: #0f9f90;
  color: #fff;
  background: #0f9f90;
}

.tutorial-journey strong {
  margin-bottom: 4px;
  font-size: 14px;
}

.tutorial-journey small {
  color: #7c7267;
  font-size: 12px;
}

.tutorial-home__grid {
  display: grid;
  grid-template-columns: minmax(0, 1.55fr) minmax(300px, 0.7fr);
  gap: 18px;
  align-items: start;
}

.tutorial-home__section-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 13px;
}

.tutorial-home__section-heading h2,
.tutorial-prep h2 {
  margin: 5px 0 0;
  color: #171411;
  font-family: "Songti SC", "STSong", Georgia, serif;
  font-size: 23px;
  font-weight: 700;
  letter-spacing: 0;
}

.tutorial-home__section-heading > span {
  color: #7c7267;
  font-size: 13px;
}

.tutorial-card-list {
  display: grid;
  gap: 9px;
}

.tutorial-card {
  display: grid;
  min-height: 90px;
  grid-template-columns: 30px 38px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  padding: 11px 14px;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 12px;
  color: inherit;
  background: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  transition: transform 0.16s ease, border-color 0.16s ease, background 0.16s ease;
}

.tutorial-card:hover {
  border-color: rgba(15, 159, 144, 0.28);
  background: rgba(255, 255, 255, 0.86);
  transform: translateY(-1px);
}

.tutorial-card__index {
  color: #0f8176;
  font-family: Georgia, serif;
  font-size: 14px;
  font-weight: 700;
}

.tutorial-card__icon {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border-radius: 9px;
  color: #087b70;
  background: #dcece7;
}

.tutorial-card__content {
  min-width: 0;
}

.tutorial-card__content strong,
.tutorial-card__content small {
  display: block;
}

.tutorial-card__content strong {
  margin-bottom: 5px;
  color: #231f1b;
  font-size: 15px;
}

.tutorial-card__content small {
  overflow: hidden;
  color: #746b62;
  font-size: 13px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tutorial-card__meta {
  display: flex;
  align-items: center;
  gap: 7px;
  color: #91877f;
  font-size: 12px;
  white-space: nowrap;
}

.tutorial-home__empty {
  display: grid;
  min-height: 260px;
  place-items: center;
  align-content: center;
  gap: 8px;
  border: 1px dashed rgba(23, 20, 17, 0.16);
  border-radius: 12px;
  color: #7c7267;
  text-align: center;
}

.tutorial-home__empty strong,
.tutorial-home__empty span {
  display: block;
}

.tutorial-home__empty span {
  font-size: 13px;
}

.tutorial-prep {
  padding: 18px 19px;
  border: 1px solid rgba(23, 20, 17, 0.1);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
}

.tutorial-prep > p {
  margin: 7px 0 15px;
  color: #746b62;
  font-size: 13px;
  line-height: 1.55;
}

.tutorial-prep__list {
  border-top: 1px solid rgba(23, 20, 17, 0.09);
}

.tutorial-prep__list a {
  display: grid;
  min-height: 78px;
  grid-template-columns: 30px minmax(0, 1fr) 14px;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.09);
  color: inherit;
  text-decoration: none;
}

.tutorial-prep__list a > span {
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  border-radius: 50%;
  color: #087b70;
  background: #e4f2ee;
  font-family: Georgia, serif;
  font-size: 12px;
  font-weight: 700;
}

.tutorial-prep__list strong,
.tutorial-prep__list small {
  display: block;
}

.tutorial-prep__list strong {
  margin-bottom: 4px;
  font-size: 14px;
}

.tutorial-prep__list small {
  color: #746b62;
  font-size: 12px;
  line-height: 1.45;
}

.tutorial-prep__tip {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  gap: 9px;
  margin-top: 15px;
  padding: 12px;
  border-left: 3px solid #b88646;
  color: #6c4a24;
  background: rgba(229, 203, 164, 0.24);
}

.tutorial-prep__tip p {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
}

:global(.dark .tutorial-home__hero h1),
:global(.dark .tutorial-home__section-heading h2),
:global(.dark .tutorial-prep h2),
:global(.dark .tutorial-card__content strong) {
  color: #f8fafc;
}

:global(.dark .tutorial-home__hero p),
:global(.dark .tutorial-journey small),
:global(.dark .tutorial-home__section-heading > span),
:global(.dark .tutorial-card__content small),
:global(.dark .tutorial-card__meta),
:global(.dark .tutorial-prep > p),
:global(.dark .tutorial-prep__list small) {
  color: #94a3b8;
}

:global(.dark .tutorial-home__filters),
:global(.dark .tutorial-prep__list),
:global(.dark .tutorial-prep__list a) {
  border-color: rgba(148, 163, 184, 0.14);
}

:global(.dark .tutorial-home__search),
:global(.dark .tutorial-home__tabs button),
:global(.dark .tutorial-card),
:global(.dark .tutorial-prep) {
  border-color: rgba(148, 163, 184, 0.16);
  color: #e5e7eb;
  background: rgba(15, 23, 42, 0.72);
}

:global(.dark .tutorial-home__search input) {
  color: #f8fafc;
}

:global(.dark .tutorial-journey::before) {
  background: rgba(148, 163, 184, 0.22);
}

:global(.dark .tutorial-journey a) {
  color: #e5e7eb;
}

:global(.dark .tutorial-journey__number) {
  border-color: rgba(148, 163, 184, 0.2);
  color: #5eead4;
  background: #0b1220;
}

@media (max-width: 1020px) {
  .tutorial-home__filters {
    grid-template-columns: 1fr;
  }

  .tutorial-home__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .tutorial-home__hero {
    align-items: stretch;
    flex-direction: column;
    gap: 17px;
  }

  .tutorial-home__continue {
    width: 100%;
  }

  .tutorial-home__tabs {
    display: flex;
    overflow-x: auto;
    padding-bottom: 3px;
    scrollbar-width: none;
  }

  .tutorial-home__tabs button {
    min-width: 112px;
  }

  .tutorial-journey {
    grid-template-columns: repeat(2, 1fr);
    gap: 18px 10px;
  }

  .tutorial-journey::before {
    display: none;
  }

  .tutorial-card {
    grid-template-columns: 28px 36px minmax(0, 1fr);
  }

  .tutorial-card__meta {
    display: none;
  }
}
</style>
