/**
 * Parse a space-separated argument string into an array,
 * treating single-quoted and double-quoted segments as single arguments.
 */
export function parseShellArgs(input: string): string[] {
  const args: string[] = []
  let current = ''
  let quote: string | null = null

  for (const char of input) {
    if (quote) {
      if (char === quote) {
        quote = null
      } else {
        current += char
      }
    } else if (char === "'" || char === '"') {
      quote = char
    } else if (/\s/.test(char)) {
      if (current) {
        args.push(current)
        current = ''
      }
    } else {
      current += char
    }
  }

  if (current) {
    args.push(current)
  }

  return args
}

/**
 * Serialize an argument array back to a space-separated string.
 * Arguments containing whitespace are wrapped in single quotes.
 */
export function serializeShellArgs(args: string[]): string {
  return args
    .map((arg) => (/\s/.test(arg) ? `'${arg}'` : arg))
    .join(' ')
}
