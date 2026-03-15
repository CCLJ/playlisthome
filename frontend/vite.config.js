import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Proxy all /auth and /api calls to the Go backend
      '/auth': {
        target: process.env.VITE_API_URL || 'http://backend:8080',
        changeOrigin: true,
      },
      '/api': {
        target: process.env.VITE_API_URL || 'http://backend:8080',
        changeOrigin: true,
      },
    },
  },
})
