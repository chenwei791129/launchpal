import { afterEach, describe, it, expect } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import ServiceFilterBar from '../ServiceFilterBar.vue'
import type { StatusFilterValue, TypeFilterValue } from '~/utils/serviceFilters'

// Host wires v-model the way the list pages do, so we can read back what the
// bar committed after interacting with the dropdowns.
function mountBar(initial?: {
  statusFilter?: StatusFilterValue[]
  typeFilter?: TypeFilterValue[]
}) {
  const statusFilter = ref<StatusFilterValue[]>(initial?.statusFilter ?? [])
  const typeFilter = ref<TypeFilterValue[]>(initial?.typeFilter ?? [])

  const Host = defineComponent({
    setup() {
      return () => h(ServiceFilterBar, {
        'statusFilter': statusFilter.value,
        'onUpdate:statusFilter': (v: StatusFilterValue[]) => { statusFilter.value = v },
        'typeFilter': typeFilter.value,
        'onUpdate:typeFilter': (v: TypeFilterValue[]) => { typeFilter.value = v },
      })
    },
  })

  const wrapper = mount(Host, { attachTo: document.body })
  mounted.push(wrapper)
  return { wrapper, statusFilter, typeFilter }
}

// ServiceFilterBar registers document-level click/keydown listeners in
// onMounted and these tests attach to document.body, so every mount must be
// torn down — otherwise listeners and DOM nodes accumulate across the file and
// later .trigger('click') calls re-invoke every leaked handler.
const mounted: VueWrapper[] = []

afterEach(() => {
  while (mounted.length) mounted.pop()!.unmount()
})

describe('ServiceFilterBar — dropdown options', () => {
  it('renders the four Status options', async () => {
    const { wrapper } = mountBar()
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    for (const v of ['running', 'loaded', 'stopped', 'unknown']) {
      expect(wrapper.find(`[data-testid="status-option-${v}"]`).exists()).toBe(true)
    }
    const menu = wrapper.find('[data-testid="status-filter-menu"]')
    expect(menu.text()).toContain('Running')
    expect(menu.text()).toContain('Loaded')
    expect(menu.text()).toContain('Unloaded')
    expect(menu.text()).toContain('Unknown')
  })

  it('renders the four Type options, including KeepAlive', async () => {
    const { wrapper } = mountBar()
    await wrapper.find('[data-testid="type-filter-trigger"]').trigger('click')
    for (const v of ['scheduled', 'keepAlive', 'runAtLoad', 'none']) {
      expect(wrapper.find(`[data-testid="type-option-${v}"]`).exists()).toBe(true)
    }
    expect(wrapper.find('[data-testid="type-filter-menu"]').text()).toContain('KeepAlive')
  })

  it('keeps both menus closed until their trigger is clicked', () => {
    const { wrapper } = mountBar()
    expect(wrapper.find('[data-testid="status-filter-menu"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="type-filter-menu"]').exists()).toBe(false)
  })

  it('opening one menu closes the other', async () => {
    const { wrapper } = mountBar()
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    expect(wrapper.find('[data-testid="status-filter-menu"]').exists()).toBe(true)
    await wrapper.find('[data-testid="type-filter-trigger"]').trigger('click')
    expect(wrapper.find('[data-testid="status-filter-menu"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="type-filter-menu"]').exists()).toBe(true)
  })
})

describe('ServiceFilterBar — selection', () => {
  it('emits the selected status and keeps the menu open for multi-select', async () => {
    const { wrapper, statusFilter } = mountBar()
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    await wrapper.find('[data-testid="status-option-running"]').trigger('click')
    expect(statusFilter.value).toEqual(['running'])
    expect(wrapper.find('[data-testid="status-filter-menu"]').exists()).toBe(true)

    await wrapper.find('[data-testid="status-option-loaded"]').trigger('click')
    expect(statusFilter.value).toEqual(['running', 'loaded'])
  })

  it('deselects an already selected option', async () => {
    const { wrapper, statusFilter } = mountBar({ statusFilter: ['running', 'loaded'] })
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    await wrapper.find('[data-testid="status-option-running"]').trigger('click')
    expect(statusFilter.value).toEqual(['loaded'])
  })

  it('emits the selected type', async () => {
    const { wrapper, typeFilter } = mountBar()
    await wrapper.find('[data-testid="type-filter-trigger"]').trigger('click')
    await wrapper.find('[data-testid="type-option-keepAlive"]').trigger('click')
    expect(typeFilter.value).toEqual(['keepAlive'])
  })

  it('does not mutate the incoming prop array', async () => {
    const incoming: StatusFilterValue[] = []
    const { wrapper } = mountBar({ statusFilter: incoming })
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    await wrapper.find('[data-testid="status-option-running"]').trigger('click')
    expect(incoming).toEqual([])
  })
})

