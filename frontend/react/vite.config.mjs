import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import tailwindcss from '@tailwindcss/vite'

// https://vitejs.dev/config/
export default defineConfig({
  base: '/',
  plugins: [svgr(), react(), tailwindcss()],
  server: {
    // In Kind: Frontend calls auth.localhost directly (no proxy needed)
    // For local dev outside Kind, set VITE_KRATOS_URL=http://localhost:4433
    // and use: kubectl port-forward svc/kratos 4433:4433 -n app-namespace
  },
  test: {
    globals: true,
    css: true,
    reporters: ['verbose']
  },
})