<template>
  <div
    v-if="message"
    class="mx-4 mt-4 p-3 bg-surface-400 border border-surface-100 rounded-lg flex items-start gap-3"
  >
    <p class="flex-1 text-sm text-gray-300">{{ message }}</p>
    <button
      class="text-sm text-gray-400 hover:text-white transition-colors"
      @click="$emit('dismiss')"
    >
      Dismiss
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{ message: string | null }>()
defineEmits<{ dismiss: [] }>()
</script>
