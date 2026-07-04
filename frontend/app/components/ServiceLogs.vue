<template>
  <div class="flex flex-col h-full">
    <!-- Log controls -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-surface-100">
      <div class="flex items-center gap-2">
        <!-- Log type toggle -->
        <div class="flex rounded-lg bg-surface-300 p-0.5">
          <button
            class="px-3 py-1 text-sm rounded-md transition-colors"
            :class="logType === 'stdout' ? 'bg-surface-200 text-white' : 'text-gray-400 hover:text-white'"
            @click="logType = 'stdout'"
          >
            stdout
          </button>
          <button
            class="px-3 py-1 text-sm rounded-md transition-colors"
            :class="logType === 'stderr' ? 'bg-surface-200 text-white' : 'text-gray-400 hover:text-white'"
            @click="logType = 'stderr'"
          >
            stderr
          </button>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <!-- Auto-refresh toggle: independent of Auto-scroll. When checked, the
             current stream reloads every 2s through loadLogs. -->
        <label class="flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
          <input
            v-model="autoRefresh"
            data-testid="auto-refresh-toggle"
            type="checkbox"
            class="w-4 h-4 rounded border-gray-600 bg-surface-300 text-primary-600 focus:ring-primary-600 focus:ring-offset-0"
          >
          Auto-refresh
        </label>

        <!-- Auto-scroll toggle -->
        <label class="flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
          <input
            v-model="autoScroll"
            data-testid="auto-scroll-toggle"
            type="checkbox"
            class="w-4 h-4 rounded border-gray-600 bg-surface-300 text-primary-600 focus:ring-primary-600 focus:ring-offset-0"
          >
          Auto-scroll
        </label>

        <!-- Refresh button -->
        <button
          class="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-white transition-colors"
          title="Refresh logs"
          :disabled="loading"
          @click="loadLogs"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4"
            :class="{ 'animate-spin': loading }"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>

        <!-- Clear Logs button (hidden entirely for apple-system) -->
        <button
          v-if="clearControlState.visible"
          data-testid="clear-logs-button"
          class="p-1.5 rounded text-gray-400 transition-colors"
          :class="clearControlState.enabled
            ? 'hover:bg-surface-200 hover:text-white'
            : 'opacity-50 cursor-not-allowed'"
          :title="clearControlState.tooltip ?? 'Clear current log'"
          :disabled="!clearControlState.enabled"
          @click="openClearDialog"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4"
            viewBox="0 0 24 24"
            fill="currentColor"
          >
            <path d="M19.36 2.72l1.42 1.42-5.72 5.71c1.07 1.54 1.22 3.39.32 4.59L9.06 8.12c1.2-.9 3.05-.75 4.59.32l5.71-5.72M5.93 17.57c-2.01-2.01-3.24-4.41-3.58-6.65l4.88-2.09 7.44 7.44-2.09 4.88c-2.24-.34-4.64-1.57-6.65-3.58z"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Transient feedback row -->
    <div v-if="clearError || clearSuccess" class="px-4 py-2 border-b border-surface-100 text-sm">
      <span v-if="clearError" class="text-red-400">{{ clearError }}</span>
      <span v-else-if="clearSuccess" class="text-green-400">Log cleared</span>
    </div>

    <!-- Log content -->
    <div
      ref="logContainer"
      class="flex-1 overflow-auto bg-surface-500 p-4 font-mono text-sm"
    >
      <div v-if="loading && !logs" class="flex items-center justify-center h-full">
        <div class="flex items-center gap-3 text-gray-400">
          <svg class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
          </svg>
          <span>Loading logs...</span>
        </div>
      </div>

      <div v-else-if="error" class="text-red-400">
        {{ error }}
      </div>

      <!-- Branch on the raw content, not renderedLogs: content made of pure
           ANSI/control sequences renders to an empty string but the file is
           not empty, so it must still get the <pre>, not the placeholder.
           renderedLogs comes only from ansiToHtml, which HTML-escapes all log
           text and emits only a four-property style whitelist, so this v-html
           is the single controlled XSS surface. -->
      <!-- eslint-disable-next-line vue/no-v-html -->
      <pre v-else-if="logs?.content" class="text-gray-300 whitespace-pre-wrap break-all" v-html="renderedLogs" />

      <!-- Every non-content state — no-path / not-found / empty — shares this
           neutral placeholder, never the red error branch: structural states
           describe a normal condition of the service, not a failure. -->
      <div v-else class="flex items-center justify-center h-full text-gray-500">
        <div class="text-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto mb-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p>{{ placeholder.text }}</p>
          <p v-if="placeholder.subtext" class="text-xs text-gray-600 mt-1">{{ placeholder.subtext }}</p>
        </div>
      </div>
    </div>

    <!-- Clear Logs confirmation dialog. Reuses the surface from [name].vue's
         Run Now modal but uses red coloring to signal a destructive action. -->
    <Teleport to="body">
      <div
        v-if="showClearDialog"
        data-testid="clear-logs-dialog"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        @click.self="cancelClearDialog"
      >
        <div class="bg-surface-400 rounded-xl shadow-xl p-6 w-96">
          <h3 class="text-lg font-semibold text-white mb-2">Clear Logs</h3>
          <p class="text-gray-400 mb-6">
            This will permanently truncate the {{ logType }} log file for {{ serviceName }}. The file is reset to 0 bytes; existing entries cannot be recovered. Continue?
          </p>
          <div class="flex justify-end gap-3">
            <button
              data-testid="clear-logs-cancel"
              class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
              @click="cancelClearDialog"
            >
              Cancel
            </button>
            <button
              data-testid="clear-logs-confirm"
              class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="clearing"
              @click="confirmClear"
            >
              {{ clearing ? 'Clearing...' : 'Clear Logs' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ansiToHtml } from '~/utils/ansiToHtml'
import type { LogClearStatus, LogsResult } from '~/types/wails'

const props = withDefaults(defineProps<{
  serviceName: string
  serviceType?: string
  adminEnabled?: boolean
}>(), {
  serviceType: 'user',
  adminEnabled: false,
})

const logType = ref<'stdout' | 'stderr'>('stdout')
const logs = ref<LogsResult | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const autoScroll = ref(true)
// Auto-refresh is component-local and session-only (not persisted). While
// checked, pollTimer reloads the current stream every 2s; see startPolling.
const autoRefresh = ref(false)
const AUTO_REFRESH_INTERVAL_MS = 2000
let pollTimer: ReturnType<typeof setInterval> | null = null
const logContainer = ref<HTMLElement | null>(null)

// renderedLogs converts ANSI SGR escape sequences in the LogsResult content
// into styled HTML for v-html. Empty/absent content short-circuits to '' so
// the placeholder branch still owns the empty state.
const renderedLogs = computed(() => (logs.value?.content ? ansiToHtml(logs.value.content) : ''))

// Placeholder wording for the non-content states. no-path / not-found come
// from LogsResult.Status; everything else (empty content, development
// fallback) keeps the generic "No logs available" text.
const placeholder = computed(() => {
  if (logs.value?.status === 'no-path') {
    return { text: `No ${logType.value} log path configured for this service`, subtext: null }
  }
  if (logs.value?.status === 'not-found') {
    return { text: 'Log file does not exist yet', subtext: logs.value.path }
  }
  return { text: `No logs available for ${logType.value}`, subtext: null }
})

// logClearStatus is null until the first GetLogClearStatus resolves; the
// matrix below treats null as "pending" so the button stays disabled
// rather than flicker through enabled states.
const logClearStatus = ref<LogClearStatus | null>(null)
const showClearDialog = ref(false)
const clearing = ref(false)
const clearError = ref<string | null>(null)
const clearSuccess = ref(false)
let clearSuccessTimeout: ReturnType<typeof setTimeout> | null = null

async function loadLogs() {
  loading.value = true
  error.value = null

  // Only a LogsResult with Status "ok" is worth continuing to poll. Structural
  // outcomes (no-path / not-found), rejections, and the development fallback
  // (no LogsResult at all) will never change on their own, so they turn
  // Auto-refresh off below. This is the shared post-load path used by polling,
  // the manual Refresh button, and stream switches, so all three behave alike.
  let loadOk = false

  try {
    // System-domain services live under /Library/LaunchDaemons or
    // /System/Library/LaunchDaemons — UserManager.GetLogs can't resolve
    // those, so route through the system-aware binding instead.
    const app = window.go?.main?.App
    if (props.serviceType === 'user' && app?.GetLogs) {
      logs.value = await app.GetLogs(props.serviceName, logType.value)
    } else if (props.serviceType !== 'user' && app?.GetSystemLogs) {
      logs.value = await app.GetSystemLogs(props.serviceName, props.serviceType, logType.value)
    } else {
      // Development fallback (no Wails bindings available)
      logs.value = null
    }

    // Read the lowercase runtime key: the Wails runtime object carries
    // lowercase json keys (Status → status).
    loadOk = logs.value?.status === 'ok'

    if (autoScroll.value) {
      await nextTick()
      scrollToBottom()
    }
  } catch (e) {
    // Wails v2 rejects with the Go error as a plain string, so the string
    // check must come first — `instanceof Error` alone would discard the
    // backend message and always fall back to the generic text.
    error.value = typeof e === 'string' ? e : e instanceof Error ? e.message : 'Failed to load logs'
    console.error('Failed to load logs:', e)
  } finally {
    loading.value = false
  }

  // Auto-disable on any non-ok outcome. Setting autoRefresh false triggers the
  // watcher below, which clears the interval — there is no automatic resume;
  // the user must re-check the box. The rendered feedback (placeholder / error)
  // is unchanged: this adds no error surface of its own.
  if (autoRefresh.value && !loadOk) {
    autoRefresh.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(() => {
    // Skip this tick if a load is still in flight; don't queue or overlap.
    if (loading.value) return
    loadLogs()
  }, AUTO_REFRESH_INTERVAL_MS)
}

function stopPolling() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function loadLogClearStatus() {
  // Apple-system has no Clear button, so the status query is wasted. Skip
  // entirely to keep List/detail loads fast.
  if (props.serviceType === 'apple-system') {
    logClearStatus.value = null
    return
  }
  const app = window.go?.main?.App
  if (!app?.GetLogClearStatus) return
  try {
    logClearStatus.value = await app.GetLogClearStatus(
      props.serviceName,
      props.serviceType,
      logType.value,
    )
  } catch {
    // Silent fail: the disabled button + "Loading status..." tooltip is
    // the user-visible feedback. Writing to clearError here would mix
    // with confirm-clear failures (and clobber a recent success indicator)
    // even though the user took no clear action.
    logClearStatus.value = null
  }
}

function scrollToBottom() {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

interface ClearControlState {
  visible: boolean
  enabled: boolean
  tooltip: string | null
}

// Single computed so the template binds one object instead of three
// independent refs. Tooltip = null means "use the default Clear current
// log title". Reads props.adminEnabled directly so a mid-session Admin
// Mode toggle re-evaluates without an extra status query.
const clearControlState = computed<ClearControlState>(() => {
  if (props.serviceType === 'apple-system') {
    return { visible: false, enabled: false, tooltip: null }
  }
  if (!logClearStatus.value) {
    return { visible: true, enabled: false, tooltip: 'Loading status...' }
  }
  const status = logClearStatus.value
  if (!status.logPath) {
    return { visible: true, enabled: false, tooltip: 'No log path configured' }
  }
  if (!status.exists) {
    return { visible: true, enabled: false, tooltip: 'Log file does not exist' }
  }
  if (props.serviceType === 'user') {
    return { visible: true, enabled: true, tooltip: null }
  }
  if (status.userWritable || props.adminEnabled) {
    return { visible: true, enabled: true, tooltip: null }
  }
  return { visible: true, enabled: false, tooltip: 'Enable Admin Mode to clear' }
})

function openClearDialog() {
  clearError.value = null
  clearSuccess.value = false
  showClearDialog.value = true
}

function cancelClearDialog() {
  showClearDialog.value = false
}

async function confirmClear() {
  const app = window.go?.main?.App
  if (!app) return
  clearing.value = true
  clearError.value = null
  try {
    if (props.serviceType === 'user') {
      await app.ClearLogs(props.serviceName, logType.value)
    } else {
      await app.ClearSystemLogs(props.serviceName, props.serviceType, logType.value)
    }
    showClearDialog.value = false
    // Spec requires reloading via GetLogs / GetSystemLogs so the buffer
    // reflects the now-empty file from the same path the rest of the tab
    // reads from. Status query runs in parallel because writability may
    // have flipped if mode changed.
    await Promise.all([loadLogs(), loadLogClearStatus()])
    // loadLogs swallows its own failure into error.value; don't flash the
    // green success indicator on top of a red reload error.
    if (!error.value) {
      clearSuccess.value = true
      if (clearSuccessTimeout) clearTimeout(clearSuccessTimeout)
      clearSuccessTimeout = setTimeout(() => {
        clearSuccess.value = false
      }, 2000)
    }
  } catch (e) {
    // Same Wails string-first unwrap as loadLogs: Go errors arrive as
    // plain strings, so instanceof Error alone would discard the message.
    clearError.value = typeof e === 'string' ? e : e instanceof Error ? e.message : 'Failed to clear logs'
  } finally {
    clearing.value = false
  }
}

watch(logType, () => {
  // Drop the previous stream's result before fetching so the in-flight
  // switch shows the loading branch instead of the other stream's stale
  // content or placeholder. Auto-refresh state is intentionally preserved
  // across a stream switch; loadLogs auto-disables it if the new stream is
  // non-ok.
  logs.value = null
  loadLogs()
  loadLogClearStatus()
})

watch(autoRefresh, (enabled) => {
  if (enabled) startPolling()
  else stopPolling()
})

watch(() => props.serviceName, () => {
  // Service-to-service navigation reuses this component without remounting
  // (the detail page has no :key on ServiceLogs), so onBeforeUnmount won't
  // fire. Auto-refresh is a choice about the previous service, so reset it and
  // stop polling explicitly rather than silently carrying it to the new one.
  autoRefresh.value = false
  stopPolling()
})

onMounted(() => {
  loadLogs()
  loadLogClearStatus()
})

onBeforeUnmount(() => {
  if (clearSuccessTimeout) clearTimeout(clearSuccessTimeout)
  stopPolling()
})
</script>
