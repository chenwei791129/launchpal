import { describe, it, expect } from 'vitest'
import type { StatusFilterValue } from '~/utils/serviceFilters'
import {
  STATUS_FILTER_OPTIONS,
  TYPE_FILTER_OPTIONS,
  filterServices,
  hasActiveFilter,
  matchesStatusFilter,
  matchesTypeFilter,
} from '~/utils/serviceFilters'
import type { Service } from '~/types/wails'

// Minimal Service factory — only the fields the filters read carry meaning,
// the rest satisfy the type.
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

describe('filter option tables', () => {
  it('exposes the four status options with the Unloaded → stopped mapping', () => {
    expect(STATUS_FILTER_OPTIONS.map(o => o.label)).toEqual([
      'Running', 'Loaded', 'Unloaded', 'Unknown',
    ])
    expect(STATUS_FILTER_OPTIONS.map(o => o.value)).toEqual([
      'running', 'loaded', 'stopped', 'unknown',
    ])
  })

  it('exposes the four type options, including KeepAlive', () => {
    expect(TYPE_FILTER_OPTIONS.map(o => o.label)).toEqual([
      'Scheduled', 'KeepAlive', 'RunAtLoad', 'None',
    ])
    expect(TYPE_FILTER_OPTIONS.map(o => o.value)).toEqual([
      'scheduled', 'keepAlive', 'runAtLoad', 'none',
    ])
  })
})

describe('matchesStatusFilter', () => {
  it('displays all services when no status option is selected', () => {
    expect(matchesStatusFilter(svc({ status: 'running' }), [])).toBe(true)
    expect(matchesStatusFilter(svc({ status: 'unknown' }), [])).toBe(true)
  })

  it('displays only running services when Running is selected', () => {
    expect(matchesStatusFilter(svc({ status: 'running' }), ['running'])).toBe(true)
    expect(matchesStatusFilter(svc({ status: 'loaded' }), ['running'])).toBe(false)
  })

  it('ORs multiple selected statuses', () => {
    const selected: StatusFilterValue[] = ['running', 'loaded']
    expect(matchesStatusFilter(svc({ status: 'running' }), selected)).toBe(true)
    expect(matchesStatusFilter(svc({ status: 'loaded' }), selected)).toBe(true)
    expect(matchesStatusFilter(svc({ status: 'stopped' }), selected)).toBe(false)
  })

  it('maps Unloaded onto status stopped', () => {
    expect(matchesStatusFilter(svc({ status: 'stopped' }), ['stopped'])).toBe(true)
    expect(matchesStatusFilter(svc({ status: 'running' }), ['stopped'])).toBe(false)
  })
})

