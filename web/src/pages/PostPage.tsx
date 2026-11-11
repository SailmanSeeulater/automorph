import { useSuspenseQuery } from '@tanstack/react-query'
import { getRouteApi, Link } from '@tanstack/react-router'

import { postQuery } from '../api/queries'
import { formatDate, postCode } from '../lib/format'
import styles from './PostPage.module.css'
import prose from '../styles/prose.module.css'

// Route IDs include the pathless "site" layout, even though the URL is just /blog/…
const route = getRouteApi('/site/blog/$slug')

export function PostPage() {
  const { slug } = route.useParams()
  const { data: post } = useSuspenseQuery(postQuery(slug))

  return (
    <article className={styles.post}>
      <header className={styles.meta}>
        <Link to="/">← Index</Link>
        <span>{postCode(post.id)}</span>
        <time dateTime={post.publishedAt}>{formatDate(post.publishedAt)}</time>
      </header>

      <h1 className={styles.title}>{post.title}</h1>

      {post.cover && (
        <img
          className={styles.cover}
          src={post.cover.url}
          srcSet={post.cover.srcset}
          sizes="(max-width: 800px) 100vw, 800px"
          width={post.cover.width}
          height={post.cover.height}
          alt={post.cover.alt}
        />
      )}

      {/* Safe to inject: the server renders this from Markdown and drops any raw HTML. */}
      <div className={prose.prose} dangerouslySetInnerHTML={{ __html: post.bodyHtml }} />
    </article>
  )
}
