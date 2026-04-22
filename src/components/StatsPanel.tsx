import React, { useState, useEffect } from 'react';

interface CostData {
  total_input_tokens: number;
  total_output_tokens: number;
  total_cost: number;
  total_requests: number;
  by_provider: Record<string, { input_tokens: number; output_tokens: number; requests: number; cost: number }>;
  updated_at: string;
}

interface IndexStatus {
  status: string;
  index?: { file_count: number; root_path: string; built_at: string };
}

const StatsPanel: React.FC = () => {
  const [stats, setStats] = useState<CostData | null>(null);
  const [indexStatus, setIndexStatus] = useState<IndexStatus | null>(null);
  const [indexBuilding, setIndexBuilding] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]);

  useEffect(() => {
    fetch('/api/stats').then(r => r.json()).then(d => { if (d.success) setStats(d.data); }).catch(() => { });
    fetch('/api/index/status').then(r => r.json()).then(d => { if (d.success) setIndexStatus(d); }).catch(() => { });
  }, []);

  const buildIndex = async () => {
    setIndexBuilding(true);
    try {
      await fetch('/api/index/build', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({}) });
      const d = await (await fetch('/api/index/status')).json();
      if (d.success) setIndexStatus(d);
    } catch { }
    setIndexBuilding(false);
  };

  const searchIndex = async () => {
    if (!searchQuery.trim()) return;
    const res = await fetch(`/api/index/search?q=${encodeURIComponent(searchQuery)}`);
    const data = await res.json();
    if (data.success) setSearchResults(data.data || []);
  };

  const resetStats = async () => {
    await fetch('/api/stats/reset', { method: 'POST' });
    setStats(null);
  };

  const fmt = (n: number) => n?.toLocaleString() ?? '0';

  return (
    <div className="panel" style={{ padding: '1.5rem', height: '100%', display: 'flex', flexDirection: 'column', gap: '1.2rem', overflowY: 'auto' }}>
      <h2 style={{ color: 'var(--text-primary)', margin: 0 }}>📊 Stats & Cost Tracker</h2>


      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '10px', padding: '1.2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
          <h3 style={{ color: 'var(--text-primary)', margin: 0 }}>Usage Overview</h3>
          <button onClick={resetStats} style={{ background: 'none', border: '1px solid var(--border)', color: 'var(--text-muted)', padding: '0.3rem 0.6rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem' }}>Reset</button>
        </div>
        {stats ? (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.8rem' }}>
            {[
              { label: 'Total Requests', value: fmt(stats.total_requests), icon: '🔗' },
              { label: 'Input Tokens', value: fmt(stats.total_input_tokens), icon: '📥' },
              { label: 'Output Tokens', value: fmt(stats.total_output_tokens), icon: '📤' },
              { label: 'Est. Cost', value: `$${(stats.total_cost || 0).toFixed(4)}`, icon: '💰' },
            ].map(item => (
              <div key={item.label} style={{ background: 'var(--bg-input)', borderRadius: '8px', padding: '0.8rem' }}>
                <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>{item.icon} {item.label}</div>
                <div style={{ color: 'var(--text-primary)', fontWeight: 'bold', fontSize: '1.2rem', marginTop: '0.2rem' }}>{item.value}</div>
              </div>
            ))}
          </div>
        ) : (
          <div style={{ color: 'var(--text-muted)' }}>No usage data yet.</div>
        )}
        {stats && Object.entries(stats.by_provider || {}).length > 0 && (
          <div style={{ marginTop: '1rem' }}>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginBottom: '0.5rem', textTransform: 'uppercase' }}>By Provider</div>
            {Object.entries(stats.by_provider).map(([p, s]) => (
              <div key={p} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.3rem 0', borderBottom: '1px solid var(--border)', fontSize: '0.85rem' }}>
                <span style={{ color: 'var(--text-primary)' }}>{p}</span>
                <span style={{ color: 'var(--text-muted)' }}>{fmt(s.requests)} req · ${(s.cost || 0).toFixed(4)}</span>
              </div>
            ))}
          </div>
        )}
      </div>


      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '10px', padding: '1.2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.8rem' }}>
          <h3 style={{ color: 'var(--text-primary)', margin: 0 }}>🗂️ Code Index</h3>
          <button onClick={buildIndex} disabled={indexBuilding} style={{ background: 'var(--accent)', border: 'none', color: '#fff', padding: '0.3rem 0.8rem', borderRadius: '6px', cursor: 'pointer', fontSize: '0.85rem' }}>
            {indexBuilding ? 'Building...' : 'Build Index'}
          </button>
        </div>
        {indexStatus?.index ? (
          <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            <div>📁 {indexStatus.index.file_count} files indexed</div>
            <div>📍 {indexStatus.index.root_path}</div>
            <div>🕒 {new Date(indexStatus.index.built_at).toLocaleString()}</div>
          </div>
        ) : (
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>No index built. Click "Build Index" to index your workspace.</div>
        )}
        <div style={{ marginTop: '0.8rem', display: 'flex', gap: '0.5rem' }}>
          <input
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && searchIndex()}
            placeholder="Search indexed files..."
            style={{ flex: 1, background: 'var(--bg-input)', border: '1px solid var(--border)', color: 'var(--text-primary)', padding: '0.4rem 0.8rem', borderRadius: '6px', fontSize: '0.85rem' }}
          />
          <button onClick={searchIndex} style={{ background: 'var(--bg-hover)', border: 'none', color: 'var(--text-primary)', padding: '0.4rem 0.8rem', borderRadius: '6px', cursor: 'pointer' }}>🔍</button>
        </div>
        {searchResults.length > 0 && (
          <div style={{ marginTop: '0.5rem', maxHeight: '150px', overflowY: 'auto' }}>
            {searchResults.map((f, i) => (
              <div key={i} style={{ padding: '0.2rem 0', borderBottom: '1px solid var(--border)', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                <span style={{ color: 'var(--accent)' }}>{f.path}</span> <span style={{ color: '#6272a4' }}>({f.language})</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default StatsPanel;
