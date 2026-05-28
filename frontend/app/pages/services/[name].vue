<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="px-4 py-3 border-b border-surface-100">
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 text-sm mb-3">
        <NuxtLink :to="backLink" class="text-gray-400 hover:text-white transition-colors">
          {{ backLinkLabel }}
        </NuxtLink>
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-white">{{ service?.label || name }}</span>
        <span
          v-if="service?.readOnly && !canWrite"
          class="inline-flex items-center px-1.5 py-0.5 rounded text-xs bg-gray-600/30 text-gray-400"
        >
          Read-only
        </span>
      </div>

      <!-- Title and actions -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <!-- Status indicator -->
          <span
            class="w-3 h-3 rounded-full"
            :class="{
              'bg-green-500': service?.status === 'running',
              'bg-blue-500': service?.status === 'loaded',
              'bg-gray-500': service?.status === 'stopped',
              'bg-yellow-500': service?.status === 'unknown'
            }"
          />
          <h1 class="text-lg font-semibold text-white">{{ service?.label || name }}</h1>
          <span
            class="px-2 py-0.5 rounded text-xs"
            :class="{
              'bg-green-600/20 text-green-400': service?.status === 'running',
              'bg-blue-600/20 text-blue-400': service?.status === 'loaded',
              'bg-gray-600/20 text-gray-400': service?.status === 'stopped',
              'bg-yellow-600/20 text-yellow-400': service?.status === 'unknown'
            }"
          >
            {{ service?.status || 'unknown' }}
          </span>
          <StatusConfidenceIcon :confidence="service?.statusConfidence" size="md" />
        </div>

        <!-- Action buttons: user services are always writable; system
             services are gated by Admin Mode (apple-system is never
             writable). -->
        <div v-if="canWrite" class="flex items-center gap-2">
          <button
            v-if="service?.status === 'running' || service?.status === 'loaded'"
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
          <button
            class="flex items-center gap-2 px-3 py-1.5 bg-amber-600/20 hover:bg-amber-600/30 text-amber-400 text-sm rounded-lg transition-colors"
            :disabled="actionLoading"
            @click="handleRunNow"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            Run Now
          </button>
          <button
            v-if="serviceType === 'user'"
            data-testid="copy-service-button"
            class="flex items-center gap-2 px-3 py-1.5 bg-surface-200 hover:bg-surface-100 text-white text-sm rounded-lg transition-colors disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="!service"
            title="Copy this service"
            @click="handleCopy"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            Copy
          </button>
        </div>
        <!-- Lock hint when the service is read-only because Admin Mode is off. -->
        <div
          v-else-if="serviceType === 'system' && !admin.isEnabled.value"
          class="flex items-center gap-2 text-sm text-gray-400"
          title="Enable Admin Mode to manage"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <NuxtLink to="/settings" class="underline hover:text-white">
            Enable Admin Mode to manage
          </NuxtLink>
        </div>
      </div>
      <!-- Action error surfaces Start/Stop/Restart/Run Now failures that
           would otherwise vanish into the console. -->
      <div v-if="actionError" class="mt-2 p-2 bg-red-500/10 border border-red-500/30 rounded text-sm text-red-400 flex items-start gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <div class="flex-1 break-all">{{ actionError }}</div>
        <button class="text-red-400 hover:text-white" @click="actionError = ''">×</button>
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
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
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
        <div v-if="activeTab === 'summary'" class="h-full overflow-auto">
          <ServiceSummary :service="service" />
        </div>

        <!-- Edit tab (user services only) -->
        <div v-else-if="activeTab === 'edit'" class="p-6 space-y-4 h-full overflow-auto">
          <div class="space-y-4">
            <!-- Program -->
            <div>
              <label class="block text-sm text-gray-400 mb-1">Program Path</label>
              <input
                v-model="editForm.program"
                type="text"
                class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
              >
              <p class="text-xs text-gray-500 mt-1">{{ PROGRAM_PATH_HINT }}</p>
            </div>

            <!-- Arguments -->
            <div>
              <label class="block text-sm text-gray-400 mb-1">Arguments</label>
              <input
                v-model="editArgumentsText"
                type="text"
                placeholder="--daemon --port=8080"
                class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
              >
              <p class="text-xs text-gray-500 mt-1">Space-separated arguments</p>
            </div>

            <!-- Working Directory -->
            <div>
              <label class="block text-sm text-gray-400 mb-1">Working Directory</label>
              <input
                v-model="editForm.workingDirectory"
                type="text"
                class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
              >
            </div>

            <!-- Checkboxes -->
            <div class="flex gap-6">
              <label class="flex items-center gap-2 text-sm text-gray-300">
                <input v-model="editForm.runAtLoad" type="checkbox" class="rounded bg-surface-400 border-surface-100" >
                Run at Load
              </label>
              <label class="flex items-center gap-2 text-sm text-gray-300">
                <input v-model="editForm.keepAlive" type="checkbox" class="rounded bg-surface-400 border-surface-100" >
                Keep Alive
              </label>
            </div>

            <!-- Environment Variables -->
            <div>
              <label class="block text-sm text-gray-400 mb-1">Environment Variables</label>
              <div class="space-y-2">
                <div v-for="(env, index) in editEnvVars" :key="index" class="flex gap-2">
                  <input
                    v-model="env.key"
                    type="text"
                    placeholder="KEY"
                    class="flex-1 px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500 font-mono text-sm"
                  >
                  <input
                    v-model="env.value"
                    :type="editEnvVisibility.has(index) ? 'text' : 'password'"
                    placeholder="Value"
                    class="flex-1 px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500 text-sm"
                  >
                  <button
                    type="button"
                    class="px-2 text-gray-500 hover:text-gray-300 transition-colors"
                    :title="editEnvVisibility.has(index) ? 'Hide value' : 'Show value'"
                    @click="toggleEditEnvVisibility(index)"
                  >
                    <svg v-if="!editEnvVisibility.has(index)" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    class="px-2 text-gray-500 hover:text-red-400 transition-colors"
                    @click="removeEditEnvVar(index)"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="mt-2 text-sm text-primary-400 hover:text-primary-300 transition-colors"
                @click="editEnvVars.push({ key: '', value: '' })"
              >
                + Add
              </button>
            </div>

            <!-- Schedule -->
            <ScheduleForm v-model="editSchedule" v-model:wake-system="editWakeSystem" />

            <!-- Save button -->
            <div class="flex items-center gap-3 pt-2">
              <button
                :disabled="saving || !canSaveEdit"
                class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded transition-colors"
                @click="handleSave"
              >
                {{ saving ? 'Saving...' : 'Save Changes' }}
              </button>
              <p v-if="!canSaveEdit" class="text-red-400 text-sm">Program Path and Arguments cannot both be empty.</p>
              <p v-else-if="saveError" class="text-red-400 text-sm">{{ saveError }}</p>
              <p v-else-if="saveSuccess" class="text-green-400 text-sm">Saved successfully!</p>
            </div>
          </div>
        </div>

        <!-- Logs tab -->
        <ServiceLogs
          v-else-if="activeTab === 'logs'"
          :service-name="name"
          :service-type="serviceType"
          :admin-enabled="admin.isEnabled.value"
          class="h-full"
        />

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
          <!-- eslint-disable-next-line vue/no-v-html -- shiki escapes input before producing the highlighted HTML -->
          <div v-if="highlightedPlist" class="bg-surface-500 rounded-lg p-4 font-mono overflow-auto [&_pre]:!bg-transparent [&_pre]:!p-0 [&_pre]:!m-0 [&_code]:!text-sm" v-html="highlightedPlist"/>
          <pre v-else-if="plistContent" class="bg-surface-500 rounded-lg p-4 font-mono text-sm text-gray-300 whitespace-pre-wrap overflow-auto">{{ plistContent }}</pre>
          <div v-else class="flex items-center justify-center h-48 text-gray-500">
            <p>No plist content available</p>
          </div>
        </div>
      </template>
    </div>

    <!-- Run Now confirmation dialog -->
    <Teleport to="body">
      <div
        v-if="showRunNowDialog"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        @click.self="showRunNowDialog = false"
      >
        <div class="bg-surface-400 rounded-xl shadow-xl p-6 w-96">
          <h3 class="text-lg font-semibold text-white mb-2">Run Now</h3>
          <p class="text-gray-400 mb-6">
            This service is currently running. Kickstart will terminate the existing process and start a new one. Continue?
          </p>
          <div class="flex justify-end gap-3">
            <button
              class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
              @click="showRunNowDialog = false"
            >
              Cancel
            </button>
            <button
              class="px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded-lg transition-colors"
              @click="confirmRunNow"
            >
              Run Now
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <CreateServiceModal
      :is-open="showCloneModal"
      service-type="user"
      :prefill="cloneSource"
      @close="closeCloneModal"
      @created="handleCloneCreated"
    />
  </div>
