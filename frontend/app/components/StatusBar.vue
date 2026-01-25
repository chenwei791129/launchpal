<template>
  <footer class="h-8 bg-surface-500 border-t border-surface-100 flex items-center justify-between px-4 text-xs text-gray-400">
    <div class="flex items-center gap-4">
      <span>{{ serviceCount }} services</span>
      <span v-if="runningCount > 0" class="flex items-center gap-1">
        <span class="w-2 h-2 bg-green-500 rounded-full"></span>
        {{ runningCount }} running
      </span>
    </div>
    <div>
      v0.1.0
    </div>
  </footer>
</template>

<script setup lang="ts">
defineProps<{
  serviceCount: number
  runningCount: number
}>()
</script>
