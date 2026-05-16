<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Admin Mode banner: nudges the user to Settings when they land here
         without Admin Mode enabled. Hidden once Admin Mode is on. -->
    <div
      v-if="!admin.isEnabled.value"
      class="mx-4 mt-4 p-3 bg-surface-400 border border-surface-100 rounded-lg flex items-center gap-3"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>
      <div class="flex-1 text-sm text-gray-300">
        System services are read-only. Enable Admin Mode in Settings to manage them.
      </div>
      <NuxtLink
        to="/settings"
        class="text-sm px-3 py-1.5 bg-primary-600 hover:bg-primary-700 text-white rounded transition-colors"
      >
        Open Settings
      </NuxtLink>
    </div>

    <!-- Permission warning: Full Disk Access is required to see every daemon. -->
    <div
      v-if="!hasPermission"
      class="mx-4 mt-4 p-4 bg-yellow-500/10 border border-yellow-500/20 rounded-lg flex items-start gap-3"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-yellow-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <div class="flex-1">
        <h3 class="text-sm font-semibold text-yellow-500 mb-1">Limited Access</h3>
        <p class="text-sm text-gray-300 mb-2">
          Some system services may not be visible. LaunchPal needs disk access permission to read system LaunchDaemons.
        </p>
        <button
          class="text-sm px-3 py-1.5 bg-yellow-500/20 hover:bg-yellow-500/30 text-yellow-500 rounded transition-colors"
          @click="openSystemSettings"
        >
          Open System Settings
        </button>
      </div>
    </div>

    <!-- Replaces alert() so users see feedback without losing page focus. -->
    <InlineBanner :message="actionMessage" @dismiss="actionMessage = null" />

    <!-- Header with search + (Admin Mode) New Service button -->
    <header class="flex items-center justify-between px-4 py-3 border-b border-surface-100">
      <h1 class="text-lg font-semibold text-white">System Services</h1>
      <div class="flex items-center gap-3">
        <div class="relative">
          <svg xmlns="http://www.w3.org/2000/svg" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search services..."
            class="w-64 pl-10 pr-4 py-1.5 bg-surface-300 border border-surface-100 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-600 focus:border-transparent"
          >
        </div>

        <button
          class="p-2 rounded-lg text-gray-400 hover:text-white hover:bg-surface-200 transition-colors"
          title="Refresh"
          @click="loadServices"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>

        <button
          v-if="admin.isEnabled.value"
          class="flex items-center gap-2 px-3 py-1.5 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition-colors"
          @click="showCreateModal = true"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          New Service
        </button>
      </div>
    </header>

    <!-- Table header -->
    <div class="flex items-center px-4 py-2 bg-surface-400 border-b border-surface-100 text-xs text-gray-400 uppercase tracking-wider">
      <div class="w-16 shrink-0">Status</div>
      <div class="flex-1 min-w-0">Name</div>
      <div class="w-24 shrink-0 text-center">Type</div>
      <div class="w-24 shrink-0">Actions</div>
    </div>

    <div class="flex-1 overflow-y-auto">
      <div v-if="loading" class="flex items-center justify-center h-full">
        <div class="flex items-center gap-3 text-gray-400">
          <svg class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
          </svg>
          <span>Loading services...</span>
        </div>
      </div>

      <div v-else-if="error" class="flex items-center justify-center h-full">
        <div class="text-center">
          <p class="text-red-400 mb-4">{{ error }}</p>
          <button
            class="px-4 py-2 bg-surface-200 hover:bg-surface-100 text-white rounded-lg transition-colors"
            @click="loadServices"
          >
            Retry
          </button>
        </div>
      </div>

      <div v-else-if="filteredServices.length === 0" class="flex items-center justify-center h-full">
        <div class="text-center text-gray-400">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto mb-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p v-if="searchQuery">No services found matching "{{ searchQuery }}"</p>
          <p v-else>No services found</p>
        </div>
      </div>

      <template v-else>
        <ServiceRow
          v-for="service in filteredServices"
          :key="service.name"
          :service="service"
          :selected="selectedService?.name === service.name"
          @select="handleSelect"
          @delete="handleDelete"
          @refresh="loadServices"
        />
      </template>
    </div>

    <!-- Delete confirmation dialog -->
    <Teleport to="body">
      <div
        v-if="showDeleteDialog"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        @click.self="showDeleteDialog = false"
      >
        <div class="bg-surface-400 rounded-xl shadow-xl p-6 w-96">
          <h3 class="text-lg font-semibold text-white mb-2">Delete System Daemon</h3>
          <p class="text-gray-400 mb-2">
            Are you sure you want to delete "{{ serviceToDelete?.label }}"?
          </p>
          <p class="text-yellow-500 text-sm mb-4">
            The daemon will be booted out and its plist backed up before removal.
          </p>
          <label class="flex items-start gap-2 mb-6 cursor-pointer select-none">
            <input
              v-model="deleteLogsChecked"
              type="checkbox"
              class="mt-1 w-4 h-4 rounded border-surface-100 bg-surface-300 text-primary-600 focus:ring-primary-600 focus:ring-offset-0"
            >
            <span class="text-sm text-gray-300">
              Also delete log files
              <span class="block text-xs text-gray-500">Log files will be permanently deleted and cannot be recovered.</span>
            </span>
          </label>
          <div class="flex justify-end gap-3">
            <button
              class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
              @click="showDeleteDialog = false"
            >
              Cancel
            </button>
            <button
              class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
              @click="confirmDelete"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create Service Modal (system variant) -->
    <CreateServiceModal
      :is-open="showCreateModal"
      service-type="system"
      @close="showCreateModal = false"
      @created="loadServices"
    />
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'
import { useAdminMode } from '~/composables/useAdminMode'

