import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, ref, type Ref } from 'vue'
import type { Service, ServiceConfig, Settings } from '~/types/wails'
import CreateServiceModal from '../CreateServiceModal.vue'
import { useSettings } from '~/composables/useSettings'
import { serviceToConfig } from '~/utils/serviceToConfig'

// Mirrors the Copy/clone state machine in pages/services/[name].vue so we can
// exercise it without mocking useRoute, useAdminMode, etc. Production code
// imports the same `serviceToConfig` helper.
const CloneHost = defineComponent({
  components: { CreateServiceModal },
  props: {
    service: { type: Object as () => Service | null, default: null },
    serviceType: { type: String as () => 'user' | 'system' | 'apple-system', default: 'user' },
    navigate: { type: Function as unknown as () => (path: string) => void, required: true },
  },
  setup(props) {
    const showCloneModal = ref(false)
    const cloneSource: Ref<ServiceConfig | null> = ref(null)

    function handleCopy() {
      if (!props.service) return
      cloneSource.value = serviceToConfig(props.service)
      showCloneModal.value = true
    }

    function closeCloneModal() {
      showCloneModal.value = false
      cloneSource.value = null
    }

    function handleCloneCreated(label: string) {
      showCloneModal.value = false
      cloneSource.value = null
      props.navigate(`/services/${encodeURIComponent(label)}?type=user`)
    }

    return { showCloneModal, cloneSource, handleCopy, closeCloneModal, handleCloneCreated }
  },
  template: `
    <div>
      <button
        v-if="serviceType === 'user'"
        data-testid="copy-service-button"
        :disabled="!service"
        @click="handleCopy"
      >Copy</button>
      <CreateServiceModal
        :is-open="showCloneModal"
        service-type="user"
        :prefill="cloneSource"
        @close="closeCloneModal"
        @created="handleCloneCreated"
      />
    </div>
  `,
})

function makeService(overrides: Partial<Service> = {}): Service {
  return {
    name: 'com.example.copy-source',
    label: 'com.example.copy-source',
    status: 'running',
    pid: 1234,
    path: '/Users/dev/Library/LaunchAgents/com.example.copy-source.plist',
    program: '/usr/bin/foo',
    arguments: ['--port=8080'],
    runAtLoad: true,
    keepAlive: { enabled: true, mode: 'boolean' },
    wakeSystem: false,
    schedule: { interval: 60 },
    environment: { LOG_LEVEL: 'debug' },
    stdoutPath: '~/Library/Logs/com.example.copy-source/stdout.log',
    stderrPath: '~/Library/Logs/com.example.copy-source/stderr.log',
    workingDirectory: '/Users/dev/project',
    type: 'user',
    readOnly: false,
    plistFormat: 'xml',
    statusConfidence: 'verified',
    ...overrides,
  }
}

describe('Clone user service — Copy button visibility', () => {
  it('renders Copy button when serviceType is user', () => {
    const wrapper = mount(CloneHost, {
      props: { service: makeService(), serviceType: 'user', navigate: vi.fn() },
    })
    expect(wrapper.find('[data-testid="copy-service-button"]').exists()).toBe(true)
  })

  it('does NOT render Copy button when serviceType is system', () => {
    const wrapper = mount(CloneHost, {
      props: { service: makeService({ type: 'system' }), serviceType: 'system', navigate: vi.fn() },
    })
    expect(wrapper.find('[data-testid="copy-service-button"]').exists()).toBe(false)
  })

  it('does NOT render Copy button when serviceType is apple-system', () => {
    const wrapper = mount(CloneHost, {
      props: {
        service: makeService({ type: 'apple-system' }),
        serviceType: 'apple-system',
        navigate: vi.fn(),
      },
    })
    expect(wrapper.find('[data-testid="copy-service-button"]').exists()).toBe(false)
  })

  it('Copy button is disabled when service is null (still loading)', () => {
    const wrapper = mount(CloneHost, {
      props: { service: null, serviceType: 'user', navigate: vi.fn() },
    })
    const btn = wrapper.find('[data-testid="copy-service-button"]')
    expect(btn.exists()).toBe(true)
    expect(btn.attributes('disabled')).toBeDefined()
  })
})

