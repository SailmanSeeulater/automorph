import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'

import { postsQuery } from '../api/queries'
import { postCode } from '../lib/format'
import { Logo } from './Logo'
import styles from './Frame.module.css'

const year = new Date().getFullYear()

// The page chrome: logo top-left, nav top-right, post codes down the left edge,
// colophon bottom-right. On wide screens it's pinned to the window; on narrow ones it flows.
export function Frame({ children }: { children: ReactNode }) {
  const { data } = useQuery(postsQuery())
  const recent = data?.items.slice(0, 4) ?? []

  return (
    <div className={styles.page}>
      <header>
        <Link to="/" className={styles.logo} aria-label="Automorph home">
          <Logo />
        </Link>
        <nav className={styles.nav}>
          <Link to="/" activeOptions={{ exact: true }} activeProps={{ className: styles.active }}>
            Blog
          </Link>
          <Link to="/about" activeProps={{ className: styles.active }}>
            About
          </Link>
          <Link to="/shop" activeProps={{ className: styles.active }}>
            Shop
          </Link>
        </nav>
      </header>

      {recent.length > 0 && (
        <aside className={styles.rail} aria-label="Recent posts">
          {recent.map((p) => (
            <Link key={p.id} to="/blog/$slug" params={{ slug: p.slug }}>
              : {postCode(p.id)}
            </Link>
          ))}
        </aside>
      )}

      <main className={styles.main}>{children}</main>

      <footer className={styles.colophon}>
        <Link to="/about">About</Link>
        <p>
          © {year} Automorph. All rights reserved.
          <br />
          This media is under exclusive right to its creators.
        </p>
        <p>latesailor.dev</p>
        <Logo size={36} />
      </footer>
    </div>
  )
}
