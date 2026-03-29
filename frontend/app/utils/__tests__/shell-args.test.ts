import { describe, expect, it } from 'vitest'
import { parseShellArgs, serializeShellArgs } from '../shell-args'

describe('parseShellArgs', () => {
  it('parses simple unquoted arguments', () => {
    expect(parseShellArgs('--print --verbose --model opus')).toEqual([
      '--print',
      '--verbose',
      '--model',
      'opus',
    ])
  })

  it('parses single-quoted argument with spaces', () => {
    expect(
      parseShellArgs("--print -p 'run daily backup and send report'"),
    ).toEqual(['--print', '-p', 'run daily backup and send report'])
  })

  it('parses double-quoted argument with spaces', () => {
    expect(parseShellArgs('--message "hello world" --verbose')).toEqual([
      '--message',
      'hello world',
      '--verbose',
    ])
  })

  it('parses mixed quoted and unquoted arguments', () => {
    expect(
      parseShellArgs('--flag \'value one\' --other "value two" plain'),
    ).toEqual(['--flag', 'value one', '--other', 'value two', 'plain'])
  })

  it('returns empty array for empty input', () => {
    expect(parseShellArgs('')).toEqual([])
  })

  it('returns empty array for whitespace-only input', () => {
    expect(parseShellArgs('   ')).toEqual([])
  })
})

describe('serializeShellArgs', () => {
  it('serializes arguments with spaces using single quotes', () => {
    expect(
      serializeShellArgs([
        '--print',
        '-p',
        'run daily backup and send report',
      ]),
    ).toBe("--print -p 'run daily backup and send report'")
  })

  it('serializes arguments without spaces as-is', () => {
    expect(serializeShellArgs(['--verbose', '--model', 'opus'])).toBe(
      '--verbose --model opus',
    )
  })

  it('returns empty string for empty array', () => {
    expect(serializeShellArgs([])).toBe('')
  })
})

describe('round-trip consistency', () => {
  it('parse then serialize preserves the original', () => {
    const input = "--print -p 'run daily backup and send report'"
    expect(serializeShellArgs(parseShellArgs(input))).toBe(input)
  })

  it('serialize then parse preserves the array', () => {
    const args = ['--print', '-p', 'run daily backup and send report']
    expect(parseShellArgs(serializeShellArgs(args))).toEqual(args)
  })
})
