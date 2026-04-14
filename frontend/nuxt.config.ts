// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-25',
  ssr: false,
  devtools: { enabled: false },
  modules: ['@nuxtjs/tailwindcss', '@nuxt/eslint'],
  css: ['~/assets/css/main.css'],
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