const admin = useAdminMode()

const services = ref<Service[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')
const selectedService = ref<Service | null>(null)
const hasPermission = ref(true)
const showCreateModal = ref(false)
const showDeleteDialog = ref(false)
const serviceToDelete = ref<Service | null>(null)
const deleteLogsChecked = ref(false)
const actionMessage = ref<string | null>(null)

const updateCounts = inject<(total: number, running: number) => void>('updateCounts')

const filteredServices = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return services.value
  return services.value.filter(service => service.label.toLowerCase().includes(query))
})

async function loadServices() {
  loading.value = true
  error.value = null
  try {
    services.value = await window.go.main.App.ListSystemServices()
    const runningCount = services.value.filter(s => s.status === 'running').length
    updateCounts?.(services.value.length, runningCount)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load services'
    console.error('Failed to load services:', e)
  } finally {
    loading.value = false
  }
}

async function checkPermissions() {
  try {
    if (window.go?.main?.App?.CheckPermissions) {
      const perms = await window.go.main.App.CheckPermissions()
      hasPermission.value = perms.system ?? true
    }
  } catch (e) {
    console.error('Failed to check permissions:', e)
  }
}

function handleSelect(service: Service) {
  selectedService.value = service
}

function handleDelete(service: Service) {
  serviceToDelete.value = service
  deleteLogsChecked.value = false
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!serviceToDelete.value) return
  try {
    const warning = await window.go.main.App.DeleteSystemService(serviceToDelete.value.name, { deleteLogs: deleteLogsChecked.value })
    if (warning) {
      actionMessage.value = warning
    }
    await loadServices()
  } catch (e) {
    console.error('Failed to delete system service:', e)
    actionMessage.value = 'Failed to delete: ' + (e instanceof Error ? e.message : String(e))
  } finally {
    showDeleteDialog.value = false
    serviceToDelete.value = null
  }
}

function openSystemSettings() {
  actionMessage.value = 'Please open System Settings > Privacy & Security > Full Disk Access and enable access for LaunchPal'
}

onMounted(() => {
  checkPermissions()
  loadServices()
})
</script>
