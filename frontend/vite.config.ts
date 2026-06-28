import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    // The only expected chunk near this size after route/data splitting is maplibre-gl itself.
    // Keep the limit close to that vendor floor instead of broadly masking app-route bloat.
    chunkSizeWarningLimit: 1200,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return undefined
          if (id.includes('maplibre-gl') || id.includes('react-map-gl')) return 'maplibre'
          if (id.includes('@deck.gl') || id.includes('deck.gl') || id.includes('@luma.gl') || id.includes('@loaders.gl')) return 'deckgl'
          if (id.includes('@tanstack/react-query')) return 'query-vendor'
          if (id.includes('react-dom') || id.includes('react/jsx-runtime') || /node_modules[\\/]react[\\/]/.test(id)) return 'react-vendor'
          if (id.includes('date-fns')) return 'date-vendor'
          if (id.includes('lucide-react')) return 'icons-vendor'
          if (id.includes('recharts')) return 'charts-vendor'
          return undefined
        },
      },
    },
  },
})
