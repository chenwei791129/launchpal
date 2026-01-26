<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center">
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-black/60" @click="$emit('close')"></div>

    <!-- Modal -->
    <div class="relative bg-surface-300 rounded-lg shadow-xl w-full max-w-lg max-h-[80vh] overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-surface-100">
        <h2 class="text-lg font-semibold text-gray-100">Create New Service</h2>
        <button @click="$emit('close')" class="text-gray-400 hover:text-white">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="p-6 space-y-4 overflow-y-auto max-h-[60vh]">
        <!-- Label -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Service Label *</label>
          <input
            v-model="form.label"
            type="text"
            placeholder="com.example.myservice"
            required
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          />
          <p class="text-xs text-gray-500 mt-1">Unique identifier (e.g., com.yourname.servicename)</p>
        </div>

        <!-- Program -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Program Path *</label>
          <input
            v-model="form.program"
            type="text"
            placeholder="/usr/local/bin/myapp"
            required
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          />
        </div>

        <!-- Arguments -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Arguments</label>
          <input
            v-model="argumentsText"
            type="text"
            placeholder="--daemon --port=8080"
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          />
          <p class="text-xs text-gray-500 mt-1">Space-separated arguments</p>
        </div>

        <!-- Working Directory -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Working Directory</label>
          <input
            v-model="form.workingDirectory"
            type="text"
            placeholder="/Users/yourname/project"
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          />
        </div>

        <!-- Checkboxes -->
        <div class="flex gap-6">
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input v-model="form.runAtLoad" type="checkbox" class="rounded bg-surface-400 border-surface-100" />
            Run at Load
          </label>
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input v-model="form.keepAlive" type="checkbox" class="rounded bg-surface-400 border-surface-100" />
            Keep Alive
          </label>
        </div>

        <!-- Log Paths (Auto-generated) -->
        <div v-if="form.label">
          <label class="block text-sm text-gray-400 mb-1">Log Paths</label>
          <div class="space-y-1">
            <div class="px-3 py-2 bg-surface-500 rounded text-gray-400 text-sm font-mono truncate">
              stdout: {{ logPaths.stdout }}
            </div>
            <div class="px-3 py-2 bg-surface-500 rounded text-gray-400 text-sm font-mono truncate">
              stderr: {{ logPaths.stderr }}
            </div>
          </div>
          <p class="text-xs text-gray-500 mt-1">Log paths are auto-generated based on service label</p>
        </div>

        <!-- Error Message -->
        <div v-if="error" class="text-red-400 text-sm">{{ error }}</div>
      </form>

      <!-- Footer -->
      <div class="flex justify-end gap-3 px-6 py-4 border-t border-surface-100">
        <button
          type="button"
          @click="$emit('close')"
          class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
        >
          Cancel
        </button>
        <button
          @click="handleSubmit"
          :disabled="loading || !form.label || !form.program"
          class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded transition-colors"
        >
          {{ loading ? 'Creating...' : 'Create Service' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ServiceConfig } from '~/types/wails.d'

defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  close: []
  created: []
}>()

const form = reactive({
  label: '',
  program: '',
  arguments: [] as string[],
  runAtLoad: true,
  keepAlive: false,
  workingDirectory: '',
})

const argumentsText = ref('')
const loading = ref(false)
const error = ref('')

const logPaths = computed(() => ({
  stdout: `~/Library/Logs/${form.label}/stdout.log`,
  stderr: `~/Library/Logs/${form.label}/stderr.log`,
}))

async function handleSubmit() {
  if (!form.label || !form.program) return

  loading.value = true
  error.value = ''

  try {
    const config: ServiceConfig = {
      ...form,
      arguments: argumentsText.value ? argumentsText.value.split(/\s+/).filter(Boolean) : [],
      stdoutPath: logPaths.value.stdout,
      stderrPath: logPaths.value.stderr,
    }

    await window.go.main.App.CreateService(config)
    emit('created')
    emit('close')

    // Reset form
    form.label = ''
    form.program = ''
    form.runAtLoad = true
    form.keepAlive = false
    form.workingDirectory = ''
    argumentsText.value = ''
  } catch (err: any) {
    error.value = err.message || 'Failed to create service'
  } finally {
    loading.value = false
  }
}
</script>
