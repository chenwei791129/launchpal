<template>
  <footer class="h-8 bg-surface-500 border-t border-surface-100 flex items-center justify-between px-4 text-xs text-gray-400">
    <div class="flex items-center gap-4">
      <span>{{ serviceCount }} services</span>
      <span v-if="runningCount > 0" class="flex items-center gap-1">
        <span class="w-2 h-2 bg-green-500 rounded-full"/>
        {{ runningCount }} running
      </span>
    </div>
    <div>
      {{ appVersion }}
    </div>
  </footer>
</template>

<script setup lang="ts">
import { useAppVersion } from '~/composables/useAppVersion'

defineProps<{
  serviceCount: number
  runningCount: number
}>()

const appVersion = useAppVersion()
</script>
