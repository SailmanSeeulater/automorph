import { keepPreviousData, useMutation, useQuery, useQueryClient, useSuspenseQuery } from '@tanstack/react-query'
import { getRouteApi, Link, useBlocker, useNavigate } from '@tanstack/react-router'
import { useEffect, useRef, useState, type ClipboardEvent, type DragEvent } from 'react'

import { api, ApiError, uploadImage, type AdminPost, type Image } from '../api/client'
import { adminPostQuery, adminPostsQuery, imagesQuery } from '../api/queries'
import { formatDate, postCode } from '../lib/format'
import { useDebounced } from '../lib/useDebounced'
import prose from '../styles/prose.module.css'
import ui from './admin.module.css'
import styles from './Editor.module.css'
import { ImagePanel } from './ImagePanel'

type Fields = {
  title: string
  slug: string
  excerpt: string
  bodyMd: string
  coverMediaId?: number
  status: 'draft' | 'published'
}

function fieldsOf(post?: AdminPost): Fields {
  if (!post) return { title: '', slug: '', excerpt: '', bodyMd: '', status: 'draft' }
  return {
    title: post.title,
    slug: post.slug,
    excerpt: post.excerpt,
    bodyMd: post.bodyMd,
    coverMediaId: post.coverMediaId,
    status: post.status,
  }
}

const sameFields = (a: Fields, b: Fields) => JSON.stringify(a) === JSON.stringify(b)

const editRoute = getRouteApi('/admin/posts/$id')

export function EditPostPage() {
  const { id } = editRoute.useParams()
  const { data } = useSuspenseQuery(adminPostQuery(id))
  // key: switching posts starts a fresh editor instead of reusing the old one's state.
  return <Editor key={id} post={data} />
}

export function NewPostPage() {
  return <Editor key="new" />
}

