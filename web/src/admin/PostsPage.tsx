import { useSuspenseQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'

import { adminPostsQuery } from '../api/queries'
import { formatDate, postCode } from '../lib/format'
import ui from './admin.module.css'
import styles from './PostsPage.module.css'

export function PostsPage() {
  const { data } = useSuspenseQuery(adminPostsQuery())

  return (
    <div className={styles.page}>
      <div className={styles.head}>
        <span className={ui.muted}>
          {data.items.length} {data.items.length === 1 ? 'post' : 'posts'}
        </span>
        <Link to="/admin/posts/new" className={ui.button}>
          New post
        </Link>
      </div>

      {data.items.length === 0 ? (
        <p className={ui.muted}>No posts yet. Write your first one.</p>
      ) : (
        <ol className={styles.list}>
          {data.items.map((post) => (
            <li key={post.id} className={styles.row}>
              <span className={ui.muted}>{postCode(post.id)}</span>
              <Link to="/admin/posts/$id" params={{ id: post.id }} className={styles.title}>
                {post.title}
              </Link>
              <span className={post.status === 'draft' ? styles.draft : ui.muted}>{post.status}</span>
              <span className={ui.muted}>Edited {formatDate(post.updatedAt)}</span>
              {post.status === 'published' ? (
                <Link to="/blog/$slug" params={{ slug: post.slug }}>
                  View
                </Link>
              ) : (
                <span />
              )}
            </li>
          ))}
        </ol>
      )}
    </div>
  )
}
