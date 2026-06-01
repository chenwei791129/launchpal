// ansiToHtml converts log text containing ANSI SGR escape sequences into an
// HTML string suitable for v-html binding. Parsing happens entirely on the
// frontend after the raw log string arrives from GetLogs / GetSystemLogs — no
// Go-side change and no third-party ANSI library. The supported SGR subset is
// small (reset, bold, underline, and the 16-color foreground/background sets),
// so a hand-written linear-scan state machine keeps the HTML-escape and
// style-whitelist boundaries fully under our control (the v-html XSS surface).

// SGR_COLOR_MAP maps every supported color SGR code to a concrete hex value.
// Foreground (30-37 / 90-97) and background (40-47 / 100-107) codes share the
// same OneDark-derived palette chosen for contrast on LaunchPal's dark
// `bg-surface-500`. Foreground codes feed `color`, background codes feed
// `background-color`; the hex is identical for a code and its +10 partner.
export const SGR_COLOR_MAP: Record<number, string> = {
  // normal foreground
  30: '#5c6370',
  31: '#e06c75',
  32: '#98c379',
  33: '#e5c07b',
  34: '#61afef',
  35: '#c678dd',
  36: '#56b6c2',
  37: '#abb2bf',
  // bright foreground
  90: '#828896',
  91: '#ff7b85',
  92: '#b5e890',
  93: '#ffd97d',
  94: '#82c8ff',
  95: '#e08af0',
  96: '#73d1de',
  97: '#ffffff',
  // normal background (same hex as the matching foreground code minus 10)
  40: '#5c6370',
  41: '#e06c75',
  42: '#98c379',
  43: '#e5c07b',
  44: '#61afef',
  45: '#c678dd',
  46: '#56b6c2',
  47: '#abb2bf',
  // bright background (same hex as the matching bright foreground code minus 10)
  100: '#828896',
  101: '#ff7b85',
  102: '#b5e890',
  103: '#ffd97d',
  104: '#82c8ff',
  105: '#e08af0',
  106: '#73d1de',
  107: '#ffffff',
}

// HTML_ENTITIES maps the five HTML-sensitive characters to their entities so no
// log content can inject markup once mounted via v-html.
const HTML_ENTITIES: Record<string, string> = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
}

