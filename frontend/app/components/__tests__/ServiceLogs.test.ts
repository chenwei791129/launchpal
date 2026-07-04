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
    GetLogs: vi.fn().mockResolvedValue({ content: 'user log content', status: 'ok', path: '/tmp/user.log' }),
    GetSystemLogs: vi.fn().mockResolvedValue({ content: 'system log content', status: 'ok', path: '/var/log/sys.log' }),
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
      GetLogs: vi.fn().mockResolvedValue({ content: '\x1b[32mOK\x1b[0m booted', status: 'ok', path: '/tmp/log.log' }),
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
      GetLogs: vi.fn().mockResolvedValue({ content: '', status: 'ok', path: '/tmp/log.log' }),
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
    interface LogsResultShape { content: string, status: string, path: string }
    let resolveLogs!: (value: LogsResultShape) => void
    installAppMock({
      GetLogs: vi.fn().mockReturnValue(new Promise<LogsResultShape>((resolve) => {
        resolveLogs = resolve
      })),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.loading', serviceType: 'user' },
    })
    await nextTick()
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.text()).toContain('Loading logs...')
    resolveLogs({ content: 'done', status: 'ok', path: '/tmp/log.log' })
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

describe('ServiceLogs – log load result classification', () => {
  it('renders the no-path placeholder without the red error branch', async () => {
    installAppMock({
      GetLogs: vi.fn().mockResolvedValue({ content: '', status: 'no-path', path: '' }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.nopath', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('No stdout log path configured for this service')
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    wrapper.unmount()
  })

  it('renders the not-found placeholder with the path as secondary text', async () => {
    installAppMock({
      GetLogs: vi.fn().mockResolvedValue({ content: '', status: 'not-found', path: '/var/log/foo/out.log' }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.missing', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('Log file does not exist yet')
    expect(wrapper.text()).toContain('/var/log/foo/out.log')
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    wrapper.unmount()
  })

  it('keeps the existing placeholder for ok with empty content', async () => {
    installAppMock({
      GetLogs: vi.fn().mockResolvedValue({ content: '', status: 'ok', path: '/tmp/log.log' }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.empty', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('No logs available for stdout')
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    wrapper.unmount()
  })

  it('renders ok content through the ANSI pipeline in the pre element', async () => {
    installAppMock({
      GetLogs: vi.fn().mockResolvedValue({ content: 'hello\n', status: 'ok', path: '/tmp/log.log' }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.hello', serviceType: 'user' },
    })
    await flushPromises()
    const pre = wrapper.find('pre')
    expect(pre.exists()).toBe(true)
    expect(pre.text()).toContain('hello')
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    wrapper.unmount()
  })

  it('falls back to the existing placeholder when no Wails bindings exist', async () => {
    // Development fallback: window.go is absent entirely.
    ;(globalThis as unknown as { window: object }).window = {}
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.dev', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('No logs available for stdout')
    expect(wrapper.find('pre').exists()).toBe(false)
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    wrapper.unmount()
  })
})

describe('ServiceLogs – backend error passthrough', () => {
  it('shows a string rejection from Wails verbatim', async () => {
    installAppMock({
      GetLogs: vi.fn().mockRejectedValue('permission denied reading log file: /var/log/foo/out.log'),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.denied', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('.text-red-400').text()).toBe('permission denied reading log file: /var/log/foo/out.log')
    expect(wrapper.find('pre').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows the message of an Error rejection', async () => {
    installAppMock({
      GetLogs: vi.fn().mockRejectedValue(new Error('boom')),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.err', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('.text-red-400').text()).toBe('boom')
    wrapper.unmount()
  })

  it('falls back to the generic text for a rejection that is neither string nor Error', async () => {
    installAppMock({
      GetLogs: vi.fn().mockRejectedValue({ code: 1 }),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.weird', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('.text-red-400').text()).toBe('Failed to load logs')
    wrapper.unmount()
  })

  it('clears stale error state when switching to a logType that loads fine', async () => {
    installAppMock({
      GetLogs: vi.fn().mockImplementation((_name: string, logType: string) =>
        logType === 'stdout'
          ? Promise.reject('permission denied reading log file: /var/log/foo/out.log')
          : Promise.resolve({ content: 'stderr content', status: 'ok', path: '/tmp/err.log' }),
      ),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.switch', serviceType: 'user' },
    })
    await flushPromises()
    expect(wrapper.find('.text-red-400').exists()).toBe(true)

    const stderrToggle = wrapper.findAll('button').find(b => b.text() === 'stderr')
    expect(stderrToggle).toBeDefined()
    ;(stderrToggle!.element as HTMLButtonElement).click()
    await flushPromises()
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    const pre = wrapper.find('pre')
    expect(pre.exists()).toBe(true)
    expect(pre.text()).toContain('stderr content')
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
