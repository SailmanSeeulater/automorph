import { Link, type ErrorComponentProps } from '@tanstack/react-router'

// Small static pages. Replace the text with your own.

export function About() {
  return (
    <div style={{ maxWidth: 420, display: 'grid', gap: 16 }}>
      <p style={{ margin: 0 }}>Automorph is a journal by Late Sailor.</p>
      <p style={{ margin: 0 }}>Writing, photographs and things in progress.</p>
    </div>
  )
}

export function Shop() {
  return <p style={{ margin: 0 }}>Coming soon.</p>
}

export function NotFound() {
  return (
    <p style={{ margin: 0 }}>
      Not found. <Link to="/">Back to the index</Link>
    </p>
  )
}

export function ErrorPage({ error }: ErrorComponentProps) {
  const message = error instanceof Error ? error.message : String(error)
  return <p style={{ margin: 0 }}>Something went wrong: {message}</p>
}
