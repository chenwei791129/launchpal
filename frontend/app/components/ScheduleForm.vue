<template>
  <div class="space-y-3">
    <!-- Enable Schedule -->
    <label class="flex items-center gap-2 text-sm text-gray-300">
      <input v-model="enabled" type="checkbox" class="rounded bg-surface-400 border-surface-100" >
      Enable Schedule
    </label>

    <div v-if="enabled" class="space-y-3 pl-1">
      <!-- Schedule Sub-type -->
      <div class="flex gap-2">
        <button
          type="button"
          class="px-3 py-1.5 text-sm rounded transition-colors"
          :class="scheduleType === 'calendar'
            ? 'bg-primary-600 text-white'
            : 'bg-surface-400 text-gray-400 hover:text-gray-200'"
          @click="scheduleType = 'calendar'"
        >
          Calendar Interval
        </button>
        <button
          type="button"
          class="px-3 py-1.5 text-sm rounded transition-colors"
          :class="scheduleType === 'interval'
            ? 'bg-primary-600 text-white'
            : 'bg-surface-400 text-gray-400 hover:text-gray-200'"
          @click="scheduleType = 'interval'"
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
        >
        <div class="flex items-center gap-4 text-xs text-gray-500">
          <span class="font-mono">minute hour day month weekday</span>
          <span>Use <code class="text-gray-400">*</code> for any, <code class="text-gray-400">N</code> single, <code class="text-gray-400">a-b</code> range, <code class="text-gray-400">a,b,c</code> list</span>
        </div>

        <!-- Parsed preview -->
        <div v-if="cronExpression.trim()" class="px-3 py-2 bg-surface-500 rounded text-xs">
          <span v-if="parseError" class="text-red-400">{{ parseError }}</span>
          <template v-else-if="parsedCron && parsedCron.length > 1">
            <div
              class="text-gray-300 cursor-pointer select-none flex items-center gap-1"
              @click="expandedPreview = !expandedPreview"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="w-3 h-3 transition-transform"
                :class="{ 'rotate-90': expandedPreview }"
                fill="none" viewBox="0 0 24 24" stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              {{ cronDescription }}
            </div>
            <div v-if="expandedPreview" class="mt-1 pl-4 space-y-0.5 text-gray-400">
              <div v-for="(desc, i) in expandedEntries" :key="i" class="font-mono">{{ desc }}</div>
            </div>
          </template>
          <span v-else class="text-gray-300">{{ cronDescription }}</span>
        </div>

        <!-- Warning: every minute -->
        <div v-if="isEveryMinute" class="flex items-start gap-2 px-3 py-2 bg-yellow-900/30 border border-yellow-700/50 rounded text-yellow-400 text-xs">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span>This will run the service <strong>every minute</strong>.</span>
        </div>

        <!-- Warning: lossy round-trip -->
        <div v-if="cronRoundTripWarning" class="flex items-start gap-2 px-3 py-2 bg-yellow-900/30 border border-yellow-700/50 rounded text-yellow-400 text-xs">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span>This schedule was created externally with entries that cannot be exactly represented in cron syntax. Saving may alter the schedule.</span>
        </div>

        <!-- Next runs preview -->
        <div v-if="nextRunsPreview.length > 0" class="px-3 py-2 bg-surface-500 rounded text-xs space-y-1">
          <div class="text-gray-400">Next runs ({{ timezone }}):</div>
          <div v-for="(run, i) in nextRunsPreview" :key="i" class="text-gray-300 font-mono">
            {{ run }}
          </div>
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
        >
        <p class="text-xs text-gray-500 mt-1">Minimum 10 seconds. Common values: 60 (1min), 3600 (1hr), 86400 (1day)</p>
      </div>

      <!-- Wake System -->
      <label class="flex items-center gap-2 text-sm text-gray-300">
        <input v-model="wakeSystemLocal" type="checkbox" class="rounded bg-surface-400 border-surface-100" >
        Wake System
      </label>
      <p class="text-xs text-gray-500 -mt-2 pl-1">Wake the Mac from sleep when the scheduled time arrives</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ScheduleConfig, CalendarEntry } from '~/types/wails.d'
import { getNextOccurrences, formatDateTime, WEEKDAY_NAMES } from '~/composables/useNextOccurrences'

const props = defineProps<{
  modelValue?: ScheduleConfig
  wakeSystem?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ScheduleConfig | undefined]
  'update:wakeSystem': [value: boolean]
}>()

