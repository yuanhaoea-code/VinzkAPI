<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ProviderIcon from '@/components/user/monitor/ProviderIcon.vue'
import { modelMarketAPI } from '@/api'
import type {
  ModelMarketGroupPrice,
  ModelMarketModel,
  ModelMarketTokenPrice,
} from '@/api/modelMarket'
import { useAppStore } from '@/stores/app'

type BillingFilter = 'all' | 'token' | 'request'
type PriceUnit = 'M' | 'K'
type ViewMode = 'card' | 'table'

interface ImageTierPrice {
  tier: '1K' | '2K' | '4K'
  price: number
  groupName: string
}

const { t } = useI18n()
const appStore = useAppStore()

const models = ref<ModelMarketModel[]>([])
const loading = ref(true)
const searchQuery = ref('')
const selectedProvider = ref('all')
const selectedBilling = ref<BillingFilter>('all')
const selectedTag = ref('all')
const selectedModelId = ref<string | null>(null)
const priceUnit = ref<PriceUnit>('M')
const viewMode = ref<ViewMode>('card')
const imageTiers = ['1K', '2K', '4K'] as const

const providers = computed(() => {
  const values = new Map<string, string>()
  for (const model of models.value) values.set(model.provider, model.provider_label)
  return [
    { key: 'all', label: t('modelMarket.filters.allProviders') },
    ...Array.from(values, ([key, label]) => ({ key, label })).sort((a, b) => a.label.localeCompare(b.label)),
  ]
})

const tags = computed(() => {
  const values = new Set<string>()
  for (const model of models.value) for (const tag of model.tags) values.add(tag)
  return [
    { key: 'all', label: t('modelMarket.filters.allTags') },
    ...Array.from(values).sort().map((key) => ({ key, label: tagLabel(key) })),
  ]
})

const filteredModels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return models.value.filter((model) => {
    const matchesProvider = selectedProvider.value === 'all' || model.provider === selectedProvider.value
    const matchesBilling = selectedBilling.value === 'all' || model.billing === selectedBilling.value
    const matchesTag = selectedTag.value === 'all' || model.tags.includes(selectedTag.value)
    const matchesQuery = !query || [model.id, model.name, model.provider_label, ...model.tags]
      .some((value) => value.toLowerCase().includes(query))
    return matchesProvider && matchesBilling && matchesTag && matchesQuery
  })
})

const selectedModel = computed(() =>
  models.value.find((model) => model.id === selectedModelId.value) ?? null
)

const providerCount = computed(() => Math.max(0, providers.value.length - 1))
const tokenModelCount = computed(() => models.value.filter((model) => model.billing === 'token').length)

function bestTokenGroup(model: ModelMarketModel): ModelMarketGroupPrice | null {
  const groups = model.groups.filter((group) => group.token_price)
  if (groups.length === 0) return null
  return groups.reduce((best, current) => current.rate < best.rate ? current : best)
}

function bestTokenPrice(model: ModelMarketModel): ModelMarketTokenPrice | null {
  return bestTokenGroup(model)?.token_price ?? null
}

function imageTierPrices(model: ModelMarketModel): ImageTierPrice[] {
  const best = new Map<ImageTierPrice['tier'], ImageTierPrice>()
  for (const group of model.groups) {
    for (const tier of ['1K', '2K', '4K'] as const) {
      const value = group.image_prices?.[tier]
      if (value === undefined) continue
      const current = best.get(tier)
      if (!current || value < current.price) best.set(tier, { tier, price: value, groupName: group.group_name })
    }
  }
  return Array.from(best.values()).sort((a, b) => Number.parseInt(a.tier) - Number.parseInt(b.tier))
}

function formatTokenPrice(value: number | undefined): string {
  if (value === undefined) return '暂未定价'
  const divisor = priceUnit.value === 'K' ? 1000 : 1
  const unit = priceUnit.value === 'K' ? '1K Tokens' : '1M Tokens'
  return `$${(value / divisor).toFixed(4)} / ${unit}`
}

