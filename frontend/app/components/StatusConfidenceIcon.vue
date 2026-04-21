<template>
  <span
    v-if="confidence === 'unverified'"
    :title="tooltip"
    class="inline-flex shrink-0"
    @click.stop
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      :class="['text-yellow-500', sizeClass]"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <title>{{ tooltip }}</title>
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  </span>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    confidence?: 'verified' | 'unverified'
    size?: 'sm' | 'md'
  }>(),
  { size: 'sm' },
)

const sizeClass = computed(() => (props.size === 'md' ? 'w-4 h-4' : 'w-3.5 h-3.5'))

const tooltip =
  'Status detection is not fully certain: the status and PID shown may not correspond to the exact process launchd started for this daemon.'
</script>