const enabled = ref(false)
const scheduleType = ref<'calendar' | 'interval'>('calendar')
const intervalSeconds = ref<number | undefined>(undefined)
const cronExpression = ref('* * * * *')
const wakeSystemLocal = computed({
  get: () => props.wakeSystem ?? false,
  set: (val) => emit('update:wakeSystem', val),
})

const parseError = ref('')
const expandedPreview = ref(false)
const cronRoundTripWarning = ref(false)

const MAX_EXPANSION = 50

function parseField(field: string, name: string, min: number, max: number): number[] | null {
  if (field === '*') return []

  if (field.includes('-')) {
    const rangeParts = field.split('-')
    if (rangeParts.length !== 2 || rangeParts[0] === '' || rangeParts[1] === '') {
      parseError.value = `${name}: invalid range "${field}"`
      return null
    }
    const a = Number(rangeParts[0])
    const b = Number(rangeParts[1])
    if (!Number.isInteger(a) || !Number.isInteger(b)) {
      parseError.value = `${name}: invalid range "${field}"`
      return null
    }
    if (a < min || a > max || b < min || b > max) {
      parseError.value = `${name}: value out of range ${min}-${max} in "${field}"`
      return null
    }
    if (a > b) {
      parseError.value = `${name}: start must be ≤ end in "${field}"`
      return null
    }
    const result: number[] = []
    for (let i = a; i <= b; i++) result.push(i)
    return result
  }

  if (field.includes(',')) {
    const parts = field.split(',')
    const seen = new Set<number>()
    const result: number[] = []
    for (const p of parts) {
      if (p === '') {
        parseError.value = `${name}: invalid list "${field}"`
        return null
      }
      const num = Number(p)
      if (!Number.isInteger(num) || num < min || num > max) {
        parseError.value = `${name}: value out of range ${min}-${max} in "${field}"`
        return null
      }
      if (!seen.has(num)) {
        seen.add(num)
        result.push(num)
      }
    }
    return result
  }

  const num = Number(field)
  if (!Number.isInteger(num) || num < min || num > max) {
    parseError.value = `${name}: expected *, ${min}-${max}, range, or list, got "${field}"`
    return null
  }
  return [num]
}

function parseCron(expr: string): CalendarEntry[] | null {
  parseError.value = ''
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) {
    parseError.value = 'Expected 5 fields: minute hour day month weekday'
    return null
  }

  const limits: [string, number, number][] = [
    ['Minute', 0, 59],
    ['Hour', 0, 23],
    ['Day', 1, 31],
    ['Month', 1, 12],
    ['Weekday', 0, 6],
  ]

  const fieldValues: number[][] = []
  for (let i = 0; i < 5; i++) {
    const result = parseField(parts[i]!, limits[i]![0], limits[i]![1], limits[i]![2])
    if (result === null) return null
    fieldValues.push(result)
  }

  const expansionCount = fieldValues.reduce((acc, f) => acc * (f.length || 1), 1)
  if (expansionCount > MAX_EXPANSION) {
    parseError.value = `Expansion produces ${expansionCount} entries, exceeding limit of ${MAX_EXPANSION}`
    return null
  }

  const entries: CalendarEntry[] = []
  const keys = ['minute', 'hour', 'day', 'month', 'weekday'] as const

  function expand(depth: number, current: Partial<CalendarEntry>) {
    if (depth === 5) {
      entries.push({ ...current } as CalendarEntry)
      return
    }
    const vals = fieldValues[depth]!
    if (vals.length === 0) {
      // Wildcard field — omit from entry
      expand(depth + 1, current)
    } else {
      for (const v of vals) {
        expand(depth + 1, { ...current, [keys[depth] as string]: v })
      }
    }
  }
  expand(0, {})

  return entries
}

function schedulesToCron(schedules: CalendarEntry[]): string {
  if (schedules.length === 0) return '* * * * *'

  if (schedules.length === 1) {
    const s = schedules[0]!
    const f = (v: number | undefined) => v !== undefined ? String(v) : '*'
    return `${f(s.minute)} ${f(s.hour)} ${f(s.day)} ${f(s.month)} ${f(s.weekday)}`
  }

  const fields = ['minute', 'hour', 'day', 'month', 'weekday'] as const
  const parts: string[] = []
  for (const field of fields) {
    const values = [...new Set(schedules.map(s => s[field]).filter((v): v is number => v !== undefined))].sort((a, b) => a - b)
    if (values.length === 0) {
      parts.push('*')
    } else if (values.length === 1) {
      parts.push(String(values[0]))
    } else {
      const isRange = values.every((v, i) => i === 0 || v === values[i - 1]! + 1)
      if (isRange) {
        parts.push(`${values[0]}-${values[values.length - 1]}`)
      } else {
        parts.push(values.join(','))
      }
    }
  }
  return parts.join(' ')
}

