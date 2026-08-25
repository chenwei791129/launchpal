import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { useServiceListFilters } from '~/composables/useServiceListFilters'
import type { Service } from '~/types/wails'

function svc(overrides: Partial<Service> = {}): Service {
  return {
    name: 'com.example.job',
    label: 'com.example.job',
    status: 'stopped',
    path: '/Users/x/Library/LaunchAgents/com.example.job.plist',
    runAtLoad: false,
    keepAlive: { enabled: false, mode: '' },
    wakeSystem: false,
    type: 'user',
    readOnly: false,
    plistFormat: 'xml',
    statusConfidence: 'verified',
    ...overrides,
  }
}

describe('useServiceListFilters', () => {
  it('starts with both dropdowns empty and no active filter', () => {
    const { statusFilter, typeFilter, hasActiveFilter } = useServiceListFilters(ref([]), ref(''))
    expect(statusFilter.value).toEqual([])
    expect(typeFilter.value).toEqual([])
    expect(hasActiveFilter.value).toBe(false)
  })

  it('recomputes filteredServices when a filter changes', () => {
    const services = ref([svc({ name: 'a', status: 'running' }), svc({ name: 'b' })])
    const { statusFilter, filteredServices, hasActiveFilter } = useServiceListFilters(services, ref(''))
    expect(filteredServices.value).toHaveLength(2)

    statusFilter.value = ['running']
    expect(filteredServices.value.map(s => s.name)).toEqual(['a'])
    expect(hasActiveFilter.value).toBe(true)
  })

  it('recomputes when the services list changes', () => {
    const services = ref<Service[]>([])
    const { filteredServices } = useServiceListFilters(services, ref(''))
    expect(filteredServices.value).toHaveLength(0)

    services.value = [svc()]
    expect(filteredServices.value).toHaveLength(1)
  })

  it('ANDs the text search with the dropdowns', () => {
    const services = ref([
      svc({ name: 'a', label: 'com.apple.one', status: 'running' }),
      svc({ name: 'b', label: 'com.other.two', status: 'running' }),
    ])
    const searchQuery = ref('')
    const { statusFilter, filteredServices } = useServiceListFilters(services, searchQuery)

    statusFilter.value = ['running']
    searchQuery.value = 'com.apple'
    expect(filteredServices.value.map(s => s.name)).toEqual(['a'])
  })

  // This is a factory, not a module-level singleton like useAdminMode /
  // useSettings: filter selection is a per-page choice and must never leak
  // between two mounted list surfaces.
  it('returns independent state per call', () => {
    const a = useServiceListFilters(ref([]), ref(''))
    const b = useServiceListFilters(ref([]), ref(''))

    a.statusFilter.value = ['running']
    expect(b.statusFilter.value).toEqual([])
    expect(a.hasActiveFilter.value).toBe(true)
    expect(b.hasActiveFilter.value).toBe(false)
  })
})
