<template>
  <div
    class="group flex items-center px-4 py-3 hover:bg-surface-200 cursor-pointer border-b border-surface-100 transition-colors"
    :class="{ 'bg-surface-200': selected }"
    @click="navigateTo(`/services/${service.name}?type=${service.type}`)"
  >
    <!-- Status indicator -->
    <div class="w-16 shrink-0 flex items-center justify-center gap-1">
      <span
        class="w-2.5 h-2.5 rounded-full"
        :class="{
          'bg-green-500': service.status === 'running',
          'bg-blue-500': service.status === 'loaded',
          'bg-gray-500': service.status === 'stopped',
          'bg-yellow-500': service.status === 'unknown'
        }"
      ></span>
      <StatusConfidenceIcon :confidence="service.statusConfidence" size="sm" />
    </div>

    <!-- Service info -->
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2">
        <span class="text-white font-medium truncate">{{ service.label }}</span>
        <span v-if="service.pid" class="text-xs text-gray-500">PID {{ service.pid }}</span>
        <span
          v-if="service.readOnly"
          class="inline-flex items-center px-1.5 py-0.5 rounded text-xs bg-gray-600/30 text-gray-400"
        >
          Read-only
        </span>
      </div>
      <div class="text-xs text-gray-500 truncate">{{ service.path }}</div>
    </div>

    <!-- Schedule badge -->
    <div class="w-24 shrink-0 text-center">
      <span
        v-if="service.schedule"
        class="inline-flex items-center px-2 py-0.5 rounded text-xs bg-purple-600/20 text-purple-400"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Scheduled
      </span>
      <span
        v-else-if="service.runAtLoad"
        class="inline-flex items-center px-2 py-0.5 rounded text-xs bg-blue-600/20 text-blue-400"
      >
        RunAtLoad
      </span>
    </div>

    <!-- Action buttons -->
    <div class="w-24 shrink-0 flex items-center gap-1">
      <template v-if="!service.readOnly">
        <!-- Start/Stop button -->
        <button
          v-if="service.status === 'running' || service.status === 'loaded'"
          class="p-1.5 rounded hover:bg-red-600/20 text-gray-400 hover:text-red-400 transition-colors"
          title="Stop service"
          @click.stop="handleStop"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
          </svg>
        </button>
        <button
          v-else
          class="p-1.5 rounded hover:bg-green-600/20 text-gray-400 hover:text-green-400 transition-colors"
          title="Start service"
          @click.stop="handleStart"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>

        <!-- Delete button -->
        <button
          class="p-1.5 rounded hover:bg-red-600/20 text-gray-400 hover:text-red-400 transition-colors"
          title="Delete service"
          @click.stop="$emit('delete', service)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </template>

      <!-- Info button for read-only services -->
      <button
        v-else
        class="p-1.5 rounded hover:bg-surface-100 text-gray-400 hover:text-white transition-colors"
        title="View details"
        @click.stop="navigateTo(`/services/${service.name}?type=${service.type}`)"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Service } from '~/types/wails'

const props = defineProps<{
  service: Service
  selected?: boolean
}>()

const emit = defineEmits<{
  select: [service: Service]
  delete: [service: Service]
  refresh: []
}>()

async function handleStart() {
  try {
    await window.go.main.App.StartService(props.service.name)
    emit('refresh')
  } catch (error) {
    console.error('Failed to start service:', error)
  }
}

async function handleStop() {
  try {
    await window.go.main.App.StopService(props.service.name)
    emit('refresh')
  } catch (error) {
    console.error('Failed to stop service:', error)
  }
}
</script>
