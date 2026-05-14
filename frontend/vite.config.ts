import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/upload_data': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/upload_events': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/events': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/farmer_data': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/load_orders': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/clients': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/client_bonus': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/spend_bonus': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/recommendations': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/vector': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
