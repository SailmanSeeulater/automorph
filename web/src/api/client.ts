import createClient from 'openapi-fetch'

import type { components, paths } from './schema'

// Same origin as the page: Vite proxies /api in dev, Go serves it in production.
export const api = createClient<paths>({ baseUrl: '' })

export type User = components['schemas']['UserDTO']
export type PostSummary = components['schemas']['PostSummary']
export type PublicPost = components['schemas']['PublicPost']
export type AdminPost = components['schemas']['AdminPost']
export type Image = components['schemas']['ImageDTO']

// The error body huma sends for every 4xx/5xx.
type Problem = components['schemas']['ErrorModel']

export class ApiError extends Error {
  status: number

  constructor(status: number, problem?: Problem) {
    // Validation errors carry the useful part in errors[], e.g. "body.slug: expected string to match pattern".
    const details = problem?.errors?.map((e) => `${e.location}: ${e.message}`).join('; ')
    super(details || problem?.detail || `Request failed (${status})`)
    this.status = status
  }
}

// openapi-fetch can't build multipart bodies from the schema, so uploads use fetch directly.
export async function uploadImage(file: File, alt = ''): Promise<Image> {
  const form = new FormData()
  form.append('file', file)
  form.append('alt', alt)
  const res = await fetch('/api/admin/media', { method: 'POST', body: form })
  const body = await res.json()
  if (!res.ok) throw new ApiError(res.status, body)
  return body as Image
}
