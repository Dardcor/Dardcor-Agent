import React, { useState, useEffect } from 'react';

const Workspace: React.FC = () => {
  const [path, setPath] = useState('');
  const [status, setStatus] = useState<'idle' | 'success' | 'error'>('idle');

  useEffect(() => {
    const savedPath = localStorage.getItem('dardcor_workspace_path');
    if (savedPath) {
      setPath(savedPath);
    } else {
      // Saran default untuk Windows
      setPath('C:\\Users\\syahr\\Documents\\Dardcor-Workspace');
    }
  }, []);

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    if (path.trim()) {
      localStorage.setItem('dardcor_workspace_path', path.trim());
      setStatus('success');
      setTimeout(() => setStatus('idle'), 3000);
    } else {
      setStatus('error');
    }
  };

  const handleApplyDefault = () => {
    setPath('C:\\Users\\syahr\\Documents\\Dardcor-Workspace');
  };

  return (
    <div className="workspace-container">
      <div className="workspace-card">
        <div className="workspace-header">
          <div className="workspace-icon">🛠️</div>
          <div className="workspace-title">
            <h2>Konfigurasi Workspace</h2>
            <p>Tentukan direktori kerja utama agar Agent dapat mengelola proyek Anda dengan aman.</p>
          </div>
        </div>

        <form className="workspace-form" onSubmit={handleSave}>
          <div className="form-group">
            <label htmlFor="workspace-path">Path Direktori Utama</label>
            <div className="input-with-button">
              <input
                type="text"
                id="workspace-path"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="Contoh: C:\Users\Username\Documents\MyProject"
                spellCheck={false}
              />
              <button type="submit" className="save-btn">
                Simpan
              </button>
            </div>
            
            {!localStorage.getItem('dardcor_workspace_path') && (
              <div className="default-suggestion">
                <span>💡 Belum diatur?</span>
                <button type="button" onClick={handleApplyDefault} className="suggest-link">
                  Gunakan saran default
                </button>
              </div>
            )}
            
            <p className="hint">
              <strong>Tips:</strong> Gunakan path absolut (drive letter pertama) untuk performa terbaik.
            </p>
          </div>
        </form>

        {status === 'success' && (
          <div className="workspace-alert success">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
            Workspace berhasil diperbarui dan disimpan!
          </div>
        )}

        <div className="workspace-guide">
          <h3>Panduan Penggunaan</h3>
          <div className="guide-grid">
            <div className="guide-item">
              <h4>🔒 Batas Keamanan</h4>
              <p>Membatasi jangkauan Agent agar tidak sembarangan masuk ke folder sistem sensitif.</p>
            </div>
            <div className="guide-item">
              <h4>🚀 Akses Cepat</h4>
              <p>Terminal dan File Explorer akan otomatis terbuka di folder ini saat dimulai.</p>
            </div>
            <div className="guide-item">
              <h4>📂 Struktur Rapi</h4>
              <p>Memudahkan pelacakan file hasil kerja Agent karena semuanya terpusat di satu folder.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Workspace;
