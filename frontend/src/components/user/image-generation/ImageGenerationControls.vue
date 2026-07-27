<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ImageGenerationKeyCapability } from '@/api/imageGenerations'
import type {
  ImageGenerationFormState,
  ImageGenerationMode,
  ImageResolutionTier
} from './types'

const props = defineProps<{
  standardKeys: ImageGenerationKeyCapability[]
  hdKeys: ImageGenerationKeyCapability[]
  loadingKeys: boolean
  generating: boolean
  canGenerate: boolean
  sourceImageName: string
  resolvedSize: string
  estimatedTotalPrice: number | null
  tierPrices: Record<ImageResolutionTier, number | null>
}>()

const emit = defineEmits<{
  generate: []
  enqueue: []
  reset: []
  sourceImage: [file?: File]
}>()

const form = defineModel<ImageGenerationFormState>('form', { required: true })
const mode = defineModel<ImageGenerationMode>('mode', { required: true })
const standardKeyId = defineModel<string>('standardKeyId', { required: true })
const hdKeyId = defineModel<string>('hdKeyId', { required: true })
const manualApiKey = defineModel<string>('manualApiKey', { required: true })

const { t } = useI18n()
const sourceInput = ref<HTMLInputElement | null>(null)

const tiers = computed(() => [
  { value: '1K' as const, label: '1K', description: t('imageGeneration.tiers.standard') },
  { value: '2K' as const, label: '2K', description: t('imageGeneration.tiers.hd') },
  { value: '4K' as const, label: '4K', description: t('imageGeneration.tiers.ultra') }
])

function priceForTier(tier: ImageResolutionTier): string {
  return formatPrice(props.tierPrices[tier])
}

function selectTier(tier: ImageResolutionTier) {
  form.value.resolution_tier = tier
}

function selectSourceImage() {
  sourceInput.value?.click()
}

function handleSourceImage(event: Event) {
  const input = event.target as HTMLInputElement
  emit('sourceImage', input.files?.[0])
}

function handleDrop(event: DragEvent) {
  emit('sourceImage', event.dataTransfer?.files?.[0])
}

function formatPrice(value: number | null): string {
  if (value === null) return t('imageGeneration.systemPrice')
  return `$${value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')}`
}
</script>

