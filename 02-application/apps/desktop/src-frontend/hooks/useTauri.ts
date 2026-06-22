import { useEffect, useState } from 'react'

export function useTauri() {
  const [isTauri, setIsTauri] = useState(false)

  useEffect(() => {
    // Check if we're running in Tauri
    setIsTauri(typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window)
  }, [])

  return { isTauri }
}
