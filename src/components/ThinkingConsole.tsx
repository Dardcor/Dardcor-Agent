import React, { useEffect, useRef, useState } from 'react';

interface ThinkingConsoleProps {
  content: string;
  isStreaming?: boolean;
}

interface ParsedThought {
  thought: string;
  plan: string;
  reflection: string;
  isComplete: boolean;
}

const parseThought = (content: string): ParsedThought => {
  const extract = (tag: string): string => {
    const end = tag === 'THOUGHT' ? '(?:\\[PLAN\\]|\\[ACTION\\]|\\[REFLECTION\\]|\\[COMPLETE\\]|$)'
      : tag === 'PLAN' ? '(?:\\[ACTION\\]|\\[REFLECTION\\]|\\[COMPLETE\\]|$)'
      : '(?:\\[ACTION\\]|\\[COMPLETE\\]|$)';
    const re = new RegExp(`\\[${tag}\\]([\\s\\S]*?)${end}`, 'i');
    const m = content.match(re);
    return m ? m[1].trim() : '';
  };
  return {
    thought: extract('THOUGHT'),
    plan: extract('PLAN'),
    reflection: extract('REFLECTION'),
    isComplete: content.includes('[COMPLETE]'),
  };
};

const ThinkingConsole: React.FC<ThinkingConsoleProps> = ({ content, isStreaming }) => {
  const { thought, plan, reflection, isComplete } = parseThought(content);
  const [visible, setVisible] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const streamRef = useRef<HTMLDivElement>(null);

  const hasContent = thought || plan || reflection || isComplete;

  useEffect(() => {
    if (hasContent) setVisible(true);
  }, [hasContent]);

  useEffect(() => {
    if (streamRef.current) {
      streamRef.current.scrollTop = streamRef.current.scrollHeight;
    }
  }, [thought, plan, reflection]);

  if (!visible || !hasContent) return null;

  return (
    <div className={`thinking-console ${isComplete ? 'complete' : ''} ${isStreaming ? 'streaming' : ''}`}>
      <div className="thinking-header" onClick={() => setCollapsed(c => !c)}>
        <div className="thinking-header-left">
          <span className="thinking-icon">{isComplete ? '✅' : isStreaming ? '🧠' : '💭'}</span>
          <span className="thinking-title">
            {isComplete ? 'Execution Complete' : isStreaming ? 'Reasoning...' : 'Internal Monologue'}
          </span>
          {isStreaming && <span className="thinking-pulse" />}
        </div>
        <button className="thinking-collapse-btn" aria-label="toggle">
          {collapsed ? '▼' : '▲'}
        </button>
      </div>

      {!collapsed && (
        <div className="thinking-body" ref={streamRef}>
          {thought && (
            <div className="thinking-section">
              <div className="thinking-section-label">
                <span className="tl-icon">🔍</span> THOUGHT
              </div>
              <div className="thinking-text stream-text">{thought}</div>
            </div>
          )}
          {plan && (
            <div className="thinking-section">
              <div className="thinking-section-label">
                <span className="tl-icon">📋</span> PLAN
              </div>
              <div className="thinking-text plan-text">{plan}</div>
            </div>
          )}
          {reflection && (
            <div className="thinking-section">
              <div className="thinking-section-label">
                <span className="tl-icon">🪞</span> REFLECTION
              </div>
              <div className="thinking-text reflection-text">{reflection}</div>
            </div>
          )}
          {isComplete && (
            <div className="thinking-complete-badge">
              <span>★</span> Objective Mastered
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ThinkingConsole;
