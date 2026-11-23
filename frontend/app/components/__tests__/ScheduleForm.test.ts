import { mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import ScheduleForm from '../ScheduleForm.vue'

it('expands ranges and lists together, but rejects more than 50 entries', async () => {
  const wrapper = mount(ScheduleForm, {
    props: { modelValue: { schedules: [{ minute: 0, hour: 9 }] } },
  })
  try {
    const input = wrapper.get('input[placeholder="* * * * *"]')
    await input.setValue('0,30 9-10 * * *')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([{
      schedules: [
        { minute: 0, hour: 9 },
        { minute: 0, hour: 10 },
        { minute: 30, hour: 9 },
        { minute: 30, hour: 10 },
      ],
    }])
    expect(wrapper.text()).toContain('4 schedules:')
    expect(wrapper.text()).toContain('Next runs (')
    expect(wrapper.text()).toMatch(/\d{1,2}\/\d{1,2} \([A-Za-z]{3}\) (09|10):(00|30)/)

    const emitted = wrapper.emitted('update:modelValue')?.slice()
    await input.setValue('0-25 9-10 * * *')
    expect(wrapper.text()).toContain('52 entries, exceeding limit of 50')
    expect(wrapper.text()).not.toContain('Next runs (')
    expect(wrapper.emitted('update:modelValue')).toEqual(emitted)
  } finally {
    wrapper.unmount()
  }
})
