import { describe, it, expect } from 'vitest'
import { serviceToConfig } from '~/utils/serviceToConfig'
import type { Service } from '~/types/wails'

function makeService(overrides: Partial<Service> = {}): Service {
  return {
    name: 'com.example.svc',
    label: 'com.example.svc',
    status: 'stopped',
    statusConfidence: 'verified',
    path: '/Users/dev/Library/LaunchAgents/com.example.svc.plist',
    program: '/usr/bin/foo',
    arguments: ['--port=8080'],
    runAtLoad: true,
    keepAlive: { enabled: false, mode: '' },
    wakeSystem: false,
    type: 'user',
    readOnly: false,
    plistFormat: 'xml',
    ...overrides,
  }
}

describe('serviceToConfig', () => {
  it('carries the structured keepAlive object and throttleInterval', () => {
    const svc = makeService({
      keepAlive: {
        enabled: true,
        mode: 'dictionary',
        successfulExit: false,
        pathState: { '/tmp/flag': true },
      },
      throttleInterval: 30,
    })

    const config = serviceToConfig(svc)

    expect(config.keepAlive).toEqual({
      enabled: true,
      mode: 'dictionary',
      successfulExit: false,
      pathState: { '/tmp/flag': true },
    })
    expect(config.throttleInterval).toBe(30)
  })

  it('deep-copies keepAlive so mutating the config does not affect the source', () => {
    const svc = makeService({
      keepAlive: { enabled: true, mode: 'dictionary', pathState: { '/tmp/flag': true } },
    })

    const config = serviceToConfig(svc)
    config.keepAlive.enabled = false
    config.keepAlive.pathState!['/tmp/flag'] = false

    expect(svc.keepAlive.enabled).toBe(true)
    expect(svc.keepAlive.pathState!['/tmp/flag']).toBe(true)
  })

  it('leaves throttleInterval undefined when the source has none', () => {
    const config = serviceToConfig(makeService())
    expect(config.throttleInterval).toBeUndefined()
  })
})