<template>
  <section class="image-controls">
    <div class="image-controls__scroll">
    <div class="image-pane-heading">
      <div>
        <h2>{{ t('imageGeneration.settings') }}</h2>
        <p>{{ t('imageGeneration.settingsHint') }}</p>
      </div>
      <span class="key-ready" :class="{ 'key-ready--muted': loadingKeys }">
        {{ loadingKeys ? t('common.loading') : t('imageGeneration.keyReady') }}
      </span>
    </div>

    <div class="mode-switch" role="tablist" :aria-label="t('imageGeneration.mode')">
      <button type="button" :class="{ active: mode === 'text' }" @click="mode = 'text'">
        {{ t('imageGeneration.textToImage') }}
      </button>
      <button type="button" :class="{ active: mode === 'image' }" @click="mode = 'image'">
        {{ t('imageGeneration.imageToImage') }}
      </button>
    </div>

    <div class="field-heading">
      <label for="image-prompt">{{ t('imageGeneration.prompt') }}</label>
      <span>{{ form.prompt.length }} / 4000</span>
    </div>
    <textarea
      id="image-prompt"
      v-model="form.prompt"
      class="prompt-input"
      maxlength="4000"
      :placeholder="t('imageGeneration.promptPlaceholder')"
    />

    <div
      v-if="mode === 'image'"
      class="source-upload"
      @dragover.prevent
      @drop.prevent="handleDrop"
    >
      <input ref="sourceInput" type="file" accept="image/*" class="hidden" @change="handleSourceImage" />
      <button type="button" @click="selectSourceImage">
        <Icon name="upload" size="sm" />
        <span>{{ sourceImageName || t('imageGeneration.chooseSourceImage') }}</span>
      </button>
      <button v-if="sourceImageName" type="button" class="clear-source" :title="t('imageGeneration.clearSourceImage')" @click="emit('sourceImage')">
        <Icon name="x" size="xs" />
      </button>
    </div>

    <div class="field-heading">
      <label>{{ t('imageGeneration.resolution') }}</label>
      <span>{{ t('imageGeneration.billOnSuccess') }}</span>
    </div>
    <div class="tier-grid">
      <button
        v-for="tier in tiers"
        :key="tier.value"
        type="button"
        class="tier-option"
        :class="{ active: form.resolution_tier === tier.value }"
        @click="selectTier(tier.value)"
      >
        <span class="tier-name">{{ tier.label }}</span>
        <span class="tier-price">{{ priceForTier(tier.value) }} / {{ t('imageGeneration.imageUnit') }}</span>
        <span class="tier-description">{{ tier.description }}</span>
      </button>
    </div>

    <div class="field-heading">
      <label>{{ t('imageGeneration.imageKey') }}</label>
      <span>{{ t('imageGeneration.matchByResolution') }}</span>
    </div>
    <div class="key-grid">
      <label class="key-pool" :class="{ active: form.resolution_tier === '1K', muted: form.resolution_tier !== '1K' }">
        <span class="key-pool-title"><b>{{ t('imageGeneration.standardKey') }}</b><small>1K</small></span>
        <select v-model="standardKeyId" :disabled="form.resolution_tier !== '1K' || loadingKeys">
          <option value="">{{ t('imageGeneration.autoSelect') }}</option>
          <option v-for="key in standardKeys" :key="key.api_key_id" :value="String(key.api_key_id)">{{ key.key_name }}</option>
        </select>
      </label>
      <label class="key-pool" :class="{ active: form.resolution_tier !== '1K', muted: form.resolution_tier === '1K' }">
        <span class="key-pool-title"><b>{{ t('imageGeneration.hdKey') }}</b><small>2K / 4K</small></span>
        <select v-model="hdKeyId" :disabled="form.resolution_tier === '1K' || loadingKeys">
          <option value="">{{ t('imageGeneration.autoSelect') }}</option>
          <option v-for="key in hdKeys" :key="key.api_key_id" :value="String(key.api_key_id)">{{ key.key_name }}</option>
        </select>
      </label>
    </div>
    <p v-if="!loadingKeys && (form.resolution_tier === '1K' ? standardKeys.length === 0 : hdKeys.length === 0)" class="key-warning">
      {{ form.resolution_tier === '1K' ? t('imageGeneration.noStandardKey') : t('imageGeneration.noHdKey') }}
    </p>

    <details class="manual-key">
      <summary>{{ t('imageGeneration.manualKey') }}</summary>
      <input v-model="manualApiKey" type="text" autocomplete="off" spellcheck="false" :placeholder="t('imageGeneration.manualKeyPlaceholder')" />
    </details>

    <div class="field-heading">
      <label>{{ t('imageGeneration.parameters') }}</label>
      <span>{{ form.model }}</span>
    </div>
    <div class="parameter-grid">
      <label>
        <span>{{ t('imageGeneration.aspectRatio') }}</span>
        <select v-model="form.aspect_ratio">
          <option value="1:1">1 : 1</option>
          <option value="2:3">2 : 3</option>
          <option value="3:2">3 : 2</option>
        </select>
      </label>
      <label>
        <span>{{ t('imageGeneration.quantity') }}</span>
        <select v-model.number="form.n">
          <option :value="1">1 {{ t('imageGeneration.imageUnit') }}</option>
          <option :value="2">2 {{ t('imageGeneration.imageUnit') }}</option>
          <option :value="3">3 {{ t('imageGeneration.imageUnit') }}</option>
          <option :value="4">4 {{ t('imageGeneration.imageUnit') }}</option>
        </select>
      </label>
      <label>
        <span>{{ t('imageGeneration.quality') }}</span>
        <select v-model="form.quality">
          <option value="auto">Auto</option>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
      </label>
      <label>
        <span>{{ t('imageGeneration.outputFormat') }}</span>
        <select v-model="form.output_format">
          <option value="png">PNG</option>
          <option value="jpeg">JPEG</option>
          <option value="webp">WebP</option>
        </select>
      </label>
    </div>

    </div>

    <div class="image-controls__footer">
      <div class="cost-row">
      <span>{{ t('imageGeneration.estimatedCost') }} <small>· {{ resolvedSize }}</small></span>
      <strong>{{ formatPrice(estimatedTotalPrice) }}</strong>
      </div>
      <button type="button" class="generate-button" :disabled="!canGenerate" @click="emit('generate')">
        <Icon v-if="generating" name="refresh" size="sm" class="spin" />
        <Icon v-else name="sparkles" size="sm" />
        {{ generating ? t('imageGeneration.generating') : t('imageGeneration.generateTier', { tier: form.resolution_tier }) }}
      </button>
      <div class="secondary-actions">
        <button type="button" :disabled="!canGenerate" @click="emit('enqueue')">{{ t('imageGeneration.addToQueue') }}</button>
        <button type="button" @click="emit('reset')">{{ t('common.reset') }}</button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.image-controls { min-width: 0; min-height: 0; padding: 15px; border-right: 1px solid var(--ig-line); background: var(--ig-panel-soft); display: flex; flex-direction: column; overflow: hidden; }
