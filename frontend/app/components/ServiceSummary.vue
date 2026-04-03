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
        <p v-if="service.schedule.hasMultiple" class="text-xs text-yellow-400 mt-1">Multiple schedules defined; only the first is shown.</p>
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
      <div class="mt-2 bg-surface-300 rounded p-3 font-mono text-sm">
        <div v-for="(value, key) in service.environment" :key="key" class="text-gray-100">
          {{ key }}={{ value }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'
import { serializeShellArgs } from '~/utils/shell-args'
import { RevealInFinder } from '../../wailsjs/go/main/App'

const props = defineProps<{
  service: Service
}>()

const copiedField = ref<string | null>(null)
const revealError = ref(false)

const scheduleDisplay = computed(() => {
  const s = props.service.schedule
  if (!s) return ''
  if (s.interval !== undefined) {
    return `Every ${s.interval} seconds`
  }
  const parts: string[] = []
  if (s.month !== undefined) parts.push(`Month: ${s.month}`)
  if (s.day !== undefined) parts.push(`Day: ${s.day}`)
  if (s.weekday !== undefined) {
    const names = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
    parts.push(`Weekday: ${names[s.weekday] ?? s.weekday}`)
  }
  if (s.hour !== undefined) parts.push(`Hour: ${String(s.hour).padStart(2, '0')}`)
  if (s.minute !== undefined) parts.push(`Minute: ${String(s.minute).padStart(2, '0')}`)
  return parts.length > 0 ? parts.join(', ') : 'Every minute'
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
