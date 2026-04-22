<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      @click.self="close"
    >
      <div class="bg-surface-400 rounded-xl shadow-xl w-full max-w-5xl max-h-[85vh] flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-surface-100 flex-shrink-0">
          <h3 class="text-lg font-semibold text-white">Restore Preview</h3>
          <div class="mt-1 text-sm text-gray-400">
            <span class="text-white font-medium">{{ backup?.service ?? '' }}</span>
            <span v-if="backup" class="mx-2">·</span>
            <span>{{ backup ? formatTimestamp(backup.timestamp) : '' }}</span>
          </div>

          <div v-if="showNoCurrent" class="mt-3 text-xs px-3 py-2 bg-primary-600/15 text-primary-300 rounded">
            No current version exists — restoring will create the file anew.
          </div>
          <div v-if="formatWarning" class="mt-3 text-xs px-3 py-2 bg-yellow-500/15 text-yellow-300 rounded">
            Format conversion failed for {{ formatWarning }} — diff output is likely unreadable.
          </div>
          <div v-if="diff?.truncated" class="mt-3 text-xs px-3 py-2 bg-yellow-500/15 text-yellow-300 rounded">
            Diff truncated to {{ MAX_DIFF_ROWS }} rows; {{ diff.omittedRows }} additional row(s) omitted.
          </div>
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-auto px-6 py-4">
          <div v-if="loading" class="text-center text-gray-400 py-8">Loading diff…</div>

          <div v-else-if="diff && !diff.hasChanges" class="text-center text-gray-400 py-8">
            No changes — this backup is identical to the current plist.
          </div>

          <div v-else-if="diff" class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-0 font-mono text-xs">
            <div
              data-testid="diff-column"
              class="border-r border-surface-100 overflow-x-auto min-w-0"
            >
              <div class="px-3 py-1 text-[10px] uppercase tracking-wider text-gray-500 border-b border-surface-100 sticky left-0 bg-surface-400">
                Current
              </div>
              <div
                v-for="(row, idx) in diff.left"
                :key="`l-${idx}`"
                class="flex min-h-[1.25rem] w-max min-w-full"
                :class="rowClass(row.type)"
              >
                <span class="w-10 text-right pr-2 flex-shrink-0 select-none text-gray-600">
                  {{ row.lineNumber ?? '' }}
                </span>
                <pre class="whitespace-pre pr-3">{{ row.text }}</pre>
              </div>
            </div>
            <div
              data-testid="diff-column"
              class="overflow-x-auto min-w-0"
            >
              <div class="px-3 py-1 text-[10px] uppercase tracking-wider text-gray-500 border-b border-surface-100 sticky left-0 bg-surface-400">
                Backup
              </div>
              <div
                v-for="(row, idx) in diff.right"
                :key="`r-${idx}`"
                class="flex min-h-[1.25rem] w-max min-w-full"
                :class="rowClass(row.type)"
              >
                <span class="w-10 text-right pr-2 flex-shrink-0 select-none text-gray-600">
                  {{ row.lineNumber ?? '' }}
                </span>
                <pre class="whitespace-pre pr-3">{{ row.text }}</pre>
              </div>
            </div>
          </div>

          <div v-else-if="error" class="text-center text-red-400 py-8">
            {{ error }}
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-surface-100 flex justify-end gap-3 flex-shrink-0">
          <button
            data-testid="diff-cancel"
            class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
            @click="close"
          >
            Cancel
          </button>
          <button
            data-testid="diff-restore"
            class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors"
            @click="emitRestore"
          >
            Restore
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computeSideBySideDiff, MAX_DIFF_ROWS, type DiffRowType } from '~/composables/useBackupDiff'
import { formatTimestamp } from '~/utils/formatters'
import type { Backup, PlistContent } from '~/types/wails'

const props = defineProps<{
  visible: boolean
  backup: Backup | null
}>()

const emit = defineEmits<{
  close: []
  restore: [backup: Backup]
}>()

const loading = ref(false)
const error = ref<string | null>(null)
const currentContent = ref<PlistContent | null>(null)
const backupContent = ref<PlistContent | null>(null)

const diff = computed(() => {
  if (!currentContent.value || !backupContent.value) return null
  return computeSideBySideDiff(currentContent.value.data, backupContent.value.data)
})

const showNoCurrent = computed(() => currentContent.value?.data === '')

const formatWarning = computed(() => {
  const sides: string[] = []
  if (currentContent.value?.convertFailed) sides.push('current plist')
  if (backupContent.value?.convertFailed) sides.push('backup plist')
  return sides.length > 0 ? sides.join(' and ') : null
})

// System-domain backups (created by the privhelper during Admin-Mode
// Update/Delete) target /Library/LaunchDaemons/. UserManager's
// GetCurrentPlist can't read those, so it would always return an empty
// Content and the diff would show the backup as pure additions. Detect
// the domain from originalPath and pick the matching binding.
function isSystemBackup(originalPath: string | undefined): boolean {
  if (!originalPath) return false
  return originalPath.startsWith('/Library/LaunchDaemons/')
}

async function loadDiff() {
  if (!props.backup) return
  loading.value = true
  error.value = null
  try {
    const currentFn = isSystemBackup(props.backup.originalPath)
      ? window.go.main.App.GetCurrentSystemPlist
      : window.go.main.App.GetCurrentPlist
    const [current, backup] = await Promise.all([
      currentFn(props.backup.service),
      window.go.main.App.GetBackupContent(props.backup.service, props.backup.id),
    ])
    currentContent.value = current
    backupContent.value = backup
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load diff'
    currentContent.value = null
    backupContent.value = null
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.visible, props.backup?.id, props.backup?.service] as const,
  ([visible]) => {
    if (visible && props.backup) {
      loadDiff()
    } else {
      currentContent.value = null
      backupContent.value = null
      error.value = null
      loading.value = false
    }
  },
  { immediate: true },
)

function rowClass(type: DiffRowType): string {
  switch (type) {
    case 'added':
      return 'bg-green-500/10 text-green-300'
    case 'removed':
      return 'bg-red-500/10 text-red-300'
    case 'placeholder':
      return 'bg-surface-600/40'
    default:
      return 'text-gray-300'
  }
}

function close() {
  emit('close')
}

function emitRestore() {
  if (props.backup) emit('restore', props.backup)
}
</script>
