import { useMutation, useQueryClient } from '@tanstack/react-query'
import { getRouteApi, useRouter } from '@tanstack/react-router'
import { useState, type FormEvent } from 'react'

import { api, ApiError } from '../api/client'
import { meQuery } from '../api/queries'
import { Logo } from '../components/Logo'
import ui from './admin.module.css'
import styles from './LoginPage.module.css'

const route = getRouteApi('/login')

export function LoginPage() {
  const { redirect } = route.useSearch()
  const router = useRouter()
  const queryClient = useQueryClient()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const login = useMutation({
    mutationFn: async () => {
      const { data, error, response } = await api.POST('/api/auth/login', { body: { email, password } })
      if (error) throw new ApiError(response.status, error)
      return data
    },
    onSuccess: (user) => {
      queryClient.setQueryData(meQuery().queryKey, user)
      // Only follow redirects within this site, never to another domain.
      const target = redirect?.startsWith('/') && !redirect.startsWith('//') ? redirect : '/admin'
      router.history.push(target)
    },
  })

  function submit(e: FormEvent) {
    e.preventDefault()
    login.mutate()
  }

  return (
    <div className={styles.page}>
      <form className={styles.form} onSubmit={submit}>
        <Logo />
        <label className={ui.field}>
          <span className={ui.label}>Email</span>
          <input
            className={ui.input}
            type="email"
            autoComplete="username"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </label>
        <label className={ui.field}>
          <span className={ui.label}>Password</span>
          <input
            className={ui.input}
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>
        {login.error && <p className={ui.error}>{login.error.message}</p>}
        <button className={ui.button} type="submit" disabled={login.isPending}>
          {login.isPending ? 'Logging in…' : 'Log in'}
        </button>
      </form>
    </div>
  )
}
