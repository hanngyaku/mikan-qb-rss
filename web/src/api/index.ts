import type { components } from '../types/openapi'

export type Settings = components['schemas']['model.SettingsResponse']
export type UpdateSettings = components['schemas']['model.UpdateSettingsRequest']
export type Subscription = components['schemas']['model.Subscription']
export type CreateSubscription = components['schemas']['model.CreateSubscriptionRequest']
export type UpdateSubscription = components['schemas']['model.UpdateSubscriptionRequest']
export type QBTest = components['schemas']['model.QBTestResponse']
export type LogResponse = components['schemas']['model.LogResponse']
export type UpdateBroadcastDay = components['schemas']['model.UpdateBroadcastDayRequest']

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  const body = response.status === 204 ? undefined : await response.json()
  if (!response.ok) throw new Error(body.error || `HTTP ${response.status}`)
  return body as T
}

export const api = {
  getSettings: () => request<Settings>('/settings'),
  updateSettings: (body: UpdateSettings) => request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(body) }),
  testQB: () => request<QBTest>('/qb/test', { method: 'POST' }),
  listSubscriptions: () => request<Subscription[]>('/subscriptions'),
  createSubscription: (body: CreateSubscription) => request<Subscription>('/subscriptions', { method: 'POST', body: JSON.stringify(body) }),
  updateSubscription: (id: number, body: UpdateSubscription) => request<Subscription>(`/subscriptions/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteSubscription: (id: number) => request<void>(`/subscriptions/${id}`, { method: 'DELETE' }),
  syncSubscription: (id: number) => request<Subscription>(`/subscriptions/${id}/sync`, { method: 'POST' }),
  getLogs: (lines: number) => request<LogResponse>(`/logs?lines=${lines}`),
  updateBroadcastDay: (id: number, day: string) => request<Subscription>(`/subscriptions/${id}/broadcast-day`, { method: 'PUT', body: JSON.stringify({ day } satisfies UpdateBroadcastDay) }),
}
