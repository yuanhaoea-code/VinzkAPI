import { apiClient } from './client'

export interface ModelMarketTokenPrice {
  input: number
  output: number
  cache_read: number
}

export interface ModelMarketGroupPrice {
  group_id: number
  group_name: string
  platform: string
  rate: number
  token_price?: ModelMarketTokenPrice
  image_prices?: Partial<Record<'1K' | '2K' | '4K', number>>
}

export interface ModelMarketModel {
  id: string
  name: string
  provider: string
  provider_label: string
  billing: 'token' | 'request'
  official_price?: ModelMarketTokenPrice
  groups: ModelMarketGroupPrice[]
  endpoint_types: string[]
  tags: string[]
}

export interface ModelMarketCatalog {
  models: ModelMarketModel[]
}

export async function getModelMarketCatalog(): Promise<ModelMarketCatalog> {
  const { data } = await apiClient.get<ModelMarketCatalog>('/model-market/catalog')
  return data
}

export const modelMarketAPI = {
  getCatalog: getModelMarketCatalog,
}

export default modelMarketAPI
