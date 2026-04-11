import { describe, expect, it } from 'vitest'
import { getNextOccurrences, formatDateTime } from '../useNextOccurrences'

describe('getNextOccurrences', () => {
  // Helper: create a fixed "now" for deterministic tests
  const now = new Date(2026, 3, 3, 10, 30, 0) // Apr 3, 2026 10:30:00

  it('returns dates at specific hour and minute on consecutive days', () => {
    const results = getNextOccurrences(
      { schedules: [{ hour: 14, minute: 30 }] },
      3,
      now,
    )
    expect(results).toHaveLength(3)
    for (const d of results) {
      expect(d.getHours()).toBe(14)
      expect(d.getMinutes()).toBe(30)
    }
    // First occurrence should be today at 14:30 (since now is 10:30)
    expect(results[0]!.getFullYear()).toBe(2026)
    expect(results[0]!.getMonth()).toBe(3) // April
    expect(results[0]!.getDate()).toBe(3)
    // Second should be tomorrow
    expect(results[1]!.getDate()).toBe(4)
  })

  it('returns dates on specific weekday', () => {
    // weekday=1 means Monday in launchd (0=Sun, 1=Mon, ...)
    const results = getNextOccurrences(
      { schedules: [{ weekday: 1, hour: 9, minute: 0 }] },
      3,
      now,
    )
    expect(results).toHaveLength(3)
    for (const d of results) {
      expect(d.getDay()).toBe(1) // Monday
      expect(d.getHours()).toBe(9)
      expect(d.getMinutes()).toBe(0)
    }
  })

  it('returns every minute when all fields are unset', () => {
    const results = getNextOccurrences({ schedules: [{}] }, 3, now)
    expect(results).toHaveLength(3)
    // First should be now+1 minute = 10:31
    expect(results[0]!.getHours()).toBe(10)
    expect(results[0]!.getMinutes()).toBe(31)
    expect(results[1]!.getMinutes()).toBe(32)
    expect(results[2]!.getMinutes()).toBe(33)
  })

  it('returns dates on specific month and day (limited by 400-day scan)', () => {
    // From Apr 3, 2026: 400 days reaches ~May 8, 2027
    // Only Dec 25, 2026 is within range
    const results = getNextOccurrences(
      { schedules: [{ month: 12, day: 25, hour: 0, minute: 0 }] },
      3,
      now,
    )
    expect(results).toHaveLength(1)
    expect(results[0]!.getMonth()).toBe(11) // December (0-indexed)
    expect(results[0]!.getDate()).toBe(25)
    expect(results[0]!.getHours()).toBe(0)
    expect(results[0]!.getMinutes()).toBe(0)
    expect(results[0]!.getFullYear()).toBe(2026)
  })

  it('respects count parameter', () => {
    const results = getNextOccurrences({ schedules: [{ hour: 12, minute: 0 }] }, 5, now)
    expect(results).toHaveLength(5)
  })

  it('returns empty array when no match within 400 days', () => {
    // month=2 (Feb), day=30 — doesn't exist
    const results = getNextOccurrences(
      { schedules: [{ month: 2, day: 30, hour: 0, minute: 0 }] },
      1,
      now,
    )
    expect(results).toHaveLength(0)
  })

  it('starts scanning from now+1 minute, not current minute', () => {
    const exactNow = new Date(2026, 3, 3, 14, 30, 0)
    const results = getNextOccurrences(
      { schedules: [{ hour: 14, minute: 30 }] },
      1,
      exactNow,
    )
    // Should NOT include current time, should be next day
    expect(results[0]!.getDate()).toBe(4)
  })

  it('returns empty array for interval-only config', () => {
    const results = getNextOccurrences({ interval: 300 }, 3, now)
    expect(results).toHaveLength(0)
  })

  it('handles weekday=0 as Sunday', () => {
    // Apr 5, 2026 is a Sunday
    const results = getNextOccurrences(
      { schedules: [{ weekday: 0, hour: 8, minute: 0 }] },
      1,
      now,
    )
    expect(results[0]!.getDay()).toBe(0) // Sunday
    expect(results[0]!.getDate()).toBe(5) // Apr 5
  })

  it('returns empty array for empty schedules', () => {
    const results = getNextOccurrences({ schedules: [] }, 3, now)
    expect(results).toHaveLength(0)
  })

  it('handles multiple schedule entries', () => {
    // Two entries: 9:00 and 17:00 — should interleave
    const results = getNextOccurrences(
      { schedules: [{ hour: 9, minute: 0 }, { hour: 17, minute: 0 }] },
      4,
      now,
    )
    expect(results).toHaveLength(4)
    // First: today 17:00 (since now is 10:30, 9:00 already passed)
    expect(results[0]!.getHours()).toBe(17)
    expect(results[0]!.getDate()).toBe(3)
    // Second: tomorrow 9:00
    expect(results[1]!.getHours()).toBe(9)
    expect(results[1]!.getDate()).toBe(4)
  })
})

describe('formatDateTime', () => {
  it('formats date as M/D (Weekday) HH:mm', () => {
    // Apr 3, 2026 is a Friday
    const date = new Date(2026, 3, 3, 14, 30, 0)
    expect(formatDateTime(date)).toBe('4/3 (Fri) 14:30')
  })

  it('pads hours and minutes with leading zeros', () => {
    const date = new Date(2026, 0, 5, 8, 5, 0) // Jan 5, Mon, 08:05
    expect(formatDateTime(date)).toBe('1/5 (Mon) 08:05')
  })

  it('formats midnight correctly', () => {
    const date = new Date(2026, 11, 25, 0, 0, 0) // Dec 25, Fri, 00:00
    expect(formatDateTime(date)).toBe('12/25 (Fri) 00:00')
  })

  it('formats Sunday correctly', () => {
    const date = new Date(2026, 3, 5, 12, 0, 0) // Apr 5, Sun
    expect(formatDateTime(date)).toBe('4/5 (Sun) 12:00')
  })

  it('formats Saturday correctly', () => {
    const date = new Date(2026, 3, 4, 23, 59, 0) // Apr 4, Sat
    expect(formatDateTime(date)).toBe('4/4 (Sat) 23:59')
  })
})
