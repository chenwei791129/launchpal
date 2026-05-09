import { ref } from 'vue'
import type { Settings } from '~/types/wails'

const defaults: Settings = {
  userLogDir: '~/Library/Logs',
  systemLogDir: '/Library/Logs',
}

// Module-level singleton: every consumer reads the same Settings ref so
// the New Service modal sees Settings page edits without an event bus.
const settings = ref<Settings>({ ...defaults })

async function load() {
  const fn = window.go?.main?.App?.GetSettings
  if (!fn) return
  const next = await fn()
  settings.value = next
}

async function save(next: Settings) {
  const fn = window.go?.main?.App?.UpdateSettings
  if (!fn) {
    throw new Error('UpdateSettings binding is not available')
  }
  await fn(next)
  settings.value = next
}

export function useSettings() {
  return { settings, load, save, defaults }
}
