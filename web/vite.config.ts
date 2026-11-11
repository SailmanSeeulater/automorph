import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// The Go server. Override with API_TARGET=http://127.0.0.1:9000 npm run dev
const api = process.env.API_TARGET ?? 'http://127.0.0.1:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // Fixed port so it doesn't collide with other Vite projects on 5173.
    port: 5180,
    strictPort: true,
    // In dev, the browser only talks to Vite; these paths are forwarded to Go.
    // Same origin means cookies and CSRF protection work exactly like in production.
    proxy: {
      '/api': api,
      '/media': api,
    },
  },
})
