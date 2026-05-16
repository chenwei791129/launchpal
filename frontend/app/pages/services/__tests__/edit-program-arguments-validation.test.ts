import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, reactive, ref, computed, nextTick } from 'vue'
import { hasProgramOrArguments, PROGRAM_PATH_HINT } from '~/utils/serviceValidation'

// Mirrors the cross-field validation between `editForm.program` and
// `editArgumentsText` in pages/services/[name].vue. The wrapper isolates the
// logic so we can exercise it without mocking `useRoute` and the rest of the
// page.
const EditSaveGate = defineComponent({
  setup() {
    const editForm = reactive({ program: '' })
    const editArgumentsText = ref('')

    const canSave = computed(() =>
      hasProgramOrArguments(editForm.program, editArgumentsText.value),
    )

    return { editForm, editArgumentsText, canSave, hint: PROGRAM_PATH_HINT }
  },
  template: `
    <div>
      <input
        v-model="editForm.program"
        data-testid="program-input"
      />
      <p data-testid="program-hint">{{ hint }}</p>
      <input
        v-model="editArgumentsText"
        data-testid="arguments-input"
      />
      <button
        data-testid="save-btn"
        :disabled="!canSave"
      >Save Changes</button>
    </div>
  `,
})

describe('Edit Tab — Program OR Arguments cross-field validation', () => {
  it('disables Save Changes when both Program and Arguments are empty', () => {
    const wrapper = mount(EditSaveGate)
    const btn = wrapper.find('[data-testid="save-btn"]')
    expect(btn.attributes('disabled')).toBeDefined()
  })

  it('enables Save Changes when only Program is filled', async () => {
    const wrapper = mount(EditSaveGate)
    await wrapper.find('[data-testid="program-input"]').setValue('/usr/bin/true')
    await nextTick()
    const btn = wrapper.find('[data-testid="save-btn"]')
    expect(btn.attributes('disabled')).toBeUndefined()
  })

  it('enables Save Changes when only Arguments is filled', async () => {
    const wrapper = mount(EditSaveGate)
    await wrapper
      .find('[data-testid="arguments-input"]')
      .setValue("/usr/bin/open '/Applications/Foo.app'")
    await nextTick()
    const btn = wrapper.find('[data-testid="save-btn"]')
    expect(btn.attributes('disabled')).toBeUndefined()
  })

  it('disables Save Changes when Arguments contains only whitespace and Program is empty', async () => {
    const wrapper = mount(EditSaveGate)
    await wrapper.find('[data-testid="arguments-input"]').setValue('   ')
    await nextTick()
    const btn = wrapper.find('[data-testid="save-btn"]')
    expect(btn.attributes('disabled')).toBeDefined()
  })

  it('renders the shared Program Path hint constant', () => {
    const wrapper = mount(EditSaveGate)
    expect(wrapper.find('[data-testid="program-hint"]').text()).toBe(PROGRAM_PATH_HINT)
  })
})
