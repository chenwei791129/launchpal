<template>
  <div class="space-y-3">
    <!-- Enable Schedule -->
    <label class="flex items-center gap-2 text-sm text-gray-300">
      <input v-model="enabled" type="checkbox" class="rounded bg-surface-400 border-surface-100" />
      Enable Schedule
    </label>

    <div v-if="enabled" class="space-y-3 pl-1">
      <!-- Schedule Sub-type -->
      <div class="flex gap-2">
        <button
          type="button"
          @click="scheduleType = 'calendar'"
          class="px-3 py-1.5 text-sm rounded transition-colors"
          :class="scheduleType === 'calendar'
            ? 'bg-primary-600 text-white'
            : 'bg-surface-400 text-gray-400 hover:text-gray-200'"
        >
          Calendar Interval
        </button>
        <button
          type="button"
          @click="scheduleType = 'interval'"
          class="px-3 py-1.5 text-sm rounded transition-colors"
          :class="scheduleType === 'interval'
            ? 'bg-primary-600 text-white'
            : 'bg-surface-400 text-gray-400 hover:text-gray-200'"
        >
          Fixed Interval
        </button>
      </div>

      <!-- Calendar Interval: Cron Expression -->
      <div v-if="scheduleType === 'calendar'" class="space-y-2">
        <label class="block text-xs text-gray-500 mb-1">Schedule Expression</label>
        <input
          v-model="cronExpression"
          type="text"
          placeholder="* * * * *"
          class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500 font-mono"
        />
        <div class="flex items-center gap-4 text-xs text-gray-500">
          <span class="font-mono">minute hour day month weekday</span>
          <span>Use <code class="text-gray-400">*</code> for any, single values only (no ranges or <code class="text-gray-400">*/5</code>)</span>
        </div>

        <!-- Parsed preview -->
        <div v-if="cronExpression.trim()" class="px-3 py-2 bg-surface-500 rounded text-xs">
          <span v-if="parseError" class="text-red-400">{{ parseError }}</span>
          <span v-else class="text-gray-300">{{ cronDescription }}</span>
        </div>

        <!-- Warning: every minute -->
        <div v-if="isEveryMinute" class="flex items-start gap-2 px-3 py-2 bg-yellow-900/30 border border-yellow-700/50 rounded text-yellow-400 text-xs">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span>This will run the service <strong>every minute</strong>.</span>
        </div>
      </div>

      <!-- Fixed Interval Field -->
      <div v-if="scheduleType === 'interval'">
        <label class="block text-xs text-gray-500 mb-1">Interval (seconds)</label>
        <input
          v-model.number="intervalSeconds"
          type="number"
          min="10"
          placeholder="3600"
          class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
        />
        <p class="text-xs text-gray-500 mt-1">Minimum 10 seconds. Common values: 60 (1min), 3600 (1hr), 86400 (1day)</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ScheduleConfig } from '~/types/wails.d'

const props = defineProps<{
  modelValue?: ScheduleConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ScheduleConfig | undefined]
}>()

const weekdayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const enabled = ref(false)
const scheduleType = ref<'calendar' | 'interval'>('calendar')
const intervalSeconds = ref<number | undefined>(undefined)
const cronExpression = ref('* * * * *')

const parseError = ref('')

interface ParsedCron {
  minute?: number
  hour?: number
  day?: number
  month?: number
  weekday?: number
}

function parseCron(expr: string): ParsedCron | null {
  parseError.value = ''
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) {
    parseError.value = 'Expected 5 fields: minute hour day month weekday'
    return null
  }

  const limits = [
    ['Minute', 0, 59],
    ['Hour', 0, 23],
    ['Day', 1, 31],
    ['Month', 1, 12],
    ['Weekday', 0, 6],
  ] as const

  const values: (number | undefined)[] = []
  for (let i = 0; i < 5; i++) {
    const field = parts[i]
    const limit = limits[i]!
    if (field === '*') {
      values.push(undefined)
    } else {
      const num = Number(field)
      if (!Number.isInteger(num) || num < limit[1] || num > limit[2]) {
        parseError.value = `${limit[0]}: expected * or ${limit[1]}-${limit[2]}, got "${field}"`
        return null
      }
      values.push(num)
    }
  }

  return {
    minute: values[0],
    hour: values[1],
    day: values[2],
    month: values[3],
    weekday: values[4],
  }
}

function configToCron(config: ScheduleConfig): string {
  const f = (v: number | undefined) => v !== undefined ? String(v) : '*'
  return `${f(config.minute)} ${f(config.hour)} ${f(config.day)} ${f(config.month)} ${f(config.weekday)}`
}

const parsedCron = computed(() => parseCron(cronExpression.value))

const isEveryMinute = computed(() => {
  if (!parsedCron.value || parseError.value) return false
  const p = parsedCron.value
  return p.minute === undefined && p.hour === undefined && p.day === undefined && p.month === undefined && p.weekday === undefined
})

const cronDescription = computed(() => {
  const p = parsedCron.value
  if (!p || parseError.value) return ''

  const parts: string[] = []
  if (p.minute !== undefined) parts.push(`at minute ${String(p.minute).padStart(2, '0')}`)
  if (p.hour !== undefined) parts.push(`at hour ${String(p.hour).padStart(2, '0')}`)
  if (p.day !== undefined) parts.push(`on day ${p.day}`)
  if (p.month !== undefined) parts.push(`in month ${p.month}`)
  if (p.weekday !== undefined) parts.push(`on ${weekdayNames[p.weekday]}`)

  return parts.length > 0 ? parts.join(', ') : 'Every minute'
})

// Guard to prevent watch loop between modelValue and emit watchers
let updatingFromProp = false

// Initialize from modelValue
watch(() => props.modelValue, (val) => {
  updatingFromProp = true
  if (val) {
    enabled.value = true
    if (val.interval !== undefined) {
      scheduleType.value = 'interval'
      intervalSeconds.value = val.interval
    } else {
      scheduleType.value = 'calendar'
      const newCron = configToCron(val)
      if (cronExpression.value !== newCron) {
        cronExpression.value = newCron
      }
    }
  } else {
    enabled.value = false
  }
  nextTick(() => { updatingFromProp = false })
}, { immediate: true })

// Emit changes
watch([enabled, scheduleType, intervalSeconds, cronExpression], () => {
  if (updatingFromProp) return

  if (!enabled.value) {
    emit('update:modelValue', undefined)
    return
  }

  if (scheduleType.value === 'interval') {
    if (intervalSeconds.value && intervalSeconds.value >= 10) {
      emit('update:modelValue', { interval: intervalSeconds.value })
    } else {
      emit('update:modelValue', undefined)
    }
  } else {
    const parsed = parsedCron.value
    if (!parsed || parseError.value) return
    const config: ScheduleConfig = {}
    if (parsed.minute !== undefined) config.minute = parsed.minute
    if (parsed.hour !== undefined) config.hour = parsed.hour
    if (parsed.day !== undefined) config.day = parsed.day
    if (parsed.month !== undefined) config.month = parsed.month
    if (parsed.weekday !== undefined) config.weekday = parsed.weekday
    emit('update:modelValue', config)
  }
})
</script>
