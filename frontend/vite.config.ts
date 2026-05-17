import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite-plus'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import wails from '@wailsio/runtime/plugins/vite'

export default defineConfig({
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [vue(), tailwindcss(), wails('./bindings')],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  fmt: {
    singleQuote: true,
  },
  lint: {
    ignorePatterns: ['dist/**', 'bindings/**'],
    options: {
      typeAware: true,
    },
  },
})
