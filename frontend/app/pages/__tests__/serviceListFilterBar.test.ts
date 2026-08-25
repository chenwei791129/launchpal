import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

// The spec requires the filter bar on all three service list pages. The list
// surfaces are not a single component: `system.vue` was split back out of
// ReadOnlyServiceList when Admin Mode landed, so it carries its own header,
// filter state and empty state. A page dropping out of this set is invisible
// in a per-component test — this pins the whole set instead.
const LIST_SURFACES = [
  { name: 'User services page', file: 'app/pages/index.vue' },
  { name: 'System services page', file: 'app/pages/system.vue' },
  { name: 'Apple System services (via ReadOnlyServiceList)', file: 'app/components/ReadOnlyServiceList.vue' },
]

function read(file: string): string {
  return readFileSync(resolve(__dirname, '../../..', file), 'utf8')
}

describe.each(LIST_SURFACES)('$name', ({ file }) => {
  const source = read(file)

  it('renders the shared ServiceFilterBar', () => {
    expect(source).toContain('<ServiceFilterBar')
    expect(source).toContain('v-model:status-filter="statusFilter"')
    expect(source).toContain('v-model:type-filter="typeFilter"')
  })

  // Anchored on real markup (the two data-testids), not on a source comment —
  // renaming a comment must not fail this, but moving the bar out of position
  // must.
  it('places the filter bar between the header and the table header', () => {
    const barAt = source.indexOf('<ServiceFilterBar')
    const headerEndAt = source.indexOf('</header>')
    const tableHeaderAt = source.indexOf('data-testid="service-list-table-header"')
    expect(headerEndAt).toBeGreaterThan(-1)
    expect(tableHeaderAt).toBeGreaterThan(-1)
    expect(barAt).toBeGreaterThan(headerEndAt)
    expect(barAt).toBeLessThan(tableHeaderAt)
  })

  it('filters through the shared useServiceListFilters wiring rather than its own state', () => {
    expect(source).toContain('useServiceListFilters(services, searchQuery)')
    // The page must not re-declare the filter state the composable owns.
    expect(source).not.toContain('const statusFilter = ref')
    expect(source).not.toContain('const typeFilter = ref')
  })

  it('reports the filtered-to-empty case distinctly from having no services', () => {
    expect(source).toContain('hasActiveFilter')
    expect(source).toContain('No services match the selected filters')
  })
})

// Apple System reaches the list through ReadOnlyServiceList; if that ever stops
// being true the assertions above would silently stop covering it.
describe('apple-system page', () => {
  it('still delegates to ReadOnlyServiceList', () => {
    expect(read('app/pages/apple-system.vue')).toContain('<ReadOnlyServiceList')
  })
})
