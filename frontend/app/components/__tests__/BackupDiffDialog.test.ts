import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref, computed } from 'vue'
import BackupDiffDialog from '../BackupDiffDialog.vue'

vi.mock('#imports', () => ({
  ref,
  computed,
}))

interface MockPlistContent {
  data: string
  format: 'xml' | 'binary' | 'unknown' | ''
  convertFailed: boolean
}

function installAppMock(overrides: {
  current?: MockPlistContent
  backup?: MockPlistContent
  currentError?: Error
  backupError?: Error
}) {
  const getCurrentPlist = vi.fn(() => {
    if (overrides.currentError) return Promise.reject(overrides.currentError)
    return Promise.resolve(overrides.current ?? { data: '', format: '', convertFailed: false })
  })
  const getBackupContent = vi.fn(() => {
    if (overrides.backupError) return Promise.reject(overrides.backupError)
    return Promise.resolve(overrides.backup ?? { data: '', format: 'xml', convertFailed: false })
  })
  ;(globalThis as { window?: unknown }).window = Object.assign((globalThis as { window?: unknown }).window ?? {}, {
    go: { main: { App: { GetCurrentPlist: getCurrentPlist, GetBackupContent: getBackupContent } } },
  })
  return { getCurrentPlist, getBackupContent }
}

const fakeBackup = {
  id: '20260418-010000',
  service: 'com.example.service',
  timestamp: '2026-04-18T01:00:00Z',
  path: '/tmp/backup.plist',
}

describe('BackupDiffDialog', () => {
  beforeEach(() => {
    installAppMock({
      current: { data: 'line a\nline b\n', format: 'xml', convertFailed: false },
      backup: { data: 'line a\nline c\n', format: 'xml', convertFailed: false },
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders two diff columns with red and green styling', async () => {
    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    const html = wrapper.html()
    expect(html).toContain('bg-red-500/10')
    expect(html).toContain('bg-green-500/10')
    // Two columns present
    const columns = wrapper.findAll('[data-testid="diff-column"]')
    expect(columns).toHaveLength(2)
  })

  it('shows No changes message when content is identical', async () => {
    installAppMock({
      current: { data: 'same\n', format: 'xml', convertFailed: false },
      backup: { data: 'same\n', format: 'xml', convertFailed: false },
    })

    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    expect(wrapper.text()).toMatch(/no changes/i)
  })

  it('shows no-current-version indicator when current plist is absent', async () => {
    installAppMock({
      current: { data: '', format: '', convertFailed: false },
      backup: { data: 'new line\n', format: 'xml', convertFailed: false },
    })

    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    expect(wrapper.text()).toMatch(/no current version/i)
  })

  it('shows truncation notice when diff exceeds row limit', async () => {
    const bigBackup = Array.from({ length: 10_050 }, (_, i) => `row ${i}`).join('\n') + '\n'
    installAppMock({
      current: { data: '', format: '', convertFailed: false },
      backup: { data: bigBackup, format: 'xml', convertFailed: false },
    })

    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    expect(wrapper.text()).toMatch(/truncat/i)
  })

  it('shows format conversion warning banner when convertFailed is true', async () => {
    installAppMock({
      current: { data: 'x', format: 'binary', convertFailed: true },
      backup: { data: 'y', format: 'xml', convertFailed: false },
    })

    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    expect(wrapper.text()).toMatch(/format conversion/i)
  })

  it('emits close when Cancel button is clicked', async () => {
    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    await wrapper.find('[data-testid="diff-cancel"]').trigger('click')
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('emits restore with the backup when Restore button is clicked', async () => {
    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    await wrapper.find('[data-testid="diff-restore"]').trigger('click')
    const emitted = wrapper.emitted('restore')
    expect(emitted).toBeTruthy()
    expect(emitted![0]![0]).toEqual(fakeBackup)
  })

  it('renders an error message when IPC rejects', async () => {
    installAppMock({
      backupError: new Error('backend unavailable'),
    })

    const wrapper = mount(BackupDiffDialog, {
      props: { visible: true, backup: fakeBackup },
      global: { stubs: { teleport: true } },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('backend unavailable')
  })
})
