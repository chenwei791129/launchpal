// Simulate Nuxt auto-imports for vitest environment.
// Nuxt injects Vue composables as globals; tests run outside Nuxt, so we do it manually.
import { ref, computed, reactive, watch, watchEffect, toRef, toRefs, nextTick } from 'vue'

Object.assign(globalThis, {
  ref,
  computed,
  reactive,
  watch,
  watchEffect,
  toRef,
  toRefs,
  nextTick,
})
