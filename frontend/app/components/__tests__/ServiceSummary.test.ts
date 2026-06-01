import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref, computed } from 'vue'
import ServiceSummary from '../ServiceSummary.vue'
import type { Service } from '~/types/wails'

// Mock Nuxt auto-imports (ref, computed are used in the component without explicit import)
vi.mock('#imports', () => ({
  ref,
  computed,
}))

// Mock external dependencies
vi.mock('~/utils/shell-args', () => ({
  serializeShellArgs: (args: string[]) => args.join(' '),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  RevealInFinder: vi.fn(),
}))

vi.mock('~/composables/useNextOccurrences', () => ({
  getNextOccurrences: vi.fn(() => []),
  formatDateTime: vi.fn((d: unknown) => String(d)),
  WEEKDAY_NAMES: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
}))

// Minimal service fixture with environment variables
const makeService = (environment?: Record<string, string>, overrides: Partial<Service> = {}): Service => ({
  name: 'com.example.test',
  label: 'com.example.test',
  status: 'stopped',
  statusConfidence: 'verified',
  path: '/Library/LaunchAgents/com.example.test.plist',
  runAtLoad: false,
  keepAlive: { enabled: false, mode: '' },
  wakeSystem: false,
  type: 'user',
  readOnly: false,
  plistFormat: 'xml',
  environment,
  ...overrides,
})

describe('ServiceSummary – environment variable masking', () => {
  let wrapper: ReturnType<typeof mount>

  beforeEach(() => {
    wrapper = mount(ServiceSummary, {
      props: {
        service: makeService({ SECRET_KEY: 'super-secret', API_URL: 'https://example.com' }),
      },
    })
  })

  it('masks env var values with bullet characters by default', () => {
    const text = wrapper.text()
    expect(text).not.toContain('super-secret')
    expect(text).not.toContain('https://example.com')
    expect(text).toContain('••••••••')
  })

  it('shows the key name but not the value by default', () => {
    const text = wrapper.text()
    expect(text).toContain('SECRET_KEY')
    expect(text).toContain('API_URL')
    expect(text).not.toContain('super-secret')
  })

  it('reveals value when eye button is clicked', async () => {
    // Find the toggle button for SECRET_KEY (first env var row's button)
    const rows = wrapper.findAll('[data-testid="env-var-row"]')
    expect(rows.length).toBe(2)

    const secretKeyRow = rows[0]!
    const toggleBtn = secretKeyRow.find('[data-testid="env-var-toggle"]')
    expect(toggleBtn.exists()).toBe(true)

    await toggleBtn.trigger('click')

    expect(wrapper.text()).toContain('super-secret')
  })

  it('hides value again when eye button is clicked a second time', async () => {
    const rows = wrapper.findAll('[data-testid="env-var-row"]')
    const toggleBtn = rows[0]!.find('[data-testid="env-var-toggle"]')

    // Reveal
    await toggleBtn.trigger('click')
    expect(wrapper.text()).toContain('super-secret')

    // Hide again
    await toggleBtn.trigger('click')
    expect(wrapper.text()).not.toContain('super-secret')
    expect(wrapper.text()).toContain('••••••••')
  })

  it('toggles each variable independently', async () => {
    const rows = wrapper.findAll('[data-testid="env-var-row"]')
    const firstBtn = rows[0]!.find('[data-testid="env-var-toggle"]')
    const secondBtn = rows[1]!.find('[data-testid="env-var-toggle"]')

    // Reveal only first variable
    await firstBtn.trigger('click')

    expect(wrapper.text()).toContain('super-secret')
    // Second variable should still be masked
    expect(wrapper.text()).not.toContain('https://example.com')

    // Reveal second variable too
    await secondBtn.trigger('click')
    expect(wrapper.text()).toContain('https://example.com')
  })
})

describe('ServiceSummary – KeepAlive and ThrottleInterval rendering', () => {
  it('renders "No" for a disabled KeepAlive', () => {
    const wrapper = mount(ServiceSummary, { props: { service: makeService() } })
    expect(wrapper.find('[data-testid="keepalive-summary"]').text()).toBe('No')
    expect(wrapper.find('[data-testid="keepalive-conditions"]').exists()).toBe(false)
  })

  it('renders "Yes" for a boolean KeepAlive', () => {
    const wrapper = mount(ServiceSummary, {
      props: { service: makeService(undefined, { keepAlive: { enabled: true, mode: 'boolean' } }) },
    })
    expect(wrapper.find('[data-testid="keepalive-summary"]').text()).toBe('Yes')
  })

  it('renders dictionary-mode conditions including non-editable sub-keys', () => {
    const wrapper = mount(ServiceSummary, {
      props: {
        service: makeService(undefined, {
          keepAlive: {
            enabled: true,
            mode: 'dictionary',
            successfulExit: false,
            pathState: { '/tmp/flag': true },
          },
          throttleInterval: 30,
        }),
      },
    })
    expect(wrapper.find('[data-testid="keepalive-summary"]').text()).toBe('On conditions')
    const conditions = wrapper.find('[data-testid="keepalive-conditions"]').text()
    expect(conditions).toContain('SuccessfulExit: false')
    expect(conditions).toContain('PathState[/tmp/flag]: true')
    expect(wrapper.find('[data-testid="throttle-interval"]').text()).toBe('30s')
  })

  it('omits the ThrottleInterval row when unset', () => {
    const wrapper = mount(ServiceSummary, { props: { service: makeService() } })
    expect(wrapper.find('[data-testid="throttle-interval"]').exists()).toBe(false)
  })
})
