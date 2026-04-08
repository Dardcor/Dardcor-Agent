import React from 'react'

const AntigravityView: React.FC = () => {
  const accounts = [
    { email: 'afdanikomalik@gmail.com', type: 'FREE', status: '', models: [
      { name: 'GPT-05S 120B (Medium)', duration: '6d 23h', pct: 100, color: '#10b981' },
      { name: 'Gemini 3 Pro (High)', duration: '6d 10h', pct: 100, color: '#3b82f6' },
      { name: 'Gemini 3.1 Pro (High)', duration: '6d 10h', pct: 100, color: '#f59e0b' },
      { name: 'Gemini 3 Flash', duration: '6d 11h', pct: 100, color: '#8b5cf6' }
    ]},
    { email: 'dardcorxyz@gmail.com', type: 'FREE', status: 'CURRENT', models: [
      { name: 'GPT-05S 120B (Medium)', duration: '4d 5h', pct: 80, color: '#10b981' },
      { name: 'Gemini 3.1 Pro (High)', duration: '4d 8h', pct: 60, color: '#f59e0b' },
      { name: 'Gemini 3 Pro (Low)', duration: '4d 8h', pct: 60, color: '#3b82f6' },
      { name: 'Gemini 3 Flash', duration: '4d 11h', pct: 80, color: '#8b5cf6' }
    ]}
  ]

  return (
    <div className="antigravity-dashboard animate-fade">
      <div className="dashboard-top-bar">
        <div className="search-box">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
          <input type="text" placeholder="Search email..." />
        </div>
        
        <div className="filter-group">
          <button className="active">All <span>15</span></button>
          <button>PRO <span>0</span></button>
          <button>ULTRA <span>0</span></button>
          <button>FREE <span>15</span></button>
        </div>

        <div className="action-row">
          <button className="btn-refresh">🔄 Refresh All</button>
          <button className="btn-warmup">🔥 One-click Warmup</button>
          <div className="toggle-group">
             <span>Show All Quotas</span>
             <div className="toggle-switch"></div>
          </div>
          <button className="btn-icon">📤 Import</button>
          <button className="btn-icon">📥 Export</button>
        </div>
      </div>

      <div className="dashboard-table-header">
        <div className="col-check"><input type="checkbox" /></div>
        <div className="col-email">EMAIL</div>
        <div className="col-quota">MODEL QUOTA</div>
        <div className="col-last">LAST USED</div>
        <div className="col-actions">ACTIONS</div>
      </div>

      <div className="dashboard-table-body">
        {accounts.map((acc, i) => (
          <div key={i} className="account-row shadow-premium">
            <div className="col-check">
               <div className="drag-handle">⠿</div>
               <input type="checkbox" />
            </div>
            <div className="col-email">
              <div className="email-text">{acc.email}</div>
              <div className="badges">
                {acc.status === 'CURRENT' && <span className="badge current">CURRENT</span>}
                <span className="badge-outline">{acc.type}</span>
              </div>
            </div>
            <div className="col-quota">
              <div className="quota-grid">
                {acc.models.map((m, j) => (
                  <div key={j} className="quota-item">
                    <div className="quota-info">
                      <span className="model-name"><span className="dot" style={{ background: m.color }}></span> {m.name}</span>
                      <span className="duration">🕒 {m.duration}</span>
                    </div>
                    <div className="quota-bar">
                      <div className="fill" style={{ width: `${m.pct}%`, background: m.color }}></div>
                      <span className="pct">{m.pct}%</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <div className="col-last">
              <div className="last-date">08/04/2026</div>
              <div className="last-time">13.31</div>
            </div>
            <div className="col-actions">
              <button title="Info">ⓘ</button>
              <button title="Settings">⚙</button>
              <button title="Refresh">🔄</button>
              <button title="Delete">🗑</button>
            </div>
          </div>
        ))}
      </div>
      
      <div className="dashboard-footer">
        <div className="footer-left">
           Showing 1 to 10 of 15 entries
           <span className="spacer"></span>
           Per page 
           <select defaultValue="10 items"><option>10 items</option></select>
        </div>
        <div className="pagination">
           <button>&lt;</button>
           <button className="active">1</button>
           <button>2</button>
           <button>&gt;</button>
        </div>
      </div>
    </div>
  )
}

export default AntigravityView
