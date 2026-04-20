import React, { useState } from 'react';

interface Plugin {
  name: string;
  description?: string;
  version?: string;
  installed: boolean;
}

const FEATURED_PLUGINS: Plugin[] = [
  { name: 'dardcor-plugin-docker', description: 'Docker container management', version: '1.0.0', installed: false },
  { name: 'dardcor-plugin-git', description: 'Advanced git operations', version: '1.2.0', installed: false },
  { name: 'dardcor-plugin-testing', description: 'Test runner integration', version: '0.9.0', installed: false },
  { name: 'dardcor-plugin-database', description: 'Database query assistant', version: '1.1.0', installed: false },
  { name: 'dardcor-plugin-security', description: 'Security audit tools', version: '0.8.0', installed: false },
  { name: 'dardcor-plugin-api', description: 'API testing and mocking', version: '1.0.5', installed: false },
];

const PluginManager: React.FC = () => {
  const [plugins, setPlugins] = useState<Plugin[]>(FEATURED_PLUGINS);
  const [customName, setCustomName] = useState('');
  const [installing, setInstalling] = useState<string | null>(null);

  const installPlugin = async (name: string) => {
    setInstalling(name);
    // Simulate install delay
    await new Promise(r => setTimeout(r, 1000));
    setPlugins(prev => prev.map(p => p.name === name ? { ...p, installed: true } : p));
    setInstalling(null);
  };

  const removePlugin = (name: string) => {
    setPlugins(prev => prev.map(p => p.name === name ? { ...p, installed: false } : p));
  };

  const installCustom = async () => {
    if (!customName.trim()) return;
    const newPlugin: Plugin = { name: customName, description: 'Custom plugin', installed: true };
    setPlugins(prev => [...prev, newPlugin]);
    setCustomName('');
  };

  const createPlugin = () => {
    alert('Use CLI: dardcor plugin create <name>');
  };

  return (
    <div className="panel" style={{ padding: '1.5rem', height: '100%', display: 'flex', flexDirection: 'column', gap: '1rem', overflowY: 'auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ color: 'var(--text-primary)', margin: 0 }}>🧩 Plugin Manager</h2>
        <button onClick={createPlugin} style={{ background: 'var(--bg-hover)', border: '1px solid var(--border)', color: 'var(--text-muted)', padding: '0.4rem 0.8rem', borderRadius: '6px', cursor: 'pointer', fontSize: '0.85rem' }}>
          + Create Plugin
        </button>
      </div>

      {/* Custom install */}
      <div style={{ display: 'flex', gap: '0.5rem' }}>
        <input
          value={customName}
          onChange={e => setCustomName(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && installCustom()}
          placeholder="npm package name (e.g. dardcor-plugin-x)"
          style={{ flex: 1, background: 'var(--bg-input)', border: '1px solid var(--border)', color: 'var(--text-primary)', padding: '0.5rem 0.8rem', borderRadius: '6px', fontSize: '0.85rem' }}
        />
        <button onClick={installCustom} style={{ background: 'var(--accent)', border: 'none', color: '#fff', padding: '0.5rem 1rem', borderRadius: '6px', cursor: 'pointer' }}>Install</button>
      </div>

      {/* Installed plugins */}
      {plugins.filter(p => p.installed).length > 0 && (
        <div>
          <h3 style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.1em', margin: '0 0 0.5rem' }}>Installed</h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
            {plugins.filter(p => p.installed).map(plugin => (
              <div key={plugin.name} style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderLeft: '3px solid #50fa7b', borderRadius: '8px', padding: '0.7rem 1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ color: 'var(--text-primary)', fontWeight: 'bold', fontSize: '0.9rem' }}>{plugin.name}</div>
                  {plugin.description && <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem' }}>{plugin.description}</div>}
                </div>
                <button onClick={() => removePlugin(plugin.name)} style={{ background: 'none', border: '1px solid #ff5555', color: '#ff5555', padding: '0.2rem 0.6rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem' }}>Remove</button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Featured plugins */}
      <div>
        <h3 style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.1em', margin: '0 0 0.5rem' }}>Featured Plugins</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem' }}>
          {plugins.filter(p => !p.installed).map(plugin => (
            <div key={plugin.name} style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', padding: '0.8rem 1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ color: 'var(--text-primary)', fontWeight: 'bold', fontSize: '0.85rem' }}>{plugin.name.replace('dardcor-plugin-', '')}</div>
                <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>{plugin.description}</div>
              </div>
              <button
                onClick={() => installPlugin(plugin.name)}
                disabled={installing === plugin.name}
                style={{ background: 'var(--bg-hover)', border: '1px solid var(--border)', color: 'var(--text-primary)', padding: '0.2rem 0.6rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem' }}
              >
                {installing === plugin.name ? '...' : '+ Install'}
              </button>
            </div>
          ))}
        </div>
      </div>

      <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem', marginTop: 'auto' }}>
        💡 Plugins extend Dardcor with new commands, tools, and integrations. Use <code style={{ background: 'var(--bg-input)', padding: '0.1rem 0.3rem', borderRadius: '3px' }}>dardcor plugin list</code> in CLI.
      </div>
    </div>
  );
};

export default PluginManager;
