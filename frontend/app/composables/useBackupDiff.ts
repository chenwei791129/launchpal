import { diffLines, type Change } from 'diff'

export const MAX_DIFF_ROWS = 10_000

export type DiffRowType = 'context' | 'added' | 'removed' | 'placeholder'

export interface DiffRow {
  type: DiffRowType
  text: string
  lineNumber?: number
}

export interface SideBySideDiff {
  left: DiffRow[]
  right: DiffRow[]
  hasChanges: boolean
  truncated: boolean
  omittedRows: number
}

function splitLines(value: string): string[] {
  if (value === '') return []
  const lines = value.split('\n')
  if (lines.length > 0 && lines[lines.length - 1] === '') {
    lines.pop()
  }
  return lines
}

function placeholderRow(): DiffRow {
  return { type: 'placeholder', text: '' }
}

export function computeSideBySideDiff(current: string, backup: string): SideBySideDiff {
  const changes: Change[] = diffLines(current, backup)

  const left: DiffRow[] = []
  const right: DiffRow[] = []
  let leftLine = 0
  let rightLine = 0
  let hasChanges = false

  for (const change of changes) {
    const lines = splitLines(change.value)
    if (lines.length === 0) continue

    if (change.added) {
      hasChanges = true
      for (const text of lines) {
        rightLine += 1
        right.push({ type: 'added', text, lineNumber: rightLine })
        left.push(placeholderRow())
      }
    } else if (change.removed) {
      hasChanges = true
      for (const text of lines) {
        leftLine += 1
        left.push({ type: 'removed', text, lineNumber: leftLine })
        right.push(placeholderRow())
      }
    } else {
      for (const text of lines) {
        leftLine += 1
        rightLine += 1
        left.push({ type: 'context', text, lineNumber: leftLine })
        right.push({ type: 'context', text, lineNumber: rightLine })
      }
    }
  }

  const totalRows = left.length
  let truncated = false
  let omittedRows = 0
  if (totalRows > MAX_DIFF_ROWS) {
    truncated = true
    omittedRows = totalRows - MAX_DIFF_ROWS
    // Invariant: left and right arrays are pushed in lockstep above, so they
    // always have equal length; truncating both to MAX_DIFF_ROWS preserves row alignment.
    left.length = MAX_DIFF_ROWS
    right.length = MAX_DIFF_ROWS
  }

  return { left, right, hasChanges, truncated, omittedRows }
}
