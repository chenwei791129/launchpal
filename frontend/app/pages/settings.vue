<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="px-4 py-3 border-b border-surface-100">
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 text-sm mb-3">
        <NuxtLink to="/" class="text-gray-400 hover:text-white transition-colors">
          Services
        </NuxtLink>
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-white">Settings</span>
      </div>

      <!-- Title -->
      <div class="flex items-center gap-3">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <h1 class="text-lg font-semibold text-white">Settings</h1>
      </div>
    </header>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 space-y-6">
      <!-- Admin Mode Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">Admin Mode</h2>
        <div class="flex items-start gap-3 mb-4">
          <div class="w-10 h-10 bg-surface-200 rounded-lg flex items-center justify-center flex-shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" :class="adminBadgeColor" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <p class="text-white font-medium">{{ adminStateLabel }}</p>
              <span class="px-2 py-0.5 rounded text-xs" :class="adminBadgeClass">
                {{ adminStateLabel }}
              </span>
            </div>
            <p class="text-gray-400 text-sm mb-2">
              Enable Admin Mode to manage system LaunchDaemons (under /Library/LaunchDaemons). You will be prompted for your password once per session; no privileged process remains after LaunchPal exits.
            </p>
            <p
              v-if="admin.lastError.value"
              class="text-sm mb-2"
              :class="admin.isSessionEnded.value ? 'text-gray-400' : 'text-red-400'"
            >
              {{ admin.displayMessage.value }}
            </p>
            <div class="flex items-center gap-2">
              <button
                v-if="!admin.isEnabled.value && !admin.isShuttingDown.value"
                :disabled="admin.loading.value || admin.isRequesting.value"
                class="px-3 py-1.5 text-sm bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 text-white rounded-lg transition-colors"
                @click="admin.enable()"
              >
                {{ admin.isRequesting.value ? 'Requesting...' : 'Enable Admin Mode' }}
              </button>
              <button
                v-else-if="admin.isEnabled.value"
                :disabled="admin.loading.value"
                class="px-3 py-1.5 text-sm bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded-lg transition-colors"
                @click="admin.disable()"
              >
                Disable Admin Mode
              </button>
              <span
                v-else-if="admin.isShuttingDown.value"
                class="text-sm text-gray-400"
              >
                Shutting down helper...
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- Version Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">Version</h2>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-primary-600/20 rounded-lg flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <p class="text-white font-medium">LaunchPal</p>
              <p class="text-gray-400 text-sm">Version {{ appVersion }}</p>
            </div>
          </div>
          <span class="px-3 py-1 bg-primary-600/20 text-primary-400 text-sm rounded-full">
            {{ appVersion }}
          </span>
        </div>
      </section>

      <!-- Backup Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">Backup Storage</h2>
        <div class="flex items-start gap-3">
          <div class="w-10 h-10 bg-surface-200 rounded-lg flex items-center justify-center flex-shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-white font-medium mb-1">Backup Directory</p>
            <p class="text-gray-400 text-sm mb-2">Service backups are stored at:</p>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-surface-500 rounded-lg text-sm text-gray-300 font-mono truncate">
                {{ backupPath }}
              </code>
              <button
                class="p-2 rounded-lg hover:bg-surface-200 text-gray-400 hover:text-white transition-colors flex-shrink-0"
                title="Copy path"
                @click="copyBackupPath"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Backup List -->
        <div class="mt-6 border-t border-surface-100 pt-4">
          <div class="flex items-center justify-between mb-4">
            <p class="text-white font-medium">Backup History</p>
            <button
              class="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-white transition-colors"
              title="Refresh"
              @click="loadBackups"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          </div>

          <div v-if="loadingBackups" class="text-center py-4 text-gray-400">
            Loading backups...
          </div>

          <div v-else-if="backups.length === 0" class="text-center py-4 text-gray-500">
            No backups found
          </div>

          <div v-else class="space-y-2 max-h-64 overflow-y-auto">
            <div
              v-for="backup in backups"
              :key="`${backup.service}-${backup.id}`"
              class="flex items-center justify-between p-3 bg-surface-500 rounded-lg"
            >
              <div class="min-w-0 flex-1">
                <p class="text-white text-sm font-medium truncate">{{ backup.service }}</p>
                <p class="text-gray-500 text-xs">{{ formatTimestamp(backup.timestamp) }}</p>
              </div>
              <div class="ml-3 flex items-center gap-2">
                <span class="relative group" data-testid="diff-tooltip-wrapper">
                  <button
                    class="p-1.5 rounded hover:bg-surface-200 text-gray-400 hover:text-white transition-colors"
                    title="View diff"
                    @click="openDiff(backup)"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h8M8 12h8M8 17h5M3 4h4v16H3V4zm14 0h4v16h-4V4z" />
                    </svg>
                  </button>
                  <span
                    data-testid="diff-tooltip"
                    class="pointer-events-none absolute top-full right-0 mt-1 px-2 py-1 text-xs whitespace-nowrap bg-surface-600 text-white rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity duration-150 z-10"
                  >
                    Preview diff against current plist
                  </span>
                </span>
                <button
                  class="px-3 py-1.5 text-xs bg-primary-600 hover:bg-primary-700 text-white rounded transition-colors"
                  @click="confirmRestore(backup)"
                >
                  Restore
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Log Storage Section -->
      <LogStorageSection />

      <!-- Diff Preview Dialog -->
      <BackupDiffDialog
        :visible="showDiffDialog"
        :backup="backupToDiff"
        @close="closeDiff"
        @restore="onDiffRestore"
      />

      <!-- Restore Confirmation Dialog -->
      <Teleport to="body">
        <div
          v-if="showRestoreDialog"
          class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
          @click.self="closeRestoreDialog"
        >
          <div class="bg-surface-400 rounded-xl shadow-xl p-6 w-96">
            <h3 class="text-lg font-semibold text-white mb-2">Restore Backup</h3>
            <p class="text-gray-400 mb-4">
              Are you sure you want to restore this backup?
            </p>
            <div class="bg-surface-500 rounded-lg p-3 mb-4">
              <p class="text-white text-sm font-medium">{{ backupToRestore?.service }}</p>
              <p class="text-gray-500 text-xs">{{ backupToRestore ? formatTimestamp(backupToRestore.timestamp) : '' }}</p>
            </div>
            <p class="text-yellow-500 text-sm mb-4">
              This will overwrite the current plist file.
            </p>
            <div
              v-if="restoreBlockedByAdminMode"
              class="bg-yellow-500/10 border border-yellow-500/20 text-yellow-400 text-sm rounded-lg p-3 mb-4"
            >
              This backup targets <code class="text-yellow-300">/Library/LaunchDaemons/</code>.
              Enable Admin Mode above before restoring.
            </div>
            <p v-if="restoreError" class="text-red-400 text-sm mb-4">{{ restoreError }}</p>
            <div class="flex justify-end gap-3">
              <button
                class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
                :disabled="restoringBackup"
                @click="closeRestoreDialog"
              >
                Cancel
              </button>
              <button
                class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                :disabled="restoringBackup || restoreBlockedByAdminMode"
                @click="executeRestore"
              >
                {{ restoringBackup ? 'Restoring...' : 'Restore' }}
              </button>
            </div>
          </div>
        </div>
      </Teleport>

      <!-- About Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">About</h2>
        <div class="space-y-4">
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 bg-gradient-to-br from-primary-500 to-primary-700 rounded-2xl flex items-center justify-center shadow-lg">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h3 class="text-xl font-semibold text-white">LaunchPal</h3>
              <p class="text-gray-400">A GUI for managing macOS LaunchAgents</p>
            </div>
          </div>

          <div class="border-t border-surface-100 pt-4">
            <p class="text-gray-400 text-sm leading-relaxed">
              LaunchPal is a LaunchAgent management tool built for macOS, offering an intuitive
              graphical interface to manage system services. It supports starting, stopping, and
              restarting services, as well as inspecting and editing plist files.
            </p>
          </div>

          <div class="border-t border-surface-100 pt-4">
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p class="text-gray-500 mb-1">Platform</p>
                <p class="text-white">macOS</p>
              </div>
              <div>
                <p class="text-gray-500 mb-1">Framework</p>
                <p class="text-white">Wails + Vue</p>
              </div>
              <div>
                <p class="text-gray-500 mb-1">License</p>
                <p class="text-white">MIT</p>
              </div>
              <div>
                <p class="text-gray-500 mb-1">Language</p>
                <p class="text-white">Go + TypeScript</p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAppVersion } from '~/composables/useAppVersion'
