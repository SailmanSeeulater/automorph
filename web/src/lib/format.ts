// "AM0007": the catalogue-style number shown next to every post.
export function postCode(id: number): string {
  return 'AM' + String(id).padStart(4, '0')
}

const pad = (n: number) => String(n).padStart(2, '0')

// "2026.11.03", in the reader's own time zone. The server sends UTC, so a post
// published at 5pm in California would otherwise show tomorrow's date.
export function formatDate(iso: string | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}.${pad(d.getMonth() + 1)}.${pad(d.getDate())}`
}