</template>

<script setup lang="ts">
import type { Service, ServiceConfig, ScheduleConfig } from '~/types/wails'
import { highlightCode } from '~/composables/useHighlighter'
import { useAdminMode } from '~/composables/useAdminMode'
import { parseShellArgs, serializeShellArgs } from '~/utils/shell-args'
import { hasProgramOrArguments, PROGRAM_PATH_HINT } from '~/utils/serviceValidation'
import { serviceToConfig } from '~/utils/serviceToConfig'

const route = useRoute()
const name = computed(() => route.params.name as string)
const serviceType = computed(() => (route.query.type as string) || 'user')
const admin = useAdminMode()

// canWrite is true for user services and for system services when Admin
// Mode is Enabled. apple-system services are never writable.
const canWrite = computed(() => {
  if (serviceType.value === 'user') return true
  if (serviceType.value === 'system') return admin.isEnabled.value
  return false
})

const service = ref<Service | null>(null)
const plistContent = ref<string | null>(null)
const highlightedPlist = ref('')
const loading = ref(true)
const error = ref<string | null>(null)
const actionLoading = ref(false)
const activeTab = ref('summary')

const tabs = computed(() => {
  const base = [
    { id: 'summary', label: 'Summary' },
    { id: 'logs', label: 'Logs' },
    { id: 'inspect', label: 'Inspect' },
  ]
  // Edit tab follows canWrite so it appears for system daemons while Admin
  // Mode is enabled and disappears when the user disables it mid-session.
  if (canWrite.value) {
    base.splice(1, 0, { id: 'edit', label: 'Edit' })
  }
  return base
})

