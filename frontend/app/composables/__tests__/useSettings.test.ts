import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Settings } from '~/types/wails'

interface AppBindings {
  GetSettings: ReturnType<typeof vi.fn>
  UpdateSettings: ReturnType<typeof vi.fn>
}

function installBindings(get: () => Promise<Settings>, update: (s: Settings) => Promise<void>): AppBindings {
  const GetSettings = vi.fn(get)
  const UpdateSettings = vi.fn(update)
  ;(globalThis as unknown as { window: { go?: { main?: { App?: AppBindings } } } }).window = {
    go: { main: { App: { GetSettings, UpdateSettings } } },
  }
  return { GetSettings, UpdateSettings }
}

const defaults: Settings = { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' }

describe('useSettings composable', () => {
  beforeEach(() => {
    // Reset module-level singleton state between tests so each test starts
    // fresh — useSettings caches the loaded value across consumers.
    vi.resetModules()
  })

  afterEach(() => {
    delete (globalThis as { window?: unknown }).window
  })

  it('load() reads via GetSettings and exposes the result', async () => {
    const want: Settings = { userLogDir: '/tmp/u', systemLogDir: '/Library/Logs/launchpal' }
    const bindings = installBindings(async () => want, async () => {})
    const { useSettings } = await import('../useSettings')
    const composable = useSettings()
    await composable.load()
    expect(bindings.GetSettings).toHaveBeenCalledTimes(1)
    expect(composable.settings.value).toEqual(want)
  })

  it('save() updates local cache on success', async () => {
    const bindings = installBindings(async () => defaults, async () => {})
    const { useSettings } = await import('../useSettings')
    const composable = useSettings()
    await composable.load()

    const next: Settings = { userLogDir: '/tmp/u', systemLogDir: '/Library/Logs/launchpal' }
    await composable.save(next)

    expect(bindings.UpdateSettings).toHaveBeenCalledWith(next)
    expect(composable.settings.value).toEqual(next)
  })

  it('save() does not update cache and surfaces error on validation failure', async () => {
    const bindings = installBindings(
      async () => defaults,
      async () => {
        throw new Error('systemLogDir: must start with one of: /var/log/, /Library/Logs/')
      },
    )
    const { useSettings } = await import('../useSettings')
    const composable = useSettings()
    await composable.load()
    expect(composable.settings.value).toEqual(defaults)

    const bad: Settings = { userLogDir: '~/Library/Logs', systemLogDir: '/etc/foo' }
    await expect(composable.save(bad)).rejects.toThrow(/systemLogDir/)

    expect(bindings.UpdateSettings).toHaveBeenCalledWith(bad)
    expect(composable.settings.value).toEqual(defaults)
  })

  it('shares state across multiple useSettings() callers', async () => {
    installBindings(async () => defaults, async () => {})
    const { useSettings } = await import('../useSettings')
    const a = useSettings()
    const b = useSettings()
    await a.load()
    expect(b.settings.value).toEqual(defaults)
  })
})
