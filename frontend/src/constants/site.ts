export const PUBLIC_SITE_NAME = 'VinzkAPI'
export const PUBLIC_SITE_SUBTITLE = 'AI API Gateway Platform'
export const PUBLIC_SITE_URL = 'https://vinzk.cn'
export const PUBLIC_API_BASE_URL = 'https://api.vinzk.cn'
export const PUBLIC_TUTORIAL_URL = `${PUBLIC_SITE_URL}/tutorials`
export const PUBLIC_RECHARGE_URL = `${PUBLIC_SITE_URL}/recharge`
export const PUBLIC_CONTACT_INFO = '客服微信：13387544600'

function isNonProductionHostname(hostname: string): boolean {
  const normalized = hostname.toLowerCase().replace(/^\[|\]$/g, '')
  if (
    normalized === 'localhost' ||
    normalized === '0.0.0.0' ||
    normalized === '::1' ||
    normalized.endsWith('.localhost') ||
    normalized.endsWith('.local') ||
    normalized.endsWith('.test') ||
    normalized.endsWith('.invalid') ||
    normalized === 'example.com' ||
    normalized.endsWith('.example.com') ||
    normalized.includes('your-domain') ||
    normalized.includes('your-site')
  ) {
    return true
  }

  const octets = normalized.split('.').map(Number)
  if (octets.length !== 4 || octets.some((octet) => !Number.isInteger(octet) || octet < 0 || octet > 255)) {
    return false
  }

  return (
    octets[0] === 0 ||
    octets[0] === 10 ||
    octets[0] === 127 ||
    (octets[0] === 169 && octets[1] === 254) ||
    (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31) ||
    (octets[0] === 192 && octets[1] === 168)
  )
}

export function resolvePublicUrl(value: string | null | undefined, fallback: string): string {
  const candidate = value?.trim()
  if (!candidate) return fallback

  try {
    const parsed = new URL(candidate)
    if (parsed.protocol !== 'https:' || isNonProductionHostname(parsed.hostname)) {
      return fallback
    }
    return candidate.replace(/\/+$/, '')
  } catch {
    return fallback
  }
}
