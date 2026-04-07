import React, { useState, useEffect } from 'react'
import wsService from '../services/websocket'

interface SystemInfo {
  os: string
  arch: string
  hostname: string
  cpu: string
  memory: {
    total: number
    used: number
    free: number
  }
  uptime: string
  platform: string
}

const SystemDashboard: React.FC = () => {
  const [sysInfo, setSysInfo] = useState<SystemInfo | null>(null)
  const [processes, setProcesses] = useState<any[]>([])
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    const unsub = wsService.on('system_info_result', (msg: any) => {
      setSysInfo(msg.payload)
      setIsLoading(false)
    })

    const unsubProc = wsService.on('processes_result', (msg: any) => {
      setProcesses(msg.payload.processes)
      setIsLoading(false)
    })

    fetchData()

    const interval = setInterval(fetchData, 10000)

    return () => {
      unsub()
      unsubProc()
      clearInterval(interval)
    }
  }, [])

  const fetchData = () => {
    setIsLoading(true)
    wsService.send('get_system_info', {})
    wsService.send('get_processes', {})
  }

  const handleKill = (pid: number) => {
    if (window.confirm(`Kill process ${pid}?`)) {
      wsService.send('kill_process', { pid })
      fetchData()
    }
  }

  return (
    <div className="system-dashboard">
      <div className="dashboard-header">
        <div className="dashboard-title">
          <span>📊</span> SYSTEM HEALTH
        </div>
        <div className="dashboard-actions">
          <button onClick={fetchData}>Refresh Now</button>
        </div>
      </div>

      <div className="dashboard-grid">
        <div className="info-card">
          <h4>Host Info</h4>
          <div className="info-row"><span>Host</span><span>{sysInfo?.hostname}</span></div>
          <div className="info-row"><span>OS</span><span>{sysInfo?.os} ({sysInfo?.arch})</span></div>
          <div className="info-row"><span>Uptime</span><span>{sysInfo?.uptime}</span></div>
        </div>

        <div className="info-card">
          <h4>Memory</h4>
          <div className="info-row"><span>Total</span><span>{sysInfo?.memory ? (sysInfo.memory.total / (1024 ** 3)).toFixed(1) : '--'} GB</span></div>
          <div className="info-row"><span>Used</span><span>{sysInfo?.memory ? (sysInfo.memory.used / (1024 ** 3)).toFixed(1) : '--'} GB</span></div>
          <div className="info-row"><span>Free</span><span>{sysInfo?.memory ? (sysInfo.memory.free / (1024 ** 3)).toFixed(1) : '--'} GB</span></div>
          <div className="progress-bar">
            <div 
              className="progress-fill" 
              style={{ width: `${sysInfo ? (sysInfo.memory.used / sysInfo.memory.total * 100) : 0}%` }}
            ></div>
          </div>
        </div>
      </div>

      <div className="process-list card">
        <div className="card-header">
          <h4>Running Processes</h4>
          <span className="count-badge">{processes.length} total</span>
        </div>
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>PID</th>
                <th>Name</th>
                <th>CPU %</th>
                <th>MEM %</th>
                <th>Status</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {processes.slice(0, 50).map((p, i) => (
                <tr key={i}>
                  <td>{p.pid}</td>
                  <td>{p.name}</td>
                  <td>{p.cpu?.toFixed(1) || '0.0'}%</td>
                  <td>{p.memory?.toFixed(1) || '0.0'}%</td>
                  <td>{p.status}</td>
                  <td>
                    <button className="kill-btn" onClick={() => handleKill(p.pid)}>Kill</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default SystemDashboard
