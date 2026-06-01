import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ServiceRow from '../ServiceRow.vue'
import type { Service } from '~/types/wails'

// ServiceRow reads Admin Mode state and uses Nuxt's navigateTo global.
vi.mock('~/composables/useAdminMode', () => ({
  useAdminMode: () => ({ isEnabled: { value: false } }),
}))

beforeEach(() => {
  vi.stubGlobal('navigateTo', vi.fn())
})

function makeService(overrides: Partial<Service> = {}): Service {
  return {
    name: 'com.example.test',
    label: 'com.example.test',
    status: 'stopped',
    statusConfidence: 'verified',
    path: '/Users/dev/Library/LaunchAgents/com.example.test.plist',
    runAtLoad: false,
    keepAlive: { enabled: false, mode: '' },
    wakeSystem: false,
    type: 'user',
    readOnly: false,
    plistFormat: 'xml',
    ...overrides,
  }
}

function mountRow(service: Service) {
  return mount(ServiceRow, {
    props: { service },
    global: { stubs: { StatusConfidenceIcon: true } },
  })
}

describe('ServiceRow — launch-policy badge', () => {
  it('shows a KeepAlive badge for a keepAlive-enabled service', () => {
    const wrapper = mountRow(makeService({ keepAlive: { enabled: true, mode: 'dictionary', successfulExit: false } }))
    const badge = wrapper.find('[data-testid="launch-policy-badge"]')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toBe('KeepAlive')
  })

  it('shows a RunAtLoad badge for a runAtLoad-only service', () => {
    const wrapper = mountRow(makeService({ runAtLoad: true }))
    const badge = wrapper.find('[data-testid="launch-policy-badge"]')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toBe('RunAtLoad')
  })

  it('shows no launch-policy badge for an on-demand service', () => {
    const wrapper = mountRow(makeService())
    expect(wrapper.find('[data-testid="launch-policy-badge"]').exists()).toBe(false)
  })

  it('prefers the Scheduled badge over the launch-policy badge', () => {
    const wrapper = mountRow(makeService({ runAtLoad: true, schedule: { interval: 60 } }))
    expect(wrapper.text()).toContain('Scheduled')
    expect(wrapper.find('[data-testid="launch-policy-badge"]').exists()).toBe(false)
  })
})
