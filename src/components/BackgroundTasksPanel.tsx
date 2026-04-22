import React, { useState, useEffect, useCallback } from 'react';

interface BackgroundTask {
  id: string;
  prompt: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  result?: string;
  error?: string;
  created_at: string;
  completed_at?: string;
  conv_id?: string;
}

const BackgroundTasksPanel: React.FC = () => {
  const [tasks, setTasks] = useState<BackgroundTask[]>([]);
  const [newPrompt, setNewPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const [selectedTask, setSelectedTask] = useState<BackgroundTask | null>(null);

  const fetchTasks = useCallback(async () => {
    try {
      const res = await fetch('/api/background/tasks');
      const data = await res.json();
      if (data.success) setTasks(data.data || []);
    } catch { }
  }, []);

  useEffect(() => {
    fetchTasks();
    const interval = setInterval(fetchTasks, 3000);
    return () => clearInterval(interval);
  }, [fetchTasks]);

  const submitTask = async () => {
    if (!newPrompt.trim()) return;
    setLoading(true);
    try {
      const res = await fetch('/api/background/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt: newPrompt }),
      });
      const data = await res.json();
      if (data.success) {
        setNewPrompt('');
        fetchTasks();
      }
    } catch { }
    setLoading(false);
  };

  const cancelTask = async (id: string) => {
    await fetch('/api/background/cancel', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    });
    fetchTasks();
  };

  const purgeCompleted = async () => {
    await fetch('/api/background/purge', { method: 'POST' });
    fetchTasks();
    setSelectedTask(null);
  };

  const statusColor: Record<string, string> = {
    pending: '#f1fa8c',
    running: '#8be9fd',
    completed: '#50fa7b',
    failed: '#ff5555',
    cancelled: '#6272a4',
  };

  const statusIcon: Record<string, string> = {
    pending: '⏳',
    running: '🔄',
    completed: '✅',
    failed: '❌',
    cancelled: '🚫',
  };

  return (
    <div className="panel" style={{ padding: '1.5rem', height: '100%', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ color: 'var(--text-primary)', margin: 0 }}>🔁 Background Tasks</h2>
        <button onClick={purgeCompleted} style={{ background: 'var(--bg-hover)', border: 'none', color: 'var(--text-muted)', padding: '0.4rem 0.8rem', borderRadius: '6px', cursor: 'pointer' }}>
          Clear Done
        </button>
      </div>


      <div style={{ display: 'flex', gap: '0.5rem' }}>
        <input
          value={newPrompt}
          onChange={e => setNewPrompt(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && submitTask()}
          placeholder="Submit a background task (AI will run it async)..."
          style={{ flex: 1, background: 'var(--bg-input)', border: '1px solid var(--border)', color: 'var(--text-primary)', padding: '0.6rem 1rem', borderRadius: '8px', fontSize: '0.9rem' }}
        />
        <button
          onClick={submitTask}
          disabled={loading || !newPrompt.trim()}
          style={{ background: 'var(--accent)', border: 'none', color: '#fff', padding: '0.6rem 1.2rem', borderRadius: '8px', cursor: 'pointer', fontWeight: 'bold' }}
        >
          {loading ? '...' : 'Submit'}
        </button>
      </div>


      <div style={{ flex: 1, overflowY: 'auto', display: 'flex', gap: '0.5rem', flexDirection: 'column' }}>
        {tasks.length === 0 ? (
          <div style={{ color: 'var(--text-muted)', textAlign: 'center', marginTop: '2rem' }}>
            No background tasks. Submit a prompt above to run it asynchronously.
          </div>
        ) : (
          tasks.map(task => (
            <div
              key={task.id}
              onClick={() => setSelectedTask(task)}
              style={{
                background: selectedTask?.id === task.id ? 'var(--bg-hover)' : 'var(--bg-card)',
                border: `1px solid var(--border)`,
                borderLeft: `3px solid ${statusColor[task.status] || '#6272a4'}`,
                borderRadius: '8px',
                padding: '0.8rem 1rem',
                cursor: 'pointer',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span>{statusIcon[task.status]}</span>
                  <span style={{ color: 'var(--text-primary)', fontSize: '0.9rem', maxWidth: '400px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {task.prompt}
                  </span>
                </div>
                <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                  <span style={{ color: statusColor[task.status], fontSize: '0.75rem', fontWeight: 'bold' }}>
                    {task.status.toUpperCase()}
                  </span>
                  {(task.status === 'pending' || task.status === 'running') && (
                    <button
                      onClick={e => { e.stopPropagation(); cancelTask(task.id); }}
                      style={{ background: 'none', border: '1px solid var(--border)', color: 'var(--text-muted)', padding: '0.2rem 0.5rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem' }}
                    >
                      Cancel
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>


      {selectedTask && (
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: '8px', padding: '1rem', maxHeight: '250px', overflowY: 'auto' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>Task ID: {selectedTask.id.slice(0, 8)}</span>
            <button onClick={() => setSelectedTask(null)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>✕</button>
          </div>
          {selectedTask.result ? (
            <pre style={{ margin: 0, color: 'var(--text-primary)', fontSize: '0.85rem', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
              {selectedTask.result}
            </pre>
          ) : selectedTask.error ? (
            <pre style={{ margin: 0, color: '#ff5555', fontSize: '0.85rem' }}>{selectedTask.error}</pre>
          ) : (
            <span style={{ color: 'var(--text-muted)' }}>Task is {selectedTask.status}...</span>
          )}
        </div>
      )}
    </div>
  );
};

export default BackgroundTasksPanel;
