<script setup lang="ts">
defineProps<{
  kicker: string
  title: string
  description: string
}>()
</script>

<template>
  <section class="user-console-page">
    <header class="user-console-page__head">
      <div class="user-console-page__hero">
        <p class="user-console-page__kicker">{{ kicker }}</p>
        <h1 class="user-console-page__title">{{ title }}</h1>
        <p class="user-console-page__description">{{ description }}</p>
        <div v-if="$slots.heroNotes" class="user-console-page__notes">
          <slot name="heroNotes" />
        </div>
      </div>
      <aside v-if="$slots.heroAside" class="user-console-page__summary">
        <slot name="heroAside" />
      </aside>
    </header>

    <div class="user-console-page__body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.user-console-page {
  --console-paper: #f6f1e8;
  --console-paper-2: #fbf8f3;
  --console-ink: #171411;
  --console-muted: #70675c;
  --console-line: rgba(23, 20, 17, 0.08);
  --console-card: rgba(255, 255, 255, 0.72);
  --console-shadow: 0 1px 0 rgba(36, 29, 22, 0.025);

  display: flex;
  flex-direction: column;
  gap: 14px;
  max-width: 1560px;
  margin: 0 auto;
  color: var(--console-ink);
}

.user-console-page__head {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.user-console-page__hero,
.user-console-page__summary {
  border: 1px solid var(--console-line);
  box-shadow: var(--console-shadow);
}

.user-console-page__hero {
  min-width: 0;
  border-radius: 22px;
  padding: 17px 20px 15px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.88), rgba(248, 244, 236, 0.86)),
    var(--console-card);
}

.user-console-page__kicker {
  margin: 0;
  color: var(--console-muted);
  font-size: 12px;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.user-console-page__title {
  margin: 7px 0 7px;
  font-family: var(--console-font-sans);
  font-size: 27px;
  line-height: 1.18;
  letter-spacing: 0;
  font-weight: 700;
  color: var(--console-ink);
}

.user-console-page__description {
  max-width: 58rem;
  margin: 0;
  color: #5f584f;
  font-size: 14px;
  line-height: 1.56;
}

.user-console-page__notes {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 11px;
  margin-top: 15px;
  padding-top: 13px;
  border-top: 1px solid var(--console-line);
}

.user-console-page__summary {
  min-width: 0;
  border-radius: 22px;
  padding: 17px 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.84), rgba(248, 244, 236, 0.84));
}

.user-console-page__notes :deep(.console-note) {
  min-width: 0;
  border-radius: 15px;
  padding: 10px 13px 11px;
  background: rgba(255, 255, 255, 0.44);
  box-shadow: inset 0 0 0 1px rgba(23, 20, 17, 0.04);
}

.user-console-page__notes :deep(.console-note__title) {
  display: block;
  margin: 0 0 6px;
  color: var(--console-ink);
  font-size: 13px;
  font-weight: 700;
}

.user-console-page__notes :deep(.console-note__copy) {
  display: block;
  margin: 0;
  color: #61584e;
  font-size: 13px;
  line-height: 1.48;
}

.user-console-page__summary :deep(.console-summary) {
  display: grid;
  gap: 13px;
  min-height: 174px;
}

.user-console-page__summary :deep(.console-summary__label) {
  color: #7c5f43;
  font-size: 12px;
  letter-spacing: 0.24em;
  line-height: 1.2;
  text-transform: uppercase;
}

.user-console-page__summary :deep(.console-summary__value) {
  margin-top: 8px;
  color: var(--console-ink);
  font-family: var(--console-font-sans);
  font-size: 30px;
  font-weight: 700;
  line-height: 1.08;
}

.user-console-page__summary :deep(.console-summary__desc) {
  margin-top: 10px;
  color: #61584e;
  font-size: 13px;
  line-height: 1.5;
}

.user-console-page__summary :deep(.console-summary__micro) {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  border-top: 1px solid var(--console-line);
  padding-top: 13px;
}

.user-console-page__summary :deep(.console-summary__micro b) {
  display: block;
  color: var(--console-ink);
  font-size: 18px;
  line-height: 1.2;
}

.user-console-page__summary :deep(.console-summary__micro span) {
  display: block;
  margin-top: 8px;
  color: #7c7267;
  font-size: 13px;
}

.user-console-page__body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

:global(.dark) .user-console-page {
  --console-paper: #0f172a;
  --console-paper-2: #111827;
  --console-ink: #f8fafc;
  --console-muted: #94a3b8;
  --console-line: rgba(255, 255, 255, 0.08);
  --console-card: rgba(15, 23, 42, 0.78);
  --console-shadow: 0 16px 36px rgba(2, 6, 23, 0.18);
}

:global(.dark) .user-console-page__hero,
:global(.dark) .user-console-page__summary {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.78);
}

:global(.dark) .user-console-page__kicker {
  color: rgba(203, 213, 225, 0.78);
}

:global(.dark) .user-console-page__title {
  color: #f8fafc;
}

:global(.dark) .user-console-page__description {
  color: #cbd5e1;
}

:global(.dark) .user-console-page__notes :deep(.console-note) {
  background: rgba(255, 255, 255, 0.04);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.05);
}

:global(.dark) .user-console-page__notes :deep(.console-note__title),
:global(.dark) .user-console-page__summary :deep(.console-summary__value),
:global(.dark) .user-console-page__summary :deep(.console-summary__micro b) {
  color: #f8fafc;
}

:global(.dark) .user-console-page__notes :deep(.console-note__copy),
:global(.dark) .user-console-page__summary :deep(.console-summary__desc),
:global(.dark) .user-console-page__summary :deep(.console-summary__micro span) {
  color: #cbd5e1;
}

@media (min-width: 1024px) {
  .user-console-page__head {
    grid-template-columns: minmax(0, 1fr) 300px;
  }
}

@media (max-width: 640px) {
  .user-console-page__hero,
  .user-console-page__summary {
    padding: 16px;
    border-radius: 20px;
  }

  .user-console-page__title {
    font-size: 24px;
  }

  .user-console-page__description {
    font-size: 13px;
  }

  .user-console-page__notes {
    grid-template-columns: 1fr;
  }
}
</style>
