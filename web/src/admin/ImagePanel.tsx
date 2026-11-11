import { useQuery } from '@tanstack/react-query'
import { useRef } from 'react'

import type { Image } from '../api/client'
import { imagesQuery } from '../api/queries'
import ui from './admin.module.css'
import styles from './Editor.module.css'

type Props = {
  coverId?: number
  uploading: boolean
  onUpload: (files: File[]) => void
  onInsert: (img: Image) => void
  onCover: (id: number | undefined) => void
}

// Every uploaded image, newest first. Click to insert into the post, or make it the cover.
export function ImagePanel({ coverId, uploading, onUpload, onInsert, onCover }: Props) {
  const { data } = useQuery(imagesQuery())
  const input = useRef<HTMLInputElement>(null)

  return (
    <section className={styles.images}>
      <div className={styles.imagesHead}>
        <span className={ui.label}>Images</span>
        <button type="button" onClick={() => input.current?.click()} disabled={uploading}>
          {uploading ? 'Uploading…' : 'Upload'}
        </button>
        <input
          ref={input}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          hidden
          onChange={(e) => {
            onUpload([...(e.target.files ?? [])])
            e.target.value = '' // so picking the same file again still triggers onChange
          }}
        />
      </div>

      {data && data.items.length === 0 && (
        <p className={ui.muted}>Drop an image onto the text, or press Upload.</p>
      )}

      <ul className={styles.grid}>
        {data?.items.map((img) => (
          <li key={img.id} className={img.id === coverId ? styles.isCover : undefined}>
            <img src={img.url} srcSet={img.srcset} sizes="120px" alt={img.alt} loading="lazy" />
            <div className={styles.thumbActions}>
              <button type="button" onClick={() => onInsert(img)}>
                Insert
              </button>
              <button type="button" onClick={() => onCover(img.id === coverId ? undefined : img.id)}>
                {img.id === coverId ? 'Cover ✓' : 'Cover'}
              </button>
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}
