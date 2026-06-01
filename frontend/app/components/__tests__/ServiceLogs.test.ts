import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import ServiceLogs from '../ServiceLogs.vue'

// Mock Nuxt auto-imports.
vi.mock('#imports', () => ({
  ref,
  computed,
  nextTick,
  onMounted,
  onBeforeUnmount,
  watch,
}))

interface MockedApp {
  GetLogs: ReturnType<typeof vi.fn>
  GetSystemLogs: ReturnType<typeof vi.fn>
  GetLogClearStatus: ReturnType<typeof vi.fn>
  ClearLogs: ReturnType<typeof vi.fn>
  ClearSystemLogs: ReturnType<typeof vi.fn>
}

let mockedApp: MockedApp

// Provide a fresh App mock for each test so call counts and resolved values
// don't bleed across cases.
function installAppMock(overrides: Partial<MockedApp> = {}) {
  mockedApp = {
    GetLogs: vi.fn().mockResolvedValue('user log content'),
    GetSystemLogs: vi.fn().mockResolvedValue('system log content'),
    GetLogClearStatus: vi.fn().mockResolvedValue({
      logPath: '/tmp/log.log',
      exists: true,
      userWritable: true,
    }),
    ClearLogs: vi.fn().mockResolvedValue(undefined),
    ClearSystemLogs: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
  ;(globalThis as unknown as { window: { go: { main: { App: MockedApp } } } }).window = {
    go: { main: { App: mockedApp } },
  }
}

beforeEach(() => {
  installAppMock()
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ServiceLogs – Clear Logs control visibility', () => {
  it('hides the button entirely for apple-system services', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.apple.x', serviceType: 'apple-system' },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="clear-logs-button"]').exists()).toBe(false)
    // Status query must not run for apple-system: the spec says it should
    // be cheap and skip the round trip entirely.
    expect(mockedApp.GetLogClearStatus).not.toHaveBeenCalled()
  })

  it('shows the button for user services and queries status on mount', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="clear-logs-button"]').exists()).toBe(true)
    expect(mockedApp.GetLogClearStatus).toHaveBeenCalledWith('com.user.x', 'user', 'stdout')
  })

  it('re-queries GetLogClearStatus when logType switches', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await flushPromises()
    expect(mockedApp.GetLogClearStatus).toHaveBeenCalledWith('com.user.x', 'user', 'stdout')
    const stderrToggle = wrapper.findAll('button').find(b => b.text() === 'stderr')
    expect(stderrToggle).toBeDefined()
    ;(stderrToggle!.element as HTMLButtonElement).click()
    await flushPromises()
    expect(mockedApp.GetLogClearStatus).toHaveBeenCalledWith('com.user.x', 'user', 'stderr')
    wrapper.unmount()
  })
})

