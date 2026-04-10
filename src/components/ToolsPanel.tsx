import React, { useState, useEffect } from 'react';

interface Tool {
  id: string;
  name: string;
  description: string;
  icon: string;
  status: 'active' | 'inactive';
  category: string;
}

const ToolsPanel: React.FC = () => {
  const [tools, setTools] = useState<Tool[]>([
    { id: 'fs', name: 'File System', description: 'Read, write, and manage local files and directories safely.', icon: '📁', status: 'active', category: 'System' },
    { id: 'terminal', name: 'Terminal', description: 'Execute shell commands and manage system processes.', icon: '💻', status: 'active', category: 'System' },
    { id: 'web', name: 'Web Browser', description: 'Search the web, browse websites, and extract information.', icon: '🌐', status: 'active', category: 'Network' },
    { id: 'code', name: 'Code Interpreter', description: 'Execute Javascript/TypeScript code in a secure sandbox.', icon: '📜', status: 'active', category: 'Logic' },
    { id: 'git', name: 'Git Manager', description: 'Manage repositories, commits, branches, and merges.', icon: '🌿', status: 'inactive', category: 'Development' },
    { id: 'db', name: 'Database', description: 'Query and manipulate local or remote databases.', icon: '🗄️', status: 'inactive', category: 'System' },
    { id: 'ai', name: 'Image Gen', description: 'Generate and edit images using AI models.', icon: '🎨', status: 'active', category: 'Media' },
    { id: 'api', name: 'HTTP Client', description: 'Make API requests and handle various data formats.', icon: '🔗', status: 'active', category: 'Network' },
  ]);

  const [isLoading, setIsLoading] = useState(true);

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

        <div className="tools-grid">
          {tools.map(tool => (
            <div key={tool.id} className={`tool-card ${tool.status}`}>
              <div className="tool-card-header">
                <div className="tool-icon">{tool.icon}</div>
                <div className="tool-status-badge">
                  <span className="status-dot"></span>
                  {tool.status}
                </div>
              </div>
              <div className="tool-info">
                <span className="tool-category">{tool.category}</span>
                <h4>{tool.name}</h4>
                <p>{tool.description}</p>
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
      </div>
    </div>
  );
};

export default ToolsPanel;




