import React, { useState, useEffect, useMemo } from 'react';

interface Tool {
  id: string;
  name: string;
  description: string;
  icon: string;
  status: 'active' | 'inactive';
  category: string;
  mcp?: boolean;
  last_used?: string;
  usage_count?: number;
}

const ToolsPanel: React.FC = () => {
  const [tools, setTools] = useState<Tool[]>([
    { id: 'fs', name: 'File Core', description: 'Advanced file manipulation and recursive directory analysis.', icon: '📂', status: 'active', category: 'System' },
    { id: 'terminal', name: 'Shell Executor', description: 'Raw command-line interface for system-level operations.', icon: '🐚', status: 'active', category: 'System' },
    { id: 'web', name: 'Neural Search', description: 'High-speed web crawling and real-time knowledge extraction.', icon: '🧠', status: 'active', category: 'Intelligence' },
    { id: 'code', name: 'Sandbox Engine', description: 'Isolated execution environment for secure code processing.', icon: '⚡', status: 'active', category: 'Logic' },
    { id: 'git', name: 'Version Control', description: 'Repository synchronization and conflict resolution logic.', icon: '🌿', status: 'inactive', category: 'Development' },
    { id: 'db', name: 'Data Architect', description: 'Structured data querying and schema optimization.', icon: '💎', status: 'inactive', category: 'System' },
    { id: 'ai', name: 'Creative Forge', description: 'Generative AI for advanced media synthesis and editing.', icon: '✨', status: 'active', category: 'Media' },
    { id: 'api', name: 'Bridge Protocol', description: 'Secure API orchestration and protocol translation.', icon: '🧬', status: 'active', category: 'Network' },
  ]);

  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [collapsedCategories, setCollapsedCategories] = useState<Record<string, boolean>>({});

  useEffect(() => {
    fetch('/api/tools/config')
      .then(res => res.json())
      .then(res => {
        if (res.success && res.data) {
          setTools(res.data);
        }
        setIsLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch tools config:', err);
        setIsLoading(false);
      });
  }, []);

  const saveTools = (updatedTools: Tool[]) => {
    fetch('/api/tools/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updatedTools)
    }).catch(err => console.error('Failed to save tools config:', err));
  };

  const toggleTool = (id: string) => {
    const updatedTools = tools.map(tool =>
      tool.id === id
        ? { ...tool, status: (tool.status === 'active' ? 'inactive' : 'active') as 'active' | 'inactive' }
        : tool
    );
    setTools(updatedTools);
    saveTools(updatedTools);
  };

  const toggleCategory = (category: string) => {
    setCollapsedCategories(prev => ({
      ...prev,
      [category]: !prev[category]
    }));
  };

  const filteredTools = useMemo(() => {
    return tools.filter(tool =>
      tool.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tool.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tool.category.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [tools, searchQuery]);

  const groupedTools = useMemo(() => {
    const groups: Record<string, Tool[]> = {};
    filteredTools.forEach(tool => {
      if (!groups[tool.category]) {
        groups[tool.category] = [];
      }
      groups[tool.category].push(tool);
    });
    return groups;
  }, [filteredTools]);

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    return isNaN(date.getTime()) ? 'Never' : date.toLocaleString();
  };

  if (isLoading) {
    return (
      <div className="tools-panel">
        <div className="loading-spinner">
          <div className="spinner"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="tools-panel">
      <div className="panel-content-wrapper">
        <div className="panel-header-custom">
          <div className="header-info">
            <div className="title-wrapper">
              <h3>Agent Tools</h3>
              <span className="efficiency-badge">Efficiency: Ultra</span>
            </div>
            <p>Manage and configure the capabilities available to your agent. All tools are optimized for minimal resource consumption.</p>
          </div>
          <div className="header-stats">
            <div className="stat-item">
              <span className="stat-value">{tools.filter(t => t.status === 'active').length}</span>
              <span className="stat-label">Active</span>
            </div>
            <div className="stat-item">
              <span className="stat-value">{tools.length}</span>
              <span className="stat-label">Total</span>
            </div>
          </div>
        </div>

        <div className="search-box" style={{ marginBottom: '20px', height: '42px' }}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ marginRight: '10px', color: 'var(--text-dim)' }}><circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" /></svg>
          <input
            type="text"
            placeholder="Search tools by name, description, or category..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              color: 'var(--text-primary)',
              fontSize: '14px',
              outline: 'none'
            }}
          />
        </div>

        <div className="tools-list">
          {Object.entries(groupedTools).map(([category, categoryTools]) => (
            <div key={category} style={{ marginBottom: '24px' }}>
              <div
                onClick={() => toggleCategory(category)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  cursor: 'pointer',
                  padding: '10px 0',
                  borderBottom: '1px solid var(--border-subtle)',
                  marginBottom: '16px'
                }}
              >
                <h3 style={{ margin: 0, fontSize: '16px', color: 'var(--text-primary)' }}>
                  {category} <span style={{ color: 'var(--text-dim)', fontSize: '12px', marginLeft: '8px' }}>({categoryTools.length})</span>
                </h3>
                <span style={{
                  transform: collapsedCategories[category] ? 'rotate(-90deg)' : 'rotate(0deg)',
                  transition: 'transform 0.2s ease',
                  color: 'var(--text-dim)'
                }}>
                  ▼
                </span>
              </div>

              {!collapsedCategories[category] && (
                <div className="tools-grid">
                  {categoryTools.map(tool => (
                    <div key={tool.id} className={`tool-card ${tool.status}`}>
                      <div className="tool-card-header">
                        <div className="tool-icon">{tool.icon}</div>
                        <div className="tool-status-badge">
                          <span className="status-dot"></span>
                          {tool.status}
                        </div>
                      </div>
                      <div className="tool-info">
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                          <span className="tool-category">{tool.category}</span>
                          {tool.mcp && (
                            <span style={{
                              fontSize: '10px',
                              padding: '2px 6px',
                              background: 'var(--accent-glow-sm)',
                              color: 'var(--accent-primary)',
                              borderRadius: '4px',
                              fontWeight: 'bold',
                              border: '1px solid var(--accent-primary)'
                            }}>
                              MCP
                            </span>
                          )}
                        </div>
                        <h4>{tool.name}</h4>
                        <p>{tool.description}</p>
                        <div style={{ marginTop: '12px', fontSize: '11px', color: 'var(--text-dim)', display: 'flex', justifyContent: 'space-between' }}>
                          <span>Used: {tool.usage_count || 0} times</span>
                          <span>Last: {formatDate(tool.last_used)}</span>
                        </div>
                      </div>
                      <div className="tool-actions">
                        <button className="configure-btn">Configure</button>
                        <div className="toggle-switch" onClick={() => toggleTool(tool.id)}>
                          <input type="checkbox" checked={tool.status === 'active'} readOnly />
                          <span className="slider"></span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
          {Object.keys(groupedTools).length === 0 && (
            <div style={{ textAlign: 'center', padding: '40px 0', color: 'var(--text-dim)' }}>
              No tools match your search criteria.
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ToolsPanel;




