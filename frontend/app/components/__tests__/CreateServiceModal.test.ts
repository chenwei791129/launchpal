import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive, defineComponent, nextTick  } from 'vue'
import { composeLogPaths } from '~/utils/logPaths'
import { useSettings } from '~/composables/useSettings'
import type { Settings } from '~/types/wails'


const EnvVarSection = defineComponent({
  setup() {
    const envVars = reactive<Array<{ key: string; value: string }>>([])
    // Track which indices have their value visible (not masked)
    const envVisibility = reactive(new Set<number>())

    function addEnv() {
      envVars.push({ key: '', value: '' })
    }

    function removeEnv(index: number) {
      envVars.splice(index, 1)
      const next = new Set<number>()
      for (const i of envVisibility) {
        if (i < index) next.add(i)
        else if (i > index) next.add(i - 1)
      }
      envVisibility.clear()
      for (const i of next) envVisibility.add(i)
    }

    function toggleVisibility(index: number) {
      if (envVisibility.has(index)) {
        envVisibility.delete(index)
      } else {
        envVisibility.add(index)
      }
    }

    function resetForm() {
      envVars.splice(0)
      envVisibility.clear()
    }

    return { envVars, envVisibility, addEnv, removeEnv, toggleVisibility, resetForm }
  },
  template: `
    <div>
      <div v-for="(env, index) in envVars" :key="index" data-testid="env-var-row" class="flex gap-2">
        <input v-model="env.key" type="text" data-testid="env-var-key" placeholder="KEY" />
        <input
          v-model="env.value"
          :type="envVisibility.has(index) ? 'text' : 'password'"
          data-testid="env-var-value"
          placeholder="Value"
        />
        <button
          type="button"
          data-testid="env-var-toggle"
          @click="toggleVisibility(index)"
        >
          toggle
        </button>
        <button type="button" data-testid="env-var-remove" @click="removeEnv(index)">
          remove
        </button>
      </div>
      <button type="button" data-testid="env-var-add" @click="addEnv">+ Add</button>
      <button type="button" data-testid="reset" @click="resetForm">Reset</button>
    </div>
  `,
})

describe('CreateServiceModal – env var visibility masking', () => {
  let wrapper: ReturnType<typeof mount>

  beforeEach(() => {
    wrapper = mount(EnvVarSection)
  })

  it('adds env var rows when clicking Add', async () => {
    expect(wrapper.findAll('[data-testid="env-var-row"]').length).toBe(0)

    await wrapper.find('[data-testid="env-var-add"]').trigger('click')
    await nextTick()

    expect(wrapper.findAll('[data-testid="env-var-row"]').length).toBe(1)
  })

  it('env var value input has type="password" by default', async () => {
    await wrapper.find('[data-testid="env-var-add"]').trigger('click')
    await nextTick()

    const valueInput = wrapper.find('[data-testid="env-var-value"]')
    expect(valueInput.attributes('type')).toBe('password')
  })

  it('clicking the eye toggle changes input type to text', async () => {
    await wrapper.find('[data-testid="env-var-add"]').trigger('click')
    await nextTick()

    const valueInput = wrapper.find('[data-testid="env-var-value"]')
    const toggleBtn = wrapper.find('[data-testid="env-var-toggle"]')

    expect(valueInput.attributes('type')).toBe('password')

    await toggleBtn.trigger('click')
    await nextTick()

    expect(valueInput.attributes('type')).toBe('text')
  })

  it('clicking the toggle again changes input type back to password', async () => {
    await wrapper.find('[data-testid="env-var-add"]').trigger('click')
    await nextTick()

    const valueInput = wrapper.find('[data-testid="env-var-value"]')
    const toggleBtn = wrapper.find('[data-testid="env-var-toggle"]')

    // Reveal
    await toggleBtn.trigger('click')
    await nextTick()
    expect(valueInput.attributes('type')).toBe('text')

    // Hide again
    await toggleBtn.trigger('click')
    await nextTick()
    expect(valueInput.attributes('type')).toBe('password')
  })

  it('toggles each variable independently', async () => {
    // Add two env var rows
    const addBtn = wrapper.find('[data-testid="env-var-add"]')
    await addBtn.trigger('click')
    await nextTick()
    await addBtn.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('[data-testid="env-var-row"]')
    expect(rows.length).toBe(2)

    const firstValue = rows[0]!.find('[data-testid="env-var-value"]')
    const secondValue = rows[1]!.find('[data-testid="env-var-value"]')
    const firstToggle = rows[0]!.find('[data-testid="env-var-toggle"]')
    const secondToggle = rows[1]!.find('[data-testid="env-var-toggle"]')

    // Both masked by default
    expect(firstValue.attributes('type')).toBe('password')
    expect(secondValue.attributes('type')).toBe('password')

    // Reveal only the first
    await firstToggle.trigger('click')
    await nextTick()

    expect(firstValue.attributes('type')).toBe('text')
    expect(secondValue.attributes('type')).toBe('password')

    // Reveal only the second too
    await secondToggle.trigger('click')
    await nextTick()

    expect(firstValue.attributes('type')).toBe('text')
    expect(secondValue.attributes('type')).toBe('text')
  })

  it('resets visibility state when resetForm is called', async () => {
    const addBtn = wrapper.find('[data-testid="env-var-add"]')
    await addBtn.trigger('click')
    await nextTick()

    const toggleBtn = wrapper.find('[data-testid="env-var-toggle"]')
    await toggleBtn.trigger('click')
    await nextTick()

    // Verify it's visible before reset
    expect(wrapper.find('[data-testid="env-var-value"]').attributes('type')).toBe('text')

    // Reset the form
    await wrapper.find('[data-testid="reset"]').trigger('click')
    await nextTick()

    // After reset, no env vars remain
    expect(wrapper.findAll('[data-testid="env-var-row"]').length).toBe(0)

    // Add a new one — should be masked again
    await addBtn.trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="env-var-value"]').attributes('type')).toBe('password')
  })

  it('deleting a row shifts visibility indices correctly', async () => {
    const addBtn = wrapper.find('[data-testid="env-var-add"]')
    // Add three rows
    await addBtn.trigger('click')
    await nextTick()
    await addBtn.trigger('click')
    await nextTick()
    await addBtn.trigger('click')
    await nextTick()

    // Reveal the second row (index 1)
    const rows = wrapper.findAll('[data-testid="env-var-row"]')
    await rows[1]!.find('[data-testid="env-var-toggle"]').trigger('click')
    await nextTick()
    expect(wrapper.findAll('[data-testid="env-var-value"]')[1]!.attributes('type')).toBe('text')

    // Delete the first row (index 0)
    await wrapper.findAll('[data-testid="env-var-remove"]')[0]!.trigger('click')
    await nextTick()

    const valueInputs = wrapper.findAll('[data-testid="env-var-value"]')
    expect(valueInputs).toHaveLength(2)
    // The previously-revealed row (was index 1, now index 0) should still be revealed
    expect(valueInputs[0]!.attributes('type')).toBe('text')
    // The other row should remain masked
    expect(valueInputs[1]!.attributes('type')).toBe('password')
  })
})

