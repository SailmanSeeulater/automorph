import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Link, Outlet, useNavigate, useRouteContext } from '@tanstack/react-router'

import { api } from '../api/client'
import { meQuery } from '../api/queries'
import { Logo } from '../components/Logo'
import styles from './AdminLayout.module.css'

// Chrome for /admin/*: a plain top bar and the full window width for the editor.
export function AdminLayout() {
  const { me } = useRouteContext({ from: '/admin' })
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const logout = useMutation({
    mutationFn: () => api.POST('/api/auth/logout'),
    onSuccess: () => {
      queryClient.setQueryData(meQuery().queryKey, null)
      queryClient.removeQueries({ queryKey: ['admin'] })
      navigate({ to: '/login' })
    },
  })

  return (
    <div className={styles.page}>
      <header className={styles.bar}>
        <Link to="/admin" className={styles.logo} aria-label="Admin home">
          <Logo size={32} />
        </Link>
        <nav className={styles.nav}>
          <Link to="/admin" activeOptions={{ exact: true }} activeProps={{ className: styles.active }}>
            Posts
          </Link>
          <Link to="/admin/posts/new" activeProps={{ className: styles.active }}>
            New post
          </Link>
          <Link to="/">View site</Link>
          <button type="button" onClick={() => logout.mutate()} disabled={logout.isPending}>
            Log out {me.displayName || me.email}
          </button>
        </nav>
      </header>
      <main className={styles.main}>
        <Outlet />
      </main>
    </div>
  )
}
