import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import type { Settings } from '~/types/wails'
import LogStorageSection from '~/components/LogStorageSection.vue'
import { useSettings } from '~/composables/useSettings'

interface AppBindings {
  GetSettings: ReturnType<typeof vi.fn>
  UpdateSettings: ReturnType<typeof vi.fn>
}

function installBindings(opts: {
  initial?: Settings
  updateImpl?: (s: Settings) => Promise<void>
}): AppBindings {
  const initial = opts.initial ?? { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' }
  let current: Settings = { ...initial }
  const GetSettings = vi.fn(async () => current)
  const UpdateSettings = vi.fn(async (s: Settings) => {
    if (opts.updateImpl) {
      await opts.updateImpl(s)
    }
    current = s
  })
  // Mutate the existing happy-dom window — replacing it wholesale would
  // strip away Event / MouseEvent constructors that vue-test-utils relies on.
  ;(window as unknown as { go: { main: { App: AppBindings } } }).go = {
    main: { App: { GetSettings, UpdateSettings } },
  }
  return { GetSettings, UpdateSettings }
}

describe('Log Storage section', () => {
  beforeEach(() => {
    // Reset module-level state so each test starts from defaults — useSettings
    // caches across all consumers, including across tests.
    const { settings, defaults } = useSettings()
    settings.value = { ...defaults }
  })
  afterEach(() => {
    delete (window as unknown as { go?: unknown }).go
  })

  it('initial render shows defaults from GetSettings', async () => {
    installBindings({ initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' } })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    const userInput = wrapper.find('[data-testid="user-log-dir-input"]')
    const systemInput = wrapper.find('[data-testid="system-log-dir-input"]')
    expect((userInput.element as HTMLInputElement).value).toBe('~/Library/Logs')
    expect((systemInput.element as HTMLInputElement).value).toBe('/Library/Logs')
  })

  it('reset restores default for User Log Directory', async () => {
    installBindings({ initial: { userLogDir: '/tmp/userlogs', systemLogDir: '/Library/Logs' } })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    expect((wrapper.find('[data-testid="user-log-dir-input"]').element as HTMLInputElement).value).toBe('/tmp/userlogs')
    await wrapper.find('[data-testid="user-log-dir-reset"]').trigger('click')
    expect((wrapper.find('[data-testid="user-log-dir-input"]').element as HTMLInputElement).value).toBe('~/Library/Logs')
  })

  it('Save failure (client-side) shows inline error and keeps the typed value', async () => {
    installBindings({ initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' } })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    const systemInput = wrapper.find('[data-testid="system-log-dir-input"]')
    await systemInput.setValue('/etc/launchpal/logs')
    await wrapper.find('[data-testid="system-log-dir-save"]').trigger('click')
    await flushPromises()

    const err = wrapper.find('[data-testid="system-log-dir-error"]')
    expect(err.exists()).toBe(true)
    expect(err.text()).toContain('systemLogDir')
    expect((systemInput.element as HTMLInputElement).value).toBe('/etc/launchpal/logs')
  })

  it('Save failure (backend) surfaces the backend error verbatim', async () => {
    // Disable client-side validation drift detection by feeding a value the
    // frontend accepts; the backend mock then rejects to simulate a drift
    // between the two validators.
    installBindings({
      initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
      updateImpl: async () => {
        throw new Error('systemLogDir: rejected by backend (simulated drift)')
      },
    })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    const systemInput = wrapper.find('[data-testid="system-log-dir-input"]')
    await systemInput.setValue('/Library/Logs/launchpal')
    await wrapper.find('[data-testid="system-log-dir-save"]').trigger('click')
    await flushPromises()

    const err = wrapper.find('[data-testid="system-log-dir-error"]')
    expect(err.exists()).toBe(true)
    expect(err.text()).toContain('rejected by backend')
    expect((systemInput.element as HTMLInputElement).value).toBe('/Library/Logs/launchpal')
  })

  it('successful Save triggers GetSettings refresh and shows success indicator', async () => {
    const { GetSettings, UpdateSettings } = installBindings({
      initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
    })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    const initialGetCalls = GetSettings.mock.calls.length

    const systemInput = wrapper.find('[data-testid="system-log-dir-input"]')
    await systemInput.setValue('/Library/Logs/launchpal')
    await wrapper.find('[data-testid="system-log-dir-save"]').trigger('click')
    await flushPromises()

    expect(UpdateSettings).toHaveBeenCalledWith({
      userLogDir: '~/Library/Logs',
      systemLogDir: '/Library/Logs/launchpal',
    })
    // Refresh after save: GetSettings was called more than the initial mount.
    expect(GetSettings.mock.calls.length).toBeGreaterThan(initialGetCalls)
    expect(wrapper.find('[data-testid="system-log-dir-success"]').exists()).toBe(true)
  })

  it('User Save ignores unsaved System input — sends saved systemLogDir, succeeds', async () => {
    // Reproduces the reported bug: typing an invalid value into System Log
    // Directory must not break clicking User Log Directory's Save.
    const { UpdateSettings } = installBindings({
      initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
    })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    // User types a junk value into the System input but does NOT click its Save.
    await wrapper.find('[data-testid="system-log-dir-input"]').setValue('/Library/Logsdd')
    // User changes their User input and clicks the User Save.
    await wrapper.find('[data-testid="user-log-dir-input"]').setValue('/tmp/userlogs')
    await wrapper.find('[data-testid="user-log-dir-save"]').trigger('click')
    await flushPromises()

    // UpdateSettings must be called with the saved systemLogDir (last good
    // value), NOT the unsaved junk in the System input.
    expect(UpdateSettings).toHaveBeenCalledWith({
      userLogDir: '/tmp/userlogs',
      systemLogDir: '/Library/Logs',
    })
    // No error appears under either field.
    expect(wrapper.find('[data-testid="user-log-dir-error"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="system-log-dir-error"]').exists()).toBe(false)
    // The unsaved system input value is preserved so the user can keep editing.
    expect(
      (wrapper.find('[data-testid="system-log-dir-input"]').element as HTMLInputElement).value,
    ).toBe('/Library/Logsdd')
  })

  it('System Save ignores unsaved User input', async () => {
    const { UpdateSettings } = installBindings({
      initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
    })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    await wrapper.find('[data-testid="user-log-dir-input"]').setValue('') // intentionally invalid
    await wrapper.find('[data-testid="system-log-dir-input"]').setValue('/var/log/launchpal')
    await wrapper.find('[data-testid="system-log-dir-save"]').trigger('click')
    await flushPromises()

    expect(UpdateSettings).toHaveBeenCalledWith({
      userLogDir: '~/Library/Logs',
      systemLogDir: '/var/log/launchpal',
    })
    expect(wrapper.find('[data-testid="user-log-dir-error"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="system-log-dir-error"]').exists()).toBe(false)
  })

  it('Save button is disabled while in flight', async () => {
    let resolveUpdate!: () => void
    const { UpdateSettings } = installBindings({
      initial: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
      updateImpl: () =>
        new Promise<void>((res) => {
          resolveUpdate = res
        }),
    })
    const wrapper = mount(LogStorageSection)
    await flushPromises()

    const saveBtn = wrapper.find('[data-testid="system-log-dir-save"]')
    await wrapper.find('[data-testid="system-log-dir-input"]').setValue('/Library/Logs/launchpal')
    saveBtn.trigger('click')
    await flushPromises()
    expect((saveBtn.element as HTMLButtonElement).disabled).toBe(true)

    resolveUpdate()
    await flushPromises()
    expect((saveBtn.element as HTMLButtonElement).disabled).toBe(false)
    expect(UpdateSettings).toHaveBeenCalled()
  })
})
