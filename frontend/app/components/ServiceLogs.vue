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
        <!-- Auto-scroll toggle -->
        <label class="flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
          <input
            v-model="autoScroll"
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
      </div>
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

      <div v-else-if="!logs" class="flex items-center justify-center h-full text-gray-500">
        <div class="text-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto mb-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p>No logs available for {{ logType }}</p>
        </div>
      </div>

      <pre v-else class="text-gray-300 whitespace-pre-wrap break-all">{{ logs }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  serviceName: string
  serviceType?: string
}>(), {
  serviceType: 'user',
})

const logType = ref<'stdout' | 'stderr'>('stdout')
const logs = ref<string | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const autoScroll = ref(true)
const logContainer = ref<HTMLElement | null>(null)

async function loadLogs() {
  loading.value = true
  error.value = null

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

    if (autoScroll.value) {
      await nextTick()
      scrollToBottom()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load logs'
    console.error('Failed to load logs:', e)
  } finally {
    loading.value = false
  }
}

function scrollToBottom() {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

watch(logType, () => {
  loadLogs()
})

onMounted(() => {
  loadLogs()
})
</script>
