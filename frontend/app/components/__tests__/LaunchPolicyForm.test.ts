import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import LaunchPolicyForm from '../LaunchPolicyForm.vue'
import { emptyKeepAlive, type LaunchPolicy } from '~/utils/launchPolicy'
import type { KeepAliveConfig } from '~/types/wails'

// Host wires v-model the way the modal/edit page do, so we can read back the
// committed state after interacting with the controls.
function mountForm(initial?: {
  launchPolicy?: LaunchPolicy
  keepAlive?: KeepAliveConfig
  throttleInterval?: number
}) {
  const launchPolicy = ref<LaunchPolicy>(initial?.launchPolicy ?? 'runAtLoad')
  const keepAlive = ref<KeepAliveConfig>(initial?.keepAlive ?? emptyKeepAlive())
  const throttleInterval = ref<number | undefined>(initial?.throttleInterval)

  const Host = defineComponent({
    setup() {
      return () => h(LaunchPolicyForm, {
        'launchPolicy': launchPolicy.value,
        'onUpdate:launchPolicy': (v: LaunchPolicy) => { launchPolicy.value = v },
        'keepAlive': keepAlive.value,
        'onUpdate:keepAlive': (v: KeepAliveConfig) => { keepAlive.value = v },
        'throttleInterval': throttleInterval.value,
        'onUpdate:throttleInterval': (v: number | undefined) => { throttleInterval.value = v },
      })
    },
  })

  const wrapper = mount(Host)
  return { wrapper, launchPolicy, keepAlive, throttleInterval }
}

describe('LaunchPolicyForm — radio group', () => {
  it('renders exactly three launch-policy options', () => {
    const { wrapper } = mountForm()
    expect(wrapper.find('[data-testid="launch-policy-onDemand"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="launch-policy-runAtLoad"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="launch-policy-keepAlive"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('On Demand')
    expect(wrapper.text()).toContain('Run at Load')
    expect(wrapper.text()).toContain('Keep Alive')
  })

  it('hides the advanced section unless Keep Alive is selected', async () => {
    const { wrapper } = mountForm({ launchPolicy: 'runAtLoad' })
    expect(wrapper.find('[data-testid="keepalive-advanced"]').exists()).toBe(false)

    await wrapper.find('[data-testid="launch-policy-keepAlive"]').setValue(true)
    expect(wrapper.find('[data-testid="keepalive-advanced"]').exists()).toBe(true)
  })
})

describe('LaunchPolicyForm — advanced KeepAlive', () => {
  it('editing a sub-key produces a dictionary KeepAlive', async () => {
    const { wrapper, keepAlive } = mountForm({ launchPolicy: 'keepAlive' })

    await wrapper.find('[data-testid="keepalive-mode-dictionary"]').setValue(true)
    const select = wrapper.find('[data-testid="keepalive-successfulExit"]')
    await select.setValue('false')

    expect(keepAlive.value.mode).toBe('dictionary')
    expect(keepAlive.value.successfulExit).toBe(false)
  })

  it('edits ThrottleInterval into the model', async () => {
    const { wrapper, throttleInterval } = mountForm({ launchPolicy: 'keepAlive' })
    await wrapper.find('[data-testid="keepalive-throttleInterval"]').setValue('15')
    expect(throttleInterval.value).toBe(15)
  })

  it('shows OR-semantics and implies-Run-at-Load helper text', () => {
    const { wrapper } = mountForm({ launchPolicy: 'keepAlive' })
    const text = wrapper.text()
    expect(text).toContain('OR')
    expect(text).toContain('implies Run at Load')
  })
})
