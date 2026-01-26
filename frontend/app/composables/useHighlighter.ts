import { createHighlighter, type Highlighter } from 'shiki'

let highlighter: Highlighter | null = null

export async function useHighlighter() {
  if (!highlighter) {
    highlighter = await createHighlighter({
      themes: ['github-dark'],
      langs: ['xml'],
    })
  }
  return highlighter
}

export function highlightCode(code: string): Promise<string> {
  return useHighlighter().then(h =>
    h.codeToHtml(code, { lang: 'xml', theme: 'github-dark' })
  )
}
