import type { CalendarEntry, ScheduleConfig } from '~/types/wails.d'

const MAX_SCAN_DAYS = 400

/**
 * Calculate the next N occurrences for a CalendarInterval schedule.
 *
 * Fields in each schedule entry follow launchd semantics:
 * - undefined = wildcard (matches any value)
 * - month: 1-12
 * - day: 1-31
 * - weekday: 0-6 (0=Sunday)
 * - hour: 0-23
 * - minute: 0-59
 *
 * Scans minute-by-minute starting from now+1 minute, up to MAX_SCAN_DAYS.
 */
export function getNextOccurrences(
  config: ScheduleConfig,
  count: number,
  now: Date = new Date(),
): Date[] {
  // StartInterval configs have no calendar fields — nothing to scan
  if (config.interval !== undefined) return []

  const schedules = config.schedules ?? []
  if (schedules.length === 0) return []

  const results: Date[] = []
  const maxMinutes = MAX_SCAN_DAYS * 24 * 60

  // Start from now + 1 minute, with seconds/ms zeroed
  const cursor = new Date(now)
  cursor.setSeconds(0, 0)
  cursor.setMinutes(cursor.getMinutes() + 1)

  for (let i = 0; i < maxMinutes && results.length < count; i++) {
    if (schedules.some(entry => matchesEntry(cursor, entry))) {
      results.push(new Date(cursor))
    }
    cursor.setMinutes(cursor.getMinutes() + 1)
  }

  return results
}

export const WEEKDAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as const

/**
 * Format a Date as "M/D (Weekday) HH:mm".
 * Weekday uses English abbreviations (Sun, Mon, Tue, Wed, Thu, Fri, Sat).
 */
export function formatDateTime(date: Date): string {
  const month = date.getMonth() + 1
  const day = date.getDate()
  const weekday = WEEKDAY_NAMES[date.getDay()]
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${month}/${day} (${weekday}) ${hours}:${minutes}`
}

function matchesEntry(date: Date, entry: CalendarEntry): boolean {
  if (entry.minute !== undefined && date.getMinutes() !== entry.minute) return false
  if (entry.hour !== undefined && date.getHours() !== entry.hour) return false
  if (entry.day !== undefined && date.getDate() !== entry.day) return false
  if (entry.weekday !== undefined && date.getDay() !== entry.weekday) return false
  if (entry.month !== undefined && date.getMonth() + 1 !== entry.month) return false
  return true
}