describe('ServiceFilterBar — trigger label', () => {
  it('shows All when nothing is selected', () => {
    const { wrapper } = mountBar()
    expect(wrapper.find('[data-testid="status-filter-trigger"]').text()).toContain('All')
    expect(wrapper.find('[data-testid="type-filter-trigger"]').text()).toContain('All')
  })

  it('shows the option label when exactly one is selected', () => {
    const { wrapper } = mountBar({ statusFilter: ['stopped'], typeFilter: ['keepAlive'] })
    expect(wrapper.find('[data-testid="status-filter-trigger"]').text()).toContain('Unloaded')
    expect(wrapper.find('[data-testid="type-filter-trigger"]').text()).toContain('KeepAlive')
  })

  it('shows a count when more than one is selected', () => {
    const { wrapper } = mountBar({ statusFilter: ['running', 'loaded'] })
    expect(wrapper.find('[data-testid="status-filter-trigger"]').text()).toContain('2 selected')
  })
})

describe('ServiceFilterBar — clearing', () => {
  it('hides the Clear all button while no filter is active', () => {
    const { wrapper } = mountBar()
    expect(wrapper.find('[data-testid="filter-clear-all"]').exists()).toBe(false)
  })

  it('clears both dropdowns', async () => {
    const { wrapper, statusFilter, typeFilter } = mountBar({
      statusFilter: ['running'], typeFilter: ['scheduled'],
    })
    await wrapper.find('[data-testid="filter-clear-all"]').trigger('click')
    expect(statusFilter.value).toEqual([])
    expect(typeFilter.value).toEqual([])
  })
})

describe('ServiceFilterBar — accessibility', () => {
  it('announces the trigger as a popup control and tracks its open state', async () => {
    const { wrapper } = mountBar()
    const trigger = wrapper.find('[data-testid="status-filter-trigger"]')
    expect(trigger.attributes('aria-haspopup')).toBe('listbox')
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(trigger.attributes('aria-controls')).toBe('status-filter-menu')

    await trigger.trigger('click')
    expect(wrapper.find('[data-testid="status-filter-trigger"]').attributes('aria-expanded'))
      .toBe('true')
  })

  it('exposes the menu as a multi-selectable listbox', async () => {
    const { wrapper } = mountBar()
    await wrapper.find('[data-testid="type-filter-trigger"]').trigger('click')
    const menu = wrapper.find('[data-testid="type-filter-menu"]')
    expect(menu.attributes('role')).toBe('listbox')
    expect(menu.attributes('aria-multiselectable')).toBe('true')
    expect(menu.attributes('id')).toBe('type-filter-menu')
  })

  it('reflects selection state on each option', async () => {
    const { wrapper } = mountBar({ statusFilter: ['running'] })
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    expect(wrapper.find('[data-testid="status-option-running"]').attributes('aria-selected'))
      .toBe('true')
    expect(wrapper.find('[data-testid="status-option-loaded"]').attributes('aria-selected'))
      .toBe('false')
  })

  // Inside a <form> an untyped <button> defaults to type="submit".
  it('marks every control as a non-submitting button', async () => {
    const { wrapper } = mountBar({ statusFilter: ['running'] })
    await wrapper.find('[data-testid="status-filter-trigger"]').trigger('click')
    for (const button of wrapper.findAll('button')) {
      expect(button.attributes('type')).toBe('button')
    }
  })
})
