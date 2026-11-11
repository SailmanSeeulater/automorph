import type { QueryClient } from '@tanstack/react-query'
import {
  createRootRouteWithContext,
  createRoute,
  createRouter,
  notFound,
  Outlet,
  redirect,
} from '@tanstack/react-router'

import { AdminLayout } from './admin/AdminLayout'
import { EditPostPage, NewPostPage } from './admin/Editor'
import { LoginPage } from './admin/LoginPage'
import { PostsPage } from './admin/PostsPage'
import { adminPostQuery, adminPostsQuery, meQuery, postQuery, postsQuery } from './api/queries'
import { Frame } from './components/Frame'
import { BlogIndex } from './pages/BlogIndex'
import { PostPage } from './pages/PostPage'
import { About, ErrorPage, NotFound, Shop } from './pages/Simple'

// Every route can reach the QueryClient through its loader's context.
const rootRoute = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: Outlet,
  notFoundComponent: () => (
    <Frame>
      <NotFound />
    </Frame>
  ),
  errorComponent: (props) => (
    <Frame>
      <ErrorPage {...props} />
    </Frame>
  ),
})

// ---- Public site: every page inside the frame ----

// A "pathless" route: it adds the Frame around its children without adding to the URL.
const siteRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'site',
  component: () => (
    <Frame>
      <Outlet />
    </Frame>
  ),
  notFoundComponent: NotFound,
  errorComponent: ErrorPage,
})

const indexRoute = createRoute({
  getParentRoute: () => siteRoute,
  path: '/',
  loader: ({ context }) => context.queryClient.ensureQueryData(postsQuery()),
  component: BlogIndex,
})

const postRoute = createRoute({
  getParentRoute: () => siteRoute,
  path: '/blog/$slug',
  loader: ({ context, params }) => context.queryClient.ensureQueryData(postQuery(params.slug)),
  component: PostPage,
})

const aboutRoute = createRoute({ getParentRoute: () => siteRoute, path: '/about', component: About })
const shopRoute = createRoute({ getParentRoute: () => siteRoute, path: '/shop', component: Shop })

// ---- Admin ----

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  validateSearch: (search: Record<string, unknown>): { redirect?: string } => ({
    redirect: typeof search.redirect === 'string' ? search.redirect : undefined,
  }),
  component: LoginPage,
})

// Everything under /admin requires an admin. The server checks too; this just sends
// logged-out visitors to the login page instead of showing them errors.
const adminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  beforeLoad: async ({ context, location }) => {
    const me = await context.queryClient.ensureQueryData(meQuery())
    if (!me || me.role !== 'admin') {
      throw redirect({ to: '/login', search: { redirect: location.href } })
    }
    return { me }
  },
  component: AdminLayout,
})

const adminIndexRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: '/',
  loader: ({ context }) => context.queryClient.ensureQueryData(adminPostsQuery()),
  component: PostsPage,
})

const newPostRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: '/posts/new',
  component: NewPostPage,
})

const editPostRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: '/posts/$id',
  // URL params are strings; turn "7" into 7 so the rest of the code gets a number.
  params: {
    parse: ({ id }) => ({ id: Number(id) }),
    stringify: ({ id }) => ({ id: String(id) }),
  },
  loader: ({ context, params }) => {
    if (!Number.isInteger(params.id)) throw notFound()
    return context.queryClient.ensureQueryData(adminPostQuery(params.id))
  },
  component: EditPostPage,
})

const routeTree = rootRoute.addChildren([
  siteRoute.addChildren([indexRoute, postRoute, aboutRoute, shopRoute]),
  loginRoute,
  adminRoute.addChildren([adminIndexRoute, newPostRoute, editPostRoute]),
])

export function makeRouter(queryClient: QueryClient) {
  return createRouter({
    routeTree,
    context: { queryClient },
    defaultPreload: 'intent', // start loading a page when its link is hovered
    defaultPreloadStaleTime: 0, // let React Query decide what's fresh
    scrollRestoration: true,
  })
}

// Makes <Link to="..."> and useParams() type-checked against the routes above.
declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof makeRouter>
  }
}
