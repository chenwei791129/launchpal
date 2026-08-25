import type { Service } from '~/types/wails'

// StatusFilterValue is the raw `Service.status` string a filter option selects.
// The user-facing "Unloaded" label maps onto 'stopped'; keeping the raw status
// as the option value makes that mapping a single table entry rather than a
// branch in the predicate.
export type StatusFilterValue = Service['status']

// TypeFilterValue names a launch-policy bucket. 'none' is the negation of the
// other three rather than a field of its own.
export type TypeFilterValue = 'scheduled' | 'keepAlive' | 'runAtLoad' | 'none'

export interface FilterOption<T extends string> {
  value: T
  label: string
}

export const STATUS_FILTER_OPTIONS: readonly FilterOption<StatusFilterValue>[] = [
  { value: 'running', label: 'Running' },
  { value: 'loaded', label: 'Loaded' },
  { value: 'stopped', label: 'Unloaded' },
  { value: 'unknown', label: 'Unknown' },
]

// The first three options mirror the launch-policy badges ServiceRow renders in
// the Type column, so a filter selection never contradicts the badge the user
// sees. The badge picks one label by precedence (schedule > keepAlive >
// runAtLoad); the filter instead matches every option that applies, so a
// scheduled + runAtLoad service is found under either.
export const TYPE_FILTER_OPTIONS: readonly FilterOption<TypeFilterValue>[] = [
  { value: 'scheduled', label: 'Scheduled' },
  { value: 'keepAlive', label: 'KeepAlive' },
  { value: 'runAtLoad', label: 'RunAtLoad' },
  { value: 'none', label: 'None' },
]

// hasActiveFilter is the single definition of "any dropdown is narrowing the
// list". filterServices uses it for its fast path, the list pages use it to
// distinguish "filtered to nothing" from "no services at all", and the filter
// bar uses it to show Clear all — a third filter dimension only has to be
// added here.
export function hasActiveFilter(
  statusFilter: readonly StatusFilterValue[],
  typeFilter: readonly TypeFilterValue[],
): boolean {
  return statusFilter.length > 0 || typeFilter.length > 0
}

// An empty selection means "All" — the filter is inactive rather than matching
// nothing, which is what makes the cross-filter AND below degrade cleanly.
export function matchesStatusFilter(
  service: Service,
  statusFilter: readonly StatusFilterValue[],
): boolean {
  if (statusFilter.length === 0) return true
  return statusFilter.includes(service.status)
}

export function matchesTypeFilter(
  service: Service,
  typeFilter: readonly TypeFilterValue[],
): boolean {
  if (typeFilter.length === 0) return true

  const scheduled = service.schedule !== undefined && service.schedule !== null
  // A Keep Alive service carries runAtLoad === false because launchd implies
  // RunAtLoad from KeepAlive — without this bucket it would fall into None.
  const keepAlive = service.keepAlive?.enabled === true
  const runAtLoad = service.runAtLoad === true

  return typeFilter.some((option) => {
    switch (option) {
      case 'scheduled': return scheduled
      case 'keepAlive': return keepAlive
      case 'runAtLoad': return runAtLoad
      case 'none': return !scheduled && !keepAlive && !runAtLoad
    }
  })
}

export interface ServiceFilters {
  searchQuery: string
  statusFilter: readonly StatusFilterValue[]
  typeFilter: readonly TypeFilterValue[]
}

// filterServices is the single implementation shared by the User services page
// and ReadOnlyServiceList, so the three list pages cannot drift apart. Status,
// Type and the text search combine with AND; each is skipped while inactive.
export function filterServices(services: Service[], filters: ServiceFilters): Service[] {
  const query = filters.searchQuery.trim().toLowerCase()
  if (!query && !hasActiveFilter(filters.statusFilter, filters.typeFilter)) {
    return services
  }

  return services.filter(service =>
    (!query || service.label.toLowerCase().includes(query))
    && matchesStatusFilter(service, filters.statusFilter)
    && matchesTypeFilter(service, filters.typeFilter),
  )
}
