import { describe, it, expect } from 'vitest'
import {
  applyLaunchPolicy,
  cloneKeepAlive,
  cloneLaunchPolicy,
  deriveLaunchPolicy,
  emptyKeepAlive,
  hasEffectiveKeepAliveSubKey,
  normalizeKeepAliveForSubmit,
} from '~/utils/launchPolicy'
import type { KeepAliveConfig } from '~/types/wails'

describe('deriveLaunchPolicy', () => {
  it('prefers Keep Alive when keepAlive is enabled, even with runAtLoad', () => {
    expect(deriveLaunchPolicy({ runAtLoad: true, keepAlive: { enabled: true, mode: 'dictionary' } }))
      .toBe('keepAlive')
  })
  it('maps runAtLoad-only to runAtLoad', () => {
    expect(deriveLaunchPolicy({ runAtLoad: true, keepAlive: emptyKeepAlive() })).toBe('runAtLoad')
  })
  it('maps neither to onDemand', () => {
    expect(deriveLaunchPolicy({ runAtLoad: false, keepAlive: emptyKeepAlive() })).toBe('onDemand')
  })
})

describe('cloneLaunchPolicy', () => {
  it('preserves Keep Alive sources', () => {
    expect(cloneLaunchPolicy({ runAtLoad: true, keepAlive: { enabled: true, mode: 'boolean' } }))
      .toBe('keepAlive')
  })
  it('downgrades a Run at Load source to On Demand', () => {
    expect(cloneLaunchPolicy({ runAtLoad: true, keepAlive: emptyKeepAlive() })).toBe('onDemand')
  })
  it('keeps an On Demand source on On Demand', () => {
    expect(cloneLaunchPolicy({ runAtLoad: false, keepAlive: emptyKeepAlive() })).toBe('onDemand')
  })
})

describe('applyLaunchPolicy', () => {
  it('On Demand writes neither RunAtLoad nor KeepAlive', () => {
    const r = applyLaunchPolicy('onDemand', emptyKeepAlive())
    expect(r.runAtLoad).toBe(false)
    expect(r.keepAlive.enabled).toBe(false)
  })
  it('Run at Load writes RunAtLoad and no KeepAlive', () => {
    const r = applyLaunchPolicy('runAtLoad', emptyKeepAlive())
    expect(r.runAtLoad).toBe(true)
    expect(r.keepAlive.enabled).toBe(false)
  })
  it('Keep Alive writes KeepAlive and never RunAtLoad', () => {
    const r = applyLaunchPolicy('keepAlive', { enabled: true, mode: 'boolean' })
    expect(r.runAtLoad).toBe(false)
    expect(r.keepAlive.enabled).toBe(true)
    expect(r.keepAlive.mode).toBe('boolean')
  })
  it('Keep Alive with a dictionary sub-key keeps dictionary mode', () => {
    const r = applyLaunchPolicy('keepAlive', { enabled: true, mode: 'dictionary', successfulExit: false })
    expect(r.keepAlive.mode).toBe('dictionary')
    expect(r.keepAlive.successfulExit).toBe(false)
  })
  it('Keep Alive with an empty dictionary downgrades to boolean', () => {
    const r = applyLaunchPolicy('keepAlive', { enabled: true, mode: 'dictionary' })
    expect(r.keepAlive.mode).toBe('boolean')
    expect(r.keepAlive.enabled).toBe(true)
  })
  it('Keep Alive preserves non-editable sub-keys (PathState)', () => {
    const r = applyLaunchPolicy('keepAlive', {
      enabled: true, mode: 'dictionary', pathState: { '/tmp/flag': true },
    })
    expect(r.keepAlive.mode).toBe('dictionary')
    expect(r.keepAlive.pathState).toEqual({ '/tmp/flag': true })
  })
})

describe('hasEffectiveKeepAliveSubKey', () => {
  it('is false for an empty dictionary', () => {
    expect(hasEffectiveKeepAliveSubKey({ enabled: true, mode: 'dictionary' })).toBe(false)
  })
  it('is true when a preserved PathState is present', () => {
    expect(hasEffectiveKeepAliveSubKey({ enabled: true, mode: 'dictionary', pathState: { '/x': true } }))
      .toBe(true)
  })
})

describe('normalizeKeepAliveForSubmit', () => {
  it('downgrades an empty dictionary to boolean', () => {
    expect(normalizeKeepAliveForSubmit({ enabled: true, mode: 'dictionary' }))
      .toEqual({ enabled: true, mode: 'boolean' })
  })
  it('leaves a dictionary with sub-keys unchanged', () => {
    const ka: KeepAliveConfig = { enabled: true, mode: 'dictionary', crashed: true }
    expect(normalizeKeepAliveForSubmit(ka)).toBe(ka)
  })
})

describe('cloneKeepAlive', () => {
  it('deep-copies nested maps', () => {
    const src: KeepAliveConfig = { enabled: true, mode: 'dictionary', pathState: { '/x': true } }
    const copy = cloneKeepAlive(src)
    copy.pathState!['/x'] = false
    expect(src.pathState!['/x']).toBe(true)
  })
  it('returns a disabled config for undefined input', () => {
    expect(cloneKeepAlive(undefined)).toEqual({ enabled: false, mode: '' })
  })
})
