import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    ignores: ['wailsjs/**'],
  },
  {
    files: ['app/components/Sidebar.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
)
