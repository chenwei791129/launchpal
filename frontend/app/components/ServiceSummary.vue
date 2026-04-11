<template>
  <div class="p-6 space-y-4">
    <!-- Plist Path -->
    <div>
      <label class="text-xs text-gray-400 uppercase tracking-wider">Plist File</label>
      <div
        class="mt-1 flex items-start gap-2 px-3 py-2 bg-surface-400 rounded cursor-pointer hover:bg-surface-300 transition-colors group min-w-0"
        @click="copyPath"
      >
        <code class="flex-1 text-gray-100 font-mono text-sm break-words overflow-wrap-anywhere min-w-0">{{ service.path }}</code>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-4 h-4 text-gray-500 group-hover:text-gray-300 flex-shrink-0 mt-0.5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          title="Copy path"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-4 h-4 text-gray-500 group-hover:text-gray-300 flex-shrink-0 mt-0.5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          title="Reveal in Finder"
          @click.stop="revealInFinder"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z" />
        </svg>
      </div>
      <p v-if="copiedField === 'path'" class="text-xs text-primary-400 mt-1">Copied!</p>
      <p v-if="revealError" class="text-xs text-red-400 mt-1">Failed to reveal in Finder</p>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Program</label>
        <p class="text-gray-100 mt-1 font-mono text-sm break-words overflow-wrap-anywhere">{{ service.program || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Working Directory</label>
        <p class="text-gray-100 mt-1 font-mono text-sm break-words overflow-wrap-anywhere">{{ service.workingDirectory || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Arguments</label>
        <p class="text-gray-100 mt-1 font-mono text-sm break-words overflow-wrap-anywhere">{{ serializeShellArgs(service.arguments ?? []) || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">PID</label>
        <p class="text-gray-100 mt-1">{{ service.pid || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Plist Format</label>
        <p class="text-gray-100 mt-1">
          <span
            :class="{
              'px-2 py-0.5 rounded text-xs font-medium': true,
              'bg-blue-500/20 text-blue-400': service.plistFormat === 'xml',
              'bg-purple-500/20 text-purple-400': service.plistFormat === 'binary',
              'bg-gray-500/20 text-gray-400': service.plistFormat === 'unknown'
            }"
          >
            {{ service.plistFormat?.toUpperCase() || 'UNKNOWN' }}
          </span>
        </p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Run At Load</label>
        <p class="text-gray-100 mt-1">{{ service.runAtLoad ? 'Yes' : 'No' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Keep Alive</label>
        <p class="text-gray-100 mt-1">{{ service.keepAlive ? 'Yes' : 'No' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Wake System</label>
        <p class="text-gray-100 mt-1">{{ service.wakeSystem ? 'Yes' : 'No' }}</p>
      </div>
      <div v-if="service.schedule">
        <label class="text-xs text-gray-400 uppercase tracking-wider">Schedule</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ scheduleDisplay }}</p>
        <div v-if="nextRuns.length > 0" class="mt-2 text-xs space-y-0.5">
          <div class="text-gray-400">Next runs ({{ timezone }}):</div>
          <div v-for="(run, i) in nextRuns" :key="i" class="text-gray-300 font-mono">
            {{ run }}
          </div>
        </div>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Stdout Path</label>
        <div
          v-if="service.stdoutPath"
          class="mt-1 flex items-start gap-2 px-2 py-1 bg-surface-400 rounded cursor-pointer hover:bg-surface-300 transition-colors group min-w-0"
          title="Click to copy"
          @click="copyText(service.stdoutPath, 'stdout')"
        >
          <span class="flex-1 text-gray-100 font-mono text-sm break-words overflow-wrap-anywhere min-w-0">{{ service.stdoutPath }}</span>
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3 text-gray-500 group-hover:text-gray-300 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </div>
        <p v-else class="text-gray-100 mt-1 font-mono text-sm">-</p>
        <p v-if="copiedField === 'stdout'" class="text-xs text-primary-400 mt-1">Copied!</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Stderr Path</label>
        <div
          v-if="service.stderrPath"
          class="mt-1 flex items-start gap-2 px-2 py-1 bg-surface-400 rounded cursor-pointer hover:bg-surface-300 transition-colors group min-w-0"
          title="Click to copy"
          @click="copyText(service.stderrPath, 'stderr')"
        >
          <span class="flex-1 text-gray-100 font-mono text-sm break-words overflow-wrap-anywhere min-w-0">{{ service.stderrPath }}</span>
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3 text-gray-500 group-hover:text-gray-300 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </div>
        <p v-else class="text-gray-100 mt-1 font-mono text-sm">-</p>
        <p v-if="copiedField === 'stderr'" class="text-xs text-primary-400 mt-1">Copied!</p>
      </div>
    </div>

    <div v-if="service.environment && Object.keys(service.environment).length > 0">
      <label class="text-xs text-gray-400 uppercase tracking-wider">Environment Variables</label>
      <div class="mt-2 bg-surface-300 rounded p-3 font-mono text-sm space-y-1">
        <div
          v-for="(value, key) in service.environment"
          :key="key"
          class="flex items-center gap-2 text-gray-100"
          data-testid="env-var-row"
        >
          <span class="flex-1 break-all">{{ key }}={{ revealedEnvKeys.has(key as string) ? value : '••••••••' }}</span>
          <button
            type="button"
            class="flex-shrink-0 text-gray-500 hover:text-gray-300 transition-colors"
            :title="revealedEnvKeys.has(key as string) ? 'Hide value' : 'Reveal value'"
            data-testid="env-var-toggle"
            @click="toggleEnvKey(key as string)"
          >
            <svg
              v-if="!revealedEnvKeys.has(key as string)"
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
            <svg
              v-else
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'
import { serializeShellArgs } from '~/utils/shell-args'
import { RevealInFinder } from '../../wailsjs/go/main/App'
import { getNextOccurrences, formatDateTime, WEEKDAY_NAMES } from '~/composables/useNextOccurrences'

const props = defineProps<{
  service: Service
}>()

const copiedField = ref<string | null>(null)
const revealError = ref(false)
const revealedEnvKeys = reactive(new Set<string>())

watch(() => props.service, () => {
  revealedEnvKeys.clear()
})

function toggleEnvKey(key: string) {
  if (revealedEnvKeys.has(key)) {
    revealedEnvKeys.delete(key)
  } else {
    revealedEnvKeys.add(key)
  }
}

const nextRuns = computed(() => {
  const s = props.service.schedule
  if (!s || s.interval !== undefined) return []
  return getNextOccurrences(s, 3).map(formatDateTime)
})

const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

function describeEntry(e: { minute?: number, hour?: number, day?: number, weekday?: number, month?: number }): string {
  const parts: string[] = []
  if (e.month !== undefined) parts.push(`Month: ${e.month}`)
  if (e.day !== undefined) parts.push(`Day: ${e.day}`)
  if (e.weekday !== undefined) {
    parts.push(`Weekday: ${WEEKDAY_NAMES[e.weekday] ?? e.weekday}`)
  }
  if (e.hour !== undefined) parts.push(`Hour: ${String(e.hour).padStart(2, '0')}`)
  if (e.minute !== undefined) parts.push(`Minute: ${String(e.minute).padStart(2, '0')}`)
  return parts.length > 0 ? parts.join(', ') : 'Every minute'
}

const scheduleDisplay = computed(() => {
  const s = props.service.schedule
  if (!s) return ''
  if (s.interval !== undefined) {
    return `Every ${s.interval} seconds`
  }
  const schedules = s.schedules ?? []
  if (schedules.length === 0) return 'Every minute'
  if (schedules.length === 1) return describeEntry(schedules[0]!)
  return `${schedules.length} schedules`
})

async function copyText(text: string, field: string) {
  try {
    await navigator.clipboard.writeText(text)
    copiedField.value = field
    setTimeout(() => {
      copiedField.value = null
    }, 2000)
  } catch (e) {
    console.error('Failed to copy:', e)
  }
}

function copyPath() {
  copyText(props.service.path, 'path')
}

async function revealInFinder() {
  try {
    await RevealInFinder(props.service.path)
  } catch (e) {
    console.error('Failed to reveal in Finder:', e)
    revealError.value = true
    setTimeout(() => {
      revealError.value = false
    }, 2000)
  }
}
</script>
