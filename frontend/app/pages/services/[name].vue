<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="px-4 py-3 border-b border-surface-100">
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 text-sm mb-3">
        <NuxtLink to="/" class="text-gray-400 hover:text-white transition-colors">
          Services
        </NuxtLink>
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-white">{{ service?.label || name }}</span>
      </div>

      <!-- Title and actions -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <!-- Status indicator -->
          <span
            class="w-3 h-3 rounded-full"
            :class="{
              'bg-green-500': service?.status === 'running',
              'bg-gray-500': service?.status === 'stopped',
              'bg-yellow-500': service?.status === 'unknown'
            }"
          ></span>
          <h1 class="text-lg font-semibold text-white">{{ service?.label || name }}</h1>
          <span
            class="px-2 py-0.5 rounded text-xs"
            :class="{
              'bg-green-600/20 text-green-400': service?.status === 'running',
              'bg-gray-600/20 text-gray-400': service?.status === 'stopped',
              'bg-yellow-600/20 text-yellow-400': service?.status === 'unknown'
            }"
          >
            {{ service?.status || 'unknown' }}
          </span>
        </div>

        <!-- Action buttons -->
        <div class="flex items-center gap-2">
          <button
            v-if="service?.status === 'running'"
            class="flex items-center gap-2 px-3 py-1.5 bg-red-600/20 hover:bg-red-600/30 text-red-400 text-sm rounded-lg transition-colors"
            :disabled="actionLoading"
            @click="handleStop"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
            </svg>
            Stop
          </button>
          <button
            v-else
            class="flex items-center gap-2 px-3 py-1.5 bg-green-600/20 hover:bg-green-600/30 text-green-400 text-sm rounded-lg transition-colors"
            :disabled="actionLoading"
            @click="handleStart"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Start
          </button>
          <button
            class="flex items-center gap-2 px-3 py-1.5 bg-surface-200 hover:bg-surface-100 text-white text-sm rounded-lg transition-colors"
            :disabled="actionLoading"
            @click="handleRestart"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Restart
          </button>
        </div>
      </div>
    </header>

    <!-- Tabs -->
    <div class="flex border-b border-surface-100">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="px-4 py-2.5 text-sm font-medium transition-colors border-b-2 -mb-px"
        :class="activeTab === tab.id
          ? 'text-primary-400 border-primary-400'
          : 'text-gray-400 hover:text-white border-transparent'"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Tab content -->
    <div class="flex-1 overflow-hidden">
      <!-- Loading state -->
      <div v-if="loading" class="flex items-center justify-center h-full">
        <div class="flex items-center gap-3 text-gray-400">
          <svg class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span>Loading service...</span>
        </div>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="flex items-center justify-center h-full">
        <div class="text-center">
          <p class="text-red-400 mb-4">{{ error }}</p>
          <button
            class="px-4 py-2 bg-surface-200 hover:bg-surface-100 text-white rounded-lg transition-colors"
            @click="loadService"
          >
            Retry
          </button>
        </div>
      </div>

      <!-- Tab panels -->
      <template v-else-if="service">
        <!-- Summary tab -->
        <ServiceSummary v-if="activeTab === 'summary'" :service="service" />

        <!-- Logs tab -->
        <ServiceLogs v-else-if="activeTab === 'logs'" :service-name="name" class="h-full" />

        <!-- Inspect tab -->
        <div v-else-if="activeTab === 'inspect'" class="p-4 h-full overflow-auto">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wider">Plist Content</h3>
            <button
              class="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-white transition-colors"
              title="Copy to clipboard"
              @click="copyPlist"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
            </button>
          </div>
          <div v-if="highlightedPlist" v-html="highlightedPlist" class="bg-surface-500 rounded-lg p-4 font-mono overflow-auto [&_pre]:!bg-transparent [&_pre]:!p-0 [&_pre]:!m-0 [&_code]:!text-sm"></div>
          <pre v-else-if="plistContent" class="bg-surface-500 rounded-lg p-4 font-mono text-sm text-gray-300 whitespace-pre-wrap overflow-auto">{{ plistContent }}</pre>
          <div v-else class="flex items-center justify-center h-48 text-gray-500">
            <p>No plist content available</p>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'
import { highlightCode } from '~/composables/useHighlighter'

const route = useRoute()
const name = computed(() => route.params.name as string)

const service = ref<Service | null>(null)
const plistContent = ref<string | null>(null)
const highlightedPlist = ref('')
const loading = ref(true)
const error = ref<string | null>(null)
const actionLoading = ref(false)
const activeTab = ref('summary')

const tabs = [
  { id: 'summary', label: 'Summary' },
  { id: 'logs', label: 'Logs' },
  { id: 'inspect', label: 'Inspect' }
]

async function loadService() {
  loading.value = true
  error.value = null

  try {
    if (window.go?.main?.App?.GetService) {
      service.value = await window.go.main.App.GetService(name.value)
    }
    if (window.go?.main?.App?.GetPlist) {
      plistContent.value = await window.go.main.App.GetPlist(name.value)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load service'
    console.error('Failed to load service:', e)
  } finally {
    loading.value = false
  }
}

async function handleStart() {
  actionLoading.value = true
  try {
    if (window.go?.main?.App?.StartService) {
      await window.go.main.App.StartService(name.value)
      await loadService()
    }
  } catch (e) {
    console.error('Failed to start service:', e)
  } finally {
    actionLoading.value = false
  }
}

async function handleStop() {
  actionLoading.value = true
  try {
    if (window.go?.main?.App?.StopService) {
      await window.go.main.App.StopService(name.value)
      await loadService()
    }
  } catch (e) {
    console.error('Failed to stop service:', e)
  } finally {
    actionLoading.value = false
  }
}

async function handleRestart() {
  actionLoading.value = true
  try {
    if (window.go?.main?.App?.RestartService) {
      await window.go.main.App.RestartService(name.value)
      await loadService()
    }
  } catch (e) {
    console.error('Failed to restart service:', e)
  } finally {
    actionLoading.value = false
  }
}

async function copyPlist() {
  if (plistContent.value) {
    try {
      await navigator.clipboard.writeText(plistContent.value)
    } catch (e) {
      console.error('Failed to copy to clipboard:', e)
    }
  }
}

onMounted(() => {
  loadService()
})

watch(name, () => {
  loadService()
})

// Highlight plist content when it changes
watch(() => plistContent.value, async (content) => {
  if (content) {
    highlightedPlist.value = await highlightCode(content)
  }
}, { immediate: true })
</script>
