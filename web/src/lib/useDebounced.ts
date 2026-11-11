import { useEffect, useState } from 'react'

// Returns value, but only after it has stopped changing for `delay` ms.
// Used so the editor preview re-renders after you pause typing, not on every key.
export function useDebounced<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(t)
  }, [value, delay])
  return debounced
}
