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

// Auto-refresh polling suites use fake timers. flushPromises() resolves via
// setImmediate/setTimeout(0), which fake timers intercept, so these suites
// drive the clock with vi.advanceTimersByTimeAsync (which also flushes the
// pending microtasks) instead of flushPromises().
const AUTO_REFRESH_INTERVAL_MS = 2000

// Shared fake-timer lifecycle for every Auto-refresh suite. The useRealTimers()
// teardown is mandatory: the outer afterEach only runs restoreAllMocks(), and
// leaked fake timers would stall confirmClear's real setTimeout(2000).
function useAutoRefreshFakeTimers() {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })
}

// Flush the microtask queue (mount loads, watchers) without advancing wall
// time beyond the timers scheduled at the current instant.
const settle = () => vi.advanceTimersByTimeAsync(0)

function isChecked(wrapper: ReturnType<typeof mount>, testid: string) {
  return (wrapper.find(`[data-testid="${testid}"]`).element as HTMLInputElement).checked
}
// installAppMock() replaces globalThis.window with a plain object, which strips
// the Event constructors @vue/test-utils' setValue/trigger reach through
// window. Flip the checkbox's checked state and dispatch a change event with
// the top-level Event global (unaffected by the window swap) so Vue's v-model
// updates the ref — jsdom's native .click() toggles .checked without firing
// change for a checkbox, so the ref would otherwise stay stale.
async function clickToggle(wrapper: ReturnType<typeof mount>, testid: string) {
  const el = wrapper.find(`[data-testid="${testid}"]`).element as HTMLInputElement
  el.checked = !el.checked
  el.dispatchEvent(new Event('change'))
  await settle()
}

describe('ServiceLogs – Auto-refresh toggle in the Logs tab', () => {
  useAutoRefreshFakeTimers()

  it('renders an unchecked Auto-refresh checkbox that does not poll by default', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    const toggle = wrapper.find('[data-testid="auto-refresh-toggle"]')
    expect(toggle.exists()).toBe(true)
    expect((toggle.element as HTMLInputElement).checked).toBe(false)
    // Label text is English, matching the rest of the UI.
    expect(wrapper.text()).toContain('Auto-refresh')
    // Only the on-mount load happened; leaving the toggle off must not poll.
    const callsAfterMount = mockedApp.GetLogs.mock.calls.length
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterMount)
    wrapper.unmount()
  })

  it('toggles independently: Auto-refresh on with Auto-scroll off does not force scroll', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    // Auto-scroll defaults to on; turn it off so a polled reload must not move
    // the scroll position.
    await clickToggle(wrapper, 'auto-scroll-toggle')
    const container = wrapper.find('.overflow-auto').element as HTMLElement
    Object.defineProperty(container, 'scrollHeight', { value: 500, configurable: true })
    container.scrollTop = 42
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(true)
    expect(isChecked(wrapper, 'auto-scroll-toggle')).toBe(false)
    // Auto-scroll off: the reload left the user's scroll position untouched.
    expect(container.scrollTop).toBe(42)
    wrapper.unmount()
  })

  it('follow mode: Auto-refresh + Auto-scroll both on scrolls to bottom after each polled reload', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    const container = wrapper.find('.overflow-auto').element as HTMLElement
    Object.defineProperty(container, 'scrollHeight', { value: 500, configurable: true })
    container.scrollTop = 0
    // Auto-scroll stays on (default); enable Auto-refresh.
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(isChecked(wrapper, 'auto-scroll-toggle')).toBe(true)
    expect(container.scrollTop).toBe(500)
    wrapper.unmount()
  })
})

