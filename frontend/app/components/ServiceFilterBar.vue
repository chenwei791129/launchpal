<template>
  <div
    ref="root"
    data-testid="filter-bar"
    class="flex items-center gap-2 px-4 py-2 border-b border-surface-100"
  >
    <ServiceFilterDropdown
      label="Status"
      testid="status"
      :options="STATUS_FILTER_OPTIONS"
      :selected="statusFilter"
      :open="openMenu === 'status'"
      @toggle-menu="toggleMenu('status')"
      @select="toggleStatus"
    />

    <ServiceFilterDropdown
      label="Type"
      testid="type"
      :options="TYPE_FILTER_OPTIONS"
      :selected="typeFilter"
      :open="openMenu === 'type'"
      @toggle-menu="toggleMenu('type')"
      @select="toggleType"
    />

    <button
      v-if="anyActive"
      data-testid="filter-clear-all"
      type="button"
      class="px-2 py-1 text-sm text-gray-500 hover:text-white transition-colors"
      @click="clearAll"
    >
      Clear all
    </button>
  </div>
</template>

<script setup lang="ts">
// Explicit import: Nuxt auto-imports components in the app, but the component
// tests mount this file directly without that resolution step.
import ServiceFilterDropdown from './ServiceFilterDropdown.vue'
import {
  STATUS_FILTER_OPTIONS,
  TYPE_FILTER_OPTIONS,
  hasActiveFilter,
  type StatusFilterValue,
  type TypeFilterValue,
} from '~/utils/serviceFilters'

const props = defineProps<{
  statusFilter: StatusFilterValue[]
  typeFilter: TypeFilterValue[]
}>()

const emit = defineEmits<{
  'update:statusFilter': [value: StatusFilterValue[]]
  'update:typeFilter': [value: TypeFilterValue[]]
}>()

type MenuName = 'status' | 'type'

const root = ref<HTMLElement | null>(null)
// Only one menu is open at a time, so opening the second closes the first
// instead of leaving two overlapping popovers on screen.
const openMenu = ref<MenuName | null>(null)

function toggleMenu(name: MenuName) {
  openMenu.value = openMenu.value === name ? null : name
}

const anyActive = computed(() => hasActiveFilter(props.statusFilter, props.typeFilter))

// Emits a new array rather than mutating the prop, and leaves the menu open so
// several options can be picked in one visit.
function nextSelection<T extends string>(current: readonly T[], value: T): T[] {
  return current.includes(value)
    ? current.filter(v => v !== value)
    : [...current, value]
}

function toggleStatus(value: StatusFilterValue) {
  emit('update:statusFilter', nextSelection(props.statusFilter, value))
}

function toggleType(value: TypeFilterValue) {
  emit('update:typeFilter', nextSelection(props.typeFilter, value))
}

function clearAll() {
  emit('update:statusFilter', [])
  emit('update:typeFilter', [])
  openMenu.value = null
}

function handleDocumentClick(event: MouseEvent) {
  if (!openMenu.value) return
  if (root.value && !root.value.contains(event.target as Node)) {
    openMenu.value = null
  }
}

function handleEscape(event: KeyboardEvent) {
  if (event.key === 'Escape') openMenu.value = null
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  document.removeEventListener('keydown', handleEscape)
})
</script>
