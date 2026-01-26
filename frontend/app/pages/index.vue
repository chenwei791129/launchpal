<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Header with search -->
    <header class="flex items-center justify-between px-4 py-3 border-b border-surface-100">
      <h1 class="text-lg font-semibold text-white">Services</h1>
      <div class="flex items-center gap-3">
        <!-- Search -->
        <div class="relative">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search services..."
            class="w-64 pl-10 pr-4 py-1.5 bg-surface-300 border border-surface-100 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-600 focus:border-transparent"
          />
        </div>

        <!-- Refresh button -->
        <button
          class="p-2 rounded-lg text-gray-400 hover:text-white hover:bg-surface-200 transition-colors"
          title="Refresh"
          @click="loadServices"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>

        <!-- Add service button -->
        <button
          @click="showCreateModal = true"
          class="flex items-center gap-2 px-3 py-1.5 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition-colors"
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
      <div class="w-8 mr-2">
        <input
          type="checkbox"
          :checked="allChecked"
          :indeterminate="someChecked"
          class="w-4 h-4 rounded border-gray-600 bg-surface-300 text-primary-600 focus:ring-primary-600 focus:ring-offset-0"
          @change="toggleAllChecked"
        />
      </div>
      <div class="w-8">Status</div>
      <div class="flex-1">Name</div>
      <div class="w-24 text-center">Type</div>
      <div class="w-24">Actions</div>
    </div>

    <!-- Services list -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="loading" class="flex items-center justify-center h-full">
        <div class="flex items-center gap-3 text-gray-400">
          <svg class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
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
          <button
            v-if="!searchQuery"
            @click="showCreateModal = true"
            class="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Create your first service
          </button>
        </div>
      </div>

      <template v-else>
        <ServiceRow
          v-for="service in filteredServices"
          :key="service.name"
          :service="service"
          :selected="selectedService?.name === service.name"
          :checked="checkedServices.has(service.name)"
          @select="handleSelect"
          @check="handleCheck"
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
          <h3 class="text-lg font-semibold text-white mb-2">Delete Service</h3>
          <p class="text-gray-400 mb-6">
            Are you sure you want to delete "{{ serviceToDelete?.label }}"? This action cannot be undone.
          </p>
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

    <!-- Create Service Modal -->
    <CreateServiceModal
      :is-open="showCreateModal"
      @close="showCreateModal = false"
      @created="loadServices"
    />
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'

const services = ref<Service[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')
const selectedService = ref<Service | null>(null)
const checkedServices = ref<Set<string>>(new Set())
const showDeleteDialog = ref(false)
const serviceToDelete = ref<Service | null>(null)
const showCreateModal = ref(false)

const updateCounts = inject<(total: number, running: number) => void>('updateCounts')

const filteredServices = computed(() => {
  if (!searchQuery.value) return services.value
  const query = searchQuery.value.toLowerCase()
  return services.value.filter(
    service =>
      service.label.toLowerCase().includes(query) ||
      service.name.toLowerCase().includes(query) ||
      service.path.toLowerCase().includes(query)
  )
})

const allChecked = computed(() => {
  return filteredServices.value.length > 0 && filteredServices.value.every(s => checkedServices.value.has(s.name))
})

const someChecked = computed(() => {
  return filteredServices.value.some(s => checkedServices.value.has(s.name)) && !allChecked.value
})

async function loadServices() {
  loading.value = true
  error.value = null

  try {
    if (window.go?.main?.App?.ListServices) {
      services.value = await window.go.main.App.ListServices()
    } else {
      // Development fallback with mock data
      services.value = []
    }

    const runningCount = services.value.filter(s => s.status === 'running').length
    updateCounts?.(services.value.length, runningCount)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load services'
    console.error('Failed to load services:', e)
  } finally {
    loading.value = false
  }
}

function handleSelect(service: Service) {
  selectedService.value = service
}

function handleCheck(service: Service) {
  if (checkedServices.value.has(service.name)) {
    checkedServices.value.delete(service.name)
  } else {
    checkedServices.value.add(service.name)
  }
}

function toggleAllChecked() {
  if (allChecked.value) {
    checkedServices.value.clear()
  } else {
    filteredServices.value.forEach(s => checkedServices.value.add(s.name))
  }
}

function handleDelete(service: Service) {
  serviceToDelete.value = service
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!serviceToDelete.value) return

  try {
    if (window.go?.main?.App?.DeleteService) {
      await window.go.main.App.DeleteService(serviceToDelete.value.name)
    }
    await loadServices()
  } catch (e) {
    console.error('Failed to delete service:', e)
  } finally {
    showDeleteDialog.value = false
    serviceToDelete.value = null
  }
}

onMounted(() => {
  loadServices()
})
</script>
