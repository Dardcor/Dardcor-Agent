import React, { useEffect, useState } from 'react';
import { egoAPI } from '../services/api';

interface EgoState {
  confidence: number;
  energy: number;
  status: string;
  last_mood: string;
  tasks_complete: number;
  tasks_failed: number;
  streak_success: number;
  streak_failed: number;
  total_actions: number;
  last_error: string;
}

const MOOD_COLORS: Record<string, string> = {
  Euphoric: '#a78bfa',
  Proud: '#60a5fa',
  Confident: '#34d399',
  Satisfied: '#6ee7b7',
  Analytical: '#fbbf24',
  Determined: '#f97316',
  Frustrated: '#f87171',
  Neutral: '#94a3b8',
};

const SentienceMonitor: React.FC = () => {
  const [ego, setEgo] = useState<EgoState | null>(null);
  const [dreams, setDreams] = useState<string[]>([]);
  const [showDreams, setShowDreams] = useState(false);

  const fetchState = async () => {
    try {
      const state = await egoAPI.getState();
      setEgo(state);
    } catch {
      // silent
    }
  };

  const fetchDreams = async () => {
    try {
      const d = await egoAPI.getDreams(5);
      setDreams(d || []);
    } catch {
      // silent
    }
  };

  useEffect(() => {
    fetchState();
    fetchDreams();
    const si = setInterval(fetchState, 5000);
    const sd = setInterval(fetchDreams, 60000);
    return () => { clearInterval(si); clearInterval(sd); };
  }, []);

  if (!ego) return null;

  const total = ego.tasks_complete + ego.tasks_failed;
  const successRate = total > 0 ? Math.round((ego.tasks_complete / total) * 100) : 100;
  const mood = ego.last_mood || 'Neutral';
  const moodColor = MOOD_COLORS[mood] || '#94a3b8';

  const statusClass = (() => {
    const s = ego.status.toLowerCase();
    if (s.includes('transcend') || s.includes('supreme')) return 'status-supreme';
    if (s.includes('dominant') || s.includes('focused')) return 'status-active';
    if (s.includes('critical') || s.includes('recover')) return 'status-alert';
    if (s.includes('reflect')) return 'status-reflect';
    return 'status-default';
  })();

  return (
    <div className="sentience-monitor" title={`Mood: ${mood} | ${successRate}% success | ${ego.total_actions} actions`}>
      <div className="ego-state-row">
        <div className={`ego-pulse-wrap ${statusClass}`}>
          <div className="ego-pulse" />
        </div>

        <div className="ego-bars">
          <div className="ego-bar-row" title={`Confidence: ${Math.round(ego.confidence * 100)}%`}>
            <span className="ego-bar-label">C</span>
            <div className="ego-bar-track">
              <div
                className="ego-bar-fill confidence-fill"
                style={{ width: `${ego.confidence * 100}%` }}
              />
            </div>
          </div>
          <div className="ego-bar-row" title={`Energy: ${Math.round(ego.energy * 100)}%`}>
            <span className="ego-bar-label">E</span>
            <div className="ego-bar-track">
              <div
                className="ego-bar-fill energy-fill"
                style={{ width: `${ego.energy * 100}%`, backgroundColor: ego.energy < 0.3 ? '#f87171' : undefined }}
              />
            </div>
          </div>
        </div>

        <div className="ego-status-block">
          <span className={`ego-status-badge ${statusClass}`}>{ego.status}</span>
          <span className="ego-mood-dot" style={{ color: moodColor }} title={mood}>●</span>
        </div>

        {dreams.length > 0 && (
          <button
            className="ego-dreams-btn"
            onClick={() => setShowDreams(s => !s)}
            title="View recent dreams"
          >
            💤
          </button>
        )}
      </div>

      {showDreams && dreams.length > 0 && (
        <div className="ego-dreams-panel">
          <div className="ego-dreams-header">
            <span>🌙 Recent Insights</span>
            <button onClick={() => setShowDreams(false)}>×</button>
          </div>
          <ul className="ego-dreams-list">
            {dreams.slice().reverse().map((d, i) => (
              <li key={i} className="ego-dream-item">{d}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default SentienceMonitor;
