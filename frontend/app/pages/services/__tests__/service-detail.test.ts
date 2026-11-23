import { flushPromises, shallowMount } from '@vue/test-utils'
import { afterEach, expect, it, vi } from 'vitest'
import ServiceDetail from '../[name].vue'

const name = 'com.example.worker'
const service = {
  name,
  label: name,
  type: 'user',
  program: '/usr/bin/true',
  abandonProcessGroup: true,
  environment: { FIRST: 'example-a', SECOND: 'example-b' },
}

async function mountDetail(status = 'running') {
  const kickstart = vi.fn()
  const updateService = vi.fn().mockResolvedValue(undefined)
  const getService = vi.fn().mockResolvedValue({ ...service, status })
  vi.stubGlobal('useRoute', () => ({ params: { name }, query: {} }))
  vi.stubGlobal('go', { main: { App: { GetService: getService, KickstartService: kickstart, UpdateService: updateService } } })
  const wrapper = shallowMount(ServiceDetail, {
    global: {
      stubs: { teleport: true, NuxtLink: true, StatusConfidenceIcon: true, ServiceSummary: true, ServiceLogs: true, CreateServiceModal: true, ScheduleForm: false },
    },
  })
  await flushPromises()
  return { wrapper, kickstart, getService, updateService }
}

afterEach(() => vi.unstubAllGlobals())

it.each([true, false])('submits Abandon Process Group as %s from the edit form', async (enabled) => {
  const { wrapper, updateService } = await mountDetail()
  try {
    await wrapper.findAll('button').find(button => button.text() === 'Edit')!.trigger('click')
    const checkbox = wrapper.get<HTMLInputElement>('input[aria-describedby]')
    expect(checkbox.element.checked).toBe(true)
    await checkbox.setValue(enabled)
    await wrapper.findAll('button').find(button => button.text() === 'Save Changes')!.trigger('click')
    expect(updateService).toHaveBeenCalledExactlyOnceWith(name, expect.objectContaining({ abandonProcessGroup: enabled }))
    await flushPromises()
  } finally {
    wrapper.unmount()
  }
})

it('masks each existing environment value until that row is revealed', async () => {
  const { wrapper } = await mountDetail()
  try {
    await wrapper.findAll('button').find(button => button.text() === 'Edit')!.trigger('click')
    expect(wrapper.findAll('input[placeholder="Value"]').map(input => input.attributes('type'))).toEqual(['password', 'password'])
    await wrapper.get('button[title="Show value"]').trigger('click')
    expect(wrapper.findAll('input[placeholder="Value"]').map(input => input.attributes('type'))).toEqual(['text', 'password'])
  } finally {
    wrapper.unmount()
  }
})

it('does not kickstart a running service when Run Now is cancelled', async () => {
  const { wrapper, kickstart } = await mountDetail()
  try {
    await wrapper.findAll('button').find(button => button.text() === 'Run Now')!.trigger('click')
    expect(wrapper.text()).toContain('Kickstart will terminate the existing process')
    expect(kickstart).not.toHaveBeenCalled()
    await wrapper.findAll('button').find(button => button.text() === 'Cancel')!.trigger('click')
    expect(wrapper.text()).not.toContain('Kickstart will terminate the existing process')
    expect(kickstart).not.toHaveBeenCalled()
  } finally {
    wrapper.unmount()
  }
})

it.each(['running', 'loaded'])('runs and refreshes a %s service', async (status) => {
  const { wrapper, kickstart, getService } = await mountDetail(status)
  try {
    await wrapper.findAll('button').find(button => button.text() === 'Run Now')!.trigger('click')
    if (status === 'running') {
      expect(kickstart).not.toHaveBeenCalled()
      await wrapper.findAll('button').filter(button => button.text() === 'Run Now').at(-1)!.trigger('click')
    } else {
      expect(wrapper.text()).not.toContain('Kickstart will terminate the existing process')
    }
    await flushPromises()
    expect(kickstart).toHaveBeenCalledExactlyOnceWith(name)
    expect(getService).toHaveBeenCalledTimes(2)
  } finally {
    wrapper.unmount()
  }
})
