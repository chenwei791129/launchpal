import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, reactive } from 'vue'

// Minimal wrapper component that replicates the env var editing section
// from [name].vue, isolating just the behavior under test.
const EnvVarEditor = defineComponent({
  setup() {
    const envVars = reactive<Array<{ key: string; value: string }>>([])
    // Set tracking which indices have their value revealed (type="text")
    const visibility = reactive(new Set<number>())

    function toggleVisibility(index: number) {
      if (visibility.has(index)) {
        visibility.delete(index)
      } else {
        visibility.add(index)
      }
    }

    function removeEnvVar(index: number) {
      envVars.splice(index, 1)
      const next = new Set<number>()
      for (const i of visibility) {
        if (i < index) next.add(i)
        else if (i > index) next.add(i - 1)
      }
      visibility.clear()
      for (const i of next) visibility.add(i)
    }

    return { envVars, visibility, toggleVisibility, removeEnvVar }
  },
  template: `
    <div>
      <div v-for="(env, index) in envVars" :key="index" class="env-row">
        <input
          v-model="env.key"
          type="text"
          class="key-input"
          data-testid="key-input"
        />
        <input
          v-model="env.value"
          :type="visibility.has(index) ? 'text' : 'password'"
          class="value-input"
          data-testid="value-input"
        />
        <button
          type="button"
          @click="toggleVisibility(index)"
          data-testid="toggle-btn"
        >
          <svg v-if="!visibility.has(index)" data-testid="eye-icon" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          <svg v-else data-testid="eye-off-icon" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
          </svg>
        </button>
        <button
          type="button"
          @click="removeEnvVar(index)"
          data-testid="delete-btn"
        >
          ×
        </button>
      </div>
      <button type="button" @click="envVars.push({ key: '', value: '' })" data-testid="add-btn">
        + Add
      </button>
    </div>
  `,
})

describe('Edit tab — environment variable masking', () => {
  it('value input has type="password" by default', async () => {
    const wrapper = mount(EnvVarEditor)
    // Add an env var row
    await wrapper.find('[data-testid="add-btn"]').trigger('click')

    const valueInput = wrapper.find('[data-testid="value-input"]')
    expect(valueInput.attributes('type')).toBe('password')
  })

  it('clicking the eye button reveals the value (type becomes "text")', async () => {
    const wrapper = mount(EnvVarEditor)
    await wrapper.find('[data-testid="add-btn"]').trigger('click')

    const valueInput = wrapper.find('[data-testid="value-input"]')
    expect(valueInput.attributes('type')).toBe('password')

    await wrapper.find('[data-testid="toggle-btn"]').trigger('click')
    expect(valueInput.attributes('type')).toBe('text')
  })

  it('clicking the eye button again hides the value (type back to "password")', async () => {
    const wrapper = mount(EnvVarEditor)
    await wrapper.find('[data-testid="add-btn"]').trigger('click')

    const toggleBtn = wrapper.find('[data-testid="toggle-btn"]')
    // Reveal
    await toggleBtn.trigger('click')
    expect(wrapper.find('[data-testid="value-input"]').attributes('type')).toBe('text')
    // Hide again
    await toggleBtn.trigger('click')
    expect(wrapper.find('[data-testid="value-input"]').attributes('type')).toBe('password')
  })

  it('each row toggles independently', async () => {
    const wrapper = mount(EnvVarEditor)
    // Add two rows
    await wrapper.find('[data-testid="add-btn"]').trigger('click')
    await wrapper.find('[data-testid="add-btn"]').trigger('click')

    const valueInputs = wrapper.findAll('[data-testid="value-input"]')
    const toggleBtns = wrapper.findAll('[data-testid="toggle-btn"]')

    expect(valueInputs[0].attributes('type')).toBe('password')
    expect(valueInputs[1].attributes('type')).toBe('password')

    // Toggle only the first row
    await toggleBtns[0].trigger('click')
    expect(valueInputs[0].attributes('type')).toBe('text')
    expect(valueInputs[1].attributes('type')).toBe('password')

    // Toggle only the second row
    await toggleBtns[1].trigger('click')
    expect(valueInputs[0].attributes('type')).toBe('text')
    expect(valueInputs[1].attributes('type')).toBe('text')
  })

  it('eye icon is shown when value is hidden, eye-off icon when revealed', async () => {
    const wrapper = mount(EnvVarEditor)
    await wrapper.find('[data-testid="add-btn"]').trigger('click')

    // Hidden by default — should show eye icon
    expect(wrapper.find('[data-testid="eye-icon"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="eye-off-icon"]').exists()).toBe(false)

    // After revealing — should show eye-off icon
    await wrapper.find('[data-testid="toggle-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="eye-icon"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="eye-off-icon"]').exists()).toBe(true)
  })

  it('new rows added after toggle are hidden by default', async () => {
    const wrapper = mount(EnvVarEditor)
    // Add first row and reveal it
    await wrapper.find('[data-testid="add-btn"]').trigger('click')
    await wrapper.find('[data-testid="toggle-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="value-input"]').attributes('type')).toBe('text')

    // Add a second row — it should be hidden
    await wrapper.find('[data-testid="add-btn"]').trigger('click')
    const valueInputs = wrapper.findAll('[data-testid="value-input"]')
    expect(valueInputs[1].attributes('type')).toBe('password')
  })

  it('deleting a row shifts visibility indices correctly', async () => {
    const wrapper = mount(EnvVarEditor)
    // Add three rows
    const addBtn = wrapper.find('[data-testid="add-btn"]')
    await addBtn.trigger('click')
    await addBtn.trigger('click')
    await addBtn.trigger('click')

    // Reveal the second row (index 1)
    const toggleBtns = wrapper.findAll('[data-testid="toggle-btn"]')
    await toggleBtns[1].trigger('click')
    expect(wrapper.findAll('[data-testid="value-input"]')[1].attributes('type')).toBe('text')

    // Delete the first row (index 0) — the revealed row should shift from index 1 to 0
    await wrapper.findAll('[data-testid="delete-btn"]')[0].trigger('click')

    const valueInputs = wrapper.findAll('[data-testid="value-input"]')
    expect(valueInputs).toHaveLength(2)
    // The previously-revealed row (was index 1, now index 0) should still be revealed
    expect(valueInputs[0].attributes('type')).toBe('text')
    // The other row should remain masked
    expect(valueInputs[1].attributes('type')).toBe('password')
  })
})
