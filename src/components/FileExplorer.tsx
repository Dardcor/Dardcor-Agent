import React from 'react'
import { useFileExplorer } from '../hooks/useAgent'

const FileExplorer: React.FC = () => {
  const {
    currentPath,
    files,
    loading,
    error,
    drives,
    loadDirectory,
    goUp,
    refresh,
  } = useFileExplorer()

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '—'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  const getFileIcon = (name: string, isDir: boolean): string => {
    if (isDir) return '📁'
    const ext = name.split('.').pop()?.toLowerCase() || ''
    const iconMap: Record<string, string> = {
      ts: '🔷', tsx: '🔷', js: '🟡', jsx: '🟡',
      go: '🔵', py: '🐍', rs: '🦀', java: '☕',
      html: '🌐', css: '🎨', scss: '🎨',
      json: '📋', xml: '📋', yaml: '📋', yml: '📋', toml: '📋',
      md: '📝', txt: '📄', doc: '📄', docx: '📄',
      pdf: '📕', xls: '📊', xlsx: '📊',
      png: '🖼️', jpg: '🖼️', jpeg: '🖼️', gif: '🖼️', svg: '🖼️', webp: '🖼️',
      mp3: '🎵', wav: '🎵', flac: '🎵',
      mp4: '🎬', avi: '🎬', mkv: '🎬',
      zip: '📦', rar: '📦', '7z': '📦', tar: '📦', gz: '📦',
      exe: '⚙️', msi: '⚙️', dll: '🔧',
      bat: '📜', sh: '📜', ps1: '📜',
      gitignore: '🔒', env: '🔐',
      lock: '🔒', log: '📋',
    }
    return iconMap[ext] || '📄'
  }

  const pathSegments = currentPath.split(/[/\\]/).filter(Boolean)

  return (
    <div className="file-explorer">
      {/* Drive List */}
      <div className="drive-list">
        {drives.map((drive) => (
          <button
            key={drive}
            className={`drive-btn ${currentPath.startsWith(drive) ? 'active' : ''}`}
            onClick={() => loadDirectory(drive)}
          >
            💿 {drive}
          </button>
        ))}
      </div>

      {/* Toolbar */}
      <div className="file-toolbar">
        <button className="toolbar-btn" onClick={goUp} title="Naik satu level" id="file-go-up">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </button>
        <button className="toolbar-btn" onClick={refresh} title="Refresh" id="file-refresh">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="23 4 23 10 17 10" /><polyline points="1 20 1 14 7 14" />
            <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15" />
          </svg>
        </button>
        <div className="file-path-bar">
          {pathSegments.map((segment, i) => (
            <React.Fragment key={i}>
              {i > 0 && <span className="file-path-separator">/</span>}
              <span
                className="file-path-segment"
                onClick={() => {
                  const path = pathSegments.slice(0, i + 1).join('\\')
                  loadDirectory(path + (i === 0 ? '\\' : ''))
                }}
              >
                {segment}
              </span>
            </React.Fragment>
          ))}
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="error-banner">
          ⚠️ {error}
        </div>
      )}

      {/* File List */}
      <div className="file-list">
        {loading ? (
          <div className="loading-spinner">
            <div className="spinner" />
          </div>
        ) : files.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">📂</div>
            <h3>Folder Kosong</h3>
            <p>Tidak ada file atau folder di direktori ini</p>
          </div>
        ) : (
          files
            .sort((a, b) => {
              if (a.is_dir && !b.is_dir) return -1
              if (!a.is_dir && b.is_dir) return 1
              return a.name.localeCompare(b.name)
            })
            .map((file) => (
              <div
                key={file.path}
                className="file-item"
                onClick={() => {
                  if (file.is_dir) {
                    loadDirectory(file.path)
                  }
                }}
                onDoubleClick={() => {
                  if (!file.is_dir) {
                    // Could open file content viewer
                  }
                }}
              >
                <div className="file-item-icon">
                  {getFileIcon(file.name, file.is_dir)}
                </div>
                <div className="file-item-info">
                  <div className="file-item-name">{file.name}</div>
                  <div className="file-item-meta">
                    <span>{file.is_dir ? 'Folder' : file.extension || 'File'}</span>
                    <span>
                      {new Date(file.modified_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        year: 'numeric',
                      })}
                    </span>
                  </div>
                </div>
                <div className="file-item-size">
                  {file.is_dir ? '—' : formatSize(file.size)}
                </div>
              </div>
            ))
        )}
      </div>
    </div>
  )
}

export default FileExplorer
