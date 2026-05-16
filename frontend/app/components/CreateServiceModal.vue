<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center">
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-black/60" @click="$emit('close')"/>

    <!-- Modal -->
    <div class="relative bg-surface-300 rounded-lg shadow-xl w-full max-w-lg max-h-[80vh] overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-surface-100">
        <h2 class="text-lg font-semibold text-gray-100">{{ serviceType === 'system' ? 'Create New System Daemon' : 'Create New Service' }}</h2>
        <button class="text-gray-400 hover:text-white" @click="$emit('close')">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Form -->
      <form class="p-6 space-y-4 overflow-y-auto max-h-[60vh]" @submit.prevent="handleSubmit">
        <!-- Label -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Service Label *</label>
          <input
            v-model="form.label"
            type="text"
            placeholder="com.example.myservice"
            required
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          >
          <p class="text-xs text-gray-500 mt-1">Unique identifier (e.g., com.yourname.servicename)</p>
        </div>

        <!-- Program -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Program Path</label>
          <input
            v-model="form.program"
            type="text"
            placeholder="/usr/local/bin/myapp"
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          >
          <p class="text-xs text-gray-500 mt-1">{{ PROGRAM_PATH_HINT }}</p>
        </div>

        <!-- Arguments -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Arguments</label>
          <input
            v-model="argumentsText"
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
            v-model="form.workingDirectory"
            type="text"
            placeholder="/Users/yourname/project"
            class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
          >
        </div>

        <!-- Checkboxes -->
        <div class="flex gap-6">
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input v-model="form.runAtLoad" type="checkbox" class="rounded bg-surface-400 border-surface-100" >
            Run at Load
          </label>
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input v-model="form.keepAlive" type="checkbox" class="rounded bg-surface-400 border-surface-100" >
            Keep Alive
          </label>
        </div>

        <!-- Environment Variables -->
        <div>
          <label class="block text-sm text-gray-400 mb-1">Environment Variables</label>
          <div class="space-y-2">
            <div v-for="(env, index) in envVars" :key="index" class="flex gap-2">
              <input
                v-model="env.key"
                type="text"
                placeholder="KEY"
                class="flex-1 px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500 font-mono text-sm"
              >
              <input
                v-model="env.value"
                :type="envVisibility.has(index) ? 'text' : 'password'"
                placeholder="Value"
                class="flex-1 px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500 text-sm"
              >
              <button
                type="button"
                class="px-2 text-gray-500 hover:text-gray-300 transition-colors"
                :title="envVisibility.has(index) ? 'Hide value' : 'Show value'"
                @click="toggleEnvVisibility(index)"
              >
                <svg v-if="!envVisibility.has(index)" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
                @click="removeEnvVar(index)"
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
            @click="envVars.push({ key: '', value: '' })"
          >
            + Add
          </button>
        </div>

        <!-- Schedule -->
        <ScheduleForm v-model="schedule" v-model:wake-system="wakeSystem" />

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
          class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
          @click="$emit('close')"
        >
          Cancel
        </button>
        <button
          :disabled="loading || !form.label || !hasProgramOrArguments(form.program, argumentsText)"
          class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded transition-colors"
          @click="handleSubmit"
        >
          {{ loading ? 'Creating...' : 'Create Service' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { ServiceConfig, ScheduleConfig } from '~/types/wails.d'
import { parseShellArgs } from '~/utils/shell-args'
import { composeLogPaths } from '~/utils/logPaths'
import { hasProgramOrArguments, PROGRAM_PATH_HINT } from '~/utils/serviceValidation'
import { useSettings } from '~/composables/useSettings'

const props = withDefaults(defineProps<{
  isOpen: boolean
  serviceType?: 'user' | 'system'
}>(), {
  serviceType: 'user',
})

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
const envVars = reactive<Array<{ key: string; value: string }>>([])
const envVisibility = reactive(new Set<number>())
const schedule = ref<ScheduleConfig | undefined>(undefined)
const wakeSystem = ref(false)
const loading = ref(false)
const error = ref('')

function toggleEnvVisibility(index: number) {
  if (envVisibility.has(index)) {
    envVisibility.delete(index)
  } else {
    envVisibility.add(index)
  }
}

function removeEnvVar(index: number) {
  envVars.splice(index, 1)
  // Shift visibility indices to stay aligned after removal
  const next = new Set<number>()
  for (const i of envVisibility) {
    if (i < index) next.add(i)
    else if (i > index) next.add(i - 1)
  }
  envVisibility.clear()
  for (const i of next) envVisibility.add(i)
}

// Settings live as a module-level singleton via useSettings; we re-read them
// every time the modal opens (Decision 8) so changes saved on the Settings
// page take effect on the next New Service interaction without a restart.
const { settings, load: loadSettings } = useSettings()

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      loadSettings().catch(() => {
        // Defaults remain in place; the user can still create a service.
      })
    }
  },
  { immediate: true },
)

const logPaths = computed(() =>
  composeLogPaths(props.serviceType, settings.value, form.label),
)

async function handleSubmit() {
  if (!form.label || !hasProgramOrArguments(form.program, argumentsText.value)) return

  loading.value = true
  error.value = ''

  try {
    const environment: Record<string, string> = {}
    for (const env of envVars) {
      if (env.key.trim()) {
        environment[env.key.trim()] = env.value
      }
    }

    const config: ServiceConfig = {
      ...form,
      arguments: parseShellArgs(argumentsText.value),
      environment: Object.keys(environment).length > 0 ? environment : undefined,
      schedule: schedule.value,
      wakeSystem: wakeSystem.value,
      stdoutPath: logPaths.value.stdout,
      stderrPath: logPaths.value.stderr,
    }

    if (props.serviceType === 'system') {
      await window.go.main.App.CreateSystemService(config)
    } else {
      await window.go.main.App.CreateService(config)
    }
    emit('created')
    emit('close')

    // Reset form
    form.label = ''
    form.program = ''
    form.runAtLoad = true
    form.keepAlive = false
    form.workingDirectory = ''
    argumentsText.value = ''
    envVars.splice(0)
    envVisibility.clear()
    schedule.value = undefined
    wakeSystem.value = false
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : 'Failed to create service'
  } finally {
    loading.value = false
  }
}
</script>
