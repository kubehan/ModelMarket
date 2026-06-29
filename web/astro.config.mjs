import { defineConfig } from 'astro/config'
import tailwind from '@astrojs/tailwind'

// https://astro.build/config
export default defineConfig({
  // 部署前改成你的实际域名
  site: 'https://modelmarket.example.com',
  integrations: [
    tailwind({ applyBaseStyles: false })
  ],
  build: {
    inlineStylesheets: 'auto'
  },
  output: 'static',
  compressHTML: true,
  prefetch: { prefetchAll: true }
})
