import React, { useState, useEffect } from 'react'

interface FileNode {
  name: string
  path: string
  is_dir: boolean
  size: number
  mod_time: string
}

const FileExplorer: React.FC = () => {
  const [currentPath, setCurrentPath] = useState('')
  const [files, setFiles] = useState<FileNode[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [selectedFile, setSelectedFile] = useState<FileNode | null>(null)
  const [fileContent, setFileContent] = useState('')
  const [drives, setDrives] = useState<string[]>([])

  useEffect(() => {
    fetchDrives()
    fetchFiles('')
  }, [])

  const fetchDrives = async () => {
    try {
      const res = await fetch('/api/files/drives')
      const data = await res.json()
      if (data.success) setDrives(data.data || [])
    } catch (e) {
      console.error('Failed to fetch drives', e)
    }
  }

  const fetchFiles = async (path: string) => {
    setIsLoading(true)
    try {
      const cleanPath = path.trim()
      const res = await fetch(`/api/files?path=${encodeURIComponent(cleanPath)}`)
      const data = await res.json()
      if (data.success) {
        setFiles(data.data || [])
        if (data.data && data.data.length > 0) {
           const firstItem = data.data[0]
           const lastSep = firstItem.path.lastIndexOf('/') !== -1 ? firstItem.path.lastIndexOf('/') : firstItem.path.lastIndexOf('\\')
           const parent = firstItem.path.substring(0, lastSep)
           setCurrentPath(parent || cleanPath || 'Project Root')
        } else {
           setCurrentPath(cleanPath || 'Project Root')
        }
      }
    } catch (e) {
      console.error('Fetch error:', e)
    } finally {
      setIsLoading(false)
    }
  }

  const handleOpen = async (node: FileNode) => {
    if (node.is_dir) {
      fetchFiles(node.path)
    } else {
      setIsLoading(true)
      setSelectedFile(node)
      try {
        const res = await fetch(`/api/files/read?path=${encodeURIComponent(node.path)}`)
        const data = await res.json()
        if (data.success) setFileContent(data.data.content)
      } catch (e) {
        setFileContent('Error reading file')
      } finally {
        setIsLoading(false)
      }
    }
  }

  const handleBack = () => {
    const sep = currentPath.includes('\\') ? '\\' : '/'
    const parts = currentPath.split(sep).filter(Boolean)
    if (parts.length > 0) {
      parts.pop()
      const newPath = parts.join(sep)
      fetchFiles(newPath + (newPath.endsWith(':') ? sep : ''))
    }
  }

  return (
    <div className="file-explorer">
      <div className="file-toolbar">
         <div className="explorer-controls">
            <button className="toolbar-btn" onClick={handleBack} title="Back">
               <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <path d="M19 12H5M12 19l-7-7 7-7" />
               </svg>
            </button>
            <button className="toolbar-btn" onClick={() => fetchFiles(currentPath)} title="Refresh">
               <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <path d="M23 4h-6M1 20v-6h6M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15" />
               </svg>
            </button>
         </div>
         
         <div className="drive-selector">
            {drives.map(d => (
               <button key={d} className="drive-pill" onClick={() => fetchFiles(d)}>
                  💽 {d.replace(/[\\/]/g, '')}
               </button>
            ))}
         </div>

         <div className="path-display">
            <span className="path-label">Location:</span>
            <span className="path-text">{currentPath}</span>
         </div>
      </div>

      <div className="file-grid-container">
        {isLoading && <div className="loading-overlay"><div className="spinner"></div></div>}
        
        <div className="file-grid-header">
           <div className="col-name">Name</div>
           <div className="col-size">Size</div>
           <div className="col-date">Modified</div>
        </div>

        <div className="file-list-detailed">
          {files.map((file, i) => (
            <div key={i} className="file-row" onClick={() => handleOpen(file)}>
              <div className="col-name">
                 <span className="file-icon">{file.is_dir ? '📁' : '📄'}</span>
                 <span className="file-name">{file.name}</span>
              </div>
              <div className="col-size">
                 {file.is_dir ? '--' : `${(file.size / 1024).toFixed(1)} KB`}
              </div>
              <div className="col-date">
                 {file.mod_time ? new Date(file.mod_time).toLocaleDateString() : '---'}
              </div>
            </div>
          ))}
          {!isLoading && files.length === 0 && (
            <div className="empty-notif">This directory is empty.</div>
          )}
        </div>
      </div>

      {selectedFile && (
        <div className="file-modal-overlay" onClick={() => setSelectedFile(null)}>
          <div className="file-modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{selectedFile.name}</h3>
              <button className="modal-close" onClick={() => setSelectedFile(null)}>×</button>
            </div>
            <div className="modal-body">
              <pre className="code-view">{fileContent}</pre>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default FileExplorer




