import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import AntigravityView from '../model/antigravity'
import GeminiView from '../model/gemini'
import OpenRouterView from '../model/openrouter'

const ModelSelector: React.FC = () => {
  const { provider } = useParams<{ provider: string }>()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'antigravity' | 'gemini' | 'openrouter'>('antigravity')

  useEffect(() => {
    if (provider === 'gemini' || provider === 'openrouter' || provider === 'antigravity') {
      setActiveTab(provider as any)
    }
  }, [provider])

  const handleTabChange = (tab: 'antigravity' | 'gemini' | 'openrouter') => {
    setActiveTab(tab)
    navigate(`/model/${tab}`)
  }

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
        <button className={activeTab === 'antigravity' ? 'active' : ''} onClick={() => handleTabChange('antigravity')}>Antigravity</button>
        <button className={activeTab === 'gemini' ? 'active' : ''} onClick={() => handleTabChange('gemini')}>Gemini</button>
        <button className={activeTab === 'openrouter' ? 'active' : ''} onClick={() => handleTabChange('openrouter')}>OpenRouter</button>
      </div>
      
      <div className="model-content">
        {renderContent()}
      </div>
    </div>
  )
}

export default ModelSelector




