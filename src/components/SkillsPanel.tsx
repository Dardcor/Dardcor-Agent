import React, { useState, useEffect, useMemo } from 'react';

interface Skill {
  id: string;
  name: string;
  description: string;
  level: number;
  tags: string[];
  icon: string;
  category?: string;
  enabled?: boolean;
}

const SkillsPanel: React.FC = () => {
  const [skills, setSkills] = useState<Skill[]>([
    { id: 'web-dev', name: 'Web Development', description: 'Expertise in modern frontend frameworks, backend systems, and responsive design.', level: 95, tags: ['React', 'NodeJS', 'CSS'], icon: '⚛️', category: 'Development', enabled: true },
    { id: 'data-sci', name: 'Data Engineering', description: 'Processing large datasets, building pipelines, and performing complex analysis.', level: 82, tags: ['Python', 'SQL', 'Pandas'], icon: '📊', category: 'Data', enabled: true },
    { id: 'devops', name: 'Cloud Architecture', description: 'Deploying applications, managing infrastructure, and CI/CD automation.', level: 78, tags: ['Docker', 'AWS', 'K8s'], icon: '☁️', category: 'Operations', enabled: true },
    { id: 'sec', name: 'Cyber Security', description: 'Vulnerability assessment, secure coding practices, and threat mitigation.', level: 65, tags: ['PenTest', 'Auth', 'Encryption'], icon: '🛡️', category: 'Security', enabled: true },
    { id: 'ml', name: 'Machine Learning', description: 'Training models, fine-tuning LLMs, and implementing RAG systems.', level: 88, tags: ['LLM', 'PyTorch', 'VectorDB'], icon: '🧠', category: 'Data', enabled: true },
    { id: 'seo', name: 'SEO Strategy', description: 'Optimizing content for search engines and improving digital visibility.', level: 92, tags: ['Growth', 'Analysis', 'Content'], icon: '📈', category: 'Marketing', enabled: true },
  ]);

  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeCategory, setActiveCategory] = useState<string>('All');

  useEffect(() => {
    fetch('/api/skills/config')
      .then(res => res.json())
      .then(res => {
        if (res.success && res.data) {
          setSkills(res.data.map((s: any) => ({
            ...s,
            category: s.category || 'Uncategorized',
            enabled: s.enabled !== false
          })));
        }
        setIsLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch skills config:', err);
        setIsLoading(false);
      });
  }, []);

  const saveSkills = (updatedSkills: Skill[]) => {
    fetch('/api/skills/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updatedSkills)
    }).catch(err => console.error('Failed to save skills config:', err));
  };

  const updateSkillLevel = (id: string, newLevel: number) => {
    const updatedSkills = skills.map(skill =>
      skill.id === id ? { ...skill, level: Math.min(100, Math.max(0, newLevel)) } : skill
    );
    setSkills(updatedSkills);
    saveSkills(updatedSkills);
  };

  const toggleSkill = (id: string) => {
    const updatedSkills = skills.map(skill =>
      skill.id === id ? { ...skill, enabled: !skill.enabled } : skill
    );
    setSkills(updatedSkills);
    saveSkills(updatedSkills);
  };

  const categories = useMemo(() => {
    const cats = new Set<string>();
    skills.forEach(s => cats.add(s.category || 'Uncategorized'));
    return ['All', ...Array.from(cats)].sort();
  }, [skills]);

  const categoryCounts = useMemo(() => {
    const counts: Record<string, number> = { All: skills.length };
    skills.forEach(s => {
      const cat = s.category || 'Uncategorized';
      counts[cat] = (counts[cat] || 0) + 1;
    });
    return counts;
  }, [skills]);

  const filteredSkills = useMemo(() => {
    return skills.filter(skill => {
      const matchesSearch = skill.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        skill.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        skill.tags.some(t => t.toLowerCase().includes(searchQuery.toLowerCase()));
      const matchesCategory = activeCategory === 'All' || (skill.category || 'Uncategorized') === activeCategory;
      return matchesSearch && matchesCategory;
    });
  }, [skills, searchQuery, activeCategory]);

  if (isLoading) {
    return (
      <div className="skills-panel">
        <div className="loading-spinner">
          <div className="spinner"></div>
        </div>
      </div>
    );
  }

  const activeSkillsCount = skills.filter(s => s.enabled).length;
  const averageMastery = skills.length > 0 ? Math.round(skills.reduce((acc, curr) => acc + curr.level, 0) / skills.length) : 0;

  return (
    <div className="skills-panel">
      <div className="panel-content-wrapper">
        <div className="skills-header-card">
          <div className="header-info">
            <div className="title-wrapper">
              <h3>Agent Skills</h3>
              <span className="efficiency-badge">Cognitive Efficiency: +45%</span>
            </div>
            <p>Neural pathways and mastered proficiencies optimized for low-latency task execution and minimal token consumption.</p>
            <div className="header-stats" style={{ marginTop: '16px' }}>
              <div className="stat-item">
                <span className="stat-value">{activeSkillsCount}</span>
                <span className="stat-label">Active</span>
              </div>
              <div className="stat-divider" style={{ width: '1px', background: 'var(--border-subtle)', height: '24px' }}></div>
              <div className="stat-item">
                <span className="stat-value">{skills.length - activeSkillsCount}</span>
                <span className="stat-label">Dormant</span>
              </div>
              <div className="stat-divider" style={{ width: '1px', background: 'var(--border-subtle)', height: '24px' }}></div>
              <div className="stat-item">
                <span className="stat-value">{categories.length - 1}</span>
                <span className="stat-label">Domains</span>
              </div>
            </div>
          </div>
          <div className="mastery-overview">
            <div className="mastery-orb">
              <svg viewBox="0 0 100 100">
                <circle cx="50" cy="50" r="42" className="orb-bg" />
                <circle
                  cx="50"
                  cy="50"
                  r="42"
                  className="orb-fill"
                  style={{ strokeDasharray: '263.8', strokeDashoffset: 263.8 * (1 - averageMastery / 100) }}
                />
              </svg>
              <div className="orb-text">
                <span className="mastery-percent">{averageMastery}%</span>
                <span className="mastery-label">Mastery</span>
              </div>
            </div>
          </div>
        </div>

        <div style={{ marginBottom: '24px', display: 'flex', gap: '12px', alignItems: 'center' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <span style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', opacity: 0.5 }}>🔍</span>
            <input
              type="text"
              placeholder="Search skills, tags, or expertise..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{
                width: '100%',
                padding: '12px 15px 12px 40px',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)',
                background: 'var(--bg-card)',
                color: '#fff',
                fontSize: '14px',
                outline: 'none',
                transition: 'var(--transition-fast)'
              }}
            />
          </div>
          <select
            value={activeCategory}
            onChange={(e) => setActiveCategory(e.target.value)}
            style={{
              padding: '12px 15px',
              borderRadius: '12px',
              border: '1px solid var(--border-subtle)',
              background: 'var(--bg-card)',
              color: '#fff',
              fontSize: '14px',
              outline: 'none',
              cursor: 'pointer',
              minWidth: '140px'
            }}
          >
            {categories.map(cat => (
              <option key={cat} value={cat}>{cat} ({categoryCounts[cat]})</option>
            ))}
          </select>
        </div>

        <div className="skills-list">
          {filteredSkills.map(skill => (
            <div key={skill.id} className="skill-item" style={{ opacity: skill.enabled ? 1 : 0.5 }}>
              <div className="skill-icon-wrap">
                <span className="skill-icon">{skill.icon}</span>
              </div>
              <div className="skill-content">
                <div className="skill-top">
                  <div className="skill-title-group">
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                      <h4>{skill.name}</h4>
                      <span className="skill-category-badge">{skill.category}</span>
                    </div>
                    <div className="skill-tags">
                      {skill.tags.map(tag => <span key={tag} className="skill-tag">{tag}</span>)}
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '18px' }}>
                    <div className="skill-level-text" onClick={() => updateSkillLevel(skill.id, skill.level + 1)}>
                      LV {Math.floor(skill.level / 10)}
                    </div>
                    <div className="toggle-switch" onClick={() => toggleSkill(skill.id)}>
                      <input type="checkbox" checked={skill.enabled} readOnly />
                      <span className="slider"></span>
                    </div>
                  </div>
                </div>
                <p className="skill-desc">{skill.description}</p>
                <div className="skill-progress-container">
                  <div className="skill-progress-bar">
                    <div className="skill-progress-fill" style={{ width: `${skill.level}%` }}>
                      <div className="skill-progress-glow"></div>
                    </div>
                  </div>
                  <span className="skill-percent">{skill.level}%</span>
                </div>
              </div>
            </div>
          ))}
          {filteredSkills.length === 0 && (
            <div style={{ textAlign: 'center', padding: '60px 0', border: '1px dashed var(--border-subtle)', borderRadius: '16px' }}>
              <div style={{ fontSize: '32px', marginBottom: '10px' }}>🔍</div>
              <div style={{ color: 'var(--text-dim)', fontSize: '14px' }}>No internal skills found matching "{searchQuery}"</div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SkillsPanel;




