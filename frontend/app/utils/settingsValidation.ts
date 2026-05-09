// Mirrors internal/settings.Validate so the UI surfaces validation errors
// without a Wails round-trip. Backend remains source of truth: anything
// the client misses is still rejected before the file is written.
//
// Keep this list in sync with privhelper.SystemLogPathPrefixes.
export const SYSTEM_LOG_PATH_PREFIXES = [
  '/var/log/',
  '/private/var/log/',
  '/Library/Logs/',
  '/tmp/',
  '/private/tmp/',
] as const

const SHELL_META = [';', '&', '|', '$', '`', '\n', '\r', '\x00']

function containsShellMeta(s: string): boolean {
  return SHELL_META.some(c => s.includes(c))
}

export function validateUserLogDir(v: string): string | null {
  if (v === '') return 'userLogDir: must not be empty'
  if (containsShellMeta(v)) return 'userLogDir: contains shell metacharacter'
  if (!v.startsWith('~/') && !v.startsWith('/')) {
    return 'userLogDir: must be tilde-home (~/...) or absolute (/...)'
  }
  return null
}

// Trims trailing "/" off the input and reports the first violated rule.
// Mirrors the Go validator: prefix-in-allowlist (allowing the bare root)
// is sufficient — the New Service modal interpolates a label that adds depth.
export function validateSystemLogDir(v: string): string | null {
  if (v === '') return 'systemLogDir: must not be empty'
  if (containsShellMeta(v)) return 'systemLogDir: contains shell metacharacter'
  if (!v.startsWith('/')) {
    return 'systemLogDir: must be absolute (no tilde, no relative paths)'
  }
  // Normalize collapsing "//" and trailing slashes to mimic filepath.Clean
  // for the prefix check. We don't fully reproduce Clean — that's the
  // backend's job — but we do enough so the common "/Library/Logs" works.
  let clean = v.replace(/\/+/g, '/')
  if (clean.length > 1 && clean.endsWith('/')) clean = clean.slice(0, -1)
  const cleanWithSlash = clean + '/'
  for (const prefix of SYSTEM_LOG_PATH_PREFIXES) {
    if (cleanWithSlash.startsWith(prefix)) return null
  }
  return `systemLogDir: must start with one of: ${SYSTEM_LOG_PATH_PREFIXES.join(', ')}`
}
