<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import TutorialCodeBlock from './TutorialCodeBlock.vue'
import TutorialScreenshot from './TutorialScreenshot.vue'
import type { TutorialArticle } from './types'

const props = defineProps<{
  article: TutorialArticle
  articles: TutorialArticle[]
}>()

const relatedArticles = computed(() =>
  props.article.relatedSlugs
    .map((slug) => props.articles.find((article) => article.slug === slug))
    .filter((article): article is TutorialArticle => Boolean(article))
)

const nextArticle = computed(() => {
  const currentIndex = props.articles.findIndex((article) => article.slug === props.article.slug)
  return currentIndex >= 0 ? props.articles[currentIndex + 1] : undefined
})
</script>

<template>
  <article class="tutorial-article">
    <header class="tutorial-article__header">
      <div class="tutorial-article__meta">
        <span class="tutorial-article__status"></span>
        <span>{{ article.eyebrow }}</span>
        <span>·</span>
        <span>更新于 {{ article.updatedAt }}</span>
        <span>·</span>
        <span>{{ article.duration }}</span>
      </div>
      <h1>{{ article.title }}</h1>
      <p>{{ article.description }}</p>
    </header>

    <section
      v-for="section in article.sections"
      :id="section.id"
      :key="section.id"
      class="tutorial-article__section"
      data-tutorial-section
    >
      <h2>{{ section.title }}</h2>
      <p v-if="section.description" class="tutorial-article__section-description">
        {{ section.description }}
      </p>

      <div class="tutorial-article__blocks">
        <template v-for="(block, blockIndex) in section.blocks" :key="`${section.id}-${blockIndex}`">
          <p v-if="block.type === 'paragraph'" class="tutorial-article__paragraph">
            {{ block.text }}
          </p>

          <component
            :is="block.ordered ? 'ol' : 'ul'"
            v-else-if="block.type === 'list'"
            class="tutorial-article__list"
          >
            <li v-for="item in block.items" :key="item">{{ item }}</li>
          </component>

          <div v-else-if="block.type === 'steps'" class="tutorial-steps">
            <div v-for="(step, stepIndex) in block.items" :key="step.title" class="tutorial-step">
              <span>{{ stepIndex + 1 }}</span>
              <div>
                <strong>{{ step.title }}</strong>
                <p>{{ step.description }}</p>
              </div>
            </div>
          </div>

          <TutorialCodeBlock
            v-else-if="block.type === 'code'"
            :code="block.code"
            :label="block.label"
            :language="block.language"
          />

          <div
            v-else-if="block.type === 'callout'"
            class="tutorial-callout"
            :class="`tutorial-callout--${block.tone}`"
          >
            <Icon
              :name="block.tone === 'warning' ? 'exclamationTriangle' : block.tone === 'success' ? 'checkCircle' : 'infoCircle'"
              size="sm"
            />
            <div>
              <strong>{{ block.title }}</strong>
              <p>{{ block.text }}</p>
            </div>
          </div>

          <TutorialScreenshot
            v-else-if="block.type === 'image'"
            :src="block.src"
            :alt="block.alt"
            :caption="block.caption"
          />

          <div v-else-if="block.type === 'links'" class="tutorial-links">
            <template v-for="link in block.links" :key="link.to">
              <a
                v-if="link.external"
                :href="link.to"
                target="_blank"
                rel="noopener noreferrer"
                class="tutorial-link"
              >
                {{ link.label }}
                <Icon name="externalLink" size="xs" />
              </a>
              <RouterLink v-else :to="link.to" class="tutorial-link">
                {{ link.label }}
                <Icon name="arrowRight" size="xs" />
              </RouterLink>
            </template>
          </div>
        </template>
      </div>
    </section>

    <footer v-if="relatedArticles.length" class="tutorial-related">
      <div>
        <span>继续阅读</span>
        <strong>下一步教程</strong>
      </div>
      <div class="tutorial-related__links">
        <RouterLink
          v-for="related in relatedArticles"
          :key="related.slug"
          :to="`/tutorials/${related.slug}`"
        >
          <Icon :name="related.icon" size="sm" />
          <span>{{ related.shortTitle }}</span>
          <Icon name="chevronRight" size="xs" />
        </RouterLink>
      </div>
    </footer>

    <nav v-if="nextArticle" class="tutorial-next" aria-label="下一篇教程">
      <RouterLink
        :to="`/tutorials/${nextArticle.slug}`"
        :aria-label="`下一篇：${nextArticle.title}`"
        :title="`下一篇：${nextArticle.title}`"
      >
        <span class="tutorial-next__copy">
          <small>下一篇</small>
          <strong>{{ nextArticle.title }}</strong>
        </span>
        <span class="tutorial-next__button" aria-hidden="true">
          <Icon name="arrowRight" size="sm" />
        </span>
      </RouterLink>
    </nav>
  </article>
