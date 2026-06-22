'use client'

import { useState, useCallback } from 'react'
import { invoke } from '@tauri-apps/api/core'
import type { SftpFile } from '@/types/sftp'

export function useLocalFiles() {
  const [localPath, setLocalPath] = useState('')
  const [localFiles, setLocalFiles] = useState<SftpFile[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const listLocal = useCallback(async (path?: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const targetPath = path ?? (await invoke<string>('get_home_dir'))
      const files = await invoke<SftpFile[]>('list_local_directory', { path: targetPath })
      setLocalFiles(files)
      setLocalPath(targetPath)
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      console.error('Local file list error:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const navigateLocalUp = useCallback(async () => {
    try {
      const parent = await invoke<string>('get_parent_dir', { path: localPath })
      listLocal(parent)
    } catch {
      // noop
    }
  }, [localPath, listLocal])

  const navigateLocalTo = useCallback(
    (path: string) => {
      listLocal(path)
    },
    [listLocal]
  )

  const openFilePicker = useCallback(async (): Promise<string | null> => {
    try {
      const selected = await invoke<string | null>('open_file_dialog', {
        directory: true,
        multiple: false,
      })
      if (selected) {
        listLocal(selected)
      }
      return selected
    } catch (err) {
      console.error('File picker error:', err)
      return null
    }
  }, [listLocal])

  return {
    localPath,
    localFiles,
    isLoading,
    error,
    listLocal,
    navigateLocalUp,
    navigateLocalTo,
    openFilePicker,
  }
}
