<template>
  <div class="relative">
    <button
      :data-testid="`${testid}-filter-trigger`"
      type="button"
      aria-haspopup="listbox"
      :aria-expanded="open"
      :aria-controls="`${testid}-filter-menu`"
      class="flex items-center gap-1.5 px-3 py-1 rounded-lg border text-sm transition-colors"
      :class="selected.length
        ? 'bg-primary-600/20 border-primary-600/40 text-primary-400'
        : 'bg-surface-300 border-surface-100 text-gray-400 hover:text-white'"
      @click="$emit('toggleMenu')"
    >
      <span>{{ label }}: {{ summary }}</span>
      <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <div
      v-if="open"
      :id="`${testid}-filter-menu`"
      :data-testid="`${testid}-filter-menu`"
      role="listbox"
      aria-multiselectable="true"
      :aria-label="`${label} filter`"
      class="absolute left-0 z-20 mt-1 w-44 py-1 bg-surface-400 border border-surface-100 rounded-lg shadow-xl"
    >
      <button
        v-for="option in options"
        :key="option.value"
        :data-testid="`${testid}-option-${option.value}`"
        type="button"
        role="option"
        :aria-selected="selected.includes(option.value)"
        class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left text-gray-300 hover:bg-surface-200 hover:text-white transition-colors"
        @click="$emit('select', option.value)"
      >
        <span
          aria-hidden="true"
          class="w-4 h-4 shrink-0 rounded border flex items-center justify-center"
          :class="selected.includes(option.value)
            ? 'bg-primary-600 border-primary-600'
            : 'border-surface-100'"
        >
          <svg v-if="selected.includes(option.value)" xmlns="http://www.w3.org/2000/svg" class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
          </svg>
        </span>
        {{ option.label }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T extends string">
import type { FilterOption } from '~/utils/serviceFilters'

// One multi-select filter dropdown. Status and Type differ only in their option
// table and labels, so they share this rather than two parallel template blocks.
// The parent owns both the selection and which menu is open, so only one menu
// can be open at a time.
const props = defineProps<{
  label: string
  // Prefix for this dropdown's data-testid attributes ("status" / "type").
  testid: string
  options: readonly FilterOption<T>[]
  selected: readonly T[]
  open: boolean
}>()

defineEmits<{
  toggleMenu: []
  select: [value: T]
}>()

// Reads labels from the same option table the menu renders, so the trigger text
// and the Unloaded → stopped mapping cannot drift apart.
const summary = computed(() => {
  if (props.selected.length === 0) return 'All'
  if (props.selected.length === 1) {
    const value = props.selected[0]!
    return props.options.find(o => o.value === value)?.label ?? value
  }
  return `${props.selected.length} selected`
})
</script>
