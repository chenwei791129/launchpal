<template>
  <svg
    v-if="confidence === 'unverified'"
    xmlns="http://www.w3.org/2000/svg"
    :class="['text-yellow-500 shrink-0', sizeClass]"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    :title="tooltip"
    @click.stop
  >
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
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
  '狀態偵測結果不確定；顯示的狀態與 PID（若有）可能不是該 daemon 實際對應的那個。'
</script>
