import { describe, it, expect } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import LaunchPolicyForm from '~/components/LaunchPolicyForm.vue'
import {
  applyLaunchPolicy,
  cloneKeepAlive,
  deriveLaunchPolicy,
  emptyKeepAlive,
  type LaunchPolicy,
} from '~/utils/launchPolicy'
import type { KeepAliveConfig, Service, ServiceConfig } from '~/types/wails'

// Mirrors the launch-policy round-trip in pages/services/[name].vue
// (populateEditForm + handleSave) so we can exercise it without mocking
// useRoute, useAdminMode, and the rest of the page. Production code calls the
// same deriveLaunchPolicy / cloneKeepAlive / applyLaunchPolicy helpers.
const EditHost = defineComponent({
  props: {
    service: { type: Object as () => Service, required: true },
  },
  setup(props) {
    const editLaunchPolicy = ref<LaunchPolicy>('onDemand')
    const editKeepAlive = ref<KeepAliveConfig>(emptyKeepAlive())
    const editThrottleInterval = ref<number | undefined>(undefined)
    const saved = ref<ServiceConfig | null>(null)

    // populateEditForm
    editLaunchPolicy.value = deriveLaunchPolicy({
      runAtLoad: props.service.runAtLoad,
      keepAlive: props.service.keepAlive,
    })
    editKeepAlive.value = cloneKeepAlive(props.service.keepAlive)
    editThrottleInterval.value = props.service.throttleInterval

    function handleSave() {
      const policy = applyLaunchPolicy(editLaunchPolicy.value, editKeepAlive.value)
      saved.value = {
        label: props.service.label,
        runAtLoad: policy.runAtLoad,
        keepAlive: policy.keepAlive,
        throttleInterval: editThrottleInterval.value,
        wakeSystem: props.service.wakeSystem,
      }
    }

    return () => h('div', [
      h(LaunchPolicyForm, {
        'launchPolicy': editLaunchPolicy.value,
        'onUpdate:launchPolicy': (v: LaunchPolicy) => { editLaunchPolicy.value = v },
        'keepAlive': editKeepAlive.value,
        'onUpdate:keepAlive': (v: KeepAliveConfig) => { editKeepAlive.value = v },
        'throttleInterval': editThrottleInterval.value,
        'onUpdate:throttleInterval': (v: number | undefined) => { editThrottleInterval.value = v },
      }),
      h('button', { 'data-testid': 'save', 'onClick': handleSave }, 'Save'),
      h('span', { 'data-testid': 'saved', 'data-config': saved.value ? JSON.stringify(saved.value) : '' }),
    ])
  },
})

function makeService(overrides: Partial<Service> = {}): Service {
  return {
    name: 'com.example.test',
    label: 'com.example.test',
    status: 'stopped',
    statusConfidence: 'verified',
    path: '/Users/dev/Library/LaunchAgents/com.example.test.plist',
    program: '/usr/bin/foo',
    runAtLoad: false,
    keepAlive: emptyKeepAlive(),
    wakeSystem: false,
    type: 'user',
    readOnly: false,
    plistFormat: 'xml',
    ...overrides,
  }
}

function readSaved(wrapper: ReturnType<typeof mount>): ServiceConfig {
  const raw = wrapper.find('[data-testid="saved"]').attributes('data-config')
  return JSON.parse(raw || '{}') as ServiceConfig
}

describe('Edit form — launch policy render', () => {
  it('renders the radio group and the advanced section for a Keep Alive service', () => {
    const wrapper = mount(EditHost, {
      props: { service: makeService({ keepAlive: { enabled: true, mode: 'dictionary', successfulExit: false } }) },
    })
    expect(wrapper.find('[data-testid="launch-policy-group"]').exists()).toBe(true)
    expect((wrapper.find('[data-testid="launch-policy-keepAlive"]').element as HTMLInputElement).checked).toBe(true)
    expect(wrapper.find('[data-testid="keepalive-advanced"]').exists()).toBe(true)
  })
})

describe('Edit form — legacy RunAtLoad + KeepAlive (Task 14)', () => {
  it('lands on Keep Alive and saves without a standalone RunAtLoad', async () => {
    const wrapper = mount(EditHost, {
      props: {
        service: makeService({
          runAtLoad: true,
          keepAlive: { enabled: true, mode: 'dictionary', successfulExit: false },
        }),
      },
    })

    // Loads on Keep Alive despite RunAtLoad=true.
    expect((wrapper.find('[data-testid="launch-policy-keepAlive"]').element as HTMLInputElement).checked).toBe(true)

    await wrapper.find('[data-testid="save"]').trigger('click')
    await flushPromises()

    const sent = readSaved(wrapper)
    expect(sent.runAtLoad).toBe(false)
    expect(sent.keepAlive.enabled).toBe(true)
    expect(sent.keepAlive.mode).toBe('dictionary')
  })
})

describe('Edit form — ThrottleInterval is independent of launch policy', () => {
  it('preserves an existing ThrottleInterval when saving an On Demand service', async () => {
    const wrapper = mount(EditHost, {
      props: { service: makeService({ runAtLoad: false, keepAlive: emptyKeepAlive(), throttleInterval: 30 }) },
    })

    // No KeepAlive, no RunAtLoad → On Demand, advanced section hidden.
    expect((wrapper.find('[data-testid="launch-policy-onDemand"]').element as HTMLInputElement).checked).toBe(true)
    expect(wrapper.find('[data-testid="keepalive-advanced"]').exists()).toBe(false)

    await wrapper.find('[data-testid="save"]').trigger('click')
    await flushPromises()

    // ThrottleInterval must survive the save even though the policy is not Keep Alive.
    expect(readSaved(wrapper).throttleInterval).toBe(30)
  })
})

describe('Edit form — non-editable sub-keys preserved (Task 13)', () => {
  it('editing only SuccessfulExit preserves PathState and ThrottleInterval', async () => {
    const wrapper = mount(EditHost, {
      props: {
        service: makeService({
          runAtLoad: false,
          keepAlive: { enabled: true, mode: 'dictionary', pathState: { '/tmp/flag': true } },
          throttleInterval: 30,
        }),
      },
    })

    await wrapper.find('[data-testid="keepalive-successfulExit"]').setValue('false')
    await wrapper.find('[data-testid="save"]').trigger('click')
    await flushPromises()

    const sent = readSaved(wrapper)
    expect(sent.keepAlive.successfulExit).toBe(false)
    expect(sent.keepAlive.pathState).toEqual({ '/tmp/flag': true })
    expect(sent.throttleInterval).toBe(30)
  })
})
