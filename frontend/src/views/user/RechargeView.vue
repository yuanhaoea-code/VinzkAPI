<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1560px] space-y-6">
      <div class="card overflow-hidden">
        <div class="border-b border-gray-100 px-6 py-5 dark:border-dark-700">
          <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
                {{ t('recharge.title') }}
              </h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ t('recharge.description') }}
              </p>
            </div>
            <div class="flex flex-wrap gap-3">
              <a
                :href="shopUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary"
              >
                <Icon name="externalLink" size="sm" class="mr-2" />
                {{ t('recharge.openInNewTab') }}
              </a>
              <button type="button" class="btn btn-primary" @click="goRedeem">
                <Icon name="arrowRight" size="sm" class="mr-2" />
                {{ t('recharge.goRedeem') }}
              </button>
            </div>
          </div>
        </div>

        <div class="border-b border-amber-100 bg-amber-50 px-6 py-4 dark:border-amber-900/40 dark:bg-amber-900/20">
          <div class="flex items-start gap-3">
            <Icon name="infoCircle" size="md" class="mt-0.5 flex-shrink-0 text-amber-600 dark:text-amber-300" />
            <div class="text-sm text-amber-800 dark:text-amber-200">
              <p class="font-medium">{{ t('recharge.noticeTitle') }}</p>
              <p class="mt-1 leading-6">{{ t('recharge.noticeDesc') }}</p>
            </div>
          </div>
        </div>

        <div class="h-[calc(100vh-280px)] min-h-[560px] bg-white dark:bg-dark-900">
          <iframe
            :src="shopUrl"
            class="h-full w-full border-0"
            referrerpolicy="no-referrer-when-downgrade"
            :title="t('recharge.iframeTitle')"
          ></iframe>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const shopUrl = 'https://pay.ldxp.cn/shop/5CFQLFES'

const router = useRouter()
const { t } = useI18n()

const goRedeem = () => {
  router.push({ path: '/redeem', query: { from: 'recharge' } })
}
</script>
