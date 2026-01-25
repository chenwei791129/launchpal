// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-25',
  ssr: false,
  devtools: { enabled: false },
  app: {
    baseURL: './',
    buildAssetsDir: 'assets',
  },
  vite: {
    build: {
      assetsInlineLimit: 0,
    },
  },
})
