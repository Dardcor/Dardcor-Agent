import React, { useState, useEffect } from 'react';

interface ThinkModeIndicatorProps {
  message?: string;
}

const THINK_KEYWORDS = ['think', 'thinking', 'reason', 'reasoning', 'step by step', 'chain of thought', 'cot', 'ultrawork', 'supreme', 'reflect'];

const ThinkModeIndicator: React.FC<ThinkModeIndicatorProps> = ({ message = '' }) => {
  const [active, setActive] = useState(false);

  useEffect(() => {
    const lower = message.toLowerCase();
    const detected = THINK_KEYWORDS.some(kw => lower.includes(kw));
    setActive(detected);
  }, [message]);

  if (!active) return null;

  return (
    <div style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: '0.3rem',
      background: 'rgba(189, 147, 249, 0.15)',
      border: '1px solid rgba(189, 147, 249, 0.4)',
      borderRadius: '20px',
      padding: '0.2rem 0.6rem',
      fontSize: '0.75rem',
      color: '#bd93f9',
      animation: 'pulse 1.5s infinite',
    }}>
      <span style={{ animation: 'spin 2s linear infinite', display: 'inline-block' }}>🧠</span>
      Think Mode
    </div>
  );
};

export default ThinkModeIndicator;
