import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive } from 'vue'
import { defineComponent, nextTick } from 'vue'


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
