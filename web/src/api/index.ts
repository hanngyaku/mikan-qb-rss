import type { definitions } from '../types/openapi'

export type Settings = definitions['model.SettingsResponse']
export type UpdateSettings = definitions['model.UpdateSettingsRequest']
export type Subscription = definitions['model.Subscription']
export type CreateSubscription = definitions['model.CreateSubscriptionRequest']
export type QBTest = definitions['model.QBTestResponse']

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  const body = await response.json()
  if (!response.ok) throw new Error(body.error || `HTTP ${response.status}`)
  return body as T
}

export const api = {
  getSettings: () => request<Settings>('/settings'),
  updateSettings: (body: UpdateSettings) => request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(body) }),
  testQB: () => request<QBTest>('/qb/test', { method: 'POST' }),
  listSubscriptions: () => request<Subscription[]>('/subscriptions'),
  createSubscription: (body: CreateSubscription) => request<Subscription>('/subscriptions', { method: 'POST', body: JSON.stringify(body) }),
}