function entriesToScheduleConfig(entries: CalendarEntry[]): ScheduleConfig {
  return { schedules: entries.map(e => ({ ...e })) }
}

const parsedCron = computed(() => parseCron(cronExpression.value))

const isEveryMinute = computed(() => {
  const entries = parsedCron.value
  if (!entries || parseError.value || entries.length !== 1) return false
  const e = entries[0]!
  return e.minute === undefined && e.hour === undefined && e.day === undefined && e.month === undefined && e.weekday === undefined
})

const cronDescription = computed(() => {
  const entries = parsedCron.value
  if (!entries || parseError.value) return ''

  if (entries.length <= 1) {
    const p = entries[0]
    if (!p) return 'Every minute'
    const parts: string[] = []
    if (p.minute !== undefined) parts.push(`at minute ${String(p.minute).padStart(2, '0')}`)
    if (p.hour !== undefined) parts.push(`at hour ${String(p.hour).padStart(2, '0')}`)
    if (p.day !== undefined) parts.push(`on day ${p.day}`)
    if (p.month !== undefined) parts.push(`in month ${p.month}`)
    if (p.weekday !== undefined) parts.push(`on ${WEEKDAY_NAMES[p.weekday]}`)
    return parts.length > 0 ? parts.join(', ') : 'Every minute'
  }

  // Build semantic description from the cron expression fields
  const cron = cronExpression.value.trim().split(/\s+/)
  const fieldNames = ['minute', 'hour', 'day', 'month', 'weekday']
  const desc: string[] = []
  for (let i = 0; i < 5; i++) {
    const f = cron[i]
    if (f && f !== '*') desc.push(`${fieldNames[i]} ${f}`)
  }
  return `${entries.length} schedules: ${desc.join(', ')}`
})

const expandedEntries = computed(() => {
  const entries = parsedCron.value
  if (!entries || parseError.value) return []
  return entries.map(e => {
    const hh = e.hour !== undefined ? String(e.hour).padStart(2, '0') : '*'
    const mm = e.minute !== undefined ? String(e.minute).padStart(2, '0') : '*'
    const parts = [`${hh}:${mm}`]
    if (e.day !== undefined) parts.push(`day ${e.day}`)
    if (e.month !== undefined) parts.push(`month ${e.month}`)
    if (e.weekday !== undefined) parts.push(WEEKDAY_NAMES[e.weekday] ?? `weekday ${e.weekday}`)
    return parts.join(', ')
  })
})

const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const nextRunsPreview = computed(() => {
  if (scheduleType.value !== 'calendar') return []
  const entries = parsedCron.value
  if (!entries || parseError.value || entries.length === 0) return []
  return getNextOccurrences(entriesToScheduleConfig(entries), 3).map(formatDateTime)
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
      const origSchedules = val.schedules ?? []
      const newCron = schedulesToCron(origSchedules)
      if (cronExpression.value !== newCron) {
        cronExpression.value = newCron
      }
      // Detect lossy round-trip after cron expression settles
      const origLen = origSchedules.length
      nextTick(() => {
        if (origLen > 1) {
          const reparsed = parsedCron.value
          cronRoundTripWarning.value = !!(reparsed && reparsed.length !== origLen)
        } else {
          cronRoundTripWarning.value = false
        }
        updatingFromProp = false
      })
      return
    }
  } else {
    enabled.value = false
  }
  nextTick(() => { updatingFromProp = false })
}, { immediate: true })

// Emit changes
watch([enabled, scheduleType, intervalSeconds, cronExpression], () => {
  if (updatingFromProp) return

  cronRoundTripWarning.value = false

  if (!enabled.value) {
    emit('update:modelValue', undefined)
    emit('update:wakeSystem', false)
    return
  }

  if (scheduleType.value === 'interval') {
    if (intervalSeconds.value && intervalSeconds.value >= 10) {
      emit('update:modelValue', { interval: intervalSeconds.value })
    } else {
      emit('update:modelValue', undefined)
    }
  } else {
    const entries = parsedCron.value
    if (!entries || parseError.value) return
    emit('update:modelValue', entriesToScheduleConfig(entries))
  }
})
</script>
