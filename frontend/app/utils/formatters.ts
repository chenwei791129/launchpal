// Renders the timestamp in the user's local timezone — backup IDs come back
// as RFC3339 UTC and the UI shows them in wall-clock terms. Don't swap the
// getX() calls for getUTCX(); that would shift the displayed date away from
// what the user expects.
export function formatTimestamp(timestamp: string): string {
  const d = new Date(timestamp)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