describe('Clone user service — Copy click opens modal with prefill', () => {
  let GetSettings: ReturnType<typeof vi.fn>
  let CreateService: ReturnType<typeof vi.fn>
  let currentSettings: Settings

  beforeEach(() => {
    currentSettings = { userLogDir: '~/Library/Logs', systemLogDir: '/Library/Logs' }
    GetSettings = vi.fn(async () => currentSettings)
    CreateService = vi.fn(async () => {})
    ;(window as unknown as { go: { main: { App: Record<string, unknown> } } }).go = {
      main: {
        App: {
          GetSettings,
          CreateService,
          UpdateSettings: vi.fn(),
          CreateSystemService: vi.fn(),
        },
      },
    }
    const { settings, defaults } = useSettings()
    settings.value = { ...defaults }
  })

  afterEach(() => {
    delete (window as unknown as { go?: unknown }).go
  })

  it('clicking Copy passes a ServiceConfig projection as prefill to the modal', async () => {
    const navigate = vi.fn()
    const service = makeService()
    const wrapper = mount(CloneHost, {
      props: { service, serviceType: 'user', navigate },
      global: { stubs: { ScheduleForm: true } },
    })

    await wrapper.find('[data-testid="copy-service-button"]').trigger('click')
    await flushPromises()

    const modal = wrapper.findComponent(CreateServiceModal)
    expect(modal.exists()).toBe(true)
    expect(modal.props('isOpen')).toBe(true)
    const prefill = modal.props('prefill') as ServiceConfig
    expect(prefill.label).toBe('com.example.copy-source')
    expect(prefill.program).toBe('/usr/bin/foo')
    expect(prefill.arguments).toEqual(['--port=8080'])
    expect(prefill.runAtLoad).toBe(true)
    expect(prefill.keepAlive).toEqual({ enabled: true, mode: 'boolean' })
    expect(prefill.environment).toEqual({ LOG_LEVEL: 'debug' })
    expect(prefill.schedule).toEqual({ interval: 60 })
    expect(prefill.workingDirectory).toBe('/Users/dev/project')
  })

  it('emitting "created" from modal calls navigate with the new label', async () => {
    const navigate = vi.fn()
    const wrapper = mount(CloneHost, {
      props: { service: makeService(), serviceType: 'user', navigate },
      global: { stubs: { ScheduleForm: true } },
    })

    await wrapper.find('[data-testid="copy-service-button"]').trigger('click')
    await flushPromises()

    const modal = wrapper.findComponent(CreateServiceModal)
    modal.vm.$emit('created', 'com.example.copy-dest')
    await flushPromises()

    expect(navigate).toHaveBeenCalledTimes(1)
    expect(navigate).toHaveBeenCalledWith('/services/com.example.copy-dest?type=user')
  })

  it('emitted label is URI-encoded when it contains reserved characters', async () => {
    const navigate = vi.fn()
    const wrapper = mount(CloneHost, {
      props: { service: makeService(), serviceType: 'user', navigate },
      global: { stubs: { ScheduleForm: true } },
    })

    await wrapper.find('[data-testid="copy-service-button"]').trigger('click')
    await flushPromises()

    const modal = wrapper.findComponent(CreateServiceModal)
    modal.vm.$emit('created', 'com.example/needs-encoding')
    await flushPromises()

    expect(navigate).toHaveBeenCalledWith('/services/com.example%2Fneeds-encoding?type=user')
  })

  it('modal close (without create) leaves caller without a navigate call', async () => {
    const navigate = vi.fn()
    const wrapper = mount(CloneHost, {
      props: { service: makeService(), serviceType: 'user', navigate },
      global: { stubs: { ScheduleForm: true } },
    })

    await wrapper.find('[data-testid="copy-service-button"]').trigger('click')
    await flushPromises()
    expect(wrapper.findComponent(CreateServiceModal).props('isOpen')).toBe(true)

    wrapper.findComponent(CreateServiceModal).vm.$emit('close')
    await flushPromises()

    expect(wrapper.findComponent(CreateServiceModal).props('isOpen')).toBe(false)
    expect(navigate).not.toHaveBeenCalled()
  })
})
