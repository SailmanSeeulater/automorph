import { queryOptions } from '@tanstack/react-query'
import { notFound } from '@tanstack/react-router'

import { api, ApiError } from './client'

// Query options are shared by route loaders (fetch before render) and components (read the cache).

// ---- Public ----

export const postsQuery = () =>
  queryOptions({
    queryKey: ['posts'],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/api/posts', {
        params: { query: { limit: 50 } },
      })
      if (error) throw new ApiError(response.status, error)
      return data
    },
  })

export const postQuery = (slug: string) =>
  queryOptions({
    queryKey: ['posts', slug],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/api/posts/{slug}', {
        params: { path: { slug } },
      })
      if (response.status === 404) throw notFound()
      if (error) throw new ApiError(response.status, error)
      return data
    },
  })

// ---- Auth ----

// The logged-in user, or null. A 401 is an answer here, not an error.
export const meQuery = () =>
  queryOptions({
    queryKey: ['me'],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/api/me')
      if (response.status === 401) return null
      if (error) throw new ApiError(response.status, error)
      return data
    },
    staleTime: 5 * 60_000,
  })

// ---- Admin ----

export const adminPostsQuery = () =>
  queryOptions({
    queryKey: ['admin', 'posts'],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/api/admin/posts')
      if (error) throw new ApiError(response.status, error)
      return data
    },
    staleTime: 0,
  })

export const adminPostQuery = (id: number) =>
  queryOptions({
    queryKey: ['admin', 'posts', id],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/api/admin/posts/{id}', {
        params: { path: { id } },
      })
      if (response.status === 404) throw notFound()
      if (error) throw new ApiError(response.status, error)
      return data
    },
    staleTime: 0,
  })

export const imagesQuery = () =>
  queryOptions({
    queryKey: ['admin', 'media'],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/api/admin/media')
      if (error) throw new ApiError(response.status, error)
      return data
    },
  })
