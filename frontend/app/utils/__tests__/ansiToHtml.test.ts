import { describe, it, expect } from 'vitest'
import { ansiToHtml, escapeHtml } from '~/utils/ansiToHtml'

// The expected hex values mirror the spec's "SGR code to style mapping" table.
// They are duplicated here on purpose (spec-by-example): the implementation's
// SGR_COLOR_MAP must match these, so the test must not import that constant.
const FG_HEX: Record<number, string> = {
  30: '#5c6370',
  31: '#e06c75',
  32: '#98c379',
  33: '#e5c07b',
  34: '#61afef',
  35: '#c678dd',
  36: '#56b6c2',
  37: '#abb2bf',
  90: '#828896',
  91: '#ff7b85',
  92: '#b5e890',
  93: '#ffd97d',
  94: '#82c8ff',
  95: '#e08af0',
  96: '#73d1de',
  97: '#ffffff',
}

// Background codes reuse the same hex palette as their foreground counterparts.
const BG_HEX: Record<number, string> = {
  40: FG_HEX[30]!,
  41: FG_HEX[31]!,
  42: FG_HEX[32]!,
  43: FG_HEX[33]!,
  44: FG_HEX[34]!,
  45: FG_HEX[35]!,
  46: FG_HEX[36]!,
  47: FG_HEX[37]!,
  100: FG_HEX[90]!,
  101: FG_HEX[91]!,
  102: FG_HEX[92]!,
  103: FG_HEX[93]!,
  104: FG_HEX[94]!,
  105: FG_HEX[95]!,
  106: FG_HEX[96]!,
  107: FG_HEX[97]!,
}

describe('ansiToHtml – SGR code to style mapping', () => {
  for (const [code, hex] of Object.entries(FG_HEX)) {
    it(`maps foreground code ${code} to color:${hex}`, () => {
      expect(ansiToHtml(`\x1b[${code}mX\x1b[0m`)).toBe(`<span style="color:${hex}">X</span>`)
    })
  }

  for (const [code, hex] of Object.entries(BG_HEX)) {
    it(`maps background code ${code} to background-color:${hex}`, () => {
      expect(ansiToHtml(`\x1b[${code}mX\x1b[0m`)).toBe(`<span style="background-color:${hex}">X</span>`)
    })
  }

  it('maps bold (1) to font-weight:bold', () => {
    expect(ansiToHtml('\x1b[1mX\x1b[0m')).toBe('<span style="font-weight:bold">X</span>')
  })

  it('maps underline (4) to text-decoration:underline', () => {
    expect(ansiToHtml('\x1b[4mX\x1b[0m')).toBe('<span style="text-decoration:underline">X</span>')
  })

  it('treats reset (0) as closing the span so plain text is unwrapped', () => {
    expect(ansiToHtml('\x1b[0mX')).toBe('X')
  })
})

describe('ansiToHtml – plain text', () => {
  it('returns plain text unchanged with no span', () => {
    const out = ansiToHtml('hello world\n')
    expect(out).toBe('hello world\n')
    expect(out).not.toContain('<span')
  })
})

describe('escapeHtml', () => {
  it('escapes HTML special characters in plain text', () => {
    expect(ansiToHtml('<script>alert(1)</script>')).toBe('&lt;script&gt;alert(1)&lt;/script&gt;')
    expect(ansiToHtml('<script>alert(1)</script>')).not.toContain('<script>')
  })

  it('escapes quotes inside plain text', () => {
    expect(ansiToHtml('value="hi"')).toBe('value=&quot;hi&quot;')
  })

  it('escapes the full character set directly', () => {
    expect(escapeHtml(`&<>"'`)).toBe('&amp;&lt;&gt;&quot;&#39;')
  })
})

describe('ansiToHtml – core SGR parsing', () => {
  it('renders a single foreground color span', () => {
    const out = ansiToHtml('\x1b[31mERROR\x1b[0m: connection refused')
    expect(out).toContain('<span style="color:#e06c75">ERROR</span>')
    expect(out).toContain(': connection refused')
  })

  it('combines bold from a separate sequence with a color into one span', () => {
    const out = ansiToHtml('\x1b[1m\x1b[33mWARN\x1b[0m')
    expect(out.match(/<span/g)).toHaveLength(1)
    expect(out).toContain('font-weight:bold')
    expect(out).toContain('color:#e5c07b')
    expect(out).toContain('>WARN</span>')
  })

  it('applies multiple parameters in one SGR sequence', () => {
    const out = ansiToHtml('\x1b[1;31mFATAL\x1b[0m')
    expect(out.match(/<span/g)).toHaveLength(1)
    expect(out).toContain('font-weight:bold')
    expect(out).toContain('color:#e06c75')
    expect(out).toContain('>FATAL</span>')
  })
})

