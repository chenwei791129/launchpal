export function formatTimestamp(timestamp: string): string {
  const d = new Date(timestamp)
  const date = d.toLocaleDateString('en-CA')
  const time = d.toLocaleTimeString('en-CA', { hour12: false })
  return `${date} ${time}`
}
