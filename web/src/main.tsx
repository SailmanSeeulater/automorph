import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { ApiError } from './api/client'
import { meQuery } from './api/queries'
import { makeRouter } from './router'
import './styles/global.css'

// If the session expires while you're in the admin, any request comes back 401.
// Forget the user and go to the login page, returning here afterwards.
function onError(error: Error) {
  if (error instanceof ApiError && error.status === 401) {
    queryClient.setQueryData(meQuery().queryKey, null)
    const here = router.state.location.href
    if (here.startsWith('/admin')) router.navigate({ to: '/login', search: { redirect: here } })
  }
}

const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError }),
  mutationCache: new MutationCache({ onError }),
  defaultOptions: {
    queries: { staleTime: 60_000 }, // posts don't change often; refetch at most once a minute
  },
})
const router = makeRouter(queryClient)

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
