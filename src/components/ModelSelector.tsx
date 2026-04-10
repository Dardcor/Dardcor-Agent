import React, { useState } from 'react'
import AntigravityView from '../model/antigravity'
import GeminiView from '../model/gemini'
import OpenRouterView from '../model/openrouter'

const ModelSelector: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'antigravity' | 'gemini' | 'openrouter'>('antigravity')

  const renderContent = () => {
    switch(activeTab) {
      case 'antigravity': return <AntigravityView />
      case 'gemini':      return <GeminiView />
      case 'openrouter':  return <OpenRouterView />
    }
  }

  return (
    <div className="model-selector-page">
      <div className="model-navbar">
        <button className={activeTab === 'antigravity' ? 'active' : ''} onClick={() => setActiveTab('antigravity')}>Antigravity</button>
        <button className={activeTab === 'gemini' ? 'active' : ''} onClick={() => setActiveTab('gemini')}>Gemini</button>
        <button className={activeTab === 'openrouter' ? 'active' : ''} onClick={() => setActiveTab('openrouter')}>OpenRouter</button>
      </div>
      
      <div className="model-content">
        {renderContent()}
      </div>
    </div>
  )
}

export default ModelSelector




