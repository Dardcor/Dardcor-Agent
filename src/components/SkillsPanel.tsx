import React, { useState, useEffect } from 'react';

interface Skill {
  id: string;
  name: string;
  description: string;
  level: number; // 0 to 100
  tags: string[];
  icon: string;
}

const SkillsPanel: React.FC = () => {
  const [skills, setSkills] = useState<Skill[]>([
    { id: 'web-dev', name: 'Web Development', description: 'Expertise in modern frontend frameworks, backend systems, and responsive design.', level: 95, tags: ['React', 'NodeJS', 'CSS'], icon: '⚛️' },
    { id: 'data-sci', name: 'Data Engineering', description: 'Processing large datasets, building pipelines, and performing complex analysis.', level: 82, tags: ['Python', 'SQL', 'Pandas'], icon: '📊' },
    { id: 'devops', name: 'Cloud Architecture', description: 'Deploying applications, managing infrastructure, and CI/CD automation.', level: 78, tags: ['Docker', 'AWS', 'K8s'], icon: '☁️' },
    { id: 'sec', name: 'Cyber Security', description: 'Vulnerability assessment, secure coding practices, and threat mitigation.', level: 65, tags: ['PenTest', 'Auth', 'Encryption'], icon: '🛡️' },
    { id: 'ml', name: 'Machine Learning', description: 'Training models, fine-tuning LLMs, and implementing RAG systems.', level: 88, tags: ['LLM', 'PyTorch', 'VectorDB'], icon: '🧠' },
    { id: 'seo', name: 'SEO Strategy', description: 'Optimizing content for search engines and improving digital visibility.', level: 92, tags: ['Growth', 'Analysis', 'Content'], icon: '📈' },
  ]);

  const [isLoading, setIsLoading] = useState(true);

  // Fetch skills from database on mount
  useEffect(() => {
    fetch('/api/skills/config')
      .then(res => res.json())
      .then(res => {
        if (res.success && res.data) {
          setSkills(res.data);
        }
        setIsLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch skills config:', err);
        setIsLoading(false);
      });
  }, []);

  // Save skills to database
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

  if (isLoading) {
    return (
      <div className="skills-panel">
        <div className="loading-spinner">
          <div className="spinner"></div>
        </div>
      </div>
    );
  }

  const averageMastery = Math.round(skills.reduce((acc, curr) => acc + curr.level, 0) / skills.length);

  return (
    <div className="skills-panel">
      <div className="panel-content-wrapper">
        <div className="skills-header-card">
          <div className="header-info">
            <div className="title-wrapper">
              <h3>Agent Skills</h3>
              <span className="savings-badge">Cognitive Savings: 40%</span>
            </div>
            <p>Cognitive proficiencies mastered by the agent, using optimized pathways to reduce computation and token usage.</p>
          </div>
          <div className="mastery-overview">
            <div className="mastery-orb">
              <svg viewBox="0 0 100 100">
                <circle cx="50" cy="50" r="45" className="orb-bg" />
                <circle 
                  cx="50" 
                  cy="50" 
                  r="45" 
                  className="orb-fill" 
                  style={{ strokeDashoffset: 282.7 * (1 - averageMastery / 100) }} 
                />
              </svg>
              <div className="orb-text">
                <span className="mastery-percent">{averageMastery}%</span>
                <span className="mastery-label">Avg Mastery</span>
              </div>
            </div>
          </div>
        </div>

        <div className="skills-list">
          {skills.map(skill => (
            <div key={skill.id} className="skill-item">
              <div className="skill-icon-wrap">
                <span className="skill-icon">{skill.icon}</span>
              </div>
              <div className="skill-content">
                <div className="skill-top">
                  <div className="skill-title-group">
                    <h4>{skill.name}</h4>
                    <div className="skill-tags">
                      {skill.tags.map(tag => <span key={tag} className="skill-tag">{tag}</span>)}
                    </div>
                  </div>
                  <div className="skill-level-text" onClick={() => updateSkillLevel(skill.id, skill.level + 1)}>
                    Lvl {Math.floor(skill.level / 10)}
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
        </div>
      </div>
    </div>
  );
};

export default SkillsPanel;