function formatImagePrice(value: number): string {
  return `$${value.toFixed(2)} / 张`
}

function formatMultiplier(value: number): string {
  return `${Number(value.toFixed(4))}x`
}

function billingLabel(model: ModelMarketModel): string {
  return model.billing === 'token' ? t('modelMarket.filters.tokenBilling') : t('modelMarket.filters.requestBilling')
}

function modelDescription(model: ModelMarketModel): string {
  if (model.billing === 'request') return '按清晰度固定计费，价格来自当前可用生图分组。'
  return `${model.provider_label} 模型，页面价格按官网基础价与当前可用分组倍率实时计算。`
}

function endpointSummary(model: ModelMarketModel): string {
  return model.endpoint_types.length > 0 ? model.endpoint_types.join(' / ') : '-'
}

function tagLabel(tag: string): string {
  const labels: Record<string, string> = {
    reasoning: t('modelMarket.tags.reasoning'),
    tools: t('modelMarket.tags.tools'),
    files: t('modelMarket.tags.files'),
    vision: t('modelMarket.tags.vision'),
    image: '生图',
  }
  return labels[tag] ?? tag
}

function modelPriceText(model: ModelMarketModel): string {
  const lines = [`${model.name} (${model.id})`, `${model.provider_label} · ${endpointSummary(model)}`]
  if (model.billing === 'request') {
    for (const item of imageTierPrices(model)) {
      lines.push(`${item.tier} ${formatImagePrice(item.price)} · ${item.groupName}`)
    }
    return lines.join('\n')
  }
  for (const group of model.groups) {
    if (!group.token_price) continue
    lines.push(
      `${group.group_name} (${formatMultiplier(group.rate)})`,
      `输入 ${formatTokenPrice(group.token_price.input)}`,
      `输出 ${formatTokenPrice(group.token_price.output)}`,
      `缓存读取 ${formatTokenPrice(group.token_price.cache_read)}`,
    )
  }
  return lines.join('\n')
}

async function copyModel(model: ModelMarketModel): Promise<void> {
  try {
    await navigator.clipboard.writeText(modelPriceText(model))
    appStore.showSuccess(t('modelMarket.copySuccess'))
  } catch {
    appStore.showError(t('modelMarket.copyFailed'))
  }
}

async function copyVisibleModels(): Promise<void> {
  try {
    await navigator.clipboard.writeText(filteredModels.value.map(modelPriceText).join('\n\n'))
    appStore.showSuccess(t('modelMarket.copySuccess'))
  } catch {
    appStore.showError(t('modelMarket.copyFailed'))
  }
}

