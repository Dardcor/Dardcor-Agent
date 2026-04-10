import React, { useState, useEffect } from 'react'

interface SystemInfo {
  os: { platform: string; arch: string; edition: string; family: string }
  cpu: { model_name: string; cores: number; threads: number; usage_percent: number; frequency_mhz: number }
  memory: { total: number; used: number; free: number; used_percent: number }
  disks: Array<{ device: string; mount_point: string; total: number; used: number; used_percent: number }>
  network: Array<{ name: string; addresses: string[]; bytes_sent: number; bytes_recv: number }>
  bios: { gpu?: string; motherboard?: string; kernel_edition?: string }
  uptime: number
  hostname: string
}

interface ProcessInfo {
  pid: number
  name: string
  cpu_percent: number
  mem_percent: number
  memory_mb: number
  username: string
  status: string
}

const SystemDashboard: React.FC = () => {
  const [sysInfo, setSysInfo] = useState<SystemInfo | null>(null)
  const [processes, setProcesses] = useState<ProcessInfo[]>([])

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 1000)
    return () => clearInterval(interval)
  }, [])

  const fetchData = async () => {
    try {
      const [sysRes, procRes] = await Promise.all([
        fetch('/api/system'),
        fetch('/api/system/processes?limit=60')
      ])
      const sysData = await sysRes.json()
      const procData = await procRes.json()
      if (sysData.success) setSysInfo(sysData.data)
      if (procData.success) setProcesses(procData.data || [])
    } catch (e) {
      console.error('System data fetch error:', e)
    }
  }

  const formatUptime = (sec: number) => {
    const h = Math.floor(sec / 3600)
    const m = Math.floor((sec % 3600) / 60)
    const s = sec % 60
    return `${h}h ${m}m ${s}s`
  }

  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  const formatOS = () => {
    if (!sysInfo) return '---'
    if (sysInfo.os.family && sysInfo.os.edition) {
       return `${sysInfo.os.family} ${sysInfo.os.edition} (${sysInfo.os.arch})`
    }
    return `${sysInfo.os.platform} (${sysInfo.os.arch})`
  }

  return (
    <div className="system-dashboard">
      <div className="dashboard-grid">
        <div className="dashboard-card">
          <div className="card-header">
             <span className="card-icon">🏗️</span>
             <div className="card-title-group">
                <h3>Hardware & Info</h3>
                <span className="card-subtitle">{sysInfo?.hostname}</span>
             </div>
          </div>
          <div className="card-body">
             <div className="data-row"><span>OS</span><b>{formatOS()}</b></div>
             <div className="data-row"><span>GPU</span><b style={{ color: '#a78bfa' }}>{sysInfo?.bios?.gpu || 'Integrated'}</b></div>
             <div className="data-row"><span>Board</span><b>{sysInfo?.bios?.motherboard || 'Generic'}</b></div>
             <div className="data-row"><span>Kernel</span><b>{sysInfo?.bios?.kernel_edition || '---'}</b></div>
             <div className="data-row"><span>Uptime</span><b style={{ color: '#4ade80' }}>{sysInfo ? formatUptime(sysInfo.uptime) : '---'}</b></div>
          </div>
        </div>

        <div className="dashboard-card highlight">
          <div className="card-header">
             <span className="card-icon">⚡</span>
             <div className="card-title-group">
                <h3>Vitals & Resources</h3>
                <span className="card-subtitle">{sysInfo?.cpu.model_name}</span>
             </div>
          </div>
          <div className="card-body">
             <div className="resource-stat">
                <div className="stat-info"><span>CPU Usage ({sysInfo?.cpu.cores} Cores)</span><b>{sysInfo?.cpu.usage_percent.toFixed(1)}%</b></div>
                <div className="mini-progress"><div className="fill" style={{ width: `${sysInfo?.cpu.usage_percent}%` }}></div></div>
             </div>
             <div className="resource-stat">
                <div className="stat-info"><span>RAM Usage (Free: {sysInfo ? formatSize(sysInfo.memory.free) : '0 GB'})</span><b>{sysInfo?.memory.used_percent.toFixed(1)}%</b></div>
                <div className="mini-progress"><div className="fill purple" style={{ width: `${sysInfo?.memory.used_percent}%` }}></div></div>
             </div>
             <div className="data-grid-2" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px', marginTop: '10px' }}>
                <div className="data-row small" style={{ border: 'none', padding: 0 }}><span>Total RAM</span><b>{sysInfo ? formatSize(sysInfo.memory.total) : '0 GB'}</b></div>
                <div className="data-row small" style={{ border: 'none', padding: 0 }}><span>Freq</span><b>{sysInfo?.cpu.frequency_mhz} MHz</b></div>
             </div>
          </div>
        </div>

        <div className="dashboard-card">
          <div className="card-header">
             <span className="card-icon">💽</span>
             <div className="card-title-group">
                <h3>Storage</h3>
                <span className="card-subtitle">Active Volumes</span>
             </div>
          </div>
          <div className="card-body scrollable" style={{ maxHeight: '200px', overflowY: 'auto', paddingRight: '4px' }}>
             {sysInfo?.disks.map((d, i) => (
                <div key={i} className="disk-item">
                   <div className="disk-meta">
                     <span style={{ fontWeight: 600 }}>{d.mount_point}</span>
                     <div><span style={{ color: 'var(--text-dim)', marginRight: '6px' }}>{formatSize(d.used)} / {formatSize(d.total)}</span> <b>{d.used_percent.toFixed(0)}%</b></div>
                   </div>
                   <div className="mini-progress flat"><div className="fill orange" style={{ width: `${d.used_percent}%` }}></div></div>
                </div>
             ))}
          </div>
        </div>

        <div className="dashboard-card">
          <div className="card-header">
             <span className="card-icon">📡</span>
             <div className="card-title-group">
                <h3>Network</h3>
                <span className="card-subtitle">I/O Activity</span>
             </div>
          </div>
          <div className="card-body scrollable" style={{ maxHeight: '200px', overflowY: 'auto', paddingRight: '4px' }}>
             {sysInfo?.network.filter(n => n.addresses.length > 0).map((n, i) => (
                <div key={i} className="net-item" style={{ padding: '10px 12px', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                   <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                     <div className="net-name" style={{ marginBottom: 0, fontWeight: 700 }}>{n.name}</div>
                     <div style={{ fontSize: '10px', color: '#60a5fa' }}>{formatSize(n.bytes_recv)} ↓ {formatSize(n.bytes_sent)} ↑</div>
                   </div>
                   <div className="net-addrs" style={{ fontSize: '11px', opacity: 0.8 }}>{n.addresses[0]}</div>
                </div>
             ))}
          </div>
        </div>

        <div className="dashboard-card span-full">
           <div className="card-header">
              <span className="card-icon">📟</span>
              <div className="card-title-group">
                 <h3>Process Analytics</h3>
                 <span className="card-subtitle">Top usage metrics</span>
              </div>
           </div>
           <div className="table-wrap">
              <table className="proc-table">
                 <thead>
                    <tr><th>PID</th><th>Name</th><th>User</th><th>Status</th><th>CPU %</th><th>MEM %</th><th>RAM (MB)</th></tr>
                 </thead>
                 <tbody>
                    {processes.slice(0, 40).map((p, i) => (
                       <tr key={i}>
                          <td>{p.pid}</td>
                          <td className="name-cell">{p.name || '---'}</td>
                          <td style={{ fontSize: '11px' }}>{p.username || 'system'}</td>
                          <td style={{ fontSize: '11px', textTransform: 'capitalize' }}>{p.status || 'running'}</td>
                          <td style={{ color: p.cpu_percent > 10 ? '#f87171' : 'inherit' }}>{p.cpu_percent.toFixed(1)}</td>
                          <td style={{ color: p.mem_percent > 5 ? '#f87171' : 'inherit' }}>{p.mem_percent.toFixed(1)}</td>
                          <td>{p.memory_mb ? p.memory_mb.toFixed(1) : '---'}</td>
                       </tr>
                    ))}
                 </tbody>
              </table>
           </div>
        </div>
      </div>
    </div>
  )
}

export default SystemDashboard