describe('ServiceLogs – Periodic reload while Auto-refresh is enabled', () => {
  useAutoRefreshFakeTimers()

  it('reloads through the manual Refresh path once every 2 seconds', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    const callsAfterMount = mockedApp.GetLogs.mock.calls.length
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterMount + 1)
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterMount + 2)
    wrapper.unmount()
  })

  it('skips a tick while a previous load is still in flight', async () => {
    interface LogsResultShape { content: string, status: string, path: string }
    // A load that never resolves keeps loading=true, so every tick must skip.
    installAppMock({
      GetLogs: vi.fn().mockReturnValue(new Promise<LogsResultShape>(() => {})),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    // The on-mount load is still pending (loading=true).
    const callsWhilePending = mockedApp.GetLogs.mock.calls.length
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsWhilePending)
    wrapper.unmount()
  })

  it('stops polling immediately when the checkbox is unchecked', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    const callsBeforeStop = mockedApp.GetLogs.mock.calls.length
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsBeforeStop)
    wrapper.unmount()
  })

  it('clears the timer on unmount so no further loads fire', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    const callsBeforeUnmount = mockedApp.GetLogs.mock.calls.length
    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsBeforeUnmount)
  })

  it('resets the toggle to unchecked and stops polling when serviceName changes', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.a', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    // Service-to-service navigation reuses the component (no remount).
    await wrapper.setProps({ serviceName: 'com.user.b' })
    await settle()
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(false)
    const callsAfterNav = mockedApp.GetLogs.mock.calls.length
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterNav)
    wrapper.unmount()
  })

  it('keeps the toggle and continues polling the newly selected stream after a stdout→stderr switch', async () => {
    installAppMock({
      GetLogs: vi.fn().mockImplementation((_name: string, logType: string) =>
        Promise.resolve({ content: `${logType} content`, status: 'ok', path: `/tmp/${logType}.log` }),
      ),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    const stderrToggle = wrapper.findAll('button').find(b => b.text() === 'stderr')
    expect(stderrToggle).toBeDefined()
    ;(stderrToggle!.element as HTMLButtonElement).click()
    await settle()
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(true)
    mockedApp.GetLogs.mockClear()
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(mockedApp.GetLogs).toHaveBeenCalled()
    // Every polled load after the switch targets stderr.
    for (const call of mockedApp.GetLogs.mock.calls) {
      expect(call[1]).toBe('stderr')
    }
    wrapper.unmount()
  })
})

describe('ServiceLogs – Auto-refresh disables itself on a non-ok load outcome', () => {
  useAutoRefreshFakeTimers()

  it('disables and stops polling when a polled load resolves not-found, showing the placeholder', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    // The next polled load resolves not-found.
    mockedApp.GetLogs.mockResolvedValue({ content: '', status: 'not-found', path: '/var/log/foo/out.log' })
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(false)
    expect(wrapper.text()).toContain('Log file does not exist yet')
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    // No automatic resume: further ticks do not reload.
    const callsAfterDisable = mockedApp.GetLogs.mock.calls.length
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterDisable)
    wrapper.unmount()
  })

  it('disables and stops polling when a polled load rejects, showing the backend message', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    mockedApp.GetLogs.mockRejectedValue('permission denied reading log file: /var/log/foo/out.log')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(false)
    expect(wrapper.find('.text-red-400').text()).toBe('permission denied reading log file: /var/log/foo/out.log')
    const callsAfterDisable = mockedApp.GetLogs.mock.calls.length
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterDisable)
    wrapper.unmount()
  })

  it('disables when the manual Refresh button triggers a rejecting load', async () => {
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    mockedApp.GetLogs.mockRejectedValue(new Error('boom'))
    ;(wrapper.find('button[title="Refresh logs"]').element as HTMLButtonElement).click()
    await settle()
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(false)
    const callsAfterDisable = mockedApp.GetLogs.mock.calls.length
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS * 3)
    expect(mockedApp.GetLogs.mock.calls.length).toBe(callsAfterDisable)
    wrapper.unmount()
  })

  it('disables when switching to a stream whose load resolves no-path, showing the placeholder', async () => {
    installAppMock({
      GetLogs: vi.fn().mockImplementation((_name: string, logType: string) =>
        logType === 'stdout'
          ? Promise.resolve({ content: 'stdout content', status: 'ok', path: '/tmp/out.log' })
          : Promise.resolve({ content: '', status: 'no-path', path: '' }),
      ),
    })
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.x', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    const stderrToggle = wrapper.findAll('button').find(b => b.text() === 'stderr')
    expect(stderrToggle).toBeDefined()
    ;(stderrToggle!.element as HTMLButtonElement).click()
    await settle()
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(false)
    expect(wrapper.text()).toContain('No stderr log path configured for this service')
    expect(wrapper.find('.text-red-400').exists()).toBe(false)
    wrapper.unmount()
  })

  it('disables when a polled load completes via the development fallback (no bindings)', async () => {
    // Development fallback: window.go is absent entirely.
    ;(globalThis as unknown as { window: object }).window = {}
    const wrapper = mount(ServiceLogs, {
      props: { serviceName: 'com.user.dev', serviceType: 'user' },
    })
    await settle()
    await clickToggle(wrapper, 'auto-refresh-toggle')
    await vi.advanceTimersByTimeAsync(AUTO_REFRESH_INTERVAL_MS)
    expect(isChecked(wrapper, 'auto-refresh-toggle')).toBe(false)
    wrapper.unmount()
  })
})
