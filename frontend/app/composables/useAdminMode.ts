import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { AdminModeState, AdminModeStatus } from '~/types/wails'

// Shared reactive Admin Mode status. Pages/components read this to gate
// write controls; the composable keeps it in sync with the backend by
// subscribing to the "admin_mode:state" Wails event and polling on mount.
const state = ref<AdminModeState>('disabled')
const lastError = ref<string | null>(null)
const loading = ref(false)

// Event subscription is lazy: the first component that mounts uses the
// composable to register the listener; subsequent subscribers reuse it.
let subscribed = false

declare global {
  interface Window {
    runtime?: {
      EventsOn?(event: string, callback: (...args: unknown[]) => void): () => void
      EventsOff?(event: string): void
    }
  }
}

function applyStatus(status: AdminModeStatus | null | undefined) {
  if (!status) return
  // Guard writes so watchers and templates don't see a no-op tick every
  // refresh. Vue's ref notifies on every assignment even when the value
  // is identical — in a polling composable that flips to unnecessary
  // re-renders.
  if (state.value !== status.state) state.value = status.state
  if (lastError.value !== status.error) lastError.value = status.error
}

async function refresh() {
  const status = await window.go?.main?.App?.GetAdminModeStatus?.()
  applyStatus(status)
}

function ensureSubscribed() {
  if (subscribed) return
  subscribed = true
  const runtime = window.runtime
  if (runtime?.EventsOn) {
    runtime.EventsOn('admin_mode:state', (...args) => {
      const payload = args[0] as AdminModeStatus | undefined
      applyStatus(payload)
    })
  }
}

export function useAdminMode() {
  const isEnabled = computed(() => state.value === 'enabled')
  const isRequesting = computed(() => state.value === 'requesting')
  const isShuttingDown = computed(() => state.value === 'shutting_down')

  async function enable() {
    if (loading.value) return
    loading.value = true
    try {
      await window.go?.main?.App?.EnableAdminMode?.()
      await refresh()
    } catch (e) {
      lastError.value = e instanceof Error ? e.message : String(e)
      await refresh()
    } finally {
      loading.value = false
    }
  }

  async function disable() {
    if (loading.value) return
    loading.value = true
    try {
      await window.go?.main?.App?.DisableAdminMode?.()
      await refresh()
    } catch (e) {
      lastError.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    ensureSubscribed()
    refresh().catch(() => {})
  })

  onUnmounted(() => {
    // Intentional: leave global subscription running so the status stays
    // fresh when only one consumer is mounted at a time.
  })

  return {
    state,
    lastError,
    loading,
    isEnabled,
    isRequesting,
    isShuttingDown,
    enable,
    disable,
    refresh,
  }
}
