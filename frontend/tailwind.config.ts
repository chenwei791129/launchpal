import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: [],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#a855f7',
          50: '#faf5ff',
          100: '#f3e8ff',
          200: '#e9d5ff',
          300: '#d8b4fe',
          400: '#c084fc',
          500: '#a855f7',
          600: '#9333ea',
          700: '#7e22ce',
          800: '#6b21a8',
          900: '#581c87',
        },
        surface: {
          DEFAULT: '#1e1e1e',
          50: '#3d3d3d',
          100: '#2d2d2d',
          200: '#252525',
          300: '#1e1e1e',
          400: '#171717',
          500: '#121212',
        },
      },
    },
  },
  plugins: [],
} satisfies Config