// escapeHtml replaces every HTML-sensitive character in a single pass.
export function escapeHtml(text: string): string {
  return text.replace(/[&<>"']/g, ch => HTML_ENTITIES[ch]!)
}

// SgrState holds the currently-active SGR attributes. A null fg/bg means the
// attribute is not set; a non-null value is always drawn from SGR_COLOR_MAP.
interface SgrState {
  bold: boolean
  underline: boolean
  fg: string | null
  bg: string | null
}

function emptyState(): SgrState {
  return { bold: false, underline: false, fg: null, bg: null }
}

// isSupportedSgrCode reports whether a single SGR parameter belongs to the
// supported subset. An SGR sequence is only applied when *every* parameter is
// supported; a single unsupported parameter strips the whole sequence.
function isSupportedSgrCode(code: number): boolean {
  return code === 0 || code === 1 || code === 4
    || (code >= 30 && code <= 37) || (code >= 40 && code <= 47)
    || (code >= 90 && code <= 97) || (code >= 100 && code <= 107)
}

// applySgrCode mutates the active state for one supported parameter.
function applySgrCode(state: SgrState, code: number): void {
  if (code === 0) {
    Object.assign(state, emptyState())
  } else if (code === 1) {
    state.bold = true
  } else if (code === 4) {
    state.underline = true
  } else if ((code >= 30 && code <= 37) || (code >= 90 && code <= 97)) {
    state.fg = SGR_COLOR_MAP[code]!
  } else if ((code >= 40 && code <= 47) || (code >= 100 && code <= 107)) {
    state.bg = SGR_COLOR_MAP[code]!
  }
}

// styleString renders the active state into the inline style value. The
// property order is fixed (font-weight, text-decoration, color,
// background-color) and every value is a literal or a SGR_COLOR_MAP hex — no
// portion of the log text ever reaches a style attribute.
function styleString(state: SgrState): string {
  const frags: string[] = []
  if (state.bold) frags.push('font-weight:bold')
  if (state.underline) frags.push('text-decoration:underline')
  if (state.fg) frags.push(`color:${state.fg}`)
  if (state.bg) frags.push(`background-color:${state.bg}`)
  return frags.join(';')
}

export function ansiToHtml(input: string): string {
  if (!input) return ''

  const state = emptyState()
  // Output fragments are collected in an array and join()-ed once at the end,
  // rather than concatenated onto a growing string — heavily decorated logs
  // (e.g. `docker compose logs`) emit one fragment per escape sequence, and
  // array+join avoids the quadratic blow-up of repeated `+=` on a large buffer.
  const parts: string[] = []
  let openStyle: string | null = null
  let buffer = ''

  // flush emits the buffered plain text, lazily (re)opening a span only when
  // the desired style differs from the one currently open. Lazy opening lets
  // consecutive escape sequences (e.g. bold then color) collapse into a single
  // span around the text that actually follows them.
  const flush = (): void => {
    if (buffer === '') return
    const desired = styleString(state)
    if (desired !== openStyle) {
      if (openStyle !== null) parts.push('</span>')
      openStyle = desired !== '' ? desired : null
      if (desired !== '') parts.push(`<span style="${desired}">`)
    }
    parts.push(escapeHtml(buffer))
    buffer = ''
  }

  const len = input.length
  let i = 0
  while (i < len) {
    const ch = input[i]!
    if (ch !== '\x1b') {
      buffer += ch
      i += 1
      continue
    }

    // ESC begins a control sequence; the buffered text belongs to the state in
    // effect before it, so flush first.
    flush()

    const next = input[i + 1]
    if (next === undefined) {
      // Lone trailing ESC — drop it.
      i = len
      break
    }

    if (next === '[') {
      // CSI sequence: ESC [ <params 0x30-0x3F> <intermediates 0x20-0x2F> <final>
      let j = i + 2
      const paramsStart = j
      while (j < len && input.charCodeAt(j) >= 0x30 && input.charCodeAt(j) <= 0x3f) j += 1
      const paramsEnd = j
      while (j < len && input.charCodeAt(j) >= 0x20 && input.charCodeAt(j) <= 0x2f) j += 1
      if (j >= len) {
        // Unterminated CSI: strip through end of input.
        i = len
        break
      }
      const finalCode = input.charCodeAt(j)
      const final = input[j]!
      if (finalCode >= 0x40 && finalCode <= 0x6f) {
        // Standard final byte. Treat ESC [ <params> m as SGR; any other final
        // byte (or any sequence carrying intermediates) is a non-SGR CSI that
        // we strip entirely.
        if (final === 'm' && paramsEnd === j) {
          // Params live in the 0x30-0x3F byte range, which also covers `:;<=>?`.
          // A param must be purely numeric (empty = 0, the ANSI default) — any
          // non-digit byte (e.g. the colon in `38:2:...` or a stray `<`) makes
          // it NaN, which fails the every() check below and strips the whole
          // sequence rather than letting parseInt silently truncate `1<2` to 1.
          const paramStr = input.slice(paramsStart, paramsEnd)
          const codes = (paramStr === '' ? [''] : paramStr.split(';'))
            .map(p => (p === '' ? 0 : (/^[0-9]+$/.test(p) ? parseInt(p, 10) : NaN)))
          if (codes.every(c => !Number.isNaN(c) && isSupportedSgrCode(c))) {
            for (const c of codes) applySgrCode(state, c)
          }
          // Unsupported parameter(s) → strip the whole sequence (apply nothing).
        }
        i = j + 1
      } else {
        // Private-use final byte (0x70-0x7E). Per spec the whole run of such
        // bytes is stripped as a unit (e.g. "\x1b[zzzhi" → "hi"); consume the
        // contiguous private-byte run and leave the following standard text.
        while (j < len && input.charCodeAt(j) >= 0x70 && input.charCodeAt(j) <= 0x7e) j += 1
        i = j
      }
      continue
    }

    if (next === ']') {
      // OSC: ESC ] ... terminated by BEL (\x07) or ST (ESC \). Strip the whole
      // sequence including its terminator; if unterminated, strip to end.
      let j = i + 2
      while (j < len) {
        if (input.charCodeAt(j) === 0x07) {
          j += 1
          break
        }
        if (input[j] === '\x1b' && input[j + 1] === '\\') {
          j += 2
          break
        }
        j += 1
      }
      i = j
      continue
    }

    if (next === 'P' || next === 'X' || next === '^' || next === '_') {
      // DCS / SOS / PM / APC: ESC <introducer> ... terminated by ST (ESC \).
      // Strip the whole sequence including the terminator, else to end.
      let j = i + 2
      while (j < len) {
        if (input[j] === '\x1b' && input[j + 1] === '\\') {
          j += 2
          break
        }
        j += 1
      }
      i = j
      continue
    }

    // Any other Fe escape (ESC followed by a single byte, e.g. ESC c) — drop
    // the two-byte sequence.
    i += 2
  }

  flush()
  if (openStyle !== null) parts.push('</span>')
  return parts.join('')
}
