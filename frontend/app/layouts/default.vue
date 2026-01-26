<template>
  <div class="h-screen flex flex-col">
    <!-- Title bar drag region -->
    <div class="h-8 wails-drag bg-surface-500 flex items-center justify-center border-b border-surface-100">
      <span class="text-xs text-gray-400">LaunchPal</span>
    </div>

    <div class="flex-1 flex overflow-hidden">
      <Sidebar />
      <main class="flex-1 flex flex-col overflow-hidden">
        <slot />
      </main>
    </div>

    <StatusBar :service-count="serviceCount" :running-count="runningCount" />
  </div>
</template>

<script setup lang="ts">
const serviceCount = ref(0)
const runningCount = ref(0)

provide('updateCounts', (total: number, running: number) => {
  serviceCount.value = total
  runningCount.value = running
})
</script>
