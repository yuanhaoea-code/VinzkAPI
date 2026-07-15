<template>
  <AppLayout>
    <div
      data-testid="profile-shell"
      class="profile-page"
    >
      <header class="profile-header">
        <p class="console-kicker">PROFILE</p>
        <h1>{{ t('nav.profile') }}</h1>
        <p>
          直接展示账户身份、联系方式、安全设置和通知设置。页面不放大宣传语，重点让用户快速核对资料并完成密码、TOTP、余额提醒等现有操作。
        </p>
      </header>

      <div class="profile-grid">
        <main class="profile-main">
          <ProfileInfoCard
            :user="user"
            :linuxdo-enabled="linuxdoOAuthEnabled"
            :dingtalk-enabled="dingtalkOAuthEnabled"
            :oidc-enabled="oidcOAuthEnabled"
            :oidc-provider-name="oidcOAuthProviderName"
            :wechat-enabled="wechatOAuthEnabled"
            :wechat-open-enabled="wechatOAuthOpenEnabled"
            :wechat-mp-enabled="wechatOAuthMPEnabled"
            hide-auth-bindings
          />

          <section class="profile-panel profile-password-panel">
            <div class="panel-heading">
              <p class="console-kicker">SECURITY</p>
              <h2>{{ t('profile.changePassword') }}</h2>
            </div>
            <ProfilePasswordForm embedded />
          </section>

          <section class="profile-panel profile-auth-panel">
            <div class="panel-heading">
              <p class="console-kicker">OAUTH</p>
              <h2>{{ t('profile.authBindings.title') }}</h2>
              <p>{{ t('profile.authBindings.description') }}</p>
            </div>
            <ProfileIdentityBindingsSection
              :user="user"
              :linuxdo-enabled="linuxdoOAuthEnabled"
              :dingtalk-enabled="dingtalkOAuthEnabled"
              :oidc-enabled="oidcOAuthEnabled"
              :oidc-provider-name="oidcOAuthProviderName"
              :wechat-enabled="wechatOAuthEnabled"
              :wechat-open-enabled="wechatOAuthOpenEnabled"
              :wechat-mp-enabled="wechatOAuthMPEnabled"
              embedded
              compact
              hide-header
            />
          </section>
        </main>

        <aside class="profile-side">
          <section
            v-if="contactInfo"
            class="profile-panel support-panel"
          >
            <p class="console-kicker">SUPPORT</p>
            <h2>{{ t('common.contactSupport') }}</h2>
            <p>遇到充值、密码、调用异常时使用当前站点配置的联系方式。</p>
            <div class="support-line">
              <Icon name="chat" size="sm" />
              <span>{{ contactInfo }}</span>
            </div>
          </section>

          <div class="profile-panel balance-panel-wrap">
            <ProfileBalanceNotifyCard
              v-if="user && balanceLowNotifyEnabled"
              :enabled="user.balance_notify_enabled ?? true"
              :threshold="user.balance_notify_threshold"
              :extra-emails="user.balance_notify_extra_emails ?? []"
              :system-default-threshold="systemDefaultThreshold"
              :user-email="user.email"
            />
            <div v-else class="panel-empty">
              <p class="console-kicker">BALANCE NOTICE</p>
              <h2>{{ t('profile.balanceNotify.title') }}</h2>
              <p>当前站点未开启余额提醒配置。</p>
            </div>
          </div>

          <div class="profile-panel totp-panel-wrap">
            <ProfileTotpCard />
          </div>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { PUBLIC_CONTACT_INFO } from '@/constants/site'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const contactInfo = ref(PUBLIC_CONTACT_INFO)
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const dingtalkOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) {
        return
      }
      contactInfo.value = settings.contact_info || PUBLIC_CONTACT_INFO
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
</script>

<style scoped>
.profile-page {
  --profile-panel: rgba(255, 255, 255, 0.86);
  --profile-ink: #171513;
  --profile-muted: #70675e;
  --profile-line: rgba(23, 21, 19, 0.1);
  --profile-line-strong: rgba(23, 21, 19, 0.16);
  max-width: 1560px;
  margin: 0 auto;
  color: var(--profile-ink);
}

.profile-header {
  margin: 0 0 22px;
  max-width: 980px;
}

.console-kicker {
  margin: 0 0 8px;
  color: #8c7351;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.24em;
}

.profile-header h1,
.panel-heading h2,
.support-panel h2,
.panel-empty h2 {
  margin: 0;
  font-family: var(--console-font-sans);
  font-weight: 700;
  letter-spacing: 0;
}

.profile-header h1 {
  font-size: 34px;
  line-height: 1.1;
}

.profile-header p,
.panel-heading p,
.support-panel p,
.panel-empty p {
  margin: 10px 0 0;
  color: var(--profile-muted);
  font-size: 14px;
  line-height: 1.85;
}

.profile-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(360px, 0.92fr);
  gap: 16px;
  align-items: start;
}

.profile-main,
.profile-side {
  display: grid;
  gap: 16px;
}

.profile-panel {
  border: 1px solid var(--profile-line);
  border-radius: 24px;
  background: var(--profile-panel);
  box-shadow: 0 1px 0 rgba(44, 34, 24, 0.03);
}

.profile-password-panel,
.profile-auth-panel,
.support-panel {
  padding: 18px;
}

.panel-heading {
  margin-bottom: 16px;
}

.panel-heading h2,
.support-panel h2,
.panel-empty h2 {
  font-size: 24px;
  line-height: 1.2;
}

.support-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 18px;
  border: 1px solid var(--profile-line);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.7);
  padding: 11px 14px;
  color: #174b43;
  font-size: 13px;
  font-weight: 800;
}

.balance-panel-wrap,
.totp-panel-wrap {
  overflow: hidden;
}

.panel-empty {
  padding: 18px;
}

:deep(.card) {
  border-color: var(--profile-line);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: none;
}

:deep(.input) {
  min-height: 42px;
  border-color: var(--profile-line-strong);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
}

:deep(.btn) {
  border-radius: 999px;
  font-weight: 800;
}

:deep(.btn-primary) {
  border-color: #171513;
  background: #171513;
  color: #f7f3eb;
}

:deep(.btn-secondary) {
  border-color: var(--profile-line);
  background: rgba(255, 255, 255, 0.72);
  color: #171513;
}

:deep(.badge) {
  border-radius: 999px;
}

:global(.dark) .profile-page {
  --profile-panel: rgba(28, 26, 24, 0.9);
  --profile-ink: #f5efe6;
  --profile-muted: #b7aa9c;
  --profile-line: rgba(255, 255, 255, 0.1);
  --profile-line-strong: rgba(255, 255, 255, 0.16);
}

:global(.dark) .support-line,
:global(.dark) :deep(.input),
:global(.dark) :deep(.btn-secondary) {
  background: rgba(255, 255, 255, 0.06);
  color: var(--profile-ink);
}

@media (max-width: 1180px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .profile-header h1 {
    font-size: 30px;
  }

  .profile-password-panel,
  .profile-auth-panel,
  .support-panel,
  .panel-empty {
    padding: 16px;
  }
}
</style>
