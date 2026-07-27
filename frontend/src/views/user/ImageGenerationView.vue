<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import ImageGenerationActivity from '@/components/user/image-generation/ImageGenerationActivity.vue'
import ImageGenerationControls from '@/components/user/image-generation/ImageGenerationControls.vue'
import ImageGenerationResult from '@/components/user/image-generation/ImageGenerationResult.vue'
import { useImageGenerationWorkspace } from '@/composables/useImageGenerationWorkspace'

const { t, locale } = useI18n()
const workspace = useImageGenerationWorkspace()

function formatDate(value: string): string {
  return new Date(value).toLocaleString(locale.value === 'zh' ? 'zh-CN' : 'en-US', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <AppLayout>
    <main class="image-studio">
      <header class="studio-header">
        <div class="studio-copy">
          <p class="studio-kicker">{{ t('imageGeneration.kicker') }}</p>
          <h1>{{ t('imageGeneration.title') }}</h1>
          <p class="studio-description">{{ t('imageGeneration.description') }}</p>
        </div>
        <div class="studio-actions">
          <button type="button" :disabled="workspace.loadingHistory.value" @click="workspace.loadHistory">
            <Icon name="refresh" size="sm" :class="{ spin: workspace.loadingHistory.value }" />
            {{ t('imageGeneration.refreshHistory') }}
          </button>
          <button
            type="button"
            class="primary"
            :disabled="workspace.queueItems.value.length === 0"
            @click="workspace.runQueue"
          >
            <Icon name="play" size="sm" />
            {{ workspace.queueRunning.value ? t('imageGeneration.queueRunning') : t('imageGeneration.runQueue') }}
            <span v-if="workspace.queueItems.value.length">{{ workspace.queueItems.value.length }}</span>
          </button>
        </div>
      </header>

      <section class="studio-workspace">
        <ImageGenerationControls
          v-model:form="workspace.form"
          v-model:mode="workspace.activeMode.value"
          v-model:standard-key-id="workspace.selectedKeyByPool.standard"
          v-model:hd-key-id="workspace.selectedKeyByPool.hd"
          v-model:manual-api-key="workspace.manualApiKey.value"
          :standard-keys="workspace.standardKeys.value"
          :hd-keys="workspace.hdKeys.value"
          :loading-keys="workspace.loadingApiKeys.value"
          :generating="workspace.generating.value"
          :can-generate="workspace.canGenerate.value"
          :source-image-name="workspace.sourceImageName.value"
          :resolved-size="workspace.resolvedSize.value"
          :estimated-total-price="workspace.estimatedTotalPrice.value"
          :tier-prices="workspace.tierPrices.value"
          @generate="workspace.generate"
          @enqueue="workspace.enqueueCurrent"
          @reset="workspace.resetCurrentForm"
          @source-image="workspace.setSourceImage"
        />

        <ImageGenerationResult
          :record="workspace.currentRecord.value"
          :active-generation="workspace.activeGeneration.value"
          :generating="workspace.generating.value"
          :error-message="workspace.errorMessage.value"
          :preview-url="workspace.previewUrl"
          @view="workspace.openViewer"
          @download="workspace.downloadImage"
          @delete="workspace.confirmDelete"
        />

        <ImageGenerationActivity
          :queue-items="workspace.queueItems.value"
          :queue-running="workspace.queueRunning.value"
          :history="workspace.history.value"
          :loading-history="workspace.loadingHistory.value"
          :current-record-id="workspace.currentRecord.value?.id"
          :preview-url="workspace.previewUrl"
          @select-record="workspace.selectRecord"
          @remove-queue="workspace.removeQueueItem"
          @versions="workspace.openVersions"
          @delete="workspace.confirmDelete"
        />
      </section>

      <BaseDialog
        :show="!!workspace.viewer.value"
        :title="t('imageGeneration.previewTitle')"
        width="full"
        @close="workspace.closeViewer"
      >
        <div v-if="workspace.viewer.value" class="viewer-body">
          <img :src="workspace.viewerUrl.value" :alt="workspace.viewer.value.record.prompt" />
          <p>{{ workspace.viewer.value.record.prompt }}</p>
        </div>
      </BaseDialog>

      <BaseDialog
        :show="workspace.showVersions.value"
        :title="t('imageGeneration.versionTitle')"
        width="wide"
        @close="workspace.closeVersions"
      >
        <div v-if="workspace.versionLoading.value" class="dialog-state">
          {{ t('imageGeneration.versionLoading') }}
        </div>
        <div v-else-if="workspace.versionItems.value.length === 0" class="dialog-state">
          {{ t('imageGeneration.versionEmpty') }}
        </div>
        <div v-else class="version-list">
          <article v-for="item in workspace.versionItems.value" :key="item.version">
            <header>
              <strong>{{ t('imageGeneration.versionNumber', { version: item.version }) }}</strong>
              <span>{{ formatDate(item.created_at) }}</span>
            </header>
            <p>{{ item.prompt }}</p>
            <small>{{ item.model }} · {{ item.size }} · {{ item.quality }} · {{ item.output_format.toUpperCase() }}</small>
          </article>
        </div>
      </BaseDialog>

      <ConfirmDialog
        :show="!!workspace.deleteTarget.value"
        :title="t('imageGeneration.deleteTitle')"
        :message="t('imageGeneration.deleteMessage')"
        :confirm-text="t('common.delete')"
        danger
        @confirm="workspace.deleteConfirmed"
        @cancel="workspace.deleteTarget.value = null"
      />
    </main>
  </AppLayout>
</template>

<style scoped>
.image-studio {
  --ig-paper: #f6f1e8;
  --ig-panel: #fffdf8;
  --ig-panel-soft: #faf6ef;
  --ig-control: #f0ece4;
  --ig-control-faint: #fbf8f2;
  --ig-ink: #191715;
  --ig-ink-soft: #39342f;
  --ig-muted: #766e65;
  --ig-muted-light: #9b9288;
  --ig-line: rgba(38, 33, 28, 0.11);
  --ig-line-strong: rgba(38, 33, 28, 0.19);
  --ig-teal: #397d75;
  --ig-teal-ink: #28685d;
  --ig-teal-soft: #e3efeb;
  --ig-teal-faint: #f0f6f3;
  width: min(100%, 1560px);
  margin: 0 auto;
  color: var(--ig-ink);
}

.studio-header {
  min-height: 94px;
  padding: 3px 2px 15px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 28px;
}

.studio-copy {
  min-width: 0;
}

.studio-kicker {
  margin: 0 0 5px;
  color: #947653;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
}

.studio-copy h1 {
  margin: 0;
  font-family: var(--console-font-sans);
  font-size: 30px;
  font-weight: 720;
  line-height: 1.12;
  letter-spacing: 0;
}

.studio-description {
  max-width: 760px;
  margin: 7px 0 0;
  color: var(--ig-muted);
  font-size: 14px;
  line-height: 1.7;
}

.studio-actions {
  display: flex;
  flex: 0 0 auto;
  gap: 7px;
}

.studio-actions button {
  min-height: 34px;
  padding: 0 11px;
  border: 1px solid var(--ig-line);
  border-radius: 7px;
  color: var(--ig-ink-soft);
  background: rgba(255, 253, 248, 0.84);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 650;
}

.studio-actions button.primary {
  border-color: #285f58;
  color: #fff;
  background: #285f58;
}

.studio-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.44;
}

.studio-actions button span {
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.18);
  display: inline-grid;
  place-items: center;
  font-size: 11px;
}

