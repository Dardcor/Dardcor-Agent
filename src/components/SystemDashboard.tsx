import React from 'react'
import { useSystemInfo, useProcesses } from '../hooks/useAgent'

const SystemDashboard: React.FC = () => {
  const { systemInfo, loading: sysLoading, error: sysError } = useSystemInfo(5000)
  const { processes, loading: procLoading, sortBy, setSortBy, killProcess } = useProcesses(5000)

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    if (days > 0) return `${days}d ${hours}h ${mins}m`
    if (hours > 0) return `${hours}h ${mins}m`
    return `${mins}m`
  }

  const getUsageClass = (percent: number): string => {
    if (percent >= 90) return 'danger'
    if (percent >= 70) return 'warning'
    return ''
  }

  if (sysLoading) {
    return (
      <div className="system-dashboard">
        <div className="loading-spinner">
          <div className="spinner" />
        </div>
      </div>
    )
  }

  if (sysError) {
    return (
      <div className="system-dashboard">
        <div className="error-banner">⚠️ {sysError}</div>
      </div>
    )
  }

  return (
    <div className="system-dashboard">
      <div className="dashboard-grid">
        {/* OS Info Card */}
        <div className="dashboard-card">
          <div className="dashboard-card-header">
            <div className="dashboard-card-title">
              <span>🖥️</span> Sistem Operasi
            </div>
            <span className="dashboard-card-badge">Online</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Hostname</span>
            <span className="stat-value">{systemInfo?.hostname || '—'}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Platform</span>
            <span className="stat-value">{systemInfo?.os.platform || '—'}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Version</span>
            <span className="stat-value">{systemInfo?.os.version || '—'}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Architecture</span>
            <span className="stat-value">{systemInfo?.os.arch || '—'}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Uptime</span>
            <span className="stat-value">{systemInfo ? formatUptime(systemInfo.uptime) : '—'}</span>
          </div>
        </div>

        {/* CPU Card */}
        <div className="dashboard-card">
          <div className="dashboard-card-header">
            <div className="dashboard-card-title">
              <span>🔧</span> Prosesor (CPU)
            </div>
          </div>
          <div className="stat-row">
            <span className="stat-label">Model</span>
            <span className="stat-value" style={{ fontSize: '11px', maxWidth: '200px', textAlign: 'right' }}>
              {systemInfo?.cpu.model_name || '—'}
            </span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Cores / Threads</span>
            <span className="stat-value">{systemInfo?.cpu.cores || 0} / {systemInfo?.cpu.threads || 0}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Frequency</span>
            <span className="stat-value">{systemInfo?.cpu.frequency_mhz?.toFixed(0) || '0'} MHz</span>
          </div>
          <div className="progress-bar-container">
            <div className="progress-bar-label">
              <span>CPU Usage</span>
              <span>{systemInfo?.cpu.usage_percent?.toFixed(1) || '0'}%</span>
            </div>
            <div className="progress-bar">
              <div
                className={`progress-bar-fill ${getUsageClass(systemInfo?.cpu.usage_percent || 0)}`}
                style={{ width: `${systemInfo?.cpu.usage_percent || 0}%` }}
              />
            </div>
          </div>
        </div>

        {/* Memory Card */}
        <div className="dashboard-card">
          <div className="dashboard-card-header">
            <div className="dashboard-card-title">
              <span>🧠</span> Memori (RAM)
            </div>
          </div>
          <div className="stat-row">
            <span className="stat-label">Total</span>
            <span className="stat-value">{systemInfo ? formatBytes(systemInfo.memory.total) : '—'}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Used</span>
            <span className="stat-value">{systemInfo ? formatBytes(systemInfo.memory.used) : '—'}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Free</span>
            <span className="stat-value">{systemInfo ? formatBytes(systemInfo.memory.free) : '—'}</span>
          </div>
          <div className="progress-bar-container">
            <div className="progress-bar-label">
              <span>Memory Usage</span>
              <span>{systemInfo?.memory.used_percent?.toFixed(1) || '0'}%</span>
            </div>
            <div className="progress-bar">
              <div
                className={`progress-bar-fill ${getUsageClass(systemInfo?.memory.used_percent || 0)}`}
                style={{ width: `${systemInfo?.memory.used_percent || 0}%` }}
              />
            </div>
          </div>
        </div>

        {/* Disk Cards */}
        {systemInfo?.disks?.map((disk, i) => (
          <div className="dashboard-card" key={i}>
            <div className="dashboard-card-header">
              <div className="dashboard-card-title">
                <span>💽</span> Disk {disk.mount_point}
              </div>
              <span className="dashboard-card-badge" style={{ fontSize: '10px' }}>{disk.fs_type}</span>
            </div>
            <div className="stat-row">
              <span className="stat-label">Total</span>
              <span className="stat-value">{formatBytes(disk.total)}</span>
            </div>
            <div className="stat-row">
              <span className="stat-label">Used</span>
              <span className="stat-value">{formatBytes(disk.used)}</span>
            </div>
            <div className="stat-row">
              <span className="stat-label">Free</span>
              <span className="stat-value">{formatBytes(disk.free)}</span>
            </div>
            <div className="progress-bar-container">
              <div className="progress-bar-label">
                <span>Disk Usage</span>
                <span>{disk.used_percent?.toFixed(1)}%</span>
              </div>
              <div className="progress-bar">
                <div
                  className={`progress-bar-fill ${getUsageClass(disk.used_percent)}`}
                  style={{ width: `${disk.used_percent}%` }}
                />
              </div>
            </div>
          </div>
        ))}

        {/* Network Card */}
        <div className="dashboard-card">
          <div className="dashboard-card-header">
            <div className="dashboard-card-title">
              <span>🌐</span> Network
            </div>
          </div>
          {systemInfo?.network?.slice(0, 5).map((net, i) => (
            <div className="stat-row" key={i}>
              <span className="stat-label" style={{ fontSize: '11px' }}>{net.name}</span>
              <span className="stat-value" style={{ fontSize: '10px' }}>
                ↑{formatBytes(net.bytes_sent)} ↓{formatBytes(net.bytes_recv)}
              </span>
            </div>
          ))}
        </div>

        {/* Process Table */}
        <div className="dashboard-card process-table-container">
          <div className="dashboard-card-header">
            <div className="dashboard-card-title">
              <span>⚙️</span> Proses Berjalan
            </div>
            <div style={{ display: 'flex', gap: '8px' }}>
              {['cpu', 'memory', 'name'].map((sort) => (
                <button
                  key={sort}
                  className={`header-btn ${sortBy === sort ? 'active' : ''}`}
                  onClick={() => setSortBy(sort)}
                  style={{
                    background: sortBy === sort ? 'var(--accent-primary)' : undefined,
                    color: sortBy === sort ? 'white' : undefined,
                    borderColor: sortBy === sort ? 'var(--accent-primary)' : undefined,
                  }}
                >
                  {sort.toUpperCase()}
                </button>
              ))}
            </div>
          </div>

          {procLoading ? (
            <div className="loading-spinner"><div className="spinner" /></div>
          ) : (
            <table className="process-table">
              <thead>
                <tr>
                  <th>PID</th>
                  <th>Name</th>
                  <th>CPU %</th>
                  <th>Memory</th>
                  <th>User</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {processes.slice(0, 25).map((proc) => (
                  <tr key={proc.pid}>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '11px' }}>{proc.pid}</td>
                    <td>{proc.name}</td>
                    <td style={{ fontFamily: 'var(--font-mono)' }}>{proc.cpu_percent.toFixed(1)}%</td>
                    <td style={{ fontFamily: 'var(--font-mono)' }}>{proc.memory_mb.toFixed(1)} MB</td>
                    <td style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>{proc.username}</td>
                    <td>
                      <button
                        className="process-kill-btn"
                        onClick={() => {
                          if (window.confirm(`Kill process ${proc.name} (PID: ${proc.pid})?`)) {
                            killProcess(proc.pid)
                          }
                        }}
                      >
                        Kill
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  )
}

export default SystemDashboard