describe('ansiToHtml – stripping unsupported and malformed sequences', () => {
  it('strips a 256-color escape', () => {
    const out = ansiToHtml('\x1b[38;5;33mhi\x1b[0m')
    expect(out).toContain('hi')
    expect(out).not.toContain('<span')
  })

  it('strips a non-SGR CSI escape', () => {
    expect(ansiToHtml('\x1b[2Jcleared')).toBe('cleared')
  })

  it('strips an unterminated CSI at end of input', () => {
    expect(ansiToHtml('text\x1b[31')).toBe('text')
  })

  it('strips an SGR with mixed supported and unsupported parameters', () => {
    const out = ansiToHtml('\x1b[1;38;5;33mhi\x1b[0m')
    expect(out).toBe('hi')
  })

  it('strips an OSC sequence with a BEL terminator', () => {
    expect(ansiToHtml('\x1b]0;window-title\x07text')).toBe('text')
  })

  it('strips an SGR whose param has a non-digit byte (parseInt truncation guard)', () => {
    // `<` is in the CSI parameter byte range (0x30-0x3F); parseInt("1<2") would
    // truncate to 1 and wrongly apply bold. A non-numeric param strips the whole
    // sequence instead.
    expect(ansiToHtml('\x1b[1<2mhi\x1b[0m')).toBe('hi')
  })

  it('strips a colon-delimited SGR param', () => {
    expect(ansiToHtml('\x1b[1:31mhi\x1b[0m')).toBe('hi')
  })

  // Spec "stripping behavior cases" example table — one assertion per row.
  const STRIP_CASES: Array<[string, string]> = [
    ['\x1b[38;2;255;0;0mhi\x1b[0m', 'hi'],
    ['\x1b[5mblink\x1b[0m', 'blink'],
    ['\x1b[Hhi', 'hi'],
    ['\x1b[zzzhi', 'hi'],
    ['a\x1b[31', 'a'],
    ['\x1b]0;title\x1b\\after', 'after'],
  ]
  for (const [input, expected] of STRIP_CASES) {
    it(`strips ${JSON.stringify(input)} down to ${JSON.stringify(expected)}`, () => {
      expect(ansiToHtml(input)).toBe(expected)
    })
  }
})

describe('ansiToHtml – span lifecycle', () => {
  it('auto-closes an unterminated span at end of input', () => {
    expect(ansiToHtml('\x1b[31mhi')).toBe('<span style="color:#e06c75">hi</span>')
  })

  it('treats an empty-parameter sequence as reset', () => {
    expect(ansiToHtml('\x1b[mhi')).toBe('hi')
  })

  it('does not wrap plain text that follows a reset', () => {
    expect(ansiToHtml('\x1b[31mA\x1b[0mB')).toBe('<span style="color:#e06c75">A</span>B')
  })
})

describe('ansiToHtml – XSS boundary', () => {
  it('escapes an attacker payload inside an SGR-wrapped span', () => {
    const out = ansiToHtml('\x1b[31m" onmouseover=alert(1) x="\x1b[0m')
    expect(out).toBe('<span style="color:#e06c75">&quot; onmouseover=alert(1) x=&quot;</span>')
    // The only attribute emitted is the controlled style attribute.
    expect(out.match(/<span style="[^"]*">/)).not.toBeNull()
    // No raw double-quote can break out of the style or open a new attribute.
    expect(out).not.toMatch(/ onmouseover="/)
  })

  it('only emits whitelisted style properties', () => {
    const out = ansiToHtml('\x1b[1;4;31;41mx\x1b[0m')
    const style = out.match(/<span style="([^"]*)">/)?.[1] ?? ''
    const props = style.split(';').map(s => s.split(':')[0])
    for (const p of props) {
      expect(['font-weight', 'text-decoration', 'color', 'background-color']).toContain(p)
    }
  })
})
