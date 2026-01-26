<template>
  <div class="p-6 space-y-4">
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Program</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ service.program || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Working Directory</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ service.workingDirectory || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Arguments</label>
        <p class="text-gray-100 mt-1 font-mono text-sm">{{ service.arguments?.join(' ') || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">PID</label>
        <p class="text-gray-100 mt-1">{{ service.pid || '-' }}</p>
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
        <label class="text-xs text-gray-400 uppercase tracking-wider">Stdout Path</label>
        <p class="text-gray-100 mt-1 font-mono text-sm truncate">{{ service.stdoutPath || '-' }}</p>
      </div>
      <div>
        <label class="text-xs text-gray-400 uppercase tracking-wider">Stderr Path</label>
        <p class="text-gray-100 mt-1 font-mono text-sm truncate">{{ service.stderrPath || '-' }}</p>
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

defineProps<{
  service: Service
}>()
</script>