</template>

<style scoped>
.tutorial-article {
  width: min(100%, 800px);
  margin: 0 auto;
  padding: 30px 36px 70px;
}

.tutorial-article__header {
  padding-bottom: 26px;
  border-bottom: 1px solid rgba(23, 20, 17, 0.1);
}

.tutorial-article__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 7px;
  margin-bottom: 14px;
  color: #7c7267;
  font-size: 11px;
}

.tutorial-article__status {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #0f9f90;
}

.tutorial-article__header h1 {
  margin: 0;
  color: #171411;
  font-family: "Songti SC", "STSong", "Noto Serif CJK SC", Georgia, serif;
  font-size: 38px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.18;
}

.tutorial-article__header > p {
  max-width: 700px;
  margin: 13px 0 0;
  color: #645c54;
  font-size: 14px;
  line-height: 1.75;
}

.tutorial-article__section {
  scroll-margin-top: 76px;
  padding-top: 34px;
}

.tutorial-article__section h2 {
  margin: 0;
  color: #171411;
  font-family: "Songti SC", "STSong", "Noto Serif CJK SC", Georgia, serif;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.35;
}

.tutorial-article__section-description {
  margin: 7px 0 0;
  color: #7c7267;
  line-height: 1.7;
}

.tutorial-article__blocks {
  display: grid;
  gap: 16px;
  margin-top: 17px;
}

.tutorial-article__paragraph {
  margin: 0;
  color: #514b45;
  line-height: 1.8;
}

.tutorial-article__list {
  display: grid;
  gap: 9px;
  margin: 0;
  padding-left: 23px;
  color: #514b45;
  line-height: 1.7;
}

.tutorial-article__list li::marker {
  color: #0f9f90;
}

.tutorial-steps {
  display: grid;
  gap: 14px;
}

.tutorial-step {
  display: grid;
  grid-template-columns: 30px minmax(0, 1fr);
  gap: 12px;
}

.tutorial-step > span {
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border: 1px solid rgba(15, 159, 144, 0.3);
  border-radius: 50%;
  color: #087b70;
  background: #eff8f5;
  font-family: Georgia, serif;
  font-weight: 700;
}

.tutorial-step strong {
  display: block;
  margin: 3px 0 4px;
  color: #241f1a;
  font-size: 13px;
}

.tutorial-step p {
  margin: 0;
  color: #746b62;
  line-height: 1.65;
}

.tutorial-callout {
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr);
  gap: 11px;
  padding: 14px 15px;
  border-left: 3px solid #4f7d74;
  color: #27564f;
  background: rgba(220, 236, 231, 0.54);
}

.tutorial-callout > svg {
  margin-top: 2px;
}

.tutorial-callout strong {
  display: block;
  margin-bottom: 4px;
  font-size: 12px;
}

.tutorial-callout p {
  margin: 0;
  line-height: 1.65;
}

.tutorial-callout--warning {
  border-left-color: #ad722e;
  color: #704a20;
  background: rgba(229, 203, 164, 0.25);
}

.tutorial-callout--success {
  border-left-color: #0f9f90;
  color: #1f5f56;
  background: rgba(203, 239, 227, 0.46);
}