function Editor({ post }: { post?: AdminPost }) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const bodyRef = useRef<HTMLTextAreaElement>(null)

  const [saved, setSaved] = useState(() => fieldsOf(post))
  const [fields, setFields] = useState(saved)
  const dirty = !sameFields(fields, saved)

  function set<K extends keyof Fields>(key: K, value: Fields[K]) {
    setFields((f) => ({ ...f, [key]: value }))
  }

  // ---- Leaving with unsaved changes ----

  // A ref, so a navigation right after saving doesn't see the stale "dirty" from the last render.
  const dirtyRef = useRef(dirty)
  useEffect(() => {
    dirtyRef.current = dirty
  }, [dirty])
  useBlocker({
    shouldBlockFn: () => dirtyRef.current && !window.confirm('Discard unsaved changes?'),
    enableBeforeUnload: () => dirtyRef.current,
  })

  // ---- Save, publish, delete ----

  const save = useMutation({
    mutationFn: async (next: Fields) => {
      if (post) {
        const { data, error, response } = await api.PUT('/api/admin/posts/{id}', {
          params: { path: { id: post.id } },
          body: next,
        })
        if (error) throw new ApiError(response.status, error)
        return data
      }
      const { data, error, response } = await api.POST('/api/admin/posts', {
        body: { ...next, slug: next.slug || undefined }, // empty slug = generate from title
      })
      if (error) throw new ApiError(response.status, error)
      return data
    },
    onSuccess: (p, sent) => {
      const fresh = fieldsOf(p)
      setSaved(fresh)
      // Keep anything typed while the save was in flight.
      setFields((cur) => (sameFields(cur, sent) ? fresh : cur))
      queryClient.setQueryData(adminPostQuery(p.id).queryKey, p)
      queryClient.invalidateQueries({ queryKey: adminPostsQuery().queryKey, exact: true })
      queryClient.invalidateQueries({ queryKey: ['posts'] }) // public pages
      if (!post) {
        dirtyRef.current = false
        navigate({ to: '/admin/posts/$id', params: { id: p.id }, replace: true })
      }
    },
  })

  const remove = useMutation({
    mutationFn: async (id: number) => {
      const { error, response } = await api.DELETE('/api/admin/posts/{id}', { params: { path: { id } } })
      if (error) throw new ApiError(response.status, error)
    },
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: adminPostQuery(id).queryKey })
      queryClient.invalidateQueries({ queryKey: adminPostsQuery().queryKey, exact: true })
      queryClient.invalidateQueries({ queryKey: ['posts'] })
      dirtyRef.current = false
      navigate({ to: '/admin' })
    },
  })

  function saveAs(status: Fields['status']) {
    save.mutate({ ...fields, status })
  }

  // Cmd+S / Ctrl+S saves.
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        if (!save.isPending) save.mutate(fields)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [fields, save])

  // ---- Images ----

  // Puts text at the cursor in the body, on its own paragraph.
  function insertBlock(text: string) {
    const el = bodyRef.current
    setFields((f) => {
      const start = el?.selectionStart ?? f.bodyMd.length
      const end = el?.selectionEnd ?? start
      const before = f.bodyMd.slice(0, start)
      const after = f.bodyMd.slice(end)
      const lead = before === '' || before.endsWith('\n\n') ? '' : before.endsWith('\n') ? '\n' : '\n\n'
      const trail = after === '' || after.startsWith('\n\n') ? '' : after.startsWith('\n') ? '\n' : '\n\n'
      return { ...f, bodyMd: before + lead + text + trail + after }
    })
  }

  const upload = useMutation({
    mutationFn: async (files: File[]) => {
      const images: Image[] = []
      for (const file of files) images.push(await uploadImage(file)) // one at a time, like the server
      return images
    },
    onSuccess: (images) => {
      queryClient.invalidateQueries({ queryKey: imagesQuery().queryKey })
      insertBlock(images.map((img) => img.markdown).join('\n\n'))
    },
  })

  function uploadFiles(files: File[]) {
    const images = files.filter((f) => f.type.startsWith('image/'))
    if (images.length > 0) upload.mutate(images)
    return images.length > 0
  }

  function onDrop(e: DragEvent<HTMLTextAreaElement>) {
    if (uploadFiles([...e.dataTransfer.files])) e.preventDefault()
  }

  function onPaste(e: ClipboardEvent<HTMLTextAreaElement>) {
    if (uploadFiles([...e.clipboardData.files])) e.preventDefault()
  }

  // ---- Preview: rendered by the server, so it matches the published post exactly ----

  const previewSource = useDebounced(fields.bodyMd, 300)
  const preview = useQuery({
    queryKey: ['admin', 'preview', previewSource],
    queryFn: async () => {
      const { data, error, response } = await api.POST('/api/admin/preview', { body: { bodyMd: previewSource } })
      if (error) throw new ApiError(response.status, error)
      return data.html
    },
    placeholderData: keepPreviousData,
    staleTime: Infinity,
  })

  // The cover as currently chosen, saved or not.
  const images = useQuery(imagesQuery())
  const cover =
    fields.coverMediaId === undefined
      ? undefined
      : (images.data?.items.find((img) => img.id === fields.coverMediaId) ??
        (post?.cover?.id === fields.coverMediaId ? post.cover : undefined))

  // ---- Render ----

  const status = save.isPending ? 'Saving…' : dirty ? 'Unsaved changes' : post ? 'Saved' : 'New post'

  return (
    <div className={styles.editor}>
      <div className={styles.toolbar}>
        <div className={styles.toolbarInfo}>
          <Link to="/admin">← Posts</Link>
          {post && <span className={ui.muted}>{postCode(post.id)}</span>}
          <span className={ui.muted}>
            {saved.status === 'published' ? `Published ${formatDate(post?.publishedAt)}` : 'Draft'}
          </span>
          <span className={dirty ? undefined : ui.muted}>{status}</span>
        </div>
        <div className={styles.toolbarActions}>
          {post && saved.status === 'published' && (
            <Link to="/blog/$slug" params={{ slug: saved.slug }}>
              View
            </Link>
          )}
          {post && (
            <button
              type="button"
              className={ui.danger}
              disabled={remove.isPending}
              onClick={() => window.confirm(`Delete “${saved.title}”? This can't be undone.`) && remove.mutate(post.id)}
            >
              Delete
            </button>
          )}
          {saved.status === 'published' ? (
            <button type="button" disabled={save.isPending} onClick={() => saveAs('draft')}>
              Unpublish
            </button>
          ) : (
            <button type="button" disabled={save.isPending || !fields.title} onClick={() => saveAs('published')}>
              Publish
            </button>
          )}
          <button
            type="button"
            className={ui.button}
            disabled={save.isPending || !dirty || !fields.title}
            onClick={() => save.mutate(fields)}
          >
            Save
          </button>
        </div>
      </div>

      {(save.error || remove.error || upload.error) && (
        <p className={ui.error}>{(save.error ?? remove.error ?? upload.error)?.message}</p>
      )}

      <div className={styles.meta}>
        <label className={`${ui.field} ${styles.titleField}`}>
          <span className={ui.label}>Title</span>
          <input
            className={`${ui.input} ${styles.titleInput}`}
            value={fields.title}
            onChange={(e) => set('title', e.target.value)}
            placeholder="Untitled"
            autoFocus={!post}
          />
        </label>
        <label className={ui.field}>
          <span className={ui.label}>Slug {post ? '(changing it breaks old links)' : '(optional)'}</span>
          <input
            className={ui.input}
            value={fields.slug}
            onChange={(e) => set('slug', e.target.value)}
            placeholder={post ? '' : 'made from the title'}
          />
        </label>
        <label className={ui.field}>
          <span className={ui.label}>Excerpt</span>
          <input
            className={ui.input}
            value={fields.excerpt}
            onChange={(e) => set('excerpt', e.target.value)}
            maxLength={300}
          />
        </label>
      </div>

      <div className={styles.panes}>
        <label className={ui.field}>
          <span className={ui.label}>Markdown · drop or paste images</span>
          <textarea
            ref={bodyRef}
            className={`${ui.textarea} ${styles.body}`}
            value={fields.bodyMd}
            onChange={(e) => set('bodyMd', e.target.value)}
            onDrop={onDrop}
            onPaste={onPaste}
            spellCheck
          />
        </label>
        <div className={ui.field}>
          <span className={ui.label}>Preview</span>
          <div className={styles.preview}>
            {cover && <img className={styles.cover} src={cover.url} srcSet={cover.srcset} sizes="50vw" alt="" />}
            {/* Rendered by our own server from Markdown, with raw HTML dropped. */}
            <div className={prose.prose} dangerouslySetInnerHTML={{ __html: preview.data ?? '' }} />
          </div>
        </div>
      </div>

      <ImagePanel
        coverId={fields.coverMediaId}
        uploading={upload.isPending}
        onUpload={uploadFiles}
        onInsert={(img) => insertBlock(img.markdown)}
        onCover={(id) => set('coverMediaId', id)}
      />
    </div>
  )
}