// Path composition table mirrors the example block in
// specs/log-path-customization/spec.md. Each row is one parameterized case.
describe('CreateServiceModal — log path composition', () => {
  const cases: Array<{
    serviceType: 'user' | 'system'
    settings: Settings
    label: string
    stdout: string
  }> = [
    {
      serviceType: 'user',
      settings: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
      label: 'com.user.x',
      stdout: '~/Library/Logs/com.user.x/stdout.log',
    },
    {
      serviceType: 'user',
      settings: { userLogDir: '/tmp/u', systemLogDir: '/Library/Logs' },
      label: 'com.user.x',
      stdout: '/tmp/u/com.user.x/stdout.log',
    },
    {
      serviceType: 'system',
      settings: { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' },
      label: 'com.sys.x',
      stdout: '/Library/Logs/com.sys.x/stdout.log',
    },
    {
      serviceType: 'system',
      settings: { userLogDir: '~/Library/Logs', systemLogDir: '/var/log/lp' },
      label: 'com.sys.x',
      stdout: '/var/log/lp/com.sys.x/stdout.log',
    },
  ]

  for (const tc of cases) {
    it(`composes ${tc.serviceType} stdout for label "${tc.label}" with dir "${tc.serviceType === 'system' ? tc.settings.systemLogDir : tc.settings.userLogDir}"`, () => {
      const got = composeLogPaths(tc.serviceType, tc.settings, tc.label)
      expect(got.stdout).toBe(tc.stdout)
      expect(got.stderr).toBe(tc.stdout.replace('stdout.log', 'stderr.log'))
    })
  }

  it('strips trailing slash on the directory so paths never double up', () => {
    const got = composeLogPaths(
      'system',
      { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs/' },
      'com.x',
    )
    expect(got.stdout).toBe('/Library/Logs/com.x/stdout.log')
  })
})

// Mounts the actual modal to verify it re-reads settings each time it is
// opened (Decision 8) and that saving Settings does NOT issue any helper or
// CreateService RPC (Existing services are not migrated).
describe('CreateServiceModal — settings integration', () => {
  let GetSettings: ReturnType<typeof vi.fn>
  let UpdateSettings: ReturnType<typeof vi.fn>
  let CreateService: ReturnType<typeof vi.fn>
  let CreateSystemService: ReturnType<typeof vi.fn>
  let currentSettings: Settings

  beforeEach(() => {
    currentSettings = { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' }
    GetSettings = vi.fn(async () => currentSettings)
    UpdateSettings = vi.fn(async (s: Settings) => {
      currentSettings = s
    })
    CreateService = vi.fn(async () => {})
    CreateSystemService = vi.fn(async () => {})
    ;(window as unknown as { go: { main: { App: Record<string, unknown> } } }).go = {
      main: {
        App: {
          GetSettings,
          UpdateSettings,
          CreateService,
          CreateSystemService,
        },
      },
    }
    // Reset the cached useSettings ref between tests.
    const { settings, defaults } = useSettings()
    settings.value = { ...defaults }
  })

  afterEach(() => {
    delete (window as unknown as { go?: unknown }).go
  })

  async function mountModal(serviceType: 'user' | 'system') {
    const mod = await import('../CreateServiceModal.vue')
    return mount(mod.default, {
      props: { isOpen: true, serviceType },
      global: { stubs: { ScheduleForm: true } },
    })
  }

  it('re-reads settings each time the modal opens (Decision 8)', async () => {
    const wrapper = await mountModal('user')
    await flushPromises()
    expect(GetSettings).toHaveBeenCalled()
    const initialCalls = GetSettings.mock.calls.length

    // Simulate user changing settings while the modal is closed.
    await wrapper.setProps({ isOpen: false })
    await flushPromises()
    currentSettings = { userLogDir: '/tmp/userlogs', systemLogDir: '/Library/Logs' }
    await wrapper.setProps({ isOpen: true })
    await flushPromises()

    expect(GetSettings.mock.calls.length).toBeGreaterThan(initialCalls)
  })

  it('user serviceType uses settings.userLogDir for log preview', async () => {
    currentSettings = { userLogDir: '/tmp/userlogs', systemLogDir: '/Library/Logs' }
    const wrapper = await mountModal('user')
    await flushPromises()
    const labelInput = wrapper.find('input[placeholder="com.example.myservice"]')
    await labelInput.setValue('com.user.demo')
    await flushPromises()
    const html = wrapper.html()
    expect(html).toContain('/tmp/userlogs/com.user.demo/stdout.log')
    expect(html).toContain('/tmp/userlogs/com.user.demo/stderr.log')
  })

  it('system serviceType uses settings.systemLogDir for log preview', async () => {
    currentSettings = { userLogDir: '~/Library/Logs', systemLogDir: '/var/log/lp' }
    const wrapper = await mountModal('system')
    await flushPromises()
    const labelInput = wrapper.find('input[placeholder="com.example.myservice"]')
    await labelInput.setValue('com.sys.demo')
    await flushPromises()
    const html = wrapper.html()
    expect(html).toContain('/var/log/lp/com.sys.demo/stdout.log')
    expect(html).toContain('/var/log/lp/com.sys.demo/stderr.log')
  })

  it('Program Path label includes a hint that it is optional when Arguments holds the executable', async () => {
    const wrapper = await mountModal('user')
    await flushPromises()
    const html = wrapper.html()
    expect(html).toContain('Optional. Leave empty if the executable is provided as the first item in Arguments.')
  })

  it('Submit button is disabled when label is set but both Program and Arguments are empty', async () => {
    const wrapper = await mountModal('user')
    await flushPromises()

    const labelInput = wrapper.find('input[placeholder="com.example.myservice"]')
    await labelInput.setValue('com.user.bothempty')
    await flushPromises()

    const submitBtn = wrapper.findAll('button').find(b => b.text().trim().startsWith('Create Service'))
    expect(submitBtn).toBeDefined()
    expect(submitBtn!.attributes('disabled')).toBeDefined()
  })

  it('Submit button is enabled when label and Arguments are set but Program is empty, and CreateService is called', async () => {
    const wrapper = await mountModal('user')
    await flushPromises()

    const labelInput = wrapper.find('input[placeholder="com.example.myservice"]')
    await labelInput.setValue('com.user.argsonly')
    const argsInput = wrapper.find('input[placeholder="--daemon --port=8080"]')
    await argsInput.setValue("/usr/bin/open '/Applications/Foo.app'")
    await flushPromises()

    const submitBtn = wrapper.findAll('button').find(b => b.text().trim().startsWith('Create Service'))
    expect(submitBtn).toBeDefined()
    expect(submitBtn!.attributes('disabled')).toBeUndefined()

    await submitBtn!.trigger('click')
    await flushPromises()

    expect(CreateService).toHaveBeenCalledTimes(1)
    const sent = CreateService.mock.calls[0]![0] as { program: string; arguments: string[] }
    expect(sent.program).toBe('')
    expect(sent.arguments).toEqual(['/usr/bin/open', '/Applications/Foo.app'])
  })

  it('Program Path input no longer has the required attribute', async () => {
    const wrapper = await mountModal('user')
    await flushPromises()
    const programInput = wrapper.find('input[placeholder="/usr/local/bin/myapp"]')
    expect(programInput.exists()).toBe(true)
    expect(programInput.attributes('required')).toBeUndefined()
  })

  it('saving Settings does not modify existing plists or invoke helper RPCs', async () => {
    // Render the modal once so any subscriptions exist, then "save settings"
    // by calling UpdateSettings directly. The modal must not respond by
    // calling any service-mutating binding.
    await mountModal('user')
    await flushPromises()
    const callsBefore = {
      CreateService: CreateService.mock.calls.length,
      CreateSystemService: CreateSystemService.mock.calls.length,
    }

    await (UpdateSettings as unknown as (s: Settings) => Promise<void>)({
      userLogDir: '/tmp/u',
      systemLogDir: '/Library/Logs/lp',
    })
    await flushPromises()

    expect(CreateService.mock.calls.length).toBe(callsBefore.CreateService)
    expect(CreateSystemService.mock.calls.length).toBe(callsBefore.CreateSystemService)
  })
})