.image-controls__scroll { min-height: 0; flex: 1; overflow-y: auto; padding-right: 2px; }
.image-controls__footer { flex: 0 0 auto; margin: 8px -2px -2px; padding: 8px 2px 2px; border-top: 1px solid var(--ig-line); background: var(--ig-panel-soft); }
.image-pane-heading { min-height: 40px; display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.image-pane-heading h2 { margin: 0; color: var(--ig-ink); font-size: 16px; font-weight: 650; }
.image-pane-heading p { margin: 3px 0 0; color: var(--ig-muted); font-size: 13px; }
.key-ready { padding: 4px 8px; border-radius: 999px; color: var(--ig-teal-ink); background: var(--ig-teal-soft); font-size: 12px; font-weight: 650; }
.key-ready--muted { color: var(--ig-muted); background: var(--ig-control); }
.mode-switch { display: grid; grid-template-columns: 1fr 1fr; gap: 3px; margin-bottom: 11px; padding: 3px; border-radius: 7px; background: var(--ig-control); }
.mode-switch button { height: 35px; border: 0; border-radius: 5px; color: var(--ig-muted); background: transparent; font-size: 13px; }
.mode-switch button.active { color: var(--ig-ink); background: var(--ig-panel); box-shadow: 0 1px 5px rgba(40, 34, 25, 0.07); font-weight: 650; }
.field-heading { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin: 10px 0 6px; }
.field-heading label { color: var(--ig-ink-soft); font-size: 13px; font-weight: 650; }
.field-heading span { color: var(--ig-muted); font-size: 12px; }
.prompt-input { width: 100%; height: 92px; resize: vertical; border: 1px solid var(--ig-line); border-radius: 7px; outline: none; padding: 10px 11px; color: var(--ig-ink); background: var(--ig-panel); font-size: 14px; line-height: 1.65; }
.prompt-input:focus, select:focus, .manual-key input:focus { border-color: rgba(57, 125, 117, 0.45); box-shadow: 0 0 0 3px rgba(57, 125, 117, 0.09); }
.source-upload { position: relative; margin-top: 8px; border: 1px dashed var(--ig-line-strong); border-radius: 7px; background: var(--ig-panel); }
.source-upload > button:first-of-type { width: 100%; min-height: 42px; padding: 0 34px 0 10px; border: 0; color: var(--ig-muted); background: transparent; display: flex; align-items: center; gap: 8px; font-size: 13px; text-align: left; }
.clear-source { position: absolute; right: 7px; top: 7px; width: 24px; height: 24px; border: 1px solid var(--ig-line); border-radius: 5px; background: var(--ig-control); display: grid; place-items: center; }
.tier-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 7px; }
.tier-option { position: relative; min-height: 69px; padding: 9px; border: 1px solid var(--ig-line); border-radius: 7px; text-align: left; background: var(--ig-panel); }
.tier-option.active { border-color: var(--ig-teal); color: var(--ig-teal-ink); background: var(--ig-teal-soft); }
.tier-option.active::after { content: ''; position: absolute; top: 8px; right: 8px; width: 5px; height: 5px; border-radius: 50%; background: var(--ig-teal); }
.tier-name, .tier-price, .tier-description { display: block; }
.tier-name { color: inherit; font-size: 14px; font-weight: 700; }
.tier-price { margin-top: 4px; color: var(--ig-muted); font-size: 12px; }
.tier-description { margin-top: 3px; color: var(--ig-muted-light); font-size: 11px; }
.key-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 7px; }
.key-pool { min-width: 0; padding: 9px; border: 1px solid var(--ig-line); border-radius: 7px; background: var(--ig-panel); transition: opacity .2s, border-color .2s, background .2s; }
.key-pool.active { border-color: rgba(57, 125, 117, .38); background: var(--ig-teal-faint); }
.key-pool.muted { opacity: .48; }
.key-pool-title { display: flex; align-items: center; justify-content: space-between; color: var(--ig-ink-soft); font-size: 13px; }
.key-pool-title b { font-weight: 700; }
.key-pool-title small { color: var(--ig-teal); font-size: 11px; }
.key-pool select, .parameter-grid select { width: 100%; height: 38px; margin-top: 7px; border: 1px solid var(--ig-line); border-radius: 6px; padding: 0 8px; outline: none; color: var(--ig-ink-soft); background: var(--ig-panel); font-size: 13px; }
.key-warning { margin: 7px 0 0; color: #a56821; font-size: 12px; }
.manual-key { margin-top: 7px; color: var(--ig-muted); font-size: 12px; }
.manual-key summary { cursor: pointer; }
.manual-key input { width: 100%; height: 38px; margin-top: 6px; border: 1px solid var(--ig-line); border-radius: 6px; padding: 0 9px; outline: none; color: var(--ig-ink); background: var(--ig-panel); font-size: 13px; }
.parameter-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 7px; }
.parameter-grid label > span { display: block; color: var(--ig-ink-soft); font-size: 13px; font-weight: 600; }
.parameter-grid select { margin-top: 5px; font-size: 13px; }
.cost-row { padding: 2px 1px 9px; display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.cost-row span { color: var(--ig-muted); font-size: 13px; }
.cost-row small { color: var(--ig-muted-light); }
.cost-row strong { color: var(--ig-ink); font: 700 15px/1 ui-monospace, SFMono-Regular, Consolas, monospace; }
.generate-button { width: 100%; height: 43px; margin-top: 10px; border: 0; border-radius: 7px; color: #fff; background: #285f58; display: flex; align-items: center; justify-content: center; gap: 7px; font-size: 13px; font-weight: 700; }
.generate-button:disabled { cursor: not-allowed; opacity: .45; }
.secondary-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 7px; margin-top: 7px; }
.secondary-actions button { height: 38px; border: 1px solid var(--ig-line); border-radius: 6px; color: var(--ig-ink-soft); background: var(--ig-panel); font-size: 13px; }
.secondary-actions button:disabled { opacity: .45; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 820px) {
  .image-controls { border-right: 0; border-bottom: 1px solid var(--ig-line); }
}
</style>
