<script setup lang="ts">
defineProps<{
  title?: string
  description?: string
  compact?: boolean
}>()
</script>

<template>
  <section class="user-console-panel" :class="{ 'user-console-panel--compact': compact }">
    <header
      v-if="title || description || $slots.headerActions || $slots.headerMeta"
      class="user-console-panel__header"
    >
      <div class="user-console-panel__header-copy">
        <p v-if="title" class="user-console-panel__title">{{ title }}</p>
        <p v-if="description" class="user-console-panel__description">{{ description }}</p>
      </div>
      <div v-if="$slots.headerMeta || $slots.headerActions" class="user-console-panel__header-side">
        <slot name="headerMeta" />
        <slot name="headerActions" />
      </div>
    </header>
    <div class="user-console-panel__body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.user-console-panel {
  border: 1px solid rgba(23, 20, 17, 0.08);
  border-radius: 20px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.68);
  box-shadow: 0 1px 0 rgba(36, 29, 22, 0.025);
  overflow: hidden;
}

.user-console-panel--compact {
  border-radius: 20px;
}

.user-console-panel__header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 11px;
}

.user-console-panel__header-copy {
  min-width: 0;
}

.user-console-panel__title {
  margin: 0;
  font-family: var(--console-font-sans);
  font-size: 17px;
  line-height: 1.2;
  font-weight: 700;
  color: #171411;
}

.user-console-panel__description {
  margin: 0.3rem 0 0;
  font-size: 13px;
  line-height: 1.42;
  color: #7c7267;
}

.user-console-panel__header-side {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  flex-wrap: wrap;
}

.user-console-panel__body {
  min-width: 0;
}

:global(.dark) .user-console-panel {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.78);
  box-shadow: none;
}

:global(.dark) .user-console-panel__title {
  color: #f8fafc;
}

:global(.dark) .user-console-panel__description {
  color: #cbd5e1;
}

@media (max-width: 640px) {
  .user-console-panel {
    padding: 16px;
    border-radius: 22px;
  }

  .user-console-panel__header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
