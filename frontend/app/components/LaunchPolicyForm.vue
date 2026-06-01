<template>
  <div class="space-y-3">
    <!-- Launch Policy radio group -->
    <div>
      <label class="block text-sm text-gray-400 mb-1">Launch Policy</label>
      <div class="flex flex-col gap-2" data-testid="launch-policy-group">
        <label class="flex items-center gap-2 text-sm text-gray-300">
          <input
            v-model="launchPolicy"
            type="radio"
            value="onDemand"
            data-testid="launch-policy-onDemand"
            class="bg-surface-400 border-surface-100"
          >
          On Demand
        </label>
        <label class="flex items-center gap-2 text-sm text-gray-300">
          <input
            v-model="launchPolicy"
            type="radio"
            value="runAtLoad"
            data-testid="launch-policy-runAtLoad"
            class="bg-surface-400 border-surface-100"
          >
          Run at Load
        </label>
        <label class="flex items-center gap-2 text-sm text-gray-300">
          <input
            v-model="launchPolicy"
            type="radio"
            value="keepAlive"
            data-testid="launch-policy-keepAlive"
            class="bg-surface-400 border-surface-100"
          >
          Keep Alive
        </label>
      </div>
    </div>

    <!-- Advanced KeepAlive options (only when Keep Alive is selected) -->
    <div
      v-if="launchPolicy === 'keepAlive'"
      data-testid="keepalive-advanced"
      class="pl-4 border-l-2 border-surface-100 space-y-3"
    >
      <p class="text-xs text-gray-500">
        Keep Alive implies Run at Load. Multiple conditions are combined with OR semantics.
      </p>

      <!-- Boolean vs dictionary toggle -->
      <div>
        <label class="block text-sm text-gray-400 mb-1">Restart Behavior</label>
        <div class="flex gap-4">
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input
              v-model="mode"
              type="radio"
              value="boolean"
              data-testid="keepalive-mode-boolean"
              class="bg-surface-400 border-surface-100"
            >
            Always restart
          </label>
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input
              v-model="mode"
              type="radio"
              value="dictionary"
              data-testid="keepalive-mode-dictionary"
              class="bg-surface-400 border-surface-100"
            >
            Restart on conditions
          </label>
        </div>
      </div>

      <!-- Conditional sub-keys -->
      <div v-if="mode === 'dictionary'" class="space-y-2" data-testid="keepalive-conditions">
        <div v-for="cond in conditions" :key="cond.key" class="flex items-center justify-between gap-2">
          <label class="text-sm text-gray-300">{{ cond.label }}</label>
          <select
            :value="cond.model.value"
            :data-testid="`keepalive-${cond.key}`"
            class="px-2 py-1 bg-surface-400 border border-surface-100 rounded text-gray-100 text-sm"
            @change="cond.model.value = ($event.target as HTMLSelectElement).value as TriState"
          >
            <option value="unset">Not set</option>
            <option value="true">true</option>
            <option value="false">false</option>
          </select>
        </div>
      </div>

      <!-- ThrottleInterval -->
      <div>
        <label class="block text-sm text-gray-400 mb-1">Throttle Interval (seconds)</label>
        <input
          v-model="throttleText"
          type="number"
          min="0"
          placeholder="e.g. 10"
          data-testid="keepalive-throttleInterval"
          class="w-full px-3 py-2 bg-surface-400 border border-surface-100 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-primary-500"
        >
        <p class="text-xs text-gray-500 mt-1">Minimum seconds between restarts. Leave empty to use the launchd default.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import type { KeepAliveConfig } from '~/types/wails'
import type { LaunchPolicy } from '~/utils/launchPolicy'

type TriState = 'unset' | 'true' | 'false'
type BoolSubKey = 'successfulExit' | 'crashed' | 'afterInitialDemand'

const launchPolicy = defineModel<LaunchPolicy>('launchPolicy', { required: true })
const keepAlive = defineModel<KeepAliveConfig>('keepAlive', { required: true })
const throttleInterval = defineModel<number | undefined>('throttleInterval')

// When the user picks Keep Alive, make the model reflect an enabled config so
// the advanced controls bind against real state. Mode defaults to boolean
// (an unconditional "always restart") until the user opts into conditions.
watch(launchPolicy, (policy) => {
  if (policy === 'keepAlive' && !keepAlive.value.enabled) {
    keepAlive.value = { ...keepAlive.value, enabled: true, mode: keepAlive.value.mode || 'boolean' }
  }
}, { immediate: true })

const mode = computed<'boolean' | 'dictionary'>({
  get: () => (keepAlive.value.mode === 'dictionary' ? 'dictionary' : 'boolean'),
  set: (m) => { keepAlive.value = { ...keepAlive.value, enabled: true, mode: m } },
})

function triStateModel(key: BoolSubKey) {
  return computed<TriState>({
    get: () => {
      const v = keepAlive.value[key]
      return v === undefined ? 'unset' : v ? 'true' : 'false'
    },
    set: (sel) => {
      // Assign undefined rather than deleting the key: the field is optional,
      // so an undefined value is treated as "unset" by hasEffectiveKeepAliveSubKey
      // and is omitted from the JSON sent to the backend.
      keepAlive.value = { ...keepAlive.value, [key]: sel === 'unset' ? undefined : sel === 'true' }
    },
  })
}

const conditions = [
  { key: 'successfulExit' as const, label: 'Successful Exit', model: triStateModel('successfulExit') },
  { key: 'crashed' as const, label: 'Crashed', model: triStateModel('crashed') },
  { key: 'afterInitialDemand' as const, label: 'After Initial Demand', model: triStateModel('afterInitialDemand') },
]

const throttleText = computed<string>({
  get: () => (throttleInterval.value === undefined ? '' : String(throttleInterval.value)),
  // A number-typed input may hand the setter a number rather than a string,
  // so coerce before trimming.
  set: (s) => {
    const trimmed = String(s).trim()
    const n = Number.parseInt(trimmed, 10)
    throttleInterval.value = trimmed !== '' && Number.isFinite(n) && n >= 0 ? n : undefined
  },
})
</script>
