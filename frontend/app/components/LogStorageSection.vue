<template>
  <section class="bg-surface-400 rounded-xl p-4">
    <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">Log Storage</h2>

    <div class="flex items-start gap-3 mb-6">
      <div class="w-10 h-10 bg-surface-200 rounded-lg flex items-center justify-center flex-shrink-0">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-6a2 2 0 012-2h2a2 2 0 012 2v6m-7 4h12a2 2 0 002-2V7a2 2 0 00-2-2h-3l-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
        </svg>
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-white font-medium mb-1">User Log Directory</p>
        <p class="text-gray-400 text-sm mb-2">
          Default stdout/stderr location for new user services. Each service gets a sub-directory under this path.
        </p>
        <div class="flex items-center gap-2">
          <input
            v-model="userInput"
            data-testid="user-log-dir-input"
            type="text"
            spellcheck="false"
            class="flex-1 px-3 py-2 bg-surface-500 rounded-lg text-sm text-gray-100 font-mono outline-none focus:ring-2 focus:ring-primary-500"
          >
          <button
            data-testid="user-log-dir-save"
            type="button"
            :disabled="userSaving"
            class="px-3 py-1.5 text-sm bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            @click="onSaveUser"
          >
            {{ userSaving ? 'Saving...' : 'Save' }}
          </button>
          <button
            data-testid="user-log-dir-reset"
            type="button"
            class="px-3 py-1.5 text-sm bg-surface-200 hover:bg-surface-100 text-gray-300 rounded-lg transition-colors"
            @click="onResetUser"
          >
            Reset to Default
          </button>
        </div>
        <p
          v-if="userError"
          data-testid="user-log-dir-error"
          class="text-red-400 text-sm mt-2"
        >
          {{ userError }}
        </p>
        <p
          v-if="userSavedFlash"
          data-testid="user-log-dir-success"
          class="text-green-400 text-sm mt-2"
        >
          Saved.
        </p>
      </div>
    </div>

    <div class="flex items-start gap-3">
      <div class="w-10 h-10 bg-surface-200 rounded-lg flex items-center justify-center flex-shrink-0">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7" />
        </svg>
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-white font-medium mb-1">System Log Directory</p>
        <p class="text-gray-400 text-sm mb-2">
          Default stdout/stderr location for new system daemons. Must live under one of:
          <code class="font-mono text-gray-300">/var/log/</code>,
          <code class="font-mono text-gray-300">/private/var/log/</code>,
          <code class="font-mono text-gray-300">/Library/Logs/</code>,
          <code class="font-mono text-gray-300">/tmp/</code>,
          <code class="font-mono text-gray-300">/private/tmp/</code>.
        </p>
        <div class="flex items-center gap-2">
          <input
            v-model="systemInput"
            data-testid="system-log-dir-input"
            type="text"
            spellcheck="false"
            class="flex-1 px-3 py-2 bg-surface-500 rounded-lg text-sm text-gray-100 font-mono outline-none focus:ring-2 focus:ring-primary-500"
          >
          <button
            data-testid="system-log-dir-save"
            type="button"
            :disabled="systemSaving"
            class="px-3 py-1.5 text-sm bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            @click="onSaveSystem"
          >
            {{ systemSaving ? 'Saving...' : 'Save' }}
          </button>
          <button
            data-testid="system-log-dir-reset"
            type="button"
            class="px-3 py-1.5 text-sm bg-surface-200 hover:bg-surface-100 text-gray-300 rounded-lg transition-colors"
            @click="onResetSystem"
          >
            Reset to Default
          </button>
        </div>
        <p
          v-if="systemError"
          data-testid="system-log-dir-error"
          class="text-red-400 text-sm mt-2"
        >
          {{ systemError }}
        </p>
        <p
          v-if="systemSavedFlash"
          data-testid="system-log-dir-success"
          class="text-green-400 text-sm mt-2"
        >
          Saved.
        </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettings } from '~/composables/useSettings'
import { validateUserLogDir, validateSystemLogDir } from '~/utils/settingsValidation'

const { settings, load, save, defaults } = useSettings()

const userInput = ref(settings.value.userLogDir || defaults.userLogDir)
const systemInput = ref(settings.value.systemLogDir || defaults.systemLogDir)

const userError = ref<string | null>(null)
const systemError = ref<string | null>(null)
const userSaving = ref(false)
const systemSaving = ref(false)
const userSavedFlash = ref(false)
const systemSavedFlash = ref(false)

onMounted(async () => {
  // Initial mount: pull from disk, then mirror into the input refs so the
  // user sees the persisted values. After this point the inputs are local
  // edit state — subsequent loads update settings.value (used as the
  // partner-field source on Save) without clobbering in-progress edits.
  await load()
  userInput.value = settings.value.userLogDir
  systemInput.value = settings.value.systemLogDir
})

function flashSuccess(target: 'user' | 'system') {
  if (target === 'user') {
    userSavedFlash.value = true
    setTimeout(() => {
      userSavedFlash.value = false
    }, 2000)
  } else {
    systemSavedFlash.value = true
    setTimeout(() => {
      systemSavedFlash.value = false
    }, 2000)
  }
}

async function onSaveUser() {
  userError.value = null
  const localErr = validateUserLogDir(userInput.value)
  if (localErr) {
    userError.value = localErr
    return
  }
  userSaving.value = true
  try {
    // Each Save persists only its own field — the partner field reuses the
    // last-saved value, not whatever the user has currently typed but hasn't
    // committed. Without this, an invalid in-progress edit in the System
    // input would surface its validation error under the User Save button.
    await save({ userLogDir: userInput.value, systemLogDir: settings.value.systemLogDir })
    await load()
    flashSuccess('user')
  } catch (e) {
    userError.value = e instanceof Error ? e.message : String(e)
  } finally {
    userSaving.value = false
  }
}

async function onSaveSystem() {
  systemError.value = null
  const localErr = validateSystemLogDir(systemInput.value)
  if (localErr) {
    systemError.value = localErr
    return
  }
  systemSaving.value = true
  try {
    await save({ userLogDir: settings.value.userLogDir, systemLogDir: systemInput.value })
    await load()
    flashSuccess('system')
  } catch (e) {
    systemError.value = e instanceof Error ? e.message : String(e)
  } finally {
    systemSaving.value = false
  }
}

function onResetUser() {
  userInput.value = defaults.userLogDir
  userError.value = null
}

function onResetSystem() {
  systemInput.value = defaults.systemLogDir
  systemError.value = null
}
</script>