describe('ServiceLogs – disabled tooltips', () => {
  it('disables with "No log path configured" when logPath is empty', async () => {
    installAppMock({
      GetLogClearStatus: vi.fn().mockResolvedValue({
        logPath: '',
        exists: false,
        userWritable: false,
      }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.nopath', serviceType: 'user' },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]')
    expect(btn.attributes('disabled')).toBeDefined()
    expect(btn.attributes('title')).toBe('No log path configured')
  })

  it('disables with "Log file does not exist" when path exists but file is missing', async () => {
    installAppMock({
      GetLogClearStatus: vi.fn().mockResolvedValue({
        logPath: '/tmp/missing.log',
        exists: false,
        userWritable: false,
      }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.missing', serviceType: 'user' },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]')
    expect(btn.attributes('disabled')).toBeDefined()
    expect(btn.attributes('title')).toBe('Log file does not exist')
  })

  it('disables with "Enable Admin Mode to clear" for system service when not user-writable and admin off', async () => {
    installAppMock({
      GetLogClearStatus: vi.fn().mockResolvedValue({
        logPath: '/var/log/x.log',
        exists: true,
        userWritable: false,
      }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.x', serviceType: 'system', adminEnabled: false },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]')
    expect(btn.attributes('disabled')).toBeDefined()
    expect(btn.attributes('title')).toBe('Enable Admin Mode to clear')
  })

  it('enables system-service button when not user-writable but admin mode on', async () => {
    installAppMock({
      GetLogClearStatus: vi.fn().mockResolvedValue({
        logPath: '/var/log/x.log',
        exists: true,
        userWritable: false,
      }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.x', serviceType: 'system', adminEnabled: true },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]')
    expect(btn.attributes('disabled')).toBeUndefined()
  })
})

describe('ServiceLogs – ANSI color rendering', () => {
  it('renders ANSI colors as spans and drops the literal escape codes', async () => {
    installAppMock({
      GetLogs: vi.fn().mockResolvedValue('\x1b[32mOK\x1b[0m booted'),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await flushPromises()
    const pre = wrapper.find('pre')
    expect(pre.exists()).toBe(true)
    expect(pre.html()).toContain('<span style="color:#98c379">OK</span>')
    expect(pre.text()).toContain('booted')
    expect(pre.html()).not.toContain('[32m')
    expect(pre.html()).not.toContain('[0m')
    wrapper.unmount()
  })

  it('preserves the placeholder branch for an empty log', async () => {
    installAppMock({
      GetLogs: vi.fn().mockResolvedValue(''),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.empty', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.text()).toContain('No logs available for stdout')
    wrapper.unmount()
  })

  it('keeps the loading branch while GetLogs is pending', async () => {
    let resolveLogs!: (value: string) => void
    installAppMock({
      GetLogs: vi.fn().mockReturnValue(new Promise<string>((resolve) => {
        resolveLogs = resolve
      })),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.loading', serviceType: 'user' },
    })
    await nextTick()
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.text()).toContain('Loading logs...')
    resolveLogs('done')
    await flushPromises()
    wrapper.unmount()
  })

  it('keeps the error branch when GetLogs rejects', async () => {
    installAppMock({
      GetLogs: vi.fn().mockRejectedValue(new Error('boom')),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.err', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.find('.text-red-400').text()).toBe('boom')
    wrapper.unmount()
  })
})

describe('ServiceLogs – confirm dialog', () => {
  it('opens dialog on click and cancels without invoking ClearLogs', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]').element as HTMLButtonElement
    btn.click()
    await flushPromises()
    expect(document.body.querySelector('[data-testid="clear-logs-dialog"]')).not.toBeNull()
    const cancelBtn = document.body.querySelector('[data-testid="clear-logs-cancel"]') as HTMLButtonElement
    cancelBtn.click()
    await flushPromises()
    expect(document.body.querySelector('[data-testid="clear-logs-dialog"]')).toBeNull()
    expect(mockedApp.ClearLogs).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('confirms and routes user services to ClearLogs binding', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]').element as HTMLButtonElement
    btn.click()
    await flushPromises()
    const confirmBtn = document.body.querySelector('[data-testid="clear-logs-confirm"]') as HTMLButtonElement
    expect(confirmBtn).not.toBeNull()
    confirmBtn.click()
    await flushPromises()
    expect(mockedApp.ClearLogs).toHaveBeenCalledWith('com.user.x', 'stdout')
    expect(mockedApp.ClearSystemLogs).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('confirms and routes system services to ClearSystemLogs binding', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.sys.x', serviceType: 'system', adminEnabled: true },
    })
    await flushPromises()
    const btn = wrapper.find('[data-testid="clear-logs-button"]').element as HTMLButtonElement
    btn.click()
    await flushPromises()
    const confirmBtn = document.body.querySelector('[data-testid="clear-logs-confirm"]') as HTMLButtonElement
    expect(confirmBtn).not.toBeNull()
    confirmBtn.click()
    await flushPromises()
    expect(mockedApp.ClearSystemLogs).toHaveBeenCalledWith('com.sys.x', 'system', 'stdout')
    expect(mockedApp.ClearLogs).not.toHaveBeenCalled()
    wrapper.unmount()
  })
})