.studio-workspace {
  height: clamp(560px, calc(100dvh - 178px), 820px);
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--ig-line-strong);
  border-radius: 8px;
  background: var(--ig-panel);
  box-shadow: 0 18px 50px rgba(72, 57, 39, 0.08);
  display: grid;
  grid-template-columns: minmax(350px, 0.9fr) minmax(440px, 1.5fr) minmax(270px, 0.75fr);
}

.viewer-body {
  display: grid;
  gap: 12px;
}

.viewer-body img {
  width: 100%;
  max-height: 75vh;
  object-fit: contain;
}

.viewer-body p {
  margin: 0;
  color: var(--ig-muted);
  font-size: 13px;
  line-height: 1.7;
}

.dialog-state {
  padding: 38px 12px;
  color: var(--ig-muted);
  font-size: 13px;
  text-align: center;
}

.version-list {
  display: grid;
  gap: 9px;
}

.version-list article {
  padding: 14px;
  border: 1px solid var(--ig-line);
  border-radius: 7px;
  background: var(--ig-control-faint);
}

.version-list header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.version-list strong {
  color: var(--ig-ink);
  font-size: 14px;
}

.version-list span,
.version-list small {
  color: var(--ig-muted);
  font-size: 12px;
}

.version-list p {
  margin: 9px 0;
  color: var(--ig-ink-soft);
  font-size: 14px;
  line-height: 1.7;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 1439px) and (min-width: 821px) {
  .studio-workspace {
    height: clamp(620px, calc(100dvh - 178px), 820px);
    grid-template-columns: minmax(320px, 0.88fr) minmax(430px, 1.3fr);
    grid-template-rows: minmax(0, 1fr) minmax(180px, 240px);
  }
}

@media (max-width: 820px) {
  .studio-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 13px;
  }

  .studio-actions {
    width: 100%;
  }

  .studio-actions button {
    flex: 1;
  }

  .studio-workspace {
    min-height: 0;
    height: auto;
    grid-template-columns: 1fr;
    overflow: visible;
  }
}

@media (prefers-reduced-motion: reduce) {
  .spin {
    animation: none;
  }
}
</style>

<style>
.dark .image-studio {
  --ig-paper: #0a1115;
  --ig-panel: #111b1f;
  --ig-panel-soft: #0e171b;
  --ig-control: #182429;
  --ig-control-faint: #101a1e;
  --ig-ink: #f2f0eb;
  --ig-ink-soft: #d5d4cf;
  --ig-muted: #98a3a5;
  --ig-muted-light: #69777a;
  --ig-line: rgba(216, 229, 229, 0.1);
  --ig-line-strong: rgba(216, 229, 229, 0.17);
  --ig-teal: #62b8aa;
  --ig-teal-ink: #b9f1e6;
  --ig-teal-soft: rgba(45, 132, 119, 0.2);
  --ig-teal-faint: rgba(45, 132, 119, 0.1);
}

.dark .image-studio .studio-kicker {
  color: #b9a47f;
}

.dark .image-studio .studio-actions button {
  color: var(--ig-ink-soft);
  background: rgba(17, 27, 31, 0.92);
}

.dark .image-studio .studio-actions button.primary {
  border-color: #4aa797;
  color: #07110f;
  background: #63c4b2;
}

.dark .image-studio .studio-workspace {
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.22);
}
</style>
