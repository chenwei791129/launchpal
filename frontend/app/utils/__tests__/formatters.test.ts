import { describe, expect, it } from 'vitest'
import { formatTimestamp } from '../formatters'

describe('formatTimestamp', () => {
  const sample = '2026-01-03T14:05:09Z'

  it('produces ISO-style YYYY-MM-DD HH:mm:ss output', () => {
    expect(formatTimestamp(sample)).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
  })

  it('does not include CJK characters regardless of host locale', () => {
    expect(formatTimestamp(sample)).not.toMatch(/[一-鿿]/)
  })
})