// If Admin Mode turns off while the Edit tab is active, slide the user back
// to Summary so the UI doesn't get stuck on a tab that no longer exists.
watch(canWrite, (writable) => {
  if (!writable && activeTab.value === 'edit') {
    activeTab.value = 'summary'
  }
})

const backLink = computed(() => {
  switch (serviceType.value) {
    case 'system':
      return '/system'
    case 'apple-system':
      return '/apple-system'
    default:
      return '/'
  }
})

const backLinkLabel = computed(() => {
  switch (serviceType.value) {
    case 'system':
      return 'System Services'
    case 'apple-system':
      return 'Apple System Services'
    default:
      return 'Services'
  }
})

async function loadService() {
  loading.value = true
  error.value = null

  try {
    const type = serviceType.value
    if (type === 'user') {
      const [svc, plist] = await Promise.all([
        window.go?.main?.App?.GetService?.(name.value),
        window.go?.main?.App?.GetPlist?.(name.value),
      ])
      service.value = svc ?? null
      plistContent.value = plist ?? null
    } else {
      const [svc, plist] = await Promise.all([
        window.go?.main?.App?.GetSystemService?.(name.value, type),
        window.go?.main?.App?.GetSystemPlist?.(name.value, type),
      ])
      service.value = svc ?? null
      plistContent.value = plist ?? null
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load service'
    console.error('Failed to load service:', e)
  } finally {
    loading.value = false
    populateEditForm()
  }
}

// startFn / stopFn / restartFn pick the right Wails binding based on the
// service type. System-domain writes require Admin Mode; if the backend
// replies with ErrReadOnlyManager the message surfaces in the console.
const startFn = computed(() => {
  return serviceType.value === 'system'
    ? window.go?.main?.App?.StartSystemService
    : window.go?.main?.App?.StartService
})
const stopFn = computed(() => {
  return serviceType.value === 'system'
    ? window.go?.main?.App?.StopSystemService
    : window.go?.main?.App?.StopService
})
const restartFn = computed(() => {
  return serviceType.value === 'system'
    ? window.go?.main?.App?.RestartSystemService
    : window.go?.main?.App?.RestartService
})

// actionError surfaces backend errors from Start/Stop/Restart/Run Now to
// the user — silent console.error leaves the GUI looking like a no-op when
// launchd rejects the request (e.g. "file not found", plist parse error).
const actionError = ref('')

async function handleStart() {
  actionLoading.value = true
  actionError.value = ''
  try {
    if (startFn.value) {
      await startFn.value(name.value)
      await loadService()
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : String(e)
    console.error('Failed to start service:', e)
  } finally {
    actionLoading.value = false
  }
}

async function handleStop() {
  actionLoading.value = true
  actionError.value = ''
  try {
    if (stopFn.value) {
      await stopFn.value(name.value)
      await loadService()
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : String(e)
    console.error('Failed to stop service:', e)
  } finally {
    actionLoading.value = false
  }
}

async function handleRestart() {
  actionLoading.value = true
  actionError.value = ''
  try {
    if (restartFn.value) {
      await restartFn.value(name.value)
      await loadService()
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : String(e)
    console.error('Failed to restart service:', e)
  } finally {
    actionLoading.value = false
  }
}

const showRunNowDialog = ref(false)

const showCloneModal = ref(false)
const cloneSource = ref<ServiceConfig | null>(null)

function handleCopy() {
  if (!service.value) return
  cloneSource.value = serviceToConfig(service.value)
  showCloneModal.value = true
}

function closeCloneModal() {
  showCloneModal.value = false
  cloneSource.value = null
}

async function handleCloneCreated(label: string) {
  showCloneModal.value = false
  cloneSource.value = null
  await navigateTo(`/services/${encodeURIComponent(label)}?type=user`)
}

function handleRunNow() {
  if (service.value?.status === 'running') {
    showRunNowDialog.value = true
    return
  }
  executeKickstart()
}

async function confirmRunNow() {
  showRunNowDialog.value = false
  await executeKickstart()
}

async function executeKickstart() {
  actionLoading.value = true
  actionError.value = ''
  try {
    // System daemons have no dedicated kickstart binding — RestartSystemService
    // already routes to the helper's Kickstart RPC (`launchctl kickstart -k`),
    // which is the same semantic as KickstartService for user services.
    const kickFn = serviceType.value === 'system'
      ? window.go?.main?.App?.RestartSystemService
      : window.go?.main?.App?.KickstartService
    if (kickFn) {
      await kickFn(name.value)
      await loadService()
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : String(e)
    console.error('Failed to run kickstart:', e)
  } finally {
    actionLoading.value = false
  }
}

// Edit form state
const editForm = reactive({
  program: '',
  workingDirectory: '',
  runAtLoad: false,
  keepAlive: false,
})
const editArgumentsText = ref('')
const editEnvVars = reactive<Array<{ key: string; value: string }>>([])
const editEnvVisibility = reactive(new Set<number>())

function toggleEditEnvVisibility(index: number) {
  if (editEnvVisibility.has(index)) {
    editEnvVisibility.delete(index)
  } else {
    editEnvVisibility.add(index)
  }
}

function removeEditEnvVar(index: number) {
  editEnvVars.splice(index, 1)
  // Shift visibility indices to stay aligned after removal
  const next = new Set<number>()
  for (const i of editEnvVisibility) {
    if (i < index) next.add(i)
    else if (i > index) next.add(i - 1)
  }
  editEnvVisibility.clear()
  for (const i of next) editEnvVisibility.add(i)
}
const editSchedule = ref<ScheduleConfig | undefined>(undefined)
const editWakeSystem = ref(false)
const saving = ref(false)
const saveError = ref('')
const saveSuccess = ref(false)

const canSaveEdit = computed(() =>
  hasProgramOrArguments(editForm.program, editArgumentsText.value),
)

function populateEditForm() {
  if (!service.value) return
  editEnvVisibility.clear()
  editForm.program = service.value.program || ''
  editForm.workingDirectory = service.value.workingDirectory || ''
  editForm.runAtLoad = service.value.runAtLoad
  editForm.keepAlive = service.value.keepAlive
  editArgumentsText.value = serializeShellArgs(service.value.arguments ?? [])
  editEnvVars.splice(0)
  if (service.value.environment) {
    for (const [key, value] of Object.entries(service.value.environment)) {
      editEnvVars.push({ key, value })
    }
  }
  editSchedule.value = service.value.schedule ? { ...service.value.schedule } : undefined
  editWakeSystem.value = service.value.wakeSystem ?? false
}

async function handleSave() {
  if (!service.value) return
  if (!canSaveEdit.value) return
  saving.value = true
  saveError.value = ''
  saveSuccess.value = false

  try {
    const environment: Record<string, string> = {}
    for (const env of editEnvVars) {
      if (env.key.trim()) {
        environment[env.key.trim()] = env.value
      }
    }

    const config: ServiceConfig = {
      label: service.value.label,
      program: editForm.program,
      arguments: parseShellArgs(editArgumentsText.value),
      runAtLoad: editForm.runAtLoad,
      keepAlive: editForm.keepAlive,
      environment: Object.keys(environment).length > 0 ? environment : undefined,
      workingDirectory: editForm.workingDirectory,
      schedule: editSchedule.value,
      wakeSystem: editWakeSystem.value,
      stdoutPath: service.value.stdoutPath,
      stderrPath: service.value.stderrPath,
    }

    if (serviceType.value === 'system') {
      await window.go.main.App.UpdateSystemService(name.value, config)
    } else {
      await window.go.main.App.UpdateService(name.value, config)
    }
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 3000)
    await loadService()
  } catch (e: unknown) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save changes'
  } finally {
    saving.value = false
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
