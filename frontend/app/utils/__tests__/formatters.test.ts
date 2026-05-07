import { describe, expect, it } from 'vitest'
import { formatTimestamp } from '../formatters'

describe('formatTimestamp', () => {
  const sample = '2026-01-03T14:05:09Z'

  it('produces ISO-style YYYY-MM-DD HH:mm:ss output', () => {
    expect(formatTimestamp(sample)).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
  })

  it('does not include Asian-style year/month/day markers or Chinese characters', () => {
    const out = formatTimestamp(sample)
    expect(out).not.toMatch(/[一-鿿]/)
    expect(out).not.toMatch(/[年月日]/)
  })
})