.tutorial-links {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tutorial-link {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  gap: 7px;
  padding: 0 12px;
  border: 1px solid rgba(23, 20, 17, 0.11);
  border-radius: 999px;
  color: #355e57;
  background: rgba(255, 255, 255, 0.72);
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
  transition: border-color 0.16s ease, background 0.16s ease;
}

.tutorial-link:hover {
  border-color: rgba(15, 159, 144, 0.3);
  background: rgba(220, 236, 231, 0.56);
}

.tutorial-related {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 20px;
  margin-top: 48px;
  padding-top: 22px;
  border-top: 1px solid rgba(23, 20, 17, 0.1);
}

.tutorial-related > div:first-child span,
.tutorial-related > div:first-child strong {
  display: block;
}

.tutorial-related > div:first-child span {
  margin-bottom: 5px;
  color: #7c7267;
  font-size: 10px;
  letter-spacing: 2px;
}

.tutorial-related > div:first-child strong {
  font-family: "Songti SC", "STSong", Georgia, serif;
  font-size: 18px;
}

.tutorial-related__links {
  display: grid;
  gap: 7px;
}

.tutorial-related__links a {
  display: grid;
  min-height: 42px;
  grid-template-columns: 20px minmax(0, 1fr) 14px;
  align-items: center;
  gap: 9px;
  padding: 0 11px;
  border: 1px solid rgba(23, 20, 17, 0.09);
  border-radius: 9px;
  color: #2f2a25;
  background: rgba(255, 255, 255, 0.56);
  text-decoration: none;
}

.tutorial-related__links a:hover {
  border-color: rgba(15, 159, 144, 0.24);
  color: #087b70;
}

.tutorial-next {
  margin-top: 24px;
  border-top: 1px solid rgba(23, 20, 17, 0.1);
  border-bottom: 1px solid rgba(23, 20, 17, 0.1);
}

.tutorial-next > a {
  display: grid;
  min-height: 76px;
  grid-template-columns: minmax(0, 1fr) 40px;
  align-items: center;
  gap: 20px;
  color: #241f1a;
  text-decoration: none;
}

.tutorial-next__copy {
  min-width: 0;
}

.tutorial-next__copy small,
.tutorial-next__copy strong {
  display: block;
}

.tutorial-next__copy small {
  margin-bottom: 5px;
  color: #7c7267;
  font-size: 12px;
  font-weight: 600;
}

.tutorial-next__copy strong {
  overflow-wrap: anywhere;
  font-family: "Songti SC", "STSong", Georgia, serif;
  font-size: 18px;
  line-height: 1.4;
}

.tutorial-next__button {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  border: 1px solid rgba(15, 159, 144, 0.3);
  border-radius: 50%;
  color: #087b70;
  background: #eff8f5;
  transition: border-color 0.16s ease, color 0.16s ease, background 0.16s ease;
}

.tutorial-next > a:hover .tutorial-next__button {
  border-color: #0f9f90;
  color: #ffffff;
  background: #0f9f90;
}

.tutorial-next > a:focus-visible {
  outline: 2px solid rgba(15, 159, 144, 0.7);
  outline-offset: 4px;
}

:global(.dark .tutorial-article__header),
:global(.dark .tutorial-related),
:global(.dark .tutorial-next) {
  border-color: rgba(148, 163, 184, 0.16);
}

:global(.dark .tutorial-article__header h1),
:global(.dark .tutorial-article__section h2),
:global(.dark .tutorial-step strong),
:global(.dark .tutorial-related > div:first-child strong),
:global(.dark .tutorial-next__copy strong) {
  color: #f8fafc;
}

:global(.dark .tutorial-article__meta),
:global(.dark .tutorial-article__header > p),
:global(.dark .tutorial-article__section-description),
:global(.dark .tutorial-article__paragraph),
:global(.dark .tutorial-article__list),
:global(.dark .tutorial-step p),
:global(.dark .tutorial-related > div:first-child span),
:global(.dark .tutorial-next__copy small) {
  color: #94a3b8;
}

:global(.dark .tutorial-step > span) {
  border-color: rgba(45, 212, 191, 0.32);
  color: #5eead4;
  background: rgba(20, 184, 166, 0.12);
}

:global(.dark .tutorial-callout) {
  color: #bfe7df;
  background: rgba(20, 184, 166, 0.11);
}

:global(.dark .tutorial-callout--warning) {
  color: #f5d39f;
  background: rgba(180, 120, 48, 0.14);
}

:global(.dark .tutorial-link),
:global(.dark .tutorial-related__links a) {
  border-color: rgba(148, 163, 184, 0.16);
  color: #cbd5e1;
  background: rgba(15, 23, 42, 0.7);
}

:global(.dark .tutorial-next__button) {
  border-color: rgba(45, 212, 191, 0.32);
  color: #5eead4;
  background: rgba(20, 184, 166, 0.12);
}

:global(.dark .tutorial-next > a:hover .tutorial-next__button) {
  border-color: #14b8a6;
  color: #ffffff;
  background: #0f766e;
}

@media (max-width: 720px) {
  .tutorial-article {
    width: 100%;
    padding: 24px 20px 50px;
  }

  .tutorial-article__header h1 {
    font-size: 30px;
  }

  .tutorial-article__section {
    padding-top: 29px;
  }

  .tutorial-article__section h2 {
    font-size: 21px;
  }

  .tutorial-related {
    grid-template-columns: 1fr;
  }

  .tutorial-next > a {
    min-height: 72px;
    gap: 14px;
  }
}
</style>
