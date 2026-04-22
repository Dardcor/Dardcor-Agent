import React, { useState, useEffect } from 'react';

interface Conversation {
  id: string;
  title: string;
  created_at: string;
  updated_at: string;
  message_count?: number;
  source?: string;
}

const SessionsHistory: React.FC<{ onSelect?: (id: string) => void }> = ({ onSelect }) => {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('');

  useEffect(() => {
    fetch('/api/conversations')
      .then(r => r.json())
      .then(data => {
        if (data.success) setConversations(data.data || []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  const filtered = conversations.filter(c =>
    (c.title || '').toLowerCase().includes(filter.toLowerCase())
  );

  const formatDate = (d: string) => {
    try {
      return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
    } catch { return d; }
  };

  const sourceIcon = (src?: string) => src === 'cli' ? '💻' : '🌐';

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', gap: '0.8rem' }}>
      <input
        value={filter}
        onChange={e => setFilter(e.target.value)}
        placeholder="Search sessions..."
        style={{ background: 'var(--bg-input)', border: '1px solid var(--border)', color: 'var(--text-primary)', padding: '0.5rem 0.8rem', borderRadius: '6px', fontSize: '0.85rem', width: '100%', boxSizing: 'border-box' }}
      />

      <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '0.3rem' }}>
        {loading ? (
          <div style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>Loading...</div>
        ) : filtered.length === 0 ? (
          <div style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>No sessions found.</div>
        ) : (
          filtered.map(conv => (
            <div
              key={conv.id}
              onClick={() => onSelect?.(conv.id)}
              style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', padding: '0.7rem 1rem', cursor: 'pointer', transition: 'background 0.15s' }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-hover)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'var(--bg-card)')}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div style={{ flex: 1, overflow: 'hidden' }}>
                  <div style={{ color: 'var(--text-primary)', fontSize: '0.88rem', fontWeight: '500', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {sourceIcon(conv.source)} {conv.title || 'Untitled Session'}
                  </div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginTop: '2px' }}>
                    {formatDate(conv.created_at)}
                    {conv.message_count !== undefined && ` · ${conv.message_count} messages`}
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default SessionsHistory;