async function loadCatalog(silent = false): Promise<void> {
  if (!silent) loading.value = true
  try {
    const catalog = await modelMarketAPI.getCatalog()
    models.value = catalog.models ?? []
  } catch (error) {
    console.error('[ModelMarket] Failed to load catalog', error)
    appStore.showError(t('modelMarket.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => void loadCatalog())
</script>

<template>
  <div class="market-page">
    <header class="market-header">
      <div>
        <p class="market-kicker">MODEL MARKET</p>
        <div class="market-title-row">
          <h1>{{ t('modelMarket.title') }}</h1>
          <span>{{ t('modelMarket.modelCount', { count: filteredModels.length }) }}</span>
        </div>
        <p class="market-intro">
          查看当前账号可用模型与分组价格。Token 价格按官网基础价乘以分组倍率计算，卡片优先展示最低倍率分组。
        </p>
      </div>
      <div class="market-note">
        <strong>实时价格目录</strong>
        <span>模型、分组和用户专属倍率均来自后台配置；生图价格按 1K、2K、4K 分组固定价格展示。</span>
      </div>
    </header>

    <section class="market-toolbar" aria-label="模型筛选">
      <div class="toolbar-row">
        <label class="search-field" for="market-search">
          <Icon name="search" size="md" />
          <input id="market-search" v-model="searchQuery" type="search" :placeholder="t('modelMarket.searchPlaceholder')" />
        </label>
        <div class="toolbar-actions">
          <button type="button" title="刷新价格" @click="loadCatalog(true)">
            <Icon name="refresh" size="sm" />
            <span>刷新</span>
          </button>
          <button type="button" @click="copyVisibleModels">
            <Icon name="copy" size="sm" />
            <span>{{ t('common.copy') }}</span>
          </button>
          <button type="button" :class="{ active: viewMode === 'table' }" @click="viewMode = viewMode === 'card' ? 'table' : 'card'">
            <Icon :name="viewMode === 'card' ? 'menu' : 'grid'" size="sm" />
            <span>{{ viewMode === 'card' ? t('modelMarket.tableView') : '卡片视图' }}</span>
          </button>
          <button type="button" :class="{ active: priceUnit === 'K' }" @click="priceUnit = priceUnit === 'M' ? 'K' : 'M'">
            {{ priceUnit === 'M' ? '1M Tokens' : '1K Tokens' }}
          </button>
        </div>
      </div>

      <div class="filter-row">
        <div class="filter-set">
          <span>供应商</span>
          <button
            v-for="provider in providers"
            :key="provider.key"
            type="button"
            :class="{ active: selectedProvider === provider.key }"
            @click="selectedProvider = provider.key"
          >
            {{ provider.label }}
          </button>
        </div>
        <div class="filter-set filter-set-right">
          <span>计费与能力</span>
          <button type="button" :class="{ active: selectedBilling === 'all' }" @click="selectedBilling = 'all'">全部计费</button>
          <button type="button" :class="{ active: selectedBilling === 'token' }" @click="selectedBilling = 'token'">按量计费</button>
          <button type="button" :class="{ active: selectedBilling === 'request' }" @click="selectedBilling = 'request'">按次计费</button>
          <button
            v-for="tag in tags"
            :key="tag.key"
            type="button"
            :class="{ active: selectedTag === tag.key }"
            @click="selectedTag = tag.key"
          >
            {{ tag.label }}
          </button>
        </div>
      </div>
    </section>

    <section class="market-summary" aria-label="模型广场摘要">
      <div><strong>{{ filteredModels.length }}</strong><span>当前可用模型</span></div>
      <div><strong>{{ providerCount }}</strong><span>模型供应商</span></div>
      <div><strong>{{ tokenModelCount }}</strong><span>Token 计费模型</span></div>
      <div><strong>最低倍率</strong><span>卡片价格展示规则</span></div>
    </section>

    <div v-if="loading" class="market-state">正在读取模型与分组价格…</div>

    <section v-else-if="viewMode === 'card' && filteredModels.length > 0" class="model-grid" aria-label="模型卡片">
      <article v-for="model in filteredModels" :key="model.id" class="model-card">
        <button type="button" class="card-main" @click="selectedModelId = model.id">
          <div class="card-head">
            <span class="provider-icon" :class="`provider-${model.provider}`">
              <ProviderIcon :provider="model.provider" :size="22" />
            </span>
            <div>
              <h2>{{ model.name }}</h2>
              <p>{{ model.provider_label }} · {{ endpointSummary(model) }}</p>
            </div>
          </div>

          <div v-if="model.billing === 'token'" class="price-list">
            <p><span>{{ t('modelMarket.inputPrice') }}</span><strong>{{ formatTokenPrice(bestTokenPrice(model)?.input) }}</strong></p>
            <p><span>{{ t('modelMarket.outputPrice') }}</span><strong>{{ formatTokenPrice(bestTokenPrice(model)?.output) }}</strong></p>
            <p><span>{{ t('modelMarket.cacheReadPrice') }}</span><strong>{{ formatTokenPrice(bestTokenPrice(model)?.cache_read) }}</strong></p>
          </div>
          <div v-else class="price-list image-price-list">
            <p v-for="item in imageTierPrices(model)" :key="item.tier">
              <span>{{ item.tier }} 生图</span><strong>{{ formatImagePrice(item.price) }}</strong>
            </p>
          </div>

          <p class="model-description">{{ modelDescription(model) }}</p>
          <div class="model-meta">
            <span>{{ billingLabel(model) }}</span>
            <span v-if="bestTokenGroup(model)">{{ bestTokenGroup(model)?.group_name }} · {{ formatMultiplier(bestTokenGroup(model)?.rate ?? 0) }}</span>
            <span v-for="tag in model.tags.slice(0, 3)" :key="tag">{{ tagLabel(tag) }}</span>
          </div>
        </button>
        <button type="button" class="copy-button" :title="t('common.copy')" @click="copyModel(model)">
          <Icon name="copy" size="sm" />
        </button>
      </article>
    </section>

    <section v-else-if="viewMode === 'table' && filteredModels.length > 0" class="model-table-panel">
      <div class="model-table-wrap">
        <table>
          <thead><tr><th>模型</th><th>供应商</th><th>调用端点</th><th>最低分组</th><th>价格</th></tr></thead>
          <tbody>
            <tr v-for="model in filteredModels" :key="model.id" @click="selectedModelId = model.id">
              <td><strong>{{ model.name }}</strong><small>{{ model.id }}</small></td>
              <td>{{ model.provider_label }}</td>
              <td>{{ endpointSummary(model) }}</td>
              <td>{{ bestTokenGroup(model)?.group_name ?? imageTierPrices(model)[0]?.groupName ?? '-' }}</td>
              <td>
                <template v-if="model.billing === 'token'">
                  <span>输入 {{ formatTokenPrice(bestTokenPrice(model)?.input) }}</span>
                  <span>输出 {{ formatTokenPrice(bestTokenPrice(model)?.output) }}</span>
                </template>
                <template v-else>
                  <span v-for="item in imageTierPrices(model)" :key="item.tier">{{ item.tier }} {{ formatImagePrice(item.price) }}</span>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div v-else class="market-state">{{ t('modelMarket.empty') }}</div>

    <Transition name="fade">
      <div v-if="selectedModel" class="drawer-mask" @click="selectedModelId = null" />
    </Transition>
    <Transition name="drawer">
      <aside v-if="selectedModel" class="model-drawer" aria-label="模型价格详情">
        <header>
          <div class="card-head">
            <span class="provider-icon" :class="`provider-${selectedModel.provider}`">
              <ProviderIcon :provider="selectedModel.provider" :size="22" />
            </span>
            <div><h2>{{ selectedModel.name }}</h2><p>{{ selectedModel.id }}</p></div>
          </div>
          <button type="button" title="关闭" @click="selectedModelId = null"><Icon name="x" size="md" /></button>
        </header>

        <div class="drawer-content">
          <section>
            <p class="market-kicker">PRICE BASIS</p>
            <h3>价格计算方式</h3>
            <p v-if="selectedModel.billing === 'token'">官网基础价 × 当前用户分组倍率。下面列出所有可用分组，首页卡片使用倍率最低的一组。</p>
            <p v-else>生图按成功生成的图片张数计费，不乘 Token 分组倍率。</p>
            <div v-if="selectedModel.official_price" class="official-price">
              <span>官网输入 {{ formatTokenPrice(selectedModel.official_price.input) }}</span>
              <span>官网输出 {{ formatTokenPrice(selectedModel.official_price.output) }}</span>
              <span>缓存读取 {{ formatTokenPrice(selectedModel.official_price.cache_read) }}</span>
            </div>
          </section>

          <section>
            <p class="market-kicker">GROUP PRICING</p>
            <h3>可用分组价格</h3>
            <div class="group-pricing">
              <div v-for="group in selectedModel.groups" :key="group.group_id" class="group-price-row">
                <div><strong>{{ group.group_name }}</strong><span>{{ group.platform }} · {{ formatMultiplier(group.rate) }}</span></div>
                <div v-if="group.token_price">
                  <span>输入 {{ formatTokenPrice(group.token_price.input) }}</span>
                  <span>输出 {{ formatTokenPrice(group.token_price.output) }}</span>
                  <span>缓存 {{ formatTokenPrice(group.token_price.cache_read) }}</span>
                </div>
                <div v-else>
                  <span v-for="tier in imageTiers" :key="tier">
                    <template v-if="group.image_prices?.[tier] !== undefined">{{ tier }} {{ formatImagePrice(group.image_prices?.[tier] ?? 0) }}</template>
                  </span>
                </div>
              </div>
            </div>
          </section>
        </div>
      </aside>
    </Transition>
  </div>
</template>

<style scoped>
.market-page { color: #171916; font-family: Inter, "PingFang SC", "Microsoft YaHei", sans-serif; }
.market-header { display: grid; grid-template-columns: minmax(0, 1fr) 300px; gap: 20px; align-items: end; margin-bottom: 18px; }
.market-kicker { margin: 0 0 7px; color: #9a7c4f; font-size: 12px; font-weight: 700; letter-spacing: 0; text-transform: uppercase; }
.market-title-row { display: flex; align-items: center; gap: 12px; }
.market-title-row h1 { margin: 0; font-family: "Noto Serif SC", "Songti SC", SimSun, serif; font-size: 32px; line-height: 1.2; letter-spacing: 0; }
.market-title-row > span { border: 1px solid #cbd9d2; border-radius: 999px; background: #edf4ef; padding: 5px 10px; color: #456457; font-size: 12px; font-weight: 700; }
.market-intro { max-width: 760px; margin: 9px 0 0; color: #676b65; font-size: 14px; line-height: 1.75; }
.market-note { display: grid; gap: 6px; border: 1px solid #e1e0d8; border-radius: 8px; background: #fbfaf6; padding: 15px 17px; }
.market-note strong { font-family: "Noto Serif SC", "Songti SC", SimSun, serif; font-size: 16px; }
.market-note span { color: #74776f; font-size: 13px; line-height: 1.6; }
.market-toolbar, .model-table-panel { border: 1px solid #e0dfd7; border-radius: 8px; background: #fff; }
.toolbar-row { display: flex; gap: 12px; padding: 13px; border-bottom: 1px solid #ecebe5; }
.search-field { min-width: 240px; flex: 1; display: flex; align-items: center; gap: 9px; border: 1px solid #deded8; border-radius: 6px; padding: 0 13px; color: #858980; }
.search-field:focus-within { border-color: #769486; box-shadow: 0 0 0 3px rgba(83, 125, 104, .1); }
.search-field input { width: 100%; height: 40px; border: 0; outline: 0; background: transparent; color: inherit; font: inherit; }
.toolbar-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.toolbar-actions button, .filter-set button { border: 1px solid #deded8; border-radius: 6px; background: #fff; color: #353934; cursor: pointer; font-size: 13px; }
.toolbar-actions button { min-height: 40px; display: inline-flex; align-items: center; gap: 7px; padding: 0 13px; }
.toolbar-actions button:hover, .toolbar-actions button.active, .filter-set button:hover, .filter-set button.active { border-color: #8aa497; background: #edf4ef; color: #345744; }
.filter-row { display: grid; grid-template-columns: 1fr 1.6fr; gap: 20px; padding: 14px; }
.filter-set { display: flex; flex-wrap: wrap; align-items: center; gap: 7px; }
.filter-set > span { width: 100%; color: #9a7c4f; font-size: 11px; font-weight: 700; text-transform: uppercase; }
.filter-set button { padding: 7px 11px; }
.filter-set-right { justify-content: flex-end; }
.filter-set-right > span { text-align: right; }
.market-summary { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; margin: 12px 0; }
.market-summary > div { min-height: 74px; display: grid; align-content: center; gap: 4px; border: 1px solid #e2e1da; border-radius: 8px; background: #fbfaf7; padding: 12px 15px; }
.market-summary strong { font-family: "Noto Serif SC", "Songti SC", SimSun, serif; font-size: 23px; }
.market-summary span { color: #777a73; font-size: 12px; }
.model-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
.model-card { position: relative; min-width: 0; border: 1px solid #dfded6; border-radius: 8px; background: #fff; overflow: hidden; transition: transform .18s ease, border-color .18s ease, box-shadow .18s ease; }
.model-card:hover { transform: translateY(-2px); border-color: #9eb0a6; box-shadow: 0 10px 24px rgba(37, 44, 39, .07); }
.card-main { width: 100%; min-height: 300px; display: flex; flex-direction: column; border: 0; background: transparent; padding: 17px; color: inherit; cursor: pointer; text-align: left; }
.card-head { min-width: 0; display: flex; align-items: center; gap: 11px; }
.card-head > div { min-width: 0; }
.card-head h2 { margin: 0; overflow-wrap: anywhere; font-family: "Noto Serif SC", "Songti SC", SimSun, serif; font-size: 18px; letter-spacing: 0; }
.card-head p { margin: 3px 0 0; overflow-wrap: anywhere; color: #777b74; font-size: 12px; }
.provider-icon { width: 36px; height: 36px; flex: 0 0 auto; display: grid; place-items: center; border: 1px solid #e1e0d9; border-radius: 7px; background: #faf9f5; color: #2d4f40; }
.provider-anthropic { color: #c46649; }
.provider-gemini { color: #3978c8; }
.price-list { display: grid; gap: 7px; margin: 19px 0 14px; padding-top: 15px; border-top: 1px solid #ecebe5; }
.price-list p { display: flex; justify-content: space-between; gap: 12px; margin: 0; font-size: 13px; }
.price-list span { color: #6e726b; }
.price-list strong { color: #242723; font-size: 13px; text-align: right; }
.model-description { min-height: 45px; margin: 0; color: #686c66; font-size: 13px; line-height: 1.7; }
.model-meta { display: flex; flex-wrap: wrap; gap: 6px; margin-top: auto; padding-top: 15px; }
.model-meta span { border-radius: 999px; background: #f2f2ed; padding: 5px 8px; color: #5c625c; font-size: 11px; }
.copy-button { position: absolute; top: 13px; right: 13px; width: 32px; height: 32px; display: grid; place-items: center; border: 1px solid #e0dfd8; border-radius: 6px; background: #fff; color: #6c716a; cursor: pointer; }
.copy-button:hover { background: #edf4ef; color: #345744; }
.market-state { display: grid; min-height: 180px; place-items: center; border: 1px dashed #d8d7d0; border-radius: 8px; color: #797d76; }
.model-table-wrap { overflow-x: auto; }
table { width: 100%; min-width: 850px; border-collapse: collapse; }
th, td { padding: 13px 15px; border-bottom: 1px solid #ecebe5; text-align: left; vertical-align: top; font-size: 13px; }
th { color: #767a73; font-size: 11px; text-transform: uppercase; }
td strong, td small, td span { display: block; }
td small { margin-top: 3px; color: #898d85; }
tbody tr { cursor: pointer; }
tbody tr:hover { background: #f7f8f4; }
.drawer-mask { position: fixed; inset: 0; z-index: 70; background: rgba(18, 21, 18, .36); backdrop-filter: blur(2px); }
.model-drawer { position: fixed; top: 0; right: 0; z-index: 71; width: min(520px, 100vw); height: 100vh; overflow-y: auto; border-left: 1px solid #d9d8d0; background: #fbfaf6; box-shadow: -20px 0 50px rgba(22, 28, 24, .15); }
.model-drawer > header { position: sticky; top: 0; z-index: 1; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e0dfd7; background: rgba(251, 250, 246, .96); padding: 18px 20px; }
.model-drawer > header > button { width: 36px; height: 36px; display: grid; place-items: center; border: 1px solid #deddd5; border-radius: 6px; background: #fff; cursor: pointer; }
.drawer-content { display: grid; gap: 26px; padding: 22px 20px 36px; }
.drawer-content section > h3 { margin: 0 0 8px; font-family: "Noto Serif SC", "Songti SC", SimSun, serif; font-size: 18px; }
.drawer-content section > p:not(.market-kicker) { margin: 0; color: #686c65; font-size: 13px; line-height: 1.7; }
.official-price { display: grid; gap: 5px; margin-top: 12px; border-left: 2px solid #789486; padding: 8px 12px; background: #f0f4f0; font-size: 12px; }
.group-pricing { display: grid; margin-top: 12px; border-top: 1px solid #dddcd5; }
.group-price-row { display: grid; grid-template-columns: 145px minmax(0, 1fr); gap: 14px; padding: 13px 0; border-bottom: 1px solid #e5e4dd; }
.group-price-row > div { min-width: 0; display: grid; gap: 4px; }
.group-price-row strong { overflow-wrap: anywhere; font-size: 13px; }
.group-price-row span { color: #6f736d; font-size: 12px; }
.fade-enter-active, .fade-leave-active { transition: opacity .18s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.drawer-enter-active, .drawer-leave-active { transition: transform .24s ease; }
.drawer-enter-from, .drawer-leave-to { transform: translateX(100%); }
:global(.dark .market-page) { color: #e8ebe5; }
:global(.dark .market-note), :global(.dark .market-toolbar), :global(.dark .model-card), :global(.dark .model-table-panel), :global(.dark .market-summary > div) { border-color: #303833; background: #171d19; }
:global(.dark .market-intro), :global(.dark .market-note span), :global(.dark .card-head p), :global(.dark .model-description), :global(.dark .price-list span), :global(.dark td small) { color: #9ba59d; }
:global(.dark .search-field), :global(.dark .toolbar-actions button), :global(.dark .filter-set button), :global(.dark .copy-button) { border-color: #364039; background: #121713; color: #cdd4ce; }
:global(.dark .provider-icon) { border-color: #39423c; background: #111713; }
:global(.dark .price-list), :global(.dark th), :global(.dark td), :global(.dark .toolbar-row) { border-color: #303833; }
:global(.dark .price-list strong) { color: #eef0ed; }
:global(.dark .model-meta span) { background: #252d28; color: #b8c1ba; }
:global(.dark tbody tr:hover) { background: #1d2520; }
:global(.dark .model-drawer), :global(.dark .model-drawer > header) { border-color: #303833; background: #151b17; }
:global(.dark .official-price) { background: #1d2922; }
@media (max-width: 1100px) { .model-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .market-header { grid-template-columns: 1fr; } .market-note { max-width: 620px; } }
@media (max-width: 760px) { .toolbar-row { flex-direction: column; } .toolbar-actions { display: grid; grid-template-columns: repeat(2, 1fr); } .toolbar-actions button { justify-content: center; } .filter-row { grid-template-columns: 1fr; } .filter-set-right, .filter-set-right > span { justify-content: flex-start; text-align: left; } .market-summary { grid-template-columns: repeat(2, 1fr); } .model-grid { grid-template-columns: 1fr; } .card-main { min-height: 280px; } }
@media (max-width: 420px) { .market-title-row { align-items: flex-start; flex-direction: column; } .market-title-row h1 { font-size: 28px; } .toolbar-actions { grid-template-columns: 1fr; } .market-summary { grid-template-columns: 1fr; } .group-price-row { grid-template-columns: 1fr; } }
</style>
