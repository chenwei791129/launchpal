import type { Settings } from '~/types/wails'

// composeLogPaths derives the stdout/stderr file paths the New Service modal
// previews and submits. The path is always `<dir>/<label>/<stream>.log` where
// <dir> comes from Settings (userLogDir or systemLogDir depending on the
// service type) and <label> is the user-typed service identifier.
//
// Trailing slashes on the directory are stripped so callers don't get a
// `//foo/stdout.log` when settings.json contains `~/Library/Logs/`.
export function composeLogPaths(
  serviceType: 'user' | 'system',
  settings: Settings,
  label: string,
): { stdout: string, stderr: string } {
  const dir = serviceType === 'system' ? settings.systemLogDir : settings.userLogDir
  const base = dir.replace(/\/+$/, '')
  return {
    stdout: `${base}/${label}/stdout.log`,
    stderr: `${base}/${label}/stderr.log`,
  }
}
