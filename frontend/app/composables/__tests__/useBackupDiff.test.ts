import { describe, expect, it } from 'vitest'
import { computeSideBySideDiff, MAX_DIFF_ROWS } from '../useBackupDiff'

describe('computeSideBySideDiff', () => {
  it('aligns unchanged lines on both sides without markers', () => {
    const current = 'line a\nline b\nline c\n'
    const backup = 'line a\nline b\nline c\n'

    const result = computeSideBySideDiff(current, backup)

    expect(result.left).toHaveLength(result.right.length)
    expect(result.hasChanges).toBe(false)
    for (const row of result.left) {
      expect(row.type).toBe('context')
    }
    for (const row of result.right) {
      expect(row.type).toBe('context')
    }
  })

  it('marks added-only lines on the right with placeholders on the left', () => {
    const current = 'line a\nline c\n'
    const backup = 'line a\nline b\nline c\n'

    const result = computeSideBySideDiff(current, backup)

    expect(result.hasChanges).toBe(true)
    expect(result.left).toHaveLength(result.right.length)

    // Find the added row on the right
    const addedRightIdx = result.right.findIndex(r => r.type === 'added' && r.text === 'line b')
    expect(addedRightIdx).toBeGreaterThanOrEqual(0)
    // Same-index row on the left should be a placeholder
    expect(result.left[addedRightIdx]!.type).toBe('placeholder')
  })

  it('marks removed-only lines on the left with placeholders on the right', () => {
    const current = 'line a\nline b\nline c\n'
    const backup = 'line a\nline c\n'

    const result = computeSideBySideDiff(current, backup)

    expect(result.hasChanges).toBe(true)
    const removedLeftIdx = result.left.findIndex(r => r.type === 'removed' && r.text === 'line b')
    expect(removedLeftIdx).toBeGreaterThanOrEqual(0)
    expect(result.right[removedLeftIdx]!.type).toBe('placeholder')
  })

  it('treats empty current as every backup line being an addition', () => {
    const current = ''
    const backup = 'alpha\nbeta\ngamma\n'

    const result = computeSideBySideDiff(current, backup)

    expect(result.hasChanges).toBe(true)
    expect(result.right.filter(r => r.type === 'added')).toHaveLength(3)
    expect(result.left.every(r => r.type === 'placeholder')).toBe(true)
  })

  it('treats empty backup as every current line being a deletion', () => {
    const current = 'alpha\nbeta\n'
    const backup = ''

    const result = computeSideBySideDiff(current, backup)

    expect(result.hasChanges).toBe(true)
    expect(result.left.filter(r => r.type === 'removed')).toHaveLength(2)
    expect(result.right.every(r => r.type === 'placeholder')).toBe(true)
  })

  it('handles removed+added pairs by placing each on its own aligned row', () => {
    const current = 'same\nold\nsame\n'
    const backup = 'same\nnew\nsame\n'

    const result = computeSideBySideDiff(current, backup)

    expect(result.hasChanges).toBe(true)
    // Expect one removed row on left and one added row on right
    expect(result.left.filter(r => r.type === 'removed')).toHaveLength(1)
    expect(result.right.filter(r => r.type === 'added')).toHaveLength(1)
    // And two aligned context rows on each side
    expect(result.left.filter(r => r.type === 'context')).toHaveLength(2)
    expect(result.right.filter(r => r.type === 'context')).toHaveLength(2)
    // Total rows identical on both sides
    expect(result.left.length).toBe(result.right.length)
  })

  it('truncates to MAX_DIFF_ROWS and reports omitted count', () => {
    const manyLines = Array.from({ length: MAX_DIFF_ROWS + 50 }, (_, i) => `line ${i}`).join('\n') + '\n'

    const result = computeSideBySideDiff('', manyLines)

    expect(result.left.length).toBeLessThanOrEqual(MAX_DIFF_ROWS)
    expect(result.right.length).toBeLessThanOrEqual(MAX_DIFF_ROWS)
    expect(result.truncated).toBe(true)
    expect(result.omittedRows).toBeGreaterThan(0)
  })

  it('numbers left rows in current and right rows in backup, skipping placeholders', () => {
    const current = 'a\nb\n'
    const backup = 'a\nX\nb\n'

    const result = computeSideBySideDiff(current, backup)

    // Left content rows should be numbered 1, 2
    const leftContentNumbers = result.left.filter(r => r.type !== 'placeholder').map(r => r.lineNumber)
    expect(leftContentNumbers).toEqual([1, 2])
    // Right content rows should be numbered 1, 2, 3
    const rightContentNumbers = result.right.filter(r => r.type !== 'placeholder').map(r => r.lineNumber)
    expect(rightContentNumbers).toEqual([1, 2, 3])
    // Placeholders carry no line number
    for (const row of [...result.left, ...result.right]) {
      if (row.type === 'placeholder') {
        expect(row.lineNumber).toBeUndefined()
      }
    }
  })
})