describe('matchesTypeFilter', () => {
  const scheduled = svc({ schedule: { interval: 3600 } })
  const keepAlive = svc({ keepAlive: { enabled: true, mode: 'boolean' } })
  const runAtLoad = svc({ runAtLoad: true })
  const none = svc()

  it('displays all services when no type option is selected', () => {
    for (const s of [scheduled, keepAlive, runAtLoad, none]) {
      expect(matchesTypeFilter(s, [])).toBe(true)
    }
  })

  it('matches Scheduled on a defined schedule', () => {
    expect(matchesTypeFilter(scheduled, ['scheduled'])).toBe(true)
    expect(matchesTypeFilter(none, ['scheduled'])).toBe(false)
  })

  it('matches KeepAlive on keepAlive.enabled', () => {
    expect(matchesTypeFilter(keepAlive, ['keepAlive'])).toBe(true)
    expect(matchesTypeFilter(runAtLoad, ['keepAlive'])).toBe(false)
  })

  it('matches RunAtLoad on runAtLoad === true', () => {
    expect(matchesTypeFilter(runAtLoad, ['runAtLoad'])).toBe(true)
    expect(matchesTypeFilter(none, ['runAtLoad'])).toBe(false)
  })

  it('matches a service against every applicable option, not just the badge one', () => {
    const both = svc({ schedule: { interval: 60 }, runAtLoad: true })
    expect(matchesTypeFilter(both, ['scheduled'])).toBe(true)
    expect(matchesTypeFilter(both, ['runAtLoad'])).toBe(true)
  })

  it('excludes scheduled, keepAlive and runAtLoad services from None', () => {
    expect(matchesTypeFilter(none, ['none'])).toBe(true)
    expect(matchesTypeFilter(scheduled, ['none'])).toBe(false)
    expect(matchesTypeFilter(keepAlive, ['none'])).toBe(false)
    expect(matchesTypeFilter(runAtLoad, ['none'])).toBe(false)
  })

  // Spec example: a KeepAlive service carries runAtLoad === false because
  // launchd implies RunAtLoad from KeepAlive — it must not fall into None.
  it('does not classify a KeepAlive service as None', () => {
    const s = svc({ keepAlive: { enabled: true, mode: 'boolean' }, runAtLoad: false })
    expect(matchesTypeFilter(s, ['keepAlive'])).toBe(true)
    expect(matchesTypeFilter(s, ['none'])).toBe(false)
  })

  it('treats a missing keepAlive object as not KeepAlive', () => {
    const s = { ...svc(), keepAlive: undefined } as unknown as Service
    expect(matchesTypeFilter(s, ['keepAlive'])).toBe(false)
    expect(matchesTypeFilter(s, ['none'])).toBe(true)
  })
})

describe('filterServices — cross-filter AND logic', () => {
  const services: Service[] = [
    svc({ name: 'a', label: 'com.apple.runner', status: 'running', schedule: { interval: 60 } }),
    svc({ name: 'b', label: 'com.apple.idle', status: 'stopped', schedule: { interval: 60 } }),
    svc({ name: 'c', label: 'com.other.runner', status: 'running', schedule: { interval: 60 } }),
    svc({ name: 'd', label: 'com.apple.plain', status: 'running' }),
  ]

  it('returns everything when no filter is active', () => {
    expect(filterServices(services, { searchQuery: '', statusFilter: [], typeFilter: [] }))
      .toHaveLength(4)
  })

  it('ANDs Status and Type', () => {
    const out = filterServices(services, {
      searchQuery: '', statusFilter: ['running'], typeFilter: ['scheduled'],
    })
    expect(out.map(s => s.name)).toEqual(['a', 'c'])
  })

  it('ANDs Status, Type and the text search', () => {
    const out = filterServices(services, {
      searchQuery: 'com.apple', statusFilter: ['running'], typeFilter: ['scheduled'],
    })
    expect(out.map(s => s.name)).toEqual(['a'])
  })

  it('returns an empty list when the combination matches nothing', () => {
    const out = filterServices(services, {
      searchQuery: '', statusFilter: ['unknown'], typeFilter: ['none'],
    })
    expect(out).toEqual([])
  })

  it('matches the search query case-insensitively and trims it', () => {
    const out = filterServices(services, {
      searchQuery: '  COM.OTHER  ', statusFilter: [], typeFilter: [],
    })
    expect(out.map(s => s.name)).toEqual(['c'])
  })
})

describe('hasActiveFilter', () => {
  it('is false only when both dropdowns are empty', () => {
    expect(hasActiveFilter([], [])).toBe(false)
  })

  it('is true when either dropdown has a selection', () => {
    expect(hasActiveFilter(['running'], [])).toBe(true)
    expect(hasActiveFilter([], ['scheduled'])).toBe(true)
    expect(hasActiveFilter(['running'], ['scheduled'])).toBe(true)
  })

  // The search box is deliberately not part of this — the list pages use it to
  // tell "the dropdowns excluded everything" apart from "no services at all",
  // and the search box already has its own empty-state message.
  it('ignores the text search', () => {
    expect(hasActiveFilter([], [])).toBe(false)
  })
})
