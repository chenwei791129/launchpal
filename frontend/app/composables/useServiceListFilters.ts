import type { Ref } from 'vue'
import type { Service } from '~/types/wails'
import {
  filterServices,
  hasActiveFilter,
  type StatusFilterValue,
  type TypeFilterValue,
} from '~/utils/serviceFilters'

/**
 * Reactive wiring shared by the three service list surfaces (`pages/index.vue`,
 * `pages/system.vue`, and `components/ReadOnlyServiceList.vue`). Those are
 * separate files with very different shells — Admin Mode banner, New Service
 * button, permission warning — but the filter state around the list is
 * identical, so it lives here rather than being copy-pasted three times.
 *
 * Unlike `useAdminMode` / `useSettings`, this is a **factory**: each call
 * returns fresh refs. Filter selection is a per-page choice and must not be
 * shared between pages the way session-global state is.
 *
 * `hasActiveFilter` distinguishes "the filters excluded everything" from "there
 * are no services", which the pages use to pick the right empty state.
 */
export function useServiceListFilters(services: Ref<Service[]>, searchQuery: Ref<string>) {
  const statusFilter = ref<StatusFilterValue[]>([])
  const typeFilter = ref<TypeFilterValue[]>([])

  return {
    statusFilter,
    typeFilter,
    hasActiveFilter: computed(() => hasActiveFilter(statusFilter.value, typeFilter.value)),
    filteredServices: computed(() => filterServices(services.value, {
      searchQuery: searchQuery.value,
      statusFilter: statusFilter.value,
      typeFilter: typeFilter.value,
    })),
  }
}
