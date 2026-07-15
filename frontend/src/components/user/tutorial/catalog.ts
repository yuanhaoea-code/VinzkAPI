import type { TutorialArticle, TutorialCatalog, TutorialContext } from './types'
import { PUBLIC_API_BASE_URL } from '@/constants/site'

const updatedAt = '2026-07-14'

export function normalizeTutorialBaseUrl(value: string, siteUrl: string): string {
  const candidate = value.trim() || siteUrl.trim()
  let resolved = candidate
  if (candidate.startsWith('/')) {
    try {
      resolved = new URL(candidate, siteUrl).toString()
    } catch {
      resolved = candidate
    }
  }
  const normalized = resolved.replace(/\/+$/, '')
  return normalized || PUBLIC_API_BASE_URL
}

export function buildOpenAiBaseUrl(baseUrl: string): string {
  return baseUrl.endsWith('/v1') ? baseUrl : `${baseUrl}/v1`
}

export function createTutorialCatalog(context: TutorialContext): TutorialCatalog {
  const { siteName, siteUrl, apiBaseUrl, openAiBaseUrl, apiKeyPlaceholder } = context
  const providerId = siteName.replace(/[^A-Za-z0-9_]/g, '') || 'OpenAICompatible'

  const articles: TutorialArticle[] = [
    {
      slug: 'quick-start',
      group: 'start',
      tag: 'start',
      icon: 'book',
      title: '快速开始',
      shortTitle: '快速开始',
      eyebrow: 'START HERE',
      description: '按照“准备账户、创建密钥、配置客户端、验证调用”的顺序完成首次接入。',
      duration: '约 5 分钟',
      updatedAt,
      relatedSlugs: ['api-keys', 'codex', 'ccswitch'],
      sections: [
        {
          id: 'site-information',
          title: `认识 ${siteName}`,
          description: '先确认站点地址和客户端需要使用的 API 地址。',
          blocks: [
            { type: 'paragraph', text: `${siteName} 提供统一的 AI 模型调用入口，站点地址为 ${siteUrl}。登录后可以创建 API 密钥、查看使用记录、充值或兑换额度。` },
            {
              type: 'callout',
              tone: 'info',
              title: 'API Base URL',
              text: `${openAiBaseUrl}。Claude Code、Gemini CLI 或站内“使用密钥”弹窗可能显示不带 /v1 的地址，请以对应客户端示例为准。`,
            },
            {
              type: 'links',
              links: [
                { label: '打开仪表盘', to: '/dashboard' },
                { label: '查看模型广场', to: '/model-market' },
              ],
            },
          ],
        },
        {
          id: 'recommended-path',
          title: '推荐接入顺序',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '准备账户', description: '通过自助充值或兑换码获得可用余额，并确认账户状态正常。' },
                { title: '创建 API 密钥', description: '为每个客户端创建独立密钥，便于限额、停用和排查。' },
                { title: '配置客户端', description: '选择 CCSwitch 一键导入、配置文件或环境变量方式。' },
                { title: '验证调用', description: '发送一次测试请求，再到使用记录确认模型、Token 和费用。' },
              ],
            },
          ],
        },
        {
          id: 'reading-guide',
          title: '按你的工具继续阅读',
          blocks: [
            {
              type: 'list',
              items: [
                'Codex 用户：先安装 VSCode 插件或 Codex CLI，再选择 CCSwitch 或手动配置。',
                'Claude Code 用户：先完成客户端安装，再配置 ANTHROPIC_BASE_URL 与密钥。',
                'Gemini CLI 用户：安装 CLI 后配置 GOOGLE_GEMINI_BASE_URL、密钥和模型。',
                'OpenClaw 用户：有经验可手动合并配置；新用户建议借助外部 AI Agent。',
                '出现 401、模型不可用或费用异常时，先查看“常见问题与排查”。',
              ],
            },
            {
              type: 'links',
              links: [
                { label: 'Codex 教程', to: '/tutorials/codex' },
                { label: 'Claude Code 教程', to: '/tutorials/claude-code' },
                { label: 'Gemini CLI 教程', to: '/tutorials/gemini-cli' },
              ],
            },
          ],
        },
        {
          id: 'before-support',
          title: '排查前准备',
          blocks: [
            {
              type: 'callout',
              tone: 'success',
              title: '保留这些信息会更快定位问题',
              text: '记录账号、密钥名称、发生时间、客户端名称、模型名称和完整报错截图。请勿在截图或聊天中公开完整 API Key。',
            },
          ],
        },
      ],
    },
    {
      slug: 'codex',
      group: 'clients',
      tag: 'codex',
      icon: 'terminal',
      title: 'Codex 接入教程',
      shortTitle: 'Codex',
      eyebrow: 'CODEX',
      description: '安装 VSCode 插件或 Codex CLI，并通过 CCSwitch 或配置文件连接到本站。',
      duration: '约 8 分钟',
      updatedAt,
      relatedSlugs: ['ccswitch', 'api-keys', 'troubleshooting'],
      sections: [
        {
          id: 'codex-vscode',
          title: '1. 安装 VSCode 插件',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '打开扩展市场', description: '在 VSCode 左侧活动栏打开“扩展”，搜索 Codex。' },
                { title: '确认发布者', description: '选择 OpenAI 发布的 Codex 插件并安装。' },
                { title: '重新加载窗口', description: '安装完成后按提示重新加载 VSCode。' },
              ],
            },
            {
              type: 'image',
              src: '/tutorials/codex-vscode-extension.png',
              alt: 'VSCode 中的 Codex 官方插件页面',
              caption: '在扩展市场中确认插件名称与 OpenAI 发布者。',
            },
          ],
        },
        {
          id: 'codex-cli',
          title: '2. 安装 Codex CLI',
          blocks: [
            { type: 'paragraph', text: '已安装 Node.js 18 或更高版本时可以跳过环境安装，先在终端验证版本。' },
            { type: 'code', label: '验证 Node.js 与 npm', language: 'bash', code: 'node -v\nnpm -v' },
            { type: 'code', label: '安装 Codex CLI', language: 'bash', code: 'npm install -g @openai/codex' },
            { type: 'code', label: '验证安装', language: 'bash', code: 'codex --version' },
            {
              type: 'callout',
              tone: 'warning',
              title: 'Windows 权限错误',
              text: '如果 npm 更新时出现 EPERM，可关闭正在运行的 Codex/VSCode 进程后重试，必要时以管理员身份打开 PowerShell。',
            },
          ],
        },
        {
          id: 'codex-ccswitch',
          title: '3. 使用 CCSwitch 配置（推荐）',
          blocks: [
            { type: 'paragraph', text: 'CCSwitch 可以维护多个供应商并在 Codex、Claude Code、Gemini CLI 之间切换。本站 API 密钥页已支持一键导入。' },
            {
              type: 'steps',
              items: [
                { title: '安装并启动 CCSwitch', description: '使用官方安装包，首次启动时允许系统注册 ccswitch:// 协议。' },
                { title: '进入 API 密钥', description: '找到要使用的 OpenAI 分组密钥。' },
                { title: '点击“导入到 CCS”', description: '确认系统唤起 CCSwitch 后保存供应商。' },
                { title: '在 CCSwitch 中启用', description: '切换到 Codex，选择刚导入的本站供应商。' },
              ],
            },
            {
              type: 'links',
              links: [
                { label: '打开 API 密钥', to: '/keys' },
                { label: '查看完整 CCSwitch 教程', to: '/tutorials/ccswitch' },
              ],
            },
          ],
        },
        {
          id: 'codex-manual',
          title: '4. 手动配置 config.toml',
          description: '适合不使用 CCSwitch 或需要精确控制模型参数的用户。',
          blocks: [
            { type: 'paragraph', text: 'Windows 配置目录为 %USERPROFILE%\\.codex，macOS/Linux 为 ~/.codex。创建或编辑 config.toml：' },
            {
              type: 'code',
              label: 'config.toml',
              language: 'toml',
              code: `model_provider = "${providerId}"\nmodel = "gpt-5.5"\nreview_model = "gpt-5.5"\nmodel_reasoning_effort = "xhigh"\ndisable_response_storage = true\nnetwork_access = "enabled"\nwindows_wsl_setup_acknowledged = true\n\n[model_providers.${providerId}]\nname = "${siteName}"\nbase_url = "${apiBaseUrl}"\nwire_api = "responses"\nrequires_openai_auth = true\n\n[features]\ngoals = true`,
            },
            { type: 'paragraph', text: '在同一目录创建 auth.json，并填入刚刚创建的 API Key：' },
            {
              type: 'code',
              label: 'auth.json',
              language: 'json',
              code: `{\n  "OPENAI_API_KEY": "${apiKeyPlaceholder}"\n}`,
            },
            {
              type: 'callout',
              tone: 'info',
              title: '模型以模型广场为准',
              text: '示例使用 gpt-5.5。若模型广场显示的可用模型不同，请同时修改 model 与 review_model。',
            },
          ],
        },
        {
          id: 'codex-verify',
          title: '5. 启动并验证',
          blocks: [
            { type: 'code', label: '启动 Codex', language: 'bash', code: 'codex' },
            {
              type: 'steps',
              items: [
                { title: '发送一个简短任务', description: '例如让 Codex 读取当前目录并列出文件。' },
                { title: '查看使用记录', description: '确认请求模型、Token、费用和响应状态。' },
                { title: '失败时重新启动', description: '修改配置文件后必须完全退出并重新打开客户端。' },
              ],
            },
            { type: 'links', links: [{ label: '打开使用记录', to: '/usage' }] },
          ],
        },
      ],
    },
    {
      slug: 'claude-code',
      group: 'clients',
      tag: 'claude',
      icon: 'terminal',
      title: 'Claude Code 接入教程',
      shortTitle: 'Claude Code',
      eyebrow: 'CLAUDE CODE',
      description: '安装 Claude Code，并通过 CCSwitch、settings.json 或环境变量完成接入。',
      duration: '约 7 分钟',
      updatedAt,
      relatedSlugs: ['ccswitch', 'api-keys', 'troubleshooting'],
      sections: [
        {
          id: 'claude-install',
          title: '1. 安装 Claude Code',
          blocks: [
            { type: 'code', label: 'macOS / Linux', language: 'bash', code: 'curl -fsSL https://claude.ai/install.sh | bash' },
            { type: 'code', label: 'Windows PowerShell', language: 'powershell', code: 'irm https://claude.ai/install.ps1 | iex' },
            { type: 'code', label: 'Windows CMD', language: 'cmd', code: 'curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd' },
            { type: 'code', label: '验证安装', language: 'bash', code: 'claude --version' },
          ],
        },
        {
          id: 'claude-ccswitch',
          title: '2. 使用 CCSwitch 配置（推荐）',
          blocks: [
            { type: 'paragraph', text: '在 API 密钥页点击“导入到 CCS”，Anthropic 分组会直接按 Claude Code 导入；部分兼容分组会让你选择 Claude Code 或 Gemini CLI。' },
            { type: 'links', links: [{ label: '前往 API 密钥', to: '/keys' }, { label: 'CCSwitch 详细步骤', to: '/tutorials/ccswitch' }] },
          ],
        },
        {
          id: 'claude-settings',
          title: '3. 使用 settings.json 配置',
          blocks: [
            { type: 'paragraph', text: 'Windows 路径为 %USERPROFILE%\\.claude\\settings.json，macOS/Linux 路径为 ~/.claude/settings.json。' },
            {
              type: 'code',
              label: 'settings.json',
              language: 'json',
              code: `{\n  "env": {\n    "ANTHROPIC_BASE_URL": "${apiBaseUrl}",\n    "ANTHROPIC_AUTH_TOKEN": "${apiKeyPlaceholder}",\n    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",\n    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"\n  }\n}`,
            },
          ],
        },
        {
          id: 'claude-env',
          title: '4. 使用环境变量配置',
          blocks: [
            {
              type: 'code',
              label: 'PowerShell（当前窗口）',
              language: 'powershell',
              code: `$env:ANTHROPIC_BASE_URL="${apiBaseUrl}"\n$env:ANTHROPIC_AUTH_TOKEN="${apiKeyPlaceholder}"\n$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1"\n$env:CLAUDE_CODE_ATTRIBUTION_HEADER="0"\nclaude`,
            },
            {
              type: 'code',
              label: 'macOS / Linux（当前终端）',
              language: 'bash',
              code: `export ANTHROPIC_BASE_URL="${apiBaseUrl}"\nexport ANTHROPIC_AUTH_TOKEN="${apiKeyPlaceholder}"\nexport CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1\nexport CLAUDE_CODE_ATTRIBUTION_HEADER=0\nclaude`,
            },
            {
              type: 'callout',
              tone: 'warning',
              title: '环境变量默认只对当前窗口生效',
              text: '需要长期使用时，优先选择 CCSwitch 或 settings.json；修改后重新启动 Claude Code。',
            },
          ],
        },
        {
          id: 'claude-verify',
          title: '5. 验证连接',
          blocks: [
            { type: 'paragraph', text: '启动 claude 并发送一个短任务。若出现认证错误，检查密钥是否完整、是否属于可用分组，以及地址末尾是否多写了路径。' },
            { type: 'links', links: [{ label: '查看使用记录', to: '/usage' }, { label: '查看常见问题', to: '/tutorials/troubleshooting' }] },
          ],
        },
      ],
    },
    {
      slug: 'gemini-cli',
      group: 'clients',
      tag: 'gemini',
      icon: 'sparkles',
      title: 'Gemini CLI 接入教程',
      shortTitle: 'Gemini CLI',
      eyebrow: 'GEMINI CLI',
      description: '安装 Gemini CLI，并通过 CCSwitch 或环境变量接入本站 Gemini 兼容渠道。',
      duration: '约 6 分钟',
      updatedAt,
      relatedSlugs: ['ccswitch', 'api-keys', 'troubleshooting'],
      sections: [
        {
          id: 'gemini-install',
          title: '1. 安装 Gemini CLI',
          blocks: [
            { type: 'code', label: 'npm 安装', language: 'bash', code: 'npm install -g @google/gemini-cli' },
            { type: 'code', label: '验证安装', language: 'bash', code: 'gemini --version' },
          ],
        },
        {
          id: 'gemini-ccswitch',
          title: '2. 使用 CCSwitch 配置（推荐）',
          blocks: [
            { type: 'paragraph', text: 'Gemini 分组密钥会按 Gemini CLI 直接导入；兼容分组出现客户端选择时请选择 Gemini CLI。' },
            { type: 'links', links: [{ label: '前往 API 密钥', to: '/keys' }, { label: 'CCSwitch 详细步骤', to: '/tutorials/ccswitch' }] },
          ],
        },
        {
          id: 'gemini-env',
          title: '3. 使用环境变量配置',
          blocks: [
            {
              type: 'code',
              label: 'PowerShell',
              language: 'powershell',
              code: `$env:GOOGLE_GEMINI_BASE_URL="${apiBaseUrl}"\n$env:GEMINI_API_KEY="${apiKeyPlaceholder}"\n$env:GEMINI_MODEL="gemini-2.0-flash"\ngemini`,
            },
            {
              type: 'code',
              label: 'macOS / Linux',
              language: 'bash',
              code: `export GOOGLE_GEMINI_BASE_URL="${apiBaseUrl}"\nexport GEMINI_API_KEY="${apiKeyPlaceholder}"\nexport GEMINI_MODEL="gemini-2.0-flash"\ngemini`,
            },
            {
              type: 'callout',
              tone: 'info',
              title: '确认渠道与模型',
              text: '示例模型仅用于说明。请在模型广场确认 Gemini 渠道实际支持的模型名称。',
            },
          ],
        },
        {
          id: 'gemini-verify',
          title: '4. 启动并验证',
          blocks: [
            { type: 'paragraph', text: '运行 gemini 后发送一条简短消息，再到使用记录确认请求已经进入正确分组。' },
            { type: 'links', links: [{ label: '打开模型广场', to: '/model-market' }, { label: '打开使用记录', to: '/usage' }] },
          ],
        },
      ],
    },
    {
      slug: 'ccswitch',
      group: 'clients',
      tag: 'ccswitch',
      icon: 'swap',
      title: 'CCSwitch 配置教程',
      shortTitle: 'CCSwitch',
      eyebrow: 'ONE-CLICK SETUP',
      description: '通过站内深链接一键导入供应商，或在 CCSwitch 中手动添加配置。',
      duration: '约 5 分钟',
      updatedAt,
      relatedSlugs: ['codex', 'claude-code', 'gemini-cli'],
      sections: [
        {
          id: 'ccswitch-about',
          title: '1. CCSwitch 能做什么',
          blocks: [
            { type: 'paragraph', text: 'CCSwitch 是第三方开源客户端配置管理工具，可统一管理 Codex、Claude Code、Gemini CLI、OpenClaw 等工具的供应商配置。' },
            {
              type: 'callout',
              tone: 'warning',
              title: '只从官方渠道下载',
              text: 'CCSwitch 不是本站开发的软件。请核对官方仓库与 ccswitch.io，避免安装来历不明的版本。',
            },
            {
              type: 'links',
              links: [
                { label: 'CCSwitch 官方网站', to: 'https://ccswitch.io', external: true },
                { label: 'GitHub 官方仓库', to: 'https://github.com/farion1231/cc-switch', external: true },
              ],
            },
            {
              type: 'image',
              src: '/tutorials/ccswitch-main-en.png',
              alt: 'CCSwitch 官方主界面',
              caption: 'CCSwitch 官方界面示例，实际版本与语言可能不同。',
            },
          ],
        },
        {
          id: 'ccswitch-one-click',
          title: '2. 从本站一键导入',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '安装并运行 CCSwitch', description: '首次启动后保持应用运行，并允许注册自定义协议。' },
                { title: '创建独立密钥', description: '建议用 codex-main、claude-dev 等用途名称。' },
                { title: '点击“导入到 CCS”', description: '系统会自动携带地址、分组、密钥和用量查询配置。' },
                { title: '确认并启用供应商', description: '在 CCSwitch 检查应用类型与 Endpoint 后保存。' },
              ],
            },
            {
              type: 'image',
              src: '/tutorials/keys-page.png',
              alt: `${siteName} API 密钥页面`,
              caption: '在 API 密钥的操作列中使用“导入到 CCS”；若按钮被管理员隐藏，可使用手动配置。',
            },
            { type: 'links', links: [{ label: '现在创建 API 密钥', to: '/keys' }] },
          ],
        },
        {
          id: 'ccswitch-manual',
          title: '3. 在 CCSwitch 中手动添加',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '选择客户端', description: '根据密钥分组选择 Codex、Claude Code 或 Gemini CLI。' },
                { title: '新增供应商', description: `供应商名称可以填写 ${siteName}。` },
                { title: '填写 Endpoint', description: `OpenAI 兼容客户端使用 ${openAiBaseUrl}；其他客户端按对应教程填写 ${apiBaseUrl}。` },
                { title: '填写 API Key', description: '粘贴完整密钥并保存，再将该供应商设为当前配置。' },
              ],
            },
          ],
        },
        {
          id: 'ccswitch-failure',
          title: '4. 一键导入没有反应',
          blocks: [
            {
              type: 'list',
              items: [
                '确认 CCSwitch 已安装并正在运行。',
                '确认系统允许浏览器打开 ccswitch:// 链接。',
                '浏览器出现外部应用确认框时选择允许。',
                '仍然失败时使用本页的手动添加方式。',
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'api-keys',
      group: 'start',
      tag: 'start',
      icon: 'key',
      title: 'API 密钥使用',
      shortTitle: 'API 密钥',
      eyebrow: 'API KEYS',
      description: '创建、命名、限额、复制和停用密钥，并使用站内生成的客户端配置。',
      duration: '约 4 分钟',
      updatedAt,
      relatedSlugs: ['quick-start', 'ccswitch', 'troubleshooting'],
      sections: [
        {
          id: 'key-create',
          title: '1. 创建密钥',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '打开 API 密钥', description: '在左侧导航进入 API 密钥页面。' },
                { title: '点击“创建密钥”', description: '填写名称并选择与你要接入的客户端相匹配的分组。' },
                { title: '设置限制', description: '按需设置额度、速率、到期时间和 IP 规则。' },
                { title: '保存完整密钥', description: '创建后立即复制并保存在安全的密码管理器中。' },
              ],
            },
            {
              type: 'image',
              src: '/tutorials/keys-page.png',
              alt: `${siteName} API 密钥列表`,
              caption: '密钥列表支持使用密钥、导入 CCSwitch、停用、编辑和删除。',
            },
            { type: 'links', links: [{ label: '打开 API 密钥', to: '/keys' }] },
          ],
        },
        {
          id: 'key-use',
          title: '2. 使用密钥',
          blocks: [
            { type: 'paragraph', text: '点击密钥行中的“使用密钥”，站内会根据分组平台生成 Claude Code、Gemini CLI 或 Codex 的配置。复制前先确认操作系统标签。' },
            {
              type: 'callout',
              tone: 'success',
              title: '优先使用站内生成配置',
              text: `站内弹窗会使用当前配置的 API 地址 ${apiBaseUrl}，比手动输入更不容易遗漏字段。`,
            },
          ],
        },
        {
          id: 'key-operations',
          title: '3. 运维建议',
          blocks: [
            {
              type: 'list',
              items: [
                '按用途命名，例如 codex-main、claude-dev、openclaw-home。',
                '不要在多个设备和客户端中共用同一把密钥。',
                '密钥疑似泄漏时立即停用或删除，再创建新密钥。',
                '为自动化任务设置额度与到期时间，避免失控消耗。',
                '通过使用记录按密钥筛选，定位异常请求。',
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'recharge',
      group: 'start',
      tag: 'start',
      icon: 'gift',
      title: '充值与兑换',
      shortTitle: '充值与兑换',
      eyebrow: 'BALANCE',
      description: '通过站内自助充值或兑换码增加余额，并确认账户额度已经更新。',
      duration: '约 3 分钟',
      updatedAt,
      relatedSlugs: ['quick-start', 'api-keys', 'troubleshooting'],
      sections: [
        {
          id: 'direct-recharge',
          title: '1. 自助充值',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '打开自助充值', description: '选择可用的充值方式和金额。' },
                { title: '完成支付', description: '支付结果返回后不要重复提交同一订单。' },
                { title: '检查余额', description: '回到仪表盘确认余额和账户状态。' },
              ],
            },
            { type: 'links', links: [{ label: '打开自助充值', to: '/recharge' }] },
          ],
        },
        {
          id: 'redeem-code',
          title: '2. 使用兑换码',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '进入兑换中心', description: '在左侧 SERVICES 分组打开兑换中心。' },
                { title: '输入兑换码', description: '注意区分大小写，不要保留首尾空格。' },
                { title: '确认兑换结果', description: '查看余额、并发数或订阅是否发生变化。' },
              ],
            },
            {
              type: 'image',
              src: '/tutorials/redeem-page.png',
              alt: `${siteName} 兑换中心`,
              caption: '兑换成功后，页面会更新本次兑换带来的余额或权益。',
            },
            { type: 'links', links: [{ label: '打开兑换中心', to: '/redeem' }] },
          ],
        },
        {
          id: 'balance-not-updated',
          title: '3. 余额没有更新',
          blocks: [
            {
              type: 'list',
              items: [
                '刷新页面，再重新打开仪表盘。',
                '检查兑换是否增加的是订阅、并发或其他权益，而不是账户余额。',
                '检查支付订单状态，避免重复支付。',
                '保留订单号、兑换时间和结果截图后再联系管理员。',
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'openclaw',
      group: 'tools',
      tag: 'openclaw',
      icon: 'server',
      title: 'OpenClaw 接入',
      shortTitle: 'OpenClaw',
      eyebrow: 'OPENCLAW',
      description: '手动合并 OpenClaw 配置，或借助外部 AI Agent 完成配置与工具调用测试。',
      duration: '约 10 分钟',
      updatedAt,
      relatedSlugs: ['api-keys', 'codex', 'troubleshooting'],
      sections: [
        {
          id: 'openclaw-manual',
          title: '1. 手动配置',
          blocks: [
            {
              type: 'callout',
              tone: 'warning',
              title: '合并配置，不要整份覆盖',
              text: '下面只是参考片段。请先备份 openclaw.json，再把 providers 和默认模型相关字段合并到现有文件。模型参数以当前 OpenClaw 版本和模型广场为准。',
            },
            {
              type: 'code',
              label: 'openclaw.json 参考片段',
              language: 'json',
              code: `{\n  "models": {\n    "mode": "merge",\n    "providers": {\n      "openai": {\n        "baseUrl": "${apiBaseUrl}",\n        "apiKey": "${apiKeyPlaceholder}",\n        "api": "openai-responses",\n        "models": [\n          {\n            "id": "gpt-5.5",\n            "name": "GPT 5.5",\n            "reasoning": true,\n            "input": ["text", "image"]\n          }\n        ],\n        "authHeader": true\n      }\n    }\n  },\n  "agents": {\n    "defaults": {\n      "model": { "primary": "openai/gpt-5.5" },\n      "thinkingDefault": "xhigh",\n      "timeoutSeconds": 1800\n    }\n  }\n}`,
            },
          ],
        },
        {
          id: 'openclaw-agent',
          title: '2. 借助外部 AI Agent 配置',
          blocks: [
            { type: 'paragraph', text: 'OpenClaw 的字段会随版本更新。新用户可以使用 Codex、Claude Code、Cursor 等外部 Agent 阅读当前 OpenClaw 文档并修改配置。' },
            {
              type: 'callout',
              tone: 'warning',
              title: '不要让 OpenClaw 修改自己的配置',
              text: '使用独立的外部 Agent 完成修改，避免配置错误后 OpenClaw 无法启动。执行前要求 Agent 先备份文件。',
            },
            {
              type: 'code',
              label: '给外部 Agent 的任务模板',
              language: 'text',
              code: `请先备份我的 openclaw.json，再依据当前 OpenClaw 官方文档，把 ${siteName} 作为 OpenAI Responses 兼容供应商接入。\nBase URL: ${apiBaseUrl}\nAPI Key: ${apiKeyPlaceholder}\n模型示例: gpt-5.5\n完成后进行三轮工具调用测试；测试通过后再更新其他频道配置。不要覆盖现有无关配置。`,
            },
          ],
        },
        {
          id: 'openclaw-verify',
          title: '3. 验证与回滚',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '校验 JSON', description: '确认文件格式有效，没有重复键或漏写逗号。' },
                { title: '重启并测试', description: '先做普通对话，再进行三轮工具调用。' },
                { title: '检查使用记录', description: '确认模型、Token 和响应状态符合预期。' },
                { title: '失败时回滚', description: '恢复备份文件，再逐项检查 provider、api 和模型名称。' },
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'troubleshooting',
      group: 'ops',
      tag: 'ops',
      icon: 'exclamationTriangle',
      title: '常见问题与日常排查',
      shortTitle: '问题排查',
      eyebrow: 'TROUBLESHOOTING',
      description: '处理认证、模型、余额、限流和高耗用问题，并建立日常检查习惯。',
      duration: '约 8 分钟',
      updatedAt,
      relatedSlugs: ['api-keys', 'recharge', 'quick-start'],
      sections: [
        {
          id: 'common-errors',
          title: '1. 常见错误',
          blocks: [
            {
              type: 'list',
              items: [
                '401 / 未授权：检查 API Key 是否完整、是否已停用，以及客户端是否读取了新配置。',
                '404 或返回 HTML：通常是 Base URL 路径错误；OpenAI 兼容地址优先使用带 /v1 的地址。',
                '模型不存在：以模型广场和密钥分组支持的模型为准，修改后重启客户端。',
                '429 / 限流：检查密钥 RPM、并发、额度和订阅当日限制。',
                '兑换后额度不变：刷新页面并确认兑换内容是余额、订阅还是并发权益。',
              ],
            },
            {
              type: 'callout',
              tone: 'info',
              title: '先缩小问题范围',
              text: '用一把新密钥、一个短请求和一个基础模型复现。确认基础调用成功后，再恢复长上下文、工具调用或高推理强度。',
            },
          ],
        },
        {
          id: 'daily-dashboard',
          title: '2. 仪表盘每日检查',
          blocks: [
            {
              type: 'list',
              items: ['账户余额', '今日请求与今日消费', '今日与累计 Token', '平均响应时间', '平台分布与最近使用记录'],
            },
            {
              type: 'image',
              src: '/tutorials/dashboard-page.png',
              alt: `${siteName} 仪表盘`,
              caption: '仪表盘适合快速判断余额、请求量、Token 和响应时间是否异常。',
            },
            { type: 'links', links: [{ label: '打开仪表盘', to: '/dashboard' }] },
          ],
        },
        {
          id: 'usage-analysis',
          title: '3. 使用记录排查',
          blocks: [
            { type: 'paragraph', text: '在使用记录按时间、密钥、模型或分组筛选，重点检查异常状态、超长响应和突然增加的 Token。' },
            {
              type: 'image',
              src: '/tutorials/usage-page.png',
              alt: `${siteName} 使用记录`,
              caption: '使用记录可用于定位某把密钥或某个模型的异常消耗。',
            },
            { type: 'links', links: [{ label: '打开使用记录', to: '/usage' }] },
          ],
        },
        {
          id: 'high-usage',
          title: '4. 高耗用排查',
          blocks: [
            {
              type: 'list',
              items: [
                '是否有脚本或 Agent 循环请求。',
                '是否把整个仓库或超长历史反复发送给模型。',
                '是否误用高成本模型或 xhigh 推理强度。',
                '是否有密钥泄漏并在未知设备上使用。',
                '是否启用了过多并行 Agent 或子任务。',
              ],
            },
          ],
        },
        {
          id: 'support-checklist',
          title: '5. 提交问题前的检查清单',
          blocks: [
            {
              type: 'steps',
              items: [
                { title: '记录发生时间', description: '尽量精确到分钟，便于匹配后台日志。' },
                { title: '保存完整错误', description: '截图需要包含错误码、客户端和模型名称。' },
                { title: '提供密钥名称', description: '只提供名称或末四位，不要发送完整密钥。' },
                { title: '说明复现步骤', description: '写清执行了什么命令、使用什么配置以及预期结果。' },
              ],
            },
          ],
        },
      ],
    },
  ]

  return {
    articles,
    groups: [
      { id: 'start', label: '开始使用', articleSlugs: ['quick-start', 'api-keys', 'recharge'] },
      { id: 'clients', label: '客户端接入', articleSlugs: ['codex', 'claude-code', 'gemini-cli', 'ccswitch'] },
      { id: 'tools', label: '更多工具', articleSlugs: ['openclaw'] },
      { id: 'ops', label: '排查与运维', articleSlugs: ['troubleshooting'] },
    ],
  }
}
