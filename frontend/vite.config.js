import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/auth/login': 'http://localhost:8080',
      '/auth/register': 'http://localhost:8080',
      '/auth/refresh': 'http://localhost:8080',
      '/auth/logout': 'http://localhost:8080',
      '/auth/2fa': 'http://localhost:8080',
      '/auth/logout-all': 'http://localhost:8080',
      '/doctors': 'http://localhost:8080',
      '/patients': 'http://localhost:8080',
      '/clinics': 'http://localhost:8080',
      '/chats': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
    }
  }
})