import { useAdminMode } from '~/composables/useAdminMode'
import BackupDiffDialog from '~/components/BackupDiffDialog.vue'
import LogStorageSection from '~/components/LogStorageSection.vue'
import { formatTimestamp } from '~/utils/formatters'
import type { AdminModeState, Backup } from '~/types/wails'

const appVersion = useAppVersion()
const admin = useAdminMode()

// Single lookup keyed by AdminModeState. Each state maps to its label,
// badge classes, and icon color so the template can bind directly without
// triplicate switches.
const adminStateStyles: Record<AdminModeState, { label: string, badge: string, icon: string }> = {
  enabled:       { label: 'Enabled',       badge: 'bg-green-600/20 text-green-400',   icon: 'text-green-400' },
  requesting:    { label: 'Requesting',    badge: 'bg-yellow-600/20 text-yellow-400', icon: 'text-yellow-400' },
  shutting_down: { label: 'Shutting down', badge: 'bg-orange-600/20 text-orange-400', icon: 'text-orange-400' },
  disabled:      { label: 'Disabled',      badge: 'bg-gray-600/20 text-gray-400',     icon: 'text-gray-400' },
}
const adminStyle = computed(() => adminStateStyles[admin.state.value])
const adminStateLabel = computed(() => adminStyle.value.label)
const adminBadgeClass = computed(() => adminStyle.value.badge)
const adminBadgeColor = computed(() => adminStyle.value.icon)
const backupPath = '~/.launchpal/backups/'

