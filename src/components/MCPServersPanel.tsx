import React, { useState, useEffect } from 'react';

interface MCPServer {
  name: string;
  command: string;
  args: string[];
  status: string;
  available: boolean;
  added_at: string;
}

const BUILTIN_SERVERS = [
  { name: 'context7', command: 'npx', args: ['-y', '@upstash/context7-mcp@latest'], desc: 'Up-to-date library docs' },
  { name: 'grep-app', command: 'npx', args: ['-y', '@grep-app/mcp@latest'], desc: 'Web code search' },
  { name: 'websearch', command: 'npx', args: ['-y', '@modelcontextprotocol/server-brave-search'], desc: 'Brave web search' },
  { name: 'filesystem', command: 'npx', args: ['-y', '@modelcontextprotocol/server-filesystem', '.'], desc: 'Local filesystem access' },
  { name: 'github', command: 'npx', args: ['-y', '@modelcontextprotocol/server-github'], desc: 'GitHub API' },
  { name: 'memory', command: 'npx', args: ['-y', '@modelcontextprotocol/server-memory'], desc: 'Persistent memory' },
];

const MCPServersPanel: React.FC = () => {
  const [servers, setServers] = useState<MCPServer[]>([]);
  const [newName, setNewName] = useState('');
  const [newCommand, setNewCommand] = useState('');
  const [newArgs, setNewArgs] = useState('');
  const [showAdd, setShowAdd] = useState(false);
  const [loading, setLoading] = useState(false);

  const fetchServers = async () => {
    try {
      const res = await fetch('/api/mcp/servers');
      const data = await res.json();
      if (data.success) setServers(data.data || []);
    } catch {}
  };

  useEffect(() => { fetchServers(); }, []);

  const addServer = async () => {
    if (!newName.trim() || !newCommand.trim()) return;
    setLoading(true);
    try {
      await fetch('/api/mcp/servers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newName, command: newCommand, args: newArgs.split(' ').filter(Boolean) }),
      });
      setNewName(''); setNewCommand(''); setNewArgs(''); setShowAdd(false);
      fetchServers();
    } catch {}
    setLoading(false);
  };

  const addBuiltin = async (srv: typeof BUILTIN_SERVERS[0]) => {
    setLoading(true);
    try {
      await fetch('/api/mcp/servers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: srv.name, command: srv.command, args: srv.args }),
      });
      fetchServers();
    } catch {}
    setLoading(false);
  };

  const removeServer = async (name: string) => {
    await fetch(`/api/mcp/servers?name=${encodeURIComponent(name)}`, { method: 'DELETE' });
    fetchServers();
  };

  return (
    <div className="panel" style={{ padding: '1.5rem', height: '100%', display: 'flex', flexDirection: 'column', gap: '1rem', overflowY: 'auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ color: 'var(--text-primary)', margin: 0 }}>🔌 MCP Servers</h2>
        <button onClick={() => setShowAdd(!showAdd)} style={{ background: 'var(--accent)', border: 'none', color: '#fff', padding: '0.4rem 0.8rem', borderRadius: '6px', cursor: 'pointer' }}>
          + Add
        </button>
      </div>

      {showAdd && (
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="Server name" style={inputStyle} />
          <input value={newCommand} onChange={e => setNewCommand(e.target.value)} placeholder="Command (e.g. npx)" style={inputStyle} />
          <input value={newArgs} onChange={e => setNewArgs(e.target.value)} placeholder="Args (space-separated)" style={inputStyle} />
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button onClick={addServer} disabled={loading} style={{ background: 'var(--accent)', border: 'none', color: '#fff', padding: '0.4rem 1rem', borderRadius: '6px', cursor: 'pointer' }}>Add</button>
            <button onClick={() => setShowAdd(false)} style={{ background: 'var(--bg-hover)', border: 'none', color: 'var(--text-muted)', padding: '0.4rem 1rem', borderRadius: '6px', cursor: 'pointer' }}>Cancel</button>
          </div>
        </div>
      )}

      {/* Installed servers */}
      {servers.length > 0 && (
        <div>
          <h3 style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.1em', margin: '0 0 0.5rem' }}>Installed</h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
            {servers.map(srv => (
              <div key={srv.name} style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', padding: '0.7rem 1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <span style={{ width: '8px', height: '8px', borderRadius: '50%', background: srv.available ? '#50fa7b' : '#ff5555', display: 'inline-block' }}></span>
                    <span style={{ color: 'var(--text-primary)', fontWeight: 'bold', fontSize: '0.9rem' }}>{srv.name}</span>
                  </div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem', marginTop: '2px' }}>
                    {srv.command} {(srv.args || []).join(' ')}
                  </div>
                </div>
                <button onClick={() => removeServer(srv.name)} style={{ background: 'none', border: '1px solid #ff5555', color: '#ff5555', padding: '0.2rem 0.6rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem' }}>Remove</button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Builtin servers */}
      <div>
        <h3 style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.1em', margin: '0 0 0.5rem' }}>Quick Install</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
          {BUILTIN_SERVERS.map(srv => (
            <div key={srv.name} style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', padding: '0.7rem 1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ color: 'var(--text-primary)', fontWeight: 'bold', fontSize: '0.85rem' }}>{srv.name}</div>
                <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>{srv.desc}</div>
              </div>
              <button onClick={() => addBuiltin(srv)} disabled={loading} style={{ background: 'var(--bg-hover)', border: '1px solid var(--border)', color: 'var(--text-primary)', padding: '0.2rem 0.5rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem' }}>+ Add</button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

const inputStyle: React.CSSProperties = {
  background: 'var(--bg-input)',
  border: '1px solid var(--border)',
  color: 'var(--text-primary)',
  padding: '0.5rem 0.8rem',
  borderRadius: '6px',
  fontSize: '0.9rem',
  width: '100%',
  boxSizing: 'border-box',
};

export default MCPServersPanel;
