import { useSuspenseQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'

import { postsQuery } from '../api/queries'
import { formatDate, postCode } from '../lib/format'
import styles from './BlogIndex.module.css'

export function BlogIndex() {
  // The route loader already fetched this, so it's in the cache: no loading state needed.
  const { data } = useSuspenseQuery(postsQuery())

  if (data.items.length === 0) {
    return <p className={styles.empty}>Nothing here yet.</p>
  }

  return (
    <ol className={styles.list}>
      {data.items.map((post) => (
        <li key={post.id}>
          <Link to="/blog/$slug" params={{ slug: post.slug }} className={styles.row}>
            <span className={styles.code}>{postCode(post.id)}</span>
            <span className={styles.title}>{post.title}</span>
            <time className={styles.date} dateTime={post.publishedAt}>
              {formatDate(post.publishedAt)}
            </time>
          </Link>
        </li>
      ))}
    </ol>
  )
}