const backups = ref<Backup[]>([])
const loadingBackups = ref(false)
const showRestoreDialog = ref(false)
const backupToRestore = ref<Backup | null>(null)
const restoringBackup = ref(false)
const restoreError = ref<string | null>(null)
const showDiffDialog = ref(false)
const backupToDiff = ref<Backup | null>(null)

// System-domain restores must go through the privileged helper, so we block
// the action client-side when Admin Mode is off and surface a clear error
// instead of letting the backend return "read-only manager" in an alert that
// would leave the dialog stuck open.
function isSystemBackup(backup: Backup | null): boolean {
  return !!backup?.originalPath?.startsWith('/Library/LaunchDaemons/')
}
const restoreBlockedByAdminMode = computed(
  () => isSystemBackup(backupToRestore.value) && !admin.isEnabled.value,
)

async function copyBackupPath() {
  try {
    await navigator.clipboard.writeText(backupPath)
  } catch (e) {
    console.error('Failed to copy to clipboard:', e)
  }
}

async function loadBackups() {
  loadingBackups.value = true
  try {
    if (window.go?.main?.App?.ListAllBackups) {
      backups.value = await window.go.main.App.ListAllBackups()
    }
  } catch (e) {
    console.error('Failed to load backups:', e)
  } finally {
    loadingBackups.value = false
  }
}

function confirmRestore(backup: Backup) {
  backupToRestore.value = backup
  restoreError.value = null
  showRestoreDialog.value = true
}

function closeRestoreDialog() {
  showRestoreDialog.value = false
  backupToRestore.value = null
  restoreError.value = null
  restoringBackup.value = false
}

function openDiff(backup: Backup) {
  backupToDiff.value = backup
  showDiffDialog.value = true
}

function closeDiff() {
  showDiffDialog.value = false
  backupToDiff.value = null
}

function onDiffRestore(backup: Backup) {
  closeDiff()
  confirmRestore(backup)
}

async function executeRestore() {
  if (!backupToRestore.value) return
  if (restoreBlockedByAdminMode.value) {
    // Fail fast — the backend would return ErrReadOnlyManager, but surfacing
    // that via alert() leaves the dialog stuck open. Show it inline instead.
    restoreError.value =
      'This is a system service backup. Enable Admin Mode in Settings before restoring.'
    return
  }

  restoreError.value = null
  restoringBackup.value = true
  try {
    if (!window.go?.main?.App?.RestoreBackup) {
      throw new Error('RestoreBackup binding is not available')
    }
    await window.go.main.App.RestoreBackup(
      backupToRestore.value.service,
      backupToRestore.value.id,
    )
    closeRestoreDialog()
    await loadBackups()
  } catch (e) {
    console.error('Failed to restore backup:', e)
    restoreError.value = e instanceof Error ? e.message : String(e)
  } finally {
    restoringBackup.value = false
  }
}

onMounted(() => {
  loadBackups()
})
</script>
