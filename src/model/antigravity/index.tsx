import React, { useState, useEffect, useMemo, useRef } from 'react'

interface ModelQuota {
  name: string
  percentage: number
  available: boolean
  duration?: string
  color?: string
}

interface AntigravityConfig {
  selected_model: string
  temperature: number
  max_tokens: number
  thinking_budget?: number
  google_client_id: string
  google_client_secret: string
}

interface AntigravityAccount {
  email: string
  type: string
  status: string
  is_active?: boolean
  quotas: ModelQuota[]
  last_used: string
}

const AntigravityView: React.FC = () => {
  const [accounts, setAccounts] = useState<AntigravityAccount[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState<Record<string, boolean>>({})
  const [searchQuery, setSearchQuery] = useState('')
  const [filterType, setFilterType] = useState('ALL')
  const [showAllQuotas, setShowAllQuotas] = useState(true)
  const [selectedEmails, setSelectedEmails] = useState<Set<string>>(new Set())
  const [itemsPerPage, setItemsPerPage] = useState(10)
  const [currentPage, setCurrentPage] = useState(1)

  const fileInputRef = useRef<HTMLInputElement>(null)

  const [agConfig, setAgConfig] = useState<AntigravityConfig>({
    selected_model: '',
    temperature: 0.7,
    max_tokens: 8192,
    google_client_id: '',
    google_client_secret: ''
  })
  const [savingConfig, setSavingConfig] = useState(false)

  const fetchAccounts = async () => {
    try {
      const resp = await fetch('/api/antigravity/accounts')
      const data = await resp.json()
      if (data.success) setAccounts(data.data || [])
    } catch (err) {
      console.error('Failed to fetch accounts:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchConfig = async () => {
    try {
      const resp = await fetch('/api/antigravity/config')
      const data = await resp.json()
      if (data.success && data.data) setAgConfig(data.data)
    } catch { }
  }

  const fetchActiveModels = async () => {
    try {
      const resp = await fetch('/api/model/active')
      await resp.json()
    } catch { }
  }

  useEffect(() => {
    fetchAccounts()
    fetchConfig()
    fetchActiveModels()
  }, [])

  const updateModelActive = async (isActive: boolean) => {
    try {
      const resp = await fetch('/api/model/active')
      const data = await resp.json()
      const current = data.success ? data.data : {}
      await fetch('/api/model/active', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...current, antigravity: isActive })
      })
    } catch { }
  }

  const saveConfig = async (cfg: Partial<AntigravityConfig>) => {
    setSavingConfig(true)
    const newCfg = { ...agConfig, ...cfg }
    try {
      await fetch('/api/antigravity/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newCfg)
      })
      setAgConfig(newCfg)
    } catch {
    } finally {
      setSavingConfig(false)
    }
  }

  const handleRefresh = async (email: string) => {
    setRefreshing(prev => ({ ...prev, [email]: true }))
    try {
      const resp = await fetch(`/api/antigravity/refresh?email=${encodeURIComponent(email)}`)
      const data = await resp.json()
      if (data.success) setAccounts(prev => prev.map(acc => acc.email === email ? data.data : acc))
    } catch (err) {
      console.error('Refresh failed:', err)
    } finally {
      setRefreshing(prev => ({ ...prev, [email]: false }))
    }
  }

  const handleRefreshAll = async () => {
    for (const acc of filteredAccounts) {
      if (!refreshing[acc.email]) await handleRefresh(acc.email)
    }
  }

  const toggleActive = async (email: string) => {
    try {
      const acc = accounts.find(a => a.email === email)
      const currentlyActive = !!acc?.is_active

      const resp = await fetch(`/api/antigravity/active?email=${encodeURIComponent(email)}`, { method: 'POST' })
      const data = await resp.json()
      if (data.success) {
        fetchAccounts()
        updateModelActive(!currentlyActive)
      }
    } catch (err) {
      console.error('Toggle active failed:', err)
    }
  }

  const handleDelete = async (email: string) => {
    if (!window.confirm(`Delete ${email}?`)) return
    try {
      const resp = await fetch(`/api/antigravity/accounts?email=${encodeURIComponent(email)}`, { method: 'DELETE' })
      const data = await resp.json()
      if (data.success) {
        setAccounts(prev => prev.filter(a => a.email !== email))
        setSelectedEmails(prev => { const n = new Set(prev); n.delete(email); return n })
      }
    } catch (err) { console.error('Delete failed:', err) }
  }

  const handleImport = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = async (evt) => {
      try {
        const parsed = JSON.parse(evt.target?.result as string)
        if (Array.isArray(parsed)) {
          let count = 0
          for (const item of parsed) {
            if (item.email && item.refresh_token) {
              await fetch('/api/antigravity/accounts', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(item) })
              count++
            }
          }
          alert(`Imported ${count} accounts.`)
          await fetchAccounts()
        }
      } catch { alert('Invalid JSON file.') }
    }
    reader.readAsText(file)
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const handleExport = () => {
    const data = accounts.map(a => ({ email: a.email, refresh_token: (a as any).refresh_token || '', type: a.type }))
    const url = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(data, null, 2))
    const a = document.createElement('a')
    a.href = url; a.download = 'antigravity_accounts.json'
    document.body.appendChild(a); a.click(); a.remove()
  }

  const filteredAccounts = useMemo(() =>
    accounts.filter(acc => {
      if (filterType !== 'ALL' && acc.type !== filterType) return false
      if (searchQuery && !acc.email.toLowerCase().includes(searchQuery.toLowerCase())) return false
      return true
    }), [accounts, filterType, searchQuery])

  const totalPages = Math.ceil(filteredAccounts.length / itemsPerPage) || 1
  const paginatedAccounts = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage
    return filteredAccounts.slice(start, start + itemsPerPage)
  }, [filteredAccounts, currentPage, itemsPerPage])

  const counts = useMemo(() =>
    accounts.reduce((acc, curr) => {
      acc.ALL = (acc.ALL || 0) + 1
      acc[curr.type] = (acc[curr.type] || 0) + 1
      return acc
    }, { ALL: 0 } as Record<string, number>), [accounts])

  const isRefreshingAny = Object.values(refreshing).some(Boolean)
  const activeAccount = accounts.find(a => a.is_active)

  const formatResetTime = (isoStr?: string) => {
    if (!isoStr) return ''
    try {
      const d = new Date(isoStr)
      const now = new Date()
      const diffMs = d.getTime() - now.getTime()
      if (diffMs < 0) return 'Expired'
      const diffH = Math.floor(diffMs / 3600000)
      const diffD = Math.floor(diffH / 24)
      if (diffD > 0) return `${diffD}d ${diffH % 24}h`
      if (diffH > 0) return `${diffH}h`
      const diffM = Math.floor(diffMs / 60000)
      return `${diffM}m`
    } catch { return '' }
  }

  const deduplicateQuotas = (quotas: ModelQuota[]) => {
    const map = new Map<string, ModelQuota>()
    for (const q of (quotas || [])) {
      const existing = map.get(q.name)
      if (!existing || q.percentage > existing.percentage) {
        map.set(q.name, q)
      }
    }
    return Array.from(map.values())
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '80px' }}>
        <div className="typing-dots"><span></span><span></span><span></span></div>
      </div>
    )
  }

  return (
    <div className="antigravity-dashboard animate-fade">


      <div style={{
        background: activeAccount ? 'linear-gradient(135deg, rgba(16,185,129,0.1) 0%, rgba(124,58,237,0.08) 100%)' : 'rgba(255,255,255,0.03)',
        border: activeAccount ? '1px solid rgba(16,185,129,0.25)' : '1px solid rgba(255,255,255,0.08)',
        borderRadius: '12px',
        padding: '14px 22px',
        display: 'flex',
        alignItems: 'center',
        gap: '15px',
        marginBottom: '16px',
        boxShadow: activeAccount ? '0 0 25px rgba(16,185,129,0.05)' : 'none',
        transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        position: 'relative',
        overflow: 'hidden'
      }}>
        {activeAccount && (
          <div style={{
            position: 'absolute', top: 0, left: 0, width: '4px', height: '100%', background: '#10b981'
          }}></div>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
          <div style={{
            width: '10px', height: '10px', borderRadius: '50%',
            background: activeAccount ? '#10b981' : '#4b5563',
            boxShadow: activeAccount ? '0 0 12px #10b981, 0 0 20px rgba(16,185,129,0.4)' : 'none',
            flexShrink: 0,
            transition: 'all 0.3s ease'
          }}></div>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <strong style={{
              fontSize: '15px',
              letterSpacing: '0.2px',
              color: activeAccount ? '#fff' : 'rgba(255,255,255,0.6)',
              transition: 'all 0.3s ease'
            }}>{activeAccount ? 'AGENT ENGINE ACTIVE' : 'AGENT ENGINE STANDBY'}</strong>
            {activeAccount ? (
              <span style={{ fontSize: '11px', color: '#10b981', fontWeight: 600, marginTop: '2px' }}>
                Connected via {activeAccount.email}
              </span>
            ) : (
              <span style={{ fontSize: '11px', color: 'rgba(255,255,255,0.4)', marginTop: '2px' }}>
                No active instance detected
              </span>
            )}
          </div>
        </div>

        <div
          onClick={() => {
            if (activeAccount) {
              toggleActive(activeAccount.email)
            } else if (accounts.length > 0) {
              toggleActive(accounts[0].email)
            }
          }}
          style={{
            width: '48px',
            height: '26px',
            borderRadius: '13px',
            background: activeAccount ? '#10b981' : 'rgba(255,255,255,0.1)',
            border: activeAccount ? '1px solid rgba(16,185,129,0.5)' : '1px solid rgba(255,255,255,0.1)',
            position: 'relative',
            cursor: 'pointer',
            transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
            boxShadow: activeAccount ? '0 0 15px rgba(16,185,129,0.3)' : 'none'
          }}
        >
          <div style={{
            width: '18px',
            height: '18px',
            borderRadius: '50%',
            background: '#fff',
            position: 'absolute',
            top: '3px',
            left: activeAccount ? '25px' : '4px',
            transition: 'all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
            boxShadow: '0 2px 5px rgba(0,0,0,0.4)'
          }}></div>
        </div>
      </div>


      <div className="dashboard-top-bar">
        <div className="search-box">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.5"><circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" /></svg>
          <input
            type="text"
            placeholder="Search email..."
            value={searchQuery}
            onChange={e => { setSearchQuery(e.target.value); setCurrentPage(1) }}
          />
        </div>

        <div className="filter-group">
          {(['ALL', 'PRO', 'ULTRA', 'FREE'] as const).map(t => (
            <button key={t} className={filterType === t ? 'active' : ''} onClick={() => { setFilterType(t); setCurrentPage(1) }} style={{ padding: '6px 10px' }}>
              {t} <span style={{ fontSize: '10px' }}>{counts[t] || 0}</span>
            </button>
          ))}
        </div>

        <div style={{ padding: '0 8px', height: '32px', display: 'flex', alignItems: 'center', background: 'rgba(124,58,237,0.05)', borderRadius: '8px', border: '1px solid rgba(124,58,237,0.15)', flexShrink: 0 }}>
          <span style={{ fontSize: '10px', color: '#a78bfa', fontWeight: 700, marginRight: '6px' }}>Model:</span>
          <select
            value={agConfig.selected_model}
            onChange={(e) => saveConfig({ selected_model: e.target.value })}
            style={{
              background: 'none', border: 'none', color: '#fff', fontSize: '11px', fontWeight: 700, outline: 'none', cursor: 'pointer', maxWidth: '100px'
            }}
          >
            <option value="" style={{ background: '#0d0920' }}>Auto Rotation</option>
            {Array.from(new Set(accounts.flatMap(a => a.quotas || []).map(q => q.name))).sort().map(name => {
              const q = accounts.flatMap(a => a.quotas || []).find(curr => curr.name === name)
              return <option key={name} value={(q as any).key || ''} style={{ background: '#0d0920' }}>{name}</option>
            })}
          </select>
        </div>

        <div style={{ padding: '0 8px', height: '32px', display: 'flex', alignItems: 'center', background: 'rgba(16,185,129,0.05)', borderRadius: '8px', border: '1px solid rgba(16,185,129,0.15)', flexShrink: 0 }}>
          <span style={{ fontSize: '10px', color: '#10b981', fontWeight: 700, marginRight: '6px' }}>Think:</span>
          <input
            type="range" min="0" max="32000" step="1000"
            value={agConfig.thinking_budget || 0}
            onChange={(e) => saveConfig({ thinking_budget: parseInt(e.target.value) })}
            style={{ width: '50px', height: '4px', cursor: 'pointer', accentColor: '#10b981' }}
          />
          <span style={{ fontSize: '10px', color: '#fff', marginLeft: '6px', width: '22px' }}>{Math.floor((agConfig.thinking_budget || 0) / 1000)}k</span>
        </div>

        <div style={{ padding: '0 8px', height: '32px', display: 'flex', alignItems: 'center', background: 'rgba(59,130,246,0.05)', borderRadius: '8px', border: '1px solid rgba(59,130,246,0.15)', flexShrink: 0 }}>
          <span style={{ fontSize: '10px', color: '#60a5fa', fontWeight: 700, marginRight: '6px' }}>Temp:</span>
          <input
            type="range" min="0" max="1" step="0.1"
            value={agConfig.temperature || 0.7}
            onChange={(e) => saveConfig({ temperature: parseFloat(e.target.value) })}
            style={{ width: '40px', height: '4px', cursor: 'pointer', accentColor: '#60a5fa' }}
          />
          <span style={{ fontSize: '10px', color: '#fff', marginLeft: '6px' }}>{agConfig.temperature || 0.7}</span>
        </div>

        {savingConfig && <div style={{ marginLeft: '4px' }} className="typing-dots small"><span /></div>}

        <div className="action-row">
          <button
            className="btn-add"
            onClick={() => { window.location.href = `http://${window.location.host}/api/antigravity/oauth/start` }}
            style={{ background: 'linear-gradient(135deg, #5b21b6, #7c3aed)', color: '#fff', border: 'none', borderRadius: '8px', padding: '7px 12px', fontSize: '12px', fontWeight: 700, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '5px', boxShadow: '0 0 14px rgba(124,58,237,0.3)', flexShrink: 0 }}
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3.5"><path d="M12 5v14M5 12h14" /></svg>
            Add Account
          </button>

          <button
            className="btn-refresh"
            onClick={handleRefreshAll}
            disabled={isRefreshingAny || filteredAccounts.length === 0}
            style={{ background: 'rgba(124,58,237,0.1)', color: '#c4b5fd', border: '1px solid rgba(124,58,237,0.3)', borderRadius: '8px', padding: '7px 12px', fontSize: '12px', fontWeight: 700, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '5px', transition: 'all 0.2s', flexShrink: 0 }}
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ animation: isRefreshingAny ? 'spin 1s linear infinite' : 'none' }}>
              <path d="M4 4h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            {isRefreshingAny ? '...' : 'Refresh'}
          </button>

          <button
            className="btn-warmup"
            onClick={handleRefreshAll}
            style={{ background: 'rgba(124,58,237,0.1)', color: '#c4b5fd', border: '1px solid rgba(124,58,237,0.3)', borderRadius: '8px', padding: '7px 12px', fontSize: '12px', fontWeight: 700, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '5px', transition: 'all 0.2s', flexShrink: 0 }}
          >
            <span>🔥</span> Warmup
          </button>

          <div className="toggle-group" onClick={() => setShowAllQuotas(!showAllQuotas)} style={{ cursor: 'pointer', gap: '6px' }}>
            <span style={{ fontSize: '11px' }}>Quotas</span>
            <div style={{
              width: '32px', height: '18px', borderRadius: '9px',
              background: showAllQuotas ? '#7c3aed' : '#374151',
              position: 'relative', transition: 'all 0.2s', flexShrink: 0
            }}>
              <div style={{
                width: '14px', height: '14px', borderRadius: '50%', background: 'white',
                position: 'absolute', top: '2px',
                left: showAllQuotas ? '16px' : '2px',
                transition: 'all 0.2s', boxShadow: '0 1px 3px rgba(0,0,0,0.3)'
              }}></div>
            </div>
          </div>

          <input type="file" ref={fileInputRef} onChange={handleImport} accept=".json" style={{ display: 'none' }} />
          <button className="btn-icon" onClick={() => fileInputRef.current?.click()} title="Import" style={{ padding: '8px' }}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path strokeLinecap="round" strokeLinejoin="round" d="M4 16a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4" /></svg>
          </button>
          <button className="btn-icon" onClick={handleExport} title="Export" style={{ padding: '8px' }}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path strokeLinecap="round" strokeLinejoin="round" d="M4 16a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4" /></svg>
          </button>


        </div>
      </div>


      <div className="dashboard-table-header" style={{ gridTemplateColumns: '50px 1fr 1.8fr 130px 80px 90px' }}>
        <div className="col-check">
          <input
            type="checkbox"
            checked={selectedEmails.size === filteredAccounts.length && filteredAccounts.length > 0}
            onChange={() => {
              if (selectedEmails.size === filteredAccounts.length && filteredAccounts.length > 0) setSelectedEmails(new Set())
              else setSelectedEmails(new Set(filteredAccounts.map(a => a.email)))
            }}
          />
        </div>
        <div>ACCOUNT</div>
        <div>MODEL QUOTA</div>
        <div>LAST USED</div>
        <div style={{ textAlign: 'center' }}>ACTIVE</div>
        <div style={{ textAlign: 'center' }}>ACTIONS</div>
      </div>


      <div style={{ minHeight: '300px' }}>
        {paginatedAccounts.length === 0 ? (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '60px 20px', color: 'var(--text-dim)', textAlign: 'center', gap: '12px' }}>
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" style={{ opacity: 0.3 }}>
              <path d="M20 13a2 2 0 00-2-2H6a2 2 0 00-2 2m16 0a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <p>{searchQuery || filterType !== 'ALL' ? 'No accounts match your filters.' : 'No accounts found. Click "+ Add Account" to get started.'}</p>
          </div>
        ) : (
          paginatedAccounts.map((acc, i) => {
            return (
              <div
                key={i}
                className="account-row"
                style={{
                  gridTemplateColumns: '50px 1fr 1.8fr 130px 80px 90px',
                  opacity: refreshing[acc.email] ? 0.5 : 1,
                  background: selectedEmails.has(acc.email) ? 'rgba(124,58,237,0.07)' : undefined,
                  borderLeft: acc.is_active ? '2px solid #10b981' : '2px solid transparent',
                }}
              >

                <div className="col-check">
                  <div className="drag-handle">⠿</div>
                  <input type="checkbox" checked={selectedEmails.has(acc.email)} onChange={() => {
                    setSelectedEmails(prev => { const n = new Set(prev); n.has(acc.email) ? n.delete(acc.email) : n.add(acc.email); return n })
                  }} />
                </div>


                <div className="col-email">
                  <div className="email-text" title={acc.email} style={{ color: acc.is_active ? '#a78bfa' : '#60a5fa' }}>
                    {acc.email}
                  </div>
                  <div className="badges">
                    <span className="badge-outline">{acc.type || 'FREE'}</span>
                    {acc.status === 'CURRENT' && <span className="badge">CURRENT</span>}
                    {acc.is_active && (
                      <span style={{ background: 'rgba(16,185,129,0.15)', border: '1px solid rgba(16,185,129,0.4)', color: '#6ee7b7', fontSize: '9px', padding: '1px 6px', borderRadius: '4px', fontWeight: 800 }}>
                        AGENT
                      </span>
                    )}
                  </div>
                </div>


                <div className="col-quota">
                  {(() => {
                    const allQuotas = deduplicateQuotas(showAllQuotas ? acc.quotas : acc.quotas?.filter(q => q.available !== false))
                    if (!allQuotas || allQuotas.length === 0) {
                      return <div style={{ color: 'var(--text-dim)', fontSize: '12px', fontStyle: 'italic' }}>No quota data. Click refresh.</div>
                    }
                    return (
                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '5px' }}>
                        {allQuotas.map((m, j) => (
                          <div key={j} style={{ background: '#0d0920', padding: '5px 7px', borderRadius: '6px', border: '1px solid rgba(124,58,237,0.12)' }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '3px' }}>
                              <span style={{ display: 'flex', alignItems: 'center', gap: '4px', overflow: 'hidden' }}>
                                <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: m.color || '#8B5CF6', boxShadow: `0 0 5px ${m.color || '#8B5CF6'}`, flexShrink: 0 }}></span>
                                <span style={{ fontSize: '10px', color: '#94a3b8', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: '120px' }} title={m.name}>
                                  {m.name}
                                </span>
                              </span>
                              <span style={{ fontSize: '10px', color: m.color || '#7c3aed', fontWeight: 800, flexShrink: 0 }}>{m.percentage}%</span>
                            </div>
                            <div style={{ height: '4px', background: 'rgba(0,0,0,0.5)', borderRadius: '2px', overflow: 'hidden' }}>
                              <div style={{
                                width: `${Math.max(2, m.percentage)}%`,
                                height: '100%',
                                borderRadius: '2px',
                                background: `linear-gradient(90deg, ${m.color || '#6D28D9'}, ${m.color || '#8B5CF6'})`,
                                boxShadow: `0 0 6px ${m.color || '#8B5CF6'}88`,
                                transition: 'width 1s ease',
                              }}></div>
                            </div>
                            {m.duration && (
                              <div style={{ fontSize: '8px', color: '#475569', marginTop: '2px', textAlign: 'right' }}>
                                ↺ {formatResetTime(m.duration)}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )
                  })()}
                </div>


                <div className="col-last">
                  <div className="last-date">
                    {acc.last_used && acc.last_used !== '0001-01-01T00:00:00Z'
                      ? new Date(acc.last_used).toLocaleDateString()
                      : 'Never used'}
                  </div>
                  <div className="last-time">
                    {acc.last_used && acc.last_used !== '0001-01-01T00:00:00Z'
                      ? new Date(acc.last_used).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                      : '--:--'}
                  </div>
                </div>


                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <button
                    onClick={() => toggleActive(acc.email)}
                    title={acc.is_active ? 'Deactivate agent' : 'Activate agent'}
                    style={{
                      width: '40px', height: '22px', borderRadius: '11px', border: 'none', cursor: 'pointer',
                      background: acc.is_active ? '#10b981' : '#2d1b44',
                      position: 'relative', transition: 'all 0.2s',
                      boxShadow: acc.is_active ? '0 0 10px rgba(16,185,129,0.4)' : 'none',
                      flexShrink: 0,
                    }}
                  >
                    <div style={{
                      width: '16px', height: '16px', borderRadius: '50%', background: '#fff',
                      position: 'absolute', top: '3px',
                      left: acc.is_active ? '21px' : '3px',
                      transition: 'all 0.2s', boxShadow: '0 1px 3px rgba(0,0,0,0.4)'
                    }}></div>
                  </button>
                </div>


                <div className="col-actions">
                  <button
                    onClick={() => handleRefresh(acc.email)}
                    disabled={refreshing[acc.email]}
                    title="Refresh quotas"
                    style={{ color: refreshing[acc.email] ? '#4b5563' : '#9ca3af' }}
                  >
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ animation: refreshing[acc.email] ? 'spin 1s linear infinite' : 'none' }}>
                      <path d="M4 4h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                  </button>
                  <button onClick={() => handleDelete(acc.email)} style={{ color: '#ef4444' }} title="Delete">
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4m4-6m1-10a1 1 0 00-1-1h-4a1 1 0 00-1 1M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            )
          })
        )}
      </div>


      <div className="dashboard-footer">
        <div className="footer-left">
          Showing <strong style={{ color: '#a78bfa', margin: '0 3px' }}>{paginatedAccounts.length > 0 ? (currentPage - 1) * itemsPerPage + 1 : 0}</strong>
          –<strong style={{ color: '#a78bfa', margin: '0 3px' }}>{Math.min(currentPage * itemsPerPage, filteredAccounts.length)}</strong>
          of <strong style={{ color: '#a78bfa', margin: '0 3px' }}>{filteredAccounts.length}</strong>
          <div style={{ width: '1px', height: '14px', background: 'var(--border-subtle)', margin: '0 8px' }}></div>
          Rows:
          <select value={itemsPerPage} onChange={e => { setItemsPerPage(Number(e.target.value)); setCurrentPage(1) }}>
            <option value="5">5</option>
            <option value="10">10</option>
            <option value="20">20</option>
            <option value="50">50</option>
          </select>
        </div>
        <div className="pagination">
          <button onClick={() => setCurrentPage(p => Math.max(1, p - 1))} disabled={currentPage === 1}>&lt;</button>
          {[...Array(Math.min(totalPages, 7))].map((_, idx) => (
            <button key={idx} onClick={() => setCurrentPage(idx + 1)} className={currentPage === idx + 1 ? 'active' : ''}>{idx + 1}</button>
          ))}
          <button onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))} disabled={currentPage === totalPages}>&gt;</button>
        </div>
      </div>

    </div>
  )
}

export default AntigravityView




