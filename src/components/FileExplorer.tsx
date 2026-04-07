import React, { useState, useEffect } from 'react'
import wsService from '../services/websocket'

interface FileNode {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
}

const FileExplorer: React.FC = () => {
  const [currentPath, setCurrentPath] = useState('.')
  const [files, setFiles] = useState<FileNode[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [selectedFile, setSelectedFile] = useState<FileNode | null>(null)
  const [fileContent, setFileContent] = useState('')

  useEffect(() => {
    const unsub = wsService.on('list_directory_result', (msg: any) => {
      setFiles(msg.payload.files)
      setIsLoading(false)
    })

    const unsubRead = wsService.on('read_file_result', (msg: any) => {
      setFileContent(msg.payload.content)
      setIsLoading(false)
    })

    fetchFiles('.')

    return () => {
      unsub()
      unsubRead()
    }
  }, [])

  const fetchFiles = (path: string) => {
    setIsLoading(true)
    setCurrentPath(path)
    wsService.send('list_directory', { path })
  }

  const handleOpen = (node: FileNode) => {
    if (node.isDir) {
      fetchFiles(node.path)
    } else {
      setIsLoading(true)
      setSelectedFile(node)
      wsService.send('read_file', { path: node.path })
    }
  }

  const handleBack = () => {
    const parts = currentPath.split(/[\/\\]/)
    if (parts.length > 1) {
      parts.pop()
      fetchFiles(parts.join('/') || '.')
    } else {
      fetchFiles('.')
    }
  }

  return (
    <div className="file-explorer">
      <div className="explorer-header">
        <div className="explorer-path">
          <button className="back-btn" onClick={handleBack}>↑</button>
          <span>{currentPath}</span>
        </div>
        <div className="explorer-actions">
          <button onClick={() => fetchFiles(currentPath)}>↻</button>
        </div>
      </div>

      <div className="explorer-main">
        <div className="explorer-list">
          {isLoading && <div className="loader">Loading...</div>}
          {files.map((file, i) => (
            <div 
              key={i} 
              className={`file-item ${file.isDir ? 'dir' : 'file'}`}
              onClick={() => handleOpen(file)}
            >
              <span className="file-icon">{file.isDir ? '📁' : '📄'}</span>
              <span className="file-name">{file.name}</span>
              <span className="file-size">{file.isDir ? '--' : `${(file.size / 1024).toFixed(1)} KB`}</span>
            </div>
          ))}
        </div>

        {selectedFile && (
          <div className="file-viewer">
            <div className="viewer-header">
              <span className="viewer-title">{selectedFile.name}</span>
              <button className="close-viewer" onClick={() => setSelectedFile(null)}>×</button>
            </div>
            <pre className="viewer-content">
              {fileContent}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}

export default FileExplorer
